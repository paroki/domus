package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

// SearchEvents returns paginated visible events in the caller's active parish.
func (s *Service) SearchEvents(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchEventsRequest,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.EventResponse, bool, error) {
	myOrgs, err := s.orgs.ListUserOrgs(ctx, p.ParishID, p.UserID)
	if err != nil {
		return nil, false, err
	}

	orgMap := make(map[uuid.UUID]OrgInfo, len(myOrgs))
	memberOrgIDs := make([]uuid.UUID, 0, len(myOrgs))
	for _, o := range myOrgs {
		orgMap[o.ID] = o
		memberOrgIDs = append(memberOrgIDs, o.ID)
	}

	in := repo.SearchEventsInput{
		ParishID:     p.ParishID,
		MemberOrgIDs: memberOrgIDs,
		Q:            strings.TrimSpace(req.Q),
		CursorAt:     cursorAt,
		CursorID:     cursorID,
		Limit:        limit + 1,
	}
	if req.Filter != nil {
		in.OrgID = req.Filter.OrgID
		in.Visibility = req.Filter.Visibility
		in.Status = req.Filter.Status
		in.From = req.Filter.From
		in.To = req.Filter.To
		if req.Filter.MyOrgsOnly != nil {
			in.MyOrgsOnly = *req.Filter.MyOrgsOnly
		}
	}

	rows, err := s.repo.SearchEvents(ctx, in)
	if err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(rows) > limit {
		hasMore = true
		rows = rows[:limit]
	}

	now := time.Now().UTC()
	out := make([]domain.EventResponse, 0, len(rows))
	roleCache := make(map[uuid.UUID]*string)
	memberEventOrgMap := make(map[uuid.UUID]uuid.UUID)
	memberEventIDs := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		info, ok := orgMap[row.OrgID]
		if !ok {
			fetched, err := s.orgs.GetOrganization(ctx, p.ParishID, row.OrgID)
			if err != nil {
				return nil, false, err
			}
			info = *fetched
			orgMap[row.OrgID] = info
		}

		myRole, cached := roleCache[row.OrgID]
		if !cached {
			myRole, err = s.resolveMyOrgRole(ctx, p.UserID, row.OrgID, memberOrgIDs)
			if err != nil {
				return nil, false, err
			}
			roleCache[row.OrgID] = myRole
		}

		if myRole != nil {
			memberEventOrgMap[row.ID] = row.OrgID
			memberEventIDs = append(memberEventIDs, row.ID)
		}

		out = append(out, s.toEventResponse(row, info, myRole, now))
	}

	if len(memberEventIDs) > 0 {
		myRSVPs, err := s.repo.FindUserRSVPsForEvents(ctx, memberEventIDs, p.UserID)
		if err != nil {
			return nil, false, err
		}
		summaries, err := s.repo.GetRSVPSummariesForEvents(ctx, memberEventOrgMap)
		if err != nil {
			return nil, false, err
		}
		// Load caller's User + OrgMember info per org for myRsvp enrichment.
		orgIDsForCaller := make(map[uuid.UUID]struct{})
		for _, orgID := range memberEventOrgMap {
			orgIDsForCaller[orgID] = struct{}{}
		}
		callerUsers := make(map[uuid.UUID]*ent.User)
		callerMembers := make(map[uuid.UUID]*ent.OrgMember)
		for orgID := range orgIDsForCaller {
			uMap, mMap, err := s.repo.LoadUsersAndOrgMembers(ctx, orgID, []uuid.UUID{p.UserID})
			if err != nil {
				return nil, false, err
			}
			callerUsers[orgID] = uMap[p.UserID]
			callerMembers[orgID] = mMap[p.UserID]
		}

		for i := range out {
			if out[i].MyOrgRole == nil {
				continue
			}
			if rsvpRow, ok := myRSVPs[out[i].ID]; ok {
				r := toRSVPResponse(rsvpRow, callerUsers[out[i].OrgID], callerMembers[out[i].OrgID])
				out[i].MyRsvp = &r
			}
			if counts, ok := summaries[out[i].ID]; ok {
				sum := toRSVPSummaryResponse(out[i].ID, counts)
				out[i].RsvpSummary = &sum
			}
		}
	}
	return out, hasMore, nil
}

