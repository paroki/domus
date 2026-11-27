package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/keluarga"
	"github.com/paroki/domus/ent/orgjabatan"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/ent/umatsakramen"
)

// Repository provides Ent data access for the umat feature.
type Repository struct {
	client *ent.Client
}

// New constructs a Repository backed by client.
func New(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// WithTx executes fn inside a database transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(txRepo *Repository) error) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	txRepo := &Repository{client: tx.Client()}
	if err := fn(txRepo); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rerr)
		}
		return err
	}
	return tx.Commit()
}

// KeluargaFilter holds query filters for SearchKeluarga.
type KeluargaFilter struct {
	Q            string
	WilayahID    *uuid.UUID
	LingkunganID *uuid.UUID
	Aktif        *bool
	NomorKKHash  string
	Sort         string
	CursorTime   *time.Time
	CursorName   *string
	CursorID     *uuid.UUID
	Limit        int
}

// UmatFilter holds query filters for SearchUmat.
type UmatFilter struct {
	Q                string
	KeluargaID       *uuid.UUID
	WilayahID        *uuid.UUID
	LingkunganID     *uuid.UUID
	JenisKelamin     string
	Agama            string
	StatusPernikahan string
	Almarhum         *bool
	TanggalLahir     string
	NIKHash          string
	Sort             string
	CursorTime       *time.Time
	CursorName       *string
	CursorID         *uuid.UUID
	Limit            int
}

// SearchKeluarga returns keyset-paginated non-deleted Keluarga rows in parokiID.
func (r *Repository) SearchKeluarga(ctx context.Context, parokiID uuid.UUID, f KeluargaFilter) ([]*ent.Keluarga, error) {
	q := r.client.Keluarga.Query().Where(keluarga.ParokiIDEQ(parokiID), keluarga.DeletedAtIsNil())
	if f.Q != "" {
		q = q.Where(keluarga.Or(
			keluarga.NamaKeluargaContainsFold(f.Q),
			keluarga.AlamatContainsFold(f.Q),
		))
	}
	if f.WilayahID != nil && *f.WilayahID != uuid.Nil {
		q = q.Where(keluarga.WilayahIDEQ(*f.WilayahID))
	}
	if f.LingkunganID != nil && *f.LingkunganID != uuid.Nil {
		q = q.Where(keluarga.LingkunganIDEQ(*f.LingkunganID))
	}
	if f.Aktif != nil {
		q = q.Where(keluarga.AktifEQ(*f.Aktif))
	}
	if f.NomorKKHash != "" {
		q = q.Where(keluarga.NomorKkHashEQ(f.NomorKKHash))
	}

	switch f.Sort {
	case "namaKeluarga":
		if f.CursorName != nil && f.CursorID != nil {
			q = q.Where(keluarga.Or(
				keluarga.NamaKeluargaGT(*f.CursorName),
				keluarga.And(keluarga.NamaKeluargaEQ(*f.CursorName), keluarga.IDGT(*f.CursorID)),
			))
		}
		q = q.Order(ent.Asc(keluarga.FieldNamaKeluarga), ent.Asc(keluarga.FieldID))
	case "-namaKeluarga":
		if f.CursorName != nil && f.CursorID != nil {
			q = q.Where(keluarga.Or(
				keluarga.NamaKeluargaLT(*f.CursorName),
				keluarga.And(keluarga.NamaKeluargaEQ(*f.CursorName), keluarga.IDLT(*f.CursorID)),
			))
		}
		q = q.Order(ent.Desc(keluarga.FieldNamaKeluarga), ent.Desc(keluarga.FieldID))
	default:
		if f.CursorTime != nil && f.CursorID != nil {
			q = q.Where(keluarga.Or(
				keluarga.CreatedAtLT(*f.CursorTime),
				keluarga.And(keluarga.CreatedAtEQ(*f.CursorTime), keluarga.IDLT(*f.CursorID)),
			))
		}
		q = q.Order(ent.Desc(keluarga.FieldCreatedAt), ent.Desc(keluarga.FieldID))
	}
	return q.Limit(f.Limit).All(ctx)
}

// FindKeluargaByID returns a non-deleted Keluarga by ID within parokiID, with active anggota ordered by nama.
func (r *Repository) FindKeluargaByID(ctx context.Context, parokiID, id uuid.UUID) (*ent.Keluarga, error) {
	return r.client.Keluarga.Query().
		Where(keluarga.IDEQ(id), keluarga.ParokiIDEQ(parokiID), keluarga.DeletedAtIsNil()).
		WithAnggota(func(uq *ent.UmatQuery) {
			uq.Where(umat.DeletedAtIsNil()).Order(ent.Asc(umat.FieldNama), ent.Asc(umat.FieldID))
		}).
		Only(ctx)
}

// CreateKeluarga inserts a new Keluarga row.
func (r *Repository) CreateKeluarga(ctx context.Context, build func(*ent.KeluargaCreate) *ent.KeluargaCreate) (*ent.Keluarga, error) {
	return build(r.client.Keluarga.Create()).Save(ctx)
}

