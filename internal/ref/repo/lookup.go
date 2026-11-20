package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/lookup"
)

// LookupInput holds fields for creating or reactivating a Lookup.
type LookupInput struct {
	ParishID  *uuid.UUID
	Category  string
	Code      string
	Label     string
	SortOrder int
}

// ListGlobalLookups returns non-deleted global lookups matching optional filters.
func (r *Repository) ListGlobalLookups(ctx context.Context, q, category string, active *bool) ([]*ent.Lookup, error) {
	query := r.client.Lookup.Query().Where(lookup.ParishIDIsNil(), lookup.DeletedAtIsNil())
	query = applyLookupFilters(query, q, category, active)
	return query.Order(ent.Desc(lookup.FieldCreatedAt), ent.Desc(lookup.FieldID)).All(ctx)
}

// ListParishLookups returns non-deleted tenant lookups for parishID matching optional filters.
func (r *Repository) ListParishLookups(ctx context.Context, parishID uuid.UUID, q, category string, active *bool) ([]*ent.Lookup, error) {
	query := r.client.Lookup.Query().Where(lookup.ParishIDEQ(parishID), lookup.DeletedAtIsNil())
	query = applyLookupFilters(query, q, category, active)
	return query.Order(ent.Desc(lookup.FieldCreatedAt), ent.Desc(lookup.FieldID)).All(ctx)
}

func applyLookupFilters(query *ent.LookupQuery, q, category string, active *bool) *ent.LookupQuery {
	if q != "" {
		query = query.Where(lookup.Or(
			lookup.CodeContainsFold(q),
			lookup.LabelContainsFold(q),
			lookup.CategoryContainsFold(q),
		))
	}
	if category != "" {
		query = query.Where(lookup.CategoryEQ(category))
	}
	if active != nil {
		query = query.Where(lookup.ActiveEQ(*active))
	}
	return query
}

// FindLookupByID returns a non-deleted Lookup by ID.
func (r *Repository) FindLookupByID(ctx context.Context, id uuid.UUID) (*ent.Lookup, error) {
	return r.client.Lookup.Query().
		Where(lookup.IDEQ(id), lookup.DeletedAtIsNil()).
		Only(ctx)
}

// FindLookupByKeyIncludeDeleted returns a Lookup by (parishID, category, code), including soft-deleted rows.
func (r *Repository) FindLookupByKeyIncludeDeleted(ctx context.Context, parishID *uuid.UUID, category, code string) (*ent.Lookup, error) {
	query := r.client.Lookup.Query().
		Where(lookup.CategoryEQ(category), lookup.CodeEQ(code))
	if parishID == nil {
		query = query.Where(lookup.ParishIDIsNil())
	} else {
		query = query.Where(lookup.ParishIDEQ(*parishID))
	}
	return query.Only(ctx)
}

// CreateLookup inserts a new Lookup.
func (r *Repository) CreateLookup(ctx context.Context, in LookupInput) (*ent.Lookup, error) {
	return r.client.Lookup.Create().
		SetNillableParishID(in.ParishID).
		SetCategory(in.Category).
		SetCode(in.Code).
		SetLabel(in.Label).
		SetSortOrder(in.SortOrder).
		SetActive(true).
		Save(ctx)
}

// ReactivateLookup restores a soft-deleted Lookup with fresh fields.
func (r *Repository) ReactivateLookup(ctx context.Context, id uuid.UUID, in LookupInput) (*ent.Lookup, error) {
	return r.client.Lookup.UpdateOneID(id).
		ClearDeletedAt().
		SetCategory(in.Category).
		SetCode(in.Code).
		SetLabel(in.Label).
		SetSortOrder(in.SortOrder).
		SetActive(true).
		Save(ctx)
}

// UpdateLookup updates mutable fields of a Lookup.
func (r *Repository) UpdateLookup(
	ctx context.Context,
	id uuid.UUID,
	category, code, label *string,
	sortOrder *int,
	active *bool,
) (*ent.Lookup, error) {
	upd := r.client.Lookup.UpdateOneID(id)
	if category != nil {
		upd.SetCategory(*category)
	}
	if code != nil {
		upd.SetCode(*code)
	}
	if label != nil {
		upd.SetLabel(*label)
	}
	if sortOrder != nil {
		upd.SetSortOrder(*sortOrder)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteLookup sets deleted_at and marks the Lookup inactive.
func (r *Repository) SoftDeleteLookup(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Lookup.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}
