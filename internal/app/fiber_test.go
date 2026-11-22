package app

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSPACachingAndFallback(t *testing.T) {
	distFS, err := ui.DistFS()
	require.NoError(t, err)

	var firstAsset string
	entries, err := fs.ReadDir(distFS, "assets")
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".js") {
				firstAsset = "/assets/" + e.Name()
				break
			}
		}
	}

	app := GetFiber(Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

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

	if firstAsset != "" {
		t.Run("existing hashed asset is cached as immutable", func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(http.MethodGet, firstAsset, nil))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.Equal(t, "public, max-age=31536000, immutable", resp.Header.Get(fiber.HeaderCacheControl))
		})
	}
}
