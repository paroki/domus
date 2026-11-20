package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/ward"
	"github.com/paroki/domus/ent/zone"
)

// ZoneInput holds fields for creating or reactivating a Zone.
type ZoneInput struct {
	ParishID         uuid.UUID
	Type             zone.Type
	Code             string
	Name             string
	CoordinatorName  *string
	CoordinatorPhone *string
}

// WardInput holds fields for creating or reactivating a Ward.
type WardInput struct {
	ParishID    uuid.UUID
	ZoneID      uuid.UUID
	Code        string
	Name        string
	PatronSaint *string
	HeadName    *string
	HeadPhone   *string
}

// SearchZones returns paginated non-deleted zones for parishID.
func (r *Repository) SearchZones(
	ctx context.Context,
	parishID uuid.UUID,
	q, zoneType string,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Zone, error) {
	query := r.client.Zone.Query().Where(zone.ParishIDEQ(parishID), zone.DeletedAtIsNil())
	if q != "" {
		query = query.Where(zone.Or(
			zone.CodeContainsFold(q),
			zone.NameContainsFold(q),
		))
	}
	if zoneType != "" {
		query = query.Where(zone.TypeEQ(zone.Type(zoneType)))
	}
	if active != nil {
		query = query.Where(zone.ActiveEQ(*active))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(zone.Or(
			zone.CreatedAtLT(*cursorCreatedAt),
			zone.And(
				zone.CreatedAtEQ(*cursorCreatedAt),
				zone.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(zone.FieldCreatedAt), ent.Desc(zone.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindZoneByCodeIncludeDeleted returns a Zone by (parishID, code), including soft-deleted rows.
func (r *Repository) FindZoneByCodeIncludeDeleted(ctx context.Context, parishID uuid.UUID, code string) (*ent.Zone, error) {
	return r.client.Zone.Query().
		Where(zone.ParishIDEQ(parishID), zone.CodeEQ(code)).
		Only(ctx)
}

// CreateZone inserts a new Zone.
func (r *Repository) CreateZone(ctx context.Context, in ZoneInput) (*ent.Zone, error) {
	return r.client.Zone.Create().
		SetParishID(in.ParishID).
		SetType(in.Type).
		SetCode(in.Code).
		SetName(in.Name).
		SetNillableCoordinatorName(in.CoordinatorName).
		SetNillableCoordinatorPhone(in.CoordinatorPhone).
		SetActive(true).
		Save(ctx)
}

// ReactivateZone restores a soft-deleted Zone with fresh fields.
func (r *Repository) ReactivateZone(ctx context.Context, id uuid.UUID, in ZoneInput) (*ent.Zone, error) {
	upd := r.client.Zone.UpdateOneID(id).
		ClearDeletedAt().
		SetType(in.Type).
		SetCode(in.Code).
		SetName(in.Name).
		SetActive(true)
	setOrClearZoneString(in.CoordinatorName, upd.SetCoordinatorName, upd.ClearCoordinatorName)
	setOrClearZoneString(in.CoordinatorPhone, upd.SetCoordinatorPhone, upd.ClearCoordinatorPhone)
	return upd.Save(ctx)
}

// UpdateZone updates mutable fields of a Zone.
func (r *Repository) UpdateZone(
	ctx context.Context,
	id uuid.UUID,
	zoneType *zone.Type,
	code, name, coordinatorName, coordinatorPhone *string,
	active *bool,
) (*ent.Zone, error) {
	upd := r.client.Zone.UpdateOneID(id)
	if zoneType != nil {
		upd.SetType(*zoneType)
	}
	if code != nil {
		upd.SetCode(*code)
	}
	if name != nil {
		upd.SetName(*name)
	}
	if coordinatorName != nil {
		setOrClearZoneString(coordinatorName, upd.SetCoordinatorName, upd.ClearCoordinatorName)
	}
	if coordinatorPhone != nil {
		setOrClearZoneString(coordinatorPhone, upd.SetCoordinatorPhone, upd.ClearCoordinatorPhone)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteZone sets deleted_at and marks the Zone inactive.
func (r *Repository) SoftDeleteZone(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Zone.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasWardsByZone reports whether any non-deleted Ward references zoneID in parishID.
func (r *Repository) HasWardsByZone(ctx context.Context, parishID, zoneID uuid.UUID) (bool, error) {
	return r.client.Ward.Query().
		Where(ward.ParishIDEQ(parishID), ward.ZoneIDEQ(zoneID), ward.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrganizationsByZone reports whether any non-deleted Organization references zoneID in parishID.
func (r *Repository) HasOrganizationsByZone(ctx context.Context, parishID, zoneID uuid.UUID) (bool, error) {
	return r.client.Organization.Query().
		Where(organization.ParishIDEQ(parishID), organization.ZoneIDEQ(zoneID), organization.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrganizationsByWard reports whether any non-deleted Organization references wardID in parishID.
func (r *Repository) HasOrganizationsByWard(ctx context.Context, parishID, wardID uuid.UUID) (bool, error) {
	return r.client.Organization.Query().
		Where(organization.ParishIDEQ(parishID), organization.WardIDEQ(wardID), organization.DeletedAtIsNil()).
		Exist(ctx)
}

// SearchWards returns paginated non-deleted wards for parishID.
func (r *Repository) SearchWards(
	ctx context.Context,
	parishID uuid.UUID,
	q string,
	zoneID *uuid.UUID,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Ward, error) {
	query := r.client.Ward.Query().Where(ward.ParishIDEQ(parishID), ward.DeletedAtIsNil())
	if q != "" {
		query = query.Where(ward.Or(
			ward.CodeContainsFold(q),
			ward.NameContainsFold(q),
		))
	}
	if zoneID != nil && *zoneID != uuid.Nil {
		query = query.Where(ward.ZoneIDEQ(*zoneID))
	}
	if active != nil {
		query = query.Where(ward.ActiveEQ(*active))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(ward.Or(
			ward.CreatedAtLT(*cursorCreatedAt),
			ward.And(
				ward.CreatedAtEQ(*cursorCreatedAt),
				ward.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(ward.FieldCreatedAt), ent.Desc(ward.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindWardByCodeIncludeDeleted returns a Ward by (parishID, code), including soft-deleted rows.
func (r *Repository) FindWardByCodeIncludeDeleted(ctx context.Context, parishID uuid.UUID, code string) (*ent.Ward, error) {
	return r.client.Ward.Query().
		Where(ward.ParishIDEQ(parishID), ward.CodeEQ(code)).
		Only(ctx)
}

// CreateWard inserts a new Ward.
func (r *Repository) CreateWard(ctx context.Context, in WardInput) (*ent.Ward, error) {
	return r.client.Ward.Create().
		SetParishID(in.ParishID).
		SetZoneID(in.ZoneID).
		SetCode(in.Code).
		SetName(in.Name).
		SetNillablePatronSaint(in.PatronSaint).
		SetNillableHeadName(in.HeadName).
		SetNillableHeadPhone(in.HeadPhone).
		SetActive(true).
		Save(ctx)
}

// ReactivateWard restores a soft-deleted Ward with fresh fields.
func (r *Repository) ReactivateWard(ctx context.Context, id uuid.UUID, in WardInput) (*ent.Ward, error) {
	upd := r.client.Ward.UpdateOneID(id).
		ClearDeletedAt().
		SetZoneID(in.ZoneID).
		SetCode(in.Code).
		SetName(in.Name).
		SetActive(true)
	setOrClearWardString(in.PatronSaint, upd.SetPatronSaint, upd.ClearPatronSaint)
	setOrClearWardString(in.HeadName, upd.SetHeadName, upd.ClearHeadName)
	setOrClearWardString(in.HeadPhone, upd.SetHeadPhone, upd.ClearHeadPhone)
	return upd.Save(ctx)
}

// UpdateWard updates mutable fields of a Ward.
func (r *Repository) UpdateWard(
	ctx context.Context,
	id uuid.UUID,
	zoneID *uuid.UUID,
	code, name, patronSaint, headName, headPhone *string,
	active *bool,
) (*ent.Ward, error) {
	upd := r.client.Ward.UpdateOneID(id)
	if zoneID != nil {
		upd.SetZoneID(*zoneID)
	}
	if code != nil {
		upd.SetCode(*code)
	}
	if name != nil {
		upd.SetName(*name)
	}
	if patronSaint != nil {
		setOrClearWardString(patronSaint, upd.SetPatronSaint, upd.ClearPatronSaint)
	}
	if headName != nil {
		setOrClearWardString(headName, upd.SetHeadName, upd.ClearHeadName)
	}
	if headPhone != nil {
		setOrClearWardString(headPhone, upd.SetHeadPhone, upd.ClearHeadPhone)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteWard sets deleted_at and marks the Ward inactive.
func (r *Repository) SoftDeleteWard(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Ward.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

func setOrClearZoneString(val *string, setFn func(string) *ent.ZoneUpdateOne, clearFn func() *ent.ZoneUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}

func setOrClearWardString(val *string, setFn func(string) *ent.WardUpdateOne, clearFn func() *ent.WardUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
