package repo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/keluarga"
	"github.com/paroki/domus/ent/umat"
)

// FindKeluargaByNomorKKHash returns an active Keluarga matching nomorKKHash in parokiID.
func (r *Repository) FindKeluargaByNomorKKHash(ctx context.Context, parokiID uuid.UUID, hash string) (*ent.Keluarga, error) {
	return r.client.Keluarga.Query().
		Where(keluarga.ParokiIDEQ(parokiID), keluarga.NomorKkHashEQ(hash), keluarga.DeletedAtIsNil()).
		Only(ctx)
}

// FindUmatByNIKHash returns an active Umat matching nikHash in parokiID.
func (r *Repository) FindUmatByNIKHash(ctx context.Context, parokiID uuid.UUID, hash string) (*ent.Umat, error) {
	return r.client.Umat.Query().
		Where(umat.ParokiIDEQ(parokiID), umat.NikHashEQ(hash), umat.DeletedAtIsNil()).
		Only(ctx)
}

// FindUmatByNameAndBirthDate returns an active Umat matching case-insensitive nama and tanggalLahir in parokiID.
func (r *Repository) FindUmatByNameAndBirthDate(ctx context.Context, parokiID uuid.UUID, nama, tanggalLahir string) (*ent.Umat, error) {
	return r.client.Umat.Query().
		Where(
			umat.ParokiIDEQ(parokiID),
			umat.NamaEqualFold(strings.TrimSpace(nama)),
			umat.TanggalLahirEQ(strings.TrimSpace(tanggalLahir)),
			umat.DeletedAtIsNil(),
		).
		First(ctx)
}
