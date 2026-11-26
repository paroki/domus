package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgperiode"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

func (s *Service) requireOrgPeriodeRead(ctx context.Context, p authx.Principal, orgID uuid.UUID) (*ent.Organization, error) {
	orgRow, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrOrgNotFound
	}
	if err != nil {
		return nil, err
	}
	isAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return nil, err
	}
	isMem, err := s.isOrgMember(p, orgID)
	if err != nil {
		return nil, err
	}
	isParokiAdm, err := s.isParokiAdmin(p, orgRow.ParokiID)
	if err != nil {
		return nil, err
	}
	if !isAdm && !isMem && !isParokiAdm {
		return nil, domain.ErrForbidden
	}
	return orgRow, nil
}

func (s *Service) requireOrgPeriodeWrite(ctx context.Context, p authx.Principal, orgID uuid.UUID) (*ent.Organization, error) {
	orgRow, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrOrgNotFound
	}
	if err != nil {
		return nil, err
	}
	isOrgAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return nil, err
	}
	isParokiAdm, err := s.isParokiAdmin(p, orgRow.ParokiID)
	if err != nil {
		return nil, err
	}
	if !isOrgAdm && !isParokiAdm {
		return nil, domain.ErrForbidden
	}
	return orgRow, nil
}

// SearchOrgPeriode returns paginated non-deleted OrgPeriode rows for orgID.
func (s *Service) SearchOrgPeriode(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.SearchOrgPeriodeRequest,
	cursorTanggalMulai *string,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.OrgPeriodeResponse, bool, error) {
	if _, err := s.requireOrgPeriodeRead(ctx, p, orgID); err != nil {
		return nil, false, err
	}

	var status string
	var isCurrent *bool
	if req.Filter != nil {
		status = req.Filter.Status
		isCurrent = req.Filter.IsCurrent
	}
	rows, err := s.repo.SearchOrgPeriode(ctx, orgID, strings.TrimSpace(req.Q), status, isCurrent, cursorTanggalMulai, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.OrgPeriodeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOrgPeriodeResponse(row))
	}
	return items, hasMore, nil
}

// GetOrgPeriode returns a single non-deleted OrgPeriode in orgID.
func (s *Service) GetOrgPeriode(ctx context.Context, p authx.Principal, orgID, periodeID uuid.UUID) (domain.OrgPeriodeResponse, error) {
	if _, err := s.requireOrgPeriodeRead(ctx, p, orgID); err != nil {
		return domain.OrgPeriodeResponse{}, err
	}
	row, err := s.repo.FindOrgPeriodeInOrg(ctx, orgID, periodeID)
	if ent.IsNotFound(err) {
		return domain.OrgPeriodeResponse{}, domain.ErrPeriodeNotFound
	}
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}
	return toOrgPeriodeResponse(row), nil
}

// CreateOrgPeriode creates an OrgPeriode (or reactivates a soft-deleted match) with overlap validation and atomic isCurrent rotation.
func (s *Service) CreateOrgPeriode(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.CreateOrgPeriodeRequest,
) (domain.OrgPeriodeResponse, error) {
	orgRow, err := s.requireOrgPeriodeWrite(ctx, p, orgID)
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}

	kode, err := normalizeCode(req.Kode)
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.OrgPeriodeResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
	}
	tanggalMulai, tanggalSelesai, err := validatePeriodeDateRange(req.TanggalMulai, req.TanggalSelesai)
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}
	status := orgperiode.StatusOpen
	if req.Status != "" {
		status = orgperiode.Status(req.Status)
	}

	in := repo.OrgPeriodeInput{
		ParokiID:       orgRow.ParokiID,
		OrgID:          orgID,
		Kode:           kode,
		Nama:           nama,
		TanggalMulai:   tanggalMulai,
		TanggalSelesai: tanggalSelesai,
		Status:         status,
		IsCurrent:      req.IsCurrent,
	}

	var out *ent.OrgPeriode
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindOrgPeriodeByKodeIncludeDeleted(ctx, orgID, kode)
		var excludeID *uuid.UUID
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			id := existing.ID
			excludeID = &id
		case !ent.IsNotFound(findErr):
			return findErr
		}

		overlaps, overlapErr := txRepo.HasOverlappingOrgPeriode(ctx, orgID, tanggalMulai, tanggalSelesai, excludeID)
		if overlapErr != nil {
			return overlapErr
		}
		if overlaps {
			return domain.ErrPeriodeOverlap
		}

		if in.IsCurrent {
			if err := txRepo.ClearCurrentOrgPeriode(ctx, orgID, excludeID); err != nil {
				return err
			}
		}

		if excludeID != nil {
			out, findErr = txRepo.ReactivateOrgPeriode(ctx, *excludeID, in)
			return findErr
		}
		out, findErr = txRepo.CreateOrgPeriode(ctx, in)
		if findErr != nil && ent.IsConstraintError(findErr) {
			return domain.ErrCodeConflict
		}
		return findErr
	})
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}
	return toOrgPeriodeResponse(out), nil
}

