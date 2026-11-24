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

// Feature wires LLM provider configuration and model discovery for the agent feature.
type Feature struct {
	drivers   *driver.Registry
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
		f.handler = handler.New(f)
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
}

// MCPServer returns the bound MCP server, or nil if unset.
func (f *Feature) MCPServer() *mcp.Server {
	return f.mcpServer
}

// Drivers returns the multi-provider model discovery Registry.
func (f *Feature) Drivers() *driver.Registry {
	return f.drivers
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

// Register mounts /agent/providers routes under /api.
func (f *Feature) Register(api fiber.Router) {
	if f.handler == nil || f.users == nil {
		return
	}
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		f.users.Require(PermProviders, authz.ActionRead),
		f.users.Require(PermProviders, authz.ActionWrite),
	)
}
