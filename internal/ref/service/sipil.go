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

const maxKodePosLen = 10

// ValidateKota returns a non-deleted Kota by ID.
func (s *Service) ValidateKota(ctx context.Context, kotaID uuid.UUID) (*domain.KotaResponse, error) {
	k, err := s.repo.FindKotaByID(ctx, kotaID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrKotaNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toKotaResponse(k)
	return &res, nil
}

// ValidateKecamatan returns a non-deleted Kecamatan if it belongs to kotaID.
func (s *Service) ValidateKecamatan(ctx context.Context, kotaID, kecamatanID uuid.UUID) (*domain.KecamatanResponse, error) {
	k, err := s.repo.FindKecamatanInKota(ctx, kotaID, kecamatanID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrKecamatanNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toKecamatanResponse(k)
	return &res, nil
}

// ValidateKelurahan returns a non-deleted Kelurahan if it belongs to kecamatanID.
func (s *Service) ValidateKelurahan(ctx context.Context, kecamatanID, kelurahanID uuid.UUID) (*domain.KelurahanResponse, error) {
	k, err := s.repo.FindKelurahanInKecamatan(ctx, kecamatanID, kelurahanID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrKelurahanNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toKelurahanResponse(k)
	return &res, nil
}

// SearchKota returns paginated non-deleted kota.
func (s *Service) SearchKota(
	ctx context.Context,
	req domain.SearchKotaRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.KotaResponse, bool, error) {
	var active *bool
	if req.Filter != nil {
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchKota(ctx, strings.TrimSpace(req.Q), active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.KotaResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toKotaResponse(row))
	}
	return items, hasMore, nil
}

// GetKota returns a single non-deleted Kota by ID.
func (s *Service) GetKota(ctx context.Context, id uuid.UUID) (domain.KotaResponse, error) {
	k, err := s.repo.FindKotaByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KotaResponse{}, domain.ErrKotaNotFound
	}
	if err != nil {
		return domain.KotaResponse{}, err
	}
	return toKotaResponse(k), nil
}

// CreateKota creates a Kota or reactivates a soft-deleted one with the same kode.
func (s *Service) CreateKota(ctx context.Context, req domain.CreateKotaRequest) (domain.KotaResponse, error) {
	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.KotaResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.KotaResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.KotaInput{
		Kode:   kode,
		Nama:   nama,
		Active: active,
	}

	var out *ent.Kota
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindKotaByKodeIncludeDeleted(ctx, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateKota(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateKota(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.KotaResponse{}, err
	}
	return toKotaResponse(out), nil
}

// UpdateKota updates mutable fields of a non-deleted Kota.
func (s *Service) UpdateKota(ctx context.Context, id uuid.UUID, req domain.UpdateKotaRequest) (domain.KotaResponse, error) {
	existing, err := s.repo.FindKotaByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KotaResponse{}, domain.ErrKotaNotFound
	}
	if err != nil {
		return domain.KotaResponse{}, err
	}

	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.KotaResponse{}, err
		}
		if normKode != existing.Kode {
			if _, findErr := s.repo.FindKotaByKodeIncludeDeleted(ctx, normKode); findErr == nil {
				return domain.KotaResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.KotaResponse{}, findErr
			}
		}
		kodePtr = &normKode
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.KotaResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	updated, err := s.repo.UpdateKota(ctx, id, kodePtr, namaPtr, req.Active)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KotaResponse{}, domain.ErrCodeConflict
		}
		return domain.KotaResponse{}, err
	}
	return toKotaResponse(updated), nil
}

// DeleteKota soft-deletes a Kota if it has no non-deleted kecamatan.
func (s *Service) DeleteKota(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindKotaByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrKotaNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasKecamatanByKota(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteKota(ctx, id, time.Now().UTC())
}

// SearchKecamatan returns paginated non-deleted kecamatan.
func (s *Service) SearchKecamatan(
	ctx context.Context,
	req domain.SearchKecamatanRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.KecamatanResponse, bool, error) {
	var kotaID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		kotaID = req.Filter.KotaID
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchKecamatan(ctx, strings.TrimSpace(req.Q), kotaID, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.KecamatanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toKecamatanResponse(row))
	}
	return items, hasMore, nil
}

// GetKecamatan returns a single non-deleted Kecamatan by ID.
func (s *Service) GetKecamatan(ctx context.Context, id uuid.UUID) (domain.KecamatanResponse, error) {
	k, err := s.repo.FindKecamatanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KecamatanResponse{}, domain.ErrKecamatanNotFound
	}
	if err != nil {
		return domain.KecamatanResponse{}, err
	}
	return toKecamatanResponse(k), nil
}

