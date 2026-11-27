package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/liturgiministry"
	"github.com/paroki/domus/ent/liturgipetugas"
	"github.com/paroki/domus/internal/event"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/liturgi/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

func (s *Service) canManageRole(ctx context.Context, p authx.Principal, roleSlug string) (bool, error) {
	isAdmin, err := s.isLiturgiOrParokiAdmin(p)
	if err != nil || isAdmin {
		return isAdmin, err
	}
	ministry, err := s.repo.FindMinistry(ctx, p.ParokiID, liturgiministry.RoleSlug(roleSlug))
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if ministry.LeaderUserID != nil && *ministry.LeaderUserID == p.UserID {
		return true, nil
	}
	if ministry.OrgID != nil && s.orgs != nil {
		return s.orgs.IsAdmin(ctx, *ministry.OrgID, p.UserID)
	}
	return false, nil
}

func (s *Service) ensureMassScheduleExists(ctx context.Context, p authx.Principal, eventID uuid.UUID) (event.Event, error) {
	ev, err := s.events.GetEvent(ctx, p, eventID)
	if err != nil {
		return event.Event{}, mapEventError(err)
	}
	if !isParishMassEvent(ev) {
		return event.Event{}, domain.ErrScheduleNotFound
	}
	return ev, nil
}

func (s *Service) resolveTargetAssignee(
	ctx context.Context,
	parokiID uuid.UUID,
	userID, orgID *uuid.UUID,
	rawName string,
	userField, orgField, nameField string,
) (*uuid.UUID, *uuid.UUID, string, error) {
	hasUser := userID != nil && *userID != uuid.Nil
	hasOrg := orgID != nil && *orgID != uuid.Nil
	if hasUser && hasOrg {
		return nil, nil, "", domain.ValidationError(orgField, "MUTUALLY_EXCLUSIVE", "User and organization cannot both be provided.")
	}

	name := strings.TrimSpace(rawName)
	if hasUser {
		uName, ok, err := s.repo.GetActiveUserNameInParoki(ctx, parokiID, *userID)
		if err != nil {
			return nil, nil, "", err
		}
		if !ok {
			return nil, nil, "", domain.ValidationError(userField, "NOT_FOUND", "User not found in active parish.")
		}
		if name == "" {
			name = uName
		}
		uid := *userID
		return &uid, nil, name, nil
	}

	if hasOrg {
		if s.orgs == nil {
			return nil, nil, "", domain.ValidationError(orgField, "NOT_FOUND", "Organization not found in active parish.")
		}
		o, err := s.orgs.GetOrganization(ctx, parokiID, *orgID)
		if err != nil {
			return nil, nil, "", domain.ValidationError(orgField, "NOT_FOUND", "Organization not found in active parish.")
		}
		if name == "" {
			name = o.Name
		}
		oid := *orgID
		return nil, &oid, name, nil
	}

	if name == "" {
		return nil, nil, "", domain.ValidationError(nameField, "REQUIRED", "Name is required when neither user nor organization is provided.")
	}
	return nil, nil, name, nil
}

// CreatePetugas assigns an officer/minister to a parish Mass schedule.
func (s *Service) CreatePetugas(
	ctx context.Context,
	p authx.Principal,
	eventID uuid.UUID,
	req domain.CreateLiturgiPetugasRequest,
) (domain.LiturgiPetugasResponse, error) {
	if !domain.IsValidMinistryRole(req.Role) {
		return domain.LiturgiPetugasResponse{}, domain.ValidationError("role", "INVALID_ROLE_SLUG", "Invalid ministry role slug.")
	}
	allowed, err := s.canManageRole(ctx, p, req.Role)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	if !allowed {
		return domain.LiturgiPetugasResponse{}, domain.ErrForbidden
	}
	ev, err := s.ensureMassScheduleExists(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}

	uid, oid, name, err := s.resolveTargetAssignee(ctx, p.ParokiID, req.UserID, req.OrgID, req.Name, "userId", "orgId", "name")
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}

	status := liturgipetugas.StatusAssigned
	var respondedAt *time.Time
	var respondedBy *uuid.UUID
	if req.Status != nil && *req.Status != "" {
		switch *req.Status {
		case "assigned":
			status = liturgipetugas.StatusAssigned
		case "confirmed", "declined":
			status = liturgipetugas.Status(*req.Status)
			now := time.Now().UTC()
			respondedAt = &now
			by := p.UserID
			respondedBy = &by
		default:
			return domain.LiturgiPetugasResponse{}, domain.ValidationError("status", "INVALID_STATUS", "Status must be assigned, confirmed, or declined.")
		}
	}

	urutan := 0
	if req.Urutan != nil {
		urutan = *req.Urutan
	}

	created, err := s.repo.CreatePetugas(ctx, repo.CreatePetugasInput{
		ParokiID:    p.ParokiID,
		EventID:     eventID,
		Role:        liturgipetugas.Role(req.Role),
		SlotLabel:   trimOptionalPtr(req.SlotLabel),
		UserID:      uid,
		OrgID:       oid,
		Name:        name,
		Urutan:      urutan,
		Status:      status,
		Note:        trimOptionalPtr(req.Note),
		RespondedAt: respondedAt,
		RespondedBy: respondedBy,
	})
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}

	s.notifyPetugasAssigned(ctx, p.UserID, ev, created)
	return toPetugasResponse(created, true, true, true, false), nil
}

