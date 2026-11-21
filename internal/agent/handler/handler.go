package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/platform/httpx"
)

// ProviderSyncer discovers and persists available models for an AI provider.
type ProviderSyncer interface {
	SyncProviderModels(ctx context.Context, provider string) (domain.SyncProviderModelsResponse, error)
}

// Handler exposes HTTP and SSE endpoints for the AI agent feature.
type Handler struct {
	service *service.Service
	syncer  ProviderSyncer
}

// New constructs an agent Handler backed by svc and syncer.
func New(svc *service.Service, syncer ProviderSyncer) *Handler {
	return &Handler{service: svc, syncer: syncer}
}

// Register mounts /agent routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireAdminWrite fiber.Handler,
	requireSessionsRead fiber.Handler,
	requireSessionsWrite fiber.Handler,
	requireMemoriesRead fiber.Handler,
	requireMemoriesWrite fiber.Handler,
) {
	agentGroup := api.Group("/agent")
	agentGroup.Post("/webhooks/whatsapp", h.WhatsAppWebhook)

	authed := agentGroup.Group("", authMW, requireAuth)
	authed.Post("/providers/:provider/sync-models", requireAdminWrite, h.SyncProviderModels)

	tenant := authed.Group("", requireTenant)
	tenant.Post("/sessions/search", requireSessionsRead, h.SearchSessions)
	tenant.Post("/sessions", requireSessionsWrite, h.CreateSession)
	tenant.Get("/sessions/:sessionId", requireSessionsRead, h.GetSession)
	tenant.Patch("/sessions/:sessionId", requireSessionsWrite, h.UpdateSession)
	tenant.Delete("/sessions/:sessionId", requireSessionsWrite, h.DeleteSession)
	tenant.Post("/sessions/:sessionId/messages/search", requireSessionsRead, h.SearchSessionMessages)

	tenant.Post("/memories/search", requireMemoriesRead, h.SearchMemories)
	tenant.Post("/memories", requireMemoriesWrite, h.CreateMemory)
	tenant.Patch("/memories/:memoryId", requireMemoriesWrite, h.UpdateMemory)
	tenant.Delete("/memories/:memoryId", requireMemoriesWrite, h.DeleteMemory)

	tenant.Post("/chat", requireSessionsWrite, h.Chat)
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
	t := cur.CreatedAt.Local()
	return &t, &id, limit, nil
}

func toSessionResponse(s *ent.AgentSession) domain.AgentSessionResponse {
	return domain.AgentSessionResponse{
		ID:            s.ID,
		ParishID:      s.ParishID,
		UserID:        s.UserID,
		Channel:       string(s.Channel),
		Title:         s.Title,
		LastMessageAt: s.LastMessageAt,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

func toMessageResponse(m *ent.AgentMessage) domain.AgentMessageResponse {
	var tcs []domain.AgentToolCallDTO
	if len(m.ToolCalls) > 0 {
		tcs = make([]domain.AgentToolCallDTO, len(m.ToolCalls))
		for i, tc := range m.ToolCalls {
			tcs[i] = domain.AgentToolCallDTO{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
				AgentName: tc.AgentName,
			}
		}
	}
	return domain.AgentMessageResponse{
		ID:         m.ID,
		SessionID:  m.SessionID,
		AgentName:  m.AgentName,
		Role:       string(m.Role),
		Content:    m.Content,
		Thinking:   m.Thinking,
		ToolCalls:  tcs,
		ToolCallID: m.ToolCallID,
		ToolName:   m.ToolName,
		CreatedAt:  m.CreatedAt,
	}
}

func toMemoryResponse(m *ent.AgentMemory) domain.AgentMemoryResponse {
	return domain.AgentMemoryResponse{
		ID:              m.ID,
		ParishID:        m.ParishID,
		UserID:          m.UserID,
		Scope:           string(m.Scope),
		Category:        string(m.Category),
		Content:         m.Content,
		IsPinned:        m.IsPinned,
		SourceSessionID: m.SourceSessionID,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
