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

// SearchDioceses returns paginated non-deleted dioceses.
func (s *Service) SearchDioceses(
	ctx context.Context,
	req domain.SearchDiocesesRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.DioceseResponse, bool, error) {
	var active *bool
	if req.Filter != nil {
		active = req.Filter.Active
	}
	rows, err := s.repo.SearchDioceses(ctx, strings.TrimSpace(req.Q), active, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.DioceseResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDioceseResponse(row))
	}
	return items, hasMore, nil
}

// GetDiocese returns a single non-deleted Diocese by ID.
func (s *Service) GetDiocese(ctx context.Context, id uuid.UUID) (domain.DioceseResponse, error) {
	d, err := s.repo.FindDioceseByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.DioceseResponse{}, domain.ErrDioceseNotFound
	}
	if err != nil {
		return domain.DioceseResponse{}, err
	}
	return toDioceseResponse(d), nil
}

// CreateDiocese creates a Diocese or reactivates a soft-deleted one with the same code.
func (s *Service) CreateDiocese(ctx context.Context, req domain.CreateDioceseRequest) (domain.DioceseResponse, error) {
	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.DioceseResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.DioceseResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
	}
	phone, err := normalizeOptPhone(req.Phone)
	if err != nil {
		return domain.DioceseResponse{}, err
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.DioceseInput{
		Code:       code,
		Name:       name,
		BishopName: trimOptString(req.BishopName),
		Address:    trimOptString(req.Address),
		Phone:      phone,
		Email:      normalizeOptEmail(req.Email),
		Website:    trimOptString(req.Website),
		Active:     active,
	}

	var out *ent.Diocese
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindDioceseByCodeIncludeDeleted(ctx, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateDiocese(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateDiocese(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.DioceseResponse{}, err
	}
	return toDioceseResponse(out), nil
}

// UpdateDiocese updates mutable fields of a non-deleted Diocese.
func (s *Service) UpdateDiocese(ctx context.Context, id uuid.UUID, req domain.UpdateDioceseRequest) (domain.DioceseResponse, error) {
	existing, err := s.repo.FindDioceseByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.DioceseResponse{}, domain.ErrDioceseNotFound
	}
	if err != nil {
		return domain.DioceseResponse{}, err
	}

	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.DioceseResponse{}, err
		}
		if normCode != existing.Code {
			if _, findErr := s.repo.FindDioceseByCodeIncludeDeleted(ctx, normCode); findErr == nil {
				return domain.DioceseResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.DioceseResponse{}, findErr
			}
		}
		codePtr = &normCode
	}

	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.DioceseResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}

	phonePtr, err := normalizeOptPhone(req.Phone)
	if err != nil {
		return domain.DioceseResponse{}, err
	}

	updated, err := s.repo.UpdateDiocese(
		ctx,
		id,
		codePtr,
		namePtr,
		trimOptString(req.BishopName),
		trimOptString(req.Address),
		phonePtr,
		normalizeOptEmail(req.Email),
		trimOptString(req.Website),
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.DioceseResponse{}, domain.ErrCodeConflict
		}
		return domain.DioceseResponse{}, err
	}
	return toDioceseResponse(updated), nil
}

// DeleteDiocese soft-deletes a Diocese if it has no active parishes.
func (s *Service) DeleteDiocese(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindDioceseByID(ctx, id); ent.IsNotFound(err) {
		return domain.ErrDioceseNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasActiveParishesByDiocese(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteDiocese(ctx, id, time.Now().UTC())
}

func toDioceseResponse(d *ent.Diocese) domain.DioceseResponse {
	return domain.DioceseResponse{
		ID:         d.ID,
		Code:       d.Code,
		Name:       d.Name,
		BishopName: d.BishopName,
		Address:    d.Address,
		Phone:      d.Phone,
		Email:      d.Email,
		Website:    d.Website,
		Active:     d.Active,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}
