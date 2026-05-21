package middleware_test

import (
	"testing"

	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber"
	"github.com/paroki/domus/api/testutil"
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
			expectedStatus: 200,
			expectedOrigin: "*",
			expectCreds:    "", // browser limits creds with wildcard
		},
		{
			name:           "Prod mode - empty config blocks cross-origin (no headers)",
			env:            "production",
			corsConfig:     "",
			requestMethod:  "GET",
			requestOrigin:  "http://any-frontend.local",
			expectedStatus: 200,
			expectNoCORS:   true,
		},
		{
			name:           "Configured origins - allowed domain matches",
			env:            "production",
			corsConfig:     "http://localhost:3000, https://domus.paroki.com",
			requestMethod:  "GET",
			requestOrigin:  "http://localhost:3000",
			expectedStatus: 200,
			expectedOrigin: "http://localhost:3000",
			expectCreds:    "true",
		},
		{
			name:           "Configured origins - allowed domain matches second entry",
			env:            "production",
			corsConfig:     "http://localhost:3000, https://domus.paroki.com",
			requestMethod:  "GET",
			requestOrigin:  "https://domus.paroki.com",
			expectedStatus: 200,
			expectedOrigin: "https://domus.paroki.com",
			expectCreds:    "true",
		},
		{
			name:           "Configured origins - disallowed domain gets no CORS headers",
			env:            "production",
			corsConfig:     "http://localhost:3000, https://domus.paroki.com",
			requestMethod:  "GET",
			requestOrigin:  "https://malicious.com",
			expectedStatus: 200,
			expectNoCORS:   true,
		},
		{
			name:           "Preflight request - allowed domain",
			env:            "production",
			corsConfig:     "http://localhost:3000",
			requestMethod:  "OPTIONS",
			requestOrigin:  "http://localhost:3000",
			expectedStatus: 204,
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

			db, closeDB := testutil.NewTestDB(t)
			t.Cleanup(closeDB)

			gofiber.SetupRouter(app, cfg, log, db)

			reqBuilder := testutil.New(app).Method(tt.requestMethod, "/api/health")

			if tt.requestOrigin != "" {
				reqBuilder = reqBuilder.WithHeader("Origin", tt.requestOrigin)
			}
			if tt.requestMethod == "OPTIONS" {
				reqBuilder = reqBuilder.WithHeader("Access-Control-Request-Method", "GET")
			}

			assertions := reqBuilder.Expect(t).Status(tt.expectedStatus)

			if tt.expectNoCORS {
				assertions.Header("Access-Control-Allow-Origin").DoesNotExist()
			} else {
				assertions.Header("Access-Control-Allow-Origin").Eq(tt.expectedOrigin)
				assertions.Header("Access-Control-Allow-Credentials").Eq(tt.expectCreds)
			}
		})
	}
}
