package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/httpx"
)

// ListProviders returns all configured LLM provider entries.
//
//	@Summary		List AI provider configs
//	@Description	Lists all configured LLM provider entries with masked API keys and synced models.
//	@ID				listAgentProviders
//	@Tags			Agent
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[[]domain.AgentProviderResponse]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Router			/agent/providers [get]
func (h *Handler) ListProviders(c fiber.Ctx) error {
	if h.providers == nil {
		return domain.ErrProviderUnavailable
	}
	res, err := h.providers.ListProviders(c.Context())
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateProvider creates a new LLM provider configuration.
//
//	@Summary		Create AI provider config
//	@Description	Creates a new LLM provider configuration, encrypts its API key, and best-effort syncs models.
//	@ID				createAgentProvider
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateAgentProviderRequest	true	"Provider config payload"
//	@Success		201		{object}	httpx.Envelope[domain.AgentProviderResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/agent/providers [post]
func (h *Handler) CreateProvider(c fiber.Ctx) error {
	if h.providers == nil {
		return domain.ErrProviderUnavailable
	}
	var req domain.CreateAgentProviderRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.providers.CreateProvider(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateProvider updates an existing LLM provider configuration.
//
//	@Summary		Update AI provider config
//	@Description	Updates an LLM provider configuration; empty apiKey preserves the existing key.
//	@ID				updateAgentProvider
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			providerId	path		string								true	"Provider config UUID"
//	@Param			body		body		domain.UpdateAgentProviderRequest	true	"Updated fields"
//	@Success		200			{object}	httpx.Envelope[domain.AgentProviderResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/agent/providers/{providerId} [patch]
func (h *Handler) UpdateProvider(c fiber.Ctx) error {
	if h.providers == nil {
		return domain.ErrProviderUnavailable
	}
	id, err := uuid.Parse(c.Params("providerId"))
	if err != nil {
		return httpx.BadRequest("Invalid providerId.")
	}
	var req domain.UpdateAgentProviderRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.providers.UpdateProvider(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteProvider removes an LLM provider configuration.
//
//	@Summary		Delete AI provider config
//	@Description	Deletes an LLM provider configuration.
//	@ID				deleteAgentProvider
//	@Tags			Agent
//	@Produce		json
//	@Param			providerId	path		string	true	"Provider config UUID"
//	@Success		204			{object}	nil
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/providers/{providerId} [delete]
func (h *Handler) DeleteProvider(c fiber.Ctx) error {
	if h.providers == nil {
		return domain.ErrProviderUnavailable
	}
	id, err := uuid.Parse(c.Params("providerId"))
	if err != nil {
		return httpx.BadRequest("Invalid providerId.")
	}
	if err := h.providers.DeleteProvider(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SyncProviderModels discovers available models for a configured AI provider.
//
//	@Summary		Sync provider models
//	@Description	Discovers models from the configured upstream AI provider and persists the catalog.
//	@ID				syncAgentProviderModels
//	@Tags			Agent
//	@Produce		json
//	@Param			providerId	path		string	true	"Provider config UUID"
//	@Success		200			{object}	httpx.Envelope[domain.SyncProviderModelsResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/providers/{providerId}/sync-models [post]
func (h *Handler) SyncProviderModels(c fiber.Ctx) error {
	if h.providers == nil {
		return domain.ErrProviderUnavailable
	}
	id, err := uuid.Parse(c.Params("providerId"))
	if err != nil {
		return httpx.BadRequest("Invalid providerId.")
	}
	res, err := h.providers.SyncProviderModels(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}
