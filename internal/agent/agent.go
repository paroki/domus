package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/handler"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
)

// Feature-owned Casbin resource permissions.
var (
	PermSessions = authz.Permission("agent", "sessions")
	PermMemories = authz.Permission("agent", "memories")
)

// Feature wires AI agent configuration, drivers, settings, and orchestration service.
type Feature struct {
	config   *ConfigStore
	drivers  *driver.Registry
	service  *service.Service
	handler  *handler.Handler
	users    *user.Feature
	settings *setting.Store
}

type configAdapter struct {
	store *ConfigStore
}

func (a configAdapter) AgentConfig(name string) (domain.AgentConfig, bool) {
	return a.store.AgentConfig(name)
}

func (a configAdapter) EmbeddingConfig() service.EmbeddingConfigSnapshot {
	e := a.store.EmbeddingConfig()
	return service.EmbeddingConfigSnapshot{
		Provider: e.Provider,
		BaseURL:  e.BaseURL,
		APIKey:   e.APIKey,
		Model:    e.Model,
	}
}

func (a configAdapter) ProviderModels(provider string) []domain.ProviderModelItem {
	return a.store.ProviderModels(provider)
}

type userReaderAdapter struct {
	users *user.Feature
}

func (a userReaderAdapter) GetUser(ctx context.Context, userID uuid.UUID) (*service.UserProfileInfo, error) {
	u, err := a.users.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &service.UserProfileInfo{
		ID:        u.ID,
		Name:      u.Name,
		Gender:    u.Gender,
		BirthDate: u.BirthDate,
	}, nil
}

type parishReaderAdapter struct {
	refs *ref.Feature
}

func (a parishReaderAdapter) GetParish(ctx context.Context, parishID uuid.UUID) (*service.ParishInfo, error) {
	p, err := a.refs.GetParish(ctx, parishID)
	if err != nil {
		return nil, err
	}
	return &service.ParishInfo{
		ID:   p.ID,
		Code: p.Code,
		Name: p.Name,
	}, nil
}

func (a parishReaderAdapter) GetParishByCode(ctx context.Context, code string) (*service.ParishInfo, error) {
	p, err := a.refs.GetParishByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return &service.ParishInfo{
		ID:   p.ID,
		Code: p.Code,
		Name: p.Name,
	}, nil
}

type channelResolverAdapter struct {
	users *user.Feature
}

func (a channelResolverAdapter) ResolveByChannel(ctx context.Context, provider, subject string, overrideParishID uuid.UUID) (authx.Principal, error) {
	return a.users.ResolveByChannel(ctx, provider, subject, overrideParishID)
}

func (a channelResolverAdapter) SwitchChannelParish(ctx context.Context, provider, subject string, parishID uuid.UUID) error {
	return a.users.SwitchChannelParish(ctx, provider, subject, parishID)
}

func (a channelResolverAdapter) TelegramClient() *telegram.Client {
	return a.users.TelegramClient()
}

func (a channelResolverAdapter) GowaClient() *gowa.Client {
	return a.users.GowaClient()
}

type orgReaderAdapter struct {
	orgs *org.Feature
}

func (a orgReaderAdapter) ListUserOrgs(ctx context.Context, parishID, userID uuid.UUID) ([]service.OrgInfo, error) {
	rows, err := a.orgs.ListUserOrgs(ctx, parishID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]service.OrgInfo, len(rows))
	for i, o := range rows {
		out[i] = service.OrgInfo{
			ID:   o.ID,
			Code: o.Code,
			Name: o.Name,
			Kind: string(o.Kind),
			Role: "member",
		}
	}
	return out, nil
}

// New constructs the agent Feature with its ConfigStore, driver Registry, and orchestration Service.
func New(
	drivers *driver.Registry,
	logger *slog.Logger,
	opts ...Option,
) *Feature {
	if drivers == nil {
		drivers = driver.NewRegistry(nil)
	}
	cfg := NewConfigStore(drivers, logger)
	f := &Feature{
		config:  cfg,
		drivers: drivers,
	}
	var cfgOpts featureOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&cfgOpts)
		}
	}
	f.users = cfgOpts.users
	f.settings = cfgOpts.settings
	if cfgOpts.db != nil {
		var ur service.UserReader
		if cfgOpts.users != nil {
			ur = userReaderAdapter{users: cfgOpts.users}
		}
		var pr service.ParishReader
		if cfgOpts.refs != nil {
			pr = parishReaderAdapter{refs: cfgOpts.refs}
		}
		var or service.OrgReader
		if cfgOpts.orgs != nil {
			or = orgReaderAdapter{orgs: cfgOpts.orgs}
		}
		r := repo.New(cfgOpts.db, cfgOpts.sqlDB, cfgOpts.dialect)
		f.service = service.New(r, configAdapter{store: cfg}, drivers, cfgOpts.enforcer, cfgOpts.mcpServer, ur, pr, or, logger)
		if cfgOpts.users != nil {
			f.service.SetChannelResolver(channelResolverAdapter{users: cfgOpts.users})
			cfgOpts.users.SetTelegramMessageHandler(f.HandleTelegramUpdate)
		}
		f.handler = handler.New(f.service, f)
	}
	return f
}

