package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchParoki returns paginated non-deleted paroki.
func (s *Service) SearchParoki(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchParokiRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.ParokiResponse, bool, error) {
	var keuskupanID, memberUserID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		keuskupanID = req.Filter.KeuskupanID
		active = req.Filter.Active
		if req.Filter.MemberOnly != nil && *req.Filter.MemberOnly {
			uid := p.UserID
			memberUserID = &uid
		}
	}
	rows, err := s.repo.SearchParoki(ctx, strings.TrimSpace(req.Q), keuskupanID, active, memberUserID, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.ParokiResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toParokiResponse(row))
	}
	return items, hasMore, nil
}

// GetParoki returns a single non-deleted Paroki by ID.
func (s *Service) GetParoki(ctx context.Context, id uuid.UUID) (*domain.ParokiResponse, error) {
	row, err := s.repo.FindParokiByID(ctx, id)
	if ent.IsNotFound(err) {
		return nil, domain.ErrParokiNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toParokiResponse(row)
	return &res, nil
}

// GetParokiByCode returns an active, non-deleted Paroki by kode.
func (s *Service) GetParokiByCode(ctx context.Context, kode string) (*domain.ParokiResponse, error) {
	row, err := s.repo.FindParokiByKode(ctx, kode)
	if ent.IsNotFound(err) {
		return nil, domain.ErrParokiNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toParokiResponse(row)
	return &res, nil
}

// CreateParoki creates a Paroki or reactivates a soft-deleted one with the same kode.
func (s *Service) CreateParoki(ctx context.Context, req domain.CreateParokiRequest) (domain.ParokiResponse, error) {
	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.ParokiResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	telepon, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	if _, err := s.repo.FindKeuskupanByID(ctx, req.KeuskupanID); ent.IsNotFound(err) {
		return domain.ParokiResponse{}, domain.ErrKeuskupanNotFound
	} else if err != nil {
		return domain.ParokiResponse{}, err
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.ParokiInput{
		KeuskupanID: req.KeuskupanID,
		Kode:        kode,
		Nama:        nama,
		Pelindung:   trimOptString(req.Pelindung),
		Pastor:      trimOptString(req.Pastor),
		Alamat:      trimOptString(req.Alamat),
		Telepon:     telepon,
		Email:       normalizeOptEmail(req.Email),
		Website:     trimOptString(req.Website),
		LogoURL:     trimOptString(req.LogoURL),
		KopSurat:    trimOptString(req.KopSurat),
		Active:      active,
	}

	var out *ent.Paroki
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindParokiByKodeIncludeDeleted(ctx, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateParoki(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateParoki(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	return toParokiResponse(out), nil
}

// UpdateParoki updates a Paroki; super-admins may edit all fields, while paroki admins may edit profile fields only.
func (s *Service) UpdateParoki(ctx context.Context, p authx.Principal, id uuid.UUID, req domain.UpdateParokiRequest) (domain.ParokiResponse, error) {
	isSuper, err := s.requireParokiAdminOrSuperAdmin(p, id)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	if !isSuper && (req.Kode != nil || req.KeuskupanID != nil || req.Active != nil) {
		return domain.ParokiResponse{}, domain.ErrRefForbidden
	}

	existing, err := s.repo.FindParokiByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.ParokiResponse{}, domain.ErrParokiNotFound
	}
	if err != nil {
		return domain.ParokiResponse{}, err
	}

	if req.KeuskupanID != nil {
		if _, err := s.repo.FindKeuskupanByID(ctx, *req.KeuskupanID); ent.IsNotFound(err) {
			return domain.ParokiResponse{}, domain.ErrKeuskupanNotFound
		} else if err != nil {
			return domain.ParokiResponse{}, err
		}
	}

	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.ParokiResponse{}, err
		}
		if normKode != existing.Kode {
			if _, findErr := s.repo.FindParokiByKodeIncludeDeleted(ctx, normKode); findErr == nil {
				return domain.ParokiResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.ParokiResponse{}, findErr
			}
		}
		kodePtr = &normKode
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.ParokiResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	teleponPtr, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.ParokiResponse{}, err
	}

	updated, err := s.repo.UpdateParoki(
		ctx,
		id,
		req.KeuskupanID,
		kodePtr,
		namaPtr,
		trimOptString(req.Pelindung),
		trimOptString(req.Pastor),
		trimOptString(req.Alamat),
		teleponPtr,
		normalizeOptEmail(req.Email),
		trimOptString(req.Website),
		trimOptString(req.LogoURL),
		trimOptString(req.KopSurat),
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.ParokiResponse{}, domain.ErrCodeConflict
		}
		return domain.ParokiResponse{}, err
	}
	return toParokiResponse(updated), nil
}

// UpdateCurrentParokiProfile updates profile fields of the caller's active Paroki (requires ref:paroki-admin in paroki).
func (s *Service) UpdateCurrentParokiProfile(ctx context.Context, p authx.Principal, req domain.UpdateParokiProfileRequest) (domain.ParokiResponse, error) {
	isAdmin, err := s.isParokiAdmin(p, p.ParokiID)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	if !isAdmin {
		return domain.ParokiResponse{}, domain.ErrRefForbidden
	}

	if _, err := s.repo.FindParokiByID(ctx, p.ParokiID); ent.IsNotFound(err) {
		return domain.ParokiResponse{}, domain.ErrParokiNotFound
	} else if err != nil {
		return domain.ParokiResponse{}, err
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.ParokiResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}
	phonePtr, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.ParokiResponse{}, err
	}

	updated, err := s.repo.UpdateParoki(
		ctx,
		p.ParokiID,
		nil,
		nil,
		namaPtr,
		trimOptString(req.Pelindung),
		trimOptString(req.Pastor),
		trimOptString(req.Alamat),
		phonePtr,
		normalizeOptEmail(req.Email),
		trimOptString(req.Website),
		trimOptString(req.LogoURL),
		trimOptString(req.KopSurat),
		nil,
	)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	return toParokiResponse(updated), nil
}

// DeleteParoki soft-deletes a Paroki if it has no non-deleted child master-data rows.
func (s *Service) DeleteParoki(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindParokiByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrParokiNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasParokiChildren(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteParoki(ctx, id, time.Now().UTC())
}

// SearchParokiMembers returns paginated members of parokiID joined with user info.
func (s *Service) SearchParokiMembers(
	ctx context.Context,
	p authx.Principal,
	parokiID uuid.UUID,
	req domain.SearchParokiMembersRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.ParokiMemberResponse, bool, error) {
	if _, err := s.repo.FindParokiByID(ctx, parokiID); ent.IsNotFound(err) {
		return nil, false, domain.ErrParokiNotFound
	} else if err != nil {
		return nil, false, err
	}

	if !s.isSuperAdmin(p) {
		isAdmin, err := s.isParokiAdmin(p, parokiID)
		if err != nil {
			return nil, false, err
		}
		if !isAdmin {
			if _, err := s.repo.FindParokiMember(ctx, parokiID, p.UserID); ent.IsNotFound(err) {
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
	rows, usersMap, err := s.repo.SearchParokiMembers(ctx, parokiID, strings.TrimSpace(req.Q), role, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.ParokiMemberResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toParokiMemberResponse(row, usersMap[row.UserID]))
	}
	return items, hasMore, nil
}

// AddParokiMember adds an existing or invited user to parokiID and syncs Casbin roles.
func (s *Service) AddParokiMember(
	ctx context.Context,
	p authx.Principal,
	parokiID uuid.UUID,
	req domain.AddParokiMemberRequest,
) (domain.ParokiMemberResponse, error) {
	if _, err := s.requireParokiAdminOrSuperAdmin(p, parokiID); err != nil {
		return domain.ParokiMemberResponse{}, err
	}
	if _, err := s.repo.FindParokiByID(ctx, parokiID); ent.IsNotFound(err) {
		return domain.ParokiMemberResponse{}, domain.ErrParokiNotFound
	} else if err != nil {
		return domain.ParokiMemberResponse{}, err
	}

	role := req.Role
	if role == "" {
		role = string(parokimember.RoleMember)
	}

	var targetUserID uuid.UUID
	if req.UserID != nil && *req.UserID != uuid.Nil {
		u, err := s.repo.FindUserByID(ctx, *req.UserID)
		if ent.IsNotFound(err) {
			return domain.ParokiMemberResponse{}, domain.ValidationError("userId", "NOT_FOUND", "User not found.")
		}
		if err != nil {
			return domain.ParokiMemberResponse{}, err
		}
		targetUserID = u.ID
	} else {
		email := normalizeEmail(req.Email)
		if email == "" {
			return domain.ParokiMemberResponse{}, domain.ValidationError("email", "REQUIRED", "Email or userId is required.")
		}
		existingUser, err := s.repo.FindUserByEmail(ctx, email)
		if err == nil {
			targetUserID = existingUser.ID
		} else if !ent.IsNotFound(err) {
			return domain.ParokiMemberResponse{}, err
		} else {
			name := strings.TrimSpace(req.Name)
			if name == "" {
				name = email
			}
			var phoneStr string
			if req.Phone != nil {
				norm, err := normalizePhone(*req.Phone)
				if err != nil {
					return domain.ParokiMemberResponse{}, err
				}
				phoneStr = norm
			}
			uid, err := s.users.EnsureUser(ctx, email, name, phoneStr, "email", email)
			if err != nil {
				return domain.ParokiMemberResponse{}, err
			}
			targetUserID = uid
		}
	}

	var created *ent.ParokiMember
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if _, err := txRepo.FindParokiMember(ctx, parokiID, targetUserID); err == nil {
			return domain.ErrParokiMemberConflict
		} else if !ent.IsNotFound(err) {
			return err
		}
		pm, err := txRepo.CreateParokiMember(ctx, parokiID, targetUserID, parokimember.Role(role))
		if err != nil {
			if ent.IsConstraintError(err) {
				return domain.ErrParokiMemberConflict
			}
			return err
		}
		if err := txRepo.SetUserDefaultParokiIfNil(ctx, targetUserID, parokiID); err != nil {
			return err
		}
		created = pm
		return nil
	})
	if err != nil {
		return domain.ParokiMemberResponse{}, err
	}

	if role == string(parokimember.RoleAdmin) && s.enforcer != nil {
		if err := s.enforcer.AssignRole(ctx, authz.UserSubject(targetUserID), roleParokiAdmin, authz.ParokiDomain(parokiID)); err != nil {
			return domain.ParokiMemberResponse{}, err
		}
	}

	u, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.ParokiMemberResponse{}, err
	}
	return toParokiMemberResponse(created, u), nil
}

// UpdateParokiMember updates a member's role and syncs Casbin while guarding the last paroki admin.
func (s *Service) UpdateParokiMember(
	ctx context.Context,
	p authx.Principal,
	parokiID, targetUserID uuid.UUID,
	req domain.UpdateParokiMemberRequest,
) (domain.ParokiMemberResponse, error) {
	if _, err := s.requireParokiAdminOrSuperAdmin(p, parokiID); err != nil {
		return domain.ParokiMemberResponse{}, err
	}

	var updated *ent.ParokiMember
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindParokiMember(ctx, parokiID, targetUserID)
		if ent.IsNotFound(err) {
			return domain.ErrParokiMemberNotFound
		}
		if err != nil {
			return err
		}

		if existing.Role == parokimember.RoleAdmin && req.Role == string(parokimember.RoleMember) {
			if err := guardNotLastParokiAdmin(ctx, txRepo, parokiID); err != nil {
				return err
			}
		}

		u, err := txRepo.UpdateParokiMemberRole(ctx, existing.ID, parokimember.Role(req.Role))
		if err != nil {
			return err
		}
		updated = u
		return nil
	})
	if err != nil {
		return domain.ParokiMemberResponse{}, err
	}

	if s.enforcer != nil {
		sub := authz.UserSubject(targetUserID)
		dom := authz.ParokiDomain(parokiID)
		if req.Role == string(parokimember.RoleAdmin) {
			if err := s.enforcer.AssignRole(ctx, sub, roleParokiAdmin, dom); err != nil {
				return domain.ParokiMemberResponse{}, err
			}
		} else {
			if err := s.enforcer.RevokeRole(ctx, sub, roleParokiAdmin, dom); err != nil {
				return domain.ParokiMemberResponse{}, err
			}
		}
	}

	u, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.ParokiMemberResponse{}, err
	}
	return toParokiMemberResponse(updated, u), nil
}