// CreateKecamatan creates a Kecamatan or reactivates a soft-deleted one with the same (kota_id, kode).
func (s *Service) CreateKecamatan(ctx context.Context, req domain.CreateKecamatanRequest) (domain.KecamatanResponse, error) {
	if _, err := s.repo.FindKotaByID(ctx, req.KotaID); ent.IsNotFound(err) {
		return domain.KecamatanResponse{}, domain.ErrKotaNotFound
	} else if err != nil {
		return domain.KecamatanResponse{}, err
	}

	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.KecamatanResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.KecamatanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.KecamatanInput{
		KotaID: req.KotaID,
		Kode:   kode,
		Nama:   nama,
		Active: active,
	}

	var out *ent.Kecamatan
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindKecamatanByKotaAndKodeIncludeDeleted(ctx, req.KotaID, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateKecamatan(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateKecamatan(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.KecamatanResponse{}, err
	}
	return toKecamatanResponse(out), nil
}

// UpdateKecamatan updates mutable fields of a non-deleted Kecamatan.
func (s *Service) UpdateKecamatan(ctx context.Context, id uuid.UUID, req domain.UpdateKecamatanRequest) (domain.KecamatanResponse, error) {
	existing, err := s.repo.FindKecamatanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KecamatanResponse{}, domain.ErrKecamatanNotFound
	}
	if err != nil {
		return domain.KecamatanResponse{}, err
	}

	targetKotaID := existing.KotaID
	if req.KotaID != nil && *req.KotaID != existing.KotaID {
		if _, err := s.repo.FindKotaByID(ctx, *req.KotaID); ent.IsNotFound(err) {
			return domain.KecamatanResponse{}, domain.ErrKotaNotFound
		} else if err != nil {
			return domain.KecamatanResponse{}, err
		}
		targetKotaID = *req.KotaID
	}

	targetKode := existing.Kode
	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.KecamatanResponse{}, err
		}
		targetKode = normKode
		kodePtr = &normKode
	}

	if targetKotaID != existing.KotaID || targetKode != existing.Kode {
		if conflict, findErr := s.repo.FindKecamatanByKotaAndKodeIncludeDeleted(ctx, targetKotaID, targetKode); findErr == nil && conflict.ID != existing.ID {
			return domain.KecamatanResponse{}, domain.ErrCodeConflict
		} else if findErr != nil && !ent.IsNotFound(findErr) {
			return domain.KecamatanResponse{}, findErr
		}
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.KecamatanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	updated, err := s.repo.UpdateKecamatan(ctx, id, req.KotaID, kodePtr, namaPtr, req.Active)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KecamatanResponse{}, domain.ErrCodeConflict
		}
		return domain.KecamatanResponse{}, err
	}
	return toKecamatanResponse(updated), nil
}

// DeleteKecamatan soft-deletes a Kecamatan if it has no non-deleted kelurahan.
func (s *Service) DeleteKecamatan(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindKecamatanByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrKecamatanNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasKelurahanByKecamatan(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteKecamatan(ctx, id, time.Now().UTC())
}

// SearchKelurahan returns paginated non-deleted kelurahan.
func (s *Service) SearchKelurahan(
	ctx context.Context,
	req domain.SearchKelurahanRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.KelurahanResponse, bool, error) {
	var kecamatanID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		kecamatanID = req.Filter.KecamatanID
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchKelurahan(ctx, strings.TrimSpace(req.Q), kecamatanID, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.KelurahanResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toKelurahanResponse(row))
	}
	return items, hasMore, nil
}

// GetKelurahan returns a single non-deleted Kelurahan by ID.
func (s *Service) GetKelurahan(ctx context.Context, id uuid.UUID) (domain.KelurahanResponse, error) {
	k, err := s.repo.FindKelurahanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KelurahanResponse{}, domain.ErrKelurahanNotFound
	}
	if err != nil {
		return domain.KelurahanResponse{}, err
	}
	return toKelurahanResponse(k), nil
}

