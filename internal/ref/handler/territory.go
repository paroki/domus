package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchZones searches zones/stations in the active parish.
//
//	@Summary		Search zones
//	@Description	Searches zones/stations in the active parish with optional query, type/active filters, and cursor pagination.
//	@ID				searchZones
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchZonesRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.ZoneResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/zones/search [post]
func (h *Handler) SearchZones(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchZonesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchZones(c.Context(), p.ParishID, req, cursorT, cursorID, limit)
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

// GetZone fetches a single Zone in the active parish.
//
//	@Summary		Get zone
//	@Description	Returns a single non-deleted zone in the active parish.
//	@ID				getZone
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Zone UUID"
//	@Success		200	{object}	httpx.Envelope[domain.ZoneResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/zones/{id} [get]
func (h *Handler) GetZone(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid zone ID.")
	}
	res, err := h.service.GetZone(c.Context(), p.ParishID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateZone creates a Zone or Station in the active parish.
//
//	@Summary		Create zone
//	@Description	Creates a zone/station or reactivates a soft-deleted match in the active parish.
//	@ID				createZone
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateZoneRequest	true	"Zone creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.ZoneResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/zones [post]
func (h *Handler) CreateZone(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateZoneRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateZone(c.Context(), p.ParishID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateZone updates an existing Zone or Station in the active parish.
//
//	@Summary		Update zone
//	@Description	Updates mutable fields of a zone/station in the active parish.
//	@ID				updateZone
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Zone UUID"
//	@Param			body	body		domain.UpdateZoneRequest	true	"Zone update payload"
//	@Success		200		{object}	httpx.Envelope[domain.ZoneResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/zones/{id} [patch]
func (h *Handler) UpdateZone(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid zone ID.")
	}
	var req domain.UpdateZoneRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateZone(c.Context(), p.ParishID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteZone soft-deletes a Zone in the active parish if no Wards reference it.
//
//	@Summary		Delete zone
//	@Description	Soft-deletes a zone when no non-deleted wards reference it.
//	@ID				deleteZone
//	@Tags			Ref
//	@Param			id	path	string	true	"Zone UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/zones/{id} [delete]
func (h *Handler) DeleteZone(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid zone ID.")
	}
	if err := h.service.DeleteZone(c.Context(), p.ParishID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchWards searches wards in the active parish.
//
//	@Summary		Search wards
//	@Description	Searches wards in the active parish with optional query, zone/active filters, and cursor pagination.
//	@ID				searchWards
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchWardsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.WardResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/wards/search [post]
func (h *Handler) SearchWards(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchWardsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchWards(c.Context(), p.ParishID, req, cursorT, cursorID, limit)
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

// GetWard fetches a single Ward in the active parish.
//
//	@Summary		Get ward
//	@Description	Returns a single non-deleted ward in the active parish.
//	@ID				getWard
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Ward UUID"
//	@Success		200	{object}	httpx.Envelope[domain.WardResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/wards/{id} [get]
func (h *Handler) GetWard(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid ward ID.")
	}
	res, err := h.service.GetWard(c.Context(), p.ParishID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateWard creates a Ward in the active parish.
//
//	@Summary		Create ward
//	@Description	Creates a ward or reactivates a soft-deleted match in the active parish.
//	@ID				createWard
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateWardRequest	true	"Ward creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.WardResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/wards [post]
func (h *Handler) CreateWard(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateWardRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateWard(c.Context(), p.ParishID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateWard updates an existing Ward in the active parish.
//
//	@Summary		Update ward
//	@Description	Updates mutable fields of a ward in the active parish.
//	@ID				updateWard
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Ward UUID"
//	@Param			body	body		domain.UpdateWardRequest	true	"Ward update payload"
//	@Success		200		{object}	httpx.Envelope[domain.WardResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/wards/{id} [patch]
func (h *Handler) UpdateWard(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid ward ID.")
	}
	var req domain.UpdateWardRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateWard(c.Context(), p.ParishID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteWard soft-deletes a Ward in the active parish.
//
//	@Summary		Delete ward
//	@Description	Soft-deletes a ward in the active parish.
//	@ID				deleteWard
//	@Tags			Ref
//	@Param			id	path	string	true	"Ward UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/wards/{id} [delete]
func (h *Handler) DeleteWard(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid ward ID.")
	}
	if err := h.service.DeleteWard(c.Context(), p.ParishID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
