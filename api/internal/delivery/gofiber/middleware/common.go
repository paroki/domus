package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/paroki/domus/api/internal/config"
	"github.com/paroki/domus/api/internal/shared/logger"
)

func Setup(app *fiber.App, cfg *config.Config) {
	app.Use(requestid.New())
	app.Use(func(c fiber.Ctx) error {
		requestID := requestid.FromContext(c)
		if requestID != "" {
			ctx := logger.ContextWithRequestID(c.Context(), requestID)
			c.SetContext(ctx)
		}
		return c.Next()
	})
	app.Use(recover.New())
	app.Use(NewCORS(cfg))
}
