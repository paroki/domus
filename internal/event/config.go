package event

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig/v2"
)

// Config holds configuration for the event feature.
type Config struct {
	DefaultVisibility string        `env:"EVENT_DEFAULT_VISIBILITY, default=public"`
	RSVPGracePeriod   time.Duration `env:"EVENT_RSVP_GRACE_PERIOD, default=30m"`
}

// GetConfig loads event feature configuration from environment variables.
func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
