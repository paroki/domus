package httpx

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

// MapError converts standard domain/validation/framework errors into an HTTP status and ErrorBody.
func MapError(err error) (int, ErrorBody) {
	var ve validator.ValidationErrors
	var fe *fiber.Error

	switch {
	case errors.Is(err, core.ErrItemNotFound):
		return fiber.StatusNotFound, ErrorBody{Code: "NOT_FOUND", Message: "resource not found"}
	case errors.Is(err, core.ErrInvalidID):
		return fiber.StatusBadRequest, ErrorBody{Code: "INVALID_ID", Message: "invalid id parameter"}
	case errors.Is(err, core.ErrForbidden):
		return fiber.StatusForbidden, ErrorBody{Code: "FORBIDDEN", Message: "forbidden"}
	case errors.As(err, &ve):
		fields := make([]FieldError, 0, len(ve))
		for _, f := range ve {
			fields = append(fields, FieldError{Field: f.Field(), Rule: f.Tag()})
		}
		return fiber.StatusUnprocessableEntity, ErrorBody{Code: "VALIDATION_FAILED", Message: "validation failed", Fields: fields}
	case errors.As(err, &fe):
		return fe.Code, ErrorBody{Code: "HTTP_ERROR", Message: fe.Message}
	default:
		// jangan bocorin detail internal ke client
		return fiber.StatusInternalServerError, ErrorBody{Code: "INTERNAL", Message: "internal server error"}
	}
}
