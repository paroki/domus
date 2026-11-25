package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchKevikepan searches non-deleted kevikepan.
//
//	@Summary		Search kevikepan
//	@Description	Searches kevikepan with optional query, keuskupanId/active filters, and cursor pagination.
//	@ID				searchKevikepan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchKevikepanRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.KevikepanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/kevikepan/search [post]
func (h *Handler) SearchKevikepan(c fiber.Ctx) error {
	var req domain.SearchKevikepanRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchKevikepan(c.Context(), req, cursorT, cursorID, limit)
	if err != nil {
		return err
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = httpx.EncodeCursor(last.CreatedAt, last.ID.String())
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// GetKevikepan fetches a single Kevikepan by ID.
//
//	@Summary		Get kevikepan
//	@Description	Returns a single non-deleted kevikepan by ID.
//	@ID				getKevikepan
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Kevikepan UUID"
//	@Success		200	{object}	httpx.Envelope[domain.KevikepanResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/kevikepan/{id} [get]
func (h *Handler) GetKevikepan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kevikepan ID.")
	}
	res, err := h.service.GetKevikepan(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateKevikepan creates a new Kevikepan.
//
//	@Summary		Create kevikepan
//	@Description	Creates a kevikepan or reactivates a soft-deleted kevikepan with the same keuskupanId and kode.
//	@ID				createKevikepan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateKevikepanRequest	true	"Kevikepan creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.KevikepanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kevikepan [post]
func (h *Handler) CreateKevikepan(c fiber.Ctx) error {
	var req domain.CreateKevikepanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateKevikepan(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateKevikepan updates an existing Kevikepan.
//
//	@Summary		Update kevikepan
//	@Description	Updates mutable fields of a kevikepan.
//	@ID				updateKevikepan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Kevikepan UUID"
//	@Param			body	body		domain.UpdateKevikepanRequest	true	"Kevikepan update payload"
//	@Success		200		{object}	httpx.Envelope[domain.KevikepanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kevikepan/{id} [patch]
func (h *Handler) UpdateKevikepan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kevikepan ID.")
	}
	var req domain.UpdateKevikepanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateKevikepan(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteKevikepan soft-deletes a Kevikepan.
//
//	@Summary		Delete kevikepan
//	@Description	Soft-deletes a kevikepan.
//	@ID				deleteKevikepan
//	@Tags			Ref
//	@Param			id	path	string	true	"Kevikepan UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/kevikepan/{id} [delete]
func (h *Handler) DeleteKevikepan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kevikepan ID.")
	}
	if err := h.service.DeleteKevikepan(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
