package app

import (
	"errors"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/paroki/domus/ui"
)

// GetFiber initializes the HTTP server. ponytail: inline slog middleware and ErrorHandler; extract to internal/platform/httpx when multiple Fiber apps or custom error codes added.
func GetFiber(cfg Config, logger *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		BodyLimit:    cfg.BodyLimit,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := errors.AsType[*fiber.Error](err); ok {
				code = e.Code
			}
			logger.ErrorContext(c.Context(), "request error",
				"request_id", requestid.FromContext(c),
				"method", c.Method(),
				"path", c.Path(),
				"status", code,
				"err", err,
			)
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		logger.InfoContext(c.Context(), "http request",
			"request_id", requestid.FromContext(c),
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.IP(),
		)
		return err
	})

	api := app.Group("/api")
	api.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	// ponytail: serves embedded SPA without custom Cache-Control headers; add immutable asset caching when CDN/traffic requires.
	if distFS, err := ui.DistFS(); err == nil {
		app.Use(static.New("", static.Config{
			FS: distFS,
			Next: func(c fiber.Ctx) bool {
				return strings.HasPrefix(c.Path(), "/api")
			},
			NotFoundHandler: func(c fiber.Ctx) error {
				index, err := fs.ReadFile(distFS, "index.html")
				if err != nil {
					return fiber.ErrNotFound
				}
				c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
				return c.Send(index)
			},
		}))
	}

	return app
}
