package httpx_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testItem struct {
	ID string `json:"id"`
}

func newTestApp(logBuf *bytes.Buffer) *fiber.App {
	logger := slog.New(slog.NewJSONHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.ErrorHandler(logger),
	})
	app.Use(requestid.New())
	return app
}

func TestResponseHelpers(t *testing.T) {
	var logBuf bytes.Buffer
	app := newTestApp(&logBuf)

	app.Get("/ok", func(c fiber.Ctx) error {
		return httpx.OK(c, testItem{ID: "1"})
	})
	app.Post("/created", func(c fiber.Ctx) error {
		return httpx.Created(c, testItem{ID: "2"})
	})
	app.Get("/list-nil", func(c fiber.Ctx) error {
		var items []testItem
		return httpx.List(c, items, httpx.Meta{Limit: 20, HasMore: false})
	})
	app.Get("/list-total", func(c fiber.Ctx) error {
		total := 2
		return httpx.List(c, []testItem{{ID: "a"}, {ID: "b"}}, httpx.Meta{
			NextCursor: "cur-1",
			HasMore:    true,
			Limit:      2,
			Total:      &total,
		})
	})
	app.Delete("/no-content", func(c fiber.Ctx) error {
		return httpx.NoContent(c)
	})

	// OK
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ok", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	raw, _ := io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"data":{"id":"1"}}`, string(raw))

	// Created
	resp, err = app.Test(httptest.NewRequest(http.MethodPost, "/created", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	raw, _ = io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"data":{"id":"2"}}`, string(raw))

	// List with nil slice serializes [] instead of null and omits empty nextCursor
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/list-nil", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	raw, _ = io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"data":[],"meta":{"hasMore":false,"limit":20}}`, string(raw))

	// List with Total
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/list-total", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	raw, _ = io.ReadAll(resp.Body)
	assert.JSONEq(t, `{"data":[{"id":"a"},{"id":"b"}],"meta":{"nextCursor":"cur-1","hasMore":true,"limit":2,"total":2}}`, string(raw))

	// NoContent
	resp, err = app.Test(httptest.NewRequest(http.MethodDelete, "/no-content", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	raw, _ = io.ReadAll(resp.Body)
	assert.Empty(t, raw)
}

func TestPaginationAndCursor(t *testing.T) {
	var logBuf bytes.Buffer
	app := newTestApp(&logBuf)

	app.Get("/page", func(c fiber.Ctx) error {
		req, err := httpx.ParsePageRequest(c)
		if err != nil {
			return err
		}
		return httpx.OK(c, req)
	})

	// Default limit
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/page", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var okEnv httpx.Envelope[httpx.PageRequest]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&okEnv))
	assert.Equal(t, httpx.DefaultLimit, okEnv.Data.Limit)

	// Custom limit + cursor
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/page?limit=50&cursor=abc", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&okEnv))
	assert.Equal(t, 50, okEnv.Data.Limit)
	assert.Equal(t, "abc", okEnv.Data.Cursor)

	// Out-of-range limits -> 422
	for _, q := range []string{"?limit=-1", "?limit=101", "?limit=bad"} {
		resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/page"+q, nil))
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	}

	// Cursor encode/decode round-trip
	now := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)
	encoded := httpx.EncodeCursor(now, "user-42")
	decoded, err := httpx.DecodeCursor(encoded)
	require.NoError(t, err)
	assert.Equal(t, "user-42", decoded.ID)
	assert.True(t, now.Equal(decoded.CreatedAt))

	// Invalid cursors -> ErrInvalidCursor + ErrBadRequest
	for _, bad := range []string{"not-base64!", "e30", "eyJpZCI6IjEifQ"} {
		_, err := httpx.DecodeCursor(bad)
		require.Error(t, err)
		assert.ErrorIs(t, err, httpx.ErrInvalidCursor)
		assert.ErrorIs(t, err, httpx.ErrBadRequest)
	}
}

func TestAppErrorsAndHandler(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		sentinel error
		status   int
		code     string
	}{
		{"BadRequest", httpx.BadRequest("bad"), httpx.ErrBadRequest, http.StatusBadRequest, "BAD_REQUEST"},
		{"InvalidCursor", httpx.InvalidCursor("bad cursor"), httpx.ErrInvalidCursor, http.StatusBadRequest, "INVALID_CURSOR"},
		{"Unauthorized", httpx.Unauthorized("unauth"), httpx.ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"Forbidden", httpx.Forbidden("forbid"), httpx.ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{"NotFound", httpx.NotFound("missing"), httpx.ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"Conflict", httpx.Conflict("dup"), httpx.ErrConflict, http.StatusConflict, "CONFLICT"},
		{"Gone", httpx.Gone("gone"), httpx.ErrGone, http.StatusGone, "GONE"},
		{"Validation", httpx.ValidationError(httpx.FieldError{Field: "email", Code: "EMAIL", Message: "invalid"}), httpx.ErrValidation, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"RateLimited", httpx.RateLimited("slow down"), httpx.ErrRateLimited, http.StatusTooManyRequests, "RATE_LIMITED"},
		{"Internal", httpx.InternalError(errors.New("db password leaked")), httpx.ErrInternal, http.StatusInternalServerError, "INTERNAL_ERROR"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ErrorIs(t, tc.err, tc.sentinel)

			var logBuf bytes.Buffer
			app := newTestApp(&logBuf)
			app.Get("/err", func(c fiber.Ctx) error {
				return tc.err
			})

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/err", nil))
			require.NoError(t, err)
			assert.Equal(t, tc.status, resp.StatusCode)

			raw, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var rawMap map[string]any
			require.NoError(t, json.Unmarshal(raw, &rawMap))
			assert.NotContains(t, rawMap, "data")
			assert.NotContains(t, rawMap, "meta")
			assert.Contains(t, rawMap, "error")

			var env httpx.Envelope[any]
			require.NoError(t, json.Unmarshal(raw, &env))
			require.NotNil(t, env.Error)
			assert.Equal(t, tc.code, env.Error.Code)
			assert.NotEmpty(t, env.Error.RequestID)
			if tc.status >= 500 {
				assert.NotContains(t, env.Error.Message, "db password leaked")
				assert.Contains(t, logBuf.String(), `"level":"ERROR"`)
			} else {
				assert.Contains(t, logBuf.String(), `"level":"WARN"`)
			}
		})
	}
}

