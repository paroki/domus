package driver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	anthropicplugin "github.com/firebase/genkit/go/plugins/anthropic"
	oaiplugin "github.com/firebase/genkit/go/plugins/compat_oai/openai"
	openrouterplugin "github.com/firebase/genkit/go/plugins/compat_oai/openrouter"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	openaiopt "github.com/openai/openai-go/option"
	"google.golang.org/genai"

	"github.com/paroki/domus/internal/agent/domain"
)

// genkitPool caches one Genkit instance per (provider, baseURL, apiKey): plugins bind
// credentials at Init, and agent_provider configs carry per-config keys.
// ponytail: unbounded map; configs are few. Add LRU/TTL if configs churn.
type genkitPool struct {
	mu sync.Mutex
	m  map[string]*genkit.Genkit
}

// instance returns the cached Genkit instance and the model-name prefix for provider.
func (r *Registry) instance(ctx context.Context, provider, baseURL, apiKey string) (*genkit.Genkit, string, error) {
	var prefix string
	switch domain.ProviderName(provider) {
	case domain.ProviderGemini:
		prefix = "googleai/"
	case domain.ProviderAnthropic:
		prefix = "anthropic/"
	case domain.ProviderOpenAI:
		prefix = "openai/"
	case domain.ProviderOpenRouter:
		prefix = "openrouter/"
	default:
		return nil, "", fmt.Errorf("%w: unsupported provider %q", domain.ErrProviderUnavailable, provider)
	}
	sum := sha256.Sum256([]byte(provider + "\x00" + baseURL + "\x00" + apiKey))
	k := hex.EncodeToString(sum[:])

	r.pool.mu.Lock()
	defer r.pool.mu.Unlock()
	if g, ok := r.pool.m[k]; ok {
		return g, prefix, nil
	}
	var plugin api.Plugin
	switch domain.ProviderName(provider) {
	case domain.ProviderGemini:
		// Genkit appends the API version itself; strip it from the stored v1beta URL.
		base := strings.TrimSuffix(strings.TrimRight(strings.TrimSpace(baseURL), "/"), "/v1beta")
		plugin = &googlegenai.GoogleAI{APIKey: apiKey, BaseURL: base, HTTPClient: r.httpClient}
	case domain.ProviderAnthropic:
		base := strings.TrimSuffix(resolveBaseURL(baseURL, defaultAnthropicBaseURL), "/v1")
		plugin = &anthropicplugin.Anthropic{APIKey: apiKey, BaseURL: base, Opts: []anthropicopt.RequestOption{
			anthropicopt.WithHTTPClient(r.httpClient),
			anthropicopt.WithMaxRetries(0),
		}}
	case domain.ProviderOpenRouter:
		plugin = &openrouterplugin.OpenRouter{APIKey: apiKey, Opts: []openaiopt.RequestOption{
			openaiopt.WithBaseURL(openAIBaseURL(provider, baseURL)),
			openaiopt.WithHTTPClient(r.httpClient),
			openaiopt.WithMaxRetries(0),
		}}
	default:
		plugin = &oaiplugin.OpenAI{APIKey: apiKey, Opts: []openaiopt.RequestOption{
			openaiopt.WithBaseURL(openAIBaseURL(provider, baseURL)),
			openaiopt.WithHTTPClient(r.httpClient),
			openaiopt.WithMaxRetries(0),
		}}
	}
	g, err := initGenkit(ctx, plugin)
	if err != nil {
		return nil, "", err
	}
	if r.pool.m == nil {
		r.pool.m = map[string]*genkit.Genkit{}
	}
	r.pool.m[k] = g
	return g, prefix, nil
}

// initGenkit converts plugin Init panics (e.g. missing auth) into errors.
func initGenkit(ctx context.Context, p api.Plugin) (g *genkit.Genkit, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%w: init provider: %v", domain.ErrProviderUnavailable, rec)
		}
	}()
	return genkit.Init(ctx, genkit.WithPlugins(p)), nil
}

