package app

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type stubFeature struct {
	prefixes []string
}

func (s stubFeature) Register(api fiber.Router) {
	for _, p := range s.prefixes {
		prefix := p
		api.Group(prefix).Get("/ping", func(c fiber.Ctx) error {
			return c.SendString(prefix)
		})
	}
}

func TestMountFeatures(t *testing.T) {
	app := fiber.New()
	mountFeatures(
		app.Group("/api"),
		stubFeature{prefixes: []string{"/auth", "/users"}},
		stubFeature{prefixes: []string{"/parishioners"}},
	)

	for _, prefix := range []string{"/auth", "/users", "/parishioners"} {
		resp, err := app.Test(httptest.NewRequest("GET", "/api"+prefix+"/ping", nil))
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 200 || string(body) != prefix {
			t.Fatalf("GET /api%s/ping: got %d %q", prefix, resp.StatusCode, body)
		}
	}
}
