package service

import (
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/umat/repo"
)

// Service implements business logic for the umat feature.
type Service struct {
	repo     *repo.Repository
	enforcer *authz.Enforcer
	cipher   *crypt.Cipher
}

// New constructs a umat Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, cipher *crypt.Cipher) *Service {
	return &Service{
		repo:     r,
		enforcer: enforcer,
		cipher:   cipher,
	}
}
