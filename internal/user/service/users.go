package service

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// SearchUsers searches non-deleted users with optional query, status filter, and cursor pagination.
func (s *Service) SearchUsers(
	ctx context.Context,
	q, status string,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.UserResponse, bool, error) {
	rows, err := s.repo.SearchUsers(ctx, strings.TrimSpace(q), status, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.UserResponse, 0, len(rows))
	for _, row := range rows {
		res := toUserResponse(row)
		if roles, err := s.loadGlobalRoles(row.ID); err == nil {
			res.Roles = roles
		}
		items = append(items, res)
	}
	return items, hasMore, nil
}

// AdminCreateUser creates a new active user, links an email account, and assigns global roles.
func (s *Service) AdminCreateUser(ctx context.Context, req domain.AdminCreateUserRequest) (domain.UserResponse, error) {
	normEmail := normalizeEmail(req.Email)
	var phonePtr *string
	if req.Phone != nil {
		normPhone, err := normalizePhone(*req.Phone)
		if err != nil {
			return domain.UserResponse{}, err
		}
		if normPhone != "" {
			phonePtr = &normPhone
		}
	}

	var created *ent.User
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindUserByEmailIncludeDeleted(ctx, normEmail)
		var u *ent.User
		switch {
		case err == nil && existing.DeletedAt == nil:
			return domain.ErrUserEmailConflict
		case err == nil && existing.DeletedAt != nil:
			u, err = txRepo.ReactivateUser(ctx, existing.ID, strings.TrimSpace(req.Name), phonePtr)
		case ent.IsNotFound(err):
			u, err = txRepo.CreateUser(ctx, normEmail, strings.TrimSpace(req.Name), phonePtr)
		default:
			return err
		}
		if err != nil {
			if ent.IsConstraintError(err) {
				return domain.ErrUserEmailConflict
			}
			return err
		}
		if _, err := txRepo.UpsertAccount(ctx, u.ID, domain.ProviderEmail, normEmail, &u.Email, &u.Name, u.AvatarURL, time.Now().UTC()); err != nil {
			return err
		}
		created = u
		return nil
	})
	if err != nil {
		return domain.UserResponse{}, err
	}

	if s.enforcer != nil {
		sub := authz.UserSubject(created.ID)
		if err := s.enforcer.AssignRole(ctx, sub, domain.RoleUser, authz.DomainGlobal); err != nil {
			return domain.UserResponse{}, err
		}
		if req.IsSuperAdmin {
			if err := s.enforcer.AssignRole(ctx, sub, domain.RoleSuperAdmin, authz.DomainGlobal); err != nil {
				return domain.UserResponse{}, err
			}
		}
	}

	res := toUserResponse(created)
	if roles, err := s.loadGlobalRoles(created.ID); err == nil {
		res.Roles = roles
	}
	return res, nil
}

// AdminGetUser fetches a single non-deleted user by ID.
func (s *Service) AdminGetUser(ctx context.Context, id uuid.UUID) (domain.UserResponse, error) {
	u, err := s.repo.FindUserByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.UserResponse{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.UserResponse{}, err
	}
	res := toUserResponse(u)
	if roles, err := s.loadGlobalRoles(u.ID); err == nil {
		res.Roles = roles
	}
	return res, nil
}

// AdminUpdateUser updates a user's profile, status, or Super Admin role while guarding the last active Super Admin.
func (s *Service) AdminUpdateUser(ctx context.Context, id uuid.UUID, req domain.AdminUpdateUserRequest) (domain.UserResponse, error) {
	u, err := s.repo.FindUserByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.UserResponse{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.UserResponse{}, err
	}

	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.UserResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}

	var phonePtr *string
	if req.Phone != nil {
		norm, err := normalizePhone(*req.Phone)
		if err != nil {
			return domain.UserResponse{}, err
		}
		phonePtr = &norm
	}

	var statusPtr *user.Status
	if req.Status != nil {
		st := user.Status(*req.Status)
		statusPtr = &st
	}

	demoting := req.IsSuperAdmin != nil && !*req.IsSuperAdmin
	disabling := statusPtr != nil && *statusPtr == user.StatusDisabled
	if demoting || disabling {
		if err := s.guardNotLastSuperAdmin(ctx, u); err != nil {
			return domain.UserResponse{}, err
		}
	}

	var updated *ent.User
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		var txErr error
		updated, txErr = txRepo.UpdateUser(ctx, id, namePtr, phonePtr, nil, nil, statusPtr)
		if txErr != nil {
			return txErr
		}
		if disabling {
			return txRepo.DeleteSessionsByUserID(ctx, id)
		}
		return nil
	})
	if err != nil {
		return domain.UserResponse{}, err
	}

	if s.enforcer != nil && req.IsSuperAdmin != nil {
		sub := authz.UserSubject(id)
		if *req.IsSuperAdmin {
			if err := s.enforcer.AssignRole(ctx, sub, domain.RoleSuperAdmin, authz.DomainGlobal); err != nil {
				return domain.UserResponse{}, err
			}
		} else {
			if err := s.enforcer.RevokeRole(ctx, sub, domain.RoleSuperAdmin, authz.DomainGlobal); err != nil {
				return domain.UserResponse{}, err
			}
		}
	}

	res := toUserResponse(updated)
	if roles, err := s.loadGlobalRoles(updated.ID); err == nil {
		res.Roles = roles
	}
	return res, nil
}

// AdminDeleteUser soft-deletes a user and revokes all sessions while guarding the last active Super Admin.
func (s *Service) AdminDeleteUser(ctx context.Context, id uuid.UUID) error {
	u, err := s.repo.FindUserByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.ErrUserNotFound
	}
	if err != nil {
		return err
	}

	if err := s.guardNotLastSuperAdmin(ctx, u); err != nil {
		return err
	}

	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if err := txRepo.SoftDeleteUser(ctx, id, time.Now().UTC()); err != nil {
			return err
		}
		if err := txRepo.DeleteSessionsByUserID(ctx, id); err != nil {
			return err
		}
		return txRepo.DeleteAccountsByUserID(ctx, id)
	})
	if err != nil {
		return err
	}

	if s.enforcer != nil {
		sub := authz.UserSubject(id)
		for _, role := range []string{domain.RoleSuperAdmin, domain.RoleUser} {
			if err := s.enforcer.RevokeRole(ctx, sub, role, authz.DomainGlobal); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) guardNotLastSuperAdmin(ctx context.Context, u *ent.User) error {
	if s.enforcer == nil || u.Status != user.StatusActive {
		return nil
	}
	roles, err := s.loadGlobalRoles(u.ID)
	if err != nil {
		return err
	}
	if !slices.Contains(roles, domain.RoleSuperAdmin) {
		return nil
	}

	subs, err := s.enforcer.UsersForRoleInDomain(domain.RoleSuperAdmin, authz.DomainGlobal)
	if err != nil {
		return err
	}
	ids := make([]uuid.UUID, 0, len(subs))
	for _, sub := range subs {
		raw := strings.TrimPrefix(sub, "user:")
		if uid, parseErr := uuid.Parse(raw); parseErr == nil {
			ids = append(ids, uid)
		}
	}
	activeCount, err := s.repo.CountActiveUsersByIDs(ctx, ids)
	if err != nil {
		return err
	}
	if activeCount <= 1 {
		return domain.ErrLastSuperAdminRequired
	}
	return nil
}