func TestValidationAndFiberAndUnknownErrors(t *testing.T) {
	var logBuf bytes.Buffer
	app := newTestApp(&logBuf)
	validate := validator.New()

	type sampleInput struct {
		Email string `validate:"required,email"`
	}

	app.Get("/validate", func(c fiber.Ctx) error {
		return validate.Struct(sampleInput{Email: "bad"})
	})
	app.Get("/fiber-404", func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "route missing")
	})
	app.Get("/raw-err", func(c fiber.Ctx) error {
		return errors.New("secret sql error")
	})
	app.Get("/custom-err", func(c fiber.Ctx) error {
		return httpx.NewError(http.StatusTeapot, "TEAPOT", "short and stout")
	})

	// validator.ValidationErrors
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/validate", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var env httpx.Envelope[any]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	require.NotNil(t, env.Error)
	assert.Equal(t, "VALIDATION_ERROR", env.Error.Code)
	require.Len(t, env.Error.Details, 1)
	assert.Equal(t, "email", env.Error.Details[0].Field)
	assert.Equal(t, "EMAIL", env.Error.Details[0].Code)

	// *fiber.Error
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/fiber-404", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Equal(t, "NOT_FOUND", env.Error.Code)

	// raw error -> 500 without leaking message
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/raw-err", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	raw, _ := io.ReadAll(resp.Body)
	assert.False(t, strings.Contains(string(raw), "secret sql error"))

	// NewError
	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/custom-err", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	assert.Equal(t, "TEAPOT", env.Error.Code)
}

func TestErrorHandler_Errx(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.ErrorHandler(slog.New(slog.DiscardHandler)),
	})
	app.Get("/not-found", func(c fiber.Ctx) error {
		return errx.NotFound("USER_NOT_FOUND", "User not found.")
	})
	app.Get("/validation", func(c fiber.Ctx) error {
		return errx.ValidationError(errx.FieldError{
			Field:   "phone",
			Code:    "INVALID_PHONE",
			Message: "Phone must be E.164.",
		})
	})
	app.Get("/internal", func(c fiber.Ctx) error {
		return errx.Internal("DB_FAILURE", "secret sql error", errors.New("sql: conn refused"))
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/not-found", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	var env httpx.Envelope[any]
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	require.NotNil(t, env.Error)
	assert.Equal(t, "USER_NOT_FOUND", env.Error.Code)
	assert.Equal(t, "User not found.", env.Error.Message)

	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/validation", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	require.NotNil(t, env.Error)
	assert.Equal(t, "VALIDATION_ERROR", env.Error.Code)
	require.Len(t, env.Error.Details, 1)
	assert.Equal(t, "phone", env.Error.Details[0].Field)
	assert.Equal(t, "INVALID_PHONE", env.Error.Details[0].Code)

	resp, err = app.Test(httptest.NewRequest(http.MethodGet, "/internal", nil))
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	raw, _ := io.ReadAll(resp.Body)
	assert.False(t, strings.Contains(string(raw), "secret sql error"))
}