func providerConfig(cfg domain.ModelConfig) any {
	temp := cfg.Temperature
	if temp <= 0 {
		temp = domain.DefaultTemperature
	}
	budget := anthropicThinkingBudget(cfg.ReasoningEffort) // same low/medium/high budgets for all providers
	switch domain.ProviderName(cfg.Provider) {
	case domain.ProviderGemini:
		c := &genai.GenerateContentConfig{Temperature: genai.Ptr(float32(temp))}
		if budget > 0 && !strings.HasPrefix(strings.ToLower(cfg.Model), "gemma-") {
			c.ThinkingConfig = &genai.ThinkingConfig{IncludeThoughts: true, ThinkingBudget: genai.Ptr(int32(budget))}
		}
		return c
	case domain.ProviderAnthropic:
		c := anthropicsdk.MessageNewParams{MaxTokens: 4096}
		if budget > 0 {
			c.MaxTokens = int64(budget) + 4096
			c.Thinking = anthropicsdk.ThinkingConfigParamOfEnabled(int64(budget))
		} else {
			c.Temperature = anthropicsdk.Float(temp)
		}
		return c
	case domain.ProviderOpenRouter:
		c := openrouterplugin.ChatConfig{Temperature: &temp}
		if cfg.ReasoningEffort != "" && cfg.ReasoningEffort != domain.ReasoningNone {
			c.Reasoning = &openrouterplugin.ReasoningConfig{
				Effort: openrouterplugin.ReasoningEffort(cfg.ReasoningEffort),
			}
		}
		return c
	default:
		c := map[string]any{"temperature": temp}
		if cfg.ReasoningEffort != "" && cfg.ReasoningEffort != domain.ReasoningNone {
			c["reasoning_effort"] = string(cfg.ReasoningEffort)
		}
		return c
	}
}

func toolOutput(content string) any {
	var m map[string]any
	if json.Unmarshal([]byte(content), &m) == nil {
		return m
	}
	return map[string]any{"result": content}
}

func toGenkitMessages(in []ChatMessage) []*ai.Message {
	out := make([]*ai.Message, 0, len(in))
	for _, m := range in {
		switch m.Role {
		case domain.RoleUser:
			out = append(out, ai.NewUserTextMessage(m.Content))
		case domain.RoleAssistant:
			var parts []*ai.Part
			if m.Content != "" {
				parts = append(parts, ai.NewTextPart(m.Content))
			}
			for _, tc := range m.ToolCalls {
				var args any = map[string]any{}
				if tc.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Arguments), &args)
				}
				parts = append(parts, ai.NewToolRequestPart(&ai.ToolRequest{Name: tc.Name, Ref: tc.ID, Input: args}))
			}
			if len(parts) > 0 {
				out = append(out, ai.NewModelMessage(parts...))
			}
		case domain.RoleTool:
			part := ai.NewToolResponsePart(&ai.ToolResponse{Name: m.ToolName, Ref: m.ToolCallID, Output: toolOutput(m.Content)})
			// Providers want all results of one assistant turn in a single message.
			if n := len(out); n > 0 && out[n-1].Role == ai.RoleTool {
				out[n-1].Content = append(out[n-1].Content, part)
				continue
			}
			out = append(out, ai.NewMessage(ai.RoleTool, nil, part))
		}
	}
	return out
}

