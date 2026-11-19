package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user/domain"
)

// GetMe returns the authenticated user's profile.
//
//	@Summary		Get current user profile
//	@Description	Returns the authenticated user's profile, active parish, and global roles.
//	@ID				getMe
//	@Tags			Me
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.UserResponse]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Router			/me [get]
func (h *Handler) GetMe(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	res, err := h.service.GetMe(c.Context(), p)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// UpdateMe updates the authenticated user's profile fields.
//
//	@Summary		Update current user profile
//	@Description	Updates name, phone (normalized to E.164), avatarUrl, or defaultParishId.
//	@ID				updateMe
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.UpdateMeRequest	true	"Profile update payload"
//	@Success		200		{object}	httpx.Envelope[domain.UserResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/me [patch]
func (h *Handler) UpdateMe(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.UpdateMeRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateMe(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// SwitchSessionParish updates the active parish for the current browser session.
//
//	@Summary		Switch active session parish
//	@Description	Updates activeParishId on the current session and initializes user defaultParishId if unset.
//	@ID				switchMeSessionParish
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SwitchSessionParishRequest	true	"Active parish payload"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		422	{object}	httpx.Envelope[any]
//	@Router			/me/session [patch]
func (h *Handler) SwitchSessionParish(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SwitchSessionParishRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	if err := h.service.SwitchSessionParish(c.Context(), p, req.ActiveParishID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchMySessions lists active sessions for the authenticated user.
//
//	@Summary		Search my active sessions
//	@Description	Lists active non-expired sessions for the current user with cursor pagination.
//	@ID				searchMeSessions
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchSessionsRequest	false	"Pagination parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.UserSessionResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Router			/me/sessions/search [post]
func (h *Handler) SearchMySessions(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchSessionsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchMySessions(c.Context(), p, cursorT, cursorID, limit)
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

// RevokeMySession revokes a single session owned by the authenticated user.
//
//	@Summary		Revoke a user session
//	@Description	Deletes the specified session owned by the current user and clears cookie if current.
//	@ID				revokeMeSession
//	@Tags			Me
//	@Param			id	path	string	true	"Session UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/me/sessions/{id} [delete]
func (h *Handler) RevokeMySession(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid session ID.")
	}
	isCurrent, err := h.service.RevokeMySession(c.Context(), p, sessionID)
	if err != nil {
		return err
	}
	if isCurrent {
		h.clearCookie(c, h.cookieName)
	}
	return httpx.NoContent(c)
}

// SearchMyAccounts lists linked identity accounts for the authenticated user.
//
//	@Summary		Search my linked accounts
//	@Description	Lists linked OAuth2, email, and channel accounts for the current user.
//	@ID				searchMeAccounts
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchAccountsRequest	false	"Pagination parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.UserAccountResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Router			/me/accounts/search [post]
func (h *Handler) SearchMyAccounts(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchAccountsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchMyAccounts(c.Context(), p, cursorT, cursorID, limit)
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

// UnlinkMyAccount removes a linked provider account from the authenticated user.
//
//	@Summary		Unlink a connected account
//	@Description	Removes a linked account unless it is the user's last remaining authentication method.
//	@ID				unlinkMeAccount
//	@Tags			Me
//	@Param			id	path	string	true	"Account UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/me/accounts/{id} [delete]
func (h *Handler) UnlinkMyAccount(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	accountID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid account ID.")
	}
	if err := h.service.UnlinkMyAccount(c.Context(), p, accountID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
