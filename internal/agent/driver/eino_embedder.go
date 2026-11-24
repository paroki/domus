package driver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	embGemini "github.com/cloudwego/eino-ext/components/embedding/gemini"
	embOpenAI "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/paroki/domus/internal/agent/domain"
	"google.golang.org/genai"
)

// EmbeddingDimensions is the fixed vector width expected by agent_memories.embedding halfvec(768).
const EmbeddingDimensions = 768

// NewEmbedder constructs a 768-dimension Eino Embedder for gemini, openai, or openrouter.
func (r *Registry) NewEmbedder(ctx context.Context, provider, baseURL, apiKey, modelID string) (embedding.Embedder, error) {
	key := strings.TrimSpace(apiKey)
	mid := strings.TrimSpace(modelID)
	if key == "" || mid == "" {
		return nil, fmt.Errorf("%w: missing API key or model ID for embedder %q", domain.ErrProviderUnavailable, provider)
	}
	burl := strings.TrimRight(strings.TrimSpace(baseURL), "/")

	switch domain.ProviderName(provider) {
	case domain.ProviderGemini:
		gCfg := &genai.ClientConfig{
			APIKey:     key,
			HTTPClient: r.httpClient,
		}
		if burl != "" {
			gCfg.HTTPOptions = genai.HTTPOptions{BaseURL: burl}
		}
		client, err := genai.NewClient(ctx, gCfg)
		if err != nil {
			return nil, WrapProviderError(err, key)
		}
		dim := int32(EmbeddingDimensions)
		emb, err := embGemini.NewEmbedder(ctx, &embGemini.EmbeddingConfig{
			Client:               client,
			Model:                mid,
			OutputDimensionality: &dim,
		})
		if err != nil {
			return nil, WrapProviderError(err, key)
		}
		return emb, nil
	case domain.ProviderOpenAI:
		dim := EmbeddingDimensions
		emb, err := embOpenAI.NewEmbedder(ctx, &embOpenAI.EmbeddingConfig{
			APIKey:     key,
			BaseURL:    burl,
			HTTPClient: r.httpClient,
			Model:      mid,
			Dimensions: &dim,
		})
		if err != nil {
			return nil, WrapProviderError(err, key)
		}
		return emb, nil
	case domain.ProviderOpenRouter:
		if burl == "" {
			burl = defaultOpenRouterBaseURL
		}
		dim := EmbeddingDimensions
		emb, err := embOpenAI.NewEmbedder(ctx, &embOpenAI.EmbeddingConfig{
			APIKey:     key,
			BaseURL:    burl,
			HTTPClient: r.httpClient,
			Model:      mid,
			Dimensions: &dim,
		})
		if err != nil {
			return nil, WrapProviderError(err, key)
		}
		return emb, nil
	case domain.ProviderAnthropic:
		return nil, fmt.Errorf("%w: anthropic does not support embeddings", domain.ErrProviderUnavailable)
	default:
		return nil, fmt.Errorf("%w: unsupported embedding provider %q", domain.ErrProviderUnavailable, provider)
	}
}

// EmbedTextBestEffort computes a 768-dimension float32 embedding for text.
// On any error or dimension mismatch (len != 768), it logs a warning (with apiKey redacted) and returns nil for BM25-only fallback.
func (r *Registry) EmbedTextBestEffort(
	ctx context.Context,
	provider, baseURL, apiKey, modelID, text string,
	logger *slog.Logger,
) []float32 {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}
	emb, err := r.NewEmbedder(ctx, provider, baseURL, apiKey, modelID)
	if err != nil {
		logger.WarnContext(ctx, "embedding setup failed; falling back to BM25-only", "provider", provider, "err", err)
		return nil
	}
	vecs, err := emb.EmbedStrings(ctx, []string{text})
	if err != nil {
		logger.WarnContext(ctx, "embedding request failed; falling back to BM25-only", "provider", provider, "err", WrapProviderError(err, apiKey))
		return nil
	}
	if len(vecs) != 1 || len(vecs[0]) != EmbeddingDimensions {
		gotLen := 0
		if len(vecs) > 0 {
			gotLen = len(vecs[0])
		}
		logger.WarnContext(ctx, "unexpected embedding dimensions; falling back to BM25-only", "provider", provider, "got", gotLen, "want", EmbeddingDimensions)
		return nil
	}
	out := make([]float32, EmbeddingDimensions)
	for i, v := range vecs[0] {
		out[i] = float32(v)
	}
	return out
}
