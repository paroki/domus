package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchPeriods searches fiscal and liturgical periods in the active parish.
//
//	@Summary		Search periods
//	@Description	Searches fiscal and liturgical periods in the active parish with optional filters and cursor pagination.
//	@ID				searchPeriods
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchPeriodsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.PeriodResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/periods/search [post]
func (h *Handler) SearchPeriods(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchPeriodsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchPeriods(c.Context(), p.ParokiID, req, cursorT, cursorID, limit)
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

// GetPeriod fetches a single Period in the active parish.
//
//	@Summary		Get period
//	@Description	Returns a single non-deleted period in the active parish.
//	@ID				getPeriod
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Period UUID"
//	@Success		200	{object}	httpx.Envelope[domain.PeriodResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/periods/{id} [get]
func (h *Handler) GetPeriod(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid period ID.")
	}
	res, err := h.service.GetPeriod(c.Context(), p.ParokiID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreatePeriod creates a fiscal or liturgical Period in the active parish.
//
//	@Summary		Create period
//	@Description	Creates a period with date overlap validation and atomic isCurrent rotation.
//	@ID				createPeriod
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreatePeriodRequest	true	"Period creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.PeriodResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/periods [post]
func (h *Handler) CreatePeriod(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreatePeriodRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreatePeriod(c.Context(), p.ParokiID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdatePeriod updates an existing Period in the active parish.
//
//	@Summary		Update period
//	@Description	Updates mutable fields of a period with overlap validation and atomic isCurrent rotation.
//	@ID				updatePeriod
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Period UUID"
//	@Param			body	body		domain.UpdatePeriodRequest	true	"Period update payload"
//	@Success		200		{object}	httpx.Envelope[domain.PeriodResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/periods/{id} [patch]
func (h *Handler) UpdatePeriod(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid period ID.")
	}
	var req domain.UpdatePeriodRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdatePeriod(c.Context(), p.ParokiID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeletePeriod soft-deletes a Period in the active parish.
//
//	@Summary		Delete period
//	@Description	Soft-deletes a period in the active parish.
//	@ID				deletePeriod
//	@Tags			Ref
//	@Param			id	path	string	true	"Period UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/periods/{id} [delete]
func (h *Handler) DeletePeriod(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid period ID.")
	}
	if err := h.service.DeletePeriod(c.Context(), p.ParokiID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
