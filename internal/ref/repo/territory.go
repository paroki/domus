package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/keluarga"
	"github.com/paroki/domus/ent/lingkungan"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/ent/wilayah"
)

// WilayahInput holds fields for creating or reactivating a Wilayah.
type WilayahInput struct {
	ParokiID           uuid.UUID
	Jenis              wilayah.Jenis
	Kode               string
	Nama               string
	Koordinator        *string
	TeleponKoordinator *string
}

// LingkunganInput holds fields for creating or reactivating a Lingkungan.
type LingkunganInput struct {
	ParokiID     uuid.UUID
	WilayahID    uuid.UUID
	Kode         string
	Nama         string
	Pelindung    *string
	Ketua        *string
	TeleponKetua *string
}

// SearchWilayah returns paginated non-deleted wilayah for parokiID.
func (r *Repository) SearchWilayah(
	ctx context.Context,
	parokiID uuid.UUID,
	q, jenis string,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Wilayah, error) {
	query := r.client.Wilayah.Query().Where(wilayah.ParokiIDEQ(parokiID), wilayah.DeletedAtIsNil())
	if q != "" {
		query = query.Where(wilayah.Or(
			wilayah.KodeContainsFold(q),
			wilayah.NamaContainsFold(q),
		))
	}
	if jenis != "" {
		query = query.Where(wilayah.JenisEQ(wilayah.Jenis(jenis)))
	}
	if active != nil {
		query = query.Where(wilayah.ActiveEQ(*active))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(wilayah.Or(
			wilayah.CreatedAtLT(*cursorCreatedAt),
			wilayah.And(
				wilayah.CreatedAtEQ(*cursorCreatedAt),
				wilayah.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(wilayah.FieldCreatedAt), ent.Desc(wilayah.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindWilayahByKodeIncludeDeleted returns a Wilayah by (parokiID, kode), including soft-deleted rows.
func (r *Repository) FindWilayahByKodeIncludeDeleted(ctx context.Context, parokiID uuid.UUID, kode string) (*ent.Wilayah, error) {
	return r.client.Wilayah.Query().
		Where(wilayah.ParokiIDEQ(parokiID), wilayah.KodeEQ(kode)).
		Only(ctx)
}

// CreateWilayah inserts a new Wilayah.
func (r *Repository) CreateWilayah(ctx context.Context, in WilayahInput) (*ent.Wilayah, error) {
	return r.client.Wilayah.Create().
		SetParokiID(in.ParokiID).
		SetJenis(in.Jenis).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetNillableKoordinator(in.Koordinator).
		SetNillableTeleponKoordinator(in.TeleponKoordinator).
		SetActive(true).
		Save(ctx)
}

// ReactivateWilayah restores a soft-deleted Wilayah with fresh fields.
func (r *Repository) ReactivateWilayah(ctx context.Context, id uuid.UUID, in WilayahInput) (*ent.Wilayah, error) {
	upd := r.client.Wilayah.UpdateOneID(id).
		ClearDeletedAt().
		SetJenis(in.Jenis).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(true)
	setOrClearWilayahString(in.Koordinator, upd.SetKoordinator, upd.ClearKoordinator)
	setOrClearWilayahString(in.TeleponKoordinator, upd.SetTeleponKoordinator, upd.ClearTeleponKoordinator)
	return upd.Save(ctx)
}

// UpdateWilayah updates mutable fields of a Wilayah.
func (r *Repository) UpdateWilayah(
	ctx context.Context,
	id uuid.UUID,
	jenis *wilayah.Jenis,
	kode, nama, koordinator, teleponKoordinator *string,
	active *bool,
) (*ent.Wilayah, error) {
	upd := r.client.Wilayah.UpdateOneID(id)
	if jenis != nil {
		upd.SetJenis(*jenis)
	}
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if koordinator != nil {
		setOrClearWilayahString(koordinator, upd.SetKoordinator, upd.ClearKoordinator)
	}
	if teleponKoordinator != nil {
		setOrClearWilayahString(teleponKoordinator, upd.SetTeleponKoordinator, upd.ClearTeleponKoordinator)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteWilayah sets deleted_at and marks the Wilayah inactive.
func (r *Repository) SoftDeleteWilayah(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Wilayah.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasLingkunganByWilayah reports whether any non-deleted Lingkungan references wilayahID in parokiID.
func (r *Repository) HasLingkunganByWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID) (bool, error) {
	return r.client.Lingkungan.Query().
		Where(lingkungan.ParokiIDEQ(parokiID), lingkungan.WilayahIDEQ(wilayahID), lingkungan.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrganizationsByWilayah reports whether any non-deleted Organization references wilayahID in parokiID.
func (r *Repository) HasOrganizationsByWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID) (bool, error) {
	return r.client.Organization.Query().
		Where(organization.ParokiIDEQ(parokiID), organization.WilayahIDEQ(wilayahID), organization.DeletedAtIsNil()).
		Exist(ctx)
}

// HasOrganizationsByLingkungan reports whether any non-deleted Organization references lingkunganID in parokiID.
func (r *Repository) HasOrganizationsByLingkungan(ctx context.Context, parokiID, lingkunganID uuid.UUID) (bool, error) {
	return r.client.Organization.Query().
		Where(organization.ParokiIDEQ(parokiID), organization.LingkunganIDEQ(lingkunganID), organization.DeletedAtIsNil()).
		Exist(ctx)
}

// HasKeluargaByWilayah reports whether any non-deleted Keluarga references wilayahID in parokiID.
func (r *Repository) HasKeluargaByWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID) (bool, error) {
	return r.client.Keluarga.Query().
		Where(keluarga.ParokiIDEQ(parokiID), keluarga.WilayahIDEQ(wilayahID), keluarga.DeletedAtIsNil()).
		Exist(ctx)
}

// HasUmatByWilayah reports whether any non-deleted Umat references wilayahID in parokiID.
func (r *Repository) HasUmatByWilayah(ctx context.Context, parokiID, wilayahID uuid.UUID) (bool, error) {
	return r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.WilayahIDEQ(wilayahID), umat.DeletedAtIsNil()).
		Exist(ctx)
}

// HasKeluargaByLingkungan reports whether any non-deleted Keluarga references lingkunganID in parokiID.
func (r *Repository) HasKeluargaByLingkungan(ctx context.Context, parokiID, lingkunganID uuid.UUID) (bool, error) {
	return r.client.Keluarga.Query().
		Where(keluarga.ParokiIDEQ(parokiID), keluarga.LingkunganIDEQ(lingkunganID), keluarga.DeletedAtIsNil()).
		Exist(ctx)
}

// HasUmatByLingkungan reports whether any non-deleted Umat references lingkunganID in parokiID.
func (r *Repository) HasUmatByLingkungan(ctx context.Context, parokiID, lingkunganID uuid.UUID) (bool, error) {
	return r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.LingkunganIDEQ(lingkunganID), umat.DeletedAtIsNil()).
		Exist(ctx)
}

// SearchLingkungan returns paginated non-deleted lingkungan for parokiID.
func (r *Repository) SearchLingkungan(
	ctx context.Context,
	parokiID uuid.UUID,
	q string,
	wilayahID *uuid.UUID,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Lingkungan, error) {
	query := r.client.Lingkungan.Query().Where(lingkungan.ParokiIDEQ(parokiID), lingkungan.DeletedAtIsNil())
	if q != "" {
		query = query.Where(lingkungan.Or(
			lingkungan.KodeContainsFold(q),
			lingkungan.NamaContainsFold(q),
		))
	}
	if wilayahID != nil && *wilayahID != uuid.Nil {
		query = query.Where(lingkungan.WilayahIDEQ(*wilayahID))
	}
	if active != nil {
		query = query.Where(lingkungan.ActiveEQ(*active))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(lingkungan.Or(
			lingkungan.CreatedAtLT(*cursorCreatedAt),
			lingkungan.And(
				lingkungan.CreatedAtEQ(*cursorCreatedAt),
				lingkungan.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(lingkungan.FieldCreatedAt), ent.Desc(lingkungan.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindLingkunganByKodeIncludeDeleted returns a Lingkungan by (parokiID, kode), including soft-deleted rows.
func (r *Repository) FindLingkunganByKodeIncludeDeleted(ctx context.Context, parokiID uuid.UUID, kode string) (*ent.Lingkungan, error) {
	return r.client.Lingkungan.Query().
		Where(lingkungan.ParokiIDEQ(parokiID), lingkungan.KodeEQ(kode)).
		Only(ctx)
}

// CreateLingkungan inserts a new Lingkungan.
func (r *Repository) CreateLingkungan(ctx context.Context, in LingkunganInput) (*ent.Lingkungan, error) {
	return r.client.Lingkungan.Create().
		SetParokiID(in.ParokiID).
		SetWilayahID(in.WilayahID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetNillablePelindung(in.Pelindung).
		SetNillableKetua(in.Ketua).
		SetNillableTeleponKetua(in.TeleponKetua).
		SetActive(true).
		Save(ctx)
}

// ReactivateLingkungan restores a soft-deleted Lingkungan with fresh fields.
func (r *Repository) ReactivateLingkungan(ctx context.Context, id uuid.UUID, in LingkunganInput) (*ent.Lingkungan, error) {
	upd := r.client.Lingkungan.UpdateOneID(id).
		ClearDeletedAt().
		SetWilayahID(in.WilayahID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(true)
	setOrClearLingkunganString(in.Pelindung, upd.SetPelindung, upd.ClearPelindung)
	setOrClearLingkunganString(in.Ketua, upd.SetKetua, upd.ClearKetua)
	setOrClearLingkunganString(in.TeleponKetua, upd.SetTeleponKetua, upd.ClearTeleponKetua)
	return upd.Save(ctx)
}

// UpdateLingkungan updates mutable fields of a Lingkungan.
func (r *Repository) UpdateLingkungan(
	ctx context.Context,
	id uuid.UUID,
	wilayahID *uuid.UUID,
	kode, nama, pelindung, ketua, teleponKetua *string,
	active *bool,
) (*ent.Lingkungan, error) {
	upd := r.client.Lingkungan.UpdateOneID(id)
	if wilayahID != nil {
		upd.SetWilayahID(*wilayahID)
	}
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if pelindung != nil {
		setOrClearLingkunganString(pelindung, upd.SetPelindung, upd.ClearPelindung)
	}
	if ketua != nil {
		setOrClearLingkunganString(ketua, upd.SetKetua, upd.ClearKetua)
	}
	if teleponKetua != nil {
		setOrClearLingkunganString(teleponKetua, upd.SetTeleponKetua, upd.ClearTeleponKetua)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteLingkungan sets deleted_at and marks the Lingkungan inactive.
func (r *Repository) SoftDeleteLingkungan(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Lingkungan.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

func setOrClearWilayahString(val *string, setFn func(string) *ent.WilayahUpdateOne, clearFn func() *ent.WilayahUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}

func setOrClearLingkunganString(val *string, setFn func(string) *ent.LingkunganUpdateOne, clearFn func() *ent.LingkunganUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
