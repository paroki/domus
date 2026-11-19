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
	DatabaseURL     string        `env:"DATABASE_URL, default=postgres://domus:domus@localhost:5432/domus?sslmode=disable"`
	DBMaxOpenConns  int           `env:"DB_MAX_OPEN_CONNS, default=25"`
	DBMaxIdleConns  int           `env:"DB_MAX_IDLE_CONNS, default=5"`
	AutoMigrate     bool          `env:"AUTO_MIGRATE, default=true"`
	TrustProxy      bool          `env:"TRUST_PROXY, default=false"`
}

func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