// UpdateKeluarga updates a Keluarga row by ID.
func (r *Repository) UpdateKeluarga(ctx context.Context, id uuid.UUID, build func(*ent.KeluargaUpdateOne) *ent.KeluargaUpdateOne) (*ent.Keluarga, error) {
	return build(r.client.Keluarga.UpdateOneID(id)).Save(ctx)
}

// SoftDeleteKeluarga sets deleted_at and aktif=false on a Keluarga.
func (r *Repository) SoftDeleteKeluarga(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Keluarga.UpdateOneID(id).SetDeletedAt(at).SetAktif(false).Exec(ctx)
}

// HasActiveUmatByKeluarga reports whether any non-deleted Umat references keluargaID in parokiID.
func (r *Repository) HasActiveUmatByKeluarga(ctx context.Context, parokiID, keluargaID uuid.UUID) (bool, error) {
	return r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.KeluargaIDEQ(keluargaID), umat.DeletedAtIsNil()).
		Exist(ctx)
}

// SearchUmat returns keyset-paginated non-deleted Umat rows in parokiID.
func (r *Repository) SearchUmat(ctx context.Context, parokiID uuid.UUID, f UmatFilter) ([]*ent.Umat, error) {
	q := r.client.Umat.Query().Where(umat.ParokiIDEQ(parokiID), umat.DeletedAtIsNil())
	if f.Q != "" {
		q = q.Where(umat.Or(
			umat.NamaContainsFold(f.Q),
			umat.NamaBaptisContainsFold(f.Q),
		))
	}
	if f.KeluargaID != nil && *f.KeluargaID != uuid.Nil {
		q = q.Where(umat.KeluargaIDEQ(*f.KeluargaID))
	}
	if f.WilayahID != nil && *f.WilayahID != uuid.Nil {
		q = q.Where(umat.WilayahIDEQ(*f.WilayahID))
	}
	if f.LingkunganID != nil && *f.LingkunganID != uuid.Nil {
		q = q.Where(umat.LingkunganIDEQ(*f.LingkunganID))
	}
	if f.JenisKelamin != "" {
		q = q.Where(umat.JenisKelaminEQ(umat.JenisKelamin(f.JenisKelamin)))
	}
	if f.Agama != "" {
		q = q.Where(umat.AgamaEQ(f.Agama))
	}
	if f.StatusPernikahan != "" {
		q = q.Where(umat.StatusPernikahanEQ(f.StatusPernikahan))
	}
	if f.Almarhum != nil {
		q = q.Where(umat.AlmarhumEQ(*f.Almarhum))
	}
	if f.TanggalLahir != "" {
		q = q.Where(umat.TanggalLahirEQ(f.TanggalLahir))
	}
	if f.NIKHash != "" {
		q = q.Where(umat.NikHashEQ(f.NIKHash))
	}

	switch f.Sort {
	case "nama":
		if f.CursorName != nil && f.CursorID != nil {
			q = q.Where(umat.Or(
				umat.NamaGT(*f.CursorName),
				umat.And(umat.NamaEQ(*f.CursorName), umat.IDGT(*f.CursorID)),
			))
		}
		q = q.Order(ent.Asc(umat.FieldNama), ent.Asc(umat.FieldID))
	case "-nama":
		if f.CursorName != nil && f.CursorID != nil {
			q = q.Where(umat.Or(
				umat.NamaLT(*f.CursorName),
				umat.And(umat.NamaEQ(*f.CursorName), umat.IDLT(*f.CursorID)),
			))
		}
		q = q.Order(ent.Desc(umat.FieldNama), ent.Desc(umat.FieldID))
	default:
		if f.CursorTime != nil && f.CursorID != nil {
			q = q.Where(umat.Or(
				umat.CreatedAtLT(*f.CursorTime),
				umat.And(umat.CreatedAtEQ(*f.CursorTime), umat.IDLT(*f.CursorID)),
			))
		}
		q = q.Order(ent.Desc(umat.FieldCreatedAt), ent.Desc(umat.FieldID))
	}
	return q.Limit(f.Limit).All(ctx)
}

// FindUmatByID returns a non-deleted Umat by ID within parokiID, with active Keluarga and Sakramen edges.
func (r *Repository) FindUmatByID(ctx context.Context, parokiID, id uuid.UUID) (*ent.Umat, error) {
	return r.client.Umat.Query().
		Where(umat.IDEQ(id), umat.ParokiIDEQ(parokiID), umat.DeletedAtIsNil()).
		WithKeluarga(func(kq *ent.KeluargaQuery) {
			kq.Where(keluarga.DeletedAtIsNil())
		}).
		WithSakramen(func(sq *ent.UmatSakramenQuery) {
			sq.Order(ent.Asc(umatsakramen.FieldTanggal), ent.Asc(umatsakramen.FieldID))
		}).
		Only(ctx)
}

// CreateUmat inserts a new Umat row.
func (r *Repository) CreateUmat(ctx context.Context, build func(*ent.UmatCreate) *ent.UmatCreate) (*ent.Umat, error) {
	return build(r.client.Umat.Create()).Save(ctx)
}

