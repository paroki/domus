package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
)

// StreamEventType classifies incremental events emitted during StreamChat.
type StreamEventType string

const (
	StreamEventThinking StreamEventType = "thinking"
	StreamEventDelta    StreamEventType = "delta"
	StreamEventToolCall StreamEventType = "tool_call"
)

// StreamEvent represents a single streamed thinking chunk, text delta, or completed tool call.
type StreamEvent struct {
	Type     StreamEventType
	Delta    string
	ToolCall *domain.ToolCall
}

// StreamCallback receives incremental StreamEvents during StreamChat.
type StreamCallback func(StreamEvent) error

// ChatMessage represents a single turn in a provider-agnostic conversation history.
type ChatMessage struct {
	Role       domain.Role
	Content    string
	ToolCalls  []domain.ToolCall
	ToolCallID string
	ToolName   string
}

// ChatRequest is the provider-agnostic input for StreamChat.
type ChatRequest struct {
	SystemPrompt string
	Messages     []ChatMessage
	Tools        []ai.ToolRef
}

// ChatResponse is the aggregated output of StreamChat.
type ChatResponse struct {
	Content   string
	Thinking  string
	ToolCalls []domain.ToolCall
}

// Registry dispatches chat streaming, embedding, and model discovery across LLM providers.
type Registry struct {
	httpClient *http.Client
	pool       genkitPool
}

// NewRegistry creates a Registry using httpClient (or a default client with a 60s response-header timeout if nil).
func NewRegistry(httpClient *http.Client) *Registry {
	if httpClient == nil {
		// No overall Timeout: it would cut long SSE streams. Callers bound total time via ctx.
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.ResponseHeaderTimeout = 60 * time.Second
		httpClient = &http.Client{Transport: tr}
	}
	return &Registry{httpClient: httpClient}
}

// NormalizeEmbedding768 truncates vectors > 768 dims and L2-normalizes to unit length.
func NormalizeEmbedding768(vec []float32) ([]float32, error) {
	if len(vec) < domain.EmbeddingDimensions {
		return nil, fmt.Errorf("%w: embedding length %d is less than %d", domain.ErrProviderUnavailable, len(vec), domain.EmbeddingDimensions)
	}
	out := make([]float32, domain.EmbeddingDimensions)
	copy(out, vec[:domain.EmbeddingDimensions])

	var sumSq float64
	for _, v := range out {
		sumSq += float64(v) * float64(v)
	}
	norm := math.Sqrt(sumSq)
	if norm > 0 {
		for i, v := range out {
			out[i] = float32(float64(v) / norm)
		}
	}
	return out, nil
}

