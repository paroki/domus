package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the ref feature.
var (
	ErrDioceseNotFound         = errx.NotFound("REF_DIOCESE_NOT_FOUND", "Diocese not found.")
	ErrParishNotFound          = errx.NotFound("REF_PARISH_NOT_FOUND", "Parish not found.")
	ErrParishMemberNotFound    = errx.NotFound("REF_PARISH_MEMBER_NOT_FOUND", "Parish member not found.")
	ErrZoneNotFound            = errx.NotFound("REF_ZONE_NOT_FOUND", "Zone not found.")
	ErrWardNotFound            = errx.NotFound("REF_WARD_NOT_FOUND", "Ward not found.")
	ErrLookupNotFound          = errx.NotFound("REF_LOOKUP_NOT_FOUND", "Lookup not found.")
	ErrPeriodNotFound          = errx.NotFound("REF_PERIOD_NOT_FOUND", "Period not found.")
	ErrCodeConflict            = errx.Conflict("REF_CODE_CONFLICT", "Code is already in use.")
	ErrPeriodOverlap           = errx.Conflict("REF_PERIOD_OVERLAP", "Period dates overlap an existing period.")
	ErrInUse                   = errx.Conflict("REF_IN_USE", "Resource is still in use and cannot be deleted.")
	ErrLastParishAdminRequired = errx.Conflict("REF_LAST_PARISH_ADMIN_REQUIRED", "At least one parish admin is required.")
	ErrParishMemberConflict    = errx.Conflict("REF_PARISH_MEMBER_ALREADY_EXISTS", "User is already a member of this parish.")
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
