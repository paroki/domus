package domain

import "github.com/paroki/domus/internal/platform/errx"

// Sentinel errors for the umat feature.
var (
	ErrKeluargaNotFound = errx.NotFound("UMAT_KELUARGA_NOT_FOUND", "Keluarga not found.")
	ErrUmatNotFound     = errx.NotFound("UMAT_NOT_FOUND", "Umat not found.")
	ErrSakramenNotFound = errx.NotFound("UMAT_SAKRAME_NOT_FOUND", "Sakramen not found.")
)