func (r *Registry) generateGenkit(
	ctx context.Context,
	cfg domain.ModelConfig,
	req ChatRequest,
	onEvent StreamCallback,
	extraOpts ...ai.GenerateOption,
) (*ai.ModelResponse, error) {
	key := strings.TrimSpace(cfg.APIKey)
	g, prefix, err := r.instance(ctx, cfg.Provider, cfg.BaseURL, key)
	if err != nil {
		return nil, err
	}
	var streamErr error
	opts := []ai.GenerateOption{
		ai.WithModelName(prefix + cfg.Model),
		ai.WithMessages(toGenkitMessages(req.Messages)...),
		ai.WithConfig(providerConfig(cfg)),
		ai.WithStreaming(func(_ context.Context, c *ai.ModelResponseChunk) error {
			if s := c.Reasoning(); s != "" {
				if err := emitEvent(onEvent, StreamEvent{Type: StreamEventThinking, Delta: s}); err != nil {
					streamErr = err
					return err
				}
			}
			if s := c.Text(); s != "" {
				if err := emitEvent(onEvent, StreamEvent{Type: StreamEventDelta, Delta: s}); err != nil {
					streamErr = err
					return err
				}
			}
			return nil
		}),
	}
	if strings.TrimSpace(req.SystemPrompt) != "" {
		opts = append(opts, ai.WithSystem(req.SystemPrompt))
	}
	if len(req.Tools) > 0 {
		opts = append(opts, ai.WithTools(req.Tools...))
	}
	opts = append(opts, extraOpts...)
	// Innermost middleware maps raw provider SDK errors while preserving caller/stream errors.
	opts = append(opts, ai.WithUse(ai.MiddlewareFunc(func(context.Context) (*ai.Hooks, error) {
		return &ai.Hooks{
			WrapModel: func(ctx context.Context, p *ai.ModelParams, next ai.ModelNext) (*ai.ModelResponse, error) {
				resp, err := next(ctx, p)
				if err != nil {
					if streamErr != nil {
						return nil, streamErr
					}
					return nil, mapGenkitError(err, key)
				}
				return resp, nil
			},
		}, nil
	})))

	resp, err := genkit.Generate(ctx, g, opts...)
	if err != nil {
		if errors.Is(err, ai.ErrMaxTurnsExceeded) {
			return nil, domain.ErrMaxStepsExceeded
		}
		return nil, err
	}
	return resp, nil
}

func (r *Registry) embedGenkit(ctx context.Context, provider, baseURL, key, model, text string) ([]float32, error) {
	if domain.ProviderName(provider) == domain.ProviderOpenRouter {
		baseURL = openAIBaseURL(provider, baseURL)
		provider = string(domain.ProviderOpenAI)
	}
	g, prefix, err := r.instance(ctx, provider, baseURL, key)
	if err != nil {
		return nil, err
	}
	var cfg any
	if domain.ProviderName(provider) == domain.ProviderGemini {
		cfg = &genai.EmbedContentConfig{OutputDimensionality: genai.Ptr(int32(domain.EmbeddingDimensions))}
	}
	e := genkit.LookupEmbedder(g, prefix+model)
	if e == nil {
		// Embedders for dynamic model IDs are not pre-registered.
		return nil, fmt.Errorf("%w: embedder %q not available", domain.ErrProviderUnavailable, model)
	}
	opts := []ai.EmbedderOption{ai.WithEmbedder(e), ai.WithTextDocs(text)}
	if cfg != nil {
		opts = append(opts, ai.WithConfig(cfg))
	}
	res, err := genkit.Embed(ctx, g, opts...)
	if err != nil {
		return nil, mapGenkitError(err, key)
	}
	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("%w: empty embedding response", domain.ErrProviderUnavailable)
	}
	return NormalizeEmbedding768(res.Embeddings[0].Embedding)
}

// mapGenkitError maps SDK errors onto domain sentinels by HTTP status text and redacts the key.
func mapGenkitError(err error, apiKey string) error {
	msg := redactAPIKey(err.Error(), apiKey)
	low := strings.ToLower(msg)
	var sentinel error
	switch {
	case strings.Contains(low, "401"), strings.Contains(low, "403"), strings.Contains(low, "unauthenticated"), strings.Contains(low, "permission"):
		sentinel = domain.ErrProviderAuthFailed
	case strings.Contains(low, "429"), strings.Contains(low, "resource_exhausted"), strings.Contains(low, "rate limit"):
		sentinel = domain.ErrProviderRateLimited
	case strings.Contains(low, "500"), strings.Contains(low, "502"), strings.Contains(low, "503"), strings.Contains(low, "529"), strings.Contains(low, "overloaded"):
		sentinel = domain.ErrProviderOverloaded
	case strings.Contains(low, "400"), strings.Contains(low, "404"), strings.Contains(low, "invalid"):
		sentinel = domain.ErrProviderRequestInvalid
	default:
		sentinel = domain.ErrProviderUnavailable
	}
	return fmt.Errorf("%w: %s", sentinel, msg)
}
