package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the umat feature.
var (
	ErrForbidden          = errx.Forbidden("UMAT_FORBIDDEN", "You do not have permission to perform this action.")
	ErrKeluargaNotFound   = errx.NotFound("UMAT_KELUARGA_NOT_FOUND", "Keluarga not found.")
	ErrUmatNotFound       = errx.NotFound("UMAT_NOT_FOUND", "Umat not found.")
	ErrSakramenNotFound   = errx.NotFound("UMAT_SAKRAMEN_NOT_FOUND", "Sakramen not found.")
	ErrDokumenNotFound    = errx.NotFound("UMAT_DOKUMEN_NOT_FOUND", "Dokumen not found.")
	ErrKKConflict         = errx.Conflict("UMAT_KK_CONFLICT", "Nomor KK is already registered in this paroki.")
	ErrNIKConflict        = errx.Conflict("UMAT_NIK_CONFLICT", "NIK is already registered in this paroki.")
	ErrRegistrasiConflict = errx.Conflict("UMAT_REGISTRASI_CONFLICT", "Nomor registrasi is already in use in this paroki.")
	ErrNomorIndukConflict = errx.Conflict("UMAT_NOMOR_INDUK_CONFLICT", "Nomor induk is already in use in this paroki.")
	ErrUserAlreadyLinked  = errx.Conflict("UMAT_USER_ALREADY_LINKED", "User is already linked to another umat in this paroki.")
	ErrInUse              = errx.Conflict("UMAT_IN_USE", "Resource is still in use and cannot be deleted.")
	ErrSakramenConflict   = errx.Conflict("UMAT_SAKRAMEN_CONFLICT", "Sakramen of this type already exists for this umat.")
	ErrReferensiNotFound  = errx.NotFound("UMAT_REFERENSI_NOT_FOUND", "Umat referensi not found.")
	ErrReferensiConflict  = errx.Conflict("UMAT_REFERENSI_CONFLICT", "Code is already in use for this category and scope.")
)

// ValidationError constructs a 422 VALIDATION_ERROR domain error for a single field.
func ValidationError(field, code, message string) error {
	return errx.ValidationError(errx.FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}
