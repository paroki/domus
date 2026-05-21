package validator_test

import (
	"testing"

	"github.com/paroki/domus/api/internal/shared/validator"
)

// -- Test structs -------------------------------------------------------------

type loginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type registrationInput struct {
	Username string `json:"username"  validate:"required,alphanum,min=3,max=30"`
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8"`
	Confirm  string `json:"confirm"   validate:"required,eqfield=Password"`
	Age      int    `json:"age"       validate:"gte=17"`
}

type noTagStruct struct {
	Name string `validate:"required"`
}

// -- Helpers ------------------------------------------------------------------

func newValidator() validator.Validator {
	return validator.NewV10Validator(validator.DefaultMessageProvider())
}

func findError(errs []validator.FieldError, field string) (validator.FieldError, bool) {
	for _, e := range errs {
		if e.Field == field {
			return e, true
		}
	}
	return validator.FieldError{}, false
}

// -- Tests -------------------------------------------------------------------

func TestValidate_ValidStruct_ReturnsEmptySlice(t *testing.T) {
	v := newValidator()
	input := loginInput{Email: "user@example.com", Password: "secret123"}
	errs := v.Validate(input)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %+v", len(errs), errs)
	}
}

func TestValidate_RequiredFieldMissing_ReturnsFieldError(t *testing.T) {
	v := newValidator()
	input := loginInput{Email: "", Password: ""}
	errs := v.Validate(input)

	if len(errs) == 0 {
		t.Fatal("expected validation errors, got none")
	}

	emailErr, ok := findError(errs, "email")
	if !ok {
		t.Fatal("expected error on 'email' field")
	}
	if emailErr.Issue == "" {
		t.Error("expected non-empty Issue for email field")
	}

	// Field name must match json tag, not Go field name.
	for _, e := range errs {
		if e.Field == "Email" || e.Field == "Password" {
			t.Errorf("field name should be json tag, got Go name: %q", e.Field)
		}
	}
}

func TestValidate_InvalidEmail_ReturnsCorrectIssue(t *testing.T) {
	v := newValidator()
	input := loginInput{Email: "not-an-email", Password: "secret123"}
	errs := v.Validate(input)

	emailErr, ok := findError(errs, "email")
	if !ok {
		t.Fatal("expected error on 'email' field")
	}
	want := "Harus berupa alamat email yang valid."
	if emailErr.Issue != want {
		t.Errorf("Issue = %q, want %q", emailErr.Issue, want)
	}
}

func TestValidate_MinViolation_ReturnsIssueWithParam(t *testing.T) {
	v := newValidator()
	input := loginInput{Email: "user@example.com", Password: "short"}
	errs := v.Validate(input)

	pwErr, ok := findError(errs, "password")
	if !ok {
		t.Fatal("expected error on 'password' field")
	}
	want := "Minimal 8 karakter."
	if pwErr.Issue != want {
		t.Errorf("Issue = %q, want %q", pwErr.Issue, want)
	}
}

func TestValidate_EqField_ReturnsIssue(t *testing.T) {
	v := newValidator()
	input := registrationInput{
		Username: "toni123",
		Email:    "toni@example.com",
		Password: "secure123",
		Confirm:  "different",
		Age:      25,
	}
	errs := v.Validate(input)

	confirmErr, ok := findError(errs, "confirm")
	if !ok {
		t.Fatal("expected error on 'confirm' field")
	}
	if confirmErr.Issue == "" {
		t.Error("expected non-empty Issue for confirm field")
	}
}

func TestValidate_GteViolation_ReturnsIssue(t *testing.T) {
	v := newValidator()
	input := registrationInput{
		Username: "toni123",
		Email:    "toni@example.com",
		Password: "secure123",
		Confirm:  "secure123",
		Age:      16, // below gte=17
	}
	errs := v.Validate(input)

	ageErr, ok := findError(errs, "age")
	if !ok {
		t.Fatal("expected error on 'age' field")
	}
	want := "Harus lebih besar atau sama dengan 17."
	if ageErr.Issue != want {
		t.Errorf("Issue = %q, want %q", ageErr.Issue, want)
	}
}

func TestValidate_NoJsonTag_UsesGoFieldName(t *testing.T) {
	v := newValidator()
	input := noTagStruct{Name: ""}
	errs := v.Validate(input)

	if len(errs) == 0 {
		t.Fatal("expected validation errors, got none")
	}
	// Without json tag, falls back to Go field name.
	_, ok := findError(errs, "Name")
	if !ok {
		t.Fatalf("expected error on 'Name' field, got: %+v", errs)
	}
}

// -- Custom MessageProvider ---------------------------------------------------

type englishMessages struct{}

func (e *englishMessages) Message(tag, param string) string {
	switch tag {
	case "required":
		return "This field is required."
	case "email":
		return "Must be a valid email address."
	case "min":
		return "Must be at least " + param + " characters."
	default:
		return "Validation failed."
	}
}

func TestValidate_CustomMessageProvider_UsesCustomMessages(t *testing.T) {
	v := validator.NewV10Validator(&englishMessages{})
	input := loginInput{Email: "bad", Password: ""}

	errs := v.Validate(input)

	emailErr, ok := findError(errs, "email")
	if !ok {
		t.Fatal("expected error on 'email' field")
	}
	want := "Must be a valid email address."
	if emailErr.Issue != want {
		t.Errorf("Issue = %q, want %q", emailErr.Issue, want)
	}

	pwErr, ok := findError(errs, "password")
	if !ok {
		t.Fatal("expected error on 'password' field")
	}
	wantPw := "This field is required."
	if pwErr.Issue != wantPw {
		t.Errorf("Issue = %q, want %q", pwErr.Issue, wantPw)
	}
}

func TestNewV10Validator_NilMessageProvider_UsesDefault(t *testing.T) {
	v := validator.NewV10Validator(nil)
	input := loginInput{Email: "", Password: ""}
	errs := v.Validate(input)
	if len(errs) == 0 {
		t.Fatal("expected errors")
	}
	// Default (Indonesian) messages should be used.
	emailErr, ok := findError(errs, "email")
	if !ok {
		t.Fatal("expected error on 'email'")
	}
	want := "Field ini wajib diisi."
	if emailErr.Issue != want {
		t.Errorf("Issue = %q, want %q", emailErr.Issue, want)
	}
}
