package agent_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestRegisterSettings_PromptsAndModelValidation(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	store := setting.NewStore(db)

	feat := agent.New(nil, slog.New(slog.DiscardHandler), agent.WithDB(db), agent.WithSettings(store))
	feat.RegisterSettings(store)
	require.NoError(t, store.Load(ctx))

	items := store.List(ctx)
	byKey := make(map[string]setting.SettingResponse, len(items))
	for _, item := range items {
		byKey[item.Key] = item
	}

	promptKeys := []string{
		agent.SettingPromptGabriel,
		agent.SettingPromptMikael,
		agent.SettingPromptRafael,
		agent.SettingPromptMemoryExtractor,
	}
	for _, k := range promptKeys {
		item, ok := byKey[k]
		require.True(t, ok, "missing prompt key %s", k)
		require.True(t, item.Multiline, "expected multiline=true for %s", k)
		require.True(t, item.Configured, "expected configured=true for %s", k)
		require.False(t, item.Overridden, "expected overridden=false for %s", k)
		require.NotEmpty(t, item.Value)
		require.False(t, strings.ContainsAny(item.Value, "{}"), "prompt %s must not contain template placeholders", k)
	}

	// Override prompt -> overridden=true; clear -> overridden=false & default restored
	defaultGabriel := byKey[agent.SettingPromptGabriel].Value
	updated, err := store.Update(ctx, map[string]string{
		agent.SettingPromptGabriel: "Custom Gabriel prompt",
	})
	require.NoError(t, err)
	for _, item := range updated {
		if item.Key == agent.SettingPromptGabriel {
			require.True(t, item.Overridden)
			require.Equal(t, "Custom Gabriel prompt", item.Value)
		}
	}

	cleared, err := store.Update(ctx, map[string]string{
		agent.SettingPromptGabriel: "",
	})
	require.NoError(t, err)
	for _, item := range cleared {
		if item.Key == agent.SettingPromptGabriel {
			require.False(t, item.Overridden)
			require.Equal(t, defaultGabriel, item.Value)
		}
	}

	// Model routing validation
	for _, bad := range []string{
		"gpt-4o",
		"not-a-uuid:gpt-4o",
		uuid.NewString() + ":",
		uuid.NewString() + ":gpt-4o",
	} {
		_, err := store.Update(ctx, map[string]string{
			agent.SettingModelGabriel: bad,
		})
		require.Error(t, err, "expected validation error for %q", bad)
	}

	prov, err := db.AgentProvider.Create().
		SetName("Gemini").
		SetProvider(agentprovider.ProviderGemini).
		SetAPIKey("enc-key").
		Save(ctx)
	require.NoError(t, err)

	_, err = store.Update(ctx, map[string]string{
		agent.SettingModelGabriel:   prov.ID.String() + ":gemini-2.5-flash",
		agent.SettingModelEmbedding: prov.ID.String() + ":gemini-embedding-001",
	})
	require.NoError(t, err)
	require.Equal(t, prov.ID.String()+":gemini-2.5-flash", store.Get(agent.SettingModelGabriel))
}

func TestRegisterPermissions(t *testing.T) {
	ctx := context.Background()
	enforcer, err := authz.NewMemoryEnforcer()
	require.NoError(t, err)

	feat := agent.New(nil, slog.New(slog.DiscardHandler))
	require.NoError(t, feat.RegisterPermissions(ctx, enforcer))

	superID := uuid.New()
	userID := uuid.New()
	require.NoError(t, enforcer.AssignRole(ctx, authz.UserSubject(superID), user.RoleSuperAdmin, authz.DomainGlobal))
	require.NoError(t, enforcer.AssignRole(ctx, authz.UserSubject(userID), user.RoleUser, authz.DomainGlobal))

	ok, err := enforcer.Enforce(authz.UserSubject(superID), authz.DomainGlobal, agent.PermProviders, authz.ActionRead)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = enforcer.Enforce(authz.UserSubject(superID), authz.DomainGlobal, agent.PermProviders, authz.ActionWrite)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = enforcer.Enforce(authz.UserSubject(userID), authz.DomainGlobal, agent.PermProviders, authz.ActionRead)
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = enforcer.Enforce(authz.UserSubject(userID), authz.DomainGlobal, agent.PermChat, authz.ActionRead)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = enforcer.Enforce(authz.UserSubject(userID), authz.DomainGlobal, agent.PermChat, authz.ActionWrite)
	require.NoError(t, err)
	require.True(t, ok)
}
