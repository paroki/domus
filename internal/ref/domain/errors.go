package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the ref feature.
var (
	ErrKeuskupanNotFound       = errx.NotFound("REF_KEUSKUPAN_NOT_FOUND", "Keuskupan not found.")
	ErrParokiNotFound          = errx.NotFound("REF_PAROKI_NOT_FOUND", "Paroki not found.")
	ErrParokiMemberNotFound    = errx.NotFound("REF_PAROKI_MEMBER_NOT_FOUND", "Paroki member not found.")
	ErrWilayahNotFound         = errx.NotFound("REF_WILAYAH_NOT_FOUND", "Wilayah not found.")
	ErrLingkunganNotFound      = errx.NotFound("REF_LINGKUNGAN_NOT_FOUND", "Lingkungan not found.")
	ErrLookupNotFound          = errx.NotFound("REF_LOOKUP_NOT_FOUND", "Lookup not found.")
	ErrPeriodNotFound          = errx.NotFound("REF_PERIOD_NOT_FOUND", "Period not found.")
	ErrCodeConflict            = errx.Conflict("REF_CODE_CONFLICT", "Code is already in use.")
	ErrPeriodOverlap           = errx.Conflict("REF_PERIOD_OVERLAP", "Period dates overlap an existing period.")
	ErrInUse                   = errx.Conflict("REF_IN_USE", "Resource is still in use and cannot be deleted.")
	ErrLastParokiAdminRequired = errx.Conflict("REF_LAST_PAROKI_ADMIN_REQUIRED", "At least one paroki admin is required.")
	ErrParokiMemberConflict    = errx.Conflict("REF_PAROKI_MEMBER_ALREADY_EXISTS", "User is already a member of this paroki.")
	ErrRefForbidden            = errx.Forbidden("REF_FORBIDDEN", "You do not have permission to perform this action.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