type featureOptions struct {
	db        *ent.Client
	sqlDB     *sql.DB
	dialect   string
	enforcer  *authz.Enforcer
	mcpServer *mcp.Server
	users     *user.Feature
	refs      *ref.Feature
	orgs      *org.Feature
	settings  *setting.Store
}

// Option configures optional dependencies on agent.New.
type Option func(*featureOptions)

// WithService wires the persistence, authz, MCP server, and cross-feature readers into Feature.Service().
func WithService(
	db *ent.Client,
	sqlDB *sql.DB,
	dialectName string,
	enforcer *authz.Enforcer,
	mcpServer *mcp.Server,
	users *user.Feature,
	refs *ref.Feature,
	orgs *org.Feature,
) Option {
	return func(o *featureOptions) {
		o.db = db
		o.sqlDB = sqlDB
		o.dialect = dialectName
		o.enforcer = enforcer
		o.mcpServer = mcpServer
		o.users = users
		o.refs = refs
		o.orgs = orgs
	}
}

// WithSettings wires the setting.Store used for provider model discovery and persistence.
func WithSettings(store *setting.Store) Option {
	return func(o *featureOptions) {
		o.settings = store
	}
}

// Config returns the hot-reloaded ConfigStore.
func (f *Feature) Config() *ConfigStore {
	return f.config
}

// Drivers returns the multi-provider LLM and embedding Registry.
func (f *Feature) Drivers() *driver.Registry {
	return f.drivers
}

// Service returns the agent orchestration Service (nil if WithService was not provided).
func (f *Feature) Service() *service.Service {
	return f.service
}

// SetMCPServer binds the in-process MCP server used by the agent orchestrator.
func (f *Feature) SetMCPServer(s *mcp.Server) {
	if f.service != nil {
		f.service.SetMCPServer(s)
	}
}

// HandleTelegramUpdate processes non-/start Telegram bot updates via the AI agent.
func (f *Feature) HandleTelegramUpdate(ctx context.Context, update telegram.Update) error {
	if f.service == nil {
		return nil
	}
	return f.service.HandleTelegramUpdate(ctx, update)
}

// RegisterSettings registers all AI agent settings descriptors and hot-reload listeners on s.
func (f *Feature) RegisterSettings(s *setting.Store) {
	f.settings = s
	f.config.RegisterSettings(s)
}

// Register mounts /agent routes under /api.
func (f *Feature) Register(api fiber.Router) {
	if f.handler == nil || f.users == nil {
		return
	}
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(user.PermUsers, authz.ActionWrite),
		f.users.Require(PermSessions, authz.ActionRead),
		f.users.Require(PermSessions, authz.ActionWrite),
		f.users.Require(PermMemories, authz.ActionRead),
		f.users.Require(PermMemories, authz.ActionWrite),
	)
}

// RegisterPermissions registers agent feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	policies := []struct {
		perm string
		act  authz.Action
	}{
		{PermSessions, authz.ActionRead},
		{PermSessions, authz.ActionWrite},
		{PermMemories, authz.ActionRead},
		{PermMemories, authz.ActionWrite},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, user.RoleUser, authz.DomainGlobal, p.perm, p.act); err != nil {
			return err
		}
	}
	return nil
}

// SyncProviderModels discovers available models from the upstream provider and persists the catalog.
func (f *Feature) SyncProviderModels(ctx context.Context, provider string) (domain.SyncProviderModelsResponse, error) {
	var pDef *providerDef
	for i := range providerDefs {
		if providerDefs[i].name == provider {
			pDef = &providerDefs[i]
			break
		}
	}
	if pDef == nil {
		return domain.SyncProviderModelsResponse{}, httpx.BadRequest("Unsupported AI provider.")
	}
	if f.settings == nil {
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderUnavailable
	}
	apiKey := strings.TrimSpace(f.settings.Get(pDef.apiKeyKey))
	if apiKey == "" {
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderUnavailable
	}
	baseURL := strings.TrimSpace(f.settings.Get(pDef.baseURLKey))
	discovered, err := f.drivers.DiscoverModels(ctx, pDef.name, baseURL, apiKey)
	if err != nil {
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderUnavailable
	}
	b, err := json.Marshal(discovered)
	if err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}
	if err := f.settings.PersistValue(ctx, pDef.modelsKey, string(b)); err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}
	f.config.SetProviderModels(pDef.name, discovered)

	dtos := make([]domain.ProviderModelDTO, len(discovered))
	for i, m := range discovered {
		dtos[i] = domain.ProviderModelDTO(m)
	}
	return domain.SyncProviderModelsResponse{
		Provider: pDef.name,
		Models:   dtos,
	}, nil
}

func requireTenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := user.RequireTenant(c.Context()); err != nil {
			return err
		}
		return c.Next()
	}
}
