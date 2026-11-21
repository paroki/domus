package handler

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchSessions searches the caller's agent conversation sessions in the active parish.
//
//	@Summary		Search agent sessions
//	@Description	Lists non-deleted agent conversation sessions owned by the caller with cursor pagination.
//	@ID				searchAgentSessions
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchAgentSessionsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.AgentSessionResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/search [post]
func (h *Handler) SearchSessions(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchAgentSessionsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	var channel string
	if req.Filter != nil {
		channel = req.Filter.Channel
	}
	rows, err := h.service.SearchSessions(c.Context(), domain.SearchSessionsInput{
		ParishID: p.ParishID,
		UserID:   p.UserID,
		Channel:  channel,
		CursorAt: cursorT,
		CursorID: cursorID,
		Limit:    limit + 1,
	})
	if err != nil {
		return err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.AgentSessionResponse, len(rows))
	for i, r := range rows {
		items[i] = toSessionResponse(r)
	}
	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = httpx.EncodeCursor(last.LastMessageAt, last.ID.String())
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// CreateSession creates a new web agent conversation session.
//
//	@Summary		Create agent session
//	@Description	Creates a new web channel agent conversation session for the caller in the active parish.
//	@ID				createAgentSession
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateAgentSessionRequest	false	"Create session request"
//	@Success		201		{object}	httpx.Envelope[domain.AgentSessionResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/agent/sessions [post]
func (h *Handler) CreateSession(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateAgentSessionRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	title := "Percakapan Baru"
	if req.Title != nil && strings.TrimSpace(*req.Title) != "" {
		title = strings.TrimSpace(*req.Title)
	}
	sess, err := h.service.CreateSession(c.Context(), domain.CreateSessionInput{
		ParishID: p.ParishID,
		UserID:   p.UserID,
		Channel:  domain.ChannelWeb,
		Title:    title,
	})
	if err != nil {
		return err
	}
	return httpx.Created(c, toSessionResponse(sess))
}

// GetSession returns a single agent session owned by the caller.
//
//	@Summary		Get agent session
//	@Description	Returns a non-deleted agent conversation session owned by the caller in the active parish.
//	@ID				getAgentSession
//	@Tags			Agent
//	@Produce		json
//	@Param			sessionId	path		string	true	"Session UUID"
//	@Success		200			{object}	httpx.Envelope[domain.AgentSessionResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId} [get]
func (h *Handler) GetSession(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid session ID.")
	}
	sess, err := h.service.GetSession(c.Context(), p.ParishID, p.UserID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, toSessionResponse(sess))
}

// UpdateSession renames an existing agent session owned by the caller.
//
//	@Summary		Update agent session
//	@Description	Updates the title of an existing agent conversation session owned by the caller.
//	@ID				updateAgentSession
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string								true	"Session UUID"
//	@Param			body		body		domain.UpdateAgentSessionRequest	true	"Update session request"
//	@Success		200			{object}	httpx.Envelope[domain.AgentSessionResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId} [patch]
func (h *Handler) UpdateSession(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid session ID.")
	}
	var req domain.UpdateAgentSessionRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return domain.ValidationError("title", "REQUIRED", "Title cannot be empty.")
	}
	sess, err := h.service.UpdateSessionTitle(c.Context(), p.ParishID, p.UserID, id, title)
	if err != nil {
		return err
	}
	return httpx.OK(c, toSessionResponse(sess))
}

// DeleteSession soft-deletes an existing agent session owned by the caller.
//
//	@Summary		Delete agent session
//	@Description	Soft-deletes an existing agent conversation session owned by the caller.
//	@ID				deleteAgentSession
//	@Tags			Agent
//	@Produce		json
//	@Param			sessionId	path	string	true	"Session UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId} [delete]
func (h *Handler) DeleteSession(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid session ID.")
	}
	if err := h.service.DeleteSession(c.Context(), p.ParishID, p.UserID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchSessionMessages lists or searches messages within a single agent session.
//
//	@Summary		Search agent session messages
//	@Description	Lists messages in a caller-owned session with keyset pagination or BM25 full-text search.
//	@ID				searchAgentSessionMessages
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			sessionId	path		string								true	"Session UUID"
//	@Param			body		body		domain.SearchAgentMessagesRequest	false	"Search parameters"
//	@Success		200			{object}	httpx.Envelope[[]domain.AgentMessageResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/sessions/{sessionId}/messages/search [post]
func (h *Handler) SearchSessionMessages(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("sessionId"))
	if err != nil {
		return httpx.BadRequest("Invalid session ID.")
	}
	if _, err := h.service.GetSession(c.Context(), p.ParishID, p.UserID, id); err != nil {
		return err
	}
	var req domain.SearchAgentMessagesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	rows, err := h.service.SearchSessionMessages(c.Context(), domain.SearchSessionMessagesInput{
		ParishID:  p.ParishID,
		UserID:    p.UserID,
		SessionID: id,
		Q:         req.Q,
		CursorAt:  cursorT,
		CursorID:  cursorID,
		Limit:     limit + 1,
	})
	if err != nil {
		return err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.AgentMessageResponse, len(rows))
	for i, r := range rows {
		items[i] = toMessageResponse(r)
	}
	var nextCursor string
	if strings.TrimSpace(req.Q) == "" && hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor = httpx.EncodeCursor(last.CreatedAt, last.ID.String())
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}