// UpdatePetugas updates an existing officer assignment on a parish Mass schedule.
func (s *Service) UpdatePetugas(
	ctx context.Context,
	p authx.Principal,
	eventID, petugasID uuid.UUID,
	req domain.UpdateLiturgiPetugasRequest,
) (domain.LiturgiPetugasResponse, error) {
	ev, err := s.ensureMassScheduleExists(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	existing, err := s.repo.FindPetugasByID(ctx, p.ParokiID, eventID, petugasID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.LiturgiPetugasResponse{}, domain.ErrPetugasNotFound
		}
		return domain.LiturgiPetugasResponse{}, err
	}

	allowed, err := s.canManageRole(ctx, p, string(existing.Role))
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	if !allowed {
		return domain.LiturgiPetugasResponse{}, domain.ErrForbidden
	}

	var upd repo.UpdatePetugasInput
	if req.Role != nil {
		if !domain.IsValidMinistryRole(*req.Role) {
			return domain.LiturgiPetugasResponse{}, domain.ValidationError("role", "INVALID_ROLE_SLUG", "Invalid ministry role slug.")
		}
		if *req.Role != string(existing.Role) {
			canTarget, err := s.canManageRole(ctx, p, *req.Role)
			if err != nil {
				return domain.LiturgiPetugasResponse{}, err
			}
			if !canTarget {
				return domain.LiturgiPetugasResponse{}, domain.ErrForbidden
			}
		}
		r := liturgipetugas.Role(*req.Role)
		upd.Role = &r
	}

	if req.SlotLabel != nil {
		if strings.TrimSpace(*req.SlotLabel) == "" {
			upd.ClearSlotLabel = true
		} else {
			v := strings.TrimSpace(*req.SlotLabel)
			upd.SlotLabel = &v
		}
	}
	if req.Note != nil {
		if strings.TrimSpace(*req.Note) == "" {
			upd.ClearNote = true
		} else {
			v := strings.TrimSpace(*req.Note)
			upd.Note = &v
		}
	}
	if req.Urutan != nil {
		upd.Urutan = req.Urutan
	}

	hasReqUser := req.UserID != nil && *req.UserID != uuid.Nil
	hasReqOrg := req.OrgID != nil && *req.OrgID != uuid.Nil
	if hasReqUser && hasReqOrg {
		return domain.LiturgiPetugasResponse{}, domain.ValidationError("orgId", "MUTUALLY_EXCLUSIVE", "userId and orgId cannot both be provided.")
	}

	nextUserID := existing.UserID
	nextOrgID := existing.OrgID
	if req.ClearUser != nil && *req.ClearUser {
		nextUserID = nil
	}
	if req.ClearOrg != nil && *req.ClearOrg {
		nextOrgID = nil
	}
	if hasReqUser {
		id := *req.UserID
		nextUserID = &id
		nextOrgID = nil
	}
	if hasReqOrg {
		id := *req.OrgID
		nextOrgID = &id
		nextUserID = nil
	}

	assigneeChanged := !uuidPtrEqual(existing.UserID, nextUserID) || !uuidPtrEqual(existing.OrgID, nextOrgID)
	rawName := existing.Name
	if assigneeChanged {
		rawName = ""
	}
	if req.Name != nil {
		rawName = *req.Name
	}

	resolvedUser, resolvedOrg, resolvedName, err := s.resolveTargetAssignee(ctx, p.ParokiID, nextUserID, nextOrgID, rawName, "userId", "orgId", "name")
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	if resolvedUser == nil {
		upd.ClearUserID = true
	} else {
		upd.UserID = resolvedUser
	}
	if resolvedOrg == nil {
		upd.ClearOrgID = true
	} else {
		upd.OrgID = resolvedOrg
	}
	upd.Name = &resolvedName

	resetToAssigned := assigneeChanged || (req.Status != nil && *req.Status == "assigned")
	if resetToAssigned {
		st := liturgipetugas.StatusAssigned
		if req.Status != nil && *req.Status != "assigned" && !assigneeChanged {
			st = liturgipetugas.Status(*req.Status)
		}
		upd.Status = &st
		upd.ClearResponded = true
		upd.ClearSubstituteUserID = true
		upd.ClearSubstituteOrgID = true
		upd.ClearSubstituteName = true
	} else if req.Status != nil {
		switch *req.Status {
		case "confirmed", "declined":
			st := liturgipetugas.Status(*req.Status)
			now := time.Now().UTC()
			by := p.UserID
			upd.Status = &st
			upd.RespondedAt = &now
			upd.RespondedBy = &by
		default:
			return domain.LiturgiPetugasResponse{}, domain.ValidationError("status", "INVALID_STATUS", "Status must be assigned, confirmed, or declined.")
		}
	}

	updated, err := s.repo.UpdatePetugas(ctx, existing.ID, upd)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	if assigneeChanged {
		s.notifyPetugasAssigned(ctx, p.UserID, ev, updated)
	}
	return toPetugasResponse(updated, true, true, true, hasProposedSubstitute(updated)), nil
}

