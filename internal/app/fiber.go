package app

import (
	"io/fs"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/ui"
)

// GetFiber initializes the HTTP server. ponytail: inline slog middleware & hardcoded server defaults; extract or add to Config when needed.
func GetFiber(_ Config, logger *slog.Logger, features ...Feature) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:          "domus",
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     10 * time.Second,
		IdleTimeout:      60 * time.Second,
		BodyLimit:        4 * 1024 * 1024,
		ProxyHeader:      fiber.HeaderXForwardedFor,
		TrustProxy:       true,
		TrustProxyConfig: fiber.TrustProxyConfig{Private: true, Loopback: true},
		StructValidator:  newStructValidator(),
		ErrorHandler:     httpx.ErrorHandler(logger),
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(sentryMiddleware)
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
	api.Get("/health", healthHandler)
	mountFeatures(api, features...)

	mcpArgs := make([]any, 0, len(features))
	var resolver mcpx.PrincipalResolver
	for _, f := range features {
		mcpArgs = append(mcpArgs, f)
		if resolver == nil {
			if pr, ok := f.(interface{ PrincipalResolver() mcpx.PrincipalResolver }); ok {
				resolver = pr.PrincipalResolver()
			}
		}
	}
	mcpServer := mcpx.NewServer("domus", mcpArgs...)
	for _, f := range features {
		if binder, ok := f.(interface{ SetMCPServer(*mcp.Server) }); ok {
			binder.SetMCPServer(mcpServer)
		}
	}
	mcpHandler := mcpx.NewHandler(mcpServer, resolver)
	app.All("/mcp", mcpHandler)
	app.All("/mcp/*", mcpHandler)

	if distFS, err := ui.DistFS(); err == nil {
		mountSPA(app, distFS)
	}

	return app
}

func mountSPA(app *fiber.App, distFS fs.FS) {
	isBypass := func(p string) bool {
		return strings.HasPrefix(p, "/api") || p == "/mcp" || strings.HasPrefix(p, "/mcp/")
	}
	app.Use(func(c fiber.Ctx) error {
		p := c.Path()
		if isBypass(p) {
			return c.Next()
		}
		if strings.HasPrefix(p, "/assets/") {
			c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
		} else if p == "/" || p == "/index.html" || path.Ext(p) == "" {
			c.Set(fiber.HeaderCacheControl, "no-cache, no-store, must-revalidate")
		}
		return c.Next()
	})
	app.Use(static.New("", static.Config{
		FS: distFS,
		Next: func(c fiber.Ctx) bool {
			return isBypass(c.Path())
		},
		NotFoundHandler: func(c fiber.Ctx) error {
			p := c.Path()
			if strings.HasPrefix(p, "/assets/") || path.Ext(p) != "" {
				c.Set(fiber.HeaderCacheControl, "no-store")
				return fiber.ErrNotFound
			}
			index, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				return fiber.ErrNotFound
			}
			c.Set(fiber.HeaderCacheControl, "no-cache, no-store, must-revalidate")
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
			return c.Status(fiber.StatusOK).Send(index)
		},
	}))
}

type healthStatus struct {
	Status string `json:"status"`
}

// @Summary Health check
// @ID getHealth
// @Tags system
// @Produce json
// @Success 200 {object} httpx.Envelope[healthStatus]
// @Router /health [get]
func healthHandler(c fiber.Ctx) error {
	return httpx.OK(c, healthStatus{Status: "ok"})
}
