package config

import (
	"github.com/paroki/domus/api/internal/shared/validator"
)

// NewValidator returns a Validator wired with the default Indonesian MessageProvider.
// Swap the MessageProvider argument to support other locales.
func NewValidator() validator.Validator {
	return validator.NewV10Validator(validator.DefaultMessageProvider())
}