// UpdateUmat updates an Umat row by ID.
func (r *Repository) UpdateUmat(ctx context.Context, id uuid.UUID, build func(*ent.UmatUpdateOne) *ent.UmatUpdateOne) (*ent.Umat, error) {
	return build(r.client.Umat.UpdateOneID(id)).Save(ctx)
}

// SoftDeleteUmat sets deleted_at on an Umat.
func (r *Repository) SoftDeleteUmat(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Umat.UpdateOneID(id).SetDeletedAt(at).Exec(ctx)
}

// LookupUmat returns non-deleted, non-deceased Umat rows in parokiID ordered by nama ASC, id ASC.
func (r *Repository) LookupUmat(
	ctx context.Context,
	parokiID uuid.UUID,
	q string,
	wilayahID, lingkunganID *uuid.UUID,
	limit int,
) ([]*ent.Umat, error) {
	query := r.client.Umat.Query().
		Where(
			umat.ParokiIDEQ(parokiID),
			umat.DeletedAtIsNil(),
			umat.AlmarhumEQ(false),
		)
	if q != "" {
		query = query.Where(umat.Or(
			umat.NamaContainsFold(q),
			umat.NamaBaptisContainsFold(q),
		))
	}
	if wilayahID != nil && *wilayahID != uuid.Nil {
		query = query.Where(umat.WilayahIDEQ(*wilayahID))
	}
	if lingkunganID != nil && *lingkunganID != uuid.Nil {
		query = query.Where(umat.LingkunganIDEQ(*lingkunganID))
	}
	return query.
		Order(ent.Asc(umat.FieldNama), ent.Asc(umat.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindUmatBasicByID returns a non-deleted Umat by ID within parokiID.
func (r *Repository) FindUmatBasicByID(ctx context.Context, parokiID, umatID uuid.UUID) (*ent.Umat, error) {
	return r.client.Umat.Query().
		Where(umat.IDEQ(umatID), umat.ParokiIDEQ(parokiID), umat.DeletedAtIsNil()).
		Only(ctx)
}

// FindUmatByUserID returns a non-deleted Umat linked to userID within parokiID.
func (r *Repository) FindUmatByUserID(ctx context.Context, parokiID, userID uuid.UUID) (*ent.Umat, error) {
	return r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.UserIDEQ(userID), umat.DeletedAtIsNil()).
		Only(ctx)
}

// HasActiveOrgAdminInParoki reports whether userID holds an active org:admin membership in parokiID.
func (r *Repository) HasActiveOrgAdminInParoki(ctx context.Context, parokiID, userID uuid.UUID) (bool, error) {
	return r.client.OrgMember.Query().
		Where(
			orgmember.ParokiIDEQ(parokiID),
			orgmember.UserIDEQ(userID),
			orgmember.RoleEQ(orgmember.RoleAdmin),
			orgmember.StatusEQ(orgmember.StatusActive),
		).
		Exist(ctx)
}

// HasOrgJabatanByUmat reports whether any non-deleted OrgJabatan in parokiID references umatID.
func (r *Repository) HasOrgJabatanByUmat(ctx context.Context, parokiID, umatID uuid.UUID) (bool, error) {
	return r.client.OrgJabatan.Query().
		Where(
			orgjabatan.ParokiIDEQ(parokiID),
			orgjabatan.UmatIDEQ(umatID),
			orgjabatan.DeletedAtIsNil(),
		).
		Exist(ctx)
}

// HasActiveNomorKKHash reports whether an active Keluarga in parokiID already uses hash (excluding excludeID if non-nil).
func (r *Repository) HasActiveNomorKKHash(ctx context.Context, parokiID uuid.UUID, hash string, excludeID *uuid.UUID) (bool, error) {
	q := r.client.Keluarga.Query().
		Where(keluarga.ParokiIDEQ(parokiID), keluarga.NomorKkHashEQ(hash), keluarga.DeletedAtIsNil())
	if excludeID != nil {
		q = q.Where(keluarga.IDNEQ(*excludeID))
	}
	return q.Exist(ctx)
}

// HasActiveNIKHash reports whether an active Umat in parokiID already uses hash (excluding excludeID if non-nil).
func (r *Repository) HasActiveNIKHash(ctx context.Context, parokiID uuid.UUID, hash string, excludeID *uuid.UUID) (bool, error) {
	q := r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.NikHashEQ(hash), umat.DeletedAtIsNil())
	if excludeID != nil {
		q = q.Where(umat.IDNEQ(*excludeID))
	}
	return q.Exist(ctx)
}

// HasActiveUserLink reports whether an active Umat in parokiID already links userID (excluding excludeID if non-nil).
func (r *Repository) HasActiveUserLink(ctx context.Context, parokiID, userID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	q := r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.UserIDEQ(userID), umat.DeletedAtIsNil())
	if excludeID != nil {
		q = q.Where(umat.IDNEQ(*excludeID))
	}
	return q.Exist(ctx)
}
