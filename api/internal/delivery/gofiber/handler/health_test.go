package handler_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber/handler"
	"github.com/paroki/domus/api/internal/delivery/gofiber/response"
)

func TestHealthHandler_Check(t *testing.T) {
	// Create mock config
	cfg := &config.Config{
		Env: "test",
	}

	// Create fiber app for testing
	app := fiber.New()
	log := config.GetLogger(cfg)
	h := handler.NewHealthHandler(cfg, log)
	app.Get("/health", h.Check)

	// Create test request
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Run test
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to run app test: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var envelope response.Envelope[handler.HealthResponse]
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v. Body was: %s", err, string(body))
	}

	// Assertions
	if !envelope.Success {
		t.Error("expected success to be true")
	}

	if envelope.Error != nil {
		t.Errorf("expected error to be nil, got: %+v", envelope.Error)
	}

	if envelope.Data.Status != "pass" {
		t.Errorf("expected status 'pass', got: %s", envelope.Data.Status)
	}

	if envelope.Data.Version != handler.Version {
		t.Errorf("expected version '%s', got: %s", handler.Version, envelope.Data.Version)
	}

	if envelope.Data.Env != "test" {
		t.Errorf("expected env 'test', got: %s", envelope.Data.Env)
	}

	if envelope.Data.Uptime < 0 {
		t.Errorf("expected uptime >= 0, got: %f", envelope.Data.Uptime)
	}
}

