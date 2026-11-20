package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the user and authentication feature.
var (
	ErrUserNotFound           = errx.NotFound("USER_NOT_FOUND", "User not found.")
	ErrUserEmailConflict      = errx.Conflict("USER_EMAIL_CONFLICT", "Email address is already in use.")
	ErrUserDisabled           = errx.Forbidden("USER_DISABLED", "User account is disabled.")
	ErrLastSuperAdminRequired = errx.Conflict("USER_LAST_SUPER_ADMIN_REQUIRED", "At least one active super admin is required.")
	ErrOTPInvalid             = errx.Unauthorized("AUTH_OTP_INVALID", "Invalid verification code.")
	ErrOTPExpired             = errx.Unauthorized("AUTH_OTP_EXPIRED", "Verification code has expired.")
	ErrOTPRateLimited         = errx.RateLimited("AUTH_OTP_RATE_LIMITED", "Please wait before requesting another verification code.")
	ErrOAuthProviderUnknown   = errx.Invalid("AUTH_OAUTH_PROVIDER_UNKNOWN", "Unknown or unconfigured OAuth provider.")
	ErrOAuthStateInvalid      = errx.Invalid("AUTH_OAUTH_STATE_INVALID", "Invalid or expired OAuth state.")
	ErrOAuthEmailUnverified   = errx.Forbidden("AUTH_OAUTH_EMAIL_UNVERIFIED", "OAuth provider email is not verified.")
	ErrSessionNotFound        = errx.NotFound("USER_SESSION_NOT_FOUND", "Session not found.")
	ErrSessionInvalid         = errx.Unauthorized("UNAUTHORIZED", "Invalid or expired session.")
	ErrAccountNotFound        = errx.NotFound("USER_ACCOUNT_NOT_FOUND", "Linked account not found.")
	ErrLastAccountRequired    = errx.Conflict("USER_LAST_ACCOUNT_REQUIRED", "Cannot unlink the last authentication account.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