// RespondPetugas records an officer confirmation or decline with optional substitute proposal.
func (s *Service) RespondPetugas(
	ctx context.Context,
	p authx.Principal,
	eventID, petugasID uuid.UUID,
	req domain.RespondLiturgiPetugasRequest,
) (domain.LiturgiPetugasResponse, error) {
	ev, err := s.ensureMassScheduleExists(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	existing, err := s.repo.FindPetugasByID(ctx, p.ParokiID, eventID, petugasID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.LiturgiPetugasResponse{}, domain.ErrPetugasNotFound
		}
		return domain.LiturgiPetugasResponse{}, err
	}

	canEdit, err := s.canManageRole(ctx, p, string(existing.Role))
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	canRespond := canEdit || (existing.UserID != nil && *existing.UserID == p.UserID)
	if !canRespond && existing.OrgID != nil && s.orgs != nil {
		canRespond, err = s.orgs.IsAdmin(ctx, *existing.OrgID, p.UserID)
		if err != nil {
			return domain.LiturgiPetugasResponse{}, err
		}
	}
	if !canRespond {
		return domain.LiturgiPetugasResponse{}, domain.ErrForbidden
	}

	if req.Status != "confirmed" && req.Status != "declined" {
		return domain.LiturgiPetugasResponse{}, domain.ValidationError("status", "INVALID_STATUS", "Status must be confirmed or declined.")
	}
	if req.SubstituteUserID != nil && *req.SubstituteUserID != uuid.Nil && req.SubstituteOrgID != nil && *req.SubstituteOrgID != uuid.Nil {
		return domain.LiturgiPetugasResponse{}, domain.ValidationError("substituteOrgId", "MUTUALLY_EXCLUSIVE", "substituteUserId and substituteOrgId cannot both be provided.")
	}

	now := time.Now().UTC()
	by := p.UserID
	st := liturgipetugas.Status(req.Status)
	upd := repo.UpdatePetugasInput{
		Status:      &st,
		RespondedAt: &now,
		RespondedBy: &by,
	}
	if req.Note != nil {
		if strings.TrimSpace(*req.Note) == "" {
			upd.ClearNote = true
		} else {
			v := strings.TrimSpace(*req.Note)
			upd.Note = &v
		}
	}

	if req.Status == "confirmed" {
		upd.ClearSubstituteUserID = true
		upd.ClearSubstituteOrgID = true
		upd.ClearSubstituteName = true
	} else {
		hasSubUser := req.SubstituteUserID != nil && *req.SubstituteUserID != uuid.Nil
		hasSubOrg := req.SubstituteOrgID != nil && *req.SubstituteOrgID != uuid.Nil
		subName := ""
		if req.SubstituteName != nil {
			subName = strings.TrimSpace(*req.SubstituteName)
		}
		if hasSubUser || hasSubOrg || subName != "" {
			subUser, subOrg, resolvedSubName, err := s.resolveTargetAssignee(
				ctx,
				p.ParokiID,
				req.SubstituteUserID,
				req.SubstituteOrgID,
				subName,
				"substituteUserId",
				"substituteOrgId",
				"substituteName",
			)
			if err != nil {
				return domain.LiturgiPetugasResponse{}, err
			}
			if subUser != nil {
				upd.SubstituteUserID = subUser
			} else {
				upd.ClearSubstituteUserID = true
			}
			if subOrg != nil {
				upd.SubstituteOrgID = subOrg
			} else {
				upd.ClearSubstituteOrgID = true
			}
			upd.SubstituteName = &resolvedSubName
		} else {
			upd.ClearSubstituteUserID = true
			upd.ClearSubstituteOrgID = true
			upd.ClearSubstituteName = true
		}
	}

	updated, err := s.repo.UpdatePetugas(ctx, existing.ID, upd)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	if req.Status == "declined" {
		s.notifySubstituteProposed(ctx, p.UserID, ev, updated)
	}
	return toPetugasResponse(updated, true, canEdit, true, canEdit && hasProposedSubstitute(updated)), nil
}

