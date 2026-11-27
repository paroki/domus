package liturgi

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/cantus"
	"github.com/paroki/domus/internal/event"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/liturgi/handler"
	"github.com/paroki/domus/internal/liturgi/repo"
	"github.com/paroki/domus/internal/liturgi/service"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
)

// Feature-owned Casbin role and resource permissions.
var (
	RoleAdmin      = domain.RoleAdmin
	PermSchedules  = domain.PermSchedules
	PermMinistries = domain.PermMinistries
	PermPetugas    = domain.PermPetugas
	PermLagu       = domain.PermLagu
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
		Name: o.Name,
	}, nil
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
			Name: o.Name,
		})
	}
	return out, nil
}

type cantusReaderAdapter struct {
	cantus *cantus.Feature
}

func (a cantusReaderAdapter) GetCantus(ctx context.Context, id uuid.UUID) (*service.CantusInfo, error) {
	c, err := a.cantus.GetCantus(ctx, id)
	if err != nil {
		return nil, err
	}
	return &service.CantusInfo{
		ID:        c.ID,
		Sumber:    c.Sumber,
		Nomor:     c.Nomor,
		Judul:     c.Judul,
		NadaDasar: c.NadaDasar,
	}, nil
}

type refReaderAdapter struct {
	refs *ref.Feature
}

func (a refReaderAdapter) GetParokiTimezone(ctx context.Context, parokiID uuid.UUID) (string, error) {
	p, err := a.refs.GetParoki(ctx, parokiID)
	if err != nil {
		return "", err
	}
	return p.Timezone, nil
}

func (a refReaderAdapter) HasActiveLookup(ctx context.Context, parokiID uuid.UUID, category, code string) (bool, error) {
	return a.refs.HasActiveLookup(ctx, parokiID, category, code)
}

type userNotifierAdapter struct {
	users *user.Feature
}

func (a userNotifierAdapter) NotifyUser(ctx context.Context, userID uuid.UUID, parokiID *uuid.UUID, kind string, refID uuid.UUID, window string, payload service.NotificationPayload) (bool, error) {
	return a.users.NotifyUser(ctx, userID, parokiID, kind, refID, window, user.NotificationPayload{
		Title: payload.Title,
		Body:  payload.Body,
		Icon:  payload.Icon,
		Badge: payload.Badge,
		Tag:   payload.Tag,
		URL:   payload.URL,
	})
}

func (a userNotifierAdapter) NotifyUsers(ctx context.Context, userIDs []uuid.UUID, parokiID *uuid.UUID, kind string, refID uuid.UUID, window string, payload service.NotificationPayload) (int, error) {
	return a.users.NotifyUsers(ctx, userIDs, parokiID, kind, refID, window, user.NotificationPayload{
		Title: payload.Title,
		Body:  payload.Body,
		Icon:  payload.Icon,
		Badge: payload.Badge,
		Tag:   payload.Tag,
		URL:   payload.URL,
	})
}

func (a userNotifierAdapter) ResolveUserTimezones(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	return a.users.ResolveUserTimezones(ctx, userIDs)
}

func (a userNotifierAdapter) CleanupExpiredNotificationLogs(ctx context.Context, before time.Time) (int, error) {
	return a.users.CleanupExpiredNotificationLogs(ctx, before)
}

// Option configures optional dependencies on the liturgi Feature.
type Option func(*options)

type options struct {
	cantus *cantus.Feature
	refs   *ref.Feature
}

// WithCantus wires the cantus.Feature for resolving cantusId song metadata.
func WithCantus(c *cantus.Feature) Option {
	return func(o *options) {
		o.cantus = c
	}
}

// WithRef wires the ref.Feature for reading parish timezone and liturgi_song_part lookups.
func WithRef(r *ref.Feature) Option {
	return func(o *options) {
		o.refs = r
	}
}

