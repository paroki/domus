package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/parishmember"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchParishes returns paginated non-deleted parishes.
func (s *Service) SearchParishes(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchParishesRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.ParishResponse, bool, error) {
	var dioceseID, memberUserID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		dioceseID = req.Filter.DioceseID
		active = req.Filter.Active
		if req.Filter.MemberOnly != nil && *req.Filter.MemberOnly {
			uid := p.UserID
			memberUserID = &uid
		}
	}
	rows, err := s.repo.SearchParishes(ctx, strings.TrimSpace(req.Q), dioceseID, active, memberUserID, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.ParishResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toParishResponse(row))
	}
	return items, hasMore, nil
}

// GetParish returns a single non-deleted Parish by ID.
func (s *Service) GetParish(ctx context.Context, id uuid.UUID) (*domain.ParishResponse, error) {
	row, err := s.repo.FindParishByID(ctx, id)
	if ent.IsNotFound(err) {
		return nil, domain.ErrParishNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toParishResponse(row)
	return &res, nil
}

// GetParishByCode returns an active, non-deleted Parish by code.
func (s *Service) GetParishByCode(ctx context.Context, code string) (*domain.ParishResponse, error) {
	row, err := s.repo.FindParishByCode(ctx, code)
	if ent.IsNotFound(err) {
		return nil, domain.ErrParishNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toParishResponse(row)
	return &res, nil
}

// CreateParish creates a Parish or reactivates a soft-deleted one with the same code.
func (s *Service) CreateParish(ctx context.Context, req domain.CreateParishRequest) (domain.ParishResponse, error) {
	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.ParishResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.ParishResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
	}
	phone, err := normalizeOptPhone(req.Phone)
	if err != nil {
		return domain.ParishResponse{}, err
	}
	if _, err := s.repo.FindDioceseByID(ctx, req.DioceseID); ent.IsNotFound(err) {
		return domain.ParishResponse{}, domain.ErrDioceseNotFound
	} else if err != nil {
		return domain.ParishResponse{}, err
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.ParishInput{
		DioceseID:   req.DioceseID,
		Code:        code,
		Name:        name,
		PatronSaint: trimOptString(req.PatronSaint),
		PastorName:  trimOptString(req.PastorName),
		Address:     trimOptString(req.Address),
		Phone:       phone,
		Email:       normalizeOptEmail(req.Email),
		Website:     trimOptString(req.Website),
		LogoURL:     trimOptString(req.LogoURL),
		Letterhead:  trimOptString(req.Letterhead),
		Active:      active,
	}

	var out *ent.Parish
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindParishByCodeIncludeDeleted(ctx, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateParish(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateParish(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.ParishResponse{}, err
	}
	return toParishResponse(out), nil
}

// UpdateParish updates a Parish; super-admins may edit all fields, while parish admins may edit profile fields only.
func (s *Service) UpdateParish(ctx context.Context, p authx.Principal, id uuid.UUID, req domain.UpdateParishRequest) (domain.ParishResponse, error) {
	isSuper, err := s.requireParishAdminOrSuperAdmin(p, id)
	if err != nil {
		return domain.ParishResponse{}, err
	}
	if !isSuper && (req.Code != nil || req.DioceseID != nil || req.Active != nil) {
		return domain.ParishResponse{}, domain.ErrRefForbidden
	}

	existing, err := s.repo.FindParishByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.ParishResponse{}, domain.ErrParishNotFound
	}
	if err != nil {
		return domain.ParishResponse{}, err
	}

	if req.DioceseID != nil {
		if _, err := s.repo.FindDioceseByID(ctx, *req.DioceseID); ent.IsNotFound(err) {
			return domain.ParishResponse{}, domain.ErrDioceseNotFound
		} else if err != nil {
			return domain.ParishResponse{}, err
		}
	}

	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.ParishResponse{}, err
		}
		if normCode != existing.Code {
			if _, findErr := s.repo.FindParishByCodeIncludeDeleted(ctx, normCode); findErr == nil {
				return domain.ParishResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.ParishResponse{}, findErr
			}
		}
		codePtr = &normCode
	}

	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.ParishResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}

	phonePtr, err := normalizeOptPhone(req.Phone)
	if err != nil {
		return domain.ParishResponse{}, err
	}

	updated, err := s.repo.UpdateParish(
		ctx,
		id,
		req.DioceseID,
		codePtr,
		namePtr,
		trimOptString(req.PatronSaint),
		trimOptString(req.PastorName),
		trimOptString(req.Address),
		phonePtr,
		normalizeOptEmail(req.Email),
		trimOptString(req.Website),
		trimOptString(req.LogoURL),
		trimOptString(req.Letterhead),
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.ParishResponse{}, domain.ErrCodeConflict
		}
		return domain.ParishResponse{}, err
	}
	return toParishResponse(updated), nil
}