// ApproveSubstitute atomically promotes a proposed substitute onto the active assignment.
func (s *Service) ApproveSubstitute(
	ctx context.Context,
	p authx.Principal,
	eventID, petugasID uuid.UUID,
	req domain.ApproveLiturgiSubstituteRequest,
) (domain.LiturgiPetugasResponse, error) {
	ev, err := s.ensureMassScheduleExists(ctx, p, eventID)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	existing, err := s.repo.FindPetugasByID(ctx, p.ParokiID, eventID, petugasID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.LiturgiPetugasResponse{}, domain.ErrPetugasNotFound
		}
		return domain.LiturgiPetugasResponse{}, err
	}

	allowed, err := s.canManageRole(ctx, p, string(existing.Role))
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}
	if !allowed {
		return domain.LiturgiPetugasResponse{}, domain.ErrForbidden
	}
	if !hasProposedSubstitute(existing) {
		return domain.LiturgiPetugasResponse{}, domain.ValidationError("substitute", "SUBSTITUTE_REQUIRED", "No proposed substitute exists for this assignment.")
	}

	nextName := existing.Name
	if existing.SubstituteName != nil && strings.TrimSpace(*existing.SubstituteName) != "" {
		nextName = strings.TrimSpace(*existing.SubstituteName)
	}
	upd := repo.UpdatePetugasInput{
		Name:                  &nextName,
		ClearSubstituteUserID: true,
		ClearSubstituteOrgID:  true,
		ClearSubstituteName:   true,
	}
	if existing.SubstituteUserID != nil {
		uid := *existing.SubstituteUserID
		upd.UserID = &uid
		upd.ClearOrgID = true
	} else if existing.SubstituteOrgID != nil {
		oid := *existing.SubstituteOrgID
		upd.OrgID = &oid
		upd.ClearUserID = true
	} else {
		upd.ClearUserID = true
		upd.ClearOrgID = true
	}

	if req.MarkConfirmed {
		st := liturgipetugas.StatusConfirmed
		now := time.Now().UTC()
		by := p.UserID
		upd.Status = &st
		upd.RespondedAt = &now
		upd.RespondedBy = &by
	} else {
		st := liturgipetugas.StatusAssigned
		upd.Status = &st
		upd.ClearResponded = true
	}

	updated, err := s.repo.UpdatePetugas(ctx, existing.ID, upd)
	if err != nil {
		return domain.LiturgiPetugasResponse{}, err
	}

	s.notifySubstituteApproved(ctx, p.UserID, ev, existing, updated)
	return toPetugasResponse(updated, true, true, true, false), nil
}

// DeletePetugas hard-deletes an officer assignment from a parish Mass schedule.
func (s *Service) DeletePetugas(ctx context.Context, p authx.Principal, eventID, petugasID uuid.UUID) error {
	if _, err := s.ensureMassScheduleExists(ctx, p, eventID); err != nil {
		return err
	}
	existing, err := s.repo.FindPetugasByID(ctx, p.ParokiID, eventID, petugasID)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrPetugasNotFound
		}
		return err
	}
	allowed, err := s.canManageRole(ctx, p, string(existing.Role))
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return s.repo.DeletePetugas(ctx, existing.ID)
}

func hasProposedSubstitute(row *ent.LiturgiPetugas) bool {
	if row == nil {
		return false
	}
	return row.SubstituteUserID != nil ||
		row.SubstituteOrgID != nil ||
		(row.SubstituteName != nil && strings.TrimSpace(*row.SubstituteName) != "")
}

func trimOptionalPtr(v *string) *string {
	if v == nil {
		return nil
	}
	t := strings.TrimSpace(*v)
	if t == "" {
		return nil
	}
	return &t
}

func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
