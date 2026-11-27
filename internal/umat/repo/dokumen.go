package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umatdokumen"
)

// FindDokumenByID returns a UmatDokumen by ID within parokiID, with Umat and Keluarga edges.
func (r *Repository) FindDokumenByID(ctx context.Context, parokiID, id uuid.UUID) (*ent.UmatDokumen, error) {
	return r.client.UmatDokumen.Query().
		Where(umatdokumen.IDEQ(id), umatdokumen.ParokiIDEQ(parokiID)).
		WithUmat().
		WithKeluarga().
		Only(ctx)
}

// DeleteDokumen hard-deletes a UmatDokumen row by ID within parokiID.
func (r *Repository) DeleteDokumen(ctx context.Context, parokiID, id uuid.UUID) error {
	_, err := r.client.UmatDokumen.Delete().
		Where(umatdokumen.IDEQ(id), umatdokumen.ParokiIDEQ(parokiID)).
		Exec(ctx)
	return err
}

// ListDokumenByUmatIDs returns document metadata rows for the given umatIDs in parokiID.
func (r *Repository) ListDokumenByUmatIDs(ctx context.Context, parokiID uuid.UUID, umatIDs []uuid.UUID) ([]*ent.UmatDokumen, error) {
	if len(umatIDs) == 0 {
		return nil, nil
	}
	return r.client.UmatDokumen.Query().
		Where(umatdokumen.ParokiIDEQ(parokiID), umatdokumen.UmatIDIn(umatIDs...)).
		Select(
			umatdokumen.FieldID,
			umatdokumen.FieldParokiID,
			umatdokumen.FieldUmatID,
			umatdokumen.FieldKeluargaID,
			umatdokumen.FieldJenis,
			umatdokumen.FieldMimeType,
			umatdokumen.FieldSizeBytes,
			umatdokumen.FieldUploadedBy,
			umatdokumen.FieldCreatedAt,
			umatdokumen.FieldUpdatedAt,
		).
		Order(ent.Asc(umatdokumen.FieldJenis), ent.Asc(umatdokumen.FieldID)).
		All(ctx)
}

// ListDokumenByKeluargaID returns document metadata rows for keluargaID in parokiID.
func (r *Repository) ListDokumenByKeluargaID(ctx context.Context, parokiID, keluargaID uuid.UUID) ([]*ent.UmatDokumen, error) {
	return r.client.UmatDokumen.Query().
		Where(umatdokumen.ParokiIDEQ(parokiID), umatdokumen.KeluargaIDEQ(keluargaID)).
		Select(
			umatdokumen.FieldID,
			umatdokumen.FieldParokiID,
			umatdokumen.FieldUmatID,
			umatdokumen.FieldKeluargaID,
			umatdokumen.FieldJenis,
			umatdokumen.FieldMimeType,
			umatdokumen.FieldSizeBytes,
			umatdokumen.FieldUploadedBy,
			umatdokumen.FieldCreatedAt,
			umatdokumen.FieldUpdatedAt,
		).
		Order(ent.Asc(umatdokumen.FieldJenis), ent.Asc(umatdokumen.FieldID)).
		All(ctx)
}

// UpsertDokumen inserts or updates an encrypted UmatDokumen keyed by (paroki_id, umat_id, jenis) or (paroki_id, keluarga_id, jenis).
func (r *Repository) UpsertDokumen(
	ctx context.Context,
	parokiID uuid.UUID,
	umatID, keluargaID *uuid.UUID,
	jenis umatdokumen.Jenis,
	mimeType umatdokumen.MimeType,
	sizeBytes int,
	dataEnc []byte,
	uploadedBy uuid.UUID,
) (*ent.UmatDokumen, error) {
	q := r.client.UmatDokumen.Query().Where(
		umatdokumen.ParokiIDEQ(parokiID),
		umatdokumen.JenisEQ(jenis),
	)
	if umatID != nil {
		q = q.Where(umatdokumen.UmatIDEQ(*umatID))
	} else if keluargaID != nil {
		q = q.Where(umatdokumen.KeluargaIDEQ(*keluargaID))
	}
	existing, err := q.Only(ctx)
	if err == nil {
		return r.client.UmatDokumen.UpdateOneID(existing.ID).
			SetMimeType(mimeType).
			SetSizeBytes(sizeBytes).
			SetDataEnc(dataEnc).
			SetUploadedBy(uploadedBy).
			Save(ctx)
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return r.client.UmatDokumen.Create().
		SetParokiID(parokiID).
		SetNillableUmatID(umatID).
		SetNillableKeluargaID(keluargaID).
		SetJenis(jenis).
		SetMimeType(mimeType).
		SetSizeBytes(sizeBytes).
		SetDataEnc(dataEnc).
		SetUploadedBy(uploadedBy).
		Save(ctx)
}
