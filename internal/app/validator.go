package app

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type structValidator struct {
	validate *validator.Validate
}

func newStructValidator() *structValidator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &structValidator{validate: v}
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}
