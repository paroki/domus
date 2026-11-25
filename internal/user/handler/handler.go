package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user/service"
)

// Handler exposes HTTP endpoints for authentication and user operations.
type Handler struct {
	service          *service.Service
	cookieName       string
	sessionTTL       time.Duration
	tgMu             sync.RWMutex
	telegramClient   *telegram.Client
	telegramFallback func(context.Context, telegram.Update) error
}

// New constructs a Handler backed by svc and session cookie configuration.
func New(svc *service.Service, cookieName string, sessionTTL time.Duration) *Handler {
	if cookieName == "" {
		cookieName = "domus_session"
	}
	return &Handler{
		service:    svc,
		cookieName: cookieName,
		sessionTTL: sessionTTL,
	}
}

// SetTelegramClient sets the Telegram Bot client used for webhook verification and replies.
func (h *Handler) SetTelegramClient(c *telegram.Client) {
	h.tgMu.Lock()
	defer h.tgMu.Unlock()
	h.telegramClient = c
}

func (h *Handler) getTelegramClient() *telegram.Client {
	h.tgMu.RLock()
	defer h.tgMu.RUnlock()
	return h.telegramClient
}

// SetTelegramMessageHandler sets a fallback handler for non-/start Telegram webhook messages.
func (h *Handler) SetTelegramMessageHandler(fn func(context.Context, telegram.Update) error) {
	h.telegramFallback = fn
}

// Register mounts /auth, /me, and /users routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireUsersRead fiber.Handler,
	requireUsersWrite fiber.Handler,
) {
	auth := api.Group("/auth")
	// Static routes MUST be registered before parameterized /:provider routes.
	auth.Get("/providers", h.GetProviders)
	auth.Get("/roles", h.GetRoles)
	auth.Post("/otp", h.RequestOTP)
	auth.Post("/otp/verify", h.VerifyOTP)
	auth.Post("/logout", h.Logout)
	auth.Get("/:provider", h.OAuthRedirect)
	auth.Get("/:provider/callback", h.OAuthCallback)

	me := api.Group("/me", authMW, requireAuth)
	me.Get("", h.GetMe)
	me.Patch("", h.UpdateMe)
	me.Patch("/session", h.SwitchSessionParoki)
	me.Post("/sessions/search", h.SearchMySessions)
	me.Delete("/sessions/:id", h.RevokeMySession)
	me.Post("/accounts/search", h.SearchMyAccounts)
	me.Post("/accounts/otp", h.RequestLinkAccountOTP)
	me.Post("/accounts/verify", h.VerifyLinkAccount)
	me.Post("/accounts/telegram/token", h.CreateTelegramLinkToken)
	me.Delete("/accounts/:id", h.UnlinkMyAccount)

	api.Post("/telegram/hooks", h.TelegramWebhook)

	users := api.Group("/users", authMW, requireAuth)
	users.Post("/search", requireUsersRead, h.SearchUsers)
	users.Post("", requireUsersWrite, h.AdminCreateUser)
	users.Get("/:id", requireUsersRead, h.AdminGetUser)
	users.Patch("/:id", requireUsersWrite, h.AdminUpdateUser)
	users.Delete("/:id", requireUsersWrite, h.AdminDeleteUser)
}

func decodeStrictJSON(c fiber.Ctx, dst any) error {
	body := bytes.TrimSpace(c.Body())
	if len(body) > 0 {
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		if err := dec.Decode(dst); err != nil {
			return httpx.BadRequest("Invalid request body.")
		}
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			return httpx.BadRequest("Invalid request body.")
		}
	}
	if v := c.App().Config().StructValidator; v != nil {
		if err := v.Validate(dst); err != nil {
			return err
		}
	}
	return nil
}

func decodeEmptyOrStrictJSON(c fiber.Ctx, dst any) error {
	if len(bytes.TrimSpace(c.Body())) == 0 {
		return nil
	}
	return decodeStrictJSON(c, dst)
}

func parseCursorAndLimit(rawCursor string, rawLimit int) (*time.Time, *uuid.UUID, int, error) {
	limit := rawLimit
	if limit <= 0 {
		limit = httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		limit = httpx.MaxLimit
	}
	if rawCursor == "" {
		return nil, nil, limit, nil
	}
	cur, err := httpx.DecodeCursor(rawCursor)
	if err != nil {
		return nil, nil, 0, err
	}
	id, err := uuid.Parse(cur.ID)
	if err != nil {
		return nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	t := cur.CreatedAt
	return &t, &id, limit, nil
}
