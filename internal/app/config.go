package app

import (
	"context"
	"log/slog"

	"github.com/sethvargo/go-envconfig/v2"
)

type Config struct {
	Port        string     `env:"APP_PORT, default=8080"`
	LogLevel    slog.Level `env:"APP_LOG_LEVEL, default=info"`
	DatabaseURL string     `env:"APP_DB_URL, default=postgres://domus:domus@localhost:5432/domus?sslmode=disable"`

	// Sentry is disabled when APP_SENTRY_DSN is empty.
	SentryDSN         string `env:"APP_SENTRY_DSN"`
	SentryEnvironment string `env:"APP_SENTRY_ENVIRONMENT, default=development"`
	SentryRelease     string `env:"APP_SENTRY_RELEASE"`

	// SMTP is disabled when APP_SMTP_HOST is empty (falls back to LogOTPSender).
	SMTPHost string `env:"APP_SMTP_HOST"`
	SMTPPort int    `env:"APP_SMTP_PORT, default=587"`
	SMTPUser string `env:"APP_SMTP_USER"`
	SMTPPass string `env:"APP_SMTP_PASS"`
	SMTPFrom string `env:"APP_SMTP_FROM, default=Domus <no-reply@example.com>"`
}

func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
