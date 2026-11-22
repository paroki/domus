package app

import (
	"errors"
	"testing"

	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
)

func TestShouldCaptureSentry(t *testing.T) {
	assert.False(t, shouldCaptureSentry(errx.Unavailable("AGENT_PROVIDER_UNAVAILABLE", "provider missing")))
	assert.False(t, shouldCaptureSentry(httpx.BadRequest("bad input")))
	assert.True(t, shouldCaptureSentry(errx.Internal("INTERNAL_ERROR", "db down", nil)))
	assert.True(t, shouldCaptureSentry(errors.New("unexpected")))
}
