package agent

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
)

// Feature wires AI agent configuration, drivers, settings, and orchestration service.
type Feature struct {
	config  *ConfigStore
	drivers *driver.Registry
	service *service.Service
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

// RegisterSettings registers all AI agent settings descriptors and hot-reload listeners on s.
func (f *Feature) RegisterSettings(s *setting.Store) {
	f.config.RegisterSettings(s)
}
