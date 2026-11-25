package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

// SearchLookups searches global and/or parish-scoped lookups with tenant override.
func (s *Service) SearchLookups(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchLookupsRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.LookupResponse, bool, error) {
	var category, scope string
	var active *bool
	if req.Filter != nil {
		category = strings.ToLower(strings.TrimSpace(req.Filter.Category))
		active = req.Filter.Active
		scope = req.Filter.Scope
	}
	q := strings.TrimSpace(req.Q)

	if scope == "paroki" && p.ParokiID == uuid.Nil {
		return nil, false, authx.ErrTenantRequired
	}

	var globalRows, parokiRows []*ent.Lookup
	var err error
	if scope != "paroki" {
		globalRows, err = s.repo.ListGlobalLookups(ctx, q, category, active)
		if err != nil {
			return nil, false, err
		}
	}
	if scope != "global" && p.ParokiID != uuid.Nil {
		parokiRows, err = s.repo.ListParokiLookups(ctx, p.ParokiID, q, category, active)
		if err != nil {
			return nil, false, err
		}
	}

	// ponytail: in-memory merge+cursor for hybrid global+paroki lookups; move to SQL UNION when lookups > 5k
	mergedMap := make(map[string]*ent.Lookup, len(globalRows)+len(parokiRows))
	for _, row := range globalRows {
		mergedMap[row.Category+"\x00"+row.Code] = row
	}
	for _, row := range parokiRows {
		mergedMap[row.Category+"\x00"+row.Code] = row
	}

	merged := make([]*ent.Lookup, 0, len(mergedMap))
	for _, row := range mergedMap {
		merged = append(merged, row)
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].SortOrder != merged[j].SortOrder {
			return merged[i].SortOrder < merged[j].SortOrder
		}
		if !merged[i].CreatedAt.Equal(merged[j].CreatedAt) {
			return merged[i].CreatedAt.After(merged[j].CreatedAt)
		}
		return merged[i].ID.String() > merged[j].ID.String()
	})

	filtered := merged
	if cursorCreatedAt != nil && cursorID != nil {
		filtered = make([]*ent.Lookup, 0, len(merged))
		for _, row := range merged {
			if row.CreatedAt.Before(*cursorCreatedAt) ||
				(row.CreatedAt.Equal(*cursorCreatedAt) && row.ID.String() < cursorID.String()) {
				filtered = append(filtered, row)
			}
		}
	}

	hasMore := len(filtered) > limit
	if hasMore {
		filtered = filtered[:limit]
	}
	items := make([]domain.LookupResponse, 0, len(filtered))
	for _, row := range filtered {
		items = append(items, toLookupResponse(row))
	}
	return items, hasMore, nil
}

// GetLookup returns a non-deleted Lookup if global or belonging to p.ParokiID.
func (s *Service) GetLookup(ctx context.Context, p authx.Principal, id uuid.UUID) (domain.LookupResponse, error) {
	row, err := s.repo.FindLookupByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.LookupResponse{}, domain.ErrLookupNotFound
	}
	if err != nil {
		return domain.LookupResponse{}, err
	}
	if row.ParokiID != nil && *row.ParokiID != p.ParokiID {
		return domain.LookupResponse{}, domain.ErrLookupNotFound
	}
	return toLookupResponse(row), nil
}

