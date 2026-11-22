package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/mailx"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/platform/telegram"
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

// Feature-owned runtime setting keys.
var (
	SettingTelegramBotToken  = setting.Key("user", "telegram.bot_token")
	SettingGowaBaseURL       = setting.Key("user", "gowa.base_url")
	SettingGowaWebhookSecret = setting.Key("user", "gowa.webhook_secret")
)

type options struct {
	otpSender         domain.OTPSender
	channelOTPSenders map[string]domain.OTPSender
	gowaClient        *gowa.Client
	telegramClient    *telegram.Client
	oauthProviders    []domain.OAuthProvider
}

// Option customizes user Feature initialization.
type Option func(*options)

// WithOTPSender overrides the default email slog OTPSender.
func WithOTPSender(sender domain.OTPSender) Option {
	return func(o *options) {
		o.otpSender = sender
	}
}

// WithGowaClient configures WhatsApp OTP delivery via GOWA.
func WithGowaClient(client *gowa.Client) Option {
	return func(o *options) {
		o.gowaClient = client
	}
}

// WithTelegramClient configures Telegram OTP delivery and bot webhook linking.
func WithTelegramClient(client *telegram.Client) Option {
	return func(o *options) {
		o.telegramClient = client
	}
}

// WithChannelOTPSender registers a custom or test OTPSender for a specific channel.
func WithChannelOTPSender(channel string, sender domain.OTPSender) Option {
	return func(o *options) {
		ch := strings.ToLower(strings.TrimSpace(channel))
		if ch == "" || sender == nil {
			return
		}
		if o.channelOTPSenders == nil {
			o.channelOTPSenders = make(map[string]domain.OTPSender)
		}
		o.channelOTPSenders[ch] = sender
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
	service         *service.Service
	handler         *handler.Handler
	enforcer        *authz.Enforcer
	compositeSender *service.CompositeOTPSender
	otpTTL          time.Duration
	cookieName      string
	clientMu        sync.RWMutex
	tgClient        *telegram.Client
	gowaClient      *gowa.Client
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
	otpTTL := cfg.OTPTTL
	if otpTTL <= 0 {
		otpTTL = 5 * time.Minute
	}
	otpMaxAttempts := cfg.OTPMaxAttempts
	if otpMaxAttempts <= 0 {
		otpMaxAttempts = 5
	}
	sessionTTL := cfg.SessionTTL
	if sessionTTL <= 0 {
		sessionTTL = 720 * time.Hour
	}

	r := repo.New(db)
	compositeSender := service.NewCompositeOTPSender(logger)
	if opt.otpSender != nil {
		compositeSender.SetSender(domain.ProviderEmail, opt.otpSender)
	}
	if opt.gowaClient != nil {
		compositeSender.SetSender(domain.ProviderWhatsApp, service.NewGowaOTPSender(opt.gowaClient, otpTTL))
	}
	if opt.telegramClient != nil {
		compositeSender.SetSender(domain.ProviderTelegram, service.NewTelegramOTPSender(opt.telegramClient, otpTTL))
	}
	for ch, sender := range opt.channelOTPSenders {
		compositeSender.SetSender(ch, sender)
	}

	svcCfg := service.Config{
		SuperAdminEmail: cfg.SuperAdminEmail,
		SuperAdminName:  cfg.SuperAdminName,
		OTPTTL:          otpTTL,
		OTPMaxAttempts:  otpMaxAttempts,
		OTPCooldown:     cfg.OTPCooldown,
		SessionTTL:      sessionTTL,
		GoogleID:        cfg.GoogleID,
		GoogleSecret:    cfg.GoogleSecret,
		GitHubID:        cfg.GitHubID,
		GitHubSecret:    cfg.GitHubSecret,
	}
	svc := service.New(svcCfg, r, enforcer, compositeSender, logger)
	if opt.telegramClient != nil {
		svc.SetTelegramClient(opt.telegramClient)
	}
	for _, p := range opt.oauthProviders {
		svc.RegisterOAuthProvider(p)
	}
	h := handler.New(svc, cookieName, sessionTTL)
	if opt.telegramClient != nil {
		h.SetTelegramClient(opt.telegramClient)
	}
	return &Feature{
		service:         svc,
		handler:         h,
		enforcer:        enforcer,
		compositeSender: compositeSender,
		otpTTL:          otpTTL,
		cookieName:      cookieName,
		tgClient:        opt.telegramClient,
		gowaClient:      opt.gowaClient,
	}
}

// RegisterSettings registers Telegram and GOWA descriptors and a hot-reload listener on s.
func (f *Feature) RegisterSettings(s *setting.Store) {
	s.Register(setting.Descriptor{
		Key:         SettingTelegramBotToken,
		Group:       "Telegram Bot",
		Label:       "Bot Token",
		Description: "Token bot dari @BotFather untuk mengirim OTP dan menangani webhook.",
		Secret:      true,
		Placeholder: "123456:ABC-DEF...",
	})
	s.Register(setting.Descriptor{
		Key:         SettingGowaBaseURL,
		Group:       "GOWA (WhatsApp)",
		Label:       "Base URL",
		Description: "URL dasar layanan GOWA untuk pengiriman OTP WhatsApp.",
		Secret:      false,
		Placeholder: "https://gowa.example.com",
	})
	s.Register(setting.Descriptor{
		Key:         SettingGowaWebhookSecret,
		Group:       "GOWA (WhatsApp)",
		Label:       "Webhook Secret",
		Description: "Secret HMAC-SHA256 untuk memverifikasi header X-Hub-Signature-256 dari webhook GOWA.",
		Secret:      true,
		Placeholder: "webhook-secret",
	})

	s.OnChange(func(_ context.Context, values map[string]string) error {
		botToken := strings.TrimSpace(values[SettingTelegramBotToken])

		var tgClient *telegram.Client
		if botToken != "" {
			c, err := telegram.New(telegram.Config{
				BotToken: botToken,
			})
			if err != nil {
				return err
			}
			tgClient = c
		}

		baseURL := strings.TrimSpace(values[SettingGowaBaseURL])
		gowaWebhookSecret := strings.TrimSpace(values[SettingGowaWebhookSecret])

		var gowaClient *gowa.Client
		if baseURL != "" || gowaWebhookSecret != "" {
			if baseURL == "" {
				return errors.New("GOWA Base URL wajib diisi jika pengaturan GOWA digunakan")
			}
			// ponytail: BasicUser, BasicPass, DeviceID, and BotNumber omitted from UI settings until multi-device or Basic-Auth GOWA is needed.
			c, err := gowa.New(gowa.Config{
				BaseURL:       baseURL,
				WebhookSecret: gowaWebhookSecret,
			})
			if err != nil {
				return err
			}
			gowaClient = c
		}

		f.clientMu.Lock()
		f.tgClient = tgClient
		f.gowaClient = gowaClient
		f.clientMu.Unlock()

		if tgClient != nil {
			f.compositeSender.SetSender(domain.ProviderTelegram, service.NewTelegramOTPSender(tgClient, f.otpTTL))
			f.service.SetTelegramClient(tgClient)
			f.handler.SetTelegramClient(tgClient)
		} else {
			f.compositeSender.RemoveSender(domain.ProviderTelegram)
			f.service.SetTelegramClient(nil)
			f.handler.SetTelegramClient(nil)
		}

		if gowaClient != nil {
			f.compositeSender.SetSender(domain.ProviderWhatsApp, service.NewGowaOTPSender(gowaClient, f.otpTTL))
		} else {
			f.compositeSender.RemoveSender(domain.ProviderWhatsApp)
		}
		return nil
	})
}

// TelegramClient returns the active hot-reloaded Telegram Bot client (or nil if unconfigured).
func (f *Feature) TelegramClient() *telegram.Client {
	f.clientMu.RLock()
	defer f.clientMu.RUnlock()
	return f.tgClient
}

// GowaClient returns the active hot-reloaded GOWA WhatsApp client (or nil if unconfigured).
func (f *Feature) GowaClient() *gowa.Client {
	f.clientMu.RLock()
	defer f.clientMu.RUnlock()
	return f.gowaClient
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
	e.RegisterRole(authz.RoleDescriptor{
		Code:        RoleSuperAdmin,
		Name:        "Super Admin",
		Scope:       authz.RoleScopeGlobal,
		Description: "Mengelola pengguna global, keuskupan, tenant paroki, dan kamus referensi global.",
	})
	e.RegisterRole(authz.RoleDescriptor{
		Code:        RoleUser,
		Name:        "Pengguna",
		Scope:       authz.RoleScopeGlobal,
		Description: "Pengguna terdaftar aplikasi yang dapat bergabung ke paroki dan organisasi.",
	})
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

// CompleteTelegramLink consumes a Telegram deep-link token and links chatID to the owning user.
func (f *Feature) CompleteTelegramLink(ctx context.Context, token, chatID, displayName string) error {
	return f.service.CompleteTelegramLink(ctx, token, chatID, displayName)
}

// SetTelegramMessageHandler sets a fallback handler for non-/start Telegram webhook messages.
func (f *Feature) SetTelegramMessageHandler(fn func(context.Context, telegram.Update) error) {
	f.handler.SetTelegramMessageHandler(fn)
}

// GetUser returns an active user's profile by ID.
func (f *Feature) GetUser(ctx context.Context, userID uuid.UUID) (*domain.UserResponse, error) {
	return f.service.GetUser(ctx, userID)
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
