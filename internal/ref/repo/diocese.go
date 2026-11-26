package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/keuskupan"
	"github.com/paroki/domus/ent/paroki"
	"github.com/paroki/domus/ent/umatsakramen"
)

// KeuskupanInput holds mutable fields for creating or updating a Keuskupan.
type KeuskupanInput struct {
	Kode    string
	Nama    string
	Uskup   *string
	Alamat  *string
	Telepon *string
	Email   *string
	Website *string
	Active  bool
}

// SearchKeuskupan returns non-deleted keuskupan matching optional filters and cursor.
func (r *Repository) SearchKeuskupan(
	ctx context.Context,
	q string,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Keuskupan, error) {
	query := r.client.Keuskupan.Query().Where(keuskupan.DeletedAtIsNil())
	if q != "" {
		query = query.Where(keuskupan.Or(
			keuskupan.KodeContainsFold(q),
			keuskupan.NamaContainsFold(q),
		))
	}
	if active != nil {
		query = query.Where(keuskupan.ActiveEQ(*active))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(keuskupan.Or(
			keuskupan.CreatedAtLT(*cursorCreatedAt),
			keuskupan.And(
				keuskupan.CreatedAtEQ(*cursorCreatedAt),
				keuskupan.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(keuskupan.FieldCreatedAt), ent.Desc(keuskupan.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindKeuskupanByID returns a non-deleted Keuskupan by ID.
func (r *Repository) FindKeuskupanByID(ctx context.Context, id uuid.UUID) (*ent.Keuskupan, error) {
	return r.client.Keuskupan.Query().
		Where(keuskupan.IDEQ(id), keuskupan.DeletedAtIsNil()).
		Only(ctx)
}

// FindKeuskupanByKodeIncludeDeleted returns a Keuskupan by kode, including soft-deleted rows.
func (r *Repository) FindKeuskupanByKodeIncludeDeleted(ctx context.Context, kode string) (*ent.Keuskupan, error) {
	return r.client.Keuskupan.Query().
		Where(keuskupan.KodeEQ(kode)).
		Only(ctx)
}

// CreateKeuskupan inserts a new Keuskupan.
func (r *Repository) CreateKeuskupan(ctx context.Context, in KeuskupanInput) (*ent.Keuskupan, error) {
	return r.client.Keuskupan.Create().
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetNillableUskup(in.Uskup).
		SetNillableAlamat(in.Alamat).
		SetNillableTelepon(in.Telepon).
		SetNillableEmail(in.Email).
		SetNillableWebsite(in.Website).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateKeuskupan restores a soft-deleted Keuskupan with new field values.
func (r *Repository) ReactivateKeuskupan(ctx context.Context, id uuid.UUID, in KeuskupanInput) (*ent.Keuskupan, error) {
	upd := r.client.Keuskupan.UpdateOneID(id).
		ClearDeletedAt().
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(true)
	setOrClearString(in.Uskup, upd.SetUskup, upd.ClearUskup)
	setOrClearString(in.Alamat, upd.SetAlamat, upd.ClearAlamat)
	setOrClearString(in.Telepon, upd.SetTelepon, upd.ClearTelepon)
	setOrClearString(in.Email, upd.SetEmail, upd.ClearEmail)
	setOrClearString(in.Website, upd.SetWebsite, upd.ClearWebsite)
	return upd.Save(ctx)
}

// UpdateKeuskupan updates mutable fields of a Keuskupan.
func (r *Repository) UpdateKeuskupan(
	ctx context.Context,
	id uuid.UUID,
	kode, nama, uskup, alamat, telepon, email, website *string,
	active *bool,
) (*ent.Keuskupan, error) {
	upd := r.client.Keuskupan.UpdateOneID(id)
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if uskup != nil {
		setOrClearString(uskup, upd.SetUskup, upd.ClearUskup)
	}
	if alamat != nil {
		setOrClearString(alamat, upd.SetAlamat, upd.ClearAlamat)
	}
	if telepon != nil {
		setOrClearString(telepon, upd.SetTelepon, upd.ClearTelepon)
	}
	if email != nil {
		setOrClearString(email, upd.SetEmail, upd.ClearEmail)
	}
	if website != nil {
		setOrClearString(website, upd.SetWebsite, upd.ClearWebsite)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteKeuskupan sets deleted_at and marks the Keuskupan inactive.
func (r *Repository) SoftDeleteKeuskupan(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Keuskupan.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasActiveParokiByKeuskupan reports whether any non-deleted active Paroki references keuskupanID.
func (r *Repository) HasActiveParokiByKeuskupan(ctx context.Context, keuskupanID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.KeuskupanIDEQ(keuskupanID), paroki.ActiveEQ(true), paroki.DeletedAtIsNil()).
		Exist(ctx)
}

// HasParokiByKeuskupan reports whether any non-deleted Paroki references keuskupanID.
func (r *Repository) HasParokiByKeuskupan(ctx context.Context, keuskupanID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.KeuskupanIDEQ(keuskupanID), paroki.DeletedAtIsNil()).
		Exist(ctx)
}

// HasUmatSakramenByKeuskupan reports whether any UmatSakramen references keuskupanID.
func (r *Repository) HasUmatSakramenByKeuskupan(ctx context.Context, keuskupanID uuid.UUID) (bool, error) {
	return r.client.UmatSakramen.Query().
		Where(umatsakramen.KeuskupanIDEQ(keuskupanID)).
		Exist(ctx)
}

func setOrClearString(val *string, setFn func(string) *ent.KeuskupanUpdateOne, clearFn func() *ent.KeuskupanUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
