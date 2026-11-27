package umat

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/handler"
	"github.com/paroki/domus/internal/umat/repo"
	"github.com/paroki/domus/internal/umat/service"
	"github.com/paroki/domus/internal/user"
)

// Re-export read-only DTO types for downstream features.
type UmatBasic = domain.UmatLookupItemResponse

// Feature-owned Casbin roles and resource permissions.
var (
	RoleAdmin     = domain.RoleAdmin
	PermKeluarga  = domain.PermKeluarga
	PermUmat      = domain.PermUmat
	PermReferensi = domain.PermReferensi
)

// Feature wires parishioner (umat) and household (keluarga) administration.
type Feature struct {
	service  *service.Service
	handler  *handler.Handler
	enforcer *authz.Enforcer
	users    *user.Feature
	refs     *ref.Feature
}

// New constructs the umat Feature.
func New(db *ent.Client, enforcer *authz.Enforcer, users *user.Feature, refs *ref.Feature, cipher *crypt.Cipher) *Feature {
	r := repo.New(db)
	var refReader service.RefReader
	if refs != nil {
		refReader = refs
	}
	svc := service.New(r, enforcer, refReader, cipher)
	h := handler.New(svc)
	return &Feature{
		service:  svc,
		handler:  h,
		enforcer: enforcer,
		users:    users,
		refs:     refs,
	}
}

// Register mounts /umat routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(PermKeluarga, authz.ActionRead, parokiDomainFromContext),
		f.users.Require(PermKeluarga, authz.ActionWrite, parokiDomainFromContext),
		f.users.Require(PermUmat, authz.ActionRead, parokiDomainFromContext),
		f.users.Require(PermUmat, authz.ActionWrite, parokiDomainFromContext),
		f.users.Require(PermReferensi, authz.ActionRead),
	)
}

// RegisterPermissions registers umat feature roles and policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	e.RegisterRole(authz.RoleDescriptor{
		Code:        RoleAdmin,
		Name:        "Admin Data Umat",
		Scope:       authz.RoleScopeParoki,
		Description: "Mengelola buku induk kartu keluarga, profil individu umat, dan riwayat sakramen paroki.",
	})
	policies := []struct {
		role string
		perm string
		act  authz.Action
	}{
		{RoleAdmin, PermKeluarga, authz.ActionRead},
		{RoleAdmin, PermKeluarga, authz.ActionWrite},
		{RoleAdmin, PermUmat, authz.ActionRead},
		{RoleAdmin, PermUmat, authz.ActionWrite},
		{RoleAdmin, PermReferensi, authz.ActionRead},
		{RoleAdmin, PermReferensi, authz.ActionWrite},
		{user.RoleUser, PermReferensi, authz.ActionRead},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, p.role, authz.DomainGlobal, p.perm, p.act); err != nil {
			return err
		}
	}
	return nil
}

// GetBasic returns a PII-free Umat summary by ID within parokiID.
func (f *Feature) GetBasic(ctx context.Context, parokiID, umatID uuid.UUID) (*UmatBasic, error) {
	return f.service.GetBasic(ctx, parokiID, umatID)
}

// GetByUserID returns a PII-free Umat summary linked to userID within parokiID, or (nil, nil) if none exists.
func (f *Feature) GetByUserID(ctx context.Context, parokiID, userID uuid.UUID) (*UmatBasic, error) {
	return f.service.GetByUserID(ctx, parokiID, userID)
}

// PreviewVerification checks whether extracted identity data matches an existing parishioner or conflicts with another user.
func (f *Feature) PreviewVerification(ctx context.Context, parokiID, userID uuid.UUID, extraction user.IdentityExtraction) (user.VerifyPreviewResult, error) {
	return f.service.PreviewVerification(ctx, parokiID, userID, extraction)
}

// VerifyAndSyncCensus re-validates reviewed identity data and syncs keluarga, umat, and encrypted documents in one transaction.
func (f *Feature) VerifyAndSyncCensus(ctx context.Context, in user.VerifyConfirmInput) (user.VerifyConfirmResult, error) {
	return f.service.VerifyAndSyncCensus(ctx, in)
}

// GetSelfProfile returns the caller's linked umat, keluarga, members, and document metadata.
func (f *Feature) GetSelfProfile(ctx context.Context, parokiID, userID uuid.UUID) (*user.SelfUmatProfile, error) {
	return f.service.GetSelfProfile(ctx, parokiID, userID)
}

func requireTenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := user.RequireTenant(c.Context()); err != nil {
			return err
		}
		return c.Next()
	}
}

func parokiDomainFromContext(c fiber.Ctx) (string, error) {
	p, err := user.RequireTenant(c.Context())
	if err != nil {
		return "", err
	}
	return authz.ParokiDomain(p.ParokiID), nil
}
