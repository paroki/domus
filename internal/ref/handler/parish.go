package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchParoki searches non-deleted paroki.
//
//	@Summary		Search paroki
//	@Description	Searches paroki with optional query, keuskupan/active/memberOnly filters, and cursor pagination.
//	@ID				searchParoki
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchParokiRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.ParokiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/search [post]
func (h *Handler) SearchParoki(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchParokiRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchParoki(c.Context(), p, req, cursorT, cursorID, limit)
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

// GetCurrentParoki returns the caller's active Paroki.
//
//	@Summary		Get current paroki
//	@Description	Returns the paroki resolved from the active session or X-Paroki-Id header.
//	@ID				getCurrentParoki
//	@Tags			Ref
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.ParokiResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/current [get]
func (h *Handler) GetCurrentParoki(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	res, err := h.service.GetParoki(c.Context(), p.ParokiID)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// UpdateCurrentParokiProfile updates profile fields of the caller's active Paroki.
//
//	@Summary		Update current paroki profile
//	@Description	Updates mutable profile fields of the active paroki (requires ref:paroki-admin).
//	@ID				updateCurrentParokiProfile
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.UpdateParokiProfileRequest	true	"Paroki profile payload"
//	@Success		200		{object}	httpx.Envelope[domain.ParokiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/current [patch]
func (h *Handler) UpdateCurrentParokiProfile(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.UpdateParokiProfileRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateCurrentParokiProfile(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// GetParoki fetches a single Paroki by ID.
//
//	@Summary		Get paroki
//	@Description	Returns a single non-deleted paroki by ID.
//	@ID				getParoki
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Paroki UUID"
//	@Success		200	{object}	httpx.Envelope[domain.ParokiResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id} [get]
func (h *Handler) GetParoki(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	res, err := h.service.GetParoki(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateParoki creates a new Paroki.
//
//	@Summary		Create paroki
//	@Description	Creates a paroki or reactivates a soft-deleted paroki with the same kode.
//	@ID				createParoki
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateParokiRequest	true	"Paroki creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.ParokiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki [post]
func (h *Handler) CreateParoki(c fiber.Ctx) error {
	var req domain.CreateParokiRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateParoki(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateParoki updates a Paroki.
//
//	@Summary		Update paroki
//	@Description	Updates paroki fields (super-admin) or profile fields (paroki-admin).
//	@ID				updateParoki
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Paroki UUID"
//	@Param			body	body		domain.UpdateParokiRequest	true	"Paroki update payload"
//	@Success		200		{object}	httpx.Envelope[domain.ParokiResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id} [patch]
func (h *Handler) UpdateParoki(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	var req domain.UpdateParokiRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateParoki(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteParoki soft-deletes a Paroki if it has no non-deleted master-data children.
//
//	@Summary		Delete paroki
//	@Description	Soft-deletes a paroki when no wilayah, lingkungan, lookups, or periods reference it.
//	@ID				deleteParoki
//	@Tags			Ref
//	@Param			id	path	string	true	"Paroki UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id} [delete]
func (h *Handler) DeleteParoki(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	if err := h.service.DeleteParoki(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchParokiMembers searches members of a Paroki.
//
//	@Summary		Search paroki members
//	@Description	Searches members of a paroki with optional query, role filter, and cursor pagination.
//	@ID				searchParokiMembers
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Paroki UUID"
//	@Param			body	body		domain.SearchParokiMembersRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.ParokiMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id}/members/search [post]
func (h *Handler) SearchParokiMembers(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parokiID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	var req domain.SearchParokiMembersRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchParokiMembers(c.Context(), p, parokiID, req, cursorT, cursorID, limit)
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

// AddParokiMember adds an existing or invited user to a Paroki.
//
//	@Summary		Add paroki member
//	@Description	Adds a user (or creates/invites via email) as a member or admin of a paroki.
//	@ID				addParokiMember
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Paroki UUID"
//	@Param			body	body		domain.AddParokiMemberRequest	true	"Member payload"
//	@Success		201		{object}	httpx.Envelope[domain.ParokiMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id}/members [post]
func (h *Handler) AddParokiMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parokiID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	var req domain.AddParokiMemberRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.AddParokiMember(c.Context(), p, parokiID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateParokiMember updates a paroki member's role.
//
//	@Summary		Update paroki member
//	@Description	Updates a paroki member role and syncs Casbin paroki-admin policy.
//	@ID				updateParokiMember
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Paroki UUID"
//	@Param			userId	path		string								true	"User UUID"
//	@Param			body	body		domain.UpdateParokiMemberRequest	true	"Role update payload"
//	@Success		200		{object}	httpx.Envelope[domain.ParokiMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id}/members/{userId} [patch]
func (h *Handler) UpdateParokiMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parokiID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	var req domain.UpdateParokiMemberRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateParokiMember(c.Context(), p, parokiID, userID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// RemoveParokiMember removes a member from a Paroki.
//
//	@Summary		Remove paroki member
//	@Description	Removes a member from a paroki and revokes Casbin paroki-admin policy.
//	@ID				removeParokiMember
//	@Tags			Ref
//	@Param			id		path	string	true	"Paroki UUID"
//	@Param			userId	path	string	true	"User UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/paroki/{id}/members/{userId} [delete]
func (h *Handler) RemoveParokiMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parokiID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid paroki ID.")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	if err := h.service.RemoveParokiMember(c.Context(), p, parokiID, userID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
