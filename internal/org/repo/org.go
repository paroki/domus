package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/liturgiministry"
	"github.com/paroki/domus/ent/liturgipetugas"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/orgperiode"
)

// OrgInput holds fields for creating or reactivating an Organization.
type OrgInput struct {
	ParokiID     uuid.UUID
	ParentID     *uuid.UUID
	WilayahID    *uuid.UUID
	LingkunganID *uuid.UUID
	Kind         organization.Kind
	Code         string
	Name         string
	Description  *string
	Active       bool
}

// SearchOrgs returns paginated non-deleted organizations in parokiID.
func (r *Repository) SearchOrgs(
	ctx context.Context,
	parokiID uuid.UUID,
	q, kind string,
	active *bool,
	parentID, memberUserID *uuid.UUID,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Organization, error) {
	query := r.client.Organization.Query().
		Where(organization.ParokiIDEQ(parokiID), organization.DeletedAtIsNil())
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

// FindOrgInParoki returns a non-deleted Organization belonging to parokiID.
func (r *Repository) FindOrgInParoki(ctx context.Context, parokiID, orgID uuid.UUID) (*ent.Organization, error) {
	return r.client.Organization.Query().
		Where(organization.IDEQ(orgID), organization.ParokiIDEQ(parokiID), organization.DeletedAtIsNil()).
		Only(ctx)
}

// FindOrgByCodeIncludeDeleted returns an Organization in parokiID by code, including soft-deleted rows.
func (r *Repository) FindOrgByCodeIncludeDeleted(ctx context.Context, parokiID uuid.UUID, code string) (*ent.Organization, error) {
	return r.client.Organization.Query().
		Where(organization.ParokiIDEQ(parokiID), organization.CodeEQ(code)).
		Only(ctx)
}

// CreateOrg inserts a new Organization row.
func (r *Repository) CreateOrg(ctx context.Context, in OrgInput) (*ent.Organization, error) {
	q := r.client.Organization.Create().
		SetParokiID(in.ParokiID).
		SetKind(in.Kind).
		SetCode(in.Code).
		SetName(in.Name).
		SetActive(in.Active)
	if in.ParentID != nil {
		q.SetParentID(*in.ParentID)
	}
	if in.WilayahID != nil {
		q.SetWilayahID(*in.WilayahID)
	}
	if in.LingkunganID != nil {
		q.SetLingkunganID(*in.LingkunganID)
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
	setOrClearOrgUUID(in.WilayahID, upd.SetWilayahID, upd.ClearWilayahID)
	setOrClearOrgUUID(in.LingkunganID, upd.SetLingkunganID, upd.ClearLingkunganID)
	setOrClearOrgString(in.Description, upd.SetDescription, upd.ClearDescription)
	return upd.Save(ctx)
}

// UpdateOrg updates mutable fields of an Organization.
func (r *Repository) UpdateOrg(
	ctx context.Context,
	id uuid.UUID,
	parentID, wilayahID, lingkunganID *uuid.UUID,
	kind *organization.Kind,
	code, name, description *string,
	active *bool,
) (*ent.Organization, error) {
	upd := r.client.Organization.UpdateOneID(id)
	if parentID != nil {
		setOrClearOrgUUID(parentID, upd.SetParentID, upd.ClearParentID)
	}
	if wilayahID != nil {
		setOrClearOrgUUID(wilayahID, upd.SetWilayahID, upd.ClearWilayahID)
	}
	if lingkunganID != nil {
		setOrClearOrgUUID(lingkunganID, upd.SetLingkunganID, upd.ClearLingkunganID)
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

// HasOrgEvents reports whether any non-deleted Event, LiturgiMinistry, or active-event LiturgiPetugas references orgID.
func (r *Repository) HasOrgEvents(ctx context.Context, orgID uuid.UUID) (bool, error) {
	hasOwned, err := r.client.Event.Query().
		Where(event.OrgIDEQ(orgID), event.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil || hasOwned {
		return hasOwned, err
	}
	hasMinistry, err := r.client.LiturgiMinistry.Query().
		Where(liturgiministry.OrgIDEQ(orgID)).
		Exist(ctx)
	if err != nil || hasMinistry {
		return hasMinistry, err
	}
	rows, err := r.client.LiturgiPetugas.Query().
		Where(liturgipetugas.Or(
			liturgipetugas.OrgIDEQ(orgID),
			liturgipetugas.SubstituteOrgIDEQ(orgID),
		)).
		Select(liturgipetugas.FieldEventID).
		All(ctx)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, nil
	}
	evIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		evIDs = append(evIDs, row.EventID)
	}
	return r.client.Event.Query().
		Where(event.IDIn(evIDs...), event.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrgPeriode reports whether any non-deleted OrgPeriode belongs to orgID.
func (r *Repository) HasOrgPeriode(ctx context.Context, orgID uuid.UUID) (bool, error) {
	return r.client.OrgPeriode.Query().
		Where(orgperiode.OrgIDEQ(orgID), orgperiode.DeletedAtIsNil()).
		Exist(ctx)
}

// ListUserOrgsInParoki returns all active, non-deleted Organizations in parokiID where userID is an active member.
func (r *Repository) ListUserOrgsInParoki(ctx context.Context, parokiID, userID uuid.UUID) ([]*ent.Organization, error) {
	return r.client.Organization.Query().
		Where(
			organization.ParokiIDEQ(parokiID),
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
