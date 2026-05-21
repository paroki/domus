package middleware_test

import (
	"net/http"
	"testing"

	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber"
)

func TestCORS_Integration(t *testing.T) {
	tests := []struct {
		name           string
		env            string
		corsConfig     string
		requestMethod  string
		requestOrigin  string
		expectedStatus int
		expectedOrigin string
		expectCreds    string
		expectNoCORS   bool
	}{
		{
			name:           "Dev mode - fallback to wildcard",
			env:            "development",
			corsConfig:     "",
			requestMethod:  "GET",
			requestOrigin:  "http://any-frontend.local",
			expectedStatus: http.StatusOK,
			expectedOrigin: "*",
			expectCreds:    "", // browser limits creds with wildcard
		},
		{
			name:           "Prod mode - empty config blocks cross-origin (no headers)",
			env:            "production",
			corsConfig:     "",
			requestMethod:  "GET",
			requestOrigin:  "http://any-frontend.local",
			expectedStatus: http.StatusOK,
			expectNoCORS:   true,
		},
		{
			name:           "Configured origins - allowed domain matches",
			env:            "production",
			corsConfig:     "http://localhost:3000, https://domus.paroki.com",
			requestMethod:  "GET",
			requestOrigin:  "http://localhost:3000",
			expectedStatus: http.StatusOK,
			expectedOrigin: "http://localhost:3000",
			expectCreds:    "true",
		},
		{
			name:           "Configured origins - allowed domain matches second entry",
			env:            "production",
			corsConfig:     "http://localhost:3000, https://domus.paroki.com",
			requestMethod:  "GET",
			requestOrigin:  "https://domus.paroki.com",
			expectedStatus: http.StatusOK,
			expectedOrigin: "https://domus.paroki.com",
			expectCreds:    "true",
		},
		{
			name:           "Configured origins - disallowed domain gets no CORS headers",
			env:            "production",
			corsConfig:     "http://localhost:3000, https://domus.paroki.com",
			requestMethod:  "GET",
			requestOrigin:  "https://malicious.com",
			expectedStatus: http.StatusOK,
			expectNoCORS:   true,
		},
		{
			name:           "Preflight request - allowed domain",
			env:            "production",
			corsConfig:     "http://localhost:3000",
			requestMethod:  "OPTIONS",
			requestOrigin:  "http://localhost:3000",
			expectedStatus: http.StatusNoContent,
			expectedOrigin: "http://localhost:3000",
			expectCreds:    "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Env: tt.env,
				Api: config.ApiConfig{
					AppName: "Domus CORS Test",
					Cors:    tt.corsConfig,
				},
			}

			app := config.GetFiber(cfg)
			log := config.GetLogger(cfg)
			gofiber.SetupRouter(app, cfg, log)

			req, err := http.NewRequest(tt.requestMethod, "/api/health", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}
			if tt.requestMethod == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "GET")
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to run app test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			originHeader := resp.Header.Get("Access-Control-Allow-Origin")
			credsHeader := resp.Header.Get("Access-Control-Allow-Credentials")

			if tt.expectNoCORS {
				if originHeader != "" {
					t.Errorf("expected no CORS headers, but got Access-Control-Allow-Origin: %s", originHeader)
				}
			} else {
				if originHeader != tt.expectedOrigin {
					t.Errorf("expected Access-Control-Allow-Origin %q, got %q", tt.expectedOrigin, originHeader)
				}
				if credsHeader != tt.expectCreds {
					t.Errorf("expected Access-Control-Allow-Credentials %q, got %q", tt.expectCreds, credsHeader)
				}
			}
		})
	}
}
