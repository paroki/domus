package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchWilayah searches wilayah/stasi in the active paroki.
//
//	@Summary		Search wilayah
//	@Description	Searches wilayah/stasi in the active paroki with optional query, jenis/active filters, and cursor pagination.
//	@ID				searchWilayah
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchWilayahRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.WilayahResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/wilayah/search [post]
func (h *Handler) SearchWilayah(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchWilayahRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchWilayah(c.Context(), p.ParokiID, req, cursorT, cursorID, limit)
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

// GetWilayah fetches a single Wilayah in the active paroki.
//
//	@Summary		Get wilayah
//	@Description	Returns a single non-deleted wilayah in the active paroki.
//	@ID				getWilayah
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Wilayah UUID"
//	@Success		200	{object}	httpx.Envelope[domain.WilayahResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/wilayah/{id} [get]
func (h *Handler) GetWilayah(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid wilayah ID.")
	}
	res, err := h.service.GetWilayah(c.Context(), p.ParokiID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateWilayah creates a Wilayah or Stasi in the active paroki.
//
//	@Summary		Create wilayah
//	@Description	Creates a wilayah/stasi or reactivates a soft-deleted match in the active paroki.
//	@ID				createWilayah
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateWilayahRequest	true	"Wilayah creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.WilayahResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/wilayah [post]
func (h *Handler) CreateWilayah(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateWilayahRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateWilayah(c.Context(), p.ParokiID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateWilayah updates an existing Wilayah or Stasi in the active paroki.
//
//	@Summary		Update wilayah
//	@Description	Updates mutable fields of a wilayah/stasi in the active paroki.
//	@ID				updateWilayah
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Wilayah UUID"
//	@Param			body	body		domain.UpdateWilayahRequest	true	"Wilayah update payload"
//	@Success		200		{object}	httpx.Envelope[domain.WilayahResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/wilayah/{id} [patch]
func (h *Handler) UpdateWilayah(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid wilayah ID.")
	}
	var req domain.UpdateWilayahRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateWilayah(c.Context(), p.ParokiID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteWilayah soft-deletes a Wilayah in the active paroki if no Lingkungan reference it.
//
//	@Summary		Delete wilayah
//	@Description	Soft-deletes a wilayah when no non-deleted lingkungan reference it.
//	@ID				deleteWilayah
//	@Tags			Ref
//	@Param			id	path	string	true	"Wilayah UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/wilayah/{id} [delete]
func (h *Handler) DeleteWilayah(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid wilayah ID.")
	}
	if err := h.service.DeleteWilayah(c.Context(), p.ParokiID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchLingkungan searches lingkungan in the active paroki.
//
//	@Summary		Search lingkungan
//	@Description	Searches lingkungan in the active paroki with optional query, wilayah/active filters, and cursor pagination.
//	@ID				searchLingkungan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchLingkunganRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.LingkunganResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/lingkungan/search [post]
func (h *Handler) SearchLingkungan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchLingkunganRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchLingkungan(c.Context(), p.ParokiID, req, cursorT, cursorID, limit)
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

// GetLingkungan fetches a single Lingkungan in the active paroki.
//
//	@Summary		Get lingkungan
//	@Description	Returns a single non-deleted lingkungan in the active paroki.
//	@ID				getLingkungan
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Lingkungan UUID"
//	@Success		200	{object}	httpx.Envelope[domain.LingkunganResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/lingkungan/{id} [get]
func (h *Handler) GetLingkungan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid lingkungan ID.")
	}
	res, err := h.service.GetLingkungan(c.Context(), p.ParokiID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateLingkungan creates a Lingkungan in the active paroki.
//
//	@Summary		Create lingkungan
//	@Description	Creates a lingkungan or reactivates a soft-deleted match in the active paroki.
//	@ID				createLingkungan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateLingkunganRequest	true	"Lingkungan creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.LingkunganResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/lingkungan [post]
func (h *Handler) CreateLingkungan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateLingkunganRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateLingkungan(c.Context(), p.ParokiID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateLingkungan updates an existing Lingkungan in the active paroki.
//
//	@Summary		Update lingkungan
//	@Description	Updates mutable fields of a lingkungan in the active paroki.
//	@ID				updateLingkungan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Lingkungan UUID"
//	@Param			body	body		domain.UpdateLingkunganRequest	true	"Lingkungan update payload"
//	@Success		200		{object}	httpx.Envelope[domain.LingkunganResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/lingkungan/{id} [patch]
func (h *Handler) UpdateLingkungan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid lingkungan ID.")
	}
	var req domain.UpdateLingkunganRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateLingkungan(c.Context(), p.ParokiID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteLingkungan soft-deletes a Lingkungan in the active paroki.
//
//	@Summary		Delete lingkungan
//	@Description	Soft-deletes a lingkungan in the active paroki.
//	@ID				deleteLingkungan
//	@Tags			Ref
//	@Param			id	path	string	true	"Lingkungan UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/lingkungan/{id} [delete]
func (h *Handler) DeleteLingkungan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid lingkungan ID.")
	}
	if err := h.service.DeleteLingkungan(c.Context(), p.ParokiID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
