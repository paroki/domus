package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/liturgiministry"
	"github.com/paroki/domus/internal/event"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/liturgi/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
)

const roleParokiAdmin = "ref:paroki-admin"

// OrgInfo holds minimal organization metadata needed by the liturgi service.
type OrgInfo struct {
	ID   uuid.UUID
	Name string
}

// OrgReader reads organization metadata and membership roles from the org feature.
type OrgReader interface {
	GetOrganization(ctx context.Context, parokiID, orgID uuid.UUID) (*OrgInfo, error)
	IsAdmin(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	ListUserOrgs(ctx context.Context, parokiID, userID uuid.UUID) ([]OrgInfo, error)
}

// Service implements business logic for liturgical ministries and parish Mass schedules.
type Service struct {
	repo     *repo.Repository
	enforcer *authz.Enforcer
	orgs     OrgReader
	events   *event.Feature
}

// New constructs a liturgi Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, orgs OrgReader, events *event.Feature) *Service {
	return &Service{
		repo:     r,
		enforcer: enforcer,
		orgs:     orgs,
		events:   events,
	}
}

func (s *Service) isLiturgiOrParokiAdmin(p authx.Principal) (bool, error) {
	if s.enforcer == nil {
		return false, nil
	}
	sub := authz.UserSubject(p.UserID)
	dom := authz.ParokiDomain(p.ParokiID)
	ok, err := s.enforcer.HasRoleInDomain(sub, domain.RoleAdmin, dom)
	if err != nil || ok {
		return ok, err
	}
	return s.enforcer.HasRoleInDomain(sub, roleParokiAdmin, dom)
}

func (s *Service) isOrgAdminCached(ctx context.Context, userID, orgID uuid.UUID, cache map[uuid.UUID]bool) (bool, error) {
	if s.orgs == nil {
		return false, nil
	}
	if v, ok := cache[orgID]; ok {
		return v, nil
	}
	v, err := s.orgs.IsAdmin(ctx, orgID, userID)
	if err != nil {
		return false, err
	}
	cache[orgID] = v
	return v, nil
}

// ListMinistries returns all 6 liturgical ministries in canonical order for the active paroki.
func (s *Service) ListMinistries(ctx context.Context, p authx.Principal) ([]domain.LiturgiMinistryResponse, error) {
	rows, err := s.repo.ListMinistries(ctx, p.ParokiID)
	if err != nil {
		return nil, err
	}

	bySlug := make(map[string]*ent.LiturgiMinistry, len(rows))
	orgIDs := make([]uuid.UUID, 0, len(rows))
	userIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		bySlug[string(row.RoleSlug)] = row
		if row.OrgID != nil {
			orgIDs = append(orgIDs, *row.OrgID)
		}
		if row.LeaderUserID != nil {
			userIDs = append(userIDs, *row.LeaderUserID)
		}
	}

	orgsMap, err := s.repo.LoadOrgs(ctx, orgIDs)
	if err != nil {
		return nil, err
	}
	usersMap, err := s.repo.LoadUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil {
		return nil, err
	}

	orgAdminCache := make(map[uuid.UUID]bool)
	out := make([]domain.LiturgiMinistryResponse, 0, len(domain.MinistryRoles))
	for _, slug := range domain.MinistryRoles {
		row := bySlug[slug]
		res := domain.LiturgiMinistryResponse{
			ParokiID:  p.ParokiID,
			RoleSlug:  slug,
			CanManage: isAdmin,
		}
		if row != nil {
			id := row.ID
			createdAt := row.CreatedAt
			updatedAt := row.UpdatedAt
			res.ID = &id
			res.OrgID = row.OrgID
			res.LeaderUserID = row.LeaderUserID
			res.CreatedAt = &createdAt
			res.UpdatedAt = &updatedAt
			if row.OrgID != nil {
				if o, ok := orgsMap[*row.OrgID]; ok {
					name := o.Name
					res.OrgName = &name
				}
				if !res.CanManage {
					adm, err := s.isOrgAdminCached(ctx, p.UserID, *row.OrgID, orgAdminCache)
					if err != nil {
						return nil, err
					}
					res.CanManage = adm
				}
			}
			if row.LeaderUserID != nil {
				if u, ok := usersMap[*row.LeaderUserID]; ok {
					name := u.Name
					email := u.Email
					res.LeaderName = &name
					res.LeaderEmail = &email
				}
				if !res.CanManage && *row.LeaderUserID == p.UserID {
					res.CanManage = true
				}
			}
		}
		out = append(out, res)
	}
	return out, nil
}

