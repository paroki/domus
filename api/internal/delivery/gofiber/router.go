package gofiber

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/delivery/gofiber/middleware"
	"github.com/paroki/domus/api/internal/delivery/gofiber/response"
)

func SetupRouter(app *fiber.App) {
	// Global middleware
	middleware.Setup(app)

	// API Group
	api := app.Group("/api")

	// Health check example
	api.Get("/health", func(c fiber.Ctx) error {
		return response.OK(c, "OK")
	})
}
