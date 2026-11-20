package ref

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/handler"
	"github.com/paroki/domus/internal/ref/repo"
	"github.com/paroki/domus/internal/ref/service"
	"github.com/paroki/domus/internal/user"
)

// Re-export read-only DTO types for downstream features.
type (
	Diocese      = domain.DioceseResponse
	Parish       = domain.ParishResponse
	ParishMember = domain.ParishMemberResponse
	Zone         = domain.ZoneResponse
	Ward         = domain.WardResponse
	Lookup       = domain.LookupResponse
	Period       = domain.PeriodResponse
)

// Feature-owned Casbin roles and resource permissions.
var (
	RoleParishAdmin = authz.Role("ref", "parish-admin")
	PermDioceses    = authz.Permission("ref", "dioceses")
	PermParishes    = authz.Permission("ref", "parishes")
	PermMasterData  = authz.Permission("ref", "master-data")
)

type userEnsurerAdapter struct {
	users *user.Feature
}

func (a userEnsurerAdapter) EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (uuid.UUID, error) {
	u, err := a.users.EnsureUser(ctx, email, name, phone, provider, subject)
	if err != nil {
		return uuid.Nil, err
	}
	return u.ID, nil
}

// Feature wires reference master data (dioceses, parishes, members) and cross-feature helpers.
type Feature struct {
	service *service.Service
	handler *handler.Handler
	users   *user.Feature
}

// New constructs the ref Feature.
func New(db *ent.Client, enforcer *authz.Enforcer, users *user.Feature) *Feature {
	r := repo.New(db)
	var ensurer service.UserEnsurer
	if users != nil {
		ensurer = userEnsurerAdapter{users: users}
	}
	svc := service.New(r, enforcer, ensurer)
	h := handler.New(svc)
	return &Feature{
		service: svc,
		handler: h,
		users:   users,
	}
}

// Register mounts /ref/dioceses and /ref/parishes routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(PermDioceses, authz.ActionRead),
		f.users.Require(PermDioceses, authz.ActionWrite),
		f.users.Require(PermParishes, authz.ActionRead),
		f.users.Require(PermParishes, authz.ActionWrite),
		f.users.Require(PermMasterData, authz.ActionRead),
		f.users.Require(PermMasterData, authz.ActionWrite, parishDomainFromContext),
	)
}

// RegisterPermissions registers ref feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	policies := []struct {
		role string
		perm string
		act  authz.Action
	}{
		{user.RoleSuperAdmin, PermDioceses, authz.ActionRead},
		{user.RoleSuperAdmin, PermDioceses, authz.ActionWrite},
		{user.RoleSuperAdmin, PermParishes, authz.ActionRead},
		{user.RoleSuperAdmin, PermParishes, authz.ActionWrite},
		{user.RoleUser, PermDioceses, authz.ActionRead},
		{user.RoleUser, PermParishes, authz.ActionRead},
		{user.RoleUser, PermMasterData, authz.ActionRead},
		{RoleParishAdmin, PermMasterData, authz.ActionRead},
		{RoleParishAdmin, PermMasterData, authz.ActionWrite},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, p.role, authz.DomainGlobal, p.perm, p.act); err != nil {
			return err
		}
	}
	return nil
}

// GetParish returns a non-deleted Parish by ID.
func (f *Feature) GetParish(ctx context.Context, parishID uuid.UUID) (*Parish, error) {
	return f.service.GetParish(ctx, parishID)
}

// ValidateZoneInParish returns a non-deleted Zone if it belongs to parishID.
func (f *Feature) ValidateZoneInParish(ctx context.Context, parishID, zoneID uuid.UUID) (*Zone, error) {
	return f.service.ValidateZoneInParish(ctx, parishID, zoneID)
}

// ValidateWardInParish returns a non-deleted Ward if it belongs to parishID.
func (f *Feature) ValidateWardInParish(ctx context.Context, parishID, wardID uuid.UUID) (*Ward, error) {
	return f.service.ValidateWardInParish(ctx, parishID, wardID)
}

// ValidatePeriodInParish returns a non-deleted Period if it belongs to parishID.
func (f *Feature) ValidatePeriodInParish(ctx context.Context, parishID, periodID uuid.UUID) (*Period, error) {
	return f.service.ValidatePeriodInParish(ctx, parishID, periodID)
}

// EnsureParishMember idempotently ensures userID is a member of parishID.
func (f *Feature) EnsureParishMember(ctx context.Context, parishID, userID uuid.UUID) (*ParishMember, error) {
	return f.service.EnsureParishMember(ctx, parishID, userID)
}

func requireTenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := user.RequireTenant(c.Context()); err != nil {
			return err
		}
		return c.Next()
	}
}

func parishDomainFromContext(c fiber.Ctx) (string, error) {
	p, err := user.RequireTenant(c.Context())
	if err != nil {
		return "", err
	}
	return authz.ParishDomain(p.ParishID), nil
}
