package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

// SearchOrgMembers returns paginated members of orgID (requires org:admin or org:member in orgID).
func (s *Service) SearchOrgMembers(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.SearchOrgMembersRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.OrgMemberResponse, bool, error) {
	if _, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID); ent.IsNotFound(err) {
		return nil, false, domain.ErrOrgNotFound
	} else if err != nil {
		return nil, false, err
	}

	isAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return nil, false, err
	}
	isMem, err := s.isOrgMember(p, orgID)
	if err != nil {
		return nil, false, err
	}
	if !isAdm && !isMem {
		return nil, false, domain.ErrForbidden
	}

	var role, status string
	if req.Filter != nil {
		role = req.Filter.Role
		status = req.Filter.Status
	}

	rows, usersMap, err := s.repo.SearchOrgMembers(
		ctx,
		orgID,
		strings.TrimSpace(req.Q),
		role,
		status,
		cursorCreatedAt,
		cursorID,
		limit+1,
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]domain.OrgMemberResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOrgMemberResponse(row, usersMap[row.UserID]))
	}
	return items, hasMore, nil
}

// AddOrgMember adds or appoints a member to orgID and syncs Casbin roles.
func (s *Service) AddOrgMember(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.AddOrgMemberRequest,
) (domain.OrgMemberResponse, error) {
	orgRow, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID)
	if ent.IsNotFound(err) {
		return domain.OrgMemberResponse{}, domain.ErrOrgNotFound
	}
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}

	isOrgAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}
	isParokiAdm, err := s.isParokiAdmin(p, orgRow.ParokiID)
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}
	if !isOrgAdm && !isParokiAdm {
		return domain.OrgMemberResponse{}, domain.ErrForbidden
	}

	role := orgmember.RoleMember
	if req.Role != "" {
		role = orgmember.Role(req.Role)
	}
	// Paroki admin without org:admin may ONLY appoint org:admin.
	if !isOrgAdm && role != orgmember.RoleAdmin {
		return domain.OrgMemberResponse{}, domain.ErrForbidden
	}

	status := orgmember.StatusActive
	if req.Status != "" {
		status = orgmember.Status(req.Status)
	}

	targetUserID, err := s.resolveOrEnsureMemberUser(ctx, orgRow.ParokiID, req)
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}

	var saved *ent.OrgMember
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindOrgMember(ctx, orgID, targetUserID)
		switch {
		case findErr == nil:
			if existing.Status == orgmember.StatusInactive {
				activeStatus := orgmember.StatusActive
				upd, updErr := txRepo.UpdateOrgMember(
					ctx,
					existing.ID,
					&role,
					&activeStatus,
				)
				if updErr != nil {
					return updErr
				}
				saved = upd
				return nil
			}
			if !isOrgAdm && isParokiAdm {
				// Idempotent upgrade to active admin by paroki-admin.
				adminRole := orgmember.RoleAdmin
				activeStatus := orgmember.StatusActive
				upd, updErr := txRepo.UpdateOrgMember(
					ctx,
					existing.ID,
					&adminRole,
					&activeStatus,
				)
				if updErr != nil {
					return updErr
				}
				saved = upd
				return nil
			}
			return domain.ErrMemberConflict
		case !ent.IsNotFound(findErr):
			return findErr
		}

		created, createErr := txRepo.CreateOrgMember(ctx, repo.OrgMemberInput{
			ParokiID: orgRow.ParokiID,
			OrgID:    orgID,
			UserID:   targetUserID,
			Role:     role,
			Status:   status,
		})
		if createErr != nil {
			if ent.IsConstraintError(createErr) {
				return domain.ErrMemberConflict
			}
			return createErr
		}
		saved = created
		return nil
	})
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}

	if err := s.syncMemberCasbin(ctx, orgID, targetUserID, saved.Role, saved.Status); err != nil {
		return domain.OrgMemberResponse{}, err
	}

	u, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.OrgMemberResponse{}, err
	}
	return toOrgMemberResponse(saved, u), nil
}

