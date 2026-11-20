package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/zone"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchZones returns paginated non-deleted zones for parishID.
func (s *Service) SearchZones(
	ctx context.Context,
	parishID uuid.UUID,
	req domain.SearchZonesRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.ZoneResponse, bool, error) {
	var zoneType string
	var active *bool
	if req.Filter != nil {
		zoneType = req.Filter.Type
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchZones(ctx, parishID, strings.TrimSpace(req.Q), zoneType, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.ZoneResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toZoneResponse(row))
	}
	return items, hasMore, nil
}

// GetZone returns a single non-deleted Zone in parishID.
func (s *Service) GetZone(ctx context.Context, parishID, zoneID uuid.UUID) (domain.ZoneResponse, error) {
	z, err := s.repo.FindZoneInParish(ctx, parishID, zoneID)
	if ent.IsNotFound(err) {
		return domain.ZoneResponse{}, domain.ErrZoneNotFound
	}
	if err != nil {
		return domain.ZoneResponse{}, err
	}
	return toZoneResponse(z), nil
}

// CreateZone creates a Zone or reactivates a soft-deleted one with the same (parish_id, code).
func (s *Service) CreateZone(ctx context.Context, parishID uuid.UUID, req domain.CreateZoneRequest) (domain.ZoneResponse, error) {
	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.ZoneResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.ZoneResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
	}
	phone, err := normalizeOptPhone(req.CoordinatorPhone)
	if err != nil {
		return domain.ZoneResponse{}, err
	}
	zType := zone.TypeZone
	if req.Type != "" {
		zType = zone.Type(req.Type)
	}
	in := repo.ZoneInput{
		ParishID:         parishID,
		Type:             zType,
		Code:             code,
		Name:             name,
		CoordinatorName:  trimOptString(req.CoordinatorName),
		CoordinatorPhone: phone,
	}

	var out *ent.Zone
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindZoneByCodeIncludeDeleted(ctx, parishID, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateZone(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateZone(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.ZoneResponse{}, err
	}
	return toZoneResponse(out), nil
}

// UpdateZone updates mutable fields of a Zone in parishID.
func (s *Service) UpdateZone(ctx context.Context, parishID, zoneID uuid.UUID, req domain.UpdateZoneRequest) (domain.ZoneResponse, error) {
	existing, err := s.repo.FindZoneInParish(ctx, parishID, zoneID)
	if ent.IsNotFound(err) {
		return domain.ZoneResponse{}, domain.ErrZoneNotFound
	}
	if err != nil {
		return domain.ZoneResponse{}, err
	}

	var typePtr *zone.Type
	if req.Type != nil {
		t := zone.Type(*req.Type)
		typePtr = &t
	}
	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.ZoneResponse{}, err
		}
		if normCode != existing.Code {
			if _, findErr := s.repo.FindZoneByCodeIncludeDeleted(ctx, parishID, normCode); findErr == nil {
				return domain.ZoneResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.ZoneResponse{}, findErr
			}
		}
		codePtr = &normCode
	}
	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.ZoneResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}
	phonePtr, err := normalizeOptPhone(req.CoordinatorPhone)
	if err != nil {
		return domain.ZoneResponse{}, err
	}

	updated, err := s.repo.UpdateZone(
		ctx,
		zoneID,
		typePtr,
		codePtr,
		namePtr,
		trimOptString(req.CoordinatorName),
		phonePtr,
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.ZoneResponse{}, domain.ErrCodeConflict
		}
		return domain.ZoneResponse{}, err
	}
	return toZoneResponse(updated), nil
}

