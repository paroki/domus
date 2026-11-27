package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the liturgi feature.
var (
	ErrScheduleNotFound = errx.NotFound("LITURGI_SCHEDULE_NOT_FOUND", "Liturgy Mass schedule not found.")
	ErrPetugasNotFound  = errx.NotFound("LITURGI_PETUGAS_NOT_FOUND", "Liturgy officer assignment not found.")
	ErrLaguNotFound     = errx.NotFound("LITURGI_LAGU_NOT_FOUND", "Liturgy song entry not found.")
	ErrForbidden        = errx.Forbidden("LITURGI_FORBIDDEN", "You do not have permission to perform this liturgy action.")
	ErrInvalidTimeRange = errx.Invalid("LITURGI_INVALID_TIME_RANGE", "Mass end time must be after start time, and RSVP deadline cannot exceed end time.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
