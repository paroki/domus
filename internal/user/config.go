package user

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig/v2"
)

// Config holds configuration for the user and authentication feature.
type Config struct {
	SuperAdminEmail string `env:"USER_SUPER_ADMIN_EMAIL"`
	SuperAdminName  string `env:"USER_SUPER_ADMIN_NAME, default=Super Admin"`
	GoogleID        string `env:"AUTH_GOOGLE_ID"`
	GoogleSecret    string `env:"AUTH_GOOGLE_SECRET"`
	GitHubID        string `env:"AUTH_GITHUB_ID"`
	GitHubSecret    string `env:"AUTH_GITHUB_SECRET"`

	// ponytail: constant auth defaults kept in code (not env vars); promote to env if per-deployment tuning is needed.
	OTPTTL         time.Duration
	OTPMaxAttempts int
	OTPCooldown    time.Duration
	SessionTTL     time.Duration
	CookieName     string
}

// GetConfig loads user feature configuration from environment variables.
func GetConfig(ctx context.Context) (Config, error) {
	cfg := Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
		OTPCooldown:    60 * time.Second,
		SessionTTL:     720 * time.Hour,
		CookieName:     "domus_session",
	}
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
