package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// ProviderManager manages LLM provider configs and model discovery.
type ProviderManager interface {
	ListProviders(ctx context.Context) ([]domain.AgentProviderResponse, error)
	CreateProvider(ctx context.Context, req domain.CreateAgentProviderRequest) (domain.AgentProviderResponse, error)
	UpdateProvider(ctx context.Context, id uuid.UUID, req domain.UpdateAgentProviderRequest) (domain.AgentProviderResponse, error)
	DeleteProvider(ctx context.Context, id uuid.UUID) error
	SyncProviderModels(ctx context.Context, id uuid.UUID) (domain.SyncProviderModelsResponse, error)
}

// SessionManager manages AI conversation sessions for the authenticated user.
type SessionManager interface {
	SearchSessions(ctx context.Context, p authx.Principal, req domain.SearchSessionsRequest) ([]domain.SessionResponse, *string, bool, error)
	GetSession(ctx context.Context, p authx.Principal, sessionID uuid.UUID) (domain.SessionDetailResponse, error)
	UpdateSession(ctx context.Context, p authx.Principal, sessionID uuid.UUID, req domain.UpdateSessionRequest) (domain.SessionResponse, error)
	DeleteSession(ctx context.Context, p authx.Principal, sessionID uuid.UUID) error
}

// MemoryManager manages long-term AI memories for the authenticated user.
type MemoryManager interface {
	SearchMemories(ctx context.Context, p authx.Principal, req domain.SearchMemoriesRequest) ([]domain.MemoryResponse, *string, bool, error)
	CreateMemory(ctx context.Context, p authx.Principal, req domain.CreateMemoryRequest) (domain.MemoryResponse, error)
	UpdateMemory(ctx context.Context, p authx.Principal, memoryID uuid.UUID, req domain.UpdateMemoryRequest) (domain.MemoryResponse, error)
	DeleteMemory(ctx context.Context, p authx.Principal, memoryID uuid.UUID) error
}

// Handler exposes HTTP endpoints for AI agent providers, sessions, and memories.
type Handler struct {
	providers ProviderManager
	sessions  SessionManager
	memories  MemoryManager
}

// New constructs an agent Handler.
func New(providers ProviderManager, sessions SessionManager, memories MemoryManager) *Handler {
	return &Handler{
		providers: providers,
		sessions:  sessions,
		memories:  memories,
	}
}

// Register mounts /agent/* routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireAdminRead fiber.Handler,
	requireAdminWrite fiber.Handler,
	requireChatRead fiber.Handler,
	requireChatWrite fiber.Handler,
) {
	authed := api.Group("/agent", authMW, requireAuth)
	authed.Get("/providers", requireAdminRead, h.ListProviders)
	authed.Post("/providers", requireAdminWrite, h.CreateProvider)
	authed.Patch("/providers/:providerId", requireAdminWrite, h.UpdateProvider)
	authed.Delete("/providers/:providerId", requireAdminWrite, h.DeleteProvider)
	authed.Post("/providers/:providerId/sync-models", requireAdminWrite, h.SyncProviderModels)

	tenanted := authed.Group("", requireTenant)
	tenanted.Post("/sessions/search", requireChatRead, h.SearchSessions)
	tenanted.Get("/sessions/:sessionId", requireChatRead, h.GetSession)
	tenanted.Patch("/sessions/:sessionId", requireChatWrite, h.UpdateSession)
	tenanted.Delete("/sessions/:sessionId", requireChatWrite, h.DeleteSession)

	tenanted.Post("/memories/search", requireChatRead, h.SearchMemories)
	tenanted.Post("/memories", requireChatWrite, h.CreateMemory)
	tenanted.Patch("/memories/:memoryId", requireChatWrite, h.UpdateMemory)
	tenanted.Delete("/memories/:memoryId", requireChatWrite, h.DeleteMemory)
}

func decodeJSONOnly(c fiber.Ctx, dst any) error {
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
	return nil
}

func decodeStrictJSON(c fiber.Ctx, dst any) error {
	if err := decodeJSONOnly(c, dst); err != nil {
		return err
	}
	if v := c.App().Config().StructValidator; v != nil {
		if err := v.Validate(dst); err != nil {
			return err
		}
	}
	return nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		return httpx.MaxLimit
	}
	return limit
}
