package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/user/service"
)

// Handler exposes HTTP endpoints for authentication and user operations.
type Handler struct {
	service      *service.Service
	cookieName   string
	cookieSecure bool
	sessionTTL   time.Duration
}

// New constructs a Handler backed by svc and session cookie configuration.
func New(svc *service.Service, cookieName string, cookieSecure bool, sessionTTL time.Duration) *Handler {
	if cookieName == "" {
		cookieName = "domus_session"
	}
	return &Handler{
		service:      svc,
		cookieName:   cookieName,
		cookieSecure: cookieSecure,
		sessionTTL:   sessionTTL,
	}
}

// Register mounts /auth routes onto api (/api).
func (h *Handler) Register(api fiber.Router) {
	auth := api.Group("/auth")
	// Static routes MUST be registered before parameterized /:provider routes.
	auth.Get("/providers", h.GetProviders)
	auth.Post("/otp", h.RequestOTP)
	auth.Post("/otp/verify", h.VerifyOTP)
	auth.Post("/logout", h.Logout)
	auth.Get("/:provider", h.OAuthRedirect)
	auth.Get("/:provider/callback", h.OAuthCallback)
}
