package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchSessions searches the caller's AI conversation sessions in the active parish.
//
//	@Summary		Search AI sessions
//	@Description	Returns keyset-paginated AI conversation sessions owned by the caller in the active parish.
//	@ID				searchAgentSessions
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchSessionsRequest	true	"Search params"
//	@Success		200		{object}	httpx.Envelope[[]domain.SessionResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/search [post]
func (h *Handler) SearchSessions(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	var req domain.SearchSessionsRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	items, nextCursor, hasMore, err := h.sessions.SearchSessions(c.Context(), p, req)
	if err != nil {
		return err
	}
	var cur string
	if nextCursor != nil {
		cur = *nextCursor
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: cur,
		HasMore:    hasMore,
		Limit:      clampLimit(req.Limit),
	})
}

// GetSession returns an AI conversation session and its chronological messages.
//
//	@Summary		Get AI session
//	@Description	Returns an AI conversation session owned by the caller with its chronological messages.
//	@ID				getAgentSession
//	@Tags			Agent
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session UUID"
//	@Success		200			{object}	httpx.Envelope[domain.SessionDetailResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId} [get]
func (h *Handler) GetSession(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid sessionId.")
	}
	res, err := h.sessions.GetSession(c.Context(), p, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// UpdateSession updates an AI conversation session's title or model override.
//
//	@Summary		Update AI session
//	@Description	Updates session title and/or provider/model override (empty string clears override to default).
//	@ID				updateAgentSession
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string						true	"Session UUID"
//	@Param			body		body		domain.UpdateSessionRequest	true	"Update payload"
//	@Success		200			{object}	httpx.Envelope[domain.SessionResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId} [patch]
func (h *Handler) UpdateSession(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid sessionId.")
	}
	var req domain.UpdateSessionRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.sessions.UpdateSession(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteSession soft-deletes an AI conversation session.
//
//	@Summary		Delete AI session
//	@Description	Soft-deletes an AI conversation session owned by the caller.
//	@ID				deleteAgentSession
//	@Tags			Agent
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session UUID"
//	@Success		204			{string}	string	"No Content"
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId} [delete]
func (h *Handler) DeleteSession(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid sessionId.")
	}
	if err := h.sessions.DeleteSession(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
