package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/mailx"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/handler"
	"github.com/paroki/domus/internal/user/repo"
	"github.com/paroki/domus/internal/user/service"
)

// Re-export transport-agnostic principal types and helpers for downstream features.
type Principal = authx.Principal

var (
	FromContext   = authx.FromContext
	UserContext   = authx.UserContext
	WithPrincipal = authx.WithPrincipal
	RequireTenant = authx.RequireTenant
)

// Feature-owned Casbin roles and resource permissions.
var (
	RoleSuperAdmin = authz.Role("user", "super-admin")
	RoleUser       = authz.Role("user", "user")
	PermUsers      = authz.Permission("user", "users")
)

type options struct {
	otpSender      domain.OTPSender
	oauthProviders []domain.OAuthProvider
}

// Option customizes user Feature initialization.
type Option func(*options)

// WithOTPSender overrides the default slog OTPSender.
func WithOTPSender(sender domain.OTPSender) Option {
	return func(o *options) {
		o.otpSender = sender
	}
}

// NewMailOTPSender constructs an email OTPSender backed by mailer.
func NewMailOTPSender(mailer *mailx.Mailer, otpTTL time.Duration) domain.OTPSender {
	return service.NewMailOTPSender(mailer, otpTTL)
}

// WithOAuthProvider registers a custom or test OAuth2 provider driver.
func WithOAuthProvider(p domain.OAuthProvider) Option {
	return func(o *options) {
		if p != nil {
			o.oauthProviders = append(o.oauthProviders, p)
		}
	}
}

// Feature wires the user and authentication feature.
type Feature struct {
	service    *service.Service
	handler    *handler.Handler
	enforcer   *authz.Enforcer
	cookieName string
}

// New constructs the user Feature with Ent persistence, Casbin enforcer, config, and options.
func New(db *ent.Client, enforcer *authz.Enforcer, cfg Config, logger *slog.Logger, opts ...Option) *Feature {
	var opt options
	for _, o := range opts {
		if o != nil {
			o(&opt)
		}
	}

	cookieName := cfg.CookieName
	if cookieName == "" {
		cookieName = "domus_session"
	}

	r := repo.New(db)
	svcCfg := service.Config{
		SuperAdminEmail: cfg.SuperAdminEmail,
		SuperAdminName:  cfg.SuperAdminName,
		OTPTTL:          cfg.OTPTTL,
		OTPMaxAttempts:  cfg.OTPMaxAttempts,
		OTPCooldown:     cfg.OTPCooldown,
		SessionTTL:      cfg.SessionTTL,
		GoogleID:        cfg.GoogleID,
		GoogleSecret:    cfg.GoogleSecret,
		GitHubID:        cfg.GitHubID,
		GitHubSecret:    cfg.GitHubSecret,
	}
	svc := service.New(svcCfg, r, enforcer, opt.otpSender, logger)
	for _, p := range opt.oauthProviders {
		svc.RegisterOAuthProvider(p)
	}
	h := handler.New(svc, cookieName, cfg.CookieSecure, cfg.SessionTTL)
	return &Feature{
		service:    svc,
		handler:    h,
		enforcer:   enforcer,
		cookieName: cookieName,
	}
}

// Register mounts user and authentication routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.AuthenticateMiddleware(),
		f.RequireAuth(),
		f.Require(PermUsers, authz.ActionRead),
		f.Require(PermUsers, authz.ActionWrite),
	)
}

// RegisterPermissions registers user feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	if err := e.AddPermissionForRole(ctx, RoleSuperAdmin, authz.DomainGlobal, PermUsers, authz.ActionRead); err != nil {
		return err
	}
	return e.AddPermissionForRole(ctx, RoleSuperAdmin, authz.DomainGlobal, PermUsers, authz.ActionWrite)
}

// Bootstrap idempotently seeds the initial Super Admin user when configured.
func (f *Feature) Bootstrap(ctx context.Context) error {
	return f.service.BootstrapSuperAdmin(ctx)
}

