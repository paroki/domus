package cantus

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/cantus/handler"
	"github.com/paroki/domus/internal/cantus/repo"
	"github.com/paroki/domus/internal/cantus/service"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/user"
)

// Re-export cross-feature DTOs and constants.
type (
	Cantus      = domain.CantusResponse
	CantusMedia = domain.CantusMedia
)

// Feature-owned Casbin role and resource permission constants.
var (
	RoleAdmin  = domain.RoleAdmin
	PermCantus = domain.PermCantus
)

// Feature wires the global cross-parish cantus catalog.
type Feature struct {
	service  *service.Service
	handler  *handler.Handler
	enforcer *authz.Enforcer
	users    *user.Feature
}

// New constructs the cantus Feature.
func New(db *ent.Client, enforcer *authz.Enforcer, users *user.Feature) *Feature {
	r := repo.New(db)
	svc := service.New(r, enforcer)
	h := handler.New(svc)
	return &Feature{
		service:  svc,
		handler:  h,
		enforcer: enforcer,
		users:    users,
	}
}

// Register mounts /cantus routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		f.users.Require(PermCantus, authz.ActionRead),
		f.requireWriteMiddleware(),
	)
}

func (f *Feature) requireWriteMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		p, ok := user.UserContext(c.Context())
		if !ok || p.UserID == uuid.Nil || f.enforcer == nil {
			return domain.ErrCantusForbidden
		}
		allowed, err := f.enforcer.Enforce(authz.UserSubject(p.UserID), authz.DomainGlobal, PermCantus, authz.ActionWrite)
		if err != nil {
			return err
		}
		if !allowed {
			return domain.ErrCantusForbidden
		}
		return c.Next()
	}
}

// RegisterPermissions registers cantus feature roles and policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	e.RegisterRole(authz.RoleDescriptor{
		Code:        RoleAdmin,
		Name:        "Admin Nyanyian",
		Scope:       authz.RoleScopeGlobal,
		Description: "Mengelola inventaris lagu, mazmur, dan BPI global lintas-paroki.",
	})
	policies := []struct {
		role string
		act  authz.Action
	}{
		{user.RoleUser, authz.ActionRead},
		{RoleAdmin, authz.ActionRead},
		{RoleAdmin, authz.ActionWrite},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, p.role, authz.DomainGlobal, PermCantus, p.act); err != nil {
			return err
		}
	}
	return nil
}

// GetCantus returns a single non-deleted Cantus entry by ID for cross-feature consumers.
func (f *Feature) GetCantus(ctx context.Context, id uuid.UUID) (Cantus, error) {
	return f.service.Get(ctx, uuid.Nil, id)
}

// SetInUseChecker registers a callback checked before soft-deleting a cantus entry.
func (f *Feature) SetInUseChecker(fn func(ctx context.Context, cantusID uuid.UUID) (bool, error)) {
	f.service.SetInUseChecker(fn)
}
