package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the cantus feature.
var (
	ErrCantusNotFound     = errx.NotFound("CANTUS_NOT_FOUND", "Cantus entry not found.")
	ErrCantusCodeConflict = errx.Conflict("CANTUS_CODE_CONFLICT", "A cantus entry with this slug already exists.")
	ErrCantusInUse        = errx.Conflict("CANTUS_IN_USE", "Cantus entry is currently in use and cannot be deleted.")
	ErrCantusForbidden    = errx.Forbidden("CANTUS_FORBIDDEN", "Cantus admin role is required.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
