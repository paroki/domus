package httpx

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// Base sentinel errors for HTTP error mapping and errors.Is checks.
var (
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrGone         = errors.New("gone")
	ErrValidation   = errors.New("validation failed")
	ErrRateLimited  = errors.New("rate limited")
	ErrInternal     = errors.New("internal error")
)

// AppError represents a structured HTTP error returned by handlers or services.
type AppError struct {
	Status  int
	Code    string
	Message string
	Details []FieldError
	Err     error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

// Unwrap returns the wrapped sentinel or underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewError constructs an AppError with explicit status, code, message, and optional field details.
func NewError(status int, code, message string, details ...FieldError) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}

// BadRequest constructs a 400 BAD_REQUEST AppError wrapping ErrBadRequest.
func BadRequest(message string) *AppError {
	return &AppError{Status: fiber.StatusBadRequest, Code: "BAD_REQUEST", Message: message, Err: ErrBadRequest}
}

// Unauthorized constructs a 401 UNAUTHORIZED AppError wrapping ErrUnauthorized.
func Unauthorized(message string) *AppError {
	return &AppError{Status: fiber.StatusUnauthorized, Code: "UNAUTHORIZED", Message: message, Err: ErrUnauthorized}
}

// Forbidden constructs a 403 FORBIDDEN AppError wrapping ErrForbidden.
func Forbidden(message string) *AppError {
	return &AppError{Status: fiber.StatusForbidden, Code: "FORBIDDEN", Message: message, Err: ErrForbidden}
}

// NotFound constructs a 404 NOT_FOUND AppError wrapping ErrNotFound.
func NotFound(message string) *AppError {
	return &AppError{Status: fiber.StatusNotFound, Code: "NOT_FOUND", Message: message, Err: ErrNotFound}
}

// Conflict constructs a 409 CONFLICT AppError wrapping ErrConflict.
func Conflict(message string) *AppError {
	return &AppError{Status: fiber.StatusConflict, Code: "CONFLICT", Message: message, Err: ErrConflict}
}

// Gone constructs a 410 GONE AppError wrapping ErrGone.
func Gone(message string) *AppError {
	return &AppError{Status: fiber.StatusGone, Code: "GONE", Message: message, Err: ErrGone}
}

// ValidationError constructs a 422 VALIDATION_ERROR AppError wrapping ErrValidation.
func ValidationError(details ...FieldError) *AppError {
	return &AppError{
		Status:  fiber.StatusUnprocessableEntity,
		Code:    "VALIDATION_ERROR",
		Message: "One or more fields failed validation.",
		Details: details,
		Err:     ErrValidation,
	}
}

// RateLimited constructs a 429 RATE_LIMITED AppError wrapping ErrRateLimited.
func RateLimited(message string) *AppError {
	return &AppError{Status: fiber.StatusTooManyRequests, Code: "RATE_LIMITED", Message: message, Err: ErrRateLimited}
}

// InternalError constructs a 500 INTERNAL_ERROR AppError without exposing internal details.
func InternalError(err error) *AppError {
	wrapped := ErrInternal
	if err != nil {
		wrapped = errors.Join(ErrInternal, err)
	}
	return &AppError{
		Status:  fiber.StatusInternalServerError,
		Code:    "INTERNAL_ERROR",
		Message: "An unexpected internal error occurred.",
		Err:     wrapped,
	}
}

// ErrorHandler returns a Fiber v3 ErrorHandler that logs and formats errors as Envelope[any].
func ErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return func(c fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		code := "INTERNAL_ERROR"
		message := "An unexpected internal error occurred."
		var details []FieldError

		if appErr, ok := errors.AsType[*AppError](err); ok {
			status = appErr.Status
			if status >= fiber.StatusInternalServerError {
				code = "INTERNAL_ERROR"
				message = "An unexpected internal error occurred."
			} else {
				code = appErr.Code
				message = appErr.Message
				details = appErr.Details
			}
		} else if valErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			status = fiber.StatusUnprocessableEntity
			code = "VALIDATION_ERROR"
			message = "One or more fields failed validation."
			details = make([]FieldError, 0, len(valErrs))
			for _, fe := range valErrs {
				details = append(details, FieldError{
					Field:   lowerFirstRune(fe.Field()),
					Code:    toUpperSnake(fe.Tag()),
					Message: formatValidationMessage(fe),
				})
			}
		} else if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			status = fiberErr.Code
			if status >= fiber.StatusInternalServerError {
				code = "INTERNAL_ERROR"
				message = "An unexpected internal error occurred."
			} else {
				code = defaultCodeForStatus(status)
				message = fiberErr.Message
			}
		}

		reqID := requestid.FromContext(c)
		attrs := []any{
			"request_id", reqID,
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"code", code,
			"err", err,
		}
		if status >= fiber.StatusInternalServerError {
			logger.ErrorContext(c.Context(), "request error", attrs...)
		} else if status >= fiber.StatusBadRequest {
			logger.WarnContext(c.Context(), "request error", attrs...)
		}

		return c.Status(status).JSON(Envelope[any]{
			Error: &ErrorPayload{
				Code:      code,
				Message:   message,
				RequestID: reqID,
				Details:   details,
			},
		})
	}
}

func defaultCodeForStatus(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "BAD_REQUEST"
	case fiber.StatusUnauthorized:
		return "UNAUTHORIZED"
	case fiber.StatusForbidden:
		return "FORBIDDEN"
	case fiber.StatusNotFound:
		return "NOT_FOUND"
	case fiber.StatusConflict:
		return "CONFLICT"
	case fiber.StatusGone:
		return "GONE"
	case fiber.StatusUnprocessableEntity:
		return "VALIDATION_ERROR"
	case fiber.StatusTooManyRequests:
		return "RATE_LIMITED"
	default:
		return "INTERNAL_ERROR"
	}
}

func lowerFirstRune(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func toUpperSnake(tag string) string {
	var b strings.Builder
	for i, r := range tag {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

func formatValidationMessage(fe validator.FieldError) string {
	if fe.Param() != "" {
		return fmt.Sprintf("failed %s=%s validation", fe.Tag(), fe.Param())
	}
	return fmt.Sprintf("failed %s validation", fe.Tag())
}
