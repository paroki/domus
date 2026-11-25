package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/wilayah"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchWilayah returns paginated non-deleted wilayah for parokiID.
func (s *Service) SearchWilayah(
	ctx context.Context,
	parokiID uuid.UUID,
	req domain.SearchWilayahRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.WilayahResponse, bool, error) {
	var jenis string
	var active *bool
	if req.Filter != nil {
		jenis = req.Filter.Jenis
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchWilayah(ctx, parokiID, strings.TrimSpace(req.Q), jenis, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.WilayahResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toWilayahResponse(row))
	}
	return items, hasMore, nil
}

// GetWilayah returns a single non-deleted Wilayah in parokiID.
func (s *Service) GetWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID) (domain.WilayahResponse, error) {
	w, err := s.repo.FindWilayahInParoki(ctx, parokiID, wilayahID)
	if ent.IsNotFound(err) {
		return domain.WilayahResponse{}, domain.ErrWilayahNotFound
	}
	if err != nil {
		return domain.WilayahResponse{}, err
	}
	return toWilayahResponse(w), nil
}

// CreateWilayah creates a Wilayah or reactivates a soft-deleted one with the same (paroki_id, kode).
func (s *Service) CreateWilayah(ctx context.Context, parokiID uuid.UUID, req domain.CreateWilayahRequest) (domain.WilayahResponse, error) {
	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.WilayahResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.WilayahResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	phone, err := normalizeOptPhone(req.TeleponKoordinator)
	if err != nil {
		return domain.WilayahResponse{}, err
	}
	wJenis := wilayah.JenisWilayah
	if req.Jenis != "" {
		wJenis = wilayah.Jenis(req.Jenis)
	}
	in := repo.WilayahInput{
		ParokiID:           parokiID,
		Jenis:              wJenis,
		Kode:               kode,
		Nama:               nama,
		Koordinator:        trimOptString(req.Koordinator),
		TeleponKoordinator: phone,
	}

	var out *ent.Wilayah
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindWilayahByKodeIncludeDeleted(ctx, parokiID, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateWilayah(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateWilayah(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.WilayahResponse{}, err
	}
	return toWilayahResponse(out), nil
}

// UpdateWilayah updates mutable fields of a Wilayah in parokiID.
func (s *Service) UpdateWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID, req domain.UpdateWilayahRequest) (domain.WilayahResponse, error) {
	existing, err := s.repo.FindWilayahInParoki(ctx, parokiID, wilayahID)
	if ent.IsNotFound(err) {
		return domain.WilayahResponse{}, domain.ErrWilayahNotFound
	}
	if err != nil {
		return domain.WilayahResponse{}, err
	}

	var jenisPtr *wilayah.Jenis
	if req.Jenis != nil {
		j := wilayah.Jenis(*req.Jenis)
		jenisPtr = &j
	}
	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.WilayahResponse{}, err
		}
		if normKode != existing.Kode {
			if _, findErr := s.repo.FindWilayahByKodeIncludeDeleted(ctx, parokiID, normKode); findErr == nil {
				return domain.WilayahResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.WilayahResponse{}, findErr
			}
		}
		kodePtr = &normKode
	}
	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.WilayahResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}
	phonePtr, err := normalizeOptPhone(req.TeleponKoordinator)
	if err != nil {
		return domain.WilayahResponse{}, err
	}

	updated, err := s.repo.UpdateWilayah(
		ctx,
		wilayahID,
		jenisPtr,
		kodePtr,
		namaPtr,
		trimOptString(req.Koordinator),
		phonePtr,
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.WilayahResponse{}, domain.ErrCodeConflict
		}
		return domain.WilayahResponse{}, err
	}
	return toWilayahResponse(updated), nil
}

