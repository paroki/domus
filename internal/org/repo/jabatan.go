package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgjabatan"
	"github.com/paroki/domus/ent/orgperiode"
)

// OrgJabatanInput holds fields for creating an OrgJabatan.
type OrgJabatanInput struct {
	ParokiID  uuid.UUID
	OrgID     uuid.UUID
	PeriodeID uuid.UUID
	ParentID  *uuid.UUID
	Jabatan   string
	UserID    *uuid.UUID
	UmatID    *uuid.UUID
	Nama      *string
	StartDate *string
	EndDate   *string
	Urutan    int
	Active    bool
}

// SearchOrgJabatan returns paginated non-deleted OrgJabatan rows for orgID, ordered by urutan ASC, created_at ASC, id ASC.
func (r *Repository) SearchOrgJabatan(
	ctx context.Context,
	orgID uuid.UUID,
	q string,
	periodeID, parentID, userID, umatID *uuid.UUID,
	jabatan string,
	active *bool,
	cursorUrutan *int,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.OrgJabatan, error) {
	query := r.client.OrgJabatan.Query().
		Where(orgjabatan.OrgIDEQ(orgID), orgjabatan.DeletedAtIsNil())
	if q != "" {
		query = query.Where(orgjabatan.Or(
			orgjabatan.JabatanContainsFold(q),
			orgjabatan.NamaContainsFold(q),
		))
	}
	if periodeID != nil && *periodeID != uuid.Nil {
		query = query.Where(orgjabatan.PeriodeIDEQ(*periodeID))
	}
	if parentID != nil && *parentID != uuid.Nil {
		query = query.Where(orgjabatan.ParentIDEQ(*parentID))
	}
	if userID != nil && *userID != uuid.Nil {
		query = query.Where(orgjabatan.UserIDEQ(*userID))
	}
	if umatID != nil && *umatID != uuid.Nil {
		query = query.Where(orgjabatan.UmatIDEQ(*umatID))
	}
	if jabatan != "" {
		query = query.Where(orgjabatan.JabatanContainsFold(jabatan))
	}
	if active != nil {
		query = query.Where(orgjabatan.ActiveEQ(*active))
	}
	if cursorUrutan != nil && cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(orgjabatan.Or(
			orgjabatan.UrutanGT(*cursorUrutan),
			orgjabatan.And(
				orgjabatan.UrutanEQ(*cursorUrutan),
				orgjabatan.CreatedAtGT(*cursorCreatedAt),
			),
			orgjabatan.And(
				orgjabatan.UrutanEQ(*cursorUrutan),
				orgjabatan.CreatedAtEQ(*cursorCreatedAt),
				orgjabatan.IDGT(*cursorID),
			),
		))
	}
	return query.
		Order(
			orgjabatan.ByUrutan(sql.OrderAsc()),
			orgjabatan.ByCreatedAt(sql.OrderAsc()),
			orgjabatan.ByID(sql.OrderAsc()),
		).
		Limit(limit).
		All(ctx)
}

// FindOrgJabatanInOrg returns a non-deleted OrgJabatan belonging to orgID.
func (r *Repository) FindOrgJabatanInOrg(ctx context.Context, orgID, jabatanID uuid.UUID) (*ent.OrgJabatan, error) {
	return r.client.OrgJabatan.Query().
		Where(orgjabatan.IDEQ(jabatanID), orgjabatan.OrgIDEQ(orgID), orgjabatan.DeletedAtIsNil()).
		Only(ctx)
}

// FindCurrentOrgPeriode returns the non-deleted OrgPeriode marked is_current == true in orgID.
func (r *Repository) FindCurrentOrgPeriode(ctx context.Context, orgID uuid.UUID) (*ent.OrgPeriode, error) {
	return r.client.OrgPeriode.Query().
		Where(orgperiode.OrgIDEQ(orgID), orgperiode.IsCurrentEQ(true), orgperiode.DeletedAtIsNil()).
		Only(ctx)
}

