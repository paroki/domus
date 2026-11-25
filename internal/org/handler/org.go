package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchOrganizations searches non-deleted organizations in the active paroki.
//
//	@Summary		Search organizations
//	@Description	Searches organizations within the active paroki with optional filters and cursor pagination.
//	@ID				searchOrganizations
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchOrganizationsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.OrganizationResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/orgs/search [post]
func (h *Handler) SearchOrganizations(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchOrganizationsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchOrganizations(c.Context(), p, req, cursorT, cursorID, limit)
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

// GetOrganization fetches a single Organization in the active paroki by ID.
//
//	@Summary		Get organization
//	@Description	Returns a single non-deleted organization in the active paroki.
//	@ID				getOrganization
//	@Tags			Org
//	@Produce		json
//	@Param			id	path		string	true	"Organization UUID"
//	@Success		200	{object}	httpx.Envelope[domain.OrganizationResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/orgs/{id} [get]
func (h *Handler) GetOrganization(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	res, err := h.service.GetOrganizationForCaller(c.Context(), p, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateOrganization creates a new Organization in the active paroki.
//
//	@Summary		Create organization
//	@Description	Creates an organization (or reactivates a soft-deleted one) in the active paroki (requires ref:paroki-admin).
//	@ID				createOrganization
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateOrganizationRequest	true	"Organization creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.OrganizationResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs [post]
func (h *Handler) CreateOrganization(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateOrganizationRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateOrganization(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateOrganization updates an existing Organization in the active paroki.
//
//	@Summary		Update organization
//	@Description	Updates an organization in the active paroki (requires org:admin or ref:paroki-admin).
//	@ID				updateOrganization
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Organization UUID"
//	@Param			body	body		domain.UpdateOrganizationRequest	true	"Organization update payload"
//	@Success		200		{object}	httpx.Envelope[domain.OrganizationResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id} [patch]
func (h *Handler) UpdateOrganization(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.UpdateOrganizationRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateOrganization(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteOrganization soft-deletes an Organization in the active paroki.
//
//	@Summary		Delete organization
//	@Description	Soft-deletes an organization if it has no active child organizations (requires ref:paroki-admin).
//	@ID				deleteOrganization
//	@Tags			Org
//	@Produce		json
//	@Param			id	path	string	true	"Organization UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/orgs/{id} [delete]
func (h *Handler) DeleteOrganization(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	if err := h.service.DeleteOrganization(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