// UpdateMinistry configures the organization and/or leader user for one of the 6 liturgical ministry roles.
func (s *Service) UpdateMinistry(
	ctx context.Context,
	p authx.Principal,
	roleSlug string,
	req domain.UpdateLiturgiMinistryRequest,
) (domain.LiturgiMinistryResponse, error) {
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil {
		return domain.LiturgiMinistryResponse{}, err
	}
	if !isAdmin {
		return domain.LiturgiMinistryResponse{}, domain.ErrForbidden
	}
	if !domain.IsValidMinistryRole(roleSlug) {
		return domain.LiturgiMinistryResponse{}, domain.ValidationError("roleSlug", "INVALID_ROLE_SLUG", "Invalid ministry role slug.")
	}

	clearOrg := req.ClearOrgID != nil && *req.ClearOrgID
	clearLeader := req.ClearLeaderUserID != nil && *req.ClearLeaderUserID

	var orgID *uuid.UUID
	if !clearOrg && req.OrgID != nil && *req.OrgID != uuid.Nil {
		if s.orgs == nil {
			return domain.LiturgiMinistryResponse{}, domain.ValidationError("orgId", "NOT_FOUND", "Organization not found in active parish.")
		}
		if _, err := s.orgs.GetOrganization(ctx, p.ParokiID, *req.OrgID); err != nil {
			return domain.LiturgiMinistryResponse{}, domain.ValidationError("orgId", "NOT_FOUND", "Organization not found in active parish.")
		}
		id := *req.OrgID
		orgID = &id
	}

	var leaderUserID *uuid.UUID
	if !clearLeader && req.LeaderUserID != nil && *req.LeaderUserID != uuid.Nil {
		ok, err := s.repo.IsActiveUserInParoki(ctx, p.ParokiID, *req.LeaderUserID)
		if err != nil {
			return domain.LiturgiMinistryResponse{}, err
		}
		if !ok {
			return domain.LiturgiMinistryResponse{}, domain.ValidationError("leaderUserId", "NOT_FOUND", "Leader user must be an active member of the active parish.")
		}
		id := *req.LeaderUserID
		leaderUserID = &id
	}

	if _, err := s.repo.UpsertMinistry(
		ctx,
		p.ParokiID,
		liturgiministry.RoleSlug(roleSlug),
		orgID,
		clearOrg,
		leaderUserID,
		clearLeader,
	); err != nil {
		return domain.LiturgiMinistryResponse{}, err
	}

	all, err := s.ListMinistries(ctx, p)
	if err != nil {
		return domain.LiturgiMinistryResponse{}, err
	}
	for _, item := range all {
		if item.RoleSlug == roleSlug {
			return item, nil
		}
	}
	return domain.LiturgiMinistryResponse{}, nil
}

