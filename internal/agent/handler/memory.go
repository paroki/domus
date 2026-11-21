package handler

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchMemories searches personal and parish-scoped long-term agent memories.
//
//	@Summary		Search agent memories
//	@Description	Searches personal and active parish long-term memories with optional filters and cursor pagination.
//	@ID				searchAgentMemories
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchAgentMemoriesRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.AgentMemoryResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/agent/memories/search [post]
func (h *Handler) SearchMemories(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchAgentMemoriesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	var scope, category string
	var isPinned *bool
	if req.Filter != nil {
		scope = req.Filter.Scope
		category = req.Filter.Category
		isPinned = req.Filter.IsPinned
	}
	rows, err := h.service.SearchMemories(c.Context(), domain.SearchMemoriesInput{
		ParishID: p.ParishID,
		UserID:   p.UserID,
		Scope:    scope,
		Category: category,
		IsPinned: isPinned,
		Q:        req.Q,
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
	items := make([]domain.AgentMemoryResponse, len(rows))
	for i, r := range rows {
		items[i] = toMemoryResponse(r)
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

// CreateMemory creates a new personal or parish-scoped long-term agent memory.
//
//	@Summary		Create agent memory
//	@Description	Creates a new long-term agent memory fact. Parish-scoped memories require ref:parish-admin.
//	@ID				createAgentMemory
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateAgentMemoryRequest	true	"Create memory request"
//	@Success		201		{object}	httpx.Envelope[domain.AgentMemoryResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/agent/memories [post]
func (h *Handler) CreateMemory(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateAgentMemoryRequest
	if err := decodeJSONOnly(c, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Scope) == "" {
		req.Scope = string(domain.MemoryScopeUser)
	}
	if v := c.App().Config().StructValidator; v != nil {
		if err := v.Validate(&req); err != nil {
			return err
		}
	}
	if strings.TrimSpace(req.Content) == "" {
		return domain.ValidationError("content", "REQUIRED", "Content cannot be empty.")
	}
	row, err := h.service.CreateMemory(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, toMemoryResponse(row))
}

// UpdateMemory updates an existing personal or parish-scoped long-term agent memory.
//
//	@Summary		Update agent memory
//	@Description	Updates an existing long-term memory fact. Parish-scoped memories require ref:parish-admin.
//	@ID				updateAgentMemory
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			memoryId	path		string							true	"Memory UUID"
//	@Param			body		body		domain.UpdateAgentMemoryRequest	true	"Update memory request"
//	@Success		200			{object}	httpx.Envelope[domain.AgentMemoryResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/agent/memories/{memoryId} [patch]
func (h *Handler) UpdateMemory(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("memoryId"))
	if err != nil {
		return httpx.BadRequest("Invalid memory ID.")
	}
	var req domain.UpdateAgentMemoryRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	if req.Content != nil && strings.TrimSpace(*req.Content) == "" {
		return domain.ValidationError("content", "REQUIRED", "Content cannot be empty.")
	}
	row, err := h.service.UpdateMemory(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, toMemoryResponse(row))
}

// DeleteMemory soft-deletes an existing personal or parish-scoped long-term agent memory.
//
//	@Summary		Delete agent memory
//	@Description	Soft-deletes an existing long-term memory fact. Parish-scoped memories require ref:parish-admin.
//	@ID				deleteAgentMemory
//	@Tags			Agent
//	@Produce		json
//	@Param			memoryId	path	string	true	"Memory UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/agent/memories/{memoryId} [delete]
func (h *Handler) DeleteMemory(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("memoryId"))
	if err != nil {
		return httpx.BadRequest("Invalid memory ID.")
	}
	if err := h.service.DeleteMemory(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