// UpdateOrgMember updates a member's role or status and syncs Casbin.
func (s *Service) UpdateOrgMember(
	ctx context.Context,
	p authx.Principal,
	orgID, targetUserID uuid.UUID,
	req domain.UpdateOrgMemberRequest,
) (domain.OrgMemberResponse, error) {
	orgRow, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID)
	if ent.IsNotFound(err) {
		return domain.OrgMemberResponse{}, domain.ErrOrgNotFound
	}
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}

	isOrgAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}
	if !isOrgAdm {
		isParokiAdm, err := s.isParokiAdmin(p, orgRow.ParokiID)
		if err != nil {
			return domain.OrgMemberResponse{}, err
		}
		promotingToAdmin := req.Role != nil && *req.Role == string(orgmember.RoleAdmin)
		if !isParokiAdm || !promotingToAdmin {
			return domain.OrgMemberResponse{}, domain.ErrForbidden
		}
	}

	var rolePtr *orgmember.Role
	if req.Role != nil {
		r := orgmember.Role(*req.Role)
		rolePtr = &r
	}
	var statusPtr *orgmember.Status
	if req.Status != nil {
		st := orgmember.Status(*req.Status)
		statusPtr = &st
	}

	var updated *ent.OrgMember
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindOrgMember(ctx, orgID, targetUserID)
		if ent.IsNotFound(err) {
			return domain.ErrMemberNotFound
		}
		if err != nil {
			return err
		}

		newRole := existing.Role
		if rolePtr != nil {
			newRole = *rolePtr
		}
		newStatus := existing.Status
		if statusPtr != nil {
			newStatus = *statusPtr
		}

		wasActiveAdmin := existing.Role == orgmember.RoleAdmin && existing.Status == orgmember.StatusActive
		remainsActiveAdmin := newRole == orgmember.RoleAdmin && newStatus == orgmember.StatusActive
		if wasActiveAdmin && !remainsActiveAdmin {
			if err := guardNotLastOrgAdmin(ctx, txRepo, orgID); err != nil {
				return err
			}
		}

		u, err := txRepo.UpdateOrgMember(
			ctx,
			existing.ID,
			rolePtr,
			statusPtr,
		)
		if err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		return domain.OrgMemberResponse{}, err
	}

	if err := s.syncMemberCasbin(ctx, orgID, targetUserID, updated.Role, updated.Status); err != nil {
		return domain.OrgMemberResponse{}, err
	}

	u, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.OrgMemberResponse{}, err
	}
	return toOrgMemberResponse(updated, u), nil
}

// RemoveOrgMember removes a member from orgID (org:admin or self-leave) and revokes Casbin roles.
func (s *Service) RemoveOrgMember(ctx context.Context, p authx.Principal, orgID, targetUserID uuid.UUID) error {
	if _, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID); ent.IsNotFound(err) {
		return domain.ErrOrgNotFound
	} else if err != nil {
		return err
	}

	if p.UserID != targetUserID {
		isOrgAdm, err := s.isOrgAdmin(p, orgID)
		if err != nil {
			return err
		}
		if !isOrgAdm {
			return domain.ErrForbidden
		}
	}

	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindOrgMember(ctx, orgID, targetUserID)
		if ent.IsNotFound(err) {
			return domain.ErrMemberNotFound
		}
		if err != nil {
			return err
		}

		if existing.Role == orgmember.RoleAdmin && existing.Status == orgmember.StatusActive {
			if err := guardNotLastOrgAdmin(ctx, txRepo, orgID); err != nil {
				return err
			}
		}
		return txRepo.DeleteOrgMember(ctx, existing.ID)
	})
	if err != nil {
		return err
	}
	return s.revokeMemberCasbin(ctx, orgID, targetUserID)
}

func (s *Service) resolveOrEnsureMemberUser(ctx context.Context, parokiID uuid.UUID, req domain.AddOrgMemberRequest) (uuid.UUID, error) {
	var targetUserID uuid.UUID
	if req.UserID != nil && *req.UserID != uuid.Nil {
		u, err := s.repo.FindUserByID(ctx, *req.UserID)
		if ent.IsNotFound(err) {
			return uuid.Nil, domain.ValidationError("userId", "NOT_FOUND", "User not found.")
		}
		if err != nil {
			return uuid.Nil, err
		}
		targetUserID = u.ID
	} else {
		email := normalizeEmail(req.Email)
		if email == "" {
			return uuid.Nil, domain.ValidationError("email", "REQUIRED", "Email or userId is required.")
		}
		existingUser, err := s.repo.FindUserByEmail(ctx, email)
		if err == nil {
			targetUserID = existingUser.ID
		} else if !ent.IsNotFound(err) {
			return uuid.Nil, err
		} else {
			if s.users == nil {
				return uuid.Nil, domain.ValidationError("email", "UNAVAILABLE", "User onboarding is unavailable.")
			}
			name := strings.TrimSpace(req.Name)
			if name == "" {
				name = email
			}
			var phoneStr string
			if req.Phone != nil {
				norm, err := normalizePhone(*req.Phone)
				if err != nil {
					return uuid.Nil, err
				}
				phoneStr = norm
			}
			uid, err := s.users.EnsureUser(ctx, email, name, phoneStr, "email", email)
			if err != nil {
				return uuid.Nil, err
			}
			targetUserID = uid
		}
	}

	if s.refs != nil {
		if err := s.refs.EnsureParokiMember(ctx, parokiID, targetUserID); err != nil {
			return uuid.Nil, err
		}
	}
	return targetUserID, nil
}

func guardNotLastOrgAdmin(ctx context.Context, r *repo.Repository, orgID uuid.UUID) error {
	count, err := r.CountActiveOrgAdmins(ctx, orgID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return domain.ErrLastAdminRequired
	}
	return nil
}