// UpdateCurrentParishProfile updates profile fields of the caller's active Parish (requires ref:parish-admin in parish).
func (s *Service) UpdateCurrentParishProfile(ctx context.Context, p authx.Principal, req domain.UpdateParishProfileRequest) (domain.ParishResponse, error) {
	isAdmin, err := s.isParishAdmin(p, p.ParishID)
	if err != nil {
		return domain.ParishResponse{}, err
	}
	if !isAdmin {
		return domain.ParishResponse{}, domain.ErrRefForbidden
	}

	if _, err := s.repo.FindParishByID(ctx, p.ParishID); ent.IsNotFound(err) {
		return domain.ParishResponse{}, domain.ErrParishNotFound
	} else if err != nil {
		return domain.ParishResponse{}, err
	}

	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.ParishResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}
	phonePtr, err := normalizeOptPhone(req.Phone)
	if err != nil {
		return domain.ParishResponse{}, err
	}

	updated, err := s.repo.UpdateParish(
		ctx,
		p.ParishID,
		nil,
		nil,
		namePtr,
		trimOptString(req.PatronSaint),
		trimOptString(req.PastorName),
		trimOptString(req.Address),
		phonePtr,
		normalizeOptEmail(req.Email),
		trimOptString(req.Website),
		trimOptString(req.LogoURL),
		trimOptString(req.Letterhead),
		nil,
	)
	if err != nil {
		return domain.ParishResponse{}, err
	}
	return toParishResponse(updated), nil
}

// DeleteParish soft-deletes a Parish if it has no non-deleted child master-data rows.
func (s *Service) DeleteParish(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindParishByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrParishNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasParishChildren(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteParish(ctx, id, time.Now().UTC())
}

// SearchParishMembers returns paginated members of parishID joined with user info.
func (s *Service) SearchParishMembers(
	ctx context.Context,
	p authx.Principal,
	parishID uuid.UUID,
	req domain.SearchParishMembersRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.ParishMemberResponse, bool, error) {
	if _, err := s.repo.FindParishByID(ctx, parishID); ent.IsNotFound(err) {
		return nil, false, domain.ErrParishNotFound
	} else if err != nil {
		return nil, false, err
	}

	if !s.isSuperAdmin(p) {
		isAdmin, err := s.isParishAdmin(p, parishID)
		if err != nil {
			return nil, false, err
		}
		if !isAdmin {
			if _, err := s.repo.FindParishMember(ctx, parishID, p.UserID); ent.IsNotFound(err) {
				return nil, false, domain.ErrRefForbidden
			} else if err != nil {
				return nil, false, err
			}
		}
	}

	var role string
	if req.Filter != nil {
		role = req.Filter.Role
	}
	rows, usersMap, err := s.repo.SearchParishMembers(ctx, parishID, strings.TrimSpace(req.Q), role, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.ParishMemberResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toParishMemberResponse(row, usersMap[row.UserID]))
	}
	return items, hasMore, nil
}

// AddParishMember adds an existing or invited user to parishID and syncs Casbin roles.
func (s *Service) AddParishMember(
	ctx context.Context,
	p authx.Principal,
	parishID uuid.UUID,
	req domain.AddParishMemberRequest,
) (domain.ParishMemberResponse, error) {
	if _, err := s.requireParishAdminOrSuperAdmin(p, parishID); err != nil {
		return domain.ParishMemberResponse{}, err
	}
	if _, err := s.repo.FindParishByID(ctx, parishID); ent.IsNotFound(err) {
		return domain.ParishMemberResponse{}, domain.ErrParishNotFound
	} else if err != nil {
		return domain.ParishMemberResponse{}, err
	}

	role := req.Role
	if role == "" {
		role = string(parishmember.RoleMember)
	}

	var targetUserID uuid.UUID
	if req.UserID != nil && *req.UserID != uuid.Nil {
		u, err := s.repo.FindUserByID(ctx, *req.UserID)
		if ent.IsNotFound(err) {
			return domain.ParishMemberResponse{}, domain.ValidationError("userId", "NOT_FOUND", "User not found.")
		}
		if err != nil {
			return domain.ParishMemberResponse{}, err
		}
		targetUserID = u.ID
	} else {
		email := normalizeEmail(req.Email)
		if email == "" {
			return domain.ParishMemberResponse{}, domain.ValidationError("email", "REQUIRED", "Email or userId is required.")
		}
		existingUser, err := s.repo.FindUserByEmail(ctx, email)
		if err == nil {
			targetUserID = existingUser.ID
		} else if !ent.IsNotFound(err) {
			return domain.ParishMemberResponse{}, err
		} else {
			name := strings.TrimSpace(req.Name)
			if name == "" {
				name = email
			}
			var phoneStr string
			if req.Phone != nil {
				norm, err := normalizePhone(*req.Phone)
				if err != nil {
					return domain.ParishMemberResponse{}, err
				}
				phoneStr = norm
			}
			uid, err := s.users.EnsureUser(ctx, email, name, phoneStr, "email", email)
			if err != nil {
				return domain.ParishMemberResponse{}, err
			}
			targetUserID = uid
		}
	}

	var created *ent.ParishMember
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if _, err := txRepo.FindParishMember(ctx, parishID, targetUserID); err == nil {
			return domain.ErrParishMemberConflict
		} else if !ent.IsNotFound(err) {
			return err
		}
		pm, err := txRepo.CreateParishMember(ctx, parishID, targetUserID, parishmember.Role(role))
		if err != nil {
			if ent.IsConstraintError(err) {
				return domain.ErrParishMemberConflict
			}
			return err
		}
		if err := txRepo.SetUserDefaultParishIfNil(ctx, targetUserID, parishID); err != nil {
			return err
		}
		created = pm
		return nil
	})
	if err != nil {
		return domain.ParishMemberResponse{}, err
	}

	if role == string(parishmember.RoleAdmin) && s.enforcer != nil {
		if err := s.enforcer.AssignRole(ctx, authz.UserSubject(targetUserID), roleParishAdmin, authz.ParishDomain(parishID)); err != nil {
			return domain.ParishMemberResponse{}, err
		}
	}

	u, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.ParishMemberResponse{}, err
	}
	return toParishMemberResponse(created, u), nil
}

