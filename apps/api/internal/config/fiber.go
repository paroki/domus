package config

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func initMiddlewares(app *fiber.App) {
	// init healthcheck
	// Use the default probe on the conventional endpoints
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New(healthcheck.Config{
		ResponseFormat: healthcheck.FormatJSON,
	}))
}

func GetFiber() *fiber.App {
	app := fiber.New()

	initMiddlewares(app)
	return app
}
