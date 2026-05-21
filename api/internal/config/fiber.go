package config

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

type ApiConfig struct {
	AppName      string        `mapstructure:"app_name"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	BodyLimit    int           `mapstructure:"body_limit"`
	Prefork      bool          `mapstructure:"prefork"`
	Cors         string        `mapstructure:"cors"`
}

func GetFiber(cfg *Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.Api.AppName,
		ReadTimeout:  cfg.Api.ReadTimeout,
		WriteTimeout: cfg.Api.WriteTimeout,
		IdleTimeout:  cfg.Api.IdleTimeout,
		BodyLimit:    cfg.Api.BodyLimit,
	})

	return app
}
