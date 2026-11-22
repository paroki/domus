package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the agent feature.
var (
	ErrSessionNotFound              = errx.NotFound("AGENT_SESSION_NOT_FOUND", "Agent conversation session not found.")
	ErrMemoryNotFound               = errx.NotFound("AGENT_MEMORY_NOT_FOUND", "Agent memory fact not found.")
	ErrForbidden                    = errx.Forbidden("AGENT_FORBIDDEN", "You do not have permission to perform this agent action.")
	ErrEmptyMessage                 = errx.Invalid("AGENT_EMPTY_MESSAGE", "Message content cannot be empty.")
	ErrWebhookUnauthorized          = errx.Unauthorized("AGENT_WEBHOOK_UNAUTHORIZED", "Invalid webhook signature or secret token.")
	ErrSessionBusy                  = errx.Conflict("AGENT_SESSION_BUSY", "Another turn is already in progress for this conversation session.")
	ErrProviderNotFound             = errx.NotFound("AGENT_PROVIDER_NOT_FOUND", "Agent provider configuration not found.")
	ErrProviderInUse                = errx.Conflict("AGENT_PROVIDER_IN_USE", "Agent provider configuration is currently in use and cannot be deleted.")
	ErrProviderEmbeddingUnsupported = errx.Invalid("AGENT_PROVIDER_EMBEDDING_UNSUPPORTED", "Anthropic provider does not support text embeddings.")
	ErrProviderUnavailable          = errx.Invalid("AGENT_PROVIDER_UNAVAILABLE", "The configured AI provider is unavailable or missing an API key.")
	ErrMaxStepsExceeded             = &errx.Error{Kind: errx.KindValidation, Code: "AGENT_MAX_STEPS_EXCEEDED", Message: "Agent exceeded the maximum number of tool-calling steps."}
	ErrMemoryLimitExceeded          = &errx.Error{Kind: errx.KindValidation, Code: "AGENT_MEMORY_LIMIT_EXCEEDED", Message: "Active or pinned memory limit exceeded."}
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
