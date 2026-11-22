package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSPACachingAndFallback(t *testing.T) {
	distFS := fstest.MapFS{
		"index.html":        {Data: []byte("<!doctype html><html><body>domus</body></html>")},
		"assets/app-123.js": {Data: []byte("console.log('ok');")},
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))),
	})
	mountSPA(app, distFS)

	t.Run("root index.html is not cached", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), "text/html")
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get(fiber.HeaderCacheControl))
	})

	t.Run("SPA route fallback returns 200 with no-cache", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/login", nil))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, resp.Header.Get(fiber.HeaderContentType), "text/html")
		assert.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get(fiber.HeaderCacheControl))
	})

	t.Run("missing asset returns 404 with no-store instead of index.html", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/assets/nonexistent-12345.js", nil))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		assert.NotContains(t, resp.Header.Get(fiber.HeaderContentType), "text/html")
		assert.Equal(t, "no-store", resp.Header.Get(fiber.HeaderCacheControl))
	})

	t.Run("existing hashed asset is cached as immutable", func(t *testing.T) {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/assets/app-123.js", nil))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, "public, max-age=31536000, immutable", resp.Header.Get(fiber.HeaderCacheControl))
	})
}