// SearchSchedules returns paginated parish-level Mass schedules enriched with ministry permissions, petugas, and lagu.
func (s *Service) SearchSchedules(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchLiturgiSchedulesRequest,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.LiturgiScheduleResponse, bool, error) {
	parokiWide := true
	evFilter := &event.EventSearchFilter{
		ParokiWideOnly: &parokiWide,
		Jenis:          "misa",
	}

	if req.Filter != nil {
		evFilter.From = req.Filter.From
		evFilter.To = req.Filter.To
		evFilter.Status = req.Filter.Status

		hasPetugasFilter := req.Filter.Role != "" ||
			(req.Filter.AssignedToMe != nil && *req.Filter.AssignedToMe) ||
			req.Filter.AssignedOrgID != nil ||
			(req.Filter.HasDeclined != nil && *req.Filter.HasDeclined)

		if hasPetugasFilter {
			var memberOrgIDs []uuid.UUID
			if req.Filter.AssignedToMe != nil && *req.Filter.AssignedToMe && s.orgs != nil {
				myOrgs, err := s.orgs.ListUserOrgs(ctx, p.ParokiID, p.UserID)
				if err != nil {
					return nil, false, err
				}
				for _, o := range myOrgs {
					memberOrgIDs = append(memberOrgIDs, o.ID)
				}
			}
			ids, err := s.repo.ListMatchingEventIDsForPetugas(ctx, repo.PetugasEventFilter{
				ParokiID:       p.ParokiID,
				Role:           req.Filter.Role,
				AssignedToMe:   req.Filter.AssignedToMe != nil && *req.Filter.AssignedToMe,
				AssignedUserID: p.UserID,
				MemberOrgIDs:   memberOrgIDs,
				AssignedOrgID:  req.Filter.AssignedOrgID,
				HasDeclined:    req.Filter.HasDeclined,
			})
			if err != nil {
				return nil, false, err
			}
			evFilter.EventIDs = ids
		}
	}

	events, hasMore, err := s.events.SearchEvents(ctx, p, event.SearchEventsRequest{
		Q:      req.Q,
		Filter: evFilter,
	}, cursorAt, cursorID, limit)
	if err != nil {
		return nil, false, mapEventError(err)
	}

	out, err := s.enrichSchedules(ctx, p, events)
	if err != nil {
		return nil, false, err
	}
	return out, hasMore, nil
}

// GetSchedule returns a single parish-level Mass schedule by eventId.
func (s *Service) GetSchedule(ctx context.Context, p authx.Principal, eventID uuid.UUID) (domain.LiturgiScheduleResponse, error) {
	ev, err := s.events.GetEvent(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiScheduleResponse{}, mapEventError(err)
	}
	if !isParishMassEvent(ev) {
		return domain.LiturgiScheduleResponse{}, domain.ErrScheduleNotFound
	}
	items, err := s.enrichSchedules(ctx, p, []event.Event{ev})
	if err != nil {
		return domain.LiturgiScheduleResponse{}, err
	}
	return items[0], nil
}

// CreateSchedule creates a new parish-level Mass schedule in events (org_id = NULL, jenis = "misa", visibility = "public").
func (s *Service) CreateSchedule(ctx context.Context, p authx.Principal, req domain.CreateLiturgiScheduleRequest) (domain.LiturgiScheduleResponse, error) {
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil {
		return domain.LiturgiScheduleResponse{}, err
	}
	if !isAdmin {
		return domain.LiturgiScheduleResponse{}, domain.ErrForbidden
	}

	misa := "misa"
	ev, err := s.events.CreateEvent(ctx, p, event.CreateEventRequest{
		Title:        req.Title,
		Jenis:        &misa,
		Description:  req.Description,
		Location:     req.Location,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		Visibility:   "public",
		RsvpEnabled:  req.RsvpEnabled,
		RsvpDeadline: req.RsvpDeadline,
	})
	if err != nil {
		return domain.LiturgiScheduleResponse{}, mapEventError(err)
	}
	items, err := s.enrichSchedules(ctx, p, []event.Event{ev})
	if err != nil {
		return domain.LiturgiScheduleResponse{}, err
	}
	return items[0], nil
}

// UpdateSchedule updates an existing parish-level Mass schedule.
func (s *Service) UpdateSchedule(ctx context.Context, p authx.Principal, eventID uuid.UUID, req domain.UpdateLiturgiScheduleRequest) (domain.LiturgiScheduleResponse, error) {
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil {
		return domain.LiturgiScheduleResponse{}, err
	}
	if !isAdmin {
		return domain.LiturgiScheduleResponse{}, domain.ErrForbidden
	}

	existing, err := s.events.GetEvent(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiScheduleResponse{}, mapEventError(err)
	}
	if !isParishMassEvent(existing) {
		return domain.LiturgiScheduleResponse{}, domain.ErrScheduleNotFound
	}

	ev, err := s.events.UpdateEvent(ctx, p, eventID, event.UpdateEventRequest{
		Title:             req.Title,
		Description:       req.Description,
		Location:          req.Location,
		StartAt:           req.StartAt,
		EndAt:             req.EndAt,
		RsvpEnabled:       req.RsvpEnabled,
		RsvpDeadline:      req.RsvpDeadline,
		ClearRsvpDeadline: req.ClearRsvpDeadline,
		Status:            req.Status,
	})
	if err != nil {
		return domain.LiturgiScheduleResponse{}, mapEventError(err)
	}
	items, err := s.enrichSchedules(ctx, p, []event.Event{ev})
	if err != nil {
		return domain.LiturgiScheduleResponse{}, err
	}
	return items[0], nil
}

