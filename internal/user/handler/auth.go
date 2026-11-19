package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user/domain"
)

const (
	oauthStateCookieName    = "domus_oauth_state"
	oauthRedirectCookieName = "domus_oauth_redirect"
	oauthCookieMaxAge       = 600 // 10 minutes
)

// GetProviders lists enabled OAuth2 providers.
//
//	@Summary		List enabled auth providers
//	@Description	Returns enabled OAuth2 provider names in deterministic order.
//	@ID				getAuthProviders
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.AuthProvidersResponse]
//	@Router			/auth/providers [get]
func (h *Handler) GetProviders(c fiber.Ctx) error {
	return httpx.OK(c, domain.AuthProvidersResponse{
		Providers: h.service.EnabledProviders(),
	})
}

// RequestOTP dispatches a 6-digit passwordless login code via email.
//
//	@Summary		Request email OTP
//	@Description	Sends a 6-digit one-time passcode to an existing active user's email.
//	@ID				requestAuthOTP
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.RequestOTPRequest	true	"Email address"
//	@Success		200		{object}	httpx.Envelope[domain.RequestOTPResponse]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Failure		429		{object}	httpx.Envelope[any]
//	@Router			/auth/otp [post]
func (h *Handler) RequestOTP(c fiber.Ctx) error {
	var req domain.RequestOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	res, err := h.service.RequestOTP(c.Context(), req.Email)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// VerifyOTP validates a 6-digit email OTP and sets the session cookie.
//
//	@Summary		Verify email OTP
//	@Description	Verifies a 6-digit OTP code and establishes an authenticated session cookie.
//	@ID				verifyAuthOTP
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.VerifyOTPRequest	true	"OTP verification payload"
//	@Success		200		{object}	httpx.Envelope[domain.AuthSessionResponse]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/auth/otp/verify [post]
func (h *Handler) VerifyOTP(c fiber.Ctx) error {
	var req domain.VerifyOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}
	rawToken, res, err := h.service.VerifyOTP(c.Context(), req.Email, req.Code, c.Get(fiber.HeaderUserAgent), c.IP())
	if err != nil {
		return err
	}
	h.setSessionCookie(c, rawToken)
	return httpx.OK(c, res)
}

// Logout revokes the current session and clears the session cookie.
//
//	@Summary		Logout current session
//	@Description	Revokes the active session token and clears the session cookie.
//	@ID				logoutAuth
//	@Tags			Auth
//	@Success		204
//	@Router			/auth/logout [post]
func (h *Handler) Logout(c fiber.Ctx) error {
	rawToken := c.Cookies(h.cookieName)
	if err := h.service.Logout(c.Context(), rawToken); err != nil {
		return err
	}
	h.clearCookie(c, h.cookieName)
	return httpx.NoContent(c)
}

// OAuthRedirect initiates an OAuth2 login flow for the given provider.
//
//	@Summary		Start OAuth2 login
//	@Description	Redirects to the configured OAuth2 provider authorization endpoint.
//	@ID				oauthRedirect
//	@Tags			Auth
//	@Param			provider	path	string	true	"OAuth2 provider (google, github)"
//	@Param			redirect	query	string	false	"Relative post-login redirect path"
//	@Success		302
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Router			/auth/{provider} [get]
func (h *Handler) OAuthRedirect(c fiber.Ctx) error {
	provider := c.Params("provider")
	state, err := generateOAuthState()
	if err != nil {
		return err
	}
	callbackURI := c.BaseURL() + "/api/auth/" + provider + "/callback"
	authURL, err := h.service.OAuthAuthURL(provider, state, callbackURI)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   oauthCookieMaxAge,
	})
	c.Cookie(&fiber.Cookie{
		Name:     oauthRedirectCookieName,
		Value:    sanitizeRedirect(c.Query("redirect")),
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   oauthCookieMaxAge,
	})
	return c.Redirect().Status(fiber.StatusFound).To(authURL)
}

// OAuthCallback completes an OAuth2 login flow and sets the session cookie.
//
//	@Summary		Complete OAuth2 login
//	@Description	Validates OAuth2 state, exchanges code for profile, and sets session cookie.
//	@ID				oauthCallback
//	@Tags			Auth
//	@Param			provider	path	string	true	"OAuth2 provider (google, github)"
//	@Param			code		query	string	true	"Authorization code"
//	@Param			state		query	string	true	"CSRF state token"
//	@Success		302
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/auth/{provider}/callback [get]
func (h *Handler) OAuthCallback(c fiber.Ctx) error {
	provider := c.Params("provider")
	stateParam := c.Query("state")
	stateCookie := c.Cookies(oauthStateCookieName)
	redirectTo := sanitizeRedirect(c.Cookies(oauthRedirectCookieName))

	h.clearCookie(c, oauthStateCookieName)
	h.clearCookie(c, oauthRedirectCookieName)

	if stateParam == "" || stateCookie == "" || subtle.ConstantTimeCompare([]byte(stateParam), []byte(stateCookie)) != 1 {
		return domain.ErrOAuthStateInvalid
	}

	code := c.Query("code")
	if code == "" {
		return domain.ErrOAuthStateInvalid
	}

	callbackURI := c.BaseURL() + "/api/auth/" + provider + "/callback"
	rawToken, _, err := h.service.OAuthCallback(c.Context(), provider, code, callbackURI, c.Get(fiber.HeaderUserAgent), c.IP())
	if err != nil {
		return err
	}

	h.setSessionCookie(c, rawToken)
	return c.Redirect().Status(fiber.StatusFound).To(redirectTo)
}

func (h *Handler) setSessionCookie(c fiber.Ctx, rawToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     h.cookieName,
		Value:    rawToken,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   int(h.sessionTTL.Seconds()),
	})
}

func (h *Handler) clearCookie(c fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func sanitizeRedirect(raw string) string {
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return "/"
	}
	return raw
}

func generateOAuthState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