// CreateKelurahan creates a Kelurahan or reactivates a soft-deleted one with the same (kecamatan_id, kode).
func (s *Service) CreateKelurahan(ctx context.Context, req domain.CreateKelurahanRequest) (domain.KelurahanResponse, error) {
	if _, err := s.repo.FindKecamatanByID(ctx, req.KecamatanID); ent.IsNotFound(err) {
		return domain.KelurahanResponse{}, domain.ErrKecamatanNotFound
	} else if err != nil {
		return domain.KelurahanResponse{}, err
	}

	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.KelurahanResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.KelurahanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	kodePos, err := normalizeKodePos(req.KodePos, false)
	if err != nil {
		return domain.KelurahanResponse{}, err
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.KelurahanInput{
		KecamatanID: req.KecamatanID,
		Kode:        kode,
		Nama:        nama,
		KodePos:     kodePos,
		Active:      active,
	}

	var out *ent.Kelurahan
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindKelurahanByKecamatanAndKodeIncludeDeleted(ctx, req.KecamatanID, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateKelurahan(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateKelurahan(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.KelurahanResponse{}, err
	}
	return toKelurahanResponse(out), nil
}

// UpdateKelurahan updates mutable fields of a non-deleted Kelurahan.
func (s *Service) UpdateKelurahan(ctx context.Context, id uuid.UUID, req domain.UpdateKelurahanRequest) (domain.KelurahanResponse, error) {
	existing, err := s.repo.FindKelurahanByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.KelurahanResponse{}, domain.ErrKelurahanNotFound
	}
	if err != nil {
		return domain.KelurahanResponse{}, err
	}

	targetKecamatanID := existing.KecamatanID
	if req.KecamatanID != nil && *req.KecamatanID != existing.KecamatanID {
		if _, err := s.repo.FindKecamatanByID(ctx, *req.KecamatanID); ent.IsNotFound(err) {
			return domain.KelurahanResponse{}, domain.ErrKecamatanNotFound
		} else if err != nil {
			return domain.KelurahanResponse{}, err
		}
		targetKecamatanID = *req.KecamatanID
	}

	targetKode := existing.Kode
	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.KelurahanResponse{}, err
		}
		targetKode = normKode
		kodePtr = &normKode
	}

	if targetKecamatanID != existing.KecamatanID || targetKode != existing.Kode {
		if conflict, findErr := s.repo.FindKelurahanByKecamatanAndKodeIncludeDeleted(ctx, targetKecamatanID, targetKode); findErr == nil && conflict.ID != existing.ID {
			return domain.KelurahanResponse{}, domain.ErrCodeConflict
		} else if findErr != nil && !ent.IsNotFound(findErr) {
			return domain.KelurahanResponse{}, findErr
		}
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.KelurahanResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	kodePosPtr, err := normalizeKodePos(req.KodePos, true)
	if err != nil {
		return domain.KelurahanResponse{}, err
	}

	updated, err := s.repo.UpdateKelurahan(
		ctx,
		id,
		req.KecamatanID,
		kodePtr,
		namaPtr,
		kodePosPtr,
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KelurahanResponse{}, domain.ErrCodeConflict
		}
		return domain.KelurahanResponse{}, err
	}
	return toKelurahanResponse(updated), nil
}

// DeleteKelurahan soft-deletes a non-deleted Kelurahan.
func (s *Service) DeleteKelurahan(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindKelurahanByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrKelurahanNotFound
	} else if err != nil {
		return err
	}
	return s.repo.SoftDeleteKelurahan(ctx, id, time.Now().UTC())
}

func normalizeKodePos(raw *string, keepEmpty bool) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if len(trimmed) > maxKodePosLen {
		return nil, domain.ValidationError("kodePos", "MAX_LENGTH", "Kode pos must be at most 10 characters.")
	}
	if trimmed == "" && !keepEmpty {
		return nil, nil
	}
	return &trimmed, nil
}

func toKotaResponse(k *ent.Kota) domain.KotaResponse {
	return domain.KotaResponse{
		ID:        k.ID,
		Kode:      k.Kode,
		Nama:      k.Nama,
		Active:    k.Active,
		CreatedAt: k.CreatedAt,
		UpdatedAt: k.UpdatedAt,
	}
}

func toKecamatanResponse(k *ent.Kecamatan) domain.KecamatanResponse {
	return domain.KecamatanResponse{
		ID:        k.ID,
		KotaID:    k.KotaID,
		Kode:      k.Kode,
		Nama:      k.Nama,
		Active:    k.Active,
		CreatedAt: k.CreatedAt,
		UpdatedAt: k.UpdatedAt,
	}
}

func toKelurahanResponse(k *ent.Kelurahan) domain.KelurahanResponse {
	return domain.KelurahanResponse{
		ID:          k.ID,
		KecamatanID: k.KecamatanID,
		Kode:        k.Kode,
		Nama:        k.Nama,
		KodePos:     k.KodePos,
		Active:      k.Active,
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}
