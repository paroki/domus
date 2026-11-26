package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/eventpetugas"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

func (s *Service) canManageEvent(ctx context.Context, p authx.Principal, ev *ent.Event) (bool, error) {
	if ev.OrgID != nil {
		return s.isOrgAdmin(ctx, p, *ev.OrgID)
	}
	return s.isParokiAdmin(ctx, p)
}

func (s *Service) canConfirmPetugas(ctx context.Context, p authx.Principal, canManage bool, row *ent.EventPetugas, orgAdminCache map[uuid.UUID]bool) (bool, error) {
	if canManage {
		return true, nil
	}
	if row.UserID != nil && *row.UserID == p.UserID {
		return true, nil
	}
	if row.OrgID != nil {
		if orgAdminCache != nil {
			if cached, ok := orgAdminCache[*row.OrgID]; ok {
				return cached, nil
			}
		}
		isAdm, err := s.isOrgAdmin(ctx, p, *row.OrgID)
		if err != nil {
			return false, err
		}
		if orgAdminCache != nil {
			orgAdminCache[*row.OrgID] = isAdm
		}
		return isAdm, nil
	}
	return false, nil
}

// CreateEventPetugas assigns an officer/minister to eventID (requires event manager).
func (s *Service) CreateEventPetugas(
	ctx context.Context,
	p authx.Principal,
	eventID uuid.UUID,
	req domain.CreateEventPetugasRequest,
) (domain.EventPetugasResponse, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventPetugasResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}

	canManage, err := s.canManageEvent(ctx, p, ev)
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}
	if !canManage {
		return domain.EventPetugasResponse{}, domain.ErrForbidden
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		return domain.EventPetugasResponse{}, domain.ValidationError("role", "REQUIRED", "Officer role is required.")
	}

	userID := normalizeOptUUID(req.UserID)
	orgID := normalizeOptUUID(req.OrgID)
	name, err := s.validatePetugasAssignee(ctx, p.ParokiID, userID, orgID, strings.TrimSpace(req.Name))
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}

	urutan := 0
	if req.Urutan != nil {
		urutan = *req.Urutan
	}
	status := eventpetugas.StatusAssigned
	if req.Status != "" {
		status = eventpetugas.Status(req.Status)
	}

	created, err := s.repo.CreatePetugas(ctx, repo.CreateEventPetugasInput{
		ParokiID: p.ParokiID,
		EventID:  eventID,
		Role:     role,
		UserID:   userID,
		OrgID:    orgID,
		Name:     name,
		Urutan:   urutan,
		Status:   status,
		Note:     trimOptString(req.Note),
	})
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}
	return toPetugasResponse(created, true), nil
}

// UpdateEventPetugas updates an officer assignment (manager) or confirms/declines duty (assignee).
func (s *Service) UpdateEventPetugas(
	ctx context.Context,
	p authx.Principal,
	eventID, petugasID uuid.UUID,
	req domain.UpdateEventPetugasRequest,
) (domain.EventPetugasResponse, error) {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.EventPetugasResponse{}, domain.ErrEventNotFound
	}
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}

	row, err := s.repo.FindPetugasInEvent(ctx, p.ParokiID, eventID, petugasID)
	if ent.IsNotFound(err) {
		return domain.EventPetugasResponse{}, domain.ErrPetugasNotFound
	}
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}

	canManage, err := s.canManageEvent(ctx, p, ev)
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}

	if !canManage {
		canConfirm, err := s.canConfirmPetugas(ctx, p, false, row, nil)
		if err != nil {
			return domain.EventPetugasResponse{}, err
		}
		if !canConfirm {
			return domain.EventPetugasResponse{}, domain.ErrForbidden
		}
		if req.Role != nil || req.UserID != nil || req.OrgID != nil || req.Name != nil || req.Urutan != nil {
			return domain.EventPetugasResponse{}, domain.ErrForbidden
		}
		if req.Status != nil && *req.Status != string(eventpetugas.StatusConfirmed) && *req.Status != string(eventpetugas.StatusDeclined) {
			return domain.EventPetugasResponse{}, domain.ErrForbidden
		}
	}

	in := repo.UpdateEventPetugasInput{
		Urutan: req.Urutan,
	}
	if req.Role != nil {
		r := strings.TrimSpace(*req.Role)
		if r == "" {
			return domain.EventPetugasResponse{}, domain.ValidationError("role", "REQUIRED", "Officer role cannot be empty.")
		}
		in.Role = &r
	}
	if req.Status != nil {
		st := eventpetugas.Status(*req.Status)
		in.Status = &st
	}
	if req.Note != nil {
		trimmed := strings.TrimSpace(*req.Note)
		if trimmed == "" {
			in.ClearNote = true
		} else {
			in.Note = &trimmed
		}
	}

	if canManage && (req.UserID != nil || req.OrgID != nil || req.Name != nil) {
		effUserID := row.UserID
		if req.UserID != nil {
			effUserID = normalizeOptUUID(req.UserID)
			if effUserID == nil {
				in.ClearUserID = true
			} else {
				in.UserID = effUserID
			}
		}
		effOrgID := row.OrgID
		if req.OrgID != nil {
			effOrgID = normalizeOptUUID(req.OrgID)
			if effOrgID == nil {
				in.ClearOrgID = true
			} else {
				in.OrgID = effOrgID
			}
		}
		effName := row.Name
		if req.Name != nil {
			effName = strings.TrimSpace(*req.Name)
		}
		resolvedName, err := s.validatePetugasAssignee(ctx, p.ParokiID, effUserID, effOrgID, effName)
		if err != nil {
			return domain.EventPetugasResponse{}, err
		}
		if req.Name != nil || resolvedName != row.Name {
			in.Name = &resolvedName
		}
	}

	updated, err := s.repo.UpdatePetugas(ctx, petugasID, in)
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}
	canConfirm, err := s.canConfirmPetugas(ctx, p, canManage, updated, nil)
	if err != nil {
		return domain.EventPetugasResponse{}, err
	}
	return toPetugasResponse(updated, canConfirm), nil
}