// CreateLookup creates a global (super-admin) or tenant (paroki-admin) Lookup, reactivating soft-deleted matches.
func (s *Service) CreateLookup(ctx context.Context, p authx.Principal, req domain.CreateLookupRequest) (domain.LookupResponse, error) {
	var parokiID *uuid.UUID
	if req.Global {
		if !s.isSuperAdmin(p) {
			return domain.LookupResponse{}, domain.ErrRefForbidden
		}
	} else {
		if p.ParokiID == uuid.Nil {
			return domain.LookupResponse{}, authx.ErrTenantRequired
		}
		isAdmin, err := s.isParokiAdmin(p, p.ParokiID)
		if err != nil {
			return domain.LookupResponse{}, err
		}
		if !isAdmin {
			return domain.LookupResponse{}, domain.ErrRefForbidden
		}
		pid := p.ParokiID
		parokiID = &pid
	}

	category, err := normalizeLookupKey("category", req.Category)
	if err != nil {
		return domain.LookupResponse{}, err
	}
	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.LookupResponse{}, err
	}
	label := strings.TrimSpace(req.Label)
	if label == "" {
		return domain.LookupResponse{}, domain.ValidationError("label", "REQUIRED", "Label cannot be empty.")
	}

	in := repo.LookupInput{
		ParokiID:  parokiID,
		Category:  category,
		Code:      code,
		Label:     label,
		SortOrder: req.SortOrder,
	}

	var out *ent.Lookup
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindLookupByKeyIncludeDeleted(ctx, parokiID, category, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateLookup(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateLookup(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrCodeConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.LookupResponse{}, err
	}
	return toLookupResponse(out), nil
}

// UpdateLookup updates a global (super-admin) or tenant (parish-admin) Lookup.
func (s *Service) UpdateLookup(ctx context.Context, p authx.Principal, id uuid.UUID, req domain.UpdateLookupRequest) (domain.LookupResponse, error) {
	existing, err := s.repo.FindLookupByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.LookupResponse{}, domain.ErrLookupNotFound
	}
	if err != nil {
		return domain.LookupResponse{}, err
	}
	if err := s.authorizeLookupWrite(p, existing); err != nil {
		return domain.LookupResponse{}, err
	}

	targetCat := existing.Category
	var catPtr *string
	if req.Category != nil {
		normCat, err := normalizeLookupKey("category", *req.Category)
		if err != nil {
			return domain.LookupResponse{}, err
		}
		targetCat = normCat
		catPtr = &normCat
	}
	targetCode := existing.Code
	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.LookupResponse{}, err
		}
		targetCode = normCode
		codePtr = &normCode
	}
	if targetCat != existing.Category || targetCode != existing.Code {
		if _, findErr := s.repo.FindLookupByKeyIncludeDeleted(ctx, existing.ParokiID, targetCat, targetCode); findErr == nil {
			return domain.LookupResponse{}, domain.ErrCodeConflict
		} else if !ent.IsNotFound(findErr) {
			return domain.LookupResponse{}, findErr
		}
	}

	var labelPtr *string
	if req.Label != nil {
		trimmed := strings.TrimSpace(*req.Label)
		if trimmed == "" {
			return domain.LookupResponse{}, domain.ValidationError("label", "REQUIRED", "Label cannot be empty.")
		}
		labelPtr = &trimmed
	}

	updated, err := s.repo.UpdateLookup(ctx, id, catPtr, codePtr, labelPtr, req.SortOrder, req.Active)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.LookupResponse{}, domain.ErrCodeConflict
		}
		return domain.LookupResponse{}, err
	}
	return toLookupResponse(updated), nil
}

// DeleteLookup soft-deletes a global (super-admin) or tenant (parish-admin) Lookup.
func (s *Service) DeleteLookup(ctx context.Context, p authx.Principal, id uuid.UUID) error {
	existing, err := s.repo.FindLookupByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.ErrLookupNotFound
	}
	if err != nil {
		return err
	}
	if err := s.authorizeLookupWrite(p, existing); err != nil {
		return err
	}
	return s.repo.SoftDeleteLookup(ctx, id, time.Now().UTC())
}

func (s *Service) authorizeLookupWrite(p authx.Principal, row *ent.Lookup) error {
	if row.ParokiID == nil {
		if !s.isSuperAdmin(p) {
			return domain.ErrRefForbidden
		}
		return nil
	}
	if p.ParokiID == uuid.Nil {
		return authx.ErrTenantRequired
	}
	if *row.ParokiID != p.ParokiID {
		return domain.ErrLookupNotFound
	}
	isAdmin, err := s.isParokiAdmin(p, p.ParokiID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return domain.ErrRefForbidden
	}
	return nil
}

func normalizeLookupKey(field, raw string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(raw))
	if !codeRe.MatchString(norm) {
		return "", domain.ValidationError(field, "INVALID_"+strings.ToUpper(field), "Must be 2-64 lowercase alphanumeric characters, hyphens, or underscores.")
	}
	return norm, nil
}

func toLookupResponse(l *ent.Lookup) domain.LookupResponse {
	return domain.LookupResponse{
		ID:        l.ID,
		ParokiID:  l.ParokiID,
		Category:  l.Category,
		Code:      l.Code,
		Label:     l.Label,
		SortOrder: l.SortOrder,
		Active:    l.Active,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
