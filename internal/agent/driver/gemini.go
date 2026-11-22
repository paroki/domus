package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
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

func (r *Registry) streamGemini(
	ctx context.Context,
	modelCfg domain.ModelConfig,
	req ChatRequest,
	onEvent StreamCallback,
) (*ChatResponse, error) {
	baseURL := resolveBaseURL(modelCfg.BaseURL, defaultGeminiBaseURL)
	endpoint := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse", baseURL, modelCfg.Model)

	contents := make([]map[string]any, 0, len(req.Messages))
	for _, msg := range req.Messages {
		switch msg.Role {
		case domain.RoleUser:
			contents = append(contents, map[string]any{
				"role":  "user",
				"parts": []map[string]any{{"text": msg.Content}},
			})
		case domain.RoleAssistant:
			parts := make([]map[string]any, 0, 1+len(msg.ToolCalls))
			if msg.Content != "" {
				parts = append(parts, map[string]any{"text": msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				var args any = map[string]any{}
				if tc.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Arguments), &args)
				}
				parts = append(parts, map[string]any{
					"functionCall": map[string]any{
						"name": tc.Name,
						"args": args,
					},
				})
			}
			contents = append(contents, map[string]any{
				"role":  "model",
				"parts": parts,
			})
		case domain.RoleTool:
			var respObj any = map[string]any{"result": msg.Content}
			var parsed any
			if json.Unmarshal([]byte(msg.Content), &parsed) == nil {
				if m, ok := parsed.(map[string]any); ok {
					respObj = m
				}
			}
			contents = append(contents, map[string]any{
				"role": "user",
				"parts": []map[string]any{
					{
						"functionResponse": map[string]any{
							"name":     msg.ToolName,
							"response": respObj,
						},
					},
				},
			})
		}
	}

	temp := modelCfg.Temperature
	if temp <= 0 {
		temp = domain.DefaultTemperature
	}

	genCfg := map[string]any{
		"temperature": temp,
	}
	if tc := geminiThinkingConfig(modelCfg.Model, modelCfg.ReasoningEffort); tc != nil {
		genCfg["thinkingConfig"] = tc
	}
	payload := map[string]any{
		"contents":         contents,
		"generationConfig": genCfg,
	}
	if strings.TrimSpace(req.SystemPrompt) != "" {
		payload["systemInstruction"] = map[string]any{
			"parts": []map[string]any{{"text": req.SystemPrompt}},
		}
	}
	if len(req.Tools) > 0 {
		decls := make([]map[string]any, 0, len(req.Tools))
		for _, t := range req.Tools {
			var params any = map[string]any{"type": "object"}
			if len(t.Parameters) > 0 {
				_ = json.Unmarshal(t.Parameters, &params)
			}
			decls = append(decls, map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  sanitizeGeminiSchema(params),
			})
		}
		payload["tools"] = []map[string]any{{"functionDeclarations": decls}}
	}

	resp, err := r.doJSONRequest(ctx, http.MethodPost, endpoint, geminiHeaders(modelCfg.APIKey), payload, modelCfg.APIKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	out := &ChatResponse{}
	scanner := newSSEScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var chunk struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Thought      bool   `json:"thought"`
						Text         string `json:"text"`
						FunctionCall *struct {
							Name string          `json:"name"`
							Args json.RawMessage `json:"args"`
						} `json:"functionCall"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, cand := range chunk.Candidates {
			for _, part := range cand.Content.Parts {
				if part.FunctionCall != nil {
					argsStr := "{}"
					if len(part.FunctionCall.Args) > 0 {
						argsStr = string(part.FunctionCall.Args)
					}
					tc := domain.ToolCall{
						ID:        "call_" + uuid.NewString(),
						Name:      part.FunctionCall.Name,
						Arguments: argsStr,
					}
					out.ToolCalls = append(out.ToolCalls, tc)
					if err := emitEvent(onEvent, StreamEvent{Type: StreamEventToolCall, ToolCall: &tc}); err != nil {
						return nil, err
					}
					continue
				}
				if part.Text == "" {
					continue
				}
				if part.Thought {
					out.Thinking += part.Text
					if err := emitEvent(onEvent, StreamEvent{Type: StreamEventThinking, Delta: part.Text}); err != nil {
						return nil, err
					}
				} else {
					out.Content += part.Text
					if err := emitEvent(onEvent, StreamEvent{Type: StreamEventDelta, Delta: part.Text}); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: read gemini stream: %v", domain.ErrProviderUnavailable, err)
	}
	return out, nil
}

func (r *Registry) embedGemini(
	ctx context.Context,
	baseURL, apiKey, model, text string,
) ([]float32, error) {
	resolved := resolveBaseURL(baseURL, defaultGeminiBaseURL)
	endpoint := fmt.Sprintf("%s/models/%s:embedContent", resolved, model)
	payload := map[string]any{
		"content": map[string]any{
			"parts": []map[string]any{{"text": text}},
		},
		"outputDimensionality": domain.EmbeddingDimensions,
	}

	resp, err := r.doJSONRequest(ctx, http.MethodPost, endpoint, geminiHeaders(apiKey), payload, apiKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode gemini embedding: %v", domain.ErrProviderUnavailable, err)
	}
	return NormalizeEmbedding768(parsed.Embedding.Values)
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
