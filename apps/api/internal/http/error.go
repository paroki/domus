package http

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/model"
)

// MapError converts standard domain/validation/framework errors into an HTTP status and ErrorBody.
func MapError(err error) (int, model.ErrorBody) {
	var ve validator.ValidationErrors
	var fe *fiber.Error

	switch {
	case errors.Is(err, core.ErrItemNotFound):
		return fiber.StatusNotFound, model.ErrorBody{Code: "NOT_FOUND", Message: "resource not found"}
	case errors.Is(err, core.ErrForbidden):
		return fiber.StatusForbidden, model.ErrorBody{Code: "FORBIDDEN", Message: "forbidden"}
	case errors.As(err, &ve):
		fields := make([]model.FieldError, 0, len(ve))
		for _, f := range ve {
			fields = append(fields, model.FieldError{Field: f.Field(), Rule: f.Tag()})
		}
		return fiber.StatusUnprocessableEntity, model.ErrorBody{Code: "VALIDATION_FAILED", Message: "validation failed", Fields: fields}
	case errors.As(err, &fe):
		return fe.Code, model.ErrorBody{Code: "HTTP_ERROR", Message: fe.Message}
	default:
		// jangan bocorin detail internal ke client
		return fiber.StatusInternalServerError, model.ErrorBody{Code: "INTERNAL", Message: "internal server error"}
	}
}
