package agent

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/handler"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/user"
)

// Re-exported Casbin permissions for the agent feature.
var (
	PermProviders = domain.PermProviders
	PermChat      = domain.PermChat
)

// EnsureMemorySearchSchema idempotently adds the halfvec(768) embedding column and
// ParadeDB BM25 + HNSW indexes on agent_memories.
func EnsureMemorySearchSchema(ctx context.Context, sqlDB *sql.DB) error {
	return repo.EnsureMemorySearchSchema(ctx, sqlDB)
}

// Feature wires LLM provider configuration, model discovery, and session/memory persistence for the agent feature.
type Feature struct {
	drivers   *driver.Registry
	repo      *repo.Repository
	service   *service.Service
	handler   *handler.Handler
	db        *ent.Client
	sqlDB     *sql.DB
	mcpServer *mcp.Server
	cipher    *crypt.Cipher
	users     *user.Feature
	settings  *setting.Store
	logger    *slog.Logger
}

type featureOptions struct {
	db       *ent.Client
	sqlDB    *sql.DB
	users    *user.Feature
	cipher   *crypt.Cipher
	settings *setting.Store
}

// Option configures optional dependencies on agent.New.
type Option func(*featureOptions)

// WithDB wires the Ent client used to persist LLM provider configurations.
func WithDB(db *ent.Client) Option {
	return func(o *featureOptions) {
		o.db = db
	}
}

// WithSQLDB wires the raw *sql.DB pool for agent database operations.
func WithSQLDB(sqlDB *sql.DB) Option {
	return func(o *featureOptions) {
		o.sqlDB = sqlDB
	}
}

// WithUsers wires the user.Feature used for auth and RBAC middleware.
func WithUsers(users *user.Feature) Option {
	return func(o *featureOptions) {
		o.users = users
	}
}

// WithCipher wires the crypt.Cipher used to encrypt provider API keys at rest.
func WithCipher(c *crypt.Cipher) Option {
	return func(o *featureOptions) {
		o.cipher = c
	}
}

// WithSettings wires the setting.Store used for agent model routing and prompt settings.
func WithSettings(s *setting.Store) Option {
	return func(o *featureOptions) {
		o.settings = s
	}
}

// New constructs the agent Feature for LLM provider management.
func New(drivers *driver.Registry, logger *slog.Logger, opts ...Option) *Feature {
	if drivers == nil {
		drivers = driver.NewRegistry(nil)
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	var cfgOpts featureOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&cfgOpts)
		}
	}
	f := &Feature{
		drivers:  drivers,
		db:       cfgOpts.db,
		sqlDB:    cfgOpts.sqlDB,
		cipher:   cfgOpts.cipher,
		users:    cfgOpts.users,
		settings: cfgOpts.settings,
		logger:   logger,
	}
	if f.db != nil {
		f.repo = repo.New(f.db, f.sqlDB)
		f.service = service.New(f.repo, f.db, f.drivers, f.cipher, f.settings, f.logger)
		f.handler = handler.New(f, f.service, f.service, f.service)
	}
	return f
}

// SQLDB returns the wired *sql.DB pool, or nil if unconfigured.
func (f *Feature) SQLDB() *sql.DB {
	return f.sqlDB
}

// SetMCPServer binds the shared MCP server instance (invoked by app.GetFiber).
func (f *Feature) SetMCPServer(s *mcp.Server) {
	f.mcpServer = s
	if f.service != nil {
		f.service.SetMCPServer(s)
	}
}

// MCPServer returns the bound MCP server, or nil if unset.
func (f *Feature) MCPServer() *mcp.Server {
	return f.mcpServer
}

// Drivers returns the multi-provider model discovery Registry.
func (f *Feature) Drivers() *driver.Registry {
	return f.drivers
}

// ExtractIdentityDocument extracts structured identity fields from document images via AI Vision.
func (f *Feature) ExtractIdentityDocument(ctx context.Context, in user.IdentityExtractInput) (user.IdentityExtraction, error) {
	return f.service.ExtractIdentityDocument(ctx, in)
}

// SetChatModelFactory overrides chat model creation (intended for unit tests with fake models).
func (f *Feature) SetChatModelFactory(fn service.ChatModelFactory) {
	f.service.SetChatModelFactory(fn)
}

// RegisterPermissions registers agent feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	if err := e.AddPermissionForRole(ctx, user.RoleSuperAdmin, authz.DomainGlobal, PermProviders, authz.ActionRead); err != nil {
		return err
	}
	if err := e.AddPermissionForRole(ctx, user.RoleSuperAdmin, authz.DomainGlobal, PermProviders, authz.ActionWrite); err != nil {
		return err
	}
	if err := e.AddPermissionForRole(ctx, user.RoleUser, authz.DomainGlobal, PermChat, authz.ActionRead); err != nil {
		return err
	}
	return e.AddPermissionForRole(ctx, user.RoleUser, authz.DomainGlobal, PermChat, authz.ActionWrite)
}

// Register mounts /agent/* routes under /api.
func (f *Feature) Register(api fiber.Router) {
	if f.handler == nil || f.users == nil {
		return
	}
	requireTenant := func(c fiber.Ctx) error {
		if _, err := user.RequireTenant(c.Context()); err != nil {
			return err
		}
		return c.Next()
	}
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenant,
		f.users.Require(PermProviders, authz.ActionRead),
		f.users.Require(PermProviders, authz.ActionWrite),
		f.users.Require(PermChat, authz.ActionRead),
		f.users.Require(PermChat, authz.ActionWrite),
	)
}
