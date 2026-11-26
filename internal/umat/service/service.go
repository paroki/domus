package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/umat/repo"
)

// RefReader abstracts cross-feature reference validations from internal/ref.
type RefReader interface {
	ValidateWilayahInParoki(ctx context.Context, parokiID, wilayahID uuid.UUID) (*ref.Wilayah, error)
	ValidateLingkunganInParoki(ctx context.Context, parokiID, lingkunganID uuid.UUID) (*ref.Lingkungan, error)
	ValidateKota(ctx context.Context, kotaID uuid.UUID) (*ref.Kota, error)
	ValidateKecamatan(ctx context.Context, kotaID, kecamatanID uuid.UUID) (*ref.Kecamatan, error)
	ValidateKelurahan(ctx context.Context, kecamatanID, kelurahanID uuid.UUID) (*ref.Kelurahan, error)
	HasActiveLookup(ctx context.Context, parokiID uuid.UUID, category, code string) (bool, error)
	HasParokiMember(ctx context.Context, parokiID, userID uuid.UUID) (bool, error)
}

// Service implements business logic for the umat feature.
type Service struct {
	repo     *repo.Repository
	enforcer *authz.Enforcer
	refs     RefReader
	cipher   *crypt.Cipher
}

// New constructs a umat Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, refs RefReader, cipher *crypt.Cipher) *Service {
	return &Service{
		repo:     r,
		enforcer: enforcer,
		refs:     refs,
		cipher:   cipher,
	}
}
