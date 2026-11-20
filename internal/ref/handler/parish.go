package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
)

// SearchParishes searches non-deleted parishes.
//
//	@Summary		Search parishes
//	@Description	Searches parishes with optional query, diocese/active/memberOnly filters, and cursor pagination.
//	@ID				searchParishes
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchParishesRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.ParishResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/search [post]
func (h *Handler) SearchParishes(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchParishesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchParishes(c.Context(), p, req, cursorT, cursorID, limit)
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

// GetCurrentParish returns the caller's active Parish.
//
//	@Summary		Get current parish
//	@Description	Returns the parish resolved from the active session or X-Parish-Id header.
//	@ID				getCurrentParish
//	@Tags			Ref
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.ParishResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/current [get]
func (h *Handler) GetCurrentParish(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	res, err := h.service.GetParish(c.Context(), p.ParishID)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// UpdateCurrentParishProfile updates profile fields of the caller's active Parish.
//
//	@Summary		Update current parish profile
//	@Description	Updates mutable profile fields of the active parish (requires ref:parish-admin).
//	@ID				updateCurrentParishProfile
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.UpdateParishProfileRequest	true	"Parish profile payload"
//	@Success		200		{object}	httpx.Envelope[domain.ParishResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/current [patch]
func (h *Handler) UpdateCurrentParishProfile(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.UpdateParishProfileRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateCurrentParishProfile(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// GetParish fetches a single Parish by ID.
//
//	@Summary		Get parish
//	@Description	Returns a single non-deleted parish by ID.
//	@ID				getParish
//	@Tags			Ref
//	@Produce		json
//	@Param			id	path		string	true	"Parish UUID"
//	@Success		200	{object}	httpx.Envelope[domain.ParishResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id} [get]
func (h *Handler) GetParish(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	res, err := h.service.GetParish(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateParish creates a new Parish.
//
//	@Summary		Create parish
//	@Description	Creates a parish or reactivates a soft-deleted parish with the same code.
//	@ID				createParish
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateParishRequest	true	"Parish creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.ParishResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes [post]
func (h *Handler) CreateParish(c fiber.Ctx) error {
	var req domain.CreateParishRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateParish(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateParish updates a Parish.
//
//	@Summary		Update parish
//	@Description	Updates parish fields (super-admin) or profile fields (parish-admin).
//	@ID				updateParish
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Parish UUID"
//	@Param			body	body		domain.UpdateParishRequest	true	"Parish update payload"
//	@Success		200		{object}	httpx.Envelope[domain.ParishResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id} [patch]
func (h *Handler) UpdateParish(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	var req domain.UpdateParishRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateParish(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteParish soft-deletes a Parish if it has no non-deleted master-data children.
//
//	@Summary		Delete parish
//	@Description	Soft-deletes a parish when no zones, wards, lookups, or periods reference it.
//	@ID				deleteParish
//	@Tags			Ref
//	@Param			id	path	string	true	"Parish UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id} [delete]
func (h *Handler) DeleteParish(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	if err := h.service.DeleteParish(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchParishMembers searches members of a Parish.
//
//	@Summary		Search parish members
//	@Description	Searches members of a parish with optional query, role filter, and cursor pagination.
//	@ID				searchParishMembers
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Parish UUID"
//	@Param			body	body		domain.SearchParishMembersRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.ParishMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id}/members/search [post]
func (h *Handler) SearchParishMembers(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parishID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	var req domain.SearchParishMembersRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchParishMembers(c.Context(), p, parishID, req, cursorT, cursorID, limit)
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

// AddParishMember adds an existing or invited user to a Parish.
//
//	@Summary		Add parish member
//	@Description	Adds a user (or creates/invites via email) as a member or admin of a parish.
//	@ID				addParishMember
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Parish UUID"
//	@Param			body	body		domain.AddParishMemberRequest	true	"Member payload"
//	@Success		201		{object}	httpx.Envelope[domain.ParishMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id}/members [post]
func (h *Handler) AddParishMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parishID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	var req domain.AddParishMemberRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.AddParishMember(c.Context(), p, parishID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateParishMember updates a parish member's role.
//
//	@Summary		Update parish member
//	@Description	Updates a parish member role and syncs Casbin parish-admin policy.
//	@ID				updateParishMember
//	@Tags			Ref
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Parish UUID"
//	@Param			userId	path		string								true	"User UUID"
//	@Param			body	body		domain.UpdateParishMemberRequest	true	"Role update payload"
//	@Success		200		{object}	httpx.Envelope[domain.ParishMemberResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id}/members/{userId} [patch]
func (h *Handler) UpdateParishMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parishID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	var req domain.UpdateParishMemberRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateParishMember(c.Context(), p, parishID, userID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// RemoveParishMember removes a member from a Parish.
//
//	@Summary		Remove parish member
//	@Description	Removes a member from a parish and revokes Casbin parish-admin policy.
//	@ID				removeParishMember
//	@Tags			Ref
//	@Param			id		path	string	true	"Parish UUID"
//	@Param			userId	path	string	true	"User UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/ref/parishes/{id}/members/{userId} [delete]
func (h *Handler) RemoveParishMember(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	parishID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid parish ID.")
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	if err := h.service.RemoveParishMember(c.Context(), p, parishID, userID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
