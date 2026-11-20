package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the org feature.
var (
	ErrOrgNotFound          = errx.NotFound("ORG_NOT_FOUND", "Organization not found.")
	ErrCodeConflict         = errx.Conflict("ORG_CODE_CONFLICT", "Organization code is already in use in this parish.")
	ErrInUse                = errx.Conflict("ORG_IN_USE", "Organization has active child organizations or records and cannot be deleted.")
	ErrForbidden            = errx.Forbidden("ORG_FORBIDDEN", "You do not have permission to perform this organization action.")
	ErrMemberNotFound       = errx.NotFound("ORG_MEMBER_NOT_FOUND", "Organization member not found.")
	ErrMemberConflict       = errx.Conflict("ORG_MEMBER_ALREADY_EXISTS", "User is already a member of this organization.")
	ErrLastAdminRequired    = errx.Conflict("ORG_LAST_ADMIN_REQUIRED", "At least one active organization admin is required.")
	ErrInviteNotFound       = errx.NotFound("ORG_INVITE_NOT_FOUND", "Organization invite link not found.")
	ErrInviteExpired        = errx.Gone("ORG_INVITE_EXPIRED", "Organization invite link has expired.")
	ErrInviteRevoked        = errx.Gone("ORG_INVITE_REVOKED", "Organization invite link has been revoked.")
	ErrInviteMaxUsesReached = errx.Conflict("ORG_INVITE_MAX_USES_REACHED", "Organization invite link has reached its maximum uses.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
