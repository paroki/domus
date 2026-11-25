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
	Keuskupan    = domain.KeuskupanResponse
	Paroki       = domain.ParokiResponse
	ParokiMember = domain.ParokiMemberResponse
	Wilayah      = domain.WilayahResponse
	Lingkungan   = domain.LingkunganResponse
	Lookup       = domain.LookupResponse
	Period       = domain.PeriodResponse
)

// Feature-owned Casbin roles and resource permissions.
var (
	RoleParokiAdmin = authz.Role("ref", "paroki-admin")
	PermKeuskupan   = authz.Permission("ref", "keuskupan")
	PermParoki      = authz.Permission("ref", "paroki")
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
	service  *service.Service
	handler  *handler.Handler
	enforcer *authz.Enforcer
	users    *user.Feature
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
		service:  svc,
		handler:  h,
		enforcer: enforcer,
		users:    users,
	}
}

// Register mounts /ref/keuskupan and /ref/paroki routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(PermKeuskupan, authz.ActionRead),
		f.users.Require(PermKeuskupan, authz.ActionWrite),
		f.users.Require(PermParoki, authz.ActionRead),
		f.users.Require(PermParoki, authz.ActionWrite),
		f.users.Require(PermMasterData, authz.ActionRead),
		f.users.Require(PermMasterData, authz.ActionWrite, parokiDomainFromContext),
	)
}

// RegisterPermissions registers ref feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	e.RegisterRole(authz.RoleDescriptor{
		Code:        RoleParokiAdmin,
		Name:        "Admin Paroki",
		Scope:       authz.RoleScopeParoki,
		Description: "Mengelola profil paroki, wilayah & lingkungan, kamus referensi, periode, dan organisasi paroki.",
	})
	policies := []struct {
		role string
		perm string
		act  authz.Action
	}{
		{user.RoleSuperAdmin, PermKeuskupan, authz.ActionRead},
		{user.RoleSuperAdmin, PermKeuskupan, authz.ActionWrite},
		{user.RoleSuperAdmin, PermParoki, authz.ActionRead},
		{user.RoleSuperAdmin, PermParoki, authz.ActionWrite},
		{user.RoleUser, PermKeuskupan, authz.ActionRead},
		{user.RoleUser, PermParoki, authz.ActionRead},
		{user.RoleUser, PermMasterData, authz.ActionRead},
		{RoleParokiAdmin, PermMasterData, authz.ActionRead},
		{RoleParokiAdmin, PermMasterData, authz.ActionWrite},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, p.role, authz.DomainGlobal, p.perm, p.act); err != nil {
			return err
		}
	}
	return nil
}

// GetParoki returns a non-deleted Paroki by ID.
func (f *Feature) GetParoki(ctx context.Context, parokiID uuid.UUID) (*Paroki, error) {
	return f.service.GetParoki(ctx, parokiID)
}

// GetParokiByCode returns an active, non-deleted Paroki by code.
func (f *Feature) GetParokiByCode(ctx context.Context, code string) (*Paroki, error) {
	return f.service.GetParokiByCode(ctx, code)
}

// ValidateWilayahInParoki returns a non-deleted Wilayah if it belongs to parokiID.
func (f *Feature) ValidateWilayahInParoki(ctx context.Context, parokiID, wilayahID uuid.UUID) (*Wilayah, error) {
	return f.service.ValidateWilayahInParoki(ctx, parokiID, wilayahID)
}

// ValidateLingkunganInParoki returns a non-deleted Lingkungan if it belongs to parokiID.
func (f *Feature) ValidateLingkunganInParoki(ctx context.Context, parokiID, lingkunganID uuid.UUID) (*Lingkungan, error) {
	return f.service.ValidateLingkunganInParoki(ctx, parokiID, lingkunganID)
}

// ValidatePeriodInParoki returns a non-deleted Period if it belongs to parokiID.
func (f *Feature) ValidatePeriodInParoki(ctx context.Context, parokiID, periodID uuid.UUID) (*Period, error) {
	return f.service.ValidatePeriodInParoki(ctx, parokiID, periodID)
}

// EnsureParokiMember idempotently ensures userID is a member of parokiID.
func (f *Feature) EnsureParokiMember(ctx context.Context, parokiID, userID uuid.UUID) (*ParokiMember, error) {
	return f.service.EnsureParokiMember(ctx, parokiID, userID)
}

// ListParokiMembers implements user.ParokiMemberDirectory for MCP tools.
func (f *Feature) ListParokiMembers(ctx context.Context, p user.Principal, parokiID uuid.UUID, query, role string, limit int) ([]user.ParokiMemberItem, error) {
	var filter *domain.ParokiMemberSearchFilter
	if role != "" {
		filter = &domain.ParokiMemberSearchFilter{Role: role}
	}
	rows, _, err := f.service.SearchParokiMembers(
		ctx,
		p,
		parokiID,
		domain.SearchParokiMembersRequest{Q: query, Filter: filter},
		nil,
		nil,
		limit,
	)
	if err != nil {
		return nil, err
	}
	out := make([]user.ParokiMemberItem, 0, len(rows))
	for _, m := range rows {
		out = append(out, user.ParokiMemberItem{
			ID:       m.ID.String(),
			ParokiID: m.ParokiID.String(),
			UserID:   m.UserID.String(),
			Email:    m.Email,
			Name:     m.Name,
			Phone:    m.Phone,
			Role:     m.Role,
		})
	}
	return out, nil
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
