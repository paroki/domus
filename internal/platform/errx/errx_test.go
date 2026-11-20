package errx_test

import (
	"errors"
	"testing"

	"github.com/paroki/domus/internal/platform/errx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstructorsAndSentinelMatching(t *testing.T) {
	sentinel := errx.NotFound("USER_NOT_FOUND", "User not found.")
	wrapped := errors.Join(sentinel, errors.New("db row missing"))

	assert.ErrorIs(t, wrapped, sentinel)

	got, ok := errors.AsType[*errx.Error](wrapped)
	require.True(t, ok)
	assert.Equal(t, errx.KindNotFound, got.Kind)
	assert.Equal(t, "USER_NOT_FOUND", got.Code)
	assert.Equal(t, "User not found.", got.Message)
}

func TestValidationError(t *testing.T) {
	err := errx.ValidationError(errx.FieldError{
		Field:   "email",
		Code:    "EMAIL",
		Message: "must be valid",
	})

	assert.Equal(t, errx.KindValidation, err.Kind)
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	require.Len(t, err.Details, 1)
	assert.Equal(t, "email", err.Details[0].Field)
}

func TestInternalUnwrap(t *testing.T) {
	cause := errors.New("connection refused")
	err := errx.Internal("DB_ERROR", "Database unavailable.", cause)

	assert.Equal(t, errx.KindInternal, err.Kind)
	assert.ErrorIs(t, err, cause)
	assert.Equal(t, "Database unavailable.", err.Error())
}
