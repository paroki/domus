package driver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeminiDriver_DiscoverModels(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gem-secret", r.Header.Get("x-goog-api-key"))
		if r.URL.Path == "/models" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"models": []map[string]any{
					{
						"name":                       "models/gemini-2.5-flash",
						"displayName":                "Gemini 2.5 Flash",
						"supportedGenerationMethods": []string{"generateContent"},
					},
					{
						"name":                       "models/text-embedding-004",
						"displayName":                "Text Embedding 004",
						"supportedGenerationMethods": []string{"embedContent"},
					},
				},
			})
		}
	}))
	defer srv.Close()

	reg := NewRegistry(srv.Client())
	models, err := reg.DiscoverModels(ctx, "gemini", srv.URL, "gem-secret")
	require.NoError(t, err)
	require.Len(t, models, 2)
	assert.Equal(t, "gemini-2.5-flash", models[0].ID)
	assert.Equal(t, []string{"chat"}, models[0].Capabilities)
	assert.Equal(t, []string{"embedding"}, models[1].Capabilities)
}

func TestAnthropicDriver_DiscoverModels(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ant-secret", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		if r.URL.Path == "/models" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": "claude-3-7-sonnet", "display_name": "Claude 3.7 Sonnet"}},
			})
		}
	}))
	defer srv.Close()

	reg := NewRegistry(srv.Client())
	models, err := reg.DiscoverModels(ctx, "anthropic", srv.URL, "ant-secret")
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "Claude 3.7 Sonnet", models[0].Name)
}

func TestOpenAIDriver_DiscoverModelsAndErrors(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer oa-secret", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/models":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "gpt-4o"},
					{"id": "text-embedding-3-small"},
				},
			})
		case "/fail/models":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid key oa-secret"))
		}
	}))
	defer srv.Close()

	reg := NewRegistry(srv.Client())
	models, err := reg.DiscoverModels(ctx, "openai", srv.URL, "oa-secret")
	require.NoError(t, err)
	assert.Equal(t, []string{"chat"}, models[0].Capabilities)
	assert.Equal(t, []string{"embedding"}, models[1].Capabilities)

	_, err = reg.DiscoverModels(ctx, "openai", srv.URL+"/fail", "oa-secret")
	require.ErrorIs(t, err, domain.ErrProviderAuthFailed)
	assert.NotContains(t, err.Error(), "oa-secret")

	_, err = reg.DiscoverModels(ctx, "openai", srv.URL, "")
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)

	assert.ErrorIs(t, providerStatusSentinel(http.StatusBadRequest), domain.ErrProviderRequestInvalid)
	assert.ErrorIs(t, providerStatusSentinel(http.StatusUnauthorized), domain.ErrProviderAuthFailed)
	assert.ErrorIs(t, providerStatusSentinel(http.StatusForbidden), domain.ErrProviderAuthFailed)
	assert.ErrorIs(t, providerStatusSentinel(http.StatusTooManyRequests), domain.ErrProviderRateLimited)
	assert.ErrorIs(t, providerStatusSentinel(http.StatusServiceUnavailable), domain.ErrProviderOverloaded)
	assert.ErrorIs(t, providerStatusSentinel(http.StatusBadGateway), domain.ErrProviderOverloaded)
}
