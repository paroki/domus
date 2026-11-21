package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEmbedding768(t *testing.T) {
	_, err := NormalizeEmbedding768(make([]float32, 100))
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)

	raw := make([]float32, 1536)
	raw[0] = 3
	raw[1] = 4
	normed, err := NormalizeEmbedding768(raw)
	require.NoError(t, err)
	require.Len(t, normed, 768)
	assert.InDelta(t, 0.6, normed[0], 1e-5)
	assert.InDelta(t, 0.8, normed[1], 1e-5)

	var sumSq float64
	for _, v := range normed {
		sumSq += float64(v) * float64(v)
	}
	assert.InDelta(t, 1.0, math.Sqrt(sumSq), 1e-5)
}

func TestGeminiDriver_StreamEmbedAndDiscover(t *testing.T) {
	ctx := context.Background()
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gem-secret", r.Header.Get("x-goog-api-key"))
		switch {
		case strings.Contains(r.URL.Path, ":streamGenerateContent"):
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &capturedBody)
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"thought\":true,\"text\":\"Mengecek...\"}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"functionCall\":{\"name\":\"event_list\",\"args\":{\"limit\":5}}}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"Halo!\"}]}}]}\n\n")
		case strings.Contains(r.URL.Path, ":embedContent"):
			vec := make([]float32, 768)
			vec[0] = 2.0
			_ = json.NewEncoder(w).Encode(map[string]any{"embedding": map[string]any{"values": vec}})
		case r.URL.Path == "/models":
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
	var events []StreamEvent
	resp, err := reg.StreamChat(ctx, domain.ModelConfig{
		Provider:        "gemini",
		BaseURL:         srv.URL,
		APIKey:          "gem-secret",
		Model:           "gemini-2.5-flash",
		ReasoningEffort: domain.ReasoningMedium,
	}, ChatRequest{
		SystemPrompt: "System",
		Messages:     []ChatMessage{{Role: domain.RoleUser, Content: "Hi"}},
		Tools:        []ToolDefinition{{Name: "event_list", Description: "list", Parameters: json.RawMessage(`{"type":"object"}`)}},
	}, func(ev StreamEvent) error {
		events = append(events, ev)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "Mengecek...", resp.Thinking)
	assert.Equal(t, "Halo!", resp.Content)
	require.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "event_list", resp.ToolCalls[0].Name)
	assert.Len(t, events, 3)

	genCfg := capturedBody["generationConfig"].(map[string]any)
	thinkingCfg := genCfg["thinkingConfig"].(map[string]any)
	assert.Equal(t, true, thinkingCfg["includeThoughts"])
	assert.Equal(t, float64(4096), thinkingCfg["thinkingBudget"])

	emb, err := reg.Embed(ctx, "gemini", srv.URL, "gem-secret", "text-embedding-004", "jadwal misa")
	require.NoError(t, err)
	assert.Len(t, emb, 768)
	assert.InDelta(t, 1.0, emb[0], 1e-5)

	models, err := reg.DiscoverModels(ctx, "gemini", srv.URL, "gem-secret")
	require.NoError(t, err)
	require.Len(t, models, 2)
	assert.Equal(t, "gemini-2.5-flash", models[0].ID)
	assert.Equal(t, []string{"chat"}, models[0].Capabilities)
	assert.Equal(t, []string{"embedding"}, models[1].Capabilities)
}

