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

// SearchKevikepan returns paginated non-deleted kevikepan.
func (s *Service) SearchKevikepan(
	ctx context.Context,
	req domain.SearchKevikepanRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.KevikepanResponse, bool, error) {
	var keuskupanID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		keuskupanID = req.Filter.KeuskupanID
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchKevikepan(ctx, strings.TrimSpace(req.Q), keuskupanID, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.KevikepanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toKevikepanResponse(row))
	}
	return items, hasMore, nil
}

// GetKevikepan returns a single non-deleted Kevikepan by ID.
func (s *Service) GetKevikepan(ctx context.Context, id uuid.UUID) (domain.KevikepanResponse, error) {
	k, err := s.repo.FindKevikepanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KevikepanResponse{}, domain.ErrKevikepanNotFound
	}
	if err != nil {
		return domain.KevikepanResponse{}, err
	}
	return toKevikepanResponse(k), nil
}

// CreateKevikepan creates a Kevikepan or reactivates a soft-deleted one with the same (keuskupan_id, kode).
func (s *Service) CreateKevikepan(ctx context.Context, req domain.CreateKevikepanRequest) (domain.KevikepanResponse, error) {
	if _, err := s.repo.FindKeuskupanByID(ctx, req.KeuskupanID); ent.IsNotFound(err) {
		return domain.KevikepanResponse{}, domain.ErrKeuskupanNotFound
	} else if err != nil {
		return domain.KevikepanResponse{}, err
	}

	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.KevikepanResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.KevikepanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	vikep := trimOptNonEmpty(req.Vikep)
	in := repo.KevikepanInput{
		KeuskupanID: req.KeuskupanID,
		Kode:        kode,
		Nama:        nama,
		Vikep:       vikep,
		Active:      active,
	}

	var out *ent.Kevikepan
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindKevikepanByKeuskupanAndKodeIncludeDeleted(ctx, req.KeuskupanID, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateKevikepan(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateKevikepan(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.KevikepanResponse{}, err
	}
	return toKevikepanResponse(out), nil
}

// UpdateKevikepan updates mutable fields of a non-deleted Kevikepan.
func (s *Service) UpdateKevikepan(ctx context.Context, id uuid.UUID, req domain.UpdateKevikepanRequest) (domain.KevikepanResponse, error) {
	existing, err := s.repo.FindKevikepanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KevikepanResponse{}, domain.ErrKevikepanNotFound
	}
	if err != nil {
		return domain.KevikepanResponse{}, err
	}

	targetKeuskupanID := existing.KeuskupanID
	if req.KeuskupanID != nil && *req.KeuskupanID != existing.KeuskupanID {
		if _, err := s.repo.FindKeuskupanByID(ctx, *req.KeuskupanID); ent.IsNotFound(err) {
			return domain.KevikepanResponse{}, domain.ErrKeuskupanNotFound
		} else if err != nil {
			return domain.KevikepanResponse{}, err
		}
		targetKeuskupanID = *req.KeuskupanID
	}

	targetKode := existing.Kode
	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.KevikepanResponse{}, err
		}
		targetKode = normKode
		kodePtr = &normKode
	}

	if targetKeuskupanID != existing.KeuskupanID || targetKode != existing.Kode {
		if conflict, findErr := s.repo.FindKevikepanByKeuskupanAndKodeIncludeDeleted(ctx, targetKeuskupanID, targetKode); findErr == nil && conflict.ID != existing.ID {
			return domain.KevikepanResponse{}, domain.ErrCodeConflict
		} else if findErr != nil && !ent.IsNotFound(findErr) {
			return domain.KevikepanResponse{}, findErr
		}
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.KevikepanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	updated, err := s.repo.UpdateKevikepan(
		ctx,
		id,
		req.KeuskupanID,
		kodePtr,
		namaPtr,
		trimOptString(req.Vikep),
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KevikepanResponse{}, domain.ErrCodeConflict
		}
		return domain.KevikepanResponse{}, err
	}
	return toKevikepanResponse(updated), nil
}

// DeleteKevikepan soft-deletes a non-deleted Kevikepan if no non-deleted paroki references it.
func (s *Service) DeleteKevikepan(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindKevikepanByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrKevikepanNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasParokiByKevikepan(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteKevikepan(ctx, id, time.Now().UTC())
}

func trimOptNonEmpty(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func toKevikepanResponse(k *ent.Kevikepan) domain.KevikepanResponse {
	return domain.KevikepanResponse{
		ID:          k.ID,
		KeuskupanID: k.KeuskupanID,
		Kode:        k.Kode,
		Nama:        k.Nama,
		Vikep:       k.Vikep,
		Active:      k.Active,
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}
