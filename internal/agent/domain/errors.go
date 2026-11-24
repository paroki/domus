package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the agent feature.
var (
	ErrProviderNotFound       = errx.NotFound("AGENT_PROVIDER_NOT_FOUND", "Agent provider configuration not found.")
	ErrSessionNotFound        = errx.NotFound("AGENT_SESSION_NOT_FOUND", "Agent session not found.")
	ErrSessionBusy            = errx.Conflict("AGENT_SESSION_BUSY", "Agent session is currently processing another request.")
	ErrMemoryNotFound         = errx.NotFound("AGENT_MEMORY_NOT_FOUND", "Agent memory not found.")
	ErrProviderInUse          = errx.Conflict("AGENT_PROVIDER_IN_USE", "AI provider is currently assigned to an active agent model setting.")
	ErrProviderUnavailable    = errx.Invalid("AGENT_PROVIDER_UNAVAILABLE", "The configured AI provider is unavailable or missing an API key.")
	ErrProviderAuthFailed     = errx.Invalid("AGENT_PROVIDER_AUTH_FAILED", "AI provider authentication failed.")
	ErrProviderRateLimited    = errx.RateLimited("AGENT_PROVIDER_RATE_LIMITED", "AI provider rate limit exceeded.")
	ErrProviderOverloaded     = errx.Invalid("AGENT_PROVIDER_OVERLOADED", "AI provider is temporarily overloaded due to high demand.")
	ErrProviderRequestInvalid = errx.Invalid("AGENT_PROVIDER_REQUEST_INVALID", "AI provider rejected the request.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
