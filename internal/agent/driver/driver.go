package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/paroki/domus/internal/agent/domain"
)

// Registry dispatches model discovery across LLM providers.
type Registry struct {
	httpClient *http.Client
}

// NewRegistry creates a Registry using httpClient (or a default client with a 60s response-header timeout if nil).
func NewRegistry(httpClient *http.Client) *Registry {
	if httpClient == nil {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.ResponseHeaderTimeout = 60 * time.Second
		httpClient = &http.Client{Transport: tr}
	}
	return &Registry{httpClient: httpClient}
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

func resolveBaseURL(custom, fallback string) string {
	trimmed := strings.TrimSpace(custom)
	if trimmed == "" {
		return strings.TrimRight(fallback, "/")
	}
	return strings.TrimRight(trimmed, "/")
}