// UpdateParishMember updates a member's role and syncs Casbin while guarding the last parish admin.
func (s *Service) UpdateParishMember(
	ctx context.Context,
	p authx.Principal,
	parishID, targetUserID uuid.UUID,
	req domain.UpdateParishMemberRequest,
) (domain.ParishMemberResponse, error) {
	if _, err := s.requireParishAdminOrSuperAdmin(p, parishID); err != nil {
		return domain.ParishMemberResponse{}, err
	}

	var updated *ent.ParishMember
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindParishMember(ctx, parishID, targetUserID)
		if ent.IsNotFound(err) {
			return domain.ErrParishMemberNotFound
		}
		if err != nil {
			return err
		}

		if existing.Role == parishmember.RoleAdmin && req.Role == string(parishmember.RoleMember) {
			if err := guardNotLastParishAdmin(ctx, txRepo, parishID); err != nil {
				return err
			}
		}

		u, err := txRepo.UpdateParishMemberRole(ctx, existing.ID, parishmember.Role(req.Role))
		if err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		return domain.ParishMemberResponse{}, err
	}

	if s.enforcer != nil {
		sub := authz.UserSubject(targetUserID)
		dom := authz.ParishDomain(parishID)
		if req.Role == string(parishmember.RoleAdmin) {
			if err := s.enforcer.AssignRole(ctx, sub, roleParishAdmin, dom); err != nil {
				return domain.ParishMemberResponse{}, err
			}
		} else {
			if err := s.enforcer.RevokeRole(ctx, sub, roleParishAdmin, dom); err != nil {
				return domain.ParishMemberResponse{}, err
			}
		}
	}

	u, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.ParishMemberResponse{}, err
	}
	return toParishMemberResponse(updated, u), nil
}

// RemoveParishMember removes a member from parishID and revokes Casbin parish-admin role while guarding the last admin.
func (s *Service) RemoveParishMember(ctx context.Context, p authx.Principal, parishID, targetUserID uuid.UUID) error {
	if _, err := s.requireParishAdminOrSuperAdmin(p, parishID); err != nil {
		return err
	}

	var wasAdmin bool
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindParishMember(ctx, parishID, targetUserID)
		if ent.IsNotFound(err) {
			return domain.ErrParishMemberNotFound
		}
		if err != nil {
			return err
		}

		if existing.Role == parishmember.RoleAdmin {
			if err := guardNotLastParishAdmin(ctx, txRepo, parishID); err != nil {
				return err
			}
			wasAdmin = true
		}

		return txRepo.DeleteParishMember(ctx, existing.ID)
	})
	if err != nil {
		return err
	}
	if wasAdmin && s.enforcer != nil {
		return s.enforcer.RevokeRole(ctx, authz.UserSubject(targetUserID), roleParishAdmin, authz.ParishDomain(parishID))
	}
	return nil
}

func guardNotLastParishAdmin(ctx context.Context, r *repo.Repository, parishID uuid.UUID) error {
	count, err := r.CountParishAdmins(ctx, parishID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return domain.ErrLastParishAdminRequired
	}
	return nil
}

func toParishResponse(p *ent.Parish) domain.ParishResponse {
	return domain.ParishResponse{
		ID:          p.ID,
		DioceseID:   p.DioceseID,
		Code:        p.Code,
		Name:        p.Name,
		PatronSaint: p.PatronSaint,
		PastorName:  p.PastorName,
		Address:     p.Address,
		Phone:       p.Phone,
		Email:       p.Email,
		Website:     p.Website,
		LogoURL:     p.LogoURL,
		Letterhead:  p.Letterhead,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toParishMemberResponse(pm *ent.ParishMember, u *ent.User) domain.ParishMemberResponse {
	res := domain.ParishMemberResponse{
		ID:        pm.ID,
		ParishID:  pm.ParishID,
		UserID:    pm.UserID,
		Role:      string(pm.Role),
		CreatedAt: pm.CreatedAt,
		UpdatedAt: pm.UpdatedAt,
	}
	if u != nil {
		res.Email = u.Email
		res.Name = u.Name
		res.Phone = u.Phone
	}
	return res
}
