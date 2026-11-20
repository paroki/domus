package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchDioceses searches non-deleted dioceses.
//
//	@Summary		Search dioceses
//	@Description	Searches dioceses with optional query, active filter, and cursor pagination.
//	@ID				searchDioceses
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchDiocesesRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.DioceseResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/dioceses/search [post]
func (h *Handler) SearchDioceses(c fiber.Ctx) error {
	var req domain.SearchDiocesesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchDioceses(c.Context(), req, cursorT, cursorID, limit)
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

// GetDiocese fetches a single Diocese by ID.
//
//	@Summary		Get diocese
//	@Description	Returns a single non-deleted diocese by ID.
//	@ID				getDiocese
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Diocese UUID"
//	@Success		200	{object}	httpx.Envelope[domain.DioceseResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/dioceses/{id} [get]
func (h *Handler) GetDiocese(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid diocese ID.")
	}
	res, err := h.service.GetDiocese(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateDiocese creates a new Diocese.
//
//	@Summary		Create diocese
//	@Description	Creates a diocese or reactivates a soft-deleted diocese with the same code.
//	@ID				createDiocese
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateDioceseRequest	true	"Diocese creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.DioceseResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/dioceses [post]
func (h *Handler) CreateDiocese(c fiber.Ctx) error {
	var req domain.CreateDioceseRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateDiocese(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateDiocese updates an existing Diocese.
//
//	@Summary		Update diocese
//	@Description	Updates mutable fields of a diocese.
//	@ID				updateDiocese
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Diocese UUID"
//	@Param			body	body		domain.UpdateDioceseRequest	true	"Diocese update payload"
//	@Success		200		{object}	httpx.Envelope[domain.DioceseResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/dioceses/{id} [patch]
func (h *Handler) UpdateDiocese(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid diocese ID.")
	}
	var req domain.UpdateDioceseRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateDiocese(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteDiocese soft-deletes a Diocese if it has no active parishes.
//
//	@Summary		Delete diocese
//	@Description	Soft-deletes a diocese when no active parishes reference it.
//	@ID				deleteDiocese
//	@Tags			Ref
//	@Param			id	path	string	true	"Diocese UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/dioceses/{id} [delete]
func (h *Handler) DeleteDiocese(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid diocese ID.")
	}
	if err := h.service.DeleteDiocese(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
