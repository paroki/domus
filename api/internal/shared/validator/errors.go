package validator

// FieldError represents a single field-level validation failure.
// Field holds the JSON field name; Issue holds the human-readable message.
type FieldError struct {
	Field string
	Issue string
}
