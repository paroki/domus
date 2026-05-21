package gofiber_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber"
	"github.com/paroki/domus/api/internal/delivery/gofiber/handler"
	"github.com/paroki/domus/api/internal/delivery/gofiber/response"
	"github.com/paroki/domus/api/testutil"
)

func TestHealth_Integration(t *testing.T) {
	// Create mock config
	cfg := &config.Config{
		Env: "integration-test",
		Api: config.ApiConfig{
			AppName: "Domus API Test",
		},
	}

	// Create and bootstrap fiber app
	app := config.GetFiber(cfg)
	log := config.GetLogger(cfg)

	db, closeDB := testutil.NewTestDB(t)
	t.Cleanup(closeDB)

	gofiber.SetupRouter(app, cfg, log, db)

	// Create request
	req, err := http.NewRequest("GET", "/api/health", nil)
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

	if envelope.Data.Env != "integration-test" {
		t.Errorf("expected env 'integration-test', got: %s", envelope.Data.Env)
	}

	if envelope.Meta == nil {
		t.Fatal("expected meta to not be nil")
	}

	if envelope.Meta.RequestID == "" {
		t.Error("expected request_id in meta to be populated by requestid middleware")
	}
}
