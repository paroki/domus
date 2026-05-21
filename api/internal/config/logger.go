package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/paroki/domus/api/internal/shared/logger"
)

// LoggerConfig defines the configuration for the logging system.
type LoggerConfig struct {
	Level   string `mapstructure:"level"`
	Adapter string `mapstructure:"adapter"`
}

// GetLogger returns the default Logger instance based on the application configuration.
func GetLogger(cfg *Config) logger.Logger {
	var level slog.Level

	switch strings.ToLower(cfg.Log.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		if cfg.Env == "development" {
			level = slog.LevelDebug
		} else {
			level = slog.LevelInfo
		}
	}

	var handler slog.Handler
	if cfg.Env == "development" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	// Default adapter is slog. If multiple adapters are supported in the future,
	// we can add a factory switch here based on cfg.Log.Adapter.
	l := slog.New(handler)
	return logger.NewSlogLogger(l)
}