// UpdateOrgPeriode updates mutable fields of an OrgPeriode with overlap validation and atomic isCurrent rotation.
func (s *Service) UpdateOrgPeriode(
	ctx context.Context,
	p authx.Principal,
	orgID, periodeID uuid.UUID,
	req domain.UpdateOrgPeriodeRequest,
) (domain.OrgPeriodeResponse, error) {
	if _, err := s.requireOrgPeriodeWrite(ctx, p, orgID); err != nil {
		return domain.OrgPeriodeResponse{}, err
	}

	existing, err := s.repo.FindOrgPeriodeInOrg(ctx, orgID, periodeID)
	if ent.IsNotFound(err) {
		return domain.OrgPeriodeResponse{}, domain.ErrPeriodeNotFound
	}
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}

	var kodePtr *string
	if req.Kode != nil {
		normKode, err := normalizeCode(*req.Kode)
		if err != nil {
			return domain.OrgPeriodeResponse{}, err
		}
		if normKode != existing.Kode {
			if _, findErr := s.repo.FindOrgPeriodeByKodeIncludeDeleted(ctx, orgID, normKode); findErr == nil {
				return domain.OrgPeriodeResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.OrgPeriodeResponse{}, findErr
			}
		}
		kodePtr = &normKode
	}

	var namaPtr *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.OrgPeriodeResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		namaPtr = &trimmed
	}

	effStart := existing.TanggalMulai
	if req.TanggalMulai != nil {
		effStart = *req.TanggalMulai
	}
	effEnd := existing.TanggalSelesai
	if req.TanggalSelesai != nil {
		effEnd = *req.TanggalSelesai
	}
	normStart, normEnd, err := validatePeriodeDateRange(effStart, effEnd)
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}

	var startPtr, endPtr *string
	if req.TanggalMulai != nil {
		startPtr = &normStart
	}
	if req.TanggalSelesai != nil {
		endPtr = &normEnd
	}

	var statusPtr *orgperiode.Status
	if req.Status != nil {
		st := orgperiode.Status(*req.Status)
		statusPtr = &st
	}

	var updated *ent.OrgPeriode
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if req.TanggalMulai != nil || req.TanggalSelesai != nil {
			overlaps, overlapErr := txRepo.HasOverlappingOrgPeriode(ctx, orgID, normStart, normEnd, &periodeID)
			if overlapErr != nil {
				return overlapErr
			}
			if overlaps {
				return domain.ErrPeriodeOverlap
			}
		}
		if req.IsCurrent != nil && *req.IsCurrent {
			if err := txRepo.ClearCurrentOrgPeriode(ctx, orgID, &periodeID); err != nil {
				return err
			}
		}
		var txErr error
		updated, txErr = txRepo.UpdateOrgPeriode(ctx, periodeID, kodePtr, namaPtr, startPtr, endPtr, statusPtr, req.IsCurrent)
		if txErr != nil && ent.IsConstraintError(txErr) {
			return domain.ErrCodeConflict
		}
		return txErr
	})
	if err != nil {
		return domain.OrgPeriodeResponse{}, err
	}
	return toOrgPeriodeResponse(updated), nil
}

// DeleteOrgPeriode soft-deletes an OrgPeriode in orgID and clears is_current.
func (s *Service) DeleteOrgPeriode(ctx context.Context, p authx.Principal, orgID, periodeID uuid.UUID) error {
	if _, err := s.requireOrgPeriodeWrite(ctx, p, orgID); err != nil {
		return err
	}
	if _, err := s.repo.FindOrgPeriodeInOrg(ctx, orgID, periodeID); ent.IsNotFound(err) {
		return domain.ErrPeriodeNotFound
	} else if err != nil {
		return err
	}
	return s.repo.SoftDeleteOrgPeriode(ctx, periodeID, time.Now().UTC())
}

func validatePeriodeDateRange(rawStart, rawEnd string) (string, string, error) {
	s := strings.TrimSpace(rawStart)
	e := strings.TrimSpace(rawEnd)
	startTime, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", "", domain.ValidationError("tanggalMulai", "INVALID_DATE", "tanggalMulai must be formatted as YYYY-MM-DD.")
	}
	endTime, err := time.Parse("2006-01-02", e)
	if err != nil {
		return "", "", domain.ValidationError("tanggalSelesai", "INVALID_DATE", "tanggalSelesai must be formatted as YYYY-MM-DD.")
	}
	if startTime.After(endTime) {
		return "", "", domain.ValidationError("tanggalSelesai", "INVALID_DATE_RANGE", "tanggalSelesai must be on or after tanggalMulai.")
	}
	return s, e, nil
}

func toOrgPeriodeResponse(p *ent.OrgPeriode) domain.OrgPeriodeResponse {
	return domain.OrgPeriodeResponse{
		ID:             p.ID,
		ParokiID:       p.ParokiID,
		OrgID:          p.OrgID,
		Kode:           p.Kode,
		Nama:           p.Nama,
		TanggalMulai:   p.TanggalMulai,
		TanggalSelesai: p.TanggalSelesai,
		Status:         string(p.Status),
		IsCurrent:      p.IsCurrent,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
