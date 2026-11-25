package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchKeuskupan searches non-deleted keuskupan.
//
//	@Summary		Search keuskupan
//	@Description	Searches keuskupan with optional query, active filter, and cursor pagination.
//	@ID				searchKeuskupan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchKeuskupanRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.KeuskupanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/keuskupan/search [post]
func (h *Handler) SearchKeuskupan(c fiber.Ctx) error {
	var req domain.SearchKeuskupanRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchKeuskupan(c.Context(), req, cursorT, cursorID, limit)
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

// GetKeuskupan fetches a single Keuskupan by ID.
//
//	@Summary		Get keuskupan
//	@Description	Returns a single non-deleted keuskupan by ID.
//	@ID				getKeuskupan
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Keuskupan UUID"
//	@Success		200	{object}	httpx.Envelope[domain.KeuskupanResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/keuskupan/{id} [get]
func (h *Handler) GetKeuskupan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid keuskupan ID.")
	}
	res, err := h.service.GetKeuskupan(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateKeuskupan creates a new Keuskupan.
//
//	@Summary		Create keuskupan
//	@Description	Creates a keuskupan or reactivates a soft-deleted keuskupan with the same kode.
//	@ID				createKeuskupan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateKeuskupanRequest	true	"Keuskupan creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.KeuskupanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/keuskupan [post]
func (h *Handler) CreateKeuskupan(c fiber.Ctx) error {
	var req domain.CreateKeuskupanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateKeuskupan(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateKeuskupan updates an existing Keuskupan.
//
//	@Summary		Update keuskupan
//	@Description	Updates mutable fields of a keuskupan.
//	@ID				updateKeuskupan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Keuskupan UUID"
//	@Param			body	body		domain.UpdateKeuskupanRequest	true	"Keuskupan update payload"
//	@Success		200		{object}	httpx.Envelope[domain.KeuskupanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/keuskupan/{id} [patch]
func (h *Handler) UpdateKeuskupan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid keuskupan ID.")
	}
	var req domain.UpdateKeuskupanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateKeuskupan(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteKeuskupan soft-deletes a Keuskupan if it has no active paroki.
//
//	@Summary		Delete keuskupan
//	@Description	Soft-deletes a keuskupan when no active paroki reference it.
//	@ID				deleteKeuskupan
//	@Tags			Ref
//	@Param			id	path	string	true	"Keuskupan UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/keuskupan/{id} [delete]
func (h *Handler) DeleteKeuskupan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid keuskupan ID.")
	}
	if err := h.service.DeleteKeuskupan(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
