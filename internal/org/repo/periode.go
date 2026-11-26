package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgperiode"
)

// OrgPeriodeInput holds fields for creating or reactivating an OrgPeriode.
type OrgPeriodeInput struct {
	ParokiID       uuid.UUID
	OrgID          uuid.UUID
	Kode           string
	Nama           string
	TanggalMulai   string
	TanggalSelesai string
	Status         orgperiode.Status
	IsCurrent      bool
}

// SearchOrgPeriode returns paginated non-deleted OrgPeriode rows for orgID, ordered by tanggal_mulai DESC, id DESC.
func (r *Repository) SearchOrgPeriode(
	ctx context.Context,
	orgID uuid.UUID,
	q, status string,
	isCurrent *bool,
	cursorTanggalMulai *string,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.OrgPeriode, error) {
	query := r.client.OrgPeriode.Query().
		Where(orgperiode.OrgIDEQ(orgID), orgperiode.DeletedAtIsNil())
	if q != "" {
		query = query.Where(orgperiode.Or(
			orgperiode.KodeContainsFold(q),
			orgperiode.NamaContainsFold(q),
		))
	}
	if status != "" {
		query = query.Where(orgperiode.StatusEQ(orgperiode.Status(status)))
	}
	if isCurrent != nil {
		query = query.Where(orgperiode.IsCurrentEQ(*isCurrent))
	}
	if cursorTanggalMulai != nil && cursorID != nil {
		query = query.Where(orgperiode.Or(
			orgperiode.TanggalMulaiLT(*cursorTanggalMulai),
			orgperiode.And(
				orgperiode.TanggalMulaiEQ(*cursorTanggalMulai),
				orgperiode.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(orgperiode.ByTanggalMulai(sql.OrderDesc()), orgperiode.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

// FindOrgPeriodeInOrg returns a non-deleted OrgPeriode belonging to orgID.
func (r *Repository) FindOrgPeriodeInOrg(ctx context.Context, orgID, periodeID uuid.UUID) (*ent.OrgPeriode, error) {
	return r.client.OrgPeriode.Query().
		Where(orgperiode.IDEQ(periodeID), orgperiode.OrgIDEQ(orgID), orgperiode.DeletedAtIsNil()).
		Only(ctx)
}

// FindOrgPeriodeByKodeIncludeDeleted returns an OrgPeriode by (orgID, kode), including soft-deleted rows.
func (r *Repository) FindOrgPeriodeByKodeIncludeDeleted(ctx context.Context, orgID uuid.UUID, kode string) (*ent.OrgPeriode, error) {
	return r.client.OrgPeriode.Query().
		Where(orgperiode.OrgIDEQ(orgID), orgperiode.KodeEQ(kode)).
		Only(ctx)
}

// HasOverlappingOrgPeriode reports whether any non-deleted OrgPeriode in orgID overlaps [tanggalMulai, tanggalSelesai].
func (r *Repository) HasOverlappingOrgPeriode(
	ctx context.Context,
	orgID uuid.UUID,
	tanggalMulai, tanggalSelesai string,
	excludeID *uuid.UUID,
) (bool, error) {
	query := r.client.OrgPeriode.Query().
		Where(
			orgperiode.OrgIDEQ(orgID),
			orgperiode.DeletedAtIsNil(),
			orgperiode.TanggalMulaiLTE(tanggalSelesai),
			orgperiode.TanggalSelesaiGTE(tanggalMulai),
		)
	if excludeID != nil {
		query = query.Where(orgperiode.IDNEQ(*excludeID))
	}
	return query.Exist(ctx)
}

// ClearCurrentOrgPeriode sets is_current = false on sibling OrgPeriode rows in orgID.
func (r *Repository) ClearCurrentOrgPeriode(ctx context.Context, orgID uuid.UUID, excludeID *uuid.UUID) error {
	upd := r.client.OrgPeriode.Update().
		Where(
			orgperiode.OrgIDEQ(orgID),
			orgperiode.IsCurrentEQ(true),
			orgperiode.DeletedAtIsNil(),
		)
	if excludeID != nil {
		upd = upd.Where(orgperiode.IDNEQ(*excludeID))
	}
	return upd.SetIsCurrent(false).Exec(ctx)
}

// CreateOrgPeriode inserts a new OrgPeriode.
func (r *Repository) CreateOrgPeriode(ctx context.Context, in OrgPeriodeInput) (*ent.OrgPeriode, error) {
	return r.client.OrgPeriode.Create().
		SetParokiID(in.ParokiID).
		SetOrgID(in.OrgID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetTanggalMulai(in.TanggalMulai).
		SetTanggalSelesai(in.TanggalSelesai).
		SetStatus(in.Status).
		SetIsCurrent(in.IsCurrent).
		Save(ctx)
}

// ReactivateOrgPeriode restores a soft-deleted OrgPeriode with fresh fields.
func (r *Repository) ReactivateOrgPeriode(ctx context.Context, id uuid.UUID, in OrgPeriodeInput) (*ent.OrgPeriode, error) {
	return r.client.OrgPeriode.UpdateOneID(id).
		ClearDeletedAt().
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetTanggalMulai(in.TanggalMulai).
		SetTanggalSelesai(in.TanggalSelesai).
		SetStatus(in.Status).
		SetIsCurrent(in.IsCurrent).
		Save(ctx)
}

// UpdateOrgPeriode updates mutable fields of an OrgPeriode.
func (r *Repository) UpdateOrgPeriode(
	ctx context.Context,
	id uuid.UUID,
	kode, nama, tanggalMulai, tanggalSelesai *string,
	status *orgperiode.Status,
	isCurrent *bool,
) (*ent.OrgPeriode, error) {
	upd := r.client.OrgPeriode.UpdateOneID(id)
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if tanggalMulai != nil {
		upd.SetTanggalMulai(*tanggalMulai)
	}
	if tanggalSelesai != nil {
		upd.SetTanggalSelesai(*tanggalSelesai)
	}
	if status != nil {
		upd.SetStatus(*status)
	}
	if isCurrent != nil {
		upd.SetIsCurrent(*isCurrent)
	}
	return upd.Save(ctx)
}

// SoftDeleteOrgPeriode sets deleted_at and clears is_current on an OrgPeriode.
func (r *Repository) SoftDeleteOrgPeriode(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.OrgPeriode.UpdateOneID(id).
		SetDeletedAt(at).
		SetIsCurrent(false).
		Exec(ctx)
}
