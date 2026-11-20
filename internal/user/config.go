package user

import (
	"context"
	"time"

	"github.com/sethvargo/go-envconfig/v2"
)

// Config holds configuration for the user and authentication feature.
type Config struct {
	SuperAdminEmail string        `env:"USER_SUPER_ADMIN_EMAIL"`
	SuperAdminName  string        `env:"USER_SUPER_ADMIN_NAME, default=Super Admin"`
	OTPTTL          time.Duration `env:"AUTH_OTP_TTL, default=5m"`
	OTPMaxAttempts  int           `env:"AUTH_OTP_MAX_ATTEMPTS, default=5"`
	OTPCooldown     time.Duration `env:"AUTH_OTP_COOLDOWN, default=60s"`
	SessionTTL      time.Duration `env:"AUTH_SESSION_TTL, default=720h"`
	CookieName      string        `env:"AUTH_SESSION_COOKIE_NAME, default=domus_session"`
	CookieSecure    bool          `env:"AUTH_SESSION_COOKIE_SECURE, default=false"`
	GoogleID        string        `env:"AUTH_GOOGLE_ID"`
	GoogleSecret    string        `env:"AUTH_GOOGLE_SECRET"`
	GitHubID        string        `env:"AUTH_GITHUB_ID"`
	GitHubSecret    string        `env:"AUTH_GITHUB_SECRET"`
}

// GetConfig loads user feature configuration from environment variables.
func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
