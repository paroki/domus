package org

import (
	"context"
	"time"
)

// Config holds configuration for the org feature.
type Config struct {
	// ponytail: fixed invite TTL defaults; add env tags when per-deployment tuning needed.
	InviteDefaultTTL time.Duration
	InviteMaxTTL     time.Duration
}

// GetConfig returns the default org feature configuration.
func GetConfig(_ context.Context) (Config, error) {
	return Config{
		InviteDefaultTTL: 24 * time.Hour,
		InviteMaxTTL:     168 * time.Hour,
	}, nil
}
