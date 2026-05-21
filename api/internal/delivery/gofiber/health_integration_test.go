package gofiber_test

import (
	"testing"

	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber"
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

	// Run test using fluent builder
	testutil.New(app).
		GET("/api/health").
		Expect(t).
		StatusOK().
		JSONPath("$.success").EqBool(true).
		JSONPath("$.error").IsNull().
		JSONPath("$.data.status").EqString("pass").
		JSONPath("$.data.env").EqString("integration-test").
		JSONPath("$.meta.request_id").Exists()
}
