package service

import (
	"context"
	"math"
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

// DefaultParokiTimezone is the default IANA timezone assigned to new Paroki records.
const DefaultParokiTimezone = "Asia/Makassar"

// SearchParoki returns paginated non-deleted paroki.
func (s *Service) SearchParoki(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchParokiRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.ParokiResponse, bool, error) {
	var keuskupanID, kevikepanID, memberUserID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		keuskupanID = req.Filter.KeuskupanID
		kevikepanID = req.Filter.KevikepanID
		active = req.Filter.Active
		if req.Filter.MemberOnly != nil && *req.Filter.MemberOnly {
			uid := p.UserID
			memberUserID = &uid
		}
	}
	rows, err := s.repo.SearchParoki(ctx, strings.TrimSpace(req.Q), keuskupanID, kevikepanID, active, memberUserID, cursorCreatedAt, cursorID, limit+1)
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
	kodeResmi, err := normalizeOptMaxString("kodeResmi", req.KodeResmi, 64, false)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	fax, err := normalizeOptMaxString("fax", req.Fax, 32, false)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	statusParoki, err := normalizeOptStatusParoki(req.StatusParoki, false)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	tz := DefaultParokiTimezone
	if req.Timezone != nil && strings.TrimSpace(*req.Timezone) != "" {
		tz, err = normalizeTimezone(*req.Timezone)
		if err != nil {
			return domain.ParokiResponse{}, err
		}
	}
	if _, err := s.repo.FindKeuskupanByID(ctx, req.KeuskupanID); ent.IsNotFound(err) {
		return domain.ParokiResponse{}, domain.ErrKeuskupanNotFound
	} else if err != nil {
		return domain.ParokiResponse{}, err
	}

	kevikepanID := nonNilUUID(req.KevikepanID)
	indukID := nonNilUUID(req.IndukID)
	kotaID := nonNilUUID(req.KotaID)
	kecamatanID := nonNilUUID(req.KecamatanID)
	kelurahanID := nonNilUUID(req.KelurahanID)

	if err := s.validateParokiHierarchy(ctx, uuid.Nil, req.KeuskupanID, kevikepanID, indukID, kotaID, kecamatanID, kelurahanID); err != nil {
		return domain.ParokiResponse{}, err
	}
	if err := validateCoordinates(req.Latitude, req.Longitude); err != nil {
		return domain.ParokiResponse{}, err
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.ParokiInput{
		KeuskupanID:  req.KeuskupanID,
		KevikepanID:  kevikepanID,
		IndukID:      indukID,
		Kode:         kode,
		KodeResmi:    kodeResmi,
		Nama:         nama,
		Pelindung:    trimOptString(req.Pelindung),
		Pastor:       trimOptString(req.Pastor),
		Alamat:       trimOptString(req.Alamat),
		KotaID:       kotaID,
		KecamatanID:  kecamatanID,
		KelurahanID:  kelurahanID,
		Timezone:     tz,
		Telepon:      telepon,
		Fax:          fax,
		Email:        normalizeOptEmail(req.Email),
		Website:      trimOptString(req.Website),
		LogoURL:      trimOptString(req.LogoURL),
		KopSurat:     trimOptString(req.KopSurat),
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		StatusParoki: statusParoki,
		Active:       active,
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
	if !isSuper && (req.KeuskupanID != nil || req.KevikepanID != nil || req.IndukID != nil ||
		req.Kode != nil || req.KodeResmi != nil || req.StatusParoki != nil || req.Active != nil) {
		return domain.ParokiResponse{}, domain.ErrRefForbidden
	}

	existing, err := s.repo.FindParokiByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.ParokiResponse{}, domain.ErrParokiNotFound
	}
	if err != nil {
		return domain.ParokiResponse{}, err
	}

	effKeuskupanID := existing.KeuskupanID
	if req.KeuskupanID != nil {
		if _, err := s.repo.FindKeuskupanByID(ctx, *req.KeuskupanID); ent.IsNotFound(err) {
			return domain.ParokiResponse{}, domain.ErrKeuskupanNotFound
		} else if err != nil {
			return domain.ParokiResponse{}, err
		}
		effKeuskupanID = *req.KeuskupanID
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
	kodeResmiPtr, err := normalizeOptMaxString("kodeResmi", req.KodeResmi, 64, true)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	faxPtr, err := normalizeOptMaxString("fax", req.Fax, 32, true)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	statusParokiPtr, err := normalizeOptStatusParoki(req.StatusParoki, true)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	var tzPtr *string
	if req.Timezone != nil {
		normTZ, err := normalizeTimezone(*req.Timezone)
		if err != nil {
			return domain.ParokiResponse{}, err
		}
		tzPtr = &normTZ
	}

	clearCoords := req.ClearCoordinates != nil && *req.ClearCoordinates
	effLat, effLng, err := resolveUpdateCoordinates(existing.Latitude, existing.Longitude, req.Latitude, req.Longitude, clearCoords)
	if err != nil {
		return domain.ParokiResponse{}, err
	}

	kotaReq, kecReq, kelReq := cascadeCivilClear(req.KotaID, req.KecamatanID, req.KelurahanID)
	effKevikepanID := resolveOptUUID(existing.KevikepanID, req.KevikepanID)
	effIndukID := resolveOptUUID(existing.IndukID, req.IndukID)
	effKotaID := resolveOptUUID(existing.KotaID, kotaReq)
	effKecamatanID := resolveOptUUID(existing.KecamatanID, kecReq)
	effKelurahanID := resolveOptUUID(existing.KelurahanID, kelReq)

	if err := s.validateParokiHierarchy(ctx, id, effKeuskupanID, effKevikepanID, effIndukID, effKotaID, effKecamatanID, effKelurahanID); err != nil {
		return domain.ParokiResponse{}, err
	}

	updated, err := s.repo.UpdateParoki(ctx, id, repo.ParokiUpdate{
		KeuskupanID:      req.KeuskupanID,
		KevikepanID:      req.KevikepanID,
		IndukID:          req.IndukID,
		Kode:             kodePtr,
		KodeResmi:        kodeResmiPtr,
		Nama:             namaPtr,
		Pelindung:        trimOptString(req.Pelindung),
		Pastor:           trimOptString(req.Pastor),
		Alamat:           trimOptString(req.Alamat),
		KotaID:           kotaReq,
		KecamatanID:      kecReq,
		KelurahanID:      kelReq,
		Timezone:         tzPtr,
		Telepon:          teleponPtr,
		Fax:              faxPtr,
		Email:            normalizeOptEmail(req.Email),
		Website:          trimOptString(req.Website),
		LogoURL:          trimOptString(req.LogoURL),
		KopSurat:         trimOptString(req.KopSurat),
		Latitude:         effLat,
		Longitude:        effLng,
		ClearCoordinates: clearCoords,
		StatusParoki:     statusParokiPtr,
		Active:           req.Active,
	})
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

	existing, err := s.repo.FindParokiByID(ctx, p.ParokiID)
	if ent.IsNotFound(err) {
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
	faxPtr, err := normalizeOptMaxString("fax", req.Fax, 32, true)
	if err != nil {
		return domain.ParokiResponse{}, err
	}
	var tzPtr *string
	if req.Timezone != nil {
		normTZ, err := normalizeTimezone(*req.Timezone)
		if err != nil {
			return domain.ParokiResponse{}, err
		}
		tzPtr = &normTZ
	}

	clearCoords := req.ClearCoordinates != nil && *req.ClearCoordinates
	effLat, effLng, err := resolveUpdateCoordinates(existing.Latitude, existing.Longitude, req.Latitude, req.Longitude, clearCoords)
	if err != nil {
		return domain.ParokiResponse{}, err
	}

	kotaReq, kecReq, kelReq := cascadeCivilClear(req.KotaID, req.KecamatanID, req.KelurahanID)
	effKotaID := resolveOptUUID(existing.KotaID, kotaReq)
	effKecamatanID := resolveOptUUID(existing.KecamatanID, kecReq)
	effKelurahanID := resolveOptUUID(existing.KelurahanID, kelReq)

	if err := s.validateParokiHierarchy(ctx, p.ParokiID, existing.KeuskupanID, existing.KevikepanID, existing.IndukID, effKotaID, effKecamatanID, effKelurahanID); err != nil {
		return domain.ParokiResponse{}, err
	}

	updated, err := s.repo.UpdateParoki(ctx, p.ParokiID, repo.ParokiUpdate{
		Nama:             namaPtr,
		Pelindung:        trimOptString(req.Pelindung),
		Pastor:           trimOptString(req.Pastor),
		Alamat:           trimOptString(req.Alamat),
		KotaID:           kotaReq,
		KecamatanID:      kecReq,
		KelurahanID:      kelReq,
		Timezone:         tzPtr,
		Telepon:          phonePtr,
		Fax:              faxPtr,
		Email:            normalizeOptEmail(req.Email),
		Website:          trimOptString(req.Website),
		LogoURL:          trimOptString(req.LogoURL),
		KopSurat:         trimOptString(req.KopSurat),
		Latitude:         effLat,
		Longitude:        effLng,
		ClearCoordinates: clearCoords,
	})
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

func (s *Service) validateParokiHierarchy(
	ctx context.Context,
	parokiID uuid.UUID,
	keuskupanID uuid.UUID,
	kevikepanID *uuid.UUID,
	indukID *uuid.UUID,
	kotaID *uuid.UUID,
	kecamatanID *uuid.UUID,
	kelurahanID *uuid.UUID,
) error {
	if kevikepanID != nil {
		kev, err := s.repo.FindKevikepanByID(ctx, *kevikepanID)
		if ent.IsNotFound(err) {
			return domain.ErrKevikepanNotFound
		}
		if err != nil {
			return err
		}
		if kev.KeuskupanID != keuskupanID {
			return domain.ValidationError("kevikepanId", "KEUSKUPAN_MISMATCH", "Kevikepan must belong to the same Keuskupan.")
		}
	}

	if indukID != nil {
		if parokiID != uuid.Nil && *indukID == parokiID {
			return domain.ValidationError("indukId", "SELF_REFERENCE", "Paroki cannot reference itself as induk.")
		}
		induk, err := s.repo.FindParokiByID(ctx, *indukID)
		if ent.IsNotFound(err) {
			return domain.ErrParokiNotFound
		}
		if err != nil {
			return err
		}
		if induk.KeuskupanID != keuskupanID {
			return domain.ValidationError("indukId", "KEUSKUPAN_MISMATCH", "Induk paroki must belong to the same Keuskupan.")
		}
		if induk.IndukID != nil {
			return domain.ValidationError("indukId", "MAX_DEPTH", "Induk paroki cannot itself be a child paroki.")
		}
		if parokiID != uuid.Nil {
			hasChildren, err := s.repo.HasChildParoki(ctx, parokiID)
			if err != nil {
				return err
			}
			if hasChildren {
				return domain.ValidationError("indukId", "HAS_CHILDREN", "Paroki with child paroki cannot become a child paroki.")
			}
		}
	}

	if kotaID != nil {
		if _, err := s.repo.FindKotaByID(ctx, *kotaID); ent.IsNotFound(err) {
			return domain.ErrKotaNotFound
		} else if err != nil {
			return err
		}
	}
	if kecamatanID != nil {
		kec, err := s.repo.FindKecamatanByID(ctx, *kecamatanID)
		if ent.IsNotFound(err) {
			return domain.ErrKecamatanNotFound
		}
		if err != nil {
			return err
		}
		if kotaID == nil || kec.KotaID != *kotaID {
			return domain.ValidationError("kecamatanId", "KOTA_MISMATCH", "Kecamatan must belong to the selected Kota.")
		}
	}
	if kelurahanID != nil {
		kel, err := s.repo.FindKelurahanByID(ctx, *kelurahanID)
		if ent.IsNotFound(err) {
			return domain.ErrKelurahanNotFound
		}
		if err != nil {
			return err
		}
		if kecamatanID == nil || kel.KecamatanID != *kecamatanID {
			return domain.ValidationError("kelurahanId", "KECAMATAN_MISMATCH", "Kelurahan must belong to the selected Kecamatan.")
		}
	}
	return nil
}

func normalizeTimezone(raw string) (string, error) {
	tz := strings.TrimSpace(raw)
	if tz == "" || strings.EqualFold(tz, "Local") {
		return "", domain.ValidationError("timezone", "TIMEZONE", "Invalid IANA timezone.")
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", domain.ValidationError("timezone", "TIMEZONE", "Invalid IANA timezone.")
	}
	return loc.String(), nil
}

func normalizeOptMaxString(field string, raw *string, maxLen int, keepEmpty bool) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if len(trimmed) > maxLen {
		return nil, domain.ValidationError(field, "MAX_LENGTH", field+" exceeds maximum length.")
	}
	if trimmed == "" && !keepEmpty {
		return nil, nil
	}
	return &trimmed, nil
}

func normalizeOptStatusParoki(raw *string, keepEmpty bool) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		if keepEmpty {
			return &trimmed, nil
		}
		return nil, nil
	}
	norm, err := normalizeKode(trimmed)
	if err != nil {
		return nil, domain.ValidationError("statusParoki", "INVALID_CODE", "Status paroki must be 2-64 lowercase alphanumeric characters, hyphens, or underscores.")
	}
	return &norm, nil
}

func validateCoordinates(lat, lng *float64) error {
	if (lat == nil) != (lng == nil) {
		return domain.ValidationError("latitude", "COORDINATES", "Both latitude and longitude are required.")
	}
	if lat == nil {
		return nil
	}
	if math.IsNaN(*lat) || math.IsInf(*lat, 0) || *lat < -90 || *lat > 90 {
		return domain.ValidationError("latitude", "RANGE", "Latitude must be between -90 and 90.")
	}
	if math.IsNaN(*lng) || math.IsInf(*lng, 0) || *lng < -180 || *lng > 180 {
		return domain.ValidationError("longitude", "RANGE", "Longitude must be between -180 and 180.")
	}
	return nil
}

func resolveUpdateCoordinates(existingLat, existingLng, reqLat, reqLng *float64, clear bool) (*float64, *float64, error) {
	if clear {
		if reqLat != nil || reqLng != nil {
			return nil, nil, domain.ValidationError("clearCoordinates", "CONFLICT", "Cannot set coordinates while clearCoordinates is true.")
		}
		return nil, nil, nil
	}
	effLat := existingLat
	if reqLat != nil {
		effLat = reqLat
	}
	effLng := existingLng
	if reqLng != nil {
		effLng = reqLng
	}
	if err := validateCoordinates(effLat, effLng); err != nil {
		return nil, nil, err
	}
	return effLat, effLng, nil
}

func nonNilUUID(id *uuid.UUID) *uuid.UUID {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	return id
}

func resolveOptUUID(existing, req *uuid.UUID) *uuid.UUID {
	if req == nil {
		return existing
	}
	if *req == uuid.Nil {
		return nil
	}
	return req
}

func cascadeCivilClear(kotaReq, kecReq, kelReq *uuid.UUID) (*uuid.UUID, *uuid.UUID, *uuid.UUID) {
	nilID := uuid.Nil
	if kotaReq != nil && *kotaReq == uuid.Nil {
		if kecReq == nil {
			kecReq = &nilID
		}
		if kelReq == nil {
			kelReq = &nilID
		}
	}
	if kecReq != nil && *kecReq == uuid.Nil {
		if kelReq == nil {
			kelReq = &nilID
		}
	}
	return kotaReq, kecReq, kelReq
}

func toParokiResponse(p *ent.Paroki) domain.ParokiResponse {
	return domain.ParokiResponse{
		ID:           p.ID,
		KeuskupanID:  p.KeuskupanID,
		KevikepanID:  p.KevikepanID,
		IndukID:      p.IndukID,
		Kode:         p.Kode,
		KodeResmi:    p.KodeResmi,
		Nama:         p.Nama,
		Pelindung:    p.Pelindung,
		Pastor:       p.Pastor,
		Alamat:       p.Alamat,
		KotaID:       p.KotaID,
		KecamatanID:  p.KecamatanID,
		KelurahanID:  p.KelurahanID,
		Timezone:     p.Timezone,
		Telepon:      p.Telepon,
		Fax:          p.Fax,
		Email:        p.Email,
		Website:      p.Website,
		LogoURL:      p.LogoURL,
		KopSurat:     p.KopSurat,
		Latitude:     p.Latitude,
		Longitude:    p.Longitude,
		StatusParoki: p.StatusParoki,
		Active:       p.Active,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
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
