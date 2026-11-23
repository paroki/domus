package agent

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/handler"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/user"
)

// Feature wires LLM provider configuration and model discovery for the agent feature.
type Feature struct {
	drivers *driver.Registry
	handler *handler.Handler
	db      *ent.Client
	cipher  *crypt.Cipher
	users   *user.Feature
	logger  *slog.Logger
}

type featureOptions struct {
	db     *ent.Client
	users  *user.Feature
	cipher *crypt.Cipher
}

// Option configures optional dependencies on agent.New.
type Option func(*featureOptions)

// WithDB wires the Ent client used to persist LLM provider configurations.
func WithDB(db *ent.Client) Option {
	return func(o *featureOptions) {
		o.db = db
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
		drivers: drivers,
		db:      cfgOpts.db,
		cipher:  cfgOpts.cipher,
		users:   cfgOpts.users,
		logger:  logger,
	}
	if f.db != nil {
		f.handler = handler.New(f)
	}
	return f
}

// Drivers returns the multi-provider model discovery Registry.
func (f *Feature) Drivers() *driver.Registry {
	return f.drivers
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
		f.users.Require(user.PermUsers, authz.ActionRead),
		f.users.Require(user.PermUsers, authz.ActionWrite),
	)
}
