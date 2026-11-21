package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

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

	// Optional messaging channels for OTP delivery and bot webhooks.
	// ponytail: APP_GOWA_DEVICE_ID and APP_GOWA_WEBHOOK_SECRET deferred to Phase 6 (MVP-6.4).
	TelegramBotToken      string `env:"APP_TELEGRAM_BOT_TOKEN"`
	TelegramBotUsername   string `env:"APP_TELEGRAM_BOT_USERNAME"`
	TelegramWebhookSecret string `env:"APP_TELEGRAM_WEBHOOK_SECRET"`
	GowaBaseURL           string `env:"APP_GOWA_BASE_URL"`
	GowaBasicUser         string `env:"APP_GOWA_BASIC_USER"`
	GowaBasicPass         string `env:"APP_GOWA_BASIC_PASS"`
}

// StartupOAuthConfig carries OAuth2 credentials checked during startup env validation.
type StartupOAuthConfig struct {
	GoogleID     string
	GoogleSecret string
	GitHubID     string
	GitHubSecret string
}

// ValidateStartupEnv validates optional integration environment variable groups,
// failing fast on partial configurations and logging warnings when fully disabled.
func ValidateStartupEnv(cfg Config, oauth StartupOAuthConfig, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	var errs []error

	// Telegram requires all 3 vars when any is set.
	var tgMissing []string
	if strings.TrimSpace(cfg.TelegramBotToken) == "" {
		tgMissing = append(tgMissing, "APP_TELEGRAM_BOT_TOKEN")
	}
	if strings.TrimSpace(cfg.TelegramBotUsername) == "" {
		tgMissing = append(tgMissing, "APP_TELEGRAM_BOT_USERNAME")
	}
	if strings.TrimSpace(cfg.TelegramWebhookSecret) == "" {
		tgMissing = append(tgMissing, "APP_TELEGRAM_WEBHOOK_SECRET")
	}
	switch len(tgMissing) {
	case 0:
	case 3:
		logger.Warn("optional integration disabled",
			"integration", "telegram",
			"missing_env", []string{"APP_TELEGRAM_BOT_TOKEN", "APP_TELEGRAM_BOT_USERNAME", "APP_TELEGRAM_WEBHOOK_SECRET"},
		)
	default:
		errs = append(errs, fmt.Errorf("partial Telegram configuration: missing %s", strings.Join(tgMissing, ", ")))
	}

	// GOWA allows base URL alone, but basic auth requires user+pass+base URL.
	errs = append(errs, validateHostWithOptionalAuth(
		"gowa",
		cfg.GowaBaseURL, "APP_GOWA_BASE_URL",
		cfg.GowaBasicUser, "APP_GOWA_BASIC_USER",
		cfg.GowaBasicPass, "APP_GOWA_BASIC_PASS",
		"optional integration disabled",
		logger,
	)...)

	// SMTP allows host alone, but auth credentials require user+pass+host.
	errs = append(errs, validateHostWithOptionalAuth(
		"smtp",
		cfg.SMTPHost, "APP_SMTP_HOST",
		cfg.SMTPUser, "APP_SMTP_USER",
		cfg.SMTPPass, "APP_SMTP_PASS",
		"optional integration disabled; using LogOTPSender fallback",
		logger,
	)...)

	// OAuth2 pairs require both client ID and secret when either is set.
	errs = append(errs, validateOAuthPair("oauth_google", oauth.GoogleID, "AUTH_GOOGLE_ID", oauth.GoogleSecret, "AUTH_GOOGLE_SECRET", logger)...)
	errs = append(errs, validateOAuthPair("oauth_github", oauth.GitHubID, "AUTH_GITHUB_ID", oauth.GitHubSecret, "AUTH_GITHUB_SECRET", logger)...)

	return errors.Join(errs...)
}

func validateHostWithOptionalAuth(
	integration, hostVal, hostEnv, userVal, userEnv, passVal, passEnv, disabledMsg string,
	logger *slog.Logger,
) []error {
	hasHost := strings.TrimSpace(hostVal) != ""
	hasUser := strings.TrimSpace(userVal) != ""
	hasPass := strings.TrimSpace(passVal) != ""

	var errs []error
	if hasUser != hasPass {
		errs = append(errs, fmt.Errorf("partial %s configuration: both %s and %s must be set together", integration, userEnv, passEnv))
	}
	if (hasUser || hasPass) && !hasHost {
		errs = append(errs, fmt.Errorf("partial %s configuration: %s is required when %s or %s is set", integration, hostEnv, userEnv, passEnv))
	}
	if !hasHost && !hasUser && !hasPass {
		logger.Warn(disabledMsg, "integration", integration, "missing_env", []string{hostEnv})
	}
	return errs
}

func validateOAuthPair(integration, idVal, idEnv, secretVal, secretEnv string, logger *slog.Logger) []error {
	hasID := strings.TrimSpace(idVal) != ""
	hasSecret := strings.TrimSpace(secretVal) != ""
	if hasID != hasSecret {
		return []error{fmt.Errorf("partial %s configuration: both %s and %s must be set together", integration, idEnv, secretEnv)}
	}
	if !hasID && !hasSecret {
		logger.Warn("optional integration disabled", "integration", integration, "missing_env", []string{idEnv, secretEnv})
	}
	return nil
}

func GetConfig(ctx context.Context) (Config, error) {
	var cfg Config
	err := envconfig.Process(ctx, &cfg)
	return cfg, err
}
