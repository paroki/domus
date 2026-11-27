package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user/domain"
)

// SearchUsers searches the global user directory.
//
//	@Summary		Search users
//	@Description	Searches non-deleted users with optional query, status filter, and cursor pagination.
//	@ID				searchUsers
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchUsersRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.UserResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/users/search [post]
func (h *Handler) SearchUsers(c fiber.Ctx) error {
	var req domain.SearchUsersRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	var status string
	var cantusAdmin *bool
	if req.Filter != nil {
		status = req.Filter.Status
		cantusAdmin = req.Filter.CantusAdmin
	}
	items, hasMore, err := h.service.SearchUsers(c.Context(), req.Q, status, cantusAdmin, cursorT, cursorID, limit)
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

// AdminCreateUser creates a new user and optionally appoints them Super Admin.
//
//	@Summary		Create user
//	@Description	Creates an active user, links an email account, and assigns global roles.
//	@ID				createUser
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.AdminCreateUserRequest	true	"User creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.UserResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/users [post]
func (h *Handler) AdminCreateUser(c fiber.Ctx) error {
	var req domain.AdminCreateUserRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.AdminCreateUser(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// AdminGetUser fetches a single user by ID.
//
//	@Summary		Get user by ID
//	@Description	Returns a single non-deleted user and their global roles.
//	@ID				getUser
//	@Tags			Users
//	@Produce		json
//	@Param			id	path		string	true	"User UUID"
//	@Success		200	{object}	httpx.Envelope[domain.UserResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/users/{id} [get]
func (h *Handler) AdminGetUser(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	res, err := h.service.AdminGetUser(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// AdminUpdateUser updates a user's profile, status, or Super Admin role.
//
//	@Summary		Update user
//	@Description	Updates user profile, status (revoking sessions on disable), or Super Admin role.
//	@ID				updateUser
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"User UUID"
//	@Param			body	body		domain.AdminUpdateUserRequest	true	"User update payload"
//	@Success		200		{object}	httpx.Envelope[domain.UserResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/users/{id} [patch]
func (h *Handler) AdminUpdateUser(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	var req domain.AdminUpdateUserRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.AdminUpdateUser(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// AdminDeleteUser soft-deletes a user and revokes all active sessions.
//
//	@Summary		Delete user
//	@Description	Soft-deletes a user and revokes all active sessions while guarding the last Super Admin.
//	@ID				deleteUser
//	@Tags			Users
//	@Param			id	path	string	true	"User UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/users/{id} [delete]
func (h *Handler) AdminDeleteUser(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid user ID.")
	}
	if err := h.service.AdminDeleteUser(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
