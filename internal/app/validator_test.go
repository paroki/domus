package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type validatorTestPayload struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type validatorTestFeature struct{}

func (validatorTestFeature) Register(api fiber.Router) {
	api.Post("/validate-test", func(c fiber.Ctx) error {
		var body validatorTestPayload
		if err := c.Bind().Body(&body); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusOK)
	})
}

func TestStructValidator(t *testing.T) {
	app := GetFiber(Config{}, slog.New(slog.DiscardHandler), validatorTestFeature{})

	req := httptest.NewRequest(http.MethodPost, "/api/validate-test", strings.NewReader(`{"full_name":"","email":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "full_name") {
		t.Fatalf("expected json tag name 'full_name' in error, got: %s", body)
	}

	okReq := httptest.NewRequest(http.MethodPost, "/api/validate-test", strings.NewReader(`{"full_name":"Toni","email":"toni@example.com"}`))
	okReq.Header.Set("Content-Type", "application/json")
	okResp, err := app.Test(okReq)
	if err != nil {
		t.Fatal(err)
	}
	if okResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", okResp.StatusCode)
	}
}
