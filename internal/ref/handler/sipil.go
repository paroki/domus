package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchKota searches non-deleted kota.
//
//	@Summary		Search kota
//	@Description	Searches kota with optional query, active filter, and cursor pagination.
//	@ID				searchKota
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchKotaRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.KotaResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/kota/search [post]
func (h *Handler) SearchKota(c fiber.Ctx) error {
	var req domain.SearchKotaRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchKota(c.Context(), req, cursorT, cursorID, limit)
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

// GetKota fetches a single Kota by ID.
//
//	@Summary		Get kota
//	@Description	Returns a single non-deleted kota by ID.
//	@ID				getKota
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Kota UUID"
//	@Success		200	{object}	httpx.Envelope[domain.KotaResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/kota/{id} [get]
func (h *Handler) GetKota(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kota ID.")
	}
	res, err := h.service.GetKota(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateKota creates a new Kota.
//
//	@Summary		Create kota
//	@Description	Creates a kota or reactivates a soft-deleted kota with the same kode.
//	@ID				createKota
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateKotaRequest	true	"Kota creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.KotaResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kota [post]
func (h *Handler) CreateKota(c fiber.Ctx) error {
	var req domain.CreateKotaRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateKota(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateKota updates an existing Kota.
//
//	@Summary		Update kota
//	@Description	Updates mutable fields of a kota.
//	@ID				updateKota
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Kota UUID"
//	@Param			body	body		domain.UpdateKotaRequest	true	"Kota update payload"
//	@Success		200		{object}	httpx.Envelope[domain.KotaResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kota/{id} [patch]
func (h *Handler) UpdateKota(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kota ID.")
	}
	var req domain.UpdateKotaRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateKota(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteKota soft-deletes a Kota if it has no non-deleted kecamatan.
//
//	@Summary		Delete kota
//	@Description	Soft-deletes a kota when no non-deleted kecamatan reference it.
//	@ID				deleteKota
//	@Tags			Ref
//	@Param			id	path	string	true	"Kota UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/kota/{id} [delete]
func (h *Handler) DeleteKota(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kota ID.")
	}
	if err := h.service.DeleteKota(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchKecamatan searches non-deleted kecamatan.
//
//	@Summary		Search kecamatan
//	@Description	Searches kecamatan with optional query, kotaId/active filters, and cursor pagination.
//	@ID				searchKecamatan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchKecamatanRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.KecamatanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/kecamatan/search [post]
func (h *Handler) SearchKecamatan(c fiber.Ctx) error {
	var req domain.SearchKecamatanRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchKecamatan(c.Context(), req, cursorT, cursorID, limit)
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

// GetKecamatan fetches a single Kecamatan by ID.
//
//	@Summary		Get kecamatan
//	@Description	Returns a single non-deleted kecamatan by ID.
//	@ID				getKecamatan
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Kecamatan UUID"
//	@Success		200	{object}	httpx.Envelope[domain.KecamatanResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/kecamatan/{id} [get]
func (h *Handler) GetKecamatan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kecamatan ID.")
	}
	res, err := h.service.GetKecamatan(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateKecamatan creates a new Kecamatan.
//
//	@Summary		Create kecamatan
//	@Description	Creates a kecamatan or reactivates a soft-deleted kecamatan with the same kotaId and kode.
//	@ID				createKecamatan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateKecamatanRequest	true	"Kecamatan creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.KecamatanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kecamatan [post]
func (h *Handler) CreateKecamatan(c fiber.Ctx) error {
	var req domain.CreateKecamatanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateKecamatan(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateKecamatan updates an existing Kecamatan.
//
//	@Summary		Update kecamatan
//	@Description	Updates mutable fields of a kecamatan.
//	@ID				updateKecamatan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Kecamatan UUID"
//	@Param			body	body		domain.UpdateKecamatanRequest	true	"Kecamatan update payload"
//	@Success		200		{object}	httpx.Envelope[domain.KecamatanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kecamatan/{id} [patch]
func (h *Handler) UpdateKecamatan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kecamatan ID.")
	}
	var req domain.UpdateKecamatanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateKecamatan(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteKecamatan soft-deletes a Kecamatan if it has no non-deleted kelurahan.
//
//	@Summary		Delete kecamatan
//	@Description	Soft-deletes a kecamatan when no non-deleted kelurahan reference it.
//	@ID				deleteKecamatan
//	@Tags			Ref
//	@Param			id	path	string	true	"Kecamatan UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/kecamatan/{id} [delete]
func (h *Handler) DeleteKecamatan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kecamatan ID.")
	}
	if err := h.service.DeleteKecamatan(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchKelurahan searches non-deleted kelurahan.
//
//	@Summary		Search kelurahan
//	@Description	Searches kelurahan with optional query, kecamatanId/active filters, and cursor pagination.
//	@ID				searchKelurahan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchKelurahanRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.KelurahanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/kelurahan/search [post]
func (h *Handler) SearchKelurahan(c fiber.Ctx) error {
	var req domain.SearchKelurahanRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchKelurahan(c.Context(), req, cursorT, cursorID, limit)
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

// GetKelurahan fetches a single Kelurahan by ID.
//
//	@Summary		Get kelurahan
//	@Description	Returns a single non-deleted kelurahan by ID.
//	@ID				getKelurahan
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Kelurahan UUID"
//	@Success		200	{object}	httpx.Envelope[domain.KelurahanResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/kelurahan/{id} [get]
func (h *Handler) GetKelurahan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kelurahan ID.")
	}
	res, err := h.service.GetKelurahan(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateKelurahan creates a new Kelurahan.
//
//	@Summary		Create kelurahan
//	@Description	Creates a kelurahan or reactivates a soft-deleted kelurahan with the same kecamatanId and kode.
//	@ID				createKelurahan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateKelurahanRequest	true	"Kelurahan creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.KelurahanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kelurahan [post]
func (h *Handler) CreateKelurahan(c fiber.Ctx) error {
	var req domain.CreateKelurahanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateKelurahan(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateKelurahan updates an existing Kelurahan.
//
//	@Summary		Update kelurahan
//	@Description	Updates mutable fields of a kelurahan.
//	@ID				updateKelurahan
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Kelurahan UUID"
//	@Param			body	body		domain.UpdateKelurahanRequest	true	"Kelurahan update payload"
//	@Success		200		{object}	httpx.Envelope[domain.KelurahanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/kelurahan/{id} [patch]
func (h *Handler) UpdateKelurahan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kelurahan ID.")
	}
	var req domain.UpdateKelurahanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateKelurahan(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteKelurahan soft-deletes a Kelurahan.
//
//	@Summary		Delete kelurahan
//	@Description	Soft-deletes a kelurahan.
//	@ID				deleteKelurahan
//	@Tags			Ref
//	@Param			id	path	string	true	"Kelurahan UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/kelurahan/{id} [delete]
func (h *Handler) DeleteKelurahan(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid kelurahan ID.")
	}
	if err := h.service.DeleteKelurahan(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
