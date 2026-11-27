package service

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/repo"
)

const roleSuperAdmin = "user:super-admin"

var referensiCodeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

// SearchReferensi searches global and/or parish-scoped umat referensi entries with tenant override.
func (s *Service) SearchReferensi(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchUmatReferensiRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.UmatReferensiResponse, bool, error) {
	var category, scope string
	var active *bool
	if req.Filter != nil {
		category = strings.ToLower(strings.TrimSpace(req.Filter.Category))
		if category != "" && !domain.IsAllowedReferensiCategory(category) {
			return nil, false, domain.ValidationError("category", "INVALID_CATEGORY", "Category must be one of the 12 allowed umat referensi categories.")
		}
		active = req.Filter.Active
		scope = req.Filter.Scope
	}
	q := strings.TrimSpace(req.Q)

	if scope == "paroki" && p.ParokiID == uuid.Nil {
		return nil, false, authx.ErrTenantRequired
	}

	var globalRows, parokiRows []*ent.UmatReferensi
	var err error
	if scope != "paroki" {
		globalRows, err = s.repo.ListGlobalReferensi(ctx, q, category, active)
		if err != nil {
			return nil, false, err
		}
	}
	if scope != "global" && p.ParokiID != uuid.Nil {
		parokiRows, err = s.repo.ListParokiReferensi(ctx, p.ParokiID, q, category, active)
		if err != nil {
			return nil, false, err
		}
	}

	// ponytail: in-memory merge+cursor for hybrid global+paroki referensi; move to SQL UNION when rows > 5k
	mergedMap := make(map[string]*ent.UmatReferensi, len(globalRows)+len(parokiRows))
	for _, row := range globalRows {
		mergedMap[row.Category+"\x00"+row.Code] = row
	}
	for _, row := range parokiRows {
		mergedMap[row.Category+"\x00"+row.Code] = row
	}

	merged := make([]*ent.UmatReferensi, 0, len(mergedMap))
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
		filtered = make([]*ent.UmatReferensi, 0, len(merged))
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
	items := make([]domain.UmatReferensiResponse, 0, len(filtered))
	for _, row := range filtered {
		items = append(items, toReferensiResponse(row))
	}
	return items, hasMore, nil
}

// hasActiveReferensi reports whether code exists and is active in category for global or parokiID (with paroki override).
func (s *Service) hasActiveReferensi(ctx context.Context, parokiID uuid.UUID, category, code string) (bool, error) {
	cat := strings.ToLower(strings.TrimSpace(category))
	c := strings.ToLower(strings.TrimSpace(code))
	if cat == "" || c == "" {
		return false, nil
	}
	globalRows, err := s.repo.ListGlobalReferensi(ctx, "", cat, nil)
	if err != nil {
		return false, err
	}
	merged := make(map[string]*ent.UmatReferensi, len(globalRows))
	for _, row := range globalRows {
		merged[row.Code] = row
	}
	if parokiID != uuid.Nil {
		parokiRows, err := s.repo.ListParokiReferensi(ctx, parokiID, "", cat, nil)
		if err != nil {
			return false, err
		}
		for _, row := range parokiRows {
			merged[row.Code] = row
		}
	}
	row, ok := merged[c]
	return ok && row.Active, nil
}

// GetReferensi returns a non-deleted UmatReferensi if global or belonging to p.ParokiID.
func (s *Service) GetReferensi(ctx context.Context, p authx.Principal, id uuid.UUID) (domain.UmatReferensiResponse, error) {
	row, err := s.repo.FindReferensiByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.UmatReferensiResponse{}, domain.ErrReferensiNotFound
	}
	if err != nil {
		return domain.UmatReferensiResponse{}, err
	}
	if row.ParokiID != nil && *row.ParokiID != p.ParokiID {
		return domain.UmatReferensiResponse{}, domain.ErrReferensiNotFound
	}
	return toReferensiResponse(row), nil
}

// CreateReferensi creates a global (super-admin) or tenant (umat:admin) UmatReferensi, reactivating soft-deleted matches.
func (s *Service) CreateReferensi(ctx context.Context, p authx.Principal, req domain.CreateUmatReferensiRequest) (domain.UmatReferensiResponse, error) {
	var parokiID *uuid.UUID
	if req.Global {
		if !s.isSuperAdmin(p) {
			return domain.UmatReferensiResponse{}, domain.ErrForbidden
		}
	} else {
		if p.ParokiID == uuid.Nil {
			return domain.UmatReferensiResponse{}, authx.ErrTenantRequired
		}
		isAdmin, err := s.isUmatAdmin(p, p.ParokiID)
		if err != nil {
			return domain.UmatReferensiResponse{}, err
		}
		if !isAdmin {
			return domain.UmatReferensiResponse{}, domain.ErrForbidden
		}
		pid := p.ParokiID
		parokiID = &pid
	}

	category, err := normalizeReferensiCategory(req.Category)
	if err != nil {
		return domain.UmatReferensiResponse{}, err
	}
	code, err := normalizeReferensiCode(req.Code)
	if err != nil {
		return domain.UmatReferensiResponse{}, err
	}
	label := strings.TrimSpace(req.Label)
	if label == "" {
		return domain.UmatReferensiResponse{}, domain.ValidationError("label", "REQUIRED", "Label cannot be empty.")
	}

	in := repo.UmatReferensiInput{
		ParokiID:  parokiID,
		Category:  category,
		Code:      code,
		Label:     label,
		SortOrder: req.SortOrder,
	}

	var out *ent.UmatReferensi
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindReferensiByKeyIncludeDeleted(ctx, parokiID, category, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrReferensiConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateReferensi(ctx, existing.ID, in)
			return findErr
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateReferensi(ctx, in)
			if findErr != nil && ent.IsConstraintError(findErr) {
				return domain.ErrReferensiConflict
			}
			return findErr
		default:
			return findErr
		}
	})
	if err != nil {
		return domain.UmatReferensiResponse{}, err
	}
	return toReferensiResponse(out), nil
}

