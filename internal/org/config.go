package org

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig/v2"
)

// Config holds configuration for the org feature.
type Config struct {
	InviteDefaultTTL time.Duration `env:"ORG_INVITE_DEFAULT_TTL, default=24h"`
	InviteMaxTTL     time.Duration `env:"ORG_INVITE_MAX_TTL, default=168h"`
}

// GetConfig loads org feature configuration from environment variables.
func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
