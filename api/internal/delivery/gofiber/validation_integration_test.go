package gofiber_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber/middleware"
	"github.com/paroki/domus/api/internal/delivery/gofiber/response"
	"github.com/paroki/domus/api/internal/shared/validator"
)

// -- Test input struct --------------------------------------------------------

type createMemberInput struct {
	Name  string `json:"name"  validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age"   validate:"gte=17"`
}

// -- Setup helpers ------------------------------------------------------------

func newTestApp() *fiber.App {
	cfg := &config.Config{
		Env: "integration-test",
		Api: config.ApiConfig{AppName: "Domus API Test"},
	}
	app := config.GetFiber(cfg)
	v := config.NewValidator()

	// Register global middleware (includes requestid) to match production bootstrap.
	middleware.Setup(app, cfg)

	// Register a minimal POST /api/test/members route for validation testing.
	app.Post("/api/test/members", func(c fiber.Ctx) error {
		var input createMemberInput
		if err := c.Bind().JSON(&input); err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "DOMUS-VAL-000", "Failed to parse request body.", nil)
		}

		errs := v.Validate(input)
		if len(errs) > 0 {
			return response.ValidationFail(c, errs)
		}

		return response.Created(c, input)
	})

	return app
}

// -- Tests --------------------------------------------------------------------

func TestValidationIntegration_ValidPayload_Returns201(t *testing.T) {
	app := newTestApp()

	payload := `{"name":"Anthonius","email":"toni@example.com","age":30}`
	req, _ := http.NewRequest(http.MethodPost, "/api/test/members", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var envelope response.Envelope[createMemberInput]
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("failed to unmarshal: %v\nBody: %s", err, body)
	}

	if !envelope.Success {
		t.Error("expected success=true")
	}
	if envelope.Error != nil {
		t.Errorf("expected error=nil, got: %+v", envelope.Error)
	}
	if envelope.Meta == nil {
		t.Fatal("expected meta to be present")
	}
	if envelope.Meta.RequestID == "" {
		t.Error("expected request_id in meta")
	}
}

func TestValidationIntegration_InvalidPayload_Returns400WithDetails(t *testing.T) {
	app := newTestApp()

	// Missing name, invalid email, age below minimum.
	payload := `{"name":"A","email":"not-an-email","age":15}`
	req, _ := http.NewRequest(http.MethodPost, "/api/test/members", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var envelope response.Envelope[any]
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("failed to unmarshal: %v\nBody: %s", err, body)
	}

	if envelope.Success {
		t.Error("expected success=false")
	}
	if envelope.Data != nil {
		t.Errorf("expected data=null, got: %+v", envelope.Data)
	}
	if envelope.Error == nil {
		t.Fatal("expected error to be present")
	}
	if envelope.Error.Code != "DOMUS-VAL-001" {
		t.Errorf("expected code=DOMUS-VAL-001, got=%q", envelope.Error.Code)
	}
	if envelope.Error.Message == "" {
		t.Error("expected non-empty error message")
	}
	if len(envelope.Error.Details) == 0 {
		t.Fatal("expected validation details to be non-empty")
	}

	// Verify at least one detail has field + issue populated.
	for _, d := range envelope.Error.Details {
		if d.Field == "" {
			t.Error("detail.field must not be empty")
		}
		if d.Issue == "" {
			t.Error("detail.issue must not be empty")
		}
	}

	// Meta must always be present per ADR-001.
	if envelope.Meta == nil {
		t.Fatal("expected meta to be present even on error responses")
	}
	if envelope.Meta.RequestID == "" {
		t.Error("expected request_id in meta")
	}
}

func TestValidationIntegration_MissingRequiredFields_Returns400(t *testing.T) {
	app := newTestApp()

	payload := `{}`
	req, _ := http.NewRequest(http.MethodPost, "/api/test/members", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var envelope response.Envelope[any]
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("failed to unmarshal: %v\nBody: %s", err, body)
	}

	if envelope.Error == nil {
		t.Fatal("expected error to be present")
	}
	if envelope.Error.Code != "DOMUS-VAL-001" {
		t.Errorf("expected DOMUS-VAL-001, got %q", envelope.Error.Code)
	}

	// name and email are required — should appear in details.
	fields := make(map[string]bool)
	for _, d := range envelope.Error.Details {
		fields[d.Field] = true
	}
	if !fields["name"] {
		t.Error("expected 'name' field in validation details")
	}
	if !fields["email"] {
		t.Error("expected 'email' field in validation details")
	}
}

// -- Compile-time check: V10Validator satisfies Validator interface -----------

var _ validator.Validator = (*validator.V10Validator)(nil)
