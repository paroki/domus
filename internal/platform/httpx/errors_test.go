package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandler_Unavailable503(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.ErrorHandler(nil),
	})
	app.Get("/unavailable", func(c fiber.Ctx) error {
		return errx.Unavailable("AGENT_PROVIDER_UNAVAILABLE", "AI provider is unavailable.")
	})

	req := httptest.NewRequest(http.MethodGet, "/unavailable", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var body httpx.Envelope[any]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.NotNil(t, body.Error)
	assert.Equal(t, "AGENT_PROVIDER_UNAVAILABLE", body.Error.Code)
	assert.Equal(t, "AI provider is unavailable.", body.Error.Message)
}

func TestErrorStatus(t *testing.T) {
	type sample struct {
		URL string `validate:"http_url"`
	}
	valErr := validator.New().Struct(sample{URL: "bad"})
	require.Error(t, valErr)

	assert.Equal(t, http.StatusBadRequest, httpx.ErrorStatus(httpx.BadRequest("bad")))
	assert.Equal(t, http.StatusNotFound, httpx.ErrorStatus(errx.NotFound("NOT_FOUND", "missing")))
	assert.Equal(t, http.StatusServiceUnavailable, httpx.ErrorStatus(errx.Unavailable("UNAVAILABLE", "down")))
	assert.Equal(t, http.StatusUnprocessableEntity, httpx.ErrorStatus(valErr))
	assert.Equal(t, http.StatusForbidden, httpx.ErrorStatus(fiber.ErrForbidden))
	assert.Equal(t, http.StatusInternalServerError, httpx.ErrorStatus(errors.New("boom")))
}