// HasOrgJabatanChildren reports whether any non-deleted OrgJabatan has parent_id == parentID.
func (r *Repository) HasOrgJabatanChildren(ctx context.Context, orgID, parentID uuid.UUID) (bool, error) {
	return r.client.OrgJabatan.Query().
		Where(orgjabatan.OrgIDEQ(orgID), orgjabatan.ParentIDEQ(parentID), orgjabatan.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrgJabatanByPeriode reports whether any non-deleted OrgJabatan references periodeID.
func (r *Repository) HasOrgJabatanByPeriode(ctx context.Context, orgID, periodeID uuid.UUID) (bool, error) {
	return r.client.OrgJabatan.Query().
		Where(orgjabatan.OrgIDEQ(orgID), orgjabatan.PeriodeIDEQ(periodeID), orgjabatan.DeletedAtIsNil()).
		Exist(ctx)
}

// CreateOrgJabatan inserts a new OrgJabatan.
func (r *Repository) CreateOrgJabatan(ctx context.Context, in OrgJabatanInput) (*ent.OrgJabatan, error) {
	q := r.client.OrgJabatan.Create().
		SetParokiID(in.ParokiID).
		SetOrgID(in.OrgID).
		SetPeriodeID(in.PeriodeID).
		SetJabatan(in.Jabatan).
		SetUrutan(in.Urutan).
		SetActive(in.Active)
	if in.ParentID != nil && *in.ParentID != uuid.Nil {
		q.SetParentID(*in.ParentID)
	}
	if in.UserID != nil && *in.UserID != uuid.Nil {
		q.SetUserID(*in.UserID)
	}
	if in.UmatID != nil && *in.UmatID != uuid.Nil {
		q.SetUmatID(*in.UmatID)
	}
	if in.Nama != nil && *in.Nama != "" {
		q.SetNama(*in.Nama)
	}
	if in.StartDate != nil && *in.StartDate != "" {
		q.SetStartDate(*in.StartDate)
	}
	if in.EndDate != nil && *in.EndDate != "" {
		q.SetEndDate(*in.EndDate)
	}
	return q.Save(ctx)
}

// UpdateOrgJabatan updates mutable fields of an OrgJabatan.
func (r *Repository) UpdateOrgJabatan(
	ctx context.Context,
	id uuid.UUID,
	parentID *uuid.UUID,
	clearParentID bool,
	jabatan *string,
	userID *uuid.UUID,
	clearUserID bool,
	umatID *uuid.UUID,
	clearUmatID bool,
	nama *string,
	clearNama bool,
	startDate *string,
	clearStartDate bool,
	endDate *string,
	clearEndDate bool,
	urutan *int,
	active *bool,
) (*ent.OrgJabatan, error) {
	upd := r.client.OrgJabatan.UpdateOneID(id)
	if clearParentID {
		upd.ClearParentID()
	} else if parentID != nil {
		if *parentID == uuid.Nil {
			upd.ClearParentID()
		} else {
			upd.SetParentID(*parentID)
		}
	}
	if jabatan != nil {
		upd.SetJabatan(*jabatan)
	}
	if clearUserID {
		upd.ClearUserID()
	} else if userID != nil {
		if *userID == uuid.Nil {
			upd.ClearUserID()
		} else {
			upd.SetUserID(*userID)
		}
	}
	if clearUmatID {
		upd.ClearUmatID()
	} else if umatID != nil {
		if *umatID == uuid.Nil {
			upd.ClearUmatID()
		} else {
			upd.SetUmatID(*umatID)
		}
	}
	if clearNama {
		upd.ClearNama()
	} else if nama != nil {
		if *nama == "" {
			upd.ClearNama()
		} else {
			upd.SetNama(*nama)
		}
	}
	if clearStartDate {
		upd.ClearStartDate()
	} else if startDate != nil {
		if *startDate == "" {
			upd.ClearStartDate()
		} else {
			upd.SetStartDate(*startDate)
		}
	}
	if clearEndDate {
		upd.ClearEndDate()
	} else if endDate != nil {
		if *endDate == "" {
			upd.ClearEndDate()
		} else {
			upd.SetEndDate(*endDate)
		}
	}
	if urutan != nil {
		upd.SetUrutan(*urutan)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteOrgJabatan sets deleted_at on an OrgJabatan.
func (r *Repository) SoftDeleteOrgJabatan(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.OrgJabatan.UpdateOneID(id).
		SetDeletedAt(at).
		Exec(ctx)
}
