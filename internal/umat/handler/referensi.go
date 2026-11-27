package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/umat/domain"
)

// SearchReferensi searches global and/or parish-scoped umat referensi entries.
//
//	@Summary		Search umat referensi
//	@Description	Searches global and/or parish-scoped Data Umat referensi entries with tenant override on matching (category, code).
//	@ID				searchUmatReferensi
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchUmatReferensiRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.UmatReferensiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/referensi/search [post]
func (h *Handler) SearchReferensi(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchUmatReferensiRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchReferensi(c.Context(), p, req, cursorT, cursorID, limit)
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

// GetReferensi fetches a single UmatReferensi (global or belonging to active parish).
//
//	@Summary		Get umat referensi
//	@Description	Returns a single non-deleted umat referensi if global or owned by the active parish.
//	@ID				getUmatReferensi
//	@Tags			Umat
//	@Produce		json
//	@Param			id	path		string	true	"Referensi UUID"
//	@Success		200	{object}	httpx.Envelope[domain.UmatReferensiResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/umat/referensi/{id} [get]
func (h *Handler) GetReferensi(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid referensi ID.")
	}
	res, err := h.service.GetReferensi(c.Context(), p, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateReferensi creates a global or parish-scoped UmatReferensi.
//
//	@Summary		Create umat referensi
//	@Description	Creates a global (super-admin) or parish-scoped (umat:admin) umat referensi, reactivating soft-deleted matches.
//	@ID				createUmatReferensi
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateUmatReferensiRequest	true	"Referensi creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.UmatReferensiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/referensi [post]
func (h *Handler) CreateReferensi(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateUmatReferensiRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateReferensi(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateReferensi updates a global or parish-scoped UmatReferensi.
//
//	@Summary		Update umat referensi
//	@Description	Updates mutable fields of a global (super-admin) or parish-scoped (umat:admin) umat referensi.
//	@ID				updateUmatReferensi
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Referensi UUID"
//	@Param			body	body		domain.UpdateUmatReferensiRequest	true	"Referensi update payload"
//	@Success		200		{object}	httpx.Envelope[domain.UmatReferensiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/referensi/{id} [patch]
func (h *Handler) UpdateReferensi(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid referensi ID.")
	}
	var req domain.UpdateUmatReferensiRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateReferensi(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteReferensi soft-deletes a global or parish-scoped UmatReferensi.
//
//	@Summary		Delete umat referensi
//	@Description	Soft-deletes a global (super-admin) or parish-scoped (umat:admin) umat referensi if not in use.
//	@ID				deleteUmatReferensi
//	@Tags			Umat
//	@Param			id	path	string	true	"Referensi UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/umat/referensi/{id} [delete]
func (h *Handler) DeleteReferensi(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid referensi ID.")
	}
	if err := h.service.DeleteReferensi(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

func parseCursorAndLimit(rawCursor string, rawLimit int) (*time.Time, *uuid.UUID, int, error) {
	limit := normalizeLimit(rawLimit)
	if rawCursor == "" {
		return nil, nil, limit, nil
	}
	cur, err := httpx.DecodeCursor(rawCursor)
	if err != nil {
		return nil, nil, 0, err
	}
	id, err := uuid.Parse(cur.ID)
	if err != nil {
		return nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	t := cur.CreatedAt
	return &t, &id, limit, nil
}
