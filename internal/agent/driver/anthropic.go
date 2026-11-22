package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paroki/domus/internal/agent/domain"
)

const (
	defaultAnthropicBaseURL = "https://api.anthropic.com/v1"
	anthropicAPIVersion     = "2023-06-01"
)

func anthropicHeaders(apiKey string) map[string]string {
	return map[string]string{
		"content-type":      "application/json",
		"x-api-key":         apiKey,
		"anthropic-version": anthropicAPIVersion,
	}
}

func anthropicThinkingBudget(effort domain.ReasoningEffort) int {
	switch effort {
	case domain.ReasoningLow:
		return 1024
	case domain.ReasoningMedium:
		return 4096
	case domain.ReasoningHigh:
		return 16384
	default:
		return 0
	}
}

func (r *Registry) discoverAnthropicModels(
	ctx context.Context,
	baseURL, apiKey string,
) ([]domain.ProviderModelItem, error) {
	resolved := resolveBaseURL(baseURL, defaultAnthropicBaseURL)
	endpoint := resolved + "/models"

	resp, err := r.doJSONRequest(ctx, http.MethodGet, endpoint, anthropicHeaders(apiKey), nil, apiKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode anthropic models: %v", domain.ErrProviderUnavailable, err)
	}
	items := make([]domain.ProviderModelItem, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		name := m.DisplayName
		if name == "" {
			name = m.ID
		}
		items = append(items, domain.ProviderModelItem{
			ID:           m.ID,
			Name:         name,
			Capabilities: []string{"chat"},
		})
	}
	return items, nil
}
