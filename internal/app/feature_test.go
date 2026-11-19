package app

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type stubFeature struct{ prefix string }

func (s stubFeature) Prefix() string { return s.prefix }

func (s stubFeature) Register(r fiber.Router) {
	r.Get("/ping", func(c fiber.Ctx) error { return c.SendString(s.prefix) })
}

func TestMountFeatures(t *testing.T) {
	tests := []struct {
		name     string
		features []Feature
		wantErr  bool
	}{
		{"unique prefixes", []Feature{stubFeature{"/a"}, stubFeature{"/b"}}, false},
		{"duplicate prefix", []Feature{stubFeature{"/a"}, stubFeature{"/a"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mountFeatures(fiber.New().Group("/api"), tt.features...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("mountFeatures() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMountFeaturesRoutesUnderPrefix(t *testing.T) {
	app := fiber.New()
	if err := mountFeatures(app.Group("/api"), stubFeature{"/parishioners"}); err != nil {
		t.Fatal(err)
	}
	resp, err := app.Test(httptest.NewRequest("GET", "/api/parishioners/ping", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 || string(body) != "/parishioners" {
		t.Fatalf("got %d %q", resp.StatusCode, body)
	}
}