// DeleteSchedule soft-deletes a parish-level Mass schedule.
func (s *Service) DeleteSchedule(ctx context.Context, p authx.Principal, eventID uuid.UUID) error {
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil {
		return err
	}
	if !isAdmin {
		return domain.ErrForbidden
	}

	existing, err := s.events.GetEvent(ctx, p, eventID)
	if err != nil {
		return mapEventError(err)
	}
	if !isParishMassEvent(existing) {
		return domain.ErrScheduleNotFound
	}
	return mapEventError(s.events.DeleteEvent(ctx, p, eventID))
}

func (s *Service) enrichSchedules(ctx context.Context, p authx.Principal, events []event.Event) ([]domain.LiturgiScheduleResponse, error) {
	if len(events) == 0 {
		return []domain.LiturgiScheduleResponse{}, nil
	}

	ministries, err := s.ListMinistries(ctx, p)
	if err != nil {
		return nil, err
	}
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil {
		return nil, err
	}

	canManageRole := make(map[string]bool, len(ministries))
	manageableRoles := make([]string, 0, len(ministries))
	for _, m := range ministries {
		canManageRole[m.RoleSlug] = m.CanManage
		if m.CanManage {
			manageableRoles = append(manageableRoles, m.RoleSlug)
		}
	}
	canManageLagu := isAdmin || canManageRole["organis"] || canManageRole["kor"]

	eventIDs := make([]uuid.UUID, 0, len(events))
	for _, ev := range events {
		eventIDs = append(eventIDs, ev.ID)
	}

	petugasRows, err := s.repo.ListPetugasForEvents(ctx, eventIDs)
	if err != nil {
		return nil, err
	}
	laguRows, err := s.repo.ListLaguForEvents(ctx, eventIDs)
	if err != nil {
		return nil, err
	}

	orgAdminCache := make(map[uuid.UUID]bool)
	petugasByEvent := make(map[uuid.UUID][]domain.LiturgiPetugasResponse, len(eventIDs))
	for _, row := range petugasRows {
		canConfirm := canManageRole[string(row.Role)]
		if !canConfirm && row.UserID != nil && *row.UserID == p.UserID {
			canConfirm = true
		}
		if !canConfirm && row.SubstituteUserID != nil && *row.SubstituteUserID == p.UserID {
			canConfirm = true
		}
		if !canConfirm && row.OrgID != nil {
			adm, err := s.isOrgAdminCached(ctx, p.UserID, *row.OrgID, orgAdminCache)
			if err != nil {
				return nil, err
			}
			canConfirm = adm
		}
		if !canConfirm && row.SubstituteOrgID != nil {
			adm, err := s.isOrgAdminCached(ctx, p.UserID, *row.SubstituteOrgID, orgAdminCache)
			if err != nil {
				return nil, err
			}
			canConfirm = adm
		}
		petugasByEvent[row.EventID] = append(petugasByEvent[row.EventID], toPetugasResponse(row, canConfirm))
	}

	laguByEvent := make(map[uuid.UUID][]domain.LiturgiLaguResponse, len(eventIDs))
	for _, row := range laguRows {
		laguByEvent[row.EventID] = append(laguByEvent[row.EventID], toLaguResponse(row))
	}

	out := make([]domain.LiturgiScheduleResponse, 0, len(events))
	for _, ev := range events {
		pList := petugasByEvent[ev.ID]
		if pList == nil {
			pList = []domain.LiturgiPetugasResponse{}
		}
		lList := laguByEvent[ev.ID]
		if lList == nil {
			lList = []domain.LiturgiLaguResponse{}
		}
		var myRsvp *domain.LiturgiRSVPResponse
		if ev.MyRsvp != nil {
			r := domain.LiturgiRSVPResponse{
				ID:        ev.MyRsvp.ID,
				ParokiID:  ev.MyRsvp.ParokiID,
				EventID:   ev.MyRsvp.EventID,
				UserID:    ev.MyRsvp.UserID,
				Name:      ev.MyRsvp.Name,
				Email:     ev.MyRsvp.Email,
				Phone:     ev.MyRsvp.Phone,
				OrgRole:   ev.MyRsvp.OrgRole,
				Status:    ev.MyRsvp.Status,
				Note:      ev.MyRsvp.Note,
				CreatedAt: ev.MyRsvp.CreatedAt,
				UpdatedAt: ev.MyRsvp.UpdatedAt,
			}
			myRsvp = &r
		}
		var rsvpSummary *domain.LiturgiRSVPSummaryResponse
		if ev.RsvpSummary != nil {
			sum := domain.LiturgiRSVPSummaryResponse{
				EventID:            ev.RsvpSummary.EventID,
				Going:              ev.RsvpSummary.Going,
				Maybe:              ev.RsvpSummary.Maybe,
				NotGoing:           ev.RsvpSummary.NotGoing,
				TotalResponded:     ev.RsvpSummary.TotalResponded,
				TotalActiveMembers: ev.RsvpSummary.TotalActiveMembers,
				Unresponded:        ev.RsvpSummary.Unresponded,
			}
			rsvpSummary = &sum
		}
		out = append(out, domain.LiturgiScheduleResponse{
			ID:                ev.ID,
			ParokiID:          ev.ParokiID,
			CreatedBy:         ev.CreatedBy,
			Title:             ev.Title,
			Description:       ev.Description,
			Location:          ev.Location,
			StartAt:           ev.StartAt,
			EndAt:             ev.EndAt,
			Visibility:        ev.Visibility,
			RsvpEnabled:       ev.RsvpEnabled,
			RsvpDeadline:      ev.RsvpDeadline,
			Status:            ev.Status,
			RsvpOpen:          ev.RsvpOpen,
			CanRsvp:           ev.CanRsvp,
			MyRsvp:            myRsvp,
			RsvpSummary:       rsvpSummary,
			CanManageSchedule: isAdmin,
			ManageableRoles:   manageableRoles,
			CanManageLagu:     canManageLagu,
			Petugas:           pList,
			Lagu:              lList,
			CreatedAt:         ev.CreatedAt,
			UpdatedAt:         ev.UpdatedAt,
		})
	}
	return out, nil
}

