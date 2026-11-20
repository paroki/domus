package errx

// Kind classifies a domain error independent of transport (HTTP, gRPC, MCP, CLI).
type Kind uint8

const (
	KindInternal Kind = iota
	KindInvalid
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
	KindGone
	KindValidation
	KindRateLimited
)

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field   string
	Code    string
	Message string
}

// Error is a transport-agnostic domain error carrying a category, stable code,
// English developer fallback message, and optional validation details or cause.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Details []FieldError
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return e.Code
	}
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return "internal error"
}

// Unwrap returns the underlying cause, if any.
func (e *Error) Unwrap() error {
	return e.Cause
}

// New constructs a domain Error with an explicit Kind, code, and message.
func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

// Invalid constructs a KindInvalid domain Error.
func Invalid(code, message string) *Error {
	return New(KindInvalid, code, message)
}

// Unauthorized constructs a KindUnauthorized domain Error.
func Unauthorized(code, message string) *Error {
	return New(KindUnauthorized, code, message)
}

// Forbidden constructs a KindForbidden domain Error.
func Forbidden(code, message string) *Error {
	return New(KindForbidden, code, message)
}

// NotFound constructs a KindNotFound domain Error.
func NotFound(code, message string) *Error {
	return New(KindNotFound, code, message)
}

// Conflict constructs a KindConflict domain Error.
func Conflict(code, message string) *Error {
	return New(KindConflict, code, message)
}

// Gone constructs a KindGone domain Error.
func Gone(code, message string) *Error {
	return New(KindGone, code, message)
}

// ValidationError constructs a KindValidation domain Error with field details.
func ValidationError(details ...FieldError) *Error {
	return &Error{
		Kind:    KindValidation,
		Code:    "VALIDATION_ERROR",
		Message: "One or more fields failed validation.",
		Details: details,
	}
}

// RateLimited constructs a KindRateLimited domain Error.
func RateLimited(code, message string) *Error {
	return New(KindRateLimited, code, message)
}

// Internal constructs a KindInternal domain Error wrapping an optional cause.
func Internal(code, message string, cause error) *Error {
	return &Error{Kind: KindInternal, Code: code, Message: message, Cause: cause}
}
