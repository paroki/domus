package gofiber

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber/handler"
	"github.com/paroki/domus/api/internal/delivery/gofiber/middleware"
)

func SetupRouter(app *fiber.App, cfg *config.Config) {
	// Global middleware
	middleware.Setup(app)

	// API Group
	api := app.Group("/api")

	// Handlers
	healthHandler := handler.NewHealthHandler(cfg)

	// Routes
	api.Get("/health", healthHandler.Check)
}
