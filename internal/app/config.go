package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	AppName         string        `env:"APP_NAME, default=domus"`
	Host            string        `env:"HOST, default=0.0.0.0"`
	Port            string        `env:"PORT, default=8080"`
	LogLevel        slog.Level    `env:"LOG_LEVEL, default=info"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT, default=10s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT, default=10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT, default=60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT, default=10s"`
	BodyLimit       int           `env:"BODY_LIMIT, default=4194304"`
	DatabaseURL     string        `env:"APP_DB_URL, default=postgres://domus:domus@localhost:5432/domus?sslmode=disable"`
	AutoMigrate     bool          `env:"AUTO_MIGRATE, default=true"`
	TrustProxy      bool          `env:"TRUST_PROXY, default=false"`

	// Sentry is disabled when APP_SENTRY_DSN is empty.
	SentryDSN              string  `env:"APP_SENTRY_DSN"`
	SentryEnvironment      string  `env:"APP_SENTRY_ENVIRONMENT, default=development"`
	SentryRelease          string  `env:"APP_SENTRY_RELEASE"`
	SentryTracesSampleRate float64 `env:"APP_SENTRY_TRACES_SAMPLE_RATE, default=0.1"`
}

func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
