package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchOrgMembers searches members of an Organization.
//
//	@Summary		Search organization members
//	@Description	Searches members of an organization (requires org:admin or org:member).
//	@ID				searchOrgMembers
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			body	body		domain.SearchOrgMembersRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.OrgMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/members/search [post]
func (h *Handler) SearchOrgMembers(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.SearchOrgMembersRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchOrgMembers(c.Context(), p, orgID, req, cursorT, cursorID, limit)
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

// AddOrgMember adds or appoints a member to an Organization.
//
//	@Summary		Add organization member
//	@Description	Adds a member to an organization with optional user onboarding (requires org:admin, or ref:parish-admin appointing admin).
//	@ID				addOrgMember
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Organization UUID"
//	@Param			body	body		domain.AddOrgMemberRequest	true	"Member payload"
//	@Success		201		{object}	httpx.Envelope[domain.OrgMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/members [post]
func (h *Handler) AddOrgMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	var req domain.AddOrgMemberRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.AddOrgMember(c.Context(), p, orgID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateOrgMember updates an organization member's role, position, term, or status.
//
//	@Summary		Update organization member
//	@Description	Updates an organization member and syncs Casbin roles while guarding the last active admin.
//	@ID				updateOrgMember
//	@Tags			Org
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Organization UUID"
//	@Param			userId	path		string							true	"User UUID"
//	@Param			body	body		domain.UpdateOrgMemberRequest	true	"Member update payload"
//	@Success		200		{object}	httpx.Envelope[domain.OrgMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/members/{userId} [patch]
func (h *Handler) UpdateOrgMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	var req domain.UpdateOrgMemberRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateOrgMember(c.Context(), p, orgID, userID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// RemoveOrgMember removes a member from an Organization.
//
//	@Summary		Remove organization member
//	@Description	Removes a member from an organization (org:admin or self-leave) while guarding the last active admin.
//	@ID				removeOrgMember
//	@Tags			Org
//	@Produce		json
//	@Param			id		path	string	true	"Organization UUID"
//	@Param			userId	path	string	true	"User UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/orgs/{id}/members/{userId} [delete]
func (h *Handler) RemoveOrgMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	orgID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid organization ID.")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	if err := h.service.RemoveOrgMember(c.Context(), p, orgID, userID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
