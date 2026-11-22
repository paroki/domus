package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentSettings_DefaultsValidationAndProviderBinding(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	var discoverCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discoverCalls++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]any{
				{
					"name":                       "models/gemini-2.5-pro",
					"displayName":                "Gemini 2.5 Pro",
					"supportedGenerationMethods": []string{"generateContent"},
				},
			},
		})
	}))
	defer srv.Close()

	store := setting.NewStore(db, setting.WithCipher(cipher))
	feat := New(
		driver.NewRegistry(srv.Client()),
		slog.New(slog.DiscardHandler),
		WithService(db, nil, "", nil, nil, nil, nil, nil),
		WithSettings(store),
		WithCipher(cipher),
	)
	setting.RegisterFeatures(store, feat)

	items := store.List(ctx)
	require.Len(t, items, 14)

	require.NoError(t, store.Load(ctx))
	assert.Equal(t, 0, discoverCalls)

	gabriel, ok := feat.Config().AgentConfig("gabriel")
	require.True(t, ok)
	assert.Equal(t, "", gabriel.Model.Provider)
	assert.Equal(t, "gemini-2.5-flash", gabriel.Model.Model)
	assert.Equal(t, domain.ReasoningLow, gabriel.Model.ReasoningEffort)
	assert.Equal(t, prompts.GabrielInstruction, gabriel.Instruction)

	emb := feat.Config().EmbeddingConfig()
	assert.Equal(t, "none", emb.Provider)
	assert.Equal(t, "text-embedding-004", emb.Model)

	// Invalid UUID / unknown UUID / effort rejected
	_, err = store.Update(ctx, map[string]string{SettingGabrielProviderID: "not-a-uuid"})
	require.Error(t, err)
	_, err = store.Update(ctx, map[string]string{SettingGabrielProviderID: uuid.NewString()})
	require.Error(t, err)
	_, err = store.Update(ctx, map[string]string{SettingGabrielEffort: "extreme"})
	require.Error(t, err)

	// Create two providers for the same upstream driver (gemini)
	gem1, err := feat.CreateProvider(ctx, domain.CreateAgentProviderRequest{
		Name:        "Gemini 1",
		Description: "Primary AI Studio key",
		Provider:    "gemini",
		APIKey:      "gem-test-key-1",
		BaseURL:     srv.URL,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, discoverCalls)
	assert.True(t, gem1.HasAPIKey)
	assert.NotContains(t, gem1.APIKeyMasked, "gem-test-key-1")
	require.Len(t, gem1.Models, 1)
	assert.Equal(t, "gemini-2.5-pro", gem1.Models[0].ID)

	anth, err := feat.CreateProvider(ctx, domain.CreateAgentProviderRequest{
		Name:     "Anthropic 1",
		Provider: "anthropic",
		APIKey:   "sk-ant-key",
		BaseURL:  srv.URL,
	})
	require.NoError(t, err)

	// Anthropic cannot be used for embedding
	_, err = store.Update(ctx, map[string]string{SettingEmbeddingProviderID: anth.ID.String()})
	require.Error(t, err)

	// Bind Gabriel & Embedding to Gemini 1 UUID
	_, err = store.Update(ctx, map[string]string{
		SettingGabrielProviderID:   gem1.ID.String(),
		SettingEmbeddingProviderID: gem1.ID.String(),
	})
	require.NoError(t, err)

	assert.True(t, feat.Config().IsProviderInUse(gem1.ID))
	assert.False(t, feat.Config().IsProviderInUse(anth.ID))

	gabriel, ok = feat.Config().AgentConfig("gabriel")
	require.True(t, ok)
	assert.Equal(t, "gemini", gabriel.Model.Provider)
	assert.Equal(t, "gem-test-key-1", gabriel.Model.APIKey)

	// Delete while in use -> ErrProviderInUse
	err = feat.DeleteProvider(ctx, gem1.ID)
	require.ErrorIs(t, err, domain.ErrProviderInUse)

	// Update API key -> propagates to bound agent & embedding
	newKey := "gem-test-key-2"
	_, err = feat.UpdateProvider(ctx, gem1.ID, domain.UpdateAgentProviderRequest{APIKey: &newKey})
	require.NoError(t, err)
	gabriel, _ = feat.Config().AgentConfig("gabriel")
	assert.Equal(t, "gem-test-key-2", gabriel.Model.APIKey)
	assert.Equal(t, "gem-test-key-2", feat.Config().EmbeddingConfig().APIKey)
}
