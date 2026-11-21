package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SyncProviderModels discovers available models for a configured AI provider.
//
//	@Summary		Sync provider models
//	@Description	Discovers models from the configured upstream AI provider and persists the catalog.
//	@ID				syncAgentProviderModels
//	@Tags			Agent
//	@Produce		json
//	@Param			provider	path		string	true	"AI provider name (gemini, anthropic, openai, openrouter)"
//	@Success		200			{object}	httpx.Envelope[domain.SyncProviderModelsResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		503			{object}	httpx.Envelope[any]
//	@Router			/agent/providers/{provider}/sync-models [post]
func (h *Handler) SyncProviderModels(c fiber.Ctx) error {
	if h.syncer == nil {
		return domain.ErrProviderUnavailable
	}
	res, err := h.syncer.SyncProviderModels(c.Context(), c.Params("provider"))
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}
