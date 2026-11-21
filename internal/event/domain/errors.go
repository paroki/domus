package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the event feature.
var (
	ErrEventNotFound    = errx.NotFound("EVENT_NOT_FOUND", "Event not found.")
	ErrForbidden        = errx.Forbidden("EVENT_FORBIDDEN", "You do not have permission to perform this event action.")
	ErrInvalidTimeRange = errx.Invalid("EVENT_INVALID_TIME_RANGE", "Event end time must be after start time, and RSVP deadline cannot exceed end time.")
	ErrRSVPClosed       = errx.Conflict("EVENT_RSVP_CLOSED", "RSVP is closed for this event.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