// DeleteEventPetugas hard-deletes an officer assignment from eventID (requires event manager).
func (s *Service) DeleteEventPetugas(ctx context.Context, p authx.Principal, eventID, petugasID uuid.UUID) error {
	ev, err := s.repo.FindEventInParoki(ctx, p.ParokiID, eventID)
	if ent.IsNotFound(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}

	canManage, err := s.canManageEvent(ctx, p, ev)
	if err != nil {
		return err
	}
	if !canManage {
		return domain.ErrForbidden
	}

	if _, err := s.repo.FindPetugasInEvent(ctx, p.ParokiID, eventID, petugasID); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrPetugasNotFound
		}
		return err
	}

	return s.repo.DeletePetugas(ctx, petugasID)
}

func (s *Service) validatePetugasAssignee(
	ctx context.Context,
	parokiID uuid.UUID,
	userID, orgID *uuid.UUID,
	rawName string,
) (string, error) {
	if userID != nil && orgID != nil {
		return "", domain.ValidationError("userId", "MUTUALLY_EXCLUSIVE", "Cannot assign both userId and orgId.")
	}
	name := strings.TrimSpace(rawName)
	if userID != nil {
		u, err := s.repo.FindActiveUser(ctx, *userID)
		if ent.IsNotFound(err) {
			return "", domain.ValidationError("userId", "INVALID_USER", "User not found.")
		}
		if err != nil {
			return "", err
		}
		isMem, err := s.repo.IsParokiMember(ctx, parokiID, *userID)
		if err != nil {
			return "", err
		}
		if !isMem {
			return "", domain.ValidationError("userId", "INVALID_USER", "User is not a member of the active parish.")
		}
		if name == "" {
			name = u.Name
		}
		return name, nil
	}
	if orgID != nil {
		orgInfo, err := s.orgs.GetOrganization(ctx, parokiID, *orgID)
		if err != nil {
			return "", domain.ValidationError("orgId", "INVALID_ORG", "Organization not found in parish.")
		}
		if name == "" {
			name = orgInfo.Name
		}
		return name, nil
	}
	if name == "" {
		return "", domain.ValidationError("name", "REQUIRED", "Officer name is required.")
	}
	return name, nil
}

func normalizeOptUUID(id *uuid.UUID) *uuid.UUID {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id
}

func toPetugasResponse(row *ent.EventPetugas, canConfirm bool) domain.EventPetugasResponse {
	return domain.EventPetugasResponse{
		ID:         row.ID,
		ParokiID:   row.ParokiID,
		EventID:    row.EventID,
		Role:       row.Role,
		UserID:     row.UserID,
		OrgID:      row.OrgID,
		Name:       row.Name,
		Urutan:     row.Urutan,
		Status:     string(row.Status),
		Note:       row.Note,
		CanConfirm: canConfirm,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
