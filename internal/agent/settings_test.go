package agent_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/repo"
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

func TestRegisterSettings_EmbeddingResetTransitionOnly(t *testing.T) {
	ctx := context.Background()
	db, sqlDB, _ := testutil.NewEntClientWithDB(t)
	require.NoError(t, repo.EnsureMemorySearchSchema(ctx, sqlDB))

	prov1, err := db.AgentProvider.Create().
		SetName("Gemini 1").
		SetProvider(agentprovider.ProviderGemini).
		SetAPIKey("enc-1").
		Save(ctx)
	require.NoError(t, err)
	prov2, err := db.AgentProvider.Create().
		SetName("OpenAI 2").
		SetProvider(agentprovider.ProviderOpenai).
		SetAPIKey("enc-2").
		Save(ctx)
	require.NoError(t, err)

	r := repo.New(db, sqlDB)
	vec := make([]float32, repo.EmbeddingDimensions)
	vec[0] = 0.5
	mem, err := r.CreateMemory(ctx, domain.CreateMemoryInput{
		ParishID: uuid.New(),
		UserID:   uuid.New(),
		Category: domain.MemoryCategoryFact,
		Content:  "Umat aktif lingkungan",
	}, vec)
	require.NoError(t, err)

	// Pre-populate embedding setting in DB before Load
	_, err = db.Setting.Create().
		SetKey(agent.SettingModelEmbedding).
		SetValue(prov1.ID.String() + ":gemini-embedding-001").
		Save(ctx)
	require.NoError(t, err)

	store := setting.NewStore(db)
	feat := agent.New(
		nil,
		slog.New(slog.DiscardHandler),
		agent.WithDB(db),
		agent.WithSQLDB(sqlDB),
		agent.WithSettings(store),
	)
	feat.RegisterSettings(store)

	// Initial Load must preserve existing embeddings
	require.NoError(t, store.Load(ctx))
	var isNull bool
	require.NoError(t, sqlDB.QueryRowContext(ctx, `SELECT embedding IS NULL FROM agent_memories WHERE id = $1`, mem.ID).Scan(&isNull))
	require.False(t, isNull, "initial Load must not clear existing embeddings")

	// Transitioning to a different embedding model must reset embeddings to NULL
	_, err = store.Update(ctx, map[string]string{
		agent.SettingModelEmbedding: prov2.ID.String() + ":text-embedding-3-small",
	})
	require.NoError(t, err)
	require.NoError(t, sqlDB.QueryRowContext(ctx, `SELECT embedding IS NULL FROM agent_memories WHERE id = $1`, mem.ID).Scan(&isNull))
	require.True(t, isNull, "changing embedding model must reset embeddings to NULL")
}
