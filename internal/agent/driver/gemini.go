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

func geminiThinkingConfig(model string, effort domain.ReasoningEffort) map[string]any {
	if strings.HasPrefix(strings.ToLower(model), "gemma-") {
		return nil
	}
	switch effort {
	case domain.ReasoningLow:
		return map[string]any{"includeThoughts": true, "thinkingBudget": 1024}
	case domain.ReasoningMedium:
		return map[string]any{"includeThoughts": true, "thinkingBudget": 4096}
	case domain.ReasoningHigh:
		return map[string]any{"includeThoughts": true, "thinkingBudget": 16384}
	default:
		return nil
	}
}

// sanitizeGeminiSchema normalizes JSON Schema objects for Gemini v1beta compatibility
// by stripping unsupported keywords and converting type arrays (e.g. ["string", "null"])
// into a scalar type plus nullable: true.
func sanitizeGeminiSchema(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, child := range val {
			if k == "additionalProperties" || k == "$schema" {
				continue
			}
			if k == "type" {
				if arr, ok := child.([]any); ok {
					var primary string
					nullable := false
					for _, item := range arr {
						if s, ok := item.(string); ok {
							if s == "null" {
								nullable = true
							} else if primary == "" {
								primary = s
							}
						}
					}
					if primary != "" {
						out["type"] = primary
					}
					if nullable {
						out["nullable"] = true
					}
					continue
				}
			}
			out[k] = sanitizeGeminiSchema(child)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, child := range val {
			out[i] = sanitizeGeminiSchema(child)
		}
		return out
	default:
		return v
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
