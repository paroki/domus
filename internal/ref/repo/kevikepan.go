package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/kevikepan"
	"github.com/paroki/domus/ent/paroki"
)

// KevikepanInput holds mutable fields for creating or reactivating a Kevikepan.
type KevikepanInput struct {
	KeuskupanID uuid.UUID
	Kode        string
	Nama        string
	Vikep       *string
	Active      bool
}

// SearchKevikepan returns non-deleted kevikepan matching optional filters and cursor.
func (r *Repository) SearchKevikepan(
	ctx context.Context,
	q string,
	keuskupanID *uuid.UUID,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Kevikepan, error) {
	query := r.client.Kevikepan.Query().Where(kevikepan.DeletedAtIsNil())
	if keuskupanID != nil {
		query = query.Where(kevikepan.KeuskupanIDEQ(*keuskupanID))
	}
	if active != nil {
		query = query.Where(kevikepan.ActiveEQ(*active))
	}
	if q != "" {
		query = query.Where(kevikepan.Or(
			kevikepan.KodeContainsFold(q),
			kevikepan.NamaContainsFold(q),
		))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(kevikepan.Or(
			kevikepan.CreatedAtLT(*cursorCreatedAt),
			kevikepan.And(
				kevikepan.CreatedAtEQ(*cursorCreatedAt),
				kevikepan.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(kevikepan.FieldCreatedAt), ent.Desc(kevikepan.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindKevikepanByID returns a non-deleted Kevikepan by ID.
func (r *Repository) FindKevikepanByID(ctx context.Context, id uuid.UUID) (*ent.Kevikepan, error) {
	return r.client.Kevikepan.Query().
		Where(kevikepan.IDEQ(id), kevikepan.DeletedAtIsNil()).
		Only(ctx)
}

// FindKevikepanByKeuskupanAndKodeIncludeDeleted returns a Kevikepan by (keuskupanID, kode), including soft-deleted rows.
func (r *Repository) FindKevikepanByKeuskupanAndKodeIncludeDeleted(ctx context.Context, keuskupanID uuid.UUID, kode string) (*ent.Kevikepan, error) {
	return r.client.Kevikepan.Query().
		Where(kevikepan.KeuskupanIDEQ(keuskupanID), kevikepan.KodeEQ(kode)).
		Only(ctx)
}

// CreateKevikepan inserts a new Kevikepan.
func (r *Repository) CreateKevikepan(ctx context.Context, in KevikepanInput) (*ent.Kevikepan, error) {
	return r.client.Kevikepan.Create().
		SetKeuskupanID(in.KeuskupanID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetNillableVikep(in.Vikep).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateKevikepan restores a soft-deleted Kevikepan with new field values.
func (r *Repository) ReactivateKevikepan(ctx context.Context, id uuid.UUID, in KevikepanInput) (*ent.Kevikepan, error) {
	upd := r.client.Kevikepan.UpdateOneID(id).
		ClearDeletedAt().
		SetKeuskupanID(in.KeuskupanID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(in.Active)
	if in.Vikep == nil || *in.Vikep == "" {
		upd.ClearVikep()
	} else {
		upd.SetVikep(*in.Vikep)
	}
	return upd.Save(ctx)
}

// UpdateKevikepan updates mutable fields of a Kevikepan.
func (r *Repository) UpdateKevikepan(
	ctx context.Context,
	id uuid.UUID,
	keuskupanID *uuid.UUID,
	kode, nama, vikep *string,
	active *bool,
) (*ent.Kevikepan, error) {
	upd := r.client.Kevikepan.UpdateOneID(id)
	if keuskupanID != nil {
		upd.SetKeuskupanID(*keuskupanID)
	}
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if vikep != nil {
		if *vikep == "" {
			upd.ClearVikep()
		} else {
			upd.SetVikep(*vikep)
		}
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteKevikepan sets deleted_at and marks the Kevikepan inactive.
func (r *Repository) SoftDeleteKevikepan(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Kevikepan.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasKevikepanByKeuskupan reports whether any non-deleted Kevikepan references keuskupanID.
func (r *Repository) HasKevikepanByKeuskupan(ctx context.Context, keuskupanID uuid.UUID) (bool, error) {
	return r.client.Kevikepan.Query().
		Where(kevikepan.KeuskupanIDEQ(keuskupanID), kevikepan.DeletedAtIsNil()).
		Exist(ctx)
}

// HasParokiByKevikepan reports whether any non-deleted Paroki references kevikepanID.
func (r *Repository) HasParokiByKevikepan(ctx context.Context, kevikepanID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.KevikepanIDEQ(kevikepanID), paroki.DeletedAtIsNil()).
		Exist(ctx)
}
