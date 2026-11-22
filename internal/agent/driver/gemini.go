package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/paroki/domus/internal/agent/domain"
)

const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

func geminiHeaders(apiKey string) map[string]string {
	return map[string]string{
		"content-type":   "application/json",
		"x-goog-api-key": apiKey,
	}
}

func (r *Registry) discoverGeminiModels(
	ctx context.Context,
	baseURL, apiKey string,
) ([]domain.ProviderModelItem, error) {
	resolved := resolveBaseURL(baseURL, defaultGeminiBaseURL)
	endpoint := resolved + "/models"

	resp, err := r.doJSONRequest(ctx, http.MethodGet, endpoint, geminiHeaders(apiKey), nil, apiKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Models []struct {
			Name                       string   `json:"name"`
			DisplayName                string   `json:"displayName"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode gemini models: %v", domain.ErrProviderUnavailable, err)
	}

	items := make([]domain.ProviderModelItem, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		id := strings.TrimPrefix(m.Name, "models/")
		var caps []string
		if slices.Contains(m.SupportedGenerationMethods, "generateContent") {
			caps = append(caps, "chat")
		}
		if slices.Contains(m.SupportedGenerationMethods, "embedContent") {
			caps = append(caps, "embedding")
		}
		name := m.DisplayName
		if name == "" {
			name = id
		}
		items = append(items, domain.ProviderModelItem{
			ID:           id,
			Name:         name,
			Capabilities: caps,
		})
	}
	return items, nil
}
