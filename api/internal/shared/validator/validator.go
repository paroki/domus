package validator

// MessageProvider translates a validator tag and its parameter into a
// human-readable message. Implement this interface to support different locales.
type MessageProvider interface {
	// Message returns a human-readable error message for the given validation
	// tag and optional parameter value (e.g., "3" for `min=3`).
	Message(tag string, param string) string
}

// Validator is the abstraction for input validation.
// Consumers depend on this interface, not on the concrete implementation.
// An empty slice return value indicates that the value is valid.
type Validator interface {
	// Validate validates the given struct and returns a slice of FieldError.
	// Returns an empty slice if the value is valid.
	Validate(v any) []FieldError
}
