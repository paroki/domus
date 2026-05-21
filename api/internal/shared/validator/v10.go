package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// V10Validator is the go-playground/validator/v10 implementation of Validator.
type V10Validator struct {
	v   *validator.Validate
	msg MessageProvider
}

// NewV10Validator creates a new V10Validator with the provided MessageProvider.
// If mp is nil, DefaultMessageProvider (Indonesian) is used.
func NewV10Validator(mp MessageProvider) *V10Validator {
	if mp == nil {
		mp = DefaultMessageProvider()
	}

	v := validator.New()

	// Register function to extract the `json` tag as the field name so that
	// FieldError.Field matches the JSON key used in the API response.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	return &V10Validator{v: v, msg: mp}
}

// Validate implements Validator.
// Returns an empty slice when v is valid; otherwise returns one FieldError per
// failing constraint.
func (val *V10Validator) Validate(v any) []FieldError {
	err := val.v.Struct(v)
	if err == nil {
		return []FieldError{}
	}

	var errs validator.ValidationErrors
	if !errorAs(err, &errs) {
		// Unexpected error type (e.g., invalid struct passed); treat as single error.
		return []FieldError{{Field: "", Issue: err.Error()}}
	}

	result := make([]FieldError, 0, len(errs))
	for _, fe := range errs {
		result = append(result, FieldError{
			Field: fe.Field(),
			Issue: val.msg.Message(fe.Tag(), fe.Param()),
		})
	}
	return result
}

// errorAs is a thin wrapper around errors.As for testability.
func errorAs(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}
