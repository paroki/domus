package event

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/event/handler"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/event/service"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
)

// Re-export read-only DTO types for downstream features.
type (
	Event        = domain.EventResponse
	RSVP         = domain.EventRSVPResponse
	RSVPSummary  = domain.EventRSVPSummaryResponse
	EventPetugas = domain.EventPetugasResponse
	EventLagu    = domain.EventLaguResponse
)

// Feature-owned Casbin resource permissions.
var (
	PermEvents = authz.Permission("event", "events")
	PermRSVPs  = authz.Permission("event", "rsvps")
)

type orgReaderAdapter struct {
	orgs *org.Feature
}

func (a orgReaderAdapter) GetOrganization(ctx context.Context, parokiID, orgID uuid.UUID) (*service.OrgInfo, error) {
	o, err := a.orgs.GetOrganization(ctx, parokiID, orgID)
	if err != nil {
		return nil, err
	}
	return &service.OrgInfo{
		ID:   o.ID,
		Code: o.Code,
		Name: o.Name,
		Kind: o.Kind,
	}, nil
}

func (a orgReaderAdapter) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return a.orgs.IsMember(ctx, orgID, userID)
}

func (a orgReaderAdapter) IsAdmin(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return a.orgs.IsAdmin(ctx, orgID, userID)
}

func (a orgReaderAdapter) ListUserOrgs(ctx context.Context, parokiID, userID uuid.UUID) ([]service.OrgInfo, error) {
	rows, err := a.orgs.ListUserOrgs(ctx, parokiID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]service.OrgInfo, 0, len(rows))
	for _, o := range rows {
		out = append(out, service.OrgInfo{
			ID:   o.ID,
			Code: o.Code,
			Name: o.Name,
			Kind: o.Kind,
		})
	}
	return out, nil
}

type parokiReaderAdapter struct {
	refs *ref.Feature
}

func (a parokiReaderAdapter) GetParoki(ctx context.Context, parokiID uuid.UUID) (*service.ParokiInfo, error) {
	p, err := a.refs.GetParoki(ctx, parokiID)
	if err != nil {
		return nil, err
	}
	return &service.ParokiInfo{
		ID:       p.ID,
		Kode:     p.Kode,
		Nama:     p.Nama,
		Timezone: p.Timezone,
	}, nil
}

func (a parokiReaderAdapter) HasActiveLookup(ctx context.Context, parokiID uuid.UUID, category, code string) (bool, error) {
	return a.refs.HasActiveLookup(ctx, parokiID, category, code)
}

// Option configures optional dependencies on the event Feature.
type Option func(*options)

type options struct {
	refs *ref.Feature
}

// WithRef wires the ref.Feature for reading parish metadata such as timezone.
func WithRef(refs *ref.Feature) Option {
	return func(o *options) {
		o.refs = refs
	}
}

// Feature wires event management and visibility/RBAC enforcement.
type Feature struct {
	service *service.Service
	handler *handler.Handler
	users   *user.Feature
	cfg     Config
}

// New constructs the event Feature.
func New(db *ent.Client, enforcer *authz.Enforcer, users *user.Feature, orgs *org.Feature, cfg Config, opts ...Option) *Feature {
	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	r := repo.New(db)
	var orgReader service.OrgReader
	if orgs != nil {
		orgReader = orgReaderAdapter{orgs: orgs}
	}
	var parokiReader service.ParokiReader
	if o.refs != nil {
		parokiReader = parokiReaderAdapter(o)
	}
	svc := service.New(r, enforcer, orgReader, parokiReader, service.Config{
		DefaultVisibility: cfg.DefaultVisibility,
		RSVPGracePeriod:   cfg.RSVPGracePeriod,
	})
	h := handler.New(svc)
	return &Feature{
		service: svc,
		handler: h,
		users:   users,
		cfg:     cfg,
	}
}

// Register mounts /events routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(PermEvents, authz.ActionRead),
	)
}

// RegisterPermissions registers event feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	policies := []struct {
		role string
		perm string
		act  authz.Action
	}{
		{user.RoleUser, PermEvents, authz.ActionRead},
		{user.RoleUser, PermRSVPs, authz.ActionRead},
		{user.RoleUser, PermRSVPs, authz.ActionWrite},
		{ref.RoleParokiAdmin, PermEvents, authz.ActionRead},
		{ref.RoleParokiAdmin, PermEvents, authz.ActionWrite},
		{ref.RoleParokiAdmin, PermRSVPs, authz.ActionRead},
		{ref.RoleParokiAdmin, PermRSVPs, authz.ActionWrite},
		{org.RoleAdmin, PermEvents, authz.ActionRead},
		{org.RoleAdmin, PermEvents, authz.ActionWrite},
		{org.RoleAdmin, PermRSVPs, authz.ActionRead},
		{org.RoleAdmin, PermRSVPs, authz.ActionWrite},
		{org.RoleMember, PermEvents, authz.ActionRead},
		{org.RoleMember, PermRSVPs, authz.ActionRead},
		{org.RoleMember, PermRSVPs, authz.ActionWrite},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, p.role, authz.DomainGlobal, p.perm, p.act); err != nil {
			return err
		}
	}
	return nil
}

func requireTenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := user.RequireTenant(c.Context()); err != nil {
			return err
		}
		return c.Next()
	}
}
