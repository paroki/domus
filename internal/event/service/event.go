package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

var (
	eventTypeCodeRe   = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)
	defaultEventTypes = map[string]struct{}{
		"misa":         {},
		"ibadat":       {},
		"rapat":        {},
		"latihan":      {},
		"rekoleksi":    {},
		"katekese":     {},
		"bakti-sosial": {},
		"lainnya":      {},
	}
)

// SearchEvents returns paginated visible events in the caller's active paroki.
func (s *Service) SearchEvents(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchEventsRequest,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.EventResponse, bool, error) {
	myOrgs, err := s.orgs.ListUserOrgs(ctx, p.ParokiID, p.UserID)
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
		ParokiID:     p.ParokiID,
		MemberOrgIDs: memberOrgIDs,
		Q:            strings.TrimSpace(req.Q),
		CursorAt:     cursorAt,
		CursorID:     cursorID,
		Limit:        limit + 1,
	}
	if req.Filter != nil {
		in.OrgID = req.Filter.OrgID
		in.Jenis = strings.ToLower(strings.TrimSpace(req.Filter.Jenis))
		in.Visibility = req.Filter.Visibility
		in.Status = req.Filter.Status
		in.From = req.Filter.From
		in.To = req.Filter.To
		if req.Filter.ParokiWideOnly != nil {
			in.ParokiWideOnly = *req.Filter.ParokiWideOnly
		}
		if req.Filter.MyOrgsOnly != nil {
			in.MyOrgsOnly = *req.Filter.MyOrgsOnly
		}
		if req.Filter.AssignedToMe != nil && *req.Filter.AssignedToMe {
			in.AssignedToMe = true
			in.AssignedUserID = p.UserID
		}
		in.AssignedOrgID = req.Filter.AssignedOrgID
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

	isParokiAdm, err := s.isParokiAdmin(ctx, p)
	if err != nil {
		return nil, false, err
	}
	isParokiMem := isParokiAdm
	if !isParokiMem {
		isParokiMem, err = s.repo.IsParokiMember(ctx, p.ParokiID, p.UserID)
		if err != nil {
			return nil, false, err
		}
	}

	now := time.Now().UTC()
	out := make([]domain.EventResponse, 0, len(rows))
	roleCache := make(map[uuid.UUID]*string)
	rsvpEventOrgMap := make(map[uuid.UUID]*uuid.UUID)
	rsvpEventIDs := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		var info OrgInfo
		var myRole *string
		var canManage, canRsvp bool

		if row.OrgID != nil {
			orgID := *row.OrgID
			cachedInfo, ok := orgMap[orgID]
			if !ok {
				fetched, err := s.orgs.GetOrganization(ctx, p.ParokiID, orgID)
				if err != nil {
					return nil, false, err
				}
				cachedInfo = *fetched
				orgMap[orgID] = cachedInfo
			}
			info = cachedInfo

			var cached bool
			myRole, cached = roleCache[orgID]
			if !cached {
				myRole, err = s.resolveMyOrgRole(ctx, p.UserID, orgID, memberOrgIDs)
				if err != nil {
					return nil, false, err
				}
				roleCache[orgID] = myRole
			}
			canManage = myRole != nil && *myRole == "admin"
			canRsvp = myRole != nil
		} else {
			canManage = isParokiAdm
			canRsvp = isParokiMem || isParokiAdm
		}

		if canRsvp {
			rsvpEventOrgMap[row.ID] = row.OrgID
			rsvpEventIDs = append(rsvpEventIDs, row.ID)
		}

		out = append(out, s.toEventResponse(row, info, myRole, canManage, canRsvp, now))
	}

	if len(rsvpEventIDs) > 0 {
		myRSVPs, err := s.repo.FindUserRSVPsForEvents(ctx, rsvpEventIDs, p.UserID)
		if err != nil {
			return nil, false, err
		}
		summaries, err := s.repo.GetRSVPSummariesForEvents(ctx, p.ParokiID, rsvpEventOrgMap)
		if err != nil {
			return nil, false, err
		}
		usersMap, err := s.repo.LoadUsers(ctx, []uuid.UUID{p.UserID})
		if err != nil {
			return nil, false, err
		}
		callerUser := usersMap[p.UserID]

		orgIDsForCaller := make(map[uuid.UUID]struct{})
		for _, orgID := range rsvpEventOrgMap {
			if orgID != nil {
				orgIDsForCaller[*orgID] = struct{}{}
			}
		}
		callerMembers := make(map[uuid.UUID]*ent.OrgMember)
		for orgID := range orgIDsForCaller {
			_, mMap, err := s.repo.LoadUsersAndOrgMembers(ctx, orgID, []uuid.UUID{p.UserID})
			if err != nil {
				return nil, false, err
			}
			callerMembers[orgID] = mMap[p.UserID]
		}

		for i := range out {
			if !out[i].CanRsvp {
				continue
			}
			if rsvpRow, ok := myRSVPs[out[i].ID]; ok {
				var member *ent.OrgMember
				if out[i].OrgID != nil {
					member = callerMembers[*out[i].OrgID]
				}
				r := toRSVPResponse(rsvpRow, callerUser, member)
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

// GetEvent returns a single non-deleted Event in the caller's active paroki, enforcing private visibility.
func (s *Service) GetEvent(ctx context.Context, p authx.Principal, eventID uuid.UUID) (domain.EventResponse, error) {
	row, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventResponse{}, err
	}

	var info OrgInfo
	var myRole *string
	var canManage, canRsvp bool

	if row.OrgID != nil {
		orgID := *row.OrgID
		myRole, err = s.resolveMemberRole(ctx, p.UserID, orgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		if row.Visibility == event.VisibilityPrivate && myRole == nil {
			return domain.EventResponse{}, domain.ErrForbidden
		}
		fetched, err := s.orgs.GetOrganization(ctx, p.ParokiID, orgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		info = *fetched
		canManage = myRole != nil && *myRole == "admin"
		canRsvp = myRole != nil
	} else {
		isParokiAdm, err := s.isParokiAdmin(ctx, p)
		if err != nil {
			return domain.EventResponse{}, err
		}
		isParokiMem := isParokiAdm
		if !isParokiMem {
			isParokiMem, err = s.repo.IsParokiMember(ctx, p.ParokiID, p.UserID)
			if err != nil {
				return domain.EventResponse{}, err
			}
		}
		canManage = isParokiAdm
		canRsvp = isParokiMem || isParokiAdm
	}

	res := s.toEventResponse(row, info, myRole, canManage, canRsvp, time.Now().UTC())
	if canRsvp {
		counts, err := s.repo.GetRSVPSummary(ctx, p.ParokiID, row.ID, row.OrgID)
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
			var u *ent.User
			var m *ent.OrgMember
			if row.OrgID != nil {
				uMap, mMap, err := s.repo.LoadUsersAndOrgMembers(ctx, *row.OrgID, []uuid.UUID{p.UserID})
				if err != nil {
					return domain.EventResponse{}, err
				}
				u = uMap[p.UserID]
				m = mMap[p.UserID]
			} else {
				uMap, err := s.repo.LoadUsers(ctx, []uuid.UUID{p.UserID})
				if err != nil {
					return domain.EventResponse{}, err
				}
				u = uMap[p.UserID]
			}
			r := toRSVPResponse(rsvpRow, u, m)
			res.MyRsvp = &r
		}
	}

	petugasRows, err := s.repo.ListPetugasForEvent(ctx, row.ID)
	if err != nil {
		return domain.EventResponse{}, err
	}
	if len(petugasRows) > 0 {
		orgAdminCache := make(map[uuid.UUID]bool)
		petugas := make([]domain.EventPetugasResponse, 0, len(petugasRows))
		for _, pr := range petugasRows {
			canConfirm, err := s.canConfirmPetugas(ctx, p, canManage, pr, orgAdminCache)
			if err != nil {
				return domain.EventResponse{}, err
			}
			petugas = append(petugas, toPetugasResponse(pr, canConfirm))
		}
		res.Petugas = petugas
	}
	return res, nil
}

// CreateEvent creates a new Event owned by req.OrgID (org:admin) or by the active paroki (ref:paroki-admin).
func (s *Service) CreateEvent(ctx context.Context, p authx.Principal, req domain.CreateEventRequest) (domain.EventResponse, error) {
	orgID := req.OrgID
	if orgID != nil && *orgID == uuid.Nil {
		orgID = nil
	}

	var info OrgInfo
	var myRole *string
	var canRsvp bool

	if orgID != nil {
		fetched, err := s.orgs.GetOrganization(ctx, p.ParokiID, *orgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		info = *fetched

		isAdm, err := s.isOrgAdmin(ctx, p, *orgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		if !isAdm {
			return domain.EventResponse{}, domain.ErrForbidden
		}
		adminRole := "admin"
		myRole = &adminRole
		canRsvp = true
	} else {
		isAdm, err := s.isParokiAdmin(ctx, p)
		if err != nil {
			return domain.EventResponse{}, err
		}
		if !isAdm {
			return domain.EventResponse{}, domain.ErrForbidden
		}
		canRsvp = true
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
	if orgID == nil {
		if req.Visibility == string(event.VisibilityPrivate) || vis != event.VisibilityPublic {
			return domain.EventResponse{}, domain.ValidationError("visibility", "PAROKI_EVENT_PUBLIC_ONLY", "Parish-level events must have public visibility.")
		}
	}

	rsvpEnabled := true
	if req.RsvpEnabled != nil {
		rsvpEnabled = *req.RsvpEnabled
	}

	jenis, _, err := s.resolveEventJenis(ctx, p.ParokiID, req.Jenis)
	if err != nil {
		return domain.EventResponse{}, err
	}

	created, err := s.repo.CreateEvent(ctx, repo.CreateEventInput{
		ParokiID:     p.ParokiID,
		OrgID:        orgID,
		CreatedBy:    p.UserID,
		Title:        title,
		Jenis:        jenis,
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

	return s.toEventResponse(created, info, myRole, true, canRsvp, time.Now().UTC()), nil
}

// UpdateEvent updates an existing Event in the active paroki.
func (s *Service) UpdateEvent(ctx context.Context, p authx.Principal, eventID uuid.UUID, req domain.UpdateEventRequest) (domain.EventResponse, error) {
	existing, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventResponse{}, err
	}

	var myRole *string
	if existing.OrgID != nil {
		isAdm, err := s.isOrgAdmin(ctx, p, *existing.OrgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		if !isAdm {
			return domain.EventResponse{}, domain.ErrForbidden
		}
		adminRole := "admin"
		myRole = &adminRole
	} else {
		isAdm, err := s.isParokiAdmin(ctx, p)
		if err != nil {
			return domain.EventResponse{}, err
		}
		if !isAdm {
			return domain.EventResponse{}, domain.ErrForbidden
		}
		if req.Visibility != nil && *req.Visibility != string(event.VisibilityPublic) {
			return domain.EventResponse{}, domain.ValidationError("visibility", "PAROKI_EVENT_PUBLIC_ONLY", "Parish-level events must have public visibility.")
		}
	}

	in := repo.UpdateEventInput{
		Description: trimOptString(req.Description),
		Location:    trimOptString(req.Location),
		RsvpEnabled: req.RsvpEnabled,
	}
	if req.Jenis != nil {
		jenis, clearJenis, err := s.resolveEventJenis(ctx, p.ParokiID, req.Jenis)
		if err != nil {
			return domain.EventResponse{}, err
		}
		in.Jenis = jenis
		in.ClearJenis = clearJenis
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

	var info OrgInfo
	if updated.OrgID != nil {
		fetched, err := s.orgs.GetOrganization(ctx, p.ParokiID, *updated.OrgID)
		if err != nil {
			return domain.EventResponse{}, err
		}
		info = *fetched
	}

	return s.toEventResponse(updated, info, myRole, true, true, time.Now().UTC()), nil
}

// DeleteEvent soft-deletes an Event in the active paroki.
func (s *Service) DeleteEvent(ctx context.Context, p authx.Principal, eventID uuid.UUID) error {
	existing, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}

	if existing.OrgID != nil {
		isAdm, err := s.isOrgAdmin(ctx, p, *existing.OrgID)
		if err != nil {
			return err
		}
		if !isAdm {
			return domain.ErrForbidden
		}
	} else {
		isAdm, err := s.isParokiAdmin(ctx, p)
		if err != nil {
			return err
		}
		if !isAdm {
			return domain.ErrForbidden
		}
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

func (s *Service) toEventResponse(row *ent.Event, info OrgInfo, myRole *string, canManage, canRsvp bool, now time.Time) domain.EventResponse {
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
		ParokiID:     row.ParokiID,
		OrgID:        row.OrgID,
		OrgCode:      info.Code,
		OrgName:      info.Name,
		OrgKind:      info.Kind,
		CreatedBy:    row.CreatedBy,
		Title:        row.Title,
		Jenis:        row.Jenis,
		Description:  row.Description,
		Location:     row.Location,
		StartAt:      row.StartAt.UTC(),
		EndAt:        row.EndAt.UTC(),
		Visibility:   string(row.Visibility),
		RsvpEnabled:  row.RsvpEnabled,
		RsvpDeadline: deadline,
		Status:       string(row.Status),
		RsvpOpen:     rsvpOpen,
		CanManage:    canManage,
		CanRsvp:      canRsvp,
		MyOrgRole:    myRole,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func (s *Service) resolveEventJenis(ctx context.Context, parokiID uuid.UUID, raw *string) (*string, bool, error) {
	if raw == nil {
		return nil, false, nil
	}
	code := strings.ToLower(strings.TrimSpace(*raw))
	if code == "" {
		return nil, true, nil
	}
	if !eventTypeCodeRe.MatchString(code) {
		return nil, false, domain.ValidationError("jenis", "INVALID_EVENT_TYPE", "Invalid event type.")
	}
	if _, ok := defaultEventTypes[code]; ok {
		return &code, false, nil
	}
	if s.parokis != nil {
		active, err := s.parokis.HasActiveLookup(ctx, parokiID, "event_type", code)
		if err != nil {
			return nil, false, err
		}
		if active {
			return &code, false, nil
		}
	}
	return nil, false, domain.ValidationError("jenis", "INVALID_EVENT_TYPE", "Invalid event type.")
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
