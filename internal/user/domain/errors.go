package domain

import (
	"net/http"

	"github.com/paroki/domus/internal/platform/httpx"
)

// Sentinel errors for the user and authentication feature.
var (
	ErrUserNotFound           = httpx.NewError(http.StatusNotFound, "USER_NOT_FOUND", "User not found.")
	ErrUserEmailConflict      = httpx.NewError(http.StatusConflict, "USER_EMAIL_CONFLICT", "Email address is already in use.")
	ErrUserDisabled           = httpx.NewError(http.StatusForbidden, "USER_DISABLED", "User account is disabled.")
	ErrLastSuperAdminRequired = httpx.NewError(http.StatusConflict, "USER_LAST_SUPER_ADMIN_REQUIRED", "At least one active super admin is required.")
	ErrOTPInvalid             = httpx.NewError(http.StatusUnauthorized, "AUTH_OTP_INVALID", "Invalid verification code.")
	ErrOTPExpired             = httpx.NewError(http.StatusUnauthorized, "AUTH_OTP_EXPIRED", "Verification code has expired.")
	ErrOTPRateLimited         = httpx.NewError(http.StatusTooManyRequests, "AUTH_OTP_RATE_LIMITED", "Please wait before requesting another verification code.")
	ErrOAuthProviderUnknown   = httpx.NewError(http.StatusBadRequest, "AUTH_OAUTH_PROVIDER_UNKNOWN", "Unknown or unconfigured OAuth provider.")
	ErrOAuthStateInvalid      = httpx.NewError(http.StatusBadRequest, "AUTH_OAUTH_STATE_INVALID", "Invalid or expired OAuth state.")
	ErrOAuthEmailUnverified   = httpx.NewError(http.StatusForbidden, "AUTH_OAUTH_EMAIL_UNVERIFIED", "OAuth provider email is not verified.")
	ErrSessionNotFound        = httpx.NewError(http.StatusNotFound, "USER_SESSION_NOT_FOUND", "Session not found.")
	ErrAccountNotFound        = httpx.NewError(http.StatusNotFound, "USER_ACCOUNT_NOT_FOUND", "Linked account not found.")
	ErrLastAccountRequired    = httpx.NewError(http.StatusConflict, "USER_LAST_ACCOUNT_REQUIRED", "Cannot unlink the last authentication account.")
)
