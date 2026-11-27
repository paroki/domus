package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user/domain"
)

// GetVAPIDPublicKey returns the application's VAPID public key for Web Push registration.
//
//	@Summary		Get VAPID public key
//	@Description	Returns the base64url-encoded ECDSA P-256 VAPID public key for PushManager.subscribe.
//	@ID				getPushVAPIDPublicKey
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.VAPIDPublicKeyResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Router			/auth/push/vapid-public-key [get]
func (h *Handler) GetVAPIDPublicKey(c fiber.Ctx) error {
	res, err := h.service.GetVAPIDPublicKey()
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// SubscribePush registers or updates a browser Web Push subscription for the authenticated user.
//
//	@Summary		Subscribe to Web Push notifications
//	@Description	Validates and upserts a browser PushSubscription for the authenticated user and current session.
//	@ID				subscribeMePush
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body	domain.SubscribePushRequest	true	"Web Push subscription payload"
//	@Success		204
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		422	{object}	httpx.Envelope[any]
//	@Router			/me/push/subscribe [post]
func (h *Handler) SubscribePush(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SubscribePushRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	if err := h.service.SubscribePush(c.Context(), p, c.Get(fiber.HeaderUserAgent), req); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// UnsubscribePush removes a browser Web Push subscription for the authenticated user.
//
//	@Summary		Unsubscribe from Web Push notifications
//	@Description	Deletes the matching Web Push subscription owned by the authenticated user.
//	@ID				unsubscribeMePush
//	@Tags			Me
//	@Accept			json
//	@Produce		json
//	@Param			body	body	domain.UnsubscribePushRequest	true	"Web Push endpoint to remove"
//	@Success		204
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		422	{object}	httpx.Envelope[any]
//	@Router			/me/push/unsubscribe [post]
func (h *Handler) UnsubscribePush(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.UnsubscribePushRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	if err := h.service.UnsubscribePush(c.Context(), p, req); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// TestPushNotification dispatches a test notification across the caller's enabled channels.
//
//	@Summary		Send test notification
//	@Description	Dispatches a test notification to the authenticated user's enabled push, WhatsApp, or Telegram channels.
//	@ID				testMePush
//	@Tags			Me
//	@Produce		json
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Router			/me/push/test [post]
func (h *Handler) TestPushNotification(c fiber.Ctx) error {
	p, ok := authx.FromContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	if err := h.service.TestPushNotification(c.Context(), p); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
