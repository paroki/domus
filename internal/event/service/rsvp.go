package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/eventrsvp"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

func (s *Service) canCallerRSVP(ctx context.Context, p authx.Principal, ev *ent.Event) (bool, error) {
	if ev.OrgID != nil {
		myRole, err := s.resolveMemberRole(ctx, p.UserID, *ev.OrgID)
		if err != nil {
			return false, err
		}
		return myRole != nil, nil
	}
	isAdm, err := s.isParokiAdmin(ctx, p)
	if err != nil {
		return false, err
	}
	if isAdm {
		return true, nil
	}
	return s.repo.IsParokiMember(ctx, p.ParokiID, p.UserID)
}

// UpsertRSVP creates or updates the caller's RSVP for eventID (requires org membership or paroki membership for parish-level events).
func (s *Service) UpsertRSVP(
	ctx context.Context,
	p authx.Principal,
	eventID uuid.UUID,
	req domain.UpsertEventRSVPRequest,
) (domain.EventRSVPResponse, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventRSVPResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventRSVPResponse{}, err
	}

	canRSVP, err := s.canCallerRSVP(ctx, p, ev)
	if err != nil {
		return domain.EventRSVPResponse{}, err
	}
	if !canRSVP {
		return domain.EventRSVPResponse{}, domain.ErrForbidden
	}

	if !s.isRSVPOpen(ev, time.Now().UTC()) {
		return domain.EventRSVPResponse{}, domain.ErrRSVPClosed
	}

	var note *string
	if req.Note != nil {
		trimmed := strings.TrimSpace(*req.Note)
		note = &trimmed
	}

	var saved *ent.EventRSVP
	if err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		var txErr error
		saved, txErr = txRepo.UpsertRSVP(ctx, p.ParokiID, eventID, p.UserID, eventrsvp.Status(req.Status), note)
		return txErr
	}); err != nil {
		return domain.EventRSVPResponse{}, err
	}

	var u *ent.User
	var m *ent.OrgMember
	if ev.OrgID != nil {
		usersMap, membersMap, err := s.repo.LoadUsersAndOrgMembers(ctx, *ev.OrgID, []uuid.UUID{p.UserID})
		if err != nil {
			return domain.EventRSVPResponse{}, err
		}
		u = usersMap[p.UserID]
		m = membersMap[p.UserID]
	} else {
		usersMap, err := s.repo.LoadUsers(ctx, []uuid.UUID{p.UserID})
		if err != nil {
			return domain.EventRSVPResponse{}, err
		}
		u = usersMap[p.UserID]
	}

	return toRSVPResponse(saved, u, m), nil
}

// DeleteRSVP hard-deletes the caller's RSVP for eventID.
func (s *Service) DeleteRSVP(ctx context.Context, p authx.Principal, eventID uuid.UUID) error {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}

	canRSVP, err := s.canCallerRSVP(ctx, p, ev)
	if err != nil {
		return err
	}
	if !canRSVP {
		return domain.ErrForbidden
	}

	if !s.isRSVPOpen(ev, time.Now().UTC()) {
		return domain.ErrRSVPClosed
	}

	return s.repo.DeleteRSVP(ctx, eventID, p.UserID)
}

// GetRSVPSummary returns RSVP attendance summary counts for eventID.
func (s *Service) GetRSVPSummary(ctx context.Context, p authx.Principal, eventID uuid.UUID) (domain.EventRSVPSummaryResponse, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventRSVPSummaryResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventRSVPSummaryResponse{}, err
	}

	canRSVP, err := s.canCallerRSVP(ctx, p, ev)
	if err != nil {
		return domain.EventRSVPSummaryResponse{}, err
	}
	if !canRSVP {
		return domain.EventRSVPSummaryResponse{}, domain.ErrForbidden
	}

	counts, err := s.repo.GetRSVPSummary(ctx, p.ParokiID, eventID, ev.OrgID)
	if err != nil {
		return domain.EventRSVPSummaryResponse{}, err
	}
	return toRSVPSummaryResponse(eventID, counts), nil
}

// SearchRSVPs returns paginated attendee RSVPs for eventID.
func (s *Service) SearchRSVPs(
	ctx context.Context,
	p authx.Principal,
	eventID uuid.UUID,
	req domain.SearchEventRSVPsRequest,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.EventRSVPResponse, bool, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return nil, false, domain.ErrEventNotFound
	}
	if err != nil {
		return nil, false, err
	}

	canRSVP, err := s.canCallerRSVP(ctx, p, ev)
	if err != nil {
		return nil, false, err
	}
	if !canRSVP {
		return nil, false, domain.ErrForbidden
	}

	var status string
	if req.Filter != nil {
		status = req.Filter.Status
	}

	rows, usersMap, membersMap, err := s.repo.SearchRSVPs(
		ctx,
		eventID,
		ev.OrgID,
		strings.TrimSpace(req.Q),
		status,
		cursorAt,
		cursorID,
		limit+1,
	)
	if err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(rows) > limit {
		hasMore = true
		rows = rows[:limit]
	}

	out := make([]domain.EventRSVPResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRSVPResponse(row, usersMap[row.UserID], membersMap[row.UserID]))
	}
	return out, hasMore, nil
}

func (s *Service) isRSVPOpen(ev *ent.Event, now time.Time) bool {
	if ev.Status == event.StatusCancelled || !ev.RsvpEnabled {
		return false
	}
	cutoff := ev.StartAt.UTC().Add(s.cfg.RSVPGracePeriod)
	if ev.RsvpDeadline != nil {
		cutoff = ev.RsvpDeadline.UTC()
	}
	return !now.After(cutoff)
}

func toRSVPResponse(row *ent.EventRSVP, u *ent.User, m *ent.OrgMember) domain.EventRSVPResponse {
	res := domain.EventRSVPResponse{
		ID:        row.ID,
		ParokiID:  row.ParokiID,
		EventID:   row.EventID,
		UserID:    row.UserID,
		Status:    string(row.Status),
		Note:      row.Note,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if u != nil {
		res.Name = u.Name
		res.Email = u.Email
		res.Phone = u.Phone
	}
	if m != nil {
		res.OrgRole = string(m.Role)
	}
	return res
}

func toRSVPSummaryResponse(eventID uuid.UUID, counts repo.RSVPSummaryCounts) domain.EventRSVPSummaryResponse {
	totalResponded := counts.Going + counts.Maybe + counts.NotGoing
	unresponded := max(counts.TotalActiveMembers-totalResponded, 0)
	return domain.EventRSVPSummaryResponse{
		EventID:            eventID,
		Going:              counts.Going,
		Maybe:              counts.Maybe,
		NotGoing:           counts.NotGoing,
		TotalResponded:     totalResponded,
		TotalActiveMembers: counts.TotalActiveMembers,
		Unresponded:        unresponded,
	}
}
