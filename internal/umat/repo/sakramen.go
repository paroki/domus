package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umatsakramen"
)

// FindSakramenByID returns an UmatSakramen row belonging to (parokiID, umatID, sakramenID).
func (r *Repository) FindSakramenByID(ctx context.Context, parokiID, umatID, sakramenID uuid.UUID) (*ent.UmatSakramen, error) {
	return r.client.UmatSakramen.Query().
		Where(
			umatsakramen.IDEQ(sakramenID),
			umatsakramen.ParokiIDEQ(parokiID),
			umatsakramen.UmatIDEQ(umatID),
		).
		Only(ctx)
}

// FindSakramenByJenis returns an UmatSakramen row belonging to (parokiID, umatID, jenis).
func (r *Repository) FindSakramenByJenis(ctx context.Context, parokiID, umatID uuid.UUID, jenis umatsakramen.Jenis) (*ent.UmatSakramen, error) {
	return r.client.UmatSakramen.Query().
		Where(
			umatsakramen.ParokiIDEQ(parokiID),
			umatsakramen.UmatIDEQ(umatID),
			umatsakramen.JenisEQ(jenis),
		).
		Only(ctx)
}

// HasSakramenByJenis reports whether an UmatSakramen row exists for (parokiID, umatID, jenis) excluding excludeID.
func (r *Repository) HasSakramenByJenis(ctx context.Context, parokiID, umatID uuid.UUID, jenis umatsakramen.Jenis, excludeID *uuid.UUID) (bool, error) {
	q := r.client.UmatSakramen.Query().
		Where(
			umatsakramen.ParokiIDEQ(parokiID),
			umatsakramen.UmatIDEQ(umatID),
			umatsakramen.JenisEQ(jenis),
		)
	if excludeID != nil {
		q = q.Where(umatsakramen.IDNEQ(*excludeID))
	}
	return q.Exist(ctx)
}

// CreateSakramen inserts a new UmatSakramen row.
func (r *Repository) CreateSakramen(ctx context.Context, build func(*ent.UmatSakramenCreate) *ent.UmatSakramenCreate) (*ent.UmatSakramen, error) {
	return build(r.client.UmatSakramen.Create()).Save(ctx)
}

// UpdateSakramen updates an existing UmatSakramen row by ID.
func (r *Repository) UpdateSakramen(ctx context.Context, id uuid.UUID, build func(*ent.UmatSakramenUpdateOne) *ent.UmatSakramenUpdateOne) (*ent.UmatSakramen, error) {
	return build(r.client.UmatSakramen.UpdateOneID(id)).Save(ctx)
}

// DeleteSakramen physically hard-deletes an UmatSakramen row by ID.
func (r *Repository) DeleteSakramen(ctx context.Context, id uuid.UUID) error {
	return r.client.UmatSakramen.DeleteOneID(id).Exec(ctx)
}
