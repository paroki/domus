package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/keluarga"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/ent/umatreferensi"
)

// UmatReferensiInput holds fields for creating or reactivating a UmatReferensi.
type UmatReferensiInput struct {
	ParokiID  *uuid.UUID
	Category  string
	Code      string
	Label     string
	SortOrder int
}

// ListGlobalReferensi returns non-deleted global umat referensi rows matching optional filters.
func (r *Repository) ListGlobalReferensi(ctx context.Context, q, category string, active *bool) ([]*ent.UmatReferensi, error) {
	query := r.client.UmatReferensi.Query().Where(umatreferensi.ParokiIDIsNil(), umatreferensi.DeletedAtIsNil())
	query = applyReferensiFilters(query, q, category, active)
	return query.Order(ent.Desc(umatreferensi.FieldCreatedAt), ent.Desc(umatreferensi.FieldID)).All(ctx)
}

// ListParokiReferensi returns non-deleted tenant umat referensi rows for parokiID matching optional filters.
func (r *Repository) ListParokiReferensi(ctx context.Context, parokiID uuid.UUID, q, category string, active *bool) ([]*ent.UmatReferensi, error) {
	query := r.client.UmatReferensi.Query().Where(umatreferensi.ParokiIDEQ(parokiID), umatreferensi.DeletedAtIsNil())
	query = applyReferensiFilters(query, q, category, active)
	return query.Order(ent.Desc(umatreferensi.FieldCreatedAt), ent.Desc(umatreferensi.FieldID)).All(ctx)
}

func applyReferensiFilters(query *ent.UmatReferensiQuery, q, category string, active *bool) *ent.UmatReferensiQuery {
	if q != "" {
		query = query.Where(umatreferensi.Or(
			umatreferensi.CodeContainsFold(q),
			umatreferensi.LabelContainsFold(q),
			umatreferensi.CategoryContainsFold(q),
		))
	}
	if category != "" {
		query = query.Where(umatreferensi.CategoryEQ(category))
	}
	if active != nil {
		query = query.Where(umatreferensi.ActiveEQ(*active))
	}
	return query
}

// FindReferensiByID returns a non-deleted UmatReferensi by ID.
func (r *Repository) FindReferensiByID(ctx context.Context, id uuid.UUID) (*ent.UmatReferensi, error) {
	return r.client.UmatReferensi.Query().
		Where(umatreferensi.IDEQ(id), umatreferensi.DeletedAtIsNil()).
		Only(ctx)
}

// FindReferensiByKeyIncludeDeleted returns a UmatReferensi by (parokiID, category, code), including soft-deleted rows.
func (r *Repository) FindReferensiByKeyIncludeDeleted(ctx context.Context, parokiID *uuid.UUID, category, code string) (*ent.UmatReferensi, error) {
	query := r.client.UmatReferensi.Query().
		Where(umatreferensi.CategoryEQ(category), umatreferensi.CodeEQ(code))
	if parokiID == nil {
		query = query.Where(umatreferensi.ParokiIDIsNil())
	} else {
		query = query.Where(umatreferensi.ParokiIDEQ(*parokiID))
	}
	return query.Only(ctx)
}

// CreateReferensi inserts a new UmatReferensi.
func (r *Repository) CreateReferensi(ctx context.Context, in UmatReferensiInput) (*ent.UmatReferensi, error) {
	return r.client.UmatReferensi.Create().
		SetNillableParokiID(in.ParokiID).
		SetCategory(in.Category).
		SetCode(in.Code).
		SetLabel(in.Label).
		SetSortOrder(in.SortOrder).
		SetActive(true).
		Save(ctx)
}

// ReactivateReferensi restores a soft-deleted UmatReferensi with fresh fields.
func (r *Repository) ReactivateReferensi(ctx context.Context, id uuid.UUID, in UmatReferensiInput) (*ent.UmatReferensi, error) {
	return r.client.UmatReferensi.UpdateOneID(id).
		ClearDeletedAt().
		SetCategory(in.Category).
		SetCode(in.Code).
		SetLabel(in.Label).
		SetSortOrder(in.SortOrder).
		SetActive(true).
		Save(ctx)
}

// UpdateReferensi updates mutable fields of a UmatReferensi.
func (r *Repository) UpdateReferensi(
	ctx context.Context,
	id uuid.UUID,
	category, code, label *string,
	sortOrder *int,
	active *bool,
) (*ent.UmatReferensi, error) {
	upd := r.client.UmatReferensi.UpdateOneID(id)
	if category != nil {
		upd.SetCategory(*category)
	}
	if code != nil {
		upd.SetCode(*code)
	}
	if label != nil {
		upd.SetLabel(*label)
	}
	if sortOrder != nil {
		upd.SetSortOrder(*sortOrder)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteReferensi sets deleted_at and marks the UmatReferensi inactive.
func (r *Repository) SoftDeleteReferensi(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.UmatReferensi.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// IsReferensiCodeInUse reports whether any non-deleted Keluarga or Umat row in scope references code for category.
func (r *Repository) IsReferensiCodeInUse(ctx context.Context, parokiID *uuid.UUID, category, code string) (bool, error) {
	switch category {
	case "status_rumah", "status_ekonomi":
		q := r.client.Keluarga.Query().Where(keluarga.DeletedAtIsNil())
		if parokiID != nil {
			q = q.Where(keluarga.ParokiIDEQ(*parokiID))
		}
		if category == "status_rumah" {
			q = q.Where(keluarga.StatusRumahEQ(code))
		} else {
			q = q.Where(keluarga.StatusEkonomiEQ(code))
		}
		return q.Exist(ctx)
	default:
		q := r.client.Umat.Query().Where(umat.DeletedAtIsNil())
		if parokiID != nil {
			q = q.Where(umat.ParokiIDEQ(*parokiID))
		}
		switch category {
		case "agama":
			q = q.Where(umat.AgamaEQ(code))
		case "golongan_darah":
			q = q.Where(umat.GolonganDarahEQ(code))
		case "suku":
			q = q.Where(umat.SukuEQ(code))
		case "hubungan_keluarga":
			q = q.Where(umat.HubunganKeluargaEQ(code))
		case "status_pernikahan":
			q = q.Where(umat.StatusPernikahanEQ(code))
		case "pendidikan":
			q = q.Where(umat.PendidikanEQ(code))
		case "pekerjaan":
			q = q.Where(umat.PekerjaanEQ(code))
		case "profesi":
			q = q.Where(umat.ProfesiEQ(code))
		case "disabilitas":
			q = q.Where(umat.DisabilitasEQ(code))
		case "aktivitas_sosial":
			q = q.Where(umat.AktivitasSosialEQ(code))
		default:
			return false, nil
		}
		return q.Exist(ctx)
	}
}
