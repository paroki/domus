package config_test

import (
	"testing"

	"github.com/paroki/domus/api/internal/config"
)

func TestGetLogger(t *testing.T) {
	t.Run("development environment fallback", func(t *testing.T) {
		cfg := &config.Config{
			Env: "development",
		}
		l := config.GetLogger(cfg)
		if l == nil {
			t.Error("expected logger to be non-nil")
		}
	})

	t.Run("production environment fallback", func(t *testing.T) {
		cfg := &config.Config{
			Env: "production",
		}
		l := config.GetLogger(cfg)
		if l == nil {
			t.Error("expected logger to be non-nil")
		}
	})

	t.Run("configured log level debug", func(t *testing.T) {
		cfg := &config.Config{
			Env: "production",
			Log: config.LoggerConfig{
				Level:   "debug",
				Adapter: "slog",
			},
		}
		l := config.GetLogger(cfg)
		if l == nil {
			t.Error("expected logger to be non-nil")
		}
	})

	t.Run("configured log level error", func(t *testing.T) {
		cfg := &config.Config{
			Env: "production",
			Log: config.LoggerConfig{
				Level:   "error",
				Adapter: "slog",
			},
		}
		l := config.GetLogger(cfg)
		if l == nil {
			t.Error("expected logger to be non-nil")
		}
	})
}