// DeleteWilayah soft-deletes a Wilayah if no non-deleted Lingkungan or Organizations reference it.
func (s *Service) DeleteWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID) error {
	if _, err := s.repo.FindWilayahInParoki(ctx, parokiID, wilayahID); ent.IsNotFound(err) {
		return domain.ErrWilayahNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasLingkunganByWilayah(ctx, parokiID, wilayahID)
	if err != nil {
		return err
	}
	if !inUse {
		inUse, err = s.repo.HasOrganizationsByWilayah(ctx, parokiID, wilayahID)
		if err != nil {
			return err
		}
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteWilayah(ctx, wilayahID, time.Now().UTC())
}

// SearchLingkungan returns paginated non-deleted lingkungan for parokiID.
func (s *Service) SearchLingkungan(
	ctx context.Context,
	parokiID uuid.UUID,
	req domain.SearchLingkunganRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.LingkunganResponse, bool, error) {
	var wilayahID *uuid.UUID
	var active *bool
	if req.Filter != nil {
		wilayahID = req.Filter.WilayahID
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchLingkungan(ctx, parokiID, strings.TrimSpace(req.Q), wilayahID, active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.LingkunganResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toLingkunganResponse(row))
	}
	return items, hasMore, nil
}

// GetLingkungan returns a single non-deleted Lingkungan in parokiID.
func (s *Service) GetLingkungan(ctx context.Context, parokiID, lingkunganID uuid.UUID) (domain.LingkunganResponse, error) {
	l, err := s.repo.FindLingkunganInParoki(ctx, parokiID, lingkunganID)
	if ent.IsNotFound(err) {
		return domain.LingkunganResponse{}, domain.ErrLingkunganNotFound
	}
	if err != nil {
		return domain.LingkunganResponse{}, err
	}
	return toLingkunganResponse(l), nil
}

// CreateLingkungan creates a Lingkungan or reactivates a soft-deleted one with the same (paroki_id, kode).
func (s *Service) CreateLingkungan(ctx context.Context, parokiID uuid.UUID, req domain.CreateLingkunganRequest) (domain.LingkunganResponse, error) {
	if _, err := s.repo.FindWilayahInParoki(ctx, parokiID, req.WilayahID); ent.IsNotFound(err) {
		return domain.LingkunganResponse{}, domain.ErrWilayahNotFound
	} else if err != nil {
		return domain.LingkunganResponse{}, err
	}

	kode, err := normalizeKode(req.Kode)
	if err != nil {
		return domain.LingkunganResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.LingkunganResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	phone, err := normalizeOptPhone(req.TeleponKetua)
	if err != nil {
		return domain.LingkunganResponse{}, err
	}
	in := repo.LingkunganInput{
		ParokiID:     parokiID,
		WilayahID:    req.WilayahID,
		Kode:         kode,
		Nama:         nama,
		Pelindung:    trimOptString(req.Pelindung),
		Ketua:        trimOptString(req.Ketua),
		TeleponKetua: phone,
	}

	var out *ent.Lingkungan
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindLingkunganByKodeIncludeDeleted(ctx, parokiID, kode)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateLingkungan(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateLingkungan(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.LingkunganResponse{}, err
	}
	return toLingkunganResponse(out), nil
}

// UpdateLingkungan updates mutable fields of a Lingkungan in parokiID.
func (s *Service) UpdateLingkungan(ctx context.Context, parokiID, lingkunganID uuid.UUID, req domain.UpdateLingkunganRequest) (domain.LingkunganResponse, error) {
	existing, err := s.repo.FindLingkunganInParoki(ctx, parokiID, lingkunganID)
	if ent.IsNotFound(err) {
		return domain.LingkunganResponse{}, domain.ErrLingkunganNotFound
	}
	if err != nil {
		return domain.LingkunganResponse{}, err
	}

	if req.WilayahID != nil {
		if _, err := s.repo.FindWilayahInParoki(ctx, parokiID, *req.WilayahID); ent.IsNotFound(err) {
			return domain.LingkunganResponse{}, domain.ErrWilayahNotFound
		} else if err != nil {
			return domain.LingkunganResponse{}, err
		}
	}

	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeKode(*req.Kode)
		if err != nil {
			return domain.LingkunganResponse{}, err
		}
		if normKode != existing.Kode {
			if _, findErr := s.repo.FindLingkunganByKodeIncludeDeleted(ctx, parokiID, normKode); findErr == nil {
				return domain.LingkunganResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.LingkunganResponse{}, findErr
			}
		}
		kodePtr = &normKode
	}
	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.LingkunganResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}
	phonePtr, err := normalizeOptPhone(req.TeleponKetua)
	if err != nil {
		return domain.LingkunganResponse{}, err
	}

	updated, err := s.repo.UpdateLingkungan(
		ctx,
		lingkunganID,
		req.WilayahID,
		kodePtr,
		namaPtr,
		trimOptString(req.Pelindung),
		trimOptString(req.Ketua),
		phonePtr,
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.LingkunganResponse{}, domain.ErrCodeConflict
		}
		return domain.LingkunganResponse{}, err
	}
	return toLingkunganResponse(updated), nil
}

// DeleteLingkungan soft-deletes a Lingkungan in parokiID if no non-deleted Organizations reference it.
func (s *Service) DeleteLingkungan(ctx context.Context, parokiID, lingkunganID uuid.UUID) error {
	if _, err := s.repo.FindLingkunganInParoki(ctx, parokiID, lingkunganID); ent.IsNotFound(err) {
		return domain.ErrLingkunganNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasOrganizationsByLingkungan(ctx, parokiID, lingkunganID)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteLingkungan(ctx, lingkunganID, time.Now().UTC())
}
