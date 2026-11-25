package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchKeuskupan returns paginated non-deleted keuskupan.
func (s *Service) SearchKeuskupan(
	ctx context.Context,
	req domain.SearchKeuskupanRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.KeuskupanResponse, bool, error) {
	var active *bool
	if req.Filter != nil {
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchKeuskupan(ctx, strings.TrimSpace(req.Q), active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.KeuskupanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toKeuskupanResponse(row))
	}
	return items, hasMore, nil
}

// GetKeuskupan returns a single non-deleted Keuskupan by ID.
func (s *Service) GetKeuskupan(ctx context.Context, id uuid.UUID) (domain.KeuskupanResponse, error) {
	d, err := s.repo.FindKeuskupanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KeuskupanResponse{}, domain.ErrKeuskupanNotFound
	}
	if err != nil {
		return domain.KeuskupanResponse{}, err
	}
	return toKeuskupanResponse(d), nil
}

// CreateKeuskupan creates a Keuskupan or reactivates a soft-deleted one with the same kode.
func (s *Service) CreateKeuskupan(ctx context.Context, req domain.CreateKeuskupanRequest) (domain.KeuskupanResponse, error) {
	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.KeuskupanResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.KeuskupanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	telepon, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.KeuskupanResponse{}, err
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.KeuskupanInput{
		Kode:    kode,
		Nama:    nama,
		Uskup:   trimOptString(req.Uskup),
		Alamat:  trimOptString(req.Alamat),
		Telepon: telepon,
		Email:   normalizeOptEmail(req.Email),
		Website: trimOptString(req.Website),
		Active:  active,
	}

	var out *ent.Keuskupan
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindKeuskupanByKodeIncludeDeleted(ctx, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateKeuskupan(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateKeuskupan(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.KeuskupanResponse{}, err
	}
	return toKeuskupanResponse(out), nil
}

// UpdateKeuskupan updates mutable fields of a non-deleted Keuskupan.
func (s *Service) UpdateKeuskupan(ctx context.Context, id uuid.UUID, req domain.UpdateKeuskupanRequest) (domain.KeuskupanResponse, error) {
	existing, err := s.repo.FindKeuskupanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KeuskupanResponse{}, domain.ErrKeuskupanNotFound
	}
	if err != nil {
		return domain.KeuskupanResponse{}, err
	}

	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.KeuskupanResponse{}, err
		}
		if normKode != existing.Kode {
			if _, findErr := s.repo.FindKeuskupanByKodeIncludeDeleted(ctx, normKode); findErr == nil {
				return domain.KeuskupanResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.KeuskupanResponse{}, findErr
			}
		}
		kodePtr = &normKode
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.KeuskupanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	teleponPtr, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.KeuskupanResponse{}, err
	}

	updated, err := s.repo.UpdateKeuskupan(
		ctx,
		id,
		kodePtr,
		namaPtr,
		trimOptString(req.Uskup),
		trimOptString(req.Alamat),
		teleponPtr,
		normalizeOptEmail(req.Email),
		trimOptString(req.Website),
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KeuskupanResponse{}, domain.ErrCodeConflict
		}
		return domain.KeuskupanResponse{}, err
	}
	return toKeuskupanResponse(updated), nil
}

// DeleteKeuskupan soft-deletes a Keuskupan if it has no non-deleted kevikepan or paroki.
func (s *Service) DeleteKeuskupan(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindKeuskupanByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrKeuskupanNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasKevikepanByKeuskupan(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	inUse, err = s.repo.HasParokiByKeuskupan(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteKeuskupan(ctx, id, time.Now().UTC())
}

func toKeuskupanResponse(d *ent.Keuskupan) domain.KeuskupanResponse {
	return domain.KeuskupanResponse{
		ID:        d.ID,
		Kode:      d.Kode,
		Nama:      d.Nama,
		Uskup:     d.Uskup,
		Alamat:    d.Alamat,
		Telepon:   d.Telepon,
		Email:     d.Email,
		Website:   d.Website,
		Active:    d.Active,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}
