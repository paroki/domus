package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchMemories lists or searches the caller's long-term AI memories in the active parish.
//
//	@Summary		Search AI memories
//	@Description	Lists memories by keyset cursor when q is empty, or runs ParadeDB RRF hybrid / BM25 search when q is provided.
//	@ID				searchAgentMemories
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchMemoriesRequest	true	"Search params"
//	@Success		200		{object}	httpx.Envelope[[]domain.MemoryResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/agent/memories/search [post]
func (h *Handler) SearchMemories(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	var req domain.SearchMemoriesRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	items, nextCursor, hasMore, err := h.memories.SearchMemories(c.Context(), p, req)
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

// CreateMemory creates a long-term AI memory item for the caller.
//
//	@Summary		Create AI memory
//	@Description	Creates a new AI memory item and best-effort computes its 768-dim embedding.
//	@ID				createAgentMemory
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateMemoryRequest	true	"Memory payload"
//	@Success		201		{object}	httpx.Envelope[domain.MemoryResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/agent/memories [post]
func (h *Handler) CreateMemory(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	var req domain.CreateMemoryRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.memories.CreateMemory(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateMemory updates an existing AI memory item owned by the caller.
//
//	@Summary		Update AI memory
//	@Description	Updates an AI memory item's content and/or category and re-computes its embedding when content changes.
//	@ID				updateAgentMemory
//	@Tags			Agent
//	@Accept			json
//	@Produce		json
//	@Param			memoryId	path		string						true	"Memory UUID"
//	@Param			body		body		domain.UpdateMemoryRequest	true	"Update payload"
//	@Success		200			{object}	httpx.Envelope[domain.MemoryResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/agent/memories/{memoryId} [patch]
func (h *Handler) UpdateMemory(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("memoryId"))
	if err != nil {
		return httpx.BadRequest("Invalid memoryId.")
	}
	var req domain.UpdateMemoryRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.memories.UpdateMemory(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteMemory hard-deletes an AI memory item owned by the caller.
//
//	@Summary		Delete AI memory
//	@Description	Hard-deletes an AI memory item owned by the caller.
//	@ID				deleteAgentMemory
//	@Tags			Agent
//	@Produce		json
//	@Param			memoryId	path		string	true	"Memory UUID"
//	@Success		204			{string}	string	"No Content"
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/agent/memories/{memoryId} [delete]
func (h *Handler) DeleteMemory(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("memoryId"))
	if err != nil {
		return httpx.BadRequest("Invalid memoryId.")
	}
	if err := h.memories.DeleteMemory(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