func TestAnthropicDriver_StreamDiscoverAndNoEmbed(t *testing.T) {
	ctx := context.Background()
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ant-secret", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))
		switch r.URL.Path {
		case "/messages":
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &capturedBody)
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = fmt.Fprint(w, "data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"thinking_delta\",\"thinking\":\"Berpikir...\"}}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tu_1\",\"name\":\"org_list\"}}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"q\\\":\\\"koor\\\"}\"}}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"type\":\"content_block_stop\",\"index\":1}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"type\":\"content_block_delta\",\"index\":2,\"delta\":{\"type\":\"text_delta\",\"text\":\"Siap!\"}}\n\n")
		case "/models":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": "claude-3-7-sonnet", "display_name": "Claude 3.7 Sonnet"}},
			})
		}
	}))
	defer srv.Close()

	reg := NewRegistry(srv.Client())
	resp, err := reg.StreamChat(ctx, domain.ModelConfig{
		Provider:        "anthropic",
		BaseURL:         srv.URL,
		APIKey:          "ant-secret",
		Model:           "claude-3-7-sonnet",
		ReasoningEffort: domain.ReasoningHigh,
	}, ChatRequest{
		SystemPrompt: "System",
		Messages:     []ChatMessage{{Role: domain.RoleUser, Content: "Hi"}},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "Berpikir...", resp.Thinking)
	assert.Equal(t, "Siap!", resp.Content)
	require.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, "tu_1", resp.ToolCalls[0].ID)
	assert.Equal(t, `{"q":"koor"}`, resp.ToolCalls[0].Arguments)
	_, hasTemp := capturedBody["temperature"]
	assert.False(t, hasTemp)

	_, err = reg.Embed(ctx, "anthropic", srv.URL, "ant-secret", "claude-3-7-sonnet", "hi")
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)

	models, err := reg.DiscoverModels(ctx, "anthropic", srv.URL, "ant-secret")
	require.NoError(t, err)
	require.Len(t, models, 1)
	assert.Equal(t, "Claude 3.7 Sonnet", models[0].Name)
}

func TestOpenAIAndOpenRouterDriver_StreamEmbedAndErrors(t *testing.T) {
	ctx := context.Background()
	var capturedBodies []map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer oa-secret", r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/chat/completions":
			b, _ := io.ReadAll(r.Body)
			var body map[string]any
			_ = json.Unmarshal(b, &body)
			capturedBodies = append(capturedBodies, body)
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"Hmm\"}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call_oa\",\"function\":{\"name\":\"memory_recall\",\"arguments\":\"{\\\"q\\\":\"}}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"\\\"misa\\\"}\"}}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Baik.\"}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		case "/embeddings":
			vec := make([]float32, 768)
			vec[0] = 5.0
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"embedding": vec}},
			})
		case "/models":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "gpt-4o"},
					{"id": "text-embedding-3-small"},
				},
			})
		case "/fail":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid key oa-secret"))
		}
	}))
	defer srv.Close()

	reg := NewRegistry(srv.Client())
	resp, err := reg.StreamChat(ctx, domain.ModelConfig{
		Provider:        "openai",
		BaseURL:         srv.URL,
		APIKey:          "oa-secret",
		Model:           "gpt-4o",
		ReasoningEffort: domain.ReasoningLow,
	}, ChatRequest{Messages: []ChatMessage{{Role: domain.RoleUser, Content: "Hi"}}}, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hmm", resp.Thinking)
	assert.Equal(t, "Baik.", resp.Content)
	require.Len(t, resp.ToolCalls, 1)
	assert.Equal(t, `{"q":"misa"}`, resp.ToolCalls[0].Arguments)
	assert.Equal(t, "low", capturedBodies[0]["reasoning_effort"])

	_, err = reg.StreamChat(ctx, domain.ModelConfig{
		Provider:        "openrouter",
		BaseURL:         srv.URL,
		APIKey:          "oa-secret",
		Model:           "openai/gpt-4o",
		ReasoningEffort: domain.ReasoningMedium,
	}, ChatRequest{Messages: []ChatMessage{{Role: domain.RoleUser, Content: "Hi"}}}, nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"effort": "medium"}, capturedBodies[1]["reasoning"])

	emb, err := reg.Embed(ctx, "openai", srv.URL, "oa-secret", "text-embedding-3-small", "hi")
	require.NoError(t, err)
	assert.InDelta(t, 1.0, emb[0], 1e-5)

	models, err := reg.DiscoverModels(ctx, "openai", srv.URL, "oa-secret")
	require.NoError(t, err)
	assert.Equal(t, []string{"chat"}, models[0].Capabilities)
	assert.Equal(t, []string{"embedding"}, models[1].Capabilities)

	// Missing API key and HTTP >= 400 wrap ErrProviderUnavailable and redact key
	_, err = reg.StreamChat(ctx, domain.ModelConfig{Provider: "openai", Model: "gpt-4o"}, ChatRequest{}, nil)
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)

	_, err = reg.DiscoverModels(ctx, "openai", srv.URL+"/fail", "oa-secret")
	require.ErrorIs(t, err, domain.ErrProviderUnavailable)
	assert.NotContains(t, err.Error(), "oa-secret")
}
