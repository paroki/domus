package event

import (
	"context"
	"time"
)

// Config holds configuration for the event feature.
type Config struct {
	// ponytail: fixed event defaults; add env tags when per-deployment tuning needed.
	DefaultVisibility string
	RSVPGracePeriod   time.Duration
}

// GetConfig returns the default event feature configuration.
func GetConfig(_ context.Context) (Config, error) {
	return Config{
		DefaultVisibility: "public",
		RSVPGracePeriod:   30 * time.Minute,
	}, nil
}
