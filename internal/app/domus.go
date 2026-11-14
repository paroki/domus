package app

import (
	"context"
	"log/slog"
	"net"

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
	addr := net.JoinHostPort(d.cfg.Host, d.cfg.Port)
	d.logger.InfoContext(ctx, "starting server", "addr", addr)
	return d.fiber.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: true,
		GracefulContext:       ctx,
		ShutdownTimeout:       d.cfg.ShutdownTimeout,
	})
}