// GetEvent returns a single non-deleted Event in the caller's active parish, enforcing private visibility.
func (s *Service) GetEvent(ctx context.Context, p authx.Principal, eventID uuid.UUID) (domain.EventResponse, error) {
	row, err := s.repo.FindEventInParish(ctx, p.ParishID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventResponse{}, err
	}

	myRole, err := s.resolveMemberRole(ctx, p.UserID, row.OrgID)
	if err != nil {
		return domain.EventResponse{}, err
	}
	if row.Visibility == event.VisibilityPrivate && myRole == nil {
		return domain.EventResponse{}, domain.ErrForbidden
	}

	info, err := s.orgs.GetOrganization(ctx, p.ParishID, row.OrgID)
	if err != nil {
		return domain.EventResponse{}, err
	}

	res := s.toEventResponse(row, *info, myRole, time.Now().UTC())
	if myRole != nil {
		counts, err := s.repo.GetRSVPSummary(ctx, row.ID, row.OrgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		sum := toRSVPSummaryResponse(row.ID, counts)
		res.RsvpSummary = &sum

		rsvpRow, err := s.repo.FindUserRSVP(ctx, row.ID, p.UserID)
		if err != nil && !ent.IsNotFound(err) {
			return domain.EventResponse{}, err
		}
		if rsvpRow != nil {
			uMap, mMap, err := s.repo.LoadUsersAndOrgMembers(ctx, row.OrgID, []uuid.UUID{p.UserID})
			if err != nil {
				return domain.EventResponse{}, err
			}
			r := toRSVPResponse(rsvpRow, uMap[p.UserID], mMap[p.UserID])
			res.MyRsvp = &r
		}
	}
	return res, nil
}

// CreateEvent creates a new Event owned by req.OrgID (requires org:admin of req.OrgID).
func (s *Service) CreateEvent(ctx context.Context, p authx.Principal, req domain.CreateEventRequest) (domain.EventResponse, error) {
	info, err := s.orgs.GetOrganization(ctx, p.ParishID, req.OrgID)
	if err != nil {
		return domain.EventResponse{}, err
	}

	isAdm, err := s.isOrgAdmin(ctx, p, req.OrgID)
	if err != nil {
		return domain.EventResponse{}, err
	}
	if !isAdm {
		return domain.EventResponse{}, domain.ErrForbidden
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return domain.EventResponse{}, domain.ValidationError("title", "REQUIRED", "Event title is required.")
	}

	startAt := req.StartAt.UTC()
	endAt := req.EndAt.UTC()
	var rsvpDeadline *time.Time
	if req.RsvpDeadline != nil {
		d := req.RsvpDeadline.UTC()
		rsvpDeadline = &d
	}
	if err := validateTimeRange(startAt, endAt, rsvpDeadline); err != nil {
		return domain.EventResponse{}, err
	}

	vis := event.Visibility(s.cfg.DefaultVisibility)
	if req.Visibility != "" {
		vis = event.Visibility(req.Visibility)
	}
	rsvpEnabled := true
	if req.RsvpEnabled != nil {
		rsvpEnabled = *req.RsvpEnabled
	}

	created, err := s.repo.CreateEvent(ctx, repo.CreateEventInput{
		ParishID:     p.ParishID,
		OrgID:        req.OrgID,
		CreatedBy:    p.UserID,
		Title:        title,
		Description:  trimOptString(req.Description),
		Location:     trimOptString(req.Location),
		StartAt:      startAt,
		EndAt:        endAt,
		Visibility:   vis,
		RsvpEnabled:  rsvpEnabled,
		RsvpDeadline: rsvpDeadline,
	})
	if err != nil {
		return domain.EventResponse{}, err
	}

	adminRole := "admin"
	return s.toEventResponse(created, *info, &adminRole, time.Now().UTC()), nil
}

// UpdateEvent updates an existing Event in the active parish (requires org:admin of owning org).
func (s *Service) UpdateEvent(ctx context.Context, p authx.Principal, eventID uuid.UUID, req domain.UpdateEventRequest) (domain.EventResponse, error) {
	existing, err := s.repo.FindEventInParish(ctx, p.ParishID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventResponse{}, err
	}

	isAdm, err := s.isOrgAdmin(ctx, p, existing.OrgID)
	if err != nil {
		return domain.EventResponse{}, err
	}
	if !isAdm {
		return domain.EventResponse{}, domain.ErrForbidden
	}

	in := repo.UpdateEventInput{
		Description: trimOptString(req.Description),
		Location:    trimOptString(req.Location),
		RsvpEnabled: req.RsvpEnabled,
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return domain.EventResponse{}, domain.ValidationError("title", "REQUIRED", "Event title cannot be empty.")
		}
		in.Title = &title
	}
	if req.Visibility != nil {
		v := event.Visibility(*req.Visibility)
		in.Visibility = &v
	}
	if req.Status != nil {
		st := event.Status(*req.Status)
		in.Status = &st
	}

	effStart := existing.StartAt.UTC()
	if req.StartAt != nil {
		t := req.StartAt.UTC()
		effStart = t
		in.StartAt = &t
	}
	effEnd := existing.EndAt.UTC()
	if req.EndAt != nil {
		t := req.EndAt.UTC()
		effEnd = t
		in.EndAt = &t
	}
	effDeadline := existing.RsvpDeadline
	if req.ClearRsvpDeadline != nil && *req.ClearRsvpDeadline {
		in.ClearRsvpDeadline = true
		effDeadline = nil
	} else if req.RsvpDeadline != nil {
		t := req.RsvpDeadline.UTC()
		in.RsvpDeadline = &t
		effDeadline = &t
	}

	if err := validateTimeRange(effStart, effEnd, effDeadline); err != nil {
		return domain.EventResponse{}, err
	}

	updated, err := s.repo.UpdateEvent(ctx, eventID, in)
	if err != nil {
		return domain.EventResponse{}, err
	}

	info, err := s.orgs.GetOrganization(ctx, p.ParishID, updated.OrgID)
	if err != nil {
		return domain.EventResponse{}, err
	}

	adminRole := "admin"
	return s.toEventResponse(updated, *info, &adminRole, time.Now().UTC()), nil
}

