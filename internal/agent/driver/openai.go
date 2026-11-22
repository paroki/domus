package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/paroki/domus/internal/agent/domain"
)

const (
	defaultOpenAIBaseURL     = "https://api.openai.com/v1"
	defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"
)

func openAIBaseURL(provider, custom string) string {
	if domain.ProviderName(provider) == domain.ProviderOpenRouter {
		return resolveBaseURL(custom, defaultOpenRouterBaseURL)
	}
	return resolveBaseURL(custom, defaultOpenAIBaseURL)
}

func openAIHeaders(apiKey string) map[string]string {
	return map[string]string{
		"content-type":  "application/json",
		"authorization": "Bearer " + apiKey,
	}
}

func (r *Registry) discoverOpenAIModels(
	ctx context.Context,
	provider, baseURL, apiKey string,
) ([]domain.ProviderModelItem, error) {
	resolved := openAIBaseURL(provider, baseURL)
	endpoint := resolved + "/models"

	resp, err := r.doJSONRequest(ctx, http.MethodGet, endpoint, openAIHeaders(apiKey), nil, apiKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode models: %v", domain.ErrProviderUnavailable, err)
	}
	items := make([]domain.ProviderModelItem, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		caps := []string{"chat"}
		if strings.Contains(strings.ToLower(m.ID), "embed") {
			caps = []string{"embedding"}
		}
		items = append(items, domain.ProviderModelItem{
			ID:           m.ID,
			Name:         m.ID,
			Capabilities: caps,
		})
	}
	return items, nil
}
