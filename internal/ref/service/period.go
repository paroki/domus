package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchPeriods returns paginated non-deleted periods for parokiID.
func (s *Service) SearchPeriods(
	ctx context.Context,
	parokiID uuid.UUID,
	req domain.SearchPeriodsRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.PeriodResponse, bool, error) {
	var pType, status string
	var isCurrent *bool
	if req.Filter != nil {
		pType = req.Filter.Type
		status = req.Filter.Status
		isCurrent = req.Filter.IsCurrent
	}
	rows, err := s.repo.SearchPeriods(ctx, parokiID, strings.TrimSpace(req.Q), pType, status, isCurrent, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.PeriodResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPeriodResponse(row))
	}
	return items, hasMore, nil
}

// GetPeriod returns a single non-deleted Period in parokiID.
func (s *Service) GetPeriod(ctx context.Context, parokiID, periodID uuid.UUID) (domain.PeriodResponse, error) {
	row, err := s.repo.FindPeriodInParoki(ctx, parokiID, periodID)
	if ent.IsNotFound(err) {
		return domain.PeriodResponse{}, domain.ErrPeriodNotFound
	}
	if err != nil {
		return domain.PeriodResponse{}, err
	}
	return toPeriodResponse(row), nil
}

// CreatePeriod creates a Period (or reactivates a soft-deleted match) with date overlap and is_current rotation.
func (s *Service) CreatePeriod(ctx context.Context, parokiID uuid.UUID, req domain.CreatePeriodRequest) (domain.PeriodResponse, error) {
	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.PeriodResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.PeriodResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
	}
	startDate, endDate, err := validateDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return domain.PeriodResponse{}, err
	}
	pType := period.TypeFiscal
	if req.Type != "" {
		pType = period.Type(req.Type)
	}

	in := repo.PeriodInput{
		ParokiID:  parokiID,
		Type:      pType,
		Code:      code,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		IsCurrent: req.IsCurrent,
	}

	var out *ent.Period
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindPeriodByCodeIncludeDeleted(ctx, parokiID, code)
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

		overlaps, overlapErr := txRepo.HasOverlappingPeriod(ctx, parokiID, pType, startDate, endDate, excludeID)
		if overlapErr != nil {
			return overlapErr
		}
		if overlaps {
			return domain.ErrPeriodOverlap
		}

		if in.IsCurrent {
			if err := txRepo.ClearCurrentPeriod(ctx, parokiID, pType, excludeID); err != nil {
				return err
			}
		}

		if excludeID != nil {
			out, findErr = txRepo.ReactivatePeriod(ctx, *excludeID, in)
			return findErr
		}
		out, findErr = txRepo.CreatePeriod(ctx, in)
		if findErr != nil && ent.IsConstraintError(findErr) {
			return domain.ErrCodeConflict
		}
		return findErr
	})
	if err != nil {
		return domain.PeriodResponse{}, err
	}
	return toPeriodResponse(out), nil
}

// UpdatePeriod updates mutable fields of a Period with date overlap and is_current rotation.
func (s *Service) UpdatePeriod(ctx context.Context, parokiID, periodID uuid.UUID, req domain.UpdatePeriodRequest) (domain.PeriodResponse, error) {
	existing, err := s.repo.FindPeriodInParoki(ctx, parokiID, periodID)
	if ent.IsNotFound(err) {
		return domain.PeriodResponse{}, domain.ErrPeriodNotFound
	}
	if err != nil {
		return domain.PeriodResponse{}, err
	}

	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.PeriodResponse{}, err
		}
		if normCode != existing.Code {
			if _, findErr := s.repo.FindPeriodByCodeIncludeDeleted(ctx, parokiID, normCode); findErr == nil {
				return domain.PeriodResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.PeriodResponse{}, findErr
			}
		}
		codePtr = &normCode
	}

	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.PeriodResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}

	effStart := existing.StartDate
	if req.StartDate != nil {
		effStart = *req.StartDate
	}
	effEnd := existing.EndDate
	if req.EndDate != nil {
		effEnd = *req.EndDate
	}
	normStart, normEnd, err := validateDateRange(effStart, effEnd)
	if err != nil {
		return domain.PeriodResponse{}, err
	}

	var startPtr, endPtr *string
	if req.StartDate != nil {
		startPtr = &normStart
	}
	if req.EndDate != nil {
		endPtr = &normEnd
	}

	var statusPtr *period.Status
	if req.Status != nil {
		st := period.Status(*req.Status)
		statusPtr = &st
	}

	var updated *ent.Period
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if req.StartDate != nil || req.EndDate != nil {
			overlaps, overlapErr := txRepo.HasOverlappingPeriod(ctx, parokiID, existing.Type, normStart, normEnd, &periodID)
			if overlapErr != nil {
				return overlapErr
			}
			if overlaps {
				return domain.ErrPeriodOverlap
			}
		}
		if req.IsCurrent != nil && *req.IsCurrent {
			if err := txRepo.ClearCurrentPeriod(ctx, parokiID, existing.Type, &periodID); err != nil {
				return err
			}
		}
		var txErr error
		updated, txErr = txRepo.UpdatePeriod(ctx, periodID, codePtr, namePtr, startPtr, endPtr, statusPtr, req.IsCurrent)
		if txErr != nil && ent.IsConstraintError(txErr) {
			return domain.ErrCodeConflict
		}
		return txErr
	})
	if err != nil {
		return domain.PeriodResponse{}, err
	}
	return toPeriodResponse(updated), nil
}

// DeletePeriod soft-deletes a Period in parokiID.
func (s *Service) DeletePeriod(ctx context.Context, parokiID, periodID uuid.UUID) error {
	if _, err := s.repo.FindPeriodInParoki(ctx, parokiID, periodID); ent.IsNotFound(err) {
		return domain.ErrPeriodNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasOrgMembersByPeriod(ctx, parokiID, periodID)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeletePeriod(ctx, periodID, time.Now().UTC())
}

func validateDateRange(rawStart, rawEnd string) (string, string, error) {
	s := strings.TrimSpace(rawStart)
	e := strings.TrimSpace(rawEnd)
	startTime, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", "", domain.ValidationError("startDate", "INVALID_DATE", "startDate must be formatted as YYYY-MM-DD.")
	}
	endTime, err := time.Parse("2006-01-02", e)
	if err != nil {
		return "", "", domain.ValidationError("endDate", "INVALID_DATE", "endDate must be formatted as YYYY-MM-DD.")
	}
	if startTime.After(endTime) {
		return "", "", domain.ValidationError("endDate", "INVALID_DATE_RANGE", "endDate must be on or after startDate.")
	}
	return s, e, nil
}
