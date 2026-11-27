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
	"github.com/paroki/domus/ent/event"
)

const (
	notifKindEventCreated   = "event_created"
	notifKindEventUpdated   = "event_updated"
	notifKindEventCancelled = "event_cancelled"
	notifKindEventReminder  = "event_reminder"

	quietStartHour = 22
	quietEndHour   = 6
)

func (s *Service) notifyEventCreated(ctx context.Context, creatorID uuid.UUID, created *ent.Event) {
	if s.notifier == nil || created == nil || created.OrgID == nil {
		return
	}
	memberIDs, err := s.repo.ListActiveOrgMemberUserIDs(ctx, *created.OrgID)
	if err != nil {
		return
	}
	recipients := make([]uuid.UUID, 0, len(memberIDs))
	for _, uid := range memberIDs {
		if uid != creatorID {
			recipients = append(recipients, uid)
		}
	}
	if len(recipients) == 0 {
		return
	}
	_, _ = s.notifier.NotifyUsers(ctx, recipients, &created.ParokiID, notifKindEventCreated, created.ID, "instant", NotificationPayload{
		Title: created.Title,
		Body:  "Kegiatan organisasi baru telah dijadwalkan.",
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "event-" + created.ID.String(),
		URL:   "/events/" + created.ID.String(),
	})
}

func (s *Service) notifyEventUpdated(ctx context.Context, existing, updated *ent.Event) {
	if s.notifier == nil || existing == nil || updated == nil {
		return
	}
	if existing.Status != event.StatusScheduled || updated.Status != event.StatusScheduled || updated.StartAt.Equal(existing.StartAt) {
		return
	}
	recipients, _, err := s.resolveEventRecipients(ctx, updated.ID)
	if err != nil || len(recipients) == 0 {
		return
	}
	window := "instant:" + strconv.FormatInt(updated.UpdatedAt.Unix(), 10)
	_, _ = s.notifier.NotifyUsers(ctx, recipients, &updated.ParokiID, notifKindEventUpdated, updated.ID, window, NotificationPayload{
		Title: updated.Title,
		Body:  "Jadwal kegiatan telah diperbarui.",
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "event-" + updated.ID.String(),
		URL:   "/events/" + updated.ID.String(),
	})
}

func (s *Service) notifyEventCancelled(ctx context.Context, existing, updated *ent.Event) {
	if s.notifier == nil || existing == nil || updated == nil {
		return
	}
	if existing.Status == event.StatusCancelled || updated.Status != event.StatusCancelled {
		return
	}
	recipients, _, err := s.resolveEventRecipients(ctx, updated.ID)
	if err != nil || len(recipients) == 0 {
		return
	}
	_, _ = s.notifier.NotifyUsers(ctx, recipients, &updated.ParokiID, notifKindEventCancelled, updated.ID, "instant", NotificationPayload{
		Title: updated.Title,
		Body:  "Kegiatan ini telah dibatalkan.",
		Icon:  "/logo-192.png",
		Badge: "/logo-192.png",
		Tag:   "event-" + updated.ID.String(),
		URL:   "/events/" + updated.ID.String(),
	})
}

func (s *Service) resolveEventRecipients(ctx context.Context, eventID uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]string, error) {
	attendeeIDs, err := s.repo.ListAttendeeUserIDsForEvent(ctx, eventID)
	if err != nil {
		return nil, nil, err
	}

	seen := make(map[uuid.UUID]struct{}, len(attendeeIDs))
	recipients := make([]uuid.UUID, 0, len(attendeeIDs))
	rolesByUser := make(map[uuid.UUID][]string)

	for _, uid := range attendeeIDs {
		if uid == uuid.Nil {
			continue
		}
		if _, ok := seen[uid]; !ok {
			seen[uid] = struct{}{}
			recipients = append(recipients, uid)
		}
	}
	return recipients, rolesByUser, nil
}

// RunReminderTick dispatches due h-24h and h-1h event reminders respecting user quiet hours (22:00-06:00).
func (s *Service) RunReminderTick(ctx context.Context, now time.Time) error {
	if s.notifier == nil {
		return nil
	}
	now = now.UTC()
	_, _ = s.notifier.CleanupExpiredNotificationLogs(ctx, now.Add(-7*24*time.Hour))

	events, err := s.repo.ListScheduledEventsBetween(ctx, now, now.Add(36*time.Hour))
	if err != nil {
		return err
	}
	fallbackLoc, _ := time.LoadLocation(DefaultParokiTimezone)

	for _, ev := range events {
		recipients, rolesByUser, err := s.resolveEventRecipients(ctx, ev.ID)
		if err != nil {
			return err
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
			roleSuffix := ""
			if len(roles) > 0 {
				roleSuffix = fmt.Sprintf(" Tugas Anda: %s.", strings.Join(roles, ", "))
			}

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
				_, _ = s.notifier.NotifyUser(ctx, uid, &ev.ParokiID, notifKindEventReminder, ev.ID, "h-24h", NotificationPayload{
					Title: "Pengingat Kegiatan: " + ev.Title,
					Body:  fmt.Sprintf("Dimulai dalam 24 jam (%s).%s", ev.StartAt.In(loc).Format("15:04 MST"), roleSuffix),
					Icon:  "/logo-192.png",
					Badge: "/logo-192.png",
					Tag:   "event-reminder-" + ev.ID.String(),
					URL:   "/events/" + ev.ID.String(),
				})
			}

			// h-1h window (held until 06:00 local if quiet hours, already gated above).
			ideal1 := ev.StartAt.Add(-1 * time.Hour)
			if !now.Before(ideal1) && now.Before(ev.StartAt) {
				_, _ = s.notifier.NotifyUser(ctx, uid, &ev.ParokiID, notifKindEventReminder, ev.ID, "h-1h", NotificationPayload{
					Title: "Segera Dimulai: " + ev.Title,
					Body:  fmt.Sprintf("Dimulai pukul %s.%s", ev.StartAt.In(loc).Format("15:04 MST"), roleSuffix),
					Icon:  "/logo-192.png",
					Badge: "/logo-192.png",
					Tag:   "event-reminder-" + ev.ID.String(),
					URL:   "/events/" + ev.ID.String(),
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
					logger.WarnContext(ctx, "event reminder tick failed", "error", err)
				}
			}
		}
	}()
}
