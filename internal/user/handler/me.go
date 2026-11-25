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
//	@Description	Returns the authenticated user's profile, active paroki, and global roles.
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
//	@Description	Updates name, phone (normalized to E.164), gender, birthDate, avatarUrl, defaultParokiId, or onboarded status.
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

// SwitchSessionParoki updates the active paroki for the current browser session.
//
//	@Summary		Switch active session paroki
//	@Description	Updates activeParokiId on the current session and initializes user defaultParokiId if unset.
//	@ID				switchMeSessionParoki
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SwitchSessionParokiRequest	true	"Active paroki payload"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		422	{object}	httpx.Envelope[any]
//	@Router			/me/session [patch]
func (h *Handler) SwitchSessionParoki(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SwitchSessionParokiRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	if err := h.service.SwitchSessionParoki(c.Context(), p, req.ActiveParokiID); err != nil {
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

// RequestLinkAccountOTP dispatches a verification OTP to link a WhatsApp number.
//
//	@Summary		Request account linking OTP
//	@Description	Sends a 6-digit verification OTP to link a WhatsApp number to the current user.
//	@ID				requestMeAccountOTP
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.LinkAccountOTPRequest	true	"Account link OTP request"
//	@Success		200		{object}	httpx.Envelope[domain.RequestOTPResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Failure		429		{object}	httpx.Envelope[any]
//	@Router			/me/accounts/otp [post]
func (h *Handler) RequestLinkAccountOTP(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.LinkAccountOTPRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.RequestLinkAccountOTP(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// VerifyLinkAccount verifies an OTP and links a WhatsApp account to the current user.
//
//	@Summary		Verify account linking OTP
//	@Description	Verifies a 6-digit OTP and links the verified WhatsApp account to the current user.
//	@ID				verifyMeAccount
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.VerifyLinkAccountRequest	true	"Account link verification payload"
//	@Success		200		{object}	httpx.Envelope[domain.UserAccountResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/me/accounts/verify [post]
func (h *Handler) VerifyLinkAccount(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.VerifyLinkAccountRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.VerifyLinkAccount(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateTelegramLinkToken generates a single-use deep link to connect Telegram.
//
//	@Summary		Create Telegram link token
//	@Description	Generates a single-use t.me deep-link token to connect Telegram to the current user.
//	@ID				createMeTelegramLinkToken
//	@Tags			Me
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.TelegramLinkTokenResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Router			/me/accounts/telegram/token [post]
func (h *Handler) CreateTelegramLinkToken(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	res, err := h.service.CreateTelegramLinkToken(c.Context(), p)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}
