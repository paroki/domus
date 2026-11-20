package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchOrgInvites searches invite links for an Organization.
//
//	@Summary		Search organization invites
//	@Description	Searches invite links for an organization (requires org:admin).
//	@ID				searchOrgInvites
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.SearchOrgInvitesRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.OrgInviteResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/invites/search [post]
func (h *Handler) SearchOrgInvites(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.SearchOrgInvitesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchOrgInvites(c.Context(), p, orgID, req, cursorT, cursorID, limit)
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

// CreateOrgInvite creates a time-limited multi-use invite link for an Organization.
//
//	@Summary		Create organization invite
//	@Description	Creates an invite link for an organization (requires org:admin).
//	@ID				createOrgInvite
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.CreateOrgInviteRequest	true	"Invite creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.OrgInviteResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/invites [post]
func (h *Handler) CreateOrgInvite(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.CreateOrgInviteRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateOrgInvite(c.Context(), p, orgID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// RevokeOrgInvite revokes an organization invite link.
//
//	@Summary		Revoke organization invite
//	@Description	Idempotently revokes an invite link for an organization (requires org:admin).
//	@ID				revokeOrgInvite
//	@Tags			Org
//	@Produce		json
//	@Param			id			path	string	true	"Organization UUID"
//	@Param			inviteId	path	string	true	"Invite UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/invites/{inviteId} [delete]
func (h *Handler) RevokeOrgInvite(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	inviteID, err := uuid.Parse(c.Params("inviteId"))
	if err != nil {
		return httpx.BadRequest("Invalid invite ID.")
	}
	if err := h.service.RevokeOrgInvite(c.Context(), p, orgID, inviteID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// PreviewOrgInvite returns public preview information for an organization invite code.
//
//	@Summary		Preview organization invite
//	@Description	Returns public metadata and validity status for an organization invite code.
//	@ID				previewOrgInvite
//	@Tags			Org
//	@Produce		json
//	@Param			code	path		string	true	"Invite code"
//	@Success		200		{object}	httpx.Envelope[domain.OrgInvitePreviewResponse]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/orgs/invites/{code} [get]
func (h *Handler) PreviewOrgInvite(c fiber.Ctx) error {
	res, err := h.service.PreviewOrgInvite(c.Context(), c.Params("code"))
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// RedeemOrgInvite redeems an organization invite code for the authenticated user.
//
//	@Summary		Redeem organization invite
//	@Description	Idempotently redeems an invite code, ensures parish/org membership, syncs Casbin, and switches active parish.
//	@ID				redeemOrgInvite
//	@Tags			Org
//	@Produce		json
//	@Param			code	path		string	true	"Invite code"
//	@Success		200		{object}	httpx.Envelope[domain.RedeemOrgInviteResponse]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		410		{object}	httpx.Envelope[any]
//	@Router			/orgs/invites/{code}/redeem [post]
func (h *Handler) RedeemOrgInvite(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	res, err := h.service.RedeemInvite(c.Context(), p, c.Params("code"))
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}