func isParishMassEvent(ev event.Event) bool {
	return ev.OrgID == nil && ev.Jenis != nil && *ev.Jenis == "misa"
}

func mapEventError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, event.ErrEventNotFound):
		return domain.ErrScheduleNotFound
	case errors.Is(err, event.ErrInvalidTimeRange):
		return domain.ErrInvalidTimeRange
	case errors.Is(err, event.ErrForbidden):
		return domain.ErrForbidden
	default:
		return err
	}
}

func toPetugasResponse(row *ent.LiturgiPetugas, canConfirm bool) domain.LiturgiPetugasResponse {
	return domain.LiturgiPetugasResponse{
		ID:               row.ID,
		ParokiID:         row.ParokiID,
		EventID:          row.EventID,
		Role:             string(row.Role),
		SlotLabel:        row.SlotLabel,
		UserID:           row.UserID,
		OrgID:            row.OrgID,
		Name:             row.Name,
		Urutan:           row.Urutan,
		Status:           string(row.Status),
		Note:             row.Note,
		RespondedAt:      row.RespondedAt,
		RespondedBy:      row.RespondedBy,
		SubstituteUserID: row.SubstituteUserID,
		SubstituteOrgID:  row.SubstituteOrgID,
		SubstituteName:   row.SubstituteName,
		CanConfirm:       canConfirm,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func toLaguResponse(row *ent.LiturgiLagu) domain.LiturgiLaguResponse {
	return domain.LiturgiLaguResponse{
		ID:        row.ID,
		ParokiID:  row.ParokiID,
		EventID:   row.EventID,
		Urutan:    row.Urutan,
		Bagian:    row.Bagian,
		CantusID:  row.CantusID,
		Sumber:    row.Sumber,
		Nomor:     row.Nomor,
		Judul:     row.Judul,
		NadaDasar: row.NadaDasar,
		Note:      row.Note,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
