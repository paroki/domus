package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchLookups searches global and/or parish-scoped lookups.
//
//	@Summary		Search lookups
//	@Description	Searches global and/or parish-scoped lookups with tenant override on matching (category, code).
//	@ID				searchLookups
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchLookupsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.LookupResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/lookups/search [post]
func (h *Handler) SearchLookups(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchLookupsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchLookups(c.Context(), p, req, cursorT, cursorID, limit)
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

// GetLookup fetches a single Lookup (global or belonging to active parish).
//
//	@Summary		Get lookup
//	@Description	Returns a single non-deleted lookup if global or owned by the active parish.
//	@ID				getLookup
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Lookup UUID"
//	@Success		200	{object}	httpx.Envelope[domain.LookupResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/lookups/{id} [get]
func (h *Handler) GetLookup(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid lookup ID.")
	}
	res, err := h.service.GetLookup(c.Context(), p, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateLookup creates a global or parish-scoped Lookup.
//
//	@Summary		Create lookup
//	@Description	Creates a global (super-admin) or parish-scoped (parish-admin) lookup, reactivating soft-deleted matches.
//	@ID				createLookup
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateLookupRequest	true	"Lookup creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.LookupResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/lookups [post]
func (h *Handler) CreateLookup(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateLookupRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateLookup(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateLookup updates a global or parish-scoped Lookup.
//
//	@Summary		Update lookup
//	@Description	Updates mutable fields of a global (super-admin) or parish-scoped (parish-admin) lookup.
//	@ID				updateLookup
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Lookup UUID"
//	@Param			body	body		domain.UpdateLookupRequest	true	"Lookup update payload"
//	@Success		200		{object}	httpx.Envelope[domain.LookupResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/lookups/{id} [patch]
func (h *Handler) UpdateLookup(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid lookup ID.")
	}
	var req domain.UpdateLookupRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateLookup(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteLookup soft-deletes a global or parish-scoped Lookup.
//
//	@Summary		Delete lookup
//	@Description	Soft-deletes a global (super-admin) or parish-scoped (parish-admin) lookup.
//	@ID				deleteLookup
//	@Tags			Ref
//	@Param			id	path	string	true	"Lookup UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/lookups/{id} [delete]
func (h *Handler) DeleteLookup(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid lookup ID.")
	}
	if err := h.service.DeleteLookup(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