// DeleteZone soft-deletes a Zone if no non-deleted Wards or Organizations reference it.
func (s *Service) DeleteZone(ctx context.Context, parishID, zoneID uuid.UUID) error {
	if _, err := s.repo.FindZoneInParish(ctx, parishID, zoneID); ent.IsNotFound(err) {
		return domain.ErrZoneNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasWardsByZone(ctx, parishID, zoneID)
	if err != nil {
		return err
	}
	if !inUse {
		inUse, err = s.repo.HasOrganizationsByZone(ctx, parishID, zoneID)
		if err != nil {
			return err
		}
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteZone(ctx, zoneID, time.Now().UTC())
}

// SearchWards returns paginated non-deleted wards for parishID.
func (s *Service) SearchWards(
	ctx context.Context,
	parishID uuid.UUID,
	req domain.SearchWardsRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.WardResponse, bool, error) {
	var zoneID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		zoneID = req.Filter.ZoneID
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchWards(ctx, parishID, strings.TrimSpace(req.Q), zoneID, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.WardResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toWardResponse(row))
	}
	return items, hasMore, nil
}

// GetWard returns a single non-deleted Ward in parishID.
func (s *Service) GetWard(ctx context.Context, parishID, wardID uuid.UUID) (domain.WardResponse, error) {
	w, err := s.repo.FindWardInParish(ctx, parishID, wardID)
	if ent.IsNotFound(err) {
		return domain.WardResponse{}, domain.ErrWardNotFound
	}
	if err != nil {
		return domain.WardResponse{}, err
	}
	return toWardResponse(w), nil
}

// CreateWard creates a Ward or reactivates a soft-deleted one with the same (parish_id, code).
func (s *Service) CreateWard(ctx context.Context, parishID uuid.UUID, req domain.CreateWardRequest) (domain.WardResponse, error) {
	if _, err := s.repo.FindZoneInParish(ctx, parishID, req.ZoneID); ent.IsNotFound(err) {
		return domain.WardResponse{}, domain.ErrZoneNotFound
	} else if err != nil {
		return domain.WardResponse{}, err
	}

	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.WardResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.WardResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
	}
	phone, err := normalizeOptPhone(req.HeadPhone)
	if err != nil {
		return domain.WardResponse{}, err
	}
	in := repo.WardInput{
		ParishID:    parishID,
		ZoneID:      req.ZoneID,
		Code:        code,
		Name:        name,
		PatronSaint: trimOptString(req.PatronSaint),
		HeadName:    trimOptString(req.HeadName),
		HeadPhone:   phone,
	}

	var out *ent.Ward
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindWardByCodeIncludeDeleted(ctx, parishID, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateWard(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateWard(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.WardResponse{}, err
	}
	return toWardResponse(out), nil
}

// UpdateWard updates mutable fields of a Ward in parishID.
func (s *Service) UpdateWard(ctx context.Context, parishID, wardID uuid.UUID, req domain.UpdateWardRequest) (domain.WardResponse, error) {
	existing, err := s.repo.FindWardInParish(ctx, parishID, wardID)
	if ent.IsNotFound(err) {
		return domain.WardResponse{}, domain.ErrWardNotFound
	}
	if err != nil {
		return domain.WardResponse{}, err
	}

	if req.ZoneID != nil {
		if _, err := s.repo.FindZoneInParish(ctx, parishID, *req.ZoneID); ent.IsNotFound(err) {
			return domain.WardResponse{}, domain.ErrZoneNotFound
		} else if err != nil {
			return domain.WardResponse{}, err
		}
	}

	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.WardResponse{}, err
		}
		if normCode != existing.Code {
			if _, findErr := s.repo.FindWardByCodeIncludeDeleted(ctx, parishID, normCode); findErr == nil {
				return domain.WardResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.WardResponse{}, findErr
			}
		}
		codePtr = &normCode
	}
	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.WardResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}
	phonePtr, err := normalizeOptPhone(req.HeadPhone)
	if err != nil {
		return domain.WardResponse{}, err
	}

	updated, err := s.repo.UpdateWard(
		ctx,
		wardID,
		req.ZoneID,
		codePtr,
		namePtr,
		trimOptString(req.PatronSaint),
		trimOptString(req.HeadName),
		phonePtr,
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.WardResponse{}, domain.ErrCodeConflict
		}
		return domain.WardResponse{}, err
	}
	return toWardResponse(updated), nil
}

// DeleteWard soft-deletes a Ward in parishID if no non-deleted Organizations reference it.
func (s *Service) DeleteWard(ctx context.Context, parishID, wardID uuid.UUID) error {
	if _, err := s.repo.FindWardInParish(ctx, parishID, wardID); ent.IsNotFound(err) {
		return domain.ErrWardNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasOrganizationsByWard(ctx, parishID, wardID)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteWard(ctx, wardID, time.Now().UTC())
}
