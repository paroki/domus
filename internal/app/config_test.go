package app

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfig_AutoMigrateDefaultAndOptIn(t *testing.T) {
	t.Setenv("APP_AUTO_MIGRATE", "")
	cfg, err := GetConfig(context.Background())
	require.NoError(t, err)
	assert.False(t, cfg.AutoMigrate)

	t.Setenv("APP_AUTO_MIGRATE", "true")
	cfg, err = GetConfig(context.Background())
	require.NoError(t, err)
	assert.True(t, cfg.AutoMigrate)
}

const testAppSecret = "0123456789abcdef0123456789abcdef"

func TestValidateStartupEnv_AllEmptyLogsWarningsAndSucceeds(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	err := ValidateStartupEnv(Config{Secret: testAppSecret}, StartupOAuthConfig{}, logger)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 3)

	seen := make(map[string]map[string]any, 3)
	for _, line := range lines {
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		assert.Equal(t, "WARN", entry["level"])
		integration, _ := entry["integration"].(string)
		seen[integration] = entry
	}

	assert.Equal(t, []any{"APP_SMTP_HOST"}, seen["smtp"]["missing_env"])
	assert.Contains(t, seen["smtp"]["msg"], "LogOTPSender")
	assert.Equal(t, []any{"AUTH_GOOGLE_ID", "AUTH_GOOGLE_SECRET"}, seen["oauth_google"]["missing_env"])
	assert.Equal(t, []any{"AUTH_GITHUB_ID", "AUTH_GITHUB_SECRET"}, seen["oauth_github"]["missing_env"])
}

func TestValidateStartupEnv_FullyConfiguredNoErrorOrWarnings(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	cfg := Config{
		Secret:   testAppSecret,
		SMTPHost: "smtp.example.com",
		SMTPUser: "mailer",
		SMTPPass: "smtp-pass",
	}
	oauth := StartupOAuthConfig{
		GoogleID:     "g-id",
		GoogleSecret: "g-secret",
		GitHubID:     "gh-id",
		GitHubSecret: "gh-secret",
	}

	err := ValidateStartupEnv(cfg, oauth, logger)
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(buf.String()))

	// Host without auth credentials is also valid for SMTP.
	cfgNoAuth := cfg
	cfgNoAuth.SMTPUser = ""
	cfgNoAuth.SMTPPass = ""
	err = ValidateStartupEnv(cfgNoAuth, oauth, logger)
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(buf.String()))
}

func TestValidateStartupEnv_PartialErrors(t *testing.T) {
	cases := []struct {
		name     string
		cfg      Config
		oauth    StartupOAuthConfig
		contains []string
	}{
		{
			name:     "secret empty",
			cfg:      Config{},
			contains: []string{"APP_SECRET"},
		},
		{
			name:     "secret too short",
			cfg:      Config{Secret: "short-secret"},
			contains: []string{"APP_SECRET"},
		},
		{
			name:     "smtp user without pass",
			cfg:      Config{Secret: testAppSecret, SMTPHost: "smtp.example.com", SMTPUser: "user"},
			contains: []string{"APP_SMTP_USER", "APP_SMTP_PASS"},
		},
		{
			name:     "smtp auth without host",
			cfg:      Config{Secret: testAppSecret, SMTPUser: "user", SMTPPass: "pass"},
			contains: []string{"APP_SMTP_HOST"},
		},
		{
			name:     "oauth google missing secret",
			cfg:      Config{Secret: testAppSecret},
			oauth:    StartupOAuthConfig{GoogleID: "g-id"},
			contains: []string{"AUTH_GOOGLE_ID", "AUTH_GOOGLE_SECRET"},
		},
		{
			name:     "oauth github missing id",
			cfg:      Config{Secret: testAppSecret},
			oauth:    StartupOAuthConfig{GitHubSecret: "gh-secret"},
			contains: []string{"AUTH_GITHUB_ID", "AUTH_GITHUB_SECRET"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStartupEnv(tc.cfg, tc.oauth, slog.New(slog.DiscardHandler))
			require.Error(t, err)
			for _, sub := range tc.contains {
				assert.Contains(t, err.Error(), sub)
			}
		})
	}
}

func TestValidateStartupEnv_CollectsAllErrorsAndHandlesNilLogger(t *testing.T) {
	err := ValidateStartupEnv(
		Config{
			SMTPPass: "p",
		},
		StartupOAuthConfig{
			GoogleID:     "gid",
			GitHubSecret: "ghsec",
		},
		nil,
	)
	require.Error(t, err)
	msg := err.Error()
	for _, envVar := range []string{
		"APP_SECRET",
		"APP_SMTP_USER",
		"APP_SMTP_PASS",
		"APP_SMTP_HOST",
		"AUTH_GOOGLE_ID",
		"AUTH_GOOGLE_SECRET",
		"AUTH_GITHUB_ID",
		"AUTH_GITHUB_SECRET",
	} {
		assert.Contains(t, msg, envVar)
	}
}
