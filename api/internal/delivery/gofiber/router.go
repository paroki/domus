package gofiber

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/delivery/gofiber/handler"
	"github.com/paroki/domus/api/internal/delivery/gofiber/middleware"
	"github.com/paroki/domus/api/internal/shared/logger"
)

func SetupRouter(app *fiber.App, cfg *config.Config, log logger.Logger, db *ent.Client) {
	// Global middleware
	middleware.Setup(app, cfg)

	// API Group
	api := app.Group("/api")

	// Handlers
	healthHandler := handler.NewHealthHandler(cfg, log)

	// Routes
	api.Get("/health", healthHandler.Check)
}
