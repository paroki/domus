package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/openrouter"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/internal/agent/domain"
	"google.golang.org/genai"
)

const defaultClaudeMaxTokens = 4096

// NewChatModel constructs a classic *schema.Message Eino ToolCallingChatModel for the given provider.
func (r *Registry) NewChatModel(ctx context.Context, provider, baseURL, apiKey, modelID string) (model.ToolCallingChatModel, error) {
	key := strings.TrimSpace(apiKey)
	mid := strings.TrimSpace(modelID)
	if key == "" || mid == "" {
		return nil, fmt.Errorf("%w: missing API key or model ID for provider %q", domain.ErrProviderUnavailable, provider)
	}
	burl := strings.TrimRight(strings.TrimSpace(baseURL), "/")

	var (
		cm          model.ToolCallingChatModel
		defaultOpts []model.Option
		err         error
	)

	switch domain.ProviderName(provider) {
	case domain.ProviderGemini:
		gCfg := &genai.ClientConfig{
			APIKey:     key,
			HTTPClient: r.httpClient,
		}
		if burl != "" {
			gCfg.HTTPOptions = genai.HTTPOptions{BaseURL: burl}
		}
		client, clientErr := genai.NewClient(ctx, gCfg)
		if clientErr != nil {
			return nil, WrapProviderError(clientErr, key)
		}
		cm, err = gemini.NewChatModel(ctx, &gemini.Config{
			Client: client,
			Model:  mid,
		})
	case domain.ProviderAnthropic:
		cfg := &claude.Config{
			APIKey:           key,
			HTTPClient:       r.httpClient,
			Model:            mid,
			MaxTokens:        defaultClaudeMaxTokens,
			AutoCacheControl: &claude.CacheControl{},
		}
		if burl != "" {
			cfg.BaseURL = &burl
		}
		cm, err = claude.NewChatModel(ctx, cfg)
	case domain.ProviderOpenAI:
		cm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:     key,
			BaseURL:    burl,
			HTTPClient: r.httpClient,
			Model:      mid,
		})
	case domain.ProviderOpenRouter:
		if burl == "" {
			burl = defaultOpenRouterBaseURL
		}
		cm, err = openrouter.NewChatModel(ctx, &openrouter.Config{
			APIKey:     key,
			BaseURL:    burl,
			HTTPClient: r.httpClient,
			Model:      mid,
		})
	default:
		return nil, fmt.Errorf("%w: unsupported provider %q", domain.ErrProviderUnavailable, provider)
	}

	if err != nil {
		return nil, WrapProviderError(err, key)
	}
	return &wrappedChatModel{
		inner:       cm,
		apiKey:      key,
		defaultOpts: defaultOpts,
	}, nil
}

type wrappedChatModel struct {
	inner       model.ToolCallingChatModel
	apiKey      string
	defaultOpts []model.Option
}

func (w *wrappedChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	merged := append(append([]model.Option(nil), w.defaultOpts...), opts...)
	msg, err := w.inner.Generate(ctx, input, merged...)
	if err != nil {
		return nil, WrapProviderError(err, w.apiKey)
	}
	return msg, nil
}

func (w *wrappedChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	merged := append(append([]model.Option(nil), w.defaultOpts...), opts...)
	sr, err := w.inner.Stream(ctx, input, merged...)
	if err != nil {
		return nil, WrapProviderError(err, w.apiKey)
	}
	return sr, nil
}

func (w *wrappedChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	next, err := w.inner.WithTools(tools)
	if err != nil {
		return nil, WrapProviderError(err, w.apiKey)
	}
	return &wrappedChatModel{
		inner:       next,
		apiKey:      w.apiKey,
		defaultOpts: w.defaultOpts,
	}, nil
}
