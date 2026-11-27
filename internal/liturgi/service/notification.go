package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/liturgiministry"
	"github.com/paroki/domus/internal/event"
)

const (
	// DefaultParokiTimezone is the fallback IANA timezone when a paroki has none configured.
	DefaultParokiTimezone = "Asia/Makassar"

	notifKindLiturgiAssigned           = "liturgi_assigned"
	notifKindLiturgiSubstituteProposed = "liturgi_substitute_proposed"
	notifKindLiturgiSubstituteApproved = "liturgi_substitute_approved"
	notifKindLiturgiReminder           = "liturgi_reminder"

	quietStartHour = 22
	quietEndHour   = 6
)

// ResolveParokiLocation resolves the IANA time.Location for parokiID, falling back to Asia/Makassar.
func (s *Service) ResolveParokiLocation(ctx context.Context, parokiID uuid.UUID) *time.Location {
	tz := DefaultParokiTimezone
	if s.refs != nil && parokiID != uuid.Nil {
		if pTZ, err := s.refs.GetParokiTimezone(ctx, parokiID); err == nil && strings.TrimSpace(pTZ) != "" {
			tz = strings.TrimSpace(pTZ)
		}
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	loc, _ := time.LoadLocation(DefaultParokiTimezone)
	return loc
}

func (s *Service) expandAssigneeRecipients(ctx context.Context, userID, orgID *uuid.UUID, excludeUserID uuid.UUID) []uuid.UUID {
	if userID != nil && *userID != uuid.Nil {
		if *userID == excludeUserID {
			return nil
		}
		return []uuid.UUID{*userID}
	}
	if orgID != nil && *orgID != uuid.Nil {
		members, err := s.repo.ListActiveOrgMemberUserIDs(ctx, *orgID)
		if err != nil {
			return nil
		}
		out := make([]uuid.UUID, 0, len(members))
		for _, uid := range members {
			if uid != uuid.Nil && uid != excludeUserID {
				out = append(out, uid)
			}
		}
		return out
	}
	return nil
}

func (s *Service) notifyPetugasAssigned(ctx context.Context, actorID uuid.UUID, ev event.Event, row *ent.LiturgiPetugas) {
	if s.notifier == nil || row == nil {
		return
	}
	recipients := s.expandAssigneeRecipients(ctx, row.UserID, row.OrgID, actorID)
	if len(recipients) == 0 {
		return
	}
	window := "instant:" + row.ID.String() + ":" + strconv.FormatInt(row.UpdatedAt.UnixNano(), 10)
	_, _ = s.notifier.NotifyUsers(ctx, recipients, &ev.ParokiID, notifKindLiturgiAssigned, ev.ID, window, NotificationPayload{
		Title: "Tugas Liturgi: " + ev.Title,
		Body:  fmt.Sprintf("Anda ditugaskan sebagai %s.", string(row.Role)),
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "liturgi-" + ev.ID.String(),
		URL:   "/liturgi/" + ev.ID.String(),
	})
}

func (s *Service) notifySubstituteProposed(ctx context.Context, actorID uuid.UUID, ev event.Event, row *ent.LiturgiPetugas) {
	if s.notifier == nil || row == nil {
		return
	}
	ministry, err := s.repo.FindMinistry(ctx, ev.ParokiID, liturgiministry.RoleSlug(row.Role))
	if err != nil {
		return
	}
	seen := make(map[uuid.UUID]struct{})
	recipients := make([]uuid.UUID, 0, 4)
	if ministry.LeaderUserID != nil && *ministry.LeaderUserID != uuid.Nil && *ministry.LeaderUserID != actorID {
		seen[*ministry.LeaderUserID] = struct{}{}
		recipients = append(recipients, *ministry.LeaderUserID)
	}
	if ministry.OrgID != nil && *ministry.OrgID != uuid.Nil && s.orgs != nil {
		members, err := s.repo.ListActiveOrgMemberUserIDs(ctx, *ministry.OrgID)
		if err == nil {
			for _, uid := range members {
				if uid == uuid.Nil || uid == actorID {
					continue
				}
				isAdm, err := s.orgs.IsAdmin(ctx, *ministry.OrgID, uid)
				if err == nil && isAdm {
					if _, ok := seen[uid]; !ok {
						seen[uid] = struct{}{}
						recipients = append(recipients, uid)
					}
				}
			}
		}
	}
	if len(recipients) == 0 {
		return
	}
	body := fmt.Sprintf("%s berhalangan untuk tugas %s.", row.Name, string(row.Role))
	if row.SubstituteName != nil && strings.TrimSpace(*row.SubstituteName) != "" {
		body = fmt.Sprintf("%s berhalangan (%s) dan mengusulkan pengganti: %s.", row.Name, string(row.Role), strings.TrimSpace(*row.SubstituteName))
	}
	window := "instant:" + row.ID.String() + ":" + strconv.FormatInt(row.UpdatedAt.UnixNano(), 10)
	_, _ = s.notifier.NotifyUsers(ctx, recipients, &ev.ParokiID, notifKindLiturgiSubstituteProposed, ev.ID, window, NotificationPayload{
		Title: "Usulan Pengganti Liturgi: " + ev.Title,
		Body:  body,
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "liturgi-sub-" + ev.ID.String(),
		URL:   "/liturgi/" + ev.ID.String(),
	})
}

func (s *Service) notifySubstituteApproved(ctx context.Context, actorID uuid.UUID, ev event.Event, prev, updated *ent.LiturgiPetugas) {
	if s.notifier == nil || prev == nil || updated == nil {
		return
	}
	seen := make(map[uuid.UUID]struct{})
	recipients := make([]uuid.UUID, 0, 4)
	for _, uid := range s.expandAssigneeRecipients(ctx, updated.UserID, updated.OrgID, actorID) {
		if _, ok := seen[uid]; !ok {
			seen[uid] = struct{}{}
			recipients = append(recipients, uid)
		}
	}
	for _, uid := range s.expandAssigneeRecipients(ctx, prev.UserID, prev.OrgID, actorID) {
		if _, ok := seen[uid]; !ok {
			seen[uid] = struct{}{}
			recipients = append(recipients, uid)
		}
	}
	if len(recipients) == 0 {
		return
	}
	window := "instant:" + updated.ID.String() + ":" + strconv.FormatInt(updated.UpdatedAt.UnixNano(), 10)
	_, _ = s.notifier.NotifyUsers(ctx, recipients, &ev.ParokiID, notifKindLiturgiSubstituteApproved, ev.ID, window, NotificationPayload{
		Title: "Pengganti Disetujui: " + ev.Title,
		Body:  fmt.Sprintf("%s disetujui bertugas sebagai %s.", updated.Name, string(updated.Role)),
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "liturgi-sub-" + ev.ID.String(),
		URL:   "/liturgi/" + ev.ID.String(),
	})
}

// RunReminderTick dispatches due h-24h and h-1h liturgy officer reminders respecting user quiet hours (22:00-06:00).
func (s *Service) RunReminderTick(ctx context.Context, now time.Time) error {
	if s.notifier == nil {
		return nil
	}
	now = now.UTC()
	_, _ = s.notifier.CleanupExpiredNotificationLogs(ctx, now.Add(-7*24*time.Hour))

	events, err := s.repo.ListScheduledMassEventsBetween(ctx, now, now.Add(36*time.Hour))
	if err != nil {
		return err
	}
	fallbackLoc, _ := time.LoadLocation(DefaultParokiTimezone)

	for _, ev := range events {
		petugasRows, err := s.repo.ListActivePetugasForEvent(ctx, ev.ID)
		if err != nil {
			return err
		}
		if len(petugasRows) == 0 {
			continue
		}

		seen := make(map[uuid.UUID]struct{})
		recipients := make([]uuid.UUID, 0, len(petugasRows))
		rolesByUser := make(map[uuid.UUID][]string)
		orgMembersCache := make(map[uuid.UUID][]uuid.UUID)

		addRole := func(uid uuid.UUID, role string) {
			if uid == uuid.Nil {
				return
			}
			if _, ok := seen[uid]; !ok {
				seen[uid] = struct{}{}
				recipients = append(recipients, uid)
			}
			for _, existing := range rolesByUser[uid] {
				if existing == role {
					return
				}
			}
			rolesByUser[uid] = append(rolesByUser[uid], role)
		}

		for _, row := range petugasRows {
			if row.UserID != nil && *row.UserID != uuid.Nil {
				addRole(*row.UserID, string(row.Role))
			} else if row.OrgID != nil && *row.OrgID != uuid.Nil {
				members, ok := orgMembersCache[*row.OrgID]
				if !ok {
					members, err = s.repo.ListActiveOrgMemberUserIDs(ctx, *row.OrgID)
					if err != nil {
						return err
					}
					orgMembersCache[*row.OrgID] = members
				}
				for _, uid := range members {
					addRole(uid, string(row.Role))
				}
			}
		}
		if len(recipients) == 0 {
			continue
		}

		tzMap, err := s.notifier.ResolveUserTimezones(ctx, recipients)
		if err != nil {
			return err
		}

		for _, uid := range recipients {
			loc := fallbackLoc
			if tz := strings.TrimSpace(tzMap[uid]); tz != "" {
				if parsed, lErr := time.LoadLocation(tz); lErr == nil {
					loc = parsed
				}
			}
			localNow := now.In(loc)
			if localNow.Hour() >= quietStartHour || localNow.Hour() < quietEndHour {
				continue
			}

			roles := rolesByUser[uid]
			roleSummary := strings.Join(roles, ", ")

			// h-24h window with quiet-hour shift earlier to 20:00 local.
			ideal24 := ev.StartAt.Add(-24 * time.Hour)
			local24 := ideal24.In(loc)
			effective24 := ideal24
			if local24.Hour() >= quietStartHour {
				y, m, d := local24.Date()
				effective24 = time.Date(y, m, d, 20, 0, 0, 0, loc)
			} else if local24.Hour() < quietEndHour {
				prev := local24.AddDate(0, 0, -1)
				y, m, d := prev.Date()
				effective24 = time.Date(y, m, d, 20, 0, 0, 0, loc)
			}
			if !now.Before(effective24) && now.Before(ev.StartAt.Add(-2*time.Hour)) {
				_, _ = s.notifier.NotifyUser(ctx, uid, &ev.ParokiID, notifKindLiturgiReminder, ev.ID, "h-24h", NotificationPayload{
					Title: "Pengingat Tugas Liturgi: " + ev.Title,
					Body:  fmt.Sprintf("Misa dimulai dalam 24 jam (%s). Tugas Anda: %s.", ev.StartAt.In(loc).Format("15:04 MST"), roleSummary),
					Icon:  "/logo-192.png",
					Badge: "/logo-192.png",
					Tag:   "liturgi-reminder-" + ev.ID.String(),
					URL:   "/liturgi/" + ev.ID.String(),
				})
			}

			// h-1h window (held until 06:00 local if quiet hours, already gated above).
			ideal1 := ev.StartAt.Add(-1 * time.Hour)
			if !now.Before(ideal1) && now.Before(ev.StartAt) {
				_, _ = s.notifier.NotifyUser(ctx, uid, &ev.ParokiID, notifKindLiturgiReminder, ev.ID, "h-1h", NotificationPayload{
					Title: "Segera Bertugas: " + ev.Title,
					Body:  fmt.Sprintf("Misa dimulai pukul %s. Tugas Anda: %s.", ev.StartAt.In(loc).Format("15:04 MST"), roleSummary),
					Icon:  "/logo-192.png",
					Badge: "/logo-192.png",
					Tag:   "liturgi-reminder-" + ev.ID.String(),
					URL:   "/liturgi/" + ev.ID.String(),
				})
			}
		}
	}
	return nil
}

// StartReminderScheduler launches a background ticker invoking RunReminderTick until ctx is cancelled.
func (s *Service) StartReminderScheduler(ctx context.Context, interval time.Duration, logger *slog.Logger) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case tickAt := <-ticker.C:
				if err := s.RunReminderTick(ctx, tickAt); err != nil && logger != nil {
					logger.WarnContext(ctx, "liturgi reminder tick failed", "error", err)
				}
			}
		}
	}()
}
