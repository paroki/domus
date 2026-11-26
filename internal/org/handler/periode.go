package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchOrgPeriode searches periods of an Organization.
//
//	@Summary		Search organization periods
//	@Description	Searches non-deleted periods for an organization with optional filters and cursor pagination, ordered by tanggalMulai DESC, id DESC.
//	@ID				searchOrgPeriode
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.SearchOrgPeriodeRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.OrgPeriodeResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/periode/search [post]
func (h *Handler) SearchOrgPeriode(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.SearchOrgPeriodeRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	var cursorTanggalMulai *string
	if cursorT != nil {
		s := cursorT.UTC().Format("2006-01-02")
		cursorTanggalMulai = &s
	}
	items, hasMore, err := h.service.SearchOrgPeriode(c.Context(), p, orgID, req, cursorTanggalMulai, cursorID, limit)
	if err != nil {
		return err
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		if t, parseErr := time.Parse("2006-01-02", last.TanggalMulai); parseErr == nil {
			nextCursor = httpx.EncodeCursor(t, last.ID.String())
		}
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// GetOrgPeriode fetches a single OrgPeriode in an Organization.
//
//	@Summary		Get organization period
//	@Description	Returns a single non-deleted period for an organization.
//	@ID				getOrgPeriode
//	@Tags			Org
//	@Produce		json
//	@Param			id			path		string	true	"Organization UUID"
//	@Param			periodeId	path		string	true	"Period UUID"
//	@Success		200			{object}	httpx.Envelope[domain.OrgPeriodeResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/periode/{periodeId} [get]
func (h *Handler) GetOrgPeriode(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	periodeID, err := uuid.Parse(c.Params("periodeId"))
	if err != nil {
		return httpx.BadRequest("Invalid period ID.")
	}
	res, err := h.service.GetOrgPeriode(c.Context(), p, orgID, periodeID)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateOrgPeriode creates a new period for an Organization.
//
//	@Summary		Create organization period
//	@Description	Creates an organization period with date overlap validation, soft-deleted kode reactivation, and atomic isCurrent rotation.
//	@ID				createOrgPeriode
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.CreateOrgPeriodeRequest	true	"Period creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.OrgPeriodeResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/periode [post]
func (h *Handler) CreateOrgPeriode(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.CreateOrgPeriodeRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateOrgPeriode(c.Context(), p, orgID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateOrgPeriode updates an existing period of an Organization.
//
//	@Summary		Update organization period
//	@Description	Updates mutable fields of an organization period with overlap validation and atomic isCurrent rotation.
//	@ID				updateOrgPeriode
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string							true	"Organization UUID"
//	@Param			periodeId	path		string							true	"Period UUID"
//	@Param			body		body		domain.UpdateOrgPeriodeRequest	true	"Period update payload"
//	@Success		200			{object}	httpx.Envelope[domain.OrgPeriodeResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		409			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/periode/{periodeId} [patch]
func (h *Handler) UpdateOrgPeriode(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	periodeID, err := uuid.Parse(c.Params("periodeId"))
	if err != nil {
		return httpx.BadRequest("Invalid period ID.")
	}
	var req domain.UpdateOrgPeriodeRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateOrgPeriode(c.Context(), p, orgID, periodeID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteOrgPeriode soft-deletes an OrgPeriode in an Organization.
//
//	@Summary		Delete organization period
//	@Description	Soft-deletes an organization period and clears isCurrent.
//	@ID				deleteOrgPeriode
//	@Tags			Org
//	@Param			id			path	string	true	"Organization UUID"
//	@Param			periodeId	path	string	true	"Period UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/periode/{periodeId} [delete]
func (h *Handler) DeleteOrgPeriode(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	periodeID, err := uuid.Parse(c.Params("periodeId"))
	if err != nil {
		return httpx.BadRequest("Invalid period ID.")
	}
	if err := h.service.DeleteOrgPeriode(c.Context(), p, orgID, periodeID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
