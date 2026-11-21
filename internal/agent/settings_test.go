package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentSettings_DefaultsValidationAndAutoSync(t *testing.T) {
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
	feat := New(driver.NewRegistry(srv.Client()), slog.New(slog.DiscardHandler))
	setting.RegisterFeatures(store, feat)

	items := store.List(ctx)
	require.Len(t, items, 26)

	require.NoError(t, store.Load(ctx))
	assert.Equal(t, 0, discoverCalls)

	gabriel, ok := feat.Config().AgentConfig("gabriel")
	require.True(t, ok)
	assert.Equal(t, "gemini", gabriel.Model.Provider)
	assert.Equal(t, "gemini-2.5-flash", gabriel.Model.Model)
	assert.Equal(t, domain.ReasoningLow, gabriel.Model.ReasoningEffort)
	assert.Equal(t, prompts.GabrielInstruction, gabriel.Instruction)

	emb := feat.Config().EmbeddingConfig()
	assert.Equal(t, "gemini", emb.Provider)
	assert.Equal(t, "text-embedding-004", emb.Model)

	// Invalid provider / effort / URL rejected
	_, err = store.Update(ctx, map[string]string{SettingGabrielProvider: "unknown"})
	require.Error(t, err)
	_, err = store.Update(ctx, map[string]string{SettingGabrielEffort: "extreme"})
	require.Error(t, err)
	_, err = store.Update(ctx, map[string]string{SettingProviderGeminiBaseURL: "not-a-url"})
	require.Error(t, err)
	_, err = store.Update(ctx, map[string]string{SettingProviderGeminiModels: "{not-json-array}"})
	require.Error(t, err)

	// Transition api_key -> triggers best-effort /models auto-sync
	_, err = store.Update(ctx, map[string]string{
		SettingProviderGeminiBaseURL: srv.URL,
		SettingProviderGeminiAPIKey:  "gem-test-key",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, discoverCalls)

	models := feat.Config().ProviderModels("gemini")
	require.Len(t, models, 1)
	assert.Equal(t, "gemini-2.5-pro", models[0].ID)

	// Clearing api_key -> clears models to []
	_, err = store.Update(ctx, map[string]string{
		SettingProviderGeminiAPIKey: "",
	})
	require.NoError(t, err)
	assert.Empty(t, feat.Config().ProviderModels("gemini"))
	assert.Equal(t, "[]", store.Get(SettingProviderGeminiModels))
}
