package user

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/handler"
	"github.com/paroki/domus/internal/user/repo"
	"github.com/paroki/domus/internal/user/service"
)

// Feature wires the user and authentication feature.
type Feature struct {
	service *service.Service
	handler *handler.Handler
}

// New constructs the user Feature with Ent persistence, config, and optional OTPSender.
func New(cfg Config, client *ent.Client, otpSender domain.OTPSender, logger *slog.Logger) *Feature {
	r := repo.New(client)
	svcCfg := service.Config{
		OTPTTL:         cfg.OTPTTL,
		OTPMaxAttempts: cfg.OTPMaxAttempts,
		OTPCooldown:    cfg.OTPCooldown,
		SessionTTL:     cfg.SessionTTL,
		GoogleID:       cfg.GoogleID,
		GoogleSecret:   cfg.GoogleSecret,
		GitHubID:       cfg.GitHubID,
		GitHubSecret:   cfg.GitHubSecret,
	}
	svc := service.New(svcCfg, r, otpSender, logger)
	h := handler.New(svc, cfg.CookieName, cfg.CookieSecure, cfg.SessionTTL)
	return &Feature{
		service: svc,
		handler: h,
	}
}

// Register mounts user and authentication routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(api)
}

// RegisterOAuthProvider registers a custom or test OAuth2 provider driver.
func (f *Feature) RegisterOAuthProvider(p domain.OAuthProvider) {
	f.service.RegisterOAuthProvider(p)
}

// EnsureUser finds or creates a user for invite-based onboarding and links the provider account.
func (f *Feature) EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (*domain.UserResponse, error) {
	return f.service.EnsureUser(ctx, email, name, phone, provider, subject)
}
