package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
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

// Handler exposes HTTP endpoints for the AI agent provider setup.
type Handler struct {
	providers ProviderManager
}

// New constructs an agent Handler backed by providers.
func New(providers ProviderManager) *Handler {
	return &Handler{providers: providers}
}

// Register mounts /agent/providers routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireAdminRead fiber.Handler,
	requireAdminWrite fiber.Handler,
) {
	authed := api.Group("/agent", authMW, requireAuth)
	authed.Get("/providers", requireAdminRead, h.ListProviders)
	authed.Post("/providers", requireAdminWrite, h.CreateProvider)
	authed.Patch("/providers/:providerId", requireAdminWrite, h.UpdateProvider)
	authed.Delete("/providers/:providerId", requireAdminWrite, h.DeleteProvider)
	authed.Post("/providers/:providerId/sync-models", requireAdminWrite, h.SyncProviderModels)
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