// DeleteEvent soft-deletes an Event in the active parish (requires org:admin of owning org).
func (s *Service) DeleteEvent(ctx context.Context, p authx.Principal, eventID uuid.UUID) error {
	existing, err := s.repo.FindEventInParish(ctx, p.ParishID, eventID)
	if ent.IsNotFound(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}

	isAdm, err := s.isOrgAdmin(ctx, p, existing.OrgID)
	if err != nil {
		return err
	}
	if !isAdm {
		return domain.ErrForbidden
	}

	return s.repo.SoftDeleteEvent(ctx, eventID, time.Now().UTC())
}

func (s *Service) resolveMyOrgRole(ctx context.Context, userID, orgID uuid.UUID, memberOrgIDs []uuid.UUID) (*string, error) {
	isMem := false
	for _, id := range memberOrgIDs {
		if id == orgID {
			isMem = true
			break
		}
	}
	if !isMem {
		return nil, nil
	}
	isAdm, err := s.orgs.IsAdmin(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	role := "member"
	if isAdm {
		role = "admin"
	}
	return &role, nil
}

func (s *Service) resolveMemberRole(ctx context.Context, userID, orgID uuid.UUID) (*string, error) {
	isAdm, err := s.orgs.IsAdmin(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if isAdm {
		role := "admin"
		return &role, nil
	}
	isMem, err := s.orgs.IsMember(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if isMem {
		role := "member"
		return &role, nil
	}
	return nil, nil
}

func (s *Service) toEventResponse(row *ent.Event, info OrgInfo, myRole *string, now time.Time) domain.EventResponse {
	cutoff := row.StartAt.UTC().Add(s.cfg.RSVPGracePeriod)
	if row.RsvpDeadline != nil {
		cutoff = row.RsvpDeadline.UTC()
	}
	rsvpOpen := row.Status != event.StatusCancelled && row.RsvpEnabled && !now.After(cutoff)

	var deadline *time.Time
	if row.RsvpDeadline != nil {
		t := row.RsvpDeadline.UTC()
		deadline = &t
	}

	return domain.EventResponse{
		ID:           row.ID,
		ParishID:     row.ParishID,
		OrgID:        row.OrgID,
		OrgCode:      info.Code,
		OrgName:      info.Name,
		OrgKind:      info.Kind,
		CreatedBy:    row.CreatedBy,
		Title:        row.Title,
		Description:  row.Description,
		Location:     row.Location,
		StartAt:      row.StartAt.UTC(),
		EndAt:        row.EndAt.UTC(),
		Visibility:   string(row.Visibility),
		RsvpEnabled:  row.RsvpEnabled,
		RsvpDeadline: deadline,
		Status:       string(row.Status),
		RsvpOpen:     rsvpOpen,
		MyOrgRole:    myRole,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func validateTimeRange(startAt, endAt time.Time, rsvpDeadline *time.Time) error {
	if !endAt.After(startAt) {
		return domain.ErrInvalidTimeRange
	}
	if rsvpDeadline != nil && rsvpDeadline.After(endAt) {
		return domain.ErrInvalidTimeRange
	}
	return nil
}

func trimOptString(val *string) *string {
	if val == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*val)
	return &trimmed
}
