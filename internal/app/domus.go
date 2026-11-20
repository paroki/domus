package app

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Domus struct {
	cfg    Config
	fiber  *fiber.App
	logger *slog.Logger
}

func New(cfg Config, fiber *fiber.App, logger *slog.Logger) *Domus {
	return &Domus{
		cfg,
		fiber,
		logger,
	}
}

func (d *Domus) Serve(ctx context.Context) error {
	// ponytail: bind all interfaces and fixed 10s shutdown timeout; add APP_HOST/APP_SHUTDOWN_TIMEOUT to Config when needed.
	addr := net.JoinHostPort("0.0.0.0", d.cfg.Port)
	d.logger.InfoContext(ctx, "starting server", "addr", addr)
	return d.fiber.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: true,
		GracefulContext:       ctx,
		ShutdownTimeout:       10 * time.Second,
	})
}