// Feature wires liturgical ministry configuration and parish Mass schedule management.
type Feature struct {
	service *service.Service
	handler *handler.Handler
	users   *user.Feature
}

// New constructs the liturgi Feature.
func New(db *ent.Client, enforcer *authz.Enforcer, users *user.Feature, orgs *org.Feature, events *event.Feature, opts ...Option) *Feature {
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
	svc := service.New(r, enforcer, orgReader, events)
	if o.cantus != nil {
		svc.SetCantus(cantusReaderAdapter{cantus: o.cantus})
	}
	if o.refs != nil {
		svc.SetRef(refReaderAdapter{refs: o.refs})
	}
	if users != nil {
		svc.SetNotifier(userNotifierAdapter{users: users})
	}
	h := handler.New(svc)
	return &Feature{
		service: svc,
		handler: h,
		users:   users,
	}
}

// IsCantusInUse reports whether cantusID is referenced by any liturgi_lagu row of a non-deleted Event.
func (f *Feature) IsCantusInUse(ctx context.Context, cantusID uuid.UUID) (bool, error) {
	return f.service.IsCantusInUse(ctx, cantusID)
}

// RunReminderTick dispatches due liturgy officer reminders for the given reference time.
func (f *Feature) RunReminderTick(ctx context.Context, now time.Time) error {
	return f.service.RunReminderTick(ctx, now)
}

// StartReminderScheduler starts the background liturgy reminder scheduler until ctx is cancelled.
func (f *Feature) StartReminderScheduler(ctx context.Context, interval time.Duration, logger *slog.Logger) {
	f.service.StartReminderScheduler(ctx, interval, logger)
}

// Register mounts /liturgi routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(PermMinistries, authz.ActionRead),
		f.users.Require(PermSchedules, authz.ActionRead),
	)
}

// RegisterPermissions registers liturgi feature roles and policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	e.RegisterRole(authz.RoleDescriptor{
		Code:        RoleAdmin,
		Name:        "Admin Liturgi",
		Scope:       authz.RoleScopeParoki,
		Description: "Mengelola jadwal perayaan Misa paroki, pemetaan 6 bidang petugas liturgi, penugasan petugas, dan daftar lagu perayaan.",
	})
	policies := []struct {
		role string
		perm string
		act  authz.Action
	}{
		{user.RoleUser, PermSchedules, authz.ActionRead},
		{user.RoleUser, PermMinistries, authz.ActionRead},
		{user.RoleUser, PermPetugas, authz.ActionRead},
		{user.RoleUser, PermLagu, authz.ActionRead},
		{user.RoleUser, PermPetugas, authz.ActionWrite},
		{user.RoleUser, PermLagu, authz.ActionWrite},
		{ref.RoleParokiAdmin, PermSchedules, authz.ActionRead},
		{ref.RoleParokiAdmin, PermSchedules, authz.ActionWrite},
		{ref.RoleParokiAdmin, PermMinistries, authz.ActionRead},
		{ref.RoleParokiAdmin, PermMinistries, authz.ActionWrite},
		{ref.RoleParokiAdmin, PermPetugas, authz.ActionRead},
		{ref.RoleParokiAdmin, PermPetugas, authz.ActionWrite},
		{ref.RoleParokiAdmin, PermLagu, authz.ActionRead},
		{ref.RoleParokiAdmin, PermLagu, authz.ActionWrite},
		{RoleAdmin, PermSchedules, authz.ActionRead},
		{RoleAdmin, PermSchedules, authz.ActionWrite},
		{RoleAdmin, PermMinistries, authz.ActionRead},
		{RoleAdmin, PermMinistries, authz.ActionWrite},
		{RoleAdmin, PermPetugas, authz.ActionRead},
		{RoleAdmin, PermPetugas, authz.ActionWrite},
		{RoleAdmin, PermLagu, authz.ActionRead},
		{RoleAdmin, PermLagu, authz.ActionWrite},
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