// AuthenticateMiddleware resolves a session cookie or Bearer token and attaches authx.Principal to the request context.
func (f *Feature) AuthenticateMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := f.extractSessionToken(c)
		if token == "" {
			return c.Next()
		}

		var headerParishID uuid.UUID
		if rawParish := strings.TrimSpace(c.Get("X-Parish-Id")); rawParish != "" {
			parsed, err := uuid.Parse(rawParish)
			if err != nil {
				return httpx.BadRequest("Invalid X-Parish-Id header.")
			}
			headerParishID = parsed
		}

		p, err := f.service.ResolveSession(c.Context(), token, headerParishID)
		if err != nil {
			if errors.Is(err, domain.ErrUserDisabled) {
				return err
			}
			if errors.Is(err, domain.ErrSessionInvalid) {
				return c.Next()
			}
			return err
		}

		c.SetContext(authx.WithPrincipal(c.Context(), p))
		return c.Next()
	}
}

// RequireAuth rejects unauthenticated requests with 401 UNAUTHORIZED.
func (f *Feature) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		p, ok := authx.FromContext(c.Context())
		if !ok || p.UserID == uuid.Nil {
			return httpx.Unauthorized("Authentication is required.")
		}
		return c.Next()
	}
}

// Require enforces Casbin permission checks for the given resource and action.
func (f *Feature) Require(obj string, act authz.Action, domFn ...func(fiber.Ctx) (string, error)) fiber.Handler {
	if f.enforcer == nil {
		return func(_ fiber.Ctx) error {
			return httpx.InternalError(errors.New("authz enforcer not configured"))
		}
	}
	return f.enforcer.Require(obj, act, domFn...)
}

// PrincipalResolver returns an mcpx.PrincipalResolver reading the authenticated Principal from Fiber context
// or resolving it from a session cookie or Bearer token when called outside AuthenticateMiddleware (e.g. /mcp).
func (f *Feature) PrincipalResolver() mcpx.PrincipalResolver {
	return func(c fiber.Ctx) (authx.Principal, bool) {
		if p, ok := authx.FromContext(c.Context()); ok && p.UserID != uuid.Nil {
			return p, true
		}
		token := f.extractSessionToken(c)
		if token == "" {
			return authx.Principal{}, false
		}
		var headerParishID uuid.UUID
		if rawParish := strings.TrimSpace(c.Get("X-Parish-Id")); rawParish != "" {
			parsed, err := uuid.Parse(rawParish)
			if err != nil {
				return authx.Principal{}, false
			}
			headerParishID = parsed
		}
		p, err := f.service.ResolveSession(c.Context(), token, headerParishID)
		if err != nil {
			return authx.Principal{}, false
		}
		c.SetContext(authx.WithPrincipal(c.Context(), p))
		return p, true
	}
}

// RegisterOAuthProvider registers a custom or test OAuth2 provider driver.
func (f *Feature) RegisterOAuthProvider(p domain.OAuthProvider) {
	f.service.RegisterOAuthProvider(p)
}

// ResolveByChannel resolves a linked messaging/bot channel identity into an authenticated Principal.
func (f *Feature) ResolveByChannel(ctx context.Context, provider, subject string, overrideParishID uuid.UUID) (Principal, error) {
	return f.service.ResolveByChannel(ctx, provider, subject, overrideParishID)
}

// SwitchChannelParish updates the active parish for a linked channel identity.
func (f *Feature) SwitchChannelParish(ctx context.Context, provider, subject string, parishID uuid.UUID) error {
	return f.service.SwitchChannelParish(ctx, provider, subject, parishID)
}

// EnsureUser finds or creates a user for invite-based onboarding and links the provider account.
func (f *Feature) EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (*domain.UserResponse, error) {
	return f.service.EnsureUser(ctx, email, name, phone, provider, subject)
}

// InviteJoiner validates and redeems organization invite codes during guest auth onboarding.
type InviteJoiner = service.InviteJoiner

// SetInviteJoiner wires the organization invite joiner into the user auth service.
func (f *Feature) SetInviteJoiner(j InviteJoiner) {
	f.service.SetInviteJoiner(j)
}

// SwitchSessionParish updates the active parish on sessionID for userID.
func (f *Feature) SwitchSessionParish(ctx context.Context, userID, sessionID, parishID uuid.UUID) error {
	return f.service.SwitchSessionParish(ctx, authx.Principal{
		UserID:    userID,
		SessionID: sessionID,
	}, parishID)
}

func (f *Feature) extractSessionToken(c fiber.Ctx) string {
	if cookie := c.Cookies(f.cookieName); cookie != "" {
		return cookie
	}
	authHeader := c.Get(fiber.HeaderAuthorization)
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
