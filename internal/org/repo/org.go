package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/orgmember"
)

// OrgInput holds fields for creating or reactivating an Organization.
type OrgInput struct {
	ParishID    uuid.UUID
	ParentID    *uuid.UUID
	ZoneID      *uuid.UUID
	WardID      *uuid.UUID
	Kind        organization.Kind
	Code        string
	Name        string
	Description *string
	Active      bool
}

// SearchOrgs returns paginated non-deleted organizations in parishID.
func (r *Repository) SearchOrgs(
	ctx context.Context,
	parishID uuid.UUID,
	q, kind string,
	active *bool,
	parentID, memberUserID *uuid.UUID,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Organization, error) {
	query := r.client.Organization.Query().
		Where(organization.ParishIDEQ(parishID), organization.DeletedAtIsNil())
	if q != "" {
		query = query.Where(organization.Or(
			organization.NameContainsFold(q),
			organization.CodeContainsFold(q),
		))
	}
	if kind != "" {
		query = query.Where(organization.KindEQ(organization.Kind(kind)))
	}
	if active != nil {
		query = query.Where(organization.ActiveEQ(*active))
	}
	if parentID != nil {
		query = query.Where(organization.ParentIDEQ(*parentID))
	}
	if memberUserID != nil {
		query = query.Where(organization.HasMembersWith(
			orgmember.UserIDEQ(*memberUserID),
			orgmember.StatusEQ(orgmember.StatusActive),
		))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(organization.Or(
			organization.CreatedAtLT(*cursorCreatedAt),
			organization.And(organization.CreatedAtEQ(*cursorCreatedAt), organization.IDLT(*cursorID)),
		))
	}
	return query.
		Order(organization.ByCreatedAt(sql.OrderDesc()), organization.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

// FindOrgInParish returns a non-deleted Organization belonging to parishID.
func (r *Repository) FindOrgInParish(ctx context.Context, parishID, orgID uuid.UUID) (*ent.Organization, error) {
	return r.client.Organization.Query().
		Where(organization.IDEQ(orgID), organization.ParishIDEQ(parishID), organization.DeletedAtIsNil()).
		Only(ctx)
}

// FindOrgByCodeIncludeDeleted returns an Organization in parishID by code, including soft-deleted rows.
func (r *Repository) FindOrgByCodeIncludeDeleted(ctx context.Context, parishID uuid.UUID, code string) (*ent.Organization, error) {
	return r.client.Organization.Query().
		Where(organization.ParishIDEQ(parishID), organization.CodeEQ(code)).
		Only(ctx)
}

// CreateOrg inserts a new Organization row.
func (r *Repository) CreateOrg(ctx context.Context, in OrgInput) (*ent.Organization, error) {
	q := r.client.Organization.Create().
		SetParishID(in.ParishID).
		SetKind(in.Kind).
		SetCode(in.Code).
		SetName(in.Name).
		SetActive(in.Active)
	if in.ParentID != nil {
		q.SetParentID(*in.ParentID)
	}
	if in.ZoneID != nil {
		q.SetZoneID(*in.ZoneID)
	}
	if in.WardID != nil {
		q.SetWardID(*in.WardID)
	}
	if in.Description != nil && *in.Description != "" {
		q.SetDescription(*in.Description)
	}
	return q.Save(ctx)
}

// ReactivateOrg clears deleted_at and updates fields of a soft-deleted Organization.
func (r *Repository) ReactivateOrg(ctx context.Context, id uuid.UUID, in OrgInput) (*ent.Organization, error) {
	upd := r.client.Organization.UpdateOneID(id).
		ClearDeletedAt().
		SetKind(in.Kind).
		SetName(in.Name).
		SetActive(in.Active)
	setOrClearOrgUUID(in.ParentID, upd.SetParentID, upd.ClearParentID)
	setOrClearOrgUUID(in.ZoneID, upd.SetZoneID, upd.ClearZoneID)
	setOrClearOrgUUID(in.WardID, upd.SetWardID, upd.ClearWardID)
	setOrClearOrgString(in.Description, upd.SetDescription, upd.ClearDescription)
	return upd.Save(ctx)
}

// UpdateOrg updates mutable fields of an Organization.
func (r *Repository) UpdateOrg(
	ctx context.Context,
	id uuid.UUID,
	parentID, zoneID, wardID *uuid.UUID,
	kind *organization.Kind,
	code, name, description *string,
	active *bool,
) (*ent.Organization, error) {
	upd := r.client.Organization.UpdateOneID(id)
	if parentID != nil {
		setOrClearOrgUUID(parentID, upd.SetParentID, upd.ClearParentID)
	}
	if zoneID != nil {
		setOrClearOrgUUID(zoneID, upd.SetZoneID, upd.ClearZoneID)
	}
	if wardID != nil {
		setOrClearOrgUUID(wardID, upd.SetWardID, upd.ClearWardID)
	}
	if kind != nil {
		upd.SetKind(*kind)
	}
	if code != nil {
		upd.SetCode(*code)
	}
	if name != nil {
		upd.SetName(*name)
	}
	if description != nil {
		setOrClearOrgString(description, upd.SetDescription, upd.ClearDescription)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteOrg sets deleted_at and marks the Organization inactive.
func (r *Repository) SoftDeleteOrg(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Organization.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasChildOrgs reports whether any non-deleted Organization has parent_id == orgID.
func (r *Repository) HasChildOrgs(ctx context.Context, orgID uuid.UUID) (bool, error) {
	return r.client.Organization.Query().
		Where(organization.ParentIDEQ(orgID), organization.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrgEvents reports whether any non-deleted Event belongs to orgID.
func (r *Repository) HasOrgEvents(ctx context.Context, orgID uuid.UUID) (bool, error) {
	return r.client.Event.Query().
		Where(event.OrgIDEQ(orgID), event.DeletedAtIsNil()).
		Exist(ctx)
}

// ListUserOrgsInParish returns all active, non-deleted Organizations in parishID where userID is an active member.
func (r *Repository) ListUserOrgsInParish(ctx context.Context, parishID, userID uuid.UUID) ([]*ent.Organization, error) {
	return r.client.Organization.Query().
		Where(
			organization.ParishIDEQ(parishID),
			organization.ActiveEQ(true),
			organization.DeletedAtIsNil(),
			organization.HasMembersWith(
				orgmember.UserIDEQ(userID),
				orgmember.StatusEQ(orgmember.StatusActive),
			),
		).
		Order(organization.ByName(sql.OrderAsc()), organization.ByID(sql.OrderAsc())).
		All(ctx)
}

func setOrClearOrgUUID(val *uuid.UUID, setFn func(uuid.UUID) *ent.OrganizationUpdateOne, clearFn func() *ent.OrganizationUpdateOne) {
	if val == nil || *val == uuid.Nil {
		clearFn()
		return
	}
	setFn(*val)
}

func setOrClearOrgString(val *string, setFn func(string) *ent.OrganizationUpdateOne, clearFn func() *ent.OrganizationUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