// StreamChat streams a single-turn chat completion from the configured provider.
func (r *Registry) StreamChat(
	ctx context.Context,
	modelCfg domain.ModelConfig,
	req ChatRequest,
	onEvent StreamCallback,
) (*ChatResponse, error) {
	resp, err := r.Generate(ctx, modelCfg, req, onEvent, ai.WithReturnToolRequests(true))
	if err != nil {
		return nil, err
	}
	out := &ChatResponse{Content: resp.Text(), Thinking: resp.Reasoning()}
	for _, tr := range resp.ToolRequests() {
		args, _ := json.Marshal(tr.ToolRequest.Input)
		id := tr.ToolRequest.Ref
		if id == "" {
			id = uuid.NewString()
		}
		tc := domain.ToolCall{ID: id, Name: tr.ToolRequest.Name, Arguments: string(args)}
		out.ToolCalls = append(out.ToolCalls, tc)
		if err := emitEvent(onEvent, StreamEvent{Type: StreamEventToolCall, ToolCall: &tc}); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Generate runs a multi-turn Genkit generation loop against the configured provider.
func (r *Registry) Generate(
	ctx context.Context,
	modelCfg domain.ModelConfig,
	req ChatRequest,
	onEvent StreamCallback,
	extraOpts ...ai.GenerateOption,
) (*ai.ModelResponse, error) {
	apiKey := strings.TrimSpace(modelCfg.APIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("%w: missing API key for provider %q", domain.ErrProviderUnavailable, modelCfg.Provider)
	}
	switch domain.ProviderName(modelCfg.Provider) {
	case domain.ProviderGemini, domain.ProviderAnthropic, domain.ProviderOpenAI, domain.ProviderOpenRouter:
		return r.generateGenkit(ctx, modelCfg, req, onEvent, extraOpts...)
	default:
		return nil, fmt.Errorf("%w: unsupported provider %q", domain.ErrProviderUnavailable, modelCfg.Provider)
	}
}

// Embed generates a normalized 768-dimension embedding vector for text.
func (r *Registry) Embed(
	ctx context.Context,
	provider, baseURL, apiKey, model, text string,
) ([]float32, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("%w: missing API key for provider %q", domain.ErrProviderUnavailable, provider)
	}
	switch domain.ProviderName(provider) {
	case domain.ProviderGemini:
		return r.embedGenkit(ctx, provider, baseURL, key, model, text)
	case domain.ProviderAnthropic:
		return nil, fmt.Errorf("%w: Anthropic does not support text embeddings", domain.ErrProviderUnavailable)
	case domain.ProviderOpenAI, domain.ProviderOpenRouter:
		return r.embedGenkit(ctx, provider, baseURL, key, model, text)
	default:
		return nil, fmt.Errorf("%w: unsupported provider %q", domain.ErrProviderUnavailable, provider)
	}
}

// DiscoverModels queries the provider's /models endpoint for available models.
func (r *Registry) DiscoverModels(
	ctx context.Context,
	provider, baseURL, apiKey string,
) ([]domain.ProviderModelItem, error) {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return nil, fmt.Errorf("%w: missing API key for provider %q", domain.ErrProviderUnavailable, provider)
	}
	switch domain.ProviderName(provider) {
	case domain.ProviderGemini:
		return r.discoverGeminiModels(ctx, baseURL, key)
	case domain.ProviderAnthropic:
		return r.discoverAnthropicModels(ctx, baseURL, key)
	case domain.ProviderOpenAI, domain.ProviderOpenRouter:
		return r.discoverOpenAIModels(ctx, provider, baseURL, key)
	default:
		return nil, fmt.Errorf("%w: unsupported provider %q", domain.ErrProviderUnavailable, provider)
	}
}

func (r *Registry) doJSONRequest(
	ctx context.Context,
	method, rawURL string,
	headers map[string]string,
	payload any,
	apiKey string,
) (*http.Response, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("%w: marshal request: %v", domain.ErrProviderUnavailable, err)
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrProviderUnavailable, redactAPIKey(err.Error(), apiKey))
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrProviderUnavailable, redactAPIKey(err.Error(), apiKey))
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer func() { _ = resp.Body.Close() }()
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := redactAPIKey(strings.TrimSpace(string(errBody)), apiKey)
		return nil, fmt.Errorf("%w: provider HTTP %d: %s", providerStatusSentinel(resp.StatusCode), resp.StatusCode, msg)
	}
	return resp, nil
}

func providerStatusSentinel(status int) error {
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return domain.ErrProviderAuthFailed
	case http.StatusTooManyRequests:
		return domain.ErrProviderRateLimited
	default:
		if status < http.StatusInternalServerError {
			return domain.ErrProviderRequestInvalid
		}
		return domain.ErrProviderOverloaded
	}
}

func redactAPIKey(s, apiKey string) string {
	if apiKey == "" {
		return s
	}
	return strings.ReplaceAll(s, apiKey, "[REDACTED]")
}

func emitEvent(cb StreamCallback, ev StreamEvent) error {
	if cb == nil {
		return nil
	}
	return cb(ev)
}

func resolveBaseURL(custom, fallback string) string {
	trimmed := strings.TrimSpace(custom)
	if trimmed == "" {
		return strings.TrimRight(fallback, "/")
	}
	return strings.TrimRight(trimmed, "/")
}