// UpdateReferensi updates a global (super-admin) or tenant (umat:admin) UmatReferensi.
func (s *Service) UpdateReferensi(ctx context.Context, p authx.Principal, id uuid.UUID, req domain.UpdateUmatReferensiRequest) (domain.UmatReferensiResponse, error) {
	existing, err := s.repo.FindReferensiByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.UmatReferensiResponse{}, domain.ErrReferensiNotFound
	}
	if err != nil {
		return domain.UmatReferensiResponse{}, err
	}
	if err := s.authorizeReferensiWrite(p, existing); err != nil {
		return domain.UmatReferensiResponse{}, err
	}

	targetCat := existing.Category
	var catPtr *string
	if req.Category != nil {
		normCat, err := normalizeReferensiCategory(*req.Category)
		if err != nil {
			return domain.UmatReferensiResponse{}, err
		}
		targetCat = normCat
		catPtr = &normCat
	}
	targetCode := existing.Code
	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeReferensiCode(*req.Code)
		if err != nil {
			return domain.UmatReferensiResponse{}, err
		}
		targetCode = normCode
		codePtr = &normCode
	}
	if targetCat != existing.Category || targetCode != existing.Code {
		if _, findErr := s.repo.FindReferensiByKeyIncludeDeleted(ctx, existing.ParokiID, targetCat, targetCode); findErr == nil {
			return domain.UmatReferensiResponse{}, domain.ErrReferensiConflict
		} else if !ent.IsNotFound(findErr) {
			return domain.UmatReferensiResponse{}, findErr
		}
	}

	var labelPtr *string
	if req.Label != nil {
		trimmed := strings.TrimSpace(*req.Label)
		if trimmed == "" {
			return domain.UmatReferensiResponse{}, domain.ValidationError("label", "REQUIRED", "Label cannot be empty.")
		}
		labelPtr = &trimmed
	}

	updated, err := s.repo.UpdateReferensi(ctx, id, catPtr, codePtr, labelPtr, req.SortOrder, req.Active)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.UmatReferensiResponse{}, domain.ErrReferensiConflict
		}
		return domain.UmatReferensiResponse{}, err
	}
	return toReferensiResponse(updated), nil
}

// DeleteReferensi soft-deletes a global (super-admin) or tenant (umat:admin) UmatReferensi if not in use.
func (s *Service) DeleteReferensi(ctx context.Context, p authx.Principal, id uuid.UUID) error {
	existing, err := s.repo.FindReferensiByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.ErrReferensiNotFound
	}
	if err != nil {
		return err
	}
	if err := s.authorizeReferensiWrite(p, existing); err != nil {
		return err
	}
	inUse, err := s.repo.IsReferensiCodeInUse(ctx, existing.ParokiID, existing.Category, existing.Code)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteReferensi(ctx, id, time.Now().UTC())
}

func (s *Service) authorizeReferensiWrite(p authx.Principal, row *ent.UmatReferensi) error {
	if row.ParokiID == nil {
		if !s.isSuperAdmin(p) {
			return domain.ErrForbidden
		}
		return nil
	}
	if p.ParokiID == uuid.Nil {
		return authx.ErrTenantRequired
	}
	if *row.ParokiID != p.ParokiID {
		return domain.ErrReferensiNotFound
	}
	isAdmin, err := s.isUmatAdmin(p, p.ParokiID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return domain.ErrForbidden
	}
	return nil
}

func (s *Service) isSuperAdmin(p authx.Principal) bool {
	if s.enforcer == nil || p.UserID == uuid.Nil {
		return false
	}
	ok, err := s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleSuperAdmin, authz.DomainGlobal)
	return err == nil && ok
}

func (s *Service) isUmatAdmin(p authx.Principal, parokiID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || parokiID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), domain.RoleAdmin, authz.ParokiDomain(parokiID))
}

func normalizeReferensiCategory(raw string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(raw))
	if !domain.IsAllowedReferensiCategory(norm) {
		return "", domain.ValidationError("category", "INVALID_CATEGORY", "Category must be one of the 12 allowed umat referensi categories.")
	}
	return norm, nil
}

func normalizeReferensiCode(raw string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(raw))
	if !referensiCodeRe.MatchString(norm) {
		return "", domain.ValidationError("code", "INVALID_CODE", "Code must be 2-64 lowercase alphanumeric characters, hyphens, or underscores.")
	}
	return norm, nil
}

func toReferensiResponse(r *ent.UmatReferensi) domain.UmatReferensiResponse {
	return domain.UmatReferensiResponse{
		ID:        r.ID,
		ParokiID:  r.ParokiID,
		Category:  r.Category,
		Code:      r.Code,
		Label:     r.Label,
		SortOrder: r.SortOrder,
		Active:    r.Active,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
