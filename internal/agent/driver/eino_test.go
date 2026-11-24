package driver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/stretchr/testify/require"
)

func TestEinoNewChatModel_AllProvidersAnd4xxErrors(t *testing.T) {
	ctx := context.Background()

	// 1. OpenAI & OpenRouter
	openAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer bad-key" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"message":"invalid key bad-key"}}`))
			return
		}
		if r.Header.Get("Authorization") == "Bearer rate-key" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"rate limit"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-1",
			"object":  "chat.completion",
			"created": 1,
			"model":   "gpt-4o",
			"choices": []map[string]any{
				{"index": 0, "message": map[string]any{"role": "assistant", "content": "Halo dari OpenAI"}, "finish_reason": "stop"},
			},
		})
	}))
	t.Cleanup(openAIServer.Close)

	reg := NewRegistry(openAIServer.Client())
	for _, prov := range []string{"openai", "openrouter"} {
		cm, err := reg.NewChatModel(ctx, prov, openAIServer.URL, "good-key", "gpt-4o")
		require.NoError(t, err)
		msg, err := cm.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
		require.NoError(t, err)
		require.Equal(t, "Halo dari OpenAI", msg.Content)
	}

	cmAuthFail, err := reg.NewChatModel(ctx, "openai", openAIServer.URL, "bad-key", "gpt-4o")
	require.NoError(t, err)
	_, err = cmAuthFail.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	require.ErrorIs(t, err, domain.ErrProviderAuthFailed)
	require.NotContains(t, err.Error(), "bad-key")

	cmRate, err := reg.NewChatModel(ctx, "openai", openAIServer.URL, "rate-key", "gpt-4o")
	require.NoError(t, err)
	_, err = cmRate.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	require.ErrorIs(t, err, domain.ErrProviderRateLimited)

	// 2. Anthropic (Claude) with auto-cache
	claudeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.Contains(t, string(body), `"max_tokens":4096`)
		require.Contains(t, string(body), `"cache_control":{"type":"ephemeral"}`)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "msg_1",
			"type":        "message",
			"role":        "assistant",
			"model":       "claude-sonnet-4-5",
			"stop_reason": "end_turn",
			"content":     []map[string]any{{"type": "text", "text": "Halo dari Claude"}},
			"usage":       map[string]any{"input_tokens": 10, "output_tokens": 5},
		})
	}))
	t.Cleanup(claudeServer.Close)

	claudeReg := NewRegistry(claudeServer.Client())
	cmClaude, err := claudeReg.NewChatModel(ctx, "anthropic", claudeServer.URL, "ant-key", "claude-sonnet-4-5")
	require.NoError(t, err)
	msg, err := cmClaude.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	require.NoError(t, err)
	require.Equal(t, "Halo dari Claude", msg.Content)

	// 3. Gemini
	geminiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{
				{"content": map[string]any{"role": "model", "parts": []map[string]any{{"text": "Halo dari Gemini"}}}},
			},
		})
	}))
	t.Cleanup(geminiServer.Close)

	geminiReg := NewRegistry(geminiServer.Client())
	cmGemini, err := geminiReg.NewChatModel(ctx, "gemini", geminiServer.URL, "gem-key", "gemini-2.5-flash")
	require.NoError(t, err)
	msg, err = cmGemini.Generate(ctx, []*schema.Message{schema.UserMessage("hi")})
	require.NoError(t, err)
	require.Equal(t, "Halo dari Gemini", msg.Content)
}

func TestEinoEmbedder_768DimensionsAndFallback(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	vec768 := make([]float64, EmbeddingDimensions)
	vec768[0] = 0.25
	returnDim := EmbeddingDimensions

	openAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.Contains(t, string(body), `"dimensions":768`)
		w.Header().Set("Content-Type", "application/json")
		vec := make([]float64, returnDim)
		if returnDim > 0 {
			vec[0] = 0.25
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"model":  "text-embedding-3-small",
			"data":   []map[string]any{{"object": "embedding", "index": 0, "embedding": vec}},
		})
	}))
	t.Cleanup(openAIServer.Close)

	reg := NewRegistry(openAIServer.Client())
	got := reg.EmbedTextBestEffort(ctx, "openai", openAIServer.URL, "sk-123", "text-embedding-3-small", "jadwal misa", logger)
	require.Len(t, got, EmbeddingDimensions)
	require.InDelta(t, 0.25, got[0], 1e-6)

	// OpenRouter also works via OpenAI embedder
	gotOR := reg.EmbedTextBestEffort(ctx, "openrouter", openAIServer.URL, "or-123", "openai/text-embedding-3-small", "jadwal misa", logger)
	require.Len(t, gotOR, EmbeddingDimensions)

	// Dimension mismatch (len != 768) falls back cleanly to nil
	returnDim = 512
	gotMismatch := reg.EmbedTextBestEffort(ctx, "openai", openAIServer.URL, "sk-123", "text-embedding-3-small", "jadwal misa", logger)
	require.Nil(t, gotMismatch)

	// Anthropic falls back cleanly to nil
	gotAnt := reg.EmbedTextBestEffort(ctx, "anthropic", "", "ant-123", "claude", "jadwal misa", logger)
	require.Nil(t, gotAnt)

	// Gemini sets outputDimensionality=768
	geminiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.Contains(t, string(body), `"outputDimensionality":768`)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"embeddings": []map[string]any{{"values": vec768}},
		})
	}))
	t.Cleanup(geminiServer.Close)

	gemReg := NewRegistry(geminiServer.Client())
	gotGem := gemReg.EmbedTextBestEffort(ctx, "gemini", geminiServer.URL, "gem-123", "gemini-embedding-001", "jadwal misa", logger)
	require.Len(t, gotGem, EmbeddingDimensions)
}