// RemoveParokiMember removes a member from parokiID and revokes Casbin paroki-admin role while guarding the last admin.
func (s *Service) RemoveParokiMember(ctx context.Context, p authx.Principal, parokiID, targetUserID uuid.UUID) error {
	if _, err := s.requireParokiAdminOrSuperAdmin(p, parokiID); err != nil {
		return err
	}

	var wasAdmin bool
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, err := txRepo.FindParokiMember(ctx, parokiID, targetUserID)
		if ent.IsNotFound(err) {
			return domain.ErrParokiMemberNotFound
		}
		if err != nil {
			return err
		}

		if existing.Role == parokimember.RoleAdmin {
			if err := guardNotLastParokiAdmin(ctx, txRepo, parokiID); err != nil {
				return err
			}
			wasAdmin = true
		}

		return txRepo.DeleteParokiMember(ctx, existing.ID)
	})
	if err != nil {
		return err
	}
	if wasAdmin && s.enforcer != nil {
		return s.enforcer.RevokeRole(ctx, authz.UserSubject(targetUserID), roleParokiAdmin, authz.ParokiDomain(parokiID))
	}
	return nil
}

func guardNotLastParokiAdmin(ctx context.Context, r *repo.Repository, parokiID uuid.UUID) error {
	count, err := r.CountParokiAdmins(ctx, parokiID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return domain.ErrLastParokiAdminRequired
	}
	return nil
}

func toParokiResponse(p *ent.Paroki) domain.ParokiResponse {
	return domain.ParokiResponse{
		ID:          p.ID,
		KeuskupanID: p.KeuskupanID,
		Kode:        p.Kode,
		Nama:        p.Nama,
		Pelindung:   p.Pelindung,
		Pastor:      p.Pastor,
		Alamat:      p.Alamat,
		Telepon:     p.Telepon,
		Email:       p.Email,
		Website:     p.Website,
		LogoURL:     p.LogoURL,
		KopSurat:    p.KopSurat,
		Active:      p.Active,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toParokiMemberResponse(pm *ent.ParokiMember, u *ent.User) domain.ParokiMemberResponse {
	res := domain.ParokiMemberResponse{
		ID:        pm.ID,
		ParokiID:  pm.ParokiID,
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
