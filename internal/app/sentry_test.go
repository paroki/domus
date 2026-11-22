package app

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
)

func TestSentryCapturesOnly5xx(t *testing.T) {
	is5xx := func(err error) bool {
		return httpx.ErrorStatus(err) >= fiber.StatusInternalServerError
	}

	assert.False(t, is5xx(errx.Invalid("AGENT_PROVIDER_UNAVAILABLE", "provider missing")))
	assert.False(t, is5xx(httpx.BadRequest("bad input")))
	assert.True(t, is5xx(errx.Unavailable("UPSTREAM_DOWN", "upstream down")))
	assert.True(t, is5xx(errx.Internal("INTERNAL_ERROR", "db down", nil)))
	assert.True(t, is5xx(errors.New("unexpected")))
}
