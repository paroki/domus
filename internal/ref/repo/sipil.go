package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/kecamatan"
	"github.com/paroki/domus/ent/kelurahan"
	"github.com/paroki/domus/ent/kota"
	"github.com/paroki/domus/ent/paroki"
)

// KotaInput holds mutable fields for creating or reactivating a Kota.
type KotaInput struct {
	Kode   string
	Nama   string
	Active bool
}

// KecamatanInput holds mutable fields for creating or reactivating a Kecamatan.
type KecamatanInput struct {
	KotaID uuid.UUID
	Kode   string
	Nama   string
	Active bool
}

// KelurahanInput holds mutable fields for creating or reactivating a Kelurahan.
type KelurahanInput struct {
	KecamatanID uuid.UUID
	Kode        string
	Nama        string
	KodePos     *string
	Active      bool
}

// SearchKota returns non-deleted kota matching optional filters and cursor.
func (r *Repository) SearchKota(
	ctx context.Context,
	q string,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Kota, error) {
	query := r.client.Kota.Query().Where(kota.DeletedAtIsNil())
	if active != nil {
		query = query.Where(kota.ActiveEQ(*active))
	}
	if q != "" {
		query = query.Where(kota.Or(
			kota.KodeContainsFold(q),
			kota.NamaContainsFold(q),
		))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(kota.Or(
			kota.CreatedAtLT(*cursorCreatedAt),
			kota.And(
				kota.CreatedAtEQ(*cursorCreatedAt),
				kota.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(kota.FieldCreatedAt), ent.Desc(kota.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindKotaByID returns a non-deleted Kota by ID.
func (r *Repository) FindKotaByID(ctx context.Context, id uuid.UUID) (*ent.Kota, error) {
	return r.client.Kota.Query().
		Where(kota.IDEQ(id), kota.DeletedAtIsNil()).
		Only(ctx)
}

// FindKotaByKodeIncludeDeleted returns a Kota by kode, including soft-deleted rows.
func (r *Repository) FindKotaByKodeIncludeDeleted(ctx context.Context, kode string) (*ent.Kota, error) {
	return r.client.Kota.Query().
		Where(kota.KodeEQ(kode)).
		Only(ctx)
}

// CreateKota inserts a new Kota.
func (r *Repository) CreateKota(ctx context.Context, in KotaInput) (*ent.Kota, error) {
	return r.client.Kota.Create().
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateKota restores a soft-deleted Kota with new field values.
func (r *Repository) ReactivateKota(ctx context.Context, id uuid.UUID, in KotaInput) (*ent.Kota, error) {
	return r.client.Kota.UpdateOneID(id).
		ClearDeletedAt().
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(in.Active).
		Save(ctx)
}

// UpdateKota updates mutable fields of a Kota.
func (r *Repository) UpdateKota(
	ctx context.Context,
	id uuid.UUID,
	kode, nama *string,
	active *bool,
) (*ent.Kota, error) {
	upd := r.client.Kota.UpdateOneID(id)
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteKota sets deleted_at and marks the Kota inactive.
func (r *Repository) SoftDeleteKota(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Kota.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasKecamatanByKota reports whether any non-deleted Kecamatan references kotaID.
func (r *Repository) HasKecamatanByKota(ctx context.Context, kotaID uuid.UUID) (bool, error) {
	return r.client.Kecamatan.Query().
		Where(kecamatan.KotaIDEQ(kotaID), kecamatan.DeletedAtIsNil()).
		Exist(ctx)
}

// SearchKecamatan returns non-deleted kecamatan matching optional filters and cursor.
func (r *Repository) SearchKecamatan(
	ctx context.Context,
	q string,
	kotaID *uuid.UUID,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Kecamatan, error) {
	query := r.client.Kecamatan.Query().Where(kecamatan.DeletedAtIsNil())
	if kotaID != nil {
		query = query.Where(kecamatan.KotaIDEQ(*kotaID))
	}
	if active != nil {
		query = query.Where(kecamatan.ActiveEQ(*active))
	}
	if q != "" {
		query = query.Where(kecamatan.Or(
			kecamatan.KodeContainsFold(q),
			kecamatan.NamaContainsFold(q),
		))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(kecamatan.Or(
			kecamatan.CreatedAtLT(*cursorCreatedAt),
			kecamatan.And(
				kecamatan.CreatedAtEQ(*cursorCreatedAt),
				kecamatan.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(kecamatan.FieldCreatedAt), ent.Desc(kecamatan.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindKecamatanByID returns a non-deleted Kecamatan by ID.
func (r *Repository) FindKecamatanByID(ctx context.Context, id uuid.UUID) (*ent.Kecamatan, error) {
	return r.client.Kecamatan.Query().
		Where(kecamatan.IDEQ(id), kecamatan.DeletedAtIsNil()).
		Only(ctx)
}

// FindKecamatanInKota returns a non-deleted Kecamatan belonging to kotaID.
func (r *Repository) FindKecamatanInKota(ctx context.Context, kotaID, kecamatanID uuid.UUID) (*ent.Kecamatan, error) {
	return r.client.Kecamatan.Query().
		Where(kecamatan.IDEQ(kecamatanID), kecamatan.KotaIDEQ(kotaID), kecamatan.DeletedAtIsNil()).
		Only(ctx)
}

// FindKecamatanByKotaAndKodeIncludeDeleted returns a Kecamatan by (kotaID, kode), including soft-deleted rows.
func (r *Repository) FindKecamatanByKotaAndKodeIncludeDeleted(ctx context.Context, kotaID uuid.UUID, kode string) (*ent.Kecamatan, error) {
	return r.client.Kecamatan.Query().
		Where(kecamatan.KotaIDEQ(kotaID), kecamatan.KodeEQ(kode)).
		Only(ctx)
}

// CreateKecamatan inserts a new Kecamatan.
func (r *Repository) CreateKecamatan(ctx context.Context, in KecamatanInput) (*ent.Kecamatan, error) {
	return r.client.Kecamatan.Create().
		SetKotaID(in.KotaID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateKecamatan restores a soft-deleted Kecamatan with new field values.
func (r *Repository) ReactivateKecamatan(ctx context.Context, id uuid.UUID, in KecamatanInput) (*ent.Kecamatan, error) {
	return r.client.Kecamatan.UpdateOneID(id).
		ClearDeletedAt().
		SetKotaID(in.KotaID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(in.Active).
		Save(ctx)
}

// UpdateKecamatan updates mutable fields of a Kecamatan.
func (r *Repository) UpdateKecamatan(
	ctx context.Context,
	id uuid.UUID,
	kotaID *uuid.UUID,
	kode, nama *string,
	active *bool,
) (*ent.Kecamatan, error) {
	upd := r.client.Kecamatan.UpdateOneID(id)
	if kotaID != nil {
		upd.SetKotaID(*kotaID)
	}
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteKecamatan sets deleted_at and marks the Kecamatan inactive.
func (r *Repository) SoftDeleteKecamatan(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Kecamatan.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasKelurahanByKecamatan reports whether any non-deleted Kelurahan references kecamatanID.
func (r *Repository) HasKelurahanByKecamatan(ctx context.Context, kecamatanID uuid.UUID) (bool, error) {
	return r.client.Kelurahan.Query().
		Where(kelurahan.KecamatanIDEQ(kecamatanID), kelurahan.DeletedAtIsNil()).
		Exist(ctx)
}

// SearchKelurahan returns non-deleted kelurahan matching optional filters and cursor.
func (r *Repository) SearchKelurahan(
	ctx context.Context,
	q string,
	kecamatanID *uuid.UUID,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Kelurahan, error) {
	query := r.client.Kelurahan.Query().Where(kelurahan.DeletedAtIsNil())
	if kecamatanID != nil {
		query = query.Where(kelurahan.KecamatanIDEQ(*kecamatanID))
	}
	if active != nil {
		query = query.Where(kelurahan.ActiveEQ(*active))
	}
	if q != "" {
		query = query.Where(kelurahan.Or(
			kelurahan.KodeContainsFold(q),
			kelurahan.NamaContainsFold(q),
		))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(kelurahan.Or(
			kelurahan.CreatedAtLT(*cursorCreatedAt),
			kelurahan.And(
				kelurahan.CreatedAtEQ(*cursorCreatedAt),
				kelurahan.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(kelurahan.FieldCreatedAt), ent.Desc(kelurahan.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindKelurahanByID returns a non-deleted Kelurahan by ID.
func (r *Repository) FindKelurahanByID(ctx context.Context, id uuid.UUID) (*ent.Kelurahan, error) {
	return r.client.Kelurahan.Query().
		Where(kelurahan.IDEQ(id), kelurahan.DeletedAtIsNil()).
		Only(ctx)
}

// FindKelurahanInKecamatan returns a non-deleted Kelurahan belonging to kecamatanID.
func (r *Repository) FindKelurahanInKecamatan(ctx context.Context, kecamatanID, kelurahanID uuid.UUID) (*ent.Kelurahan, error) {
	return r.client.Kelurahan.Query().
		Where(kelurahan.IDEQ(kelurahanID), kelurahan.KecamatanIDEQ(kecamatanID), kelurahan.DeletedAtIsNil()).
		Only(ctx)
}

// FindKelurahanByKecamatanAndKodeIncludeDeleted returns a Kelurahan by (kecamatanID, kode), including soft-deleted rows.
func (r *Repository) FindKelurahanByKecamatanAndKodeIncludeDeleted(ctx context.Context, kecamatanID uuid.UUID, kode string) (*ent.Kelurahan, error) {
	return r.client.Kelurahan.Query().
		Where(kelurahan.KecamatanIDEQ(kecamatanID), kelurahan.KodeEQ(kode)).
		Only(ctx)
}

// CreateKelurahan inserts a new Kelurahan.
func (r *Repository) CreateKelurahan(ctx context.Context, in KelurahanInput) (*ent.Kelurahan, error) {
	return r.client.Kelurahan.Create().
		SetKecamatanID(in.KecamatanID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetNillableKodePos(in.KodePos).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateKelurahan restores a soft-deleted Kelurahan with new field values.
func (r *Repository) ReactivateKelurahan(ctx context.Context, id uuid.UUID, in KelurahanInput) (*ent.Kelurahan, error) {
	upd := r.client.Kelurahan.UpdateOneID(id).
		ClearDeletedAt().
		SetKecamatanID(in.KecamatanID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(in.Active)
	if in.KodePos == nil || *in.KodePos == "" {
		upd.ClearKodePos()
	} else {
		upd.SetKodePos(*in.KodePos)
	}
	return upd.Save(ctx)
}

// UpdateKelurahan updates mutable fields of a Kelurahan.
func (r *Repository) UpdateKelurahan(
	ctx context.Context,
	id uuid.UUID,
	kecamatanID *uuid.UUID,
	kode, nama, kodePos *string,
	active *bool,
) (*ent.Kelurahan, error) {
	upd := r.client.Kelurahan.UpdateOneID(id)
	if kecamatanID != nil {
		upd.SetKecamatanID(*kecamatanID)
	}
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if kodePos != nil {
		if *kodePos == "" {
			upd.ClearKodePos()
		} else {
			upd.SetKodePos(*kodePos)
		}
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteKelurahan sets deleted_at and marks the Kelurahan inactive.
func (r *Repository) SoftDeleteKelurahan(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Kelurahan.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasParokiByKota reports whether any non-deleted Paroki references kotaID.
func (r *Repository) HasParokiByKota(ctx context.Context, kotaID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.KotaIDEQ(kotaID), paroki.DeletedAtIsNil()).
		Exist(ctx)
}

// HasParokiByKecamatan reports whether any non-deleted Paroki references kecamatanID.
func (r *Repository) HasParokiByKecamatan(ctx context.Context, kecamatanID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.KecamatanIDEQ(kecamatanID), paroki.DeletedAtIsNil()).
		Exist(ctx)
}

// HasParokiByKelurahan reports whether any non-deleted Paroki references kelurahanID.
func (r *Repository) HasParokiByKelurahan(ctx context.Context, kelurahanID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.KelurahanIDEQ(kelurahanID), paroki.DeletedAtIsNil()).
		Exist(ctx)
}
