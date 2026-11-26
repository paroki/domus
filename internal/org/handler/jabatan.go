package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchOrgJabatan searches position nodes of an Organization.
//
//	@Summary		Search organization positions
//	@Description	Searches non-deleted position nodes (org_jabatan) for an organization with optional filters and cursor pagination, ordered by urutan ASC, createdAt ASC, id ASC.
//	@ID				searchOrgJabatan
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.SearchOrgJabatanRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.OrgJabatanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/jabatan/search [post]
func (h *Handler) SearchOrgJabatan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.SearchOrgJabatanRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorUrutan, cursorT, cursorID, limit, err := parseJabatanCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchOrgJabatan(c.Context(), p, orgID, req, cursorUrutan, cursorT, cursorID, limit)
	if err != nil {
		return err
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = httpx.EncodeCursor(last.CreatedAt, fmt.Sprintf("%d:%s", last.Urutan, last.ID.String()))
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// GetOrgJabatan fetches a single OrgJabatan in an Organization.
//
//	@Summary		Get organization position
//	@Description	Returns a single non-deleted position node for an organization.
//	@ID				getOrgJabatan
//	@Tags			Org
//	@Produce		json
//	@Param			id			path		string	true	"Organization UUID"
//	@Param			jabatanId	path		string	true	"Position UUID"
//	@Success		200			{object}	httpx.Envelope[domain.OrgJabatanResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/jabatan/{jabatanId} [get]
func (h *Handler) GetOrgJabatan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	jabatanID, err := uuid.Parse(c.Params("jabatanId"))
	if err != nil {
		return httpx.BadRequest("Invalid position ID.")
	}
	res, err := h.service.GetOrgJabatan(c.Context(), p, orgID, jabatanID)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateOrgJabatan creates a new position node in an Organization period.
//
//	@Summary		Create organization position
//	@Description	Creates an organization position node with hierarchy validation and automatic userId/umatId/nama synchronization.
//	@ID				createOrgJabatan
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.CreateOrgJabatanRequest	true	"Position creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.OrgJabatanResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/jabatan [post]
func (h *Handler) CreateOrgJabatan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.CreateOrgJabatanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateOrgJabatan(c.Context(), p, orgID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateOrgJabatan updates an existing position node of an Organization.
//
//	@Summary		Update organization position
//	@Description	Updates mutable fields of an organization position node with cycle detection and automatic userId/umatId/nama synchronization.
//	@ID				updateOrgJabatan
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string							true	"Organization UUID"
//	@Param			jabatanId	path		string							true	"Position UUID"
//	@Param			body		body		domain.UpdateOrgJabatanRequest	true	"Position update payload"
//	@Success		200			{object}	httpx.Envelope[domain.OrgJabatanResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/jabatan/{jabatanId} [patch]
func (h *Handler) UpdateOrgJabatan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	jabatanID, err := uuid.Parse(c.Params("jabatanId"))
	if err != nil {
		return httpx.BadRequest("Invalid position ID.")
	}
	var req domain.UpdateOrgJabatanRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateOrgJabatan(c.Context(), p, orgID, jabatanID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteOrgJabatan soft-deletes an OrgJabatan in an Organization.
//
//	@Summary		Delete organization position
//	@Description	Soft-deletes an organization position node if it has no active child positions.
//	@ID				deleteOrgJabatan
//	@Tags			Org
//	@Param			id			path	string	true	"Organization UUID"
//	@Param			jabatanId	path	string	true	"Position UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/jabatan/{jabatanId} [delete]
func (h *Handler) DeleteOrgJabatan(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	jabatanID, err := uuid.Parse(c.Params("jabatanId"))
	if err != nil {
		return httpx.BadRequest("Invalid position ID.")
	}
	if err := h.service.DeleteOrgJabatan(c.Context(), p, orgID, jabatanID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

func parseJabatanCursorAndLimit(rawCursor string, rawLimit int) (*int, *time.Time, *uuid.UUID, int, error) {
	limit := rawLimit
	if limit <= 0 {
		limit = httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		limit = httpx.MaxLimit
	}
	if rawCursor == "" {
		return nil, nil, nil, limit, nil
	}
	cur, err := httpx.DecodeCursor(rawCursor)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	parts := strings.SplitN(cur.ID, ":", 2)
	if len(parts) != 2 {
		return nil, nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	urutan, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	t := cur.CreatedAt
	return &urutan, &t, &id, limit, nil
}
