package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

func (r *Registry) streamOpenAI(
	ctx context.Context,
	modelCfg domain.ModelConfig,
	req ChatRequest,
	onEvent StreamCallback,
) (*ChatResponse, error) {
	baseURL := openAIBaseURL(modelCfg.Provider, modelCfg.BaseURL)
	endpoint := baseURL + "/chat/completions"

	msgs := make([]map[string]any, 0, len(req.Messages)+1)
	if strings.TrimSpace(req.SystemPrompt) != "" {
		msgs = append(msgs, map[string]any{
			"role":    "system",
			"content": req.SystemPrompt,
		})
	}
	for _, msg := range req.Messages {
		switch msg.Role {
		case domain.RoleUser:
			msgs = append(msgs, map[string]any{
				"role":    "user",
				"content": msg.Content,
			})
		case domain.RoleAssistant:
			m := map[string]any{
				"role":    "assistant",
				"content": msg.Content,
			}
			if len(msg.ToolCalls) > 0 {
				tcs := make([]map[string]any, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					tcs = append(tcs, map[string]any{
						"id":   tc.ID,
						"type": "function",
						"function": map[string]any{
							"name":      tc.Name,
							"arguments": tc.Arguments,
						},
					})
				}
				m["tool_calls"] = tcs
			}
			msgs = append(msgs, m)
		case domain.RoleTool:
			msgs = append(msgs, map[string]any{
				"role":         "tool",
				"tool_call_id": msg.ToolCallID,
				"content":      msg.Content,
			})
		}
	}

	temp := modelCfg.Temperature
	if temp <= 0 {
		temp = domain.DefaultTemperature
	}

	payload := map[string]any{
		"model":       modelCfg.Model,
		"messages":    msgs,
		"stream":      true,
		"temperature": temp,
	}
	if modelCfg.ReasoningEffort != "" && modelCfg.ReasoningEffort != domain.ReasoningNone {
		if domain.ProviderName(modelCfg.Provider) == domain.ProviderOpenRouter {
			payload["reasoning"] = map[string]any{"effort": string(modelCfg.ReasoningEffort)}
		} else {
			payload["reasoning_effort"] = string(modelCfg.ReasoningEffort)
		}
	}
	if len(req.Tools) > 0 {
		tools := make([]map[string]any, 0, len(req.Tools))
		for _, t := range req.Tools {
			var params any = map[string]any{"type": "object"}
			if len(t.Parameters) > 0 {
				_ = json.Unmarshal(t.Parameters, &params)
			}
			tools = append(tools, map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        t.Name,
					"description": t.Description,
					"parameters":  params,
				},
			})
		}
		payload["tools"] = tools
	}

	resp, err := r.doJSONRequest(ctx, http.MethodPost, endpoint, openAIHeaders(modelCfg.APIKey), payload, modelCfg.APIKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	type accTool struct {
		id   string
		name string
		args strings.Builder
	}
	accTools := make(map[int]*accTool)
	out := &ChatResponse{}

	scanner := newSSEScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
					Reasoning        string `json:"reasoning"`
					ToolCalls        []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function *struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, ch := range chunk.Choices {
			thinkingDelta := ch.Delta.ReasoningContent
			if thinkingDelta == "" {
				thinkingDelta = ch.Delta.Reasoning
			}
			if thinkingDelta != "" {
				out.Thinking += thinkingDelta
				if err := emitEvent(onEvent, StreamEvent{Type: StreamEventThinking, Delta: thinkingDelta}); err != nil {
					return nil, err
				}
			}
			if ch.Delta.Content != "" {
				out.Content += ch.Delta.Content
				if err := emitEvent(onEvent, StreamEvent{Type: StreamEventDelta, Delta: ch.Delta.Content}); err != nil {
					return nil, err
				}
			}
			for _, tc := range ch.Delta.ToolCalls {
				at, ok := accTools[tc.Index]
				if !ok {
					at = &accTool{}
					accTools[tc.Index] = at
				}
				if tc.ID != "" {
					at.id = tc.ID
				}
				if tc.Function != nil {
					if tc.Function.Name != "" {
						at.name = tc.Function.Name
					}
					if tc.Function.Arguments != "" {
						at.args.WriteString(tc.Function.Arguments)
					}
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: read openai stream: %v", domain.ErrProviderUnavailable, err)
	}

	if len(accTools) > 0 {
		indices := make([]int, 0, len(accTools))
		for idx := range accTools {
			indices = append(indices, idx)
		}
		sort.Ints(indices)
		for _, idx := range indices {
			at := accTools[idx]
			argsStr := strings.TrimSpace(at.args.String())
			if argsStr == "" {
				argsStr = "{}"
			}
			tc := domain.ToolCall{
				ID:        at.id,
				Name:      at.name,
				Arguments: argsStr,
			}
			out.ToolCalls = append(out.ToolCalls, tc)
			if err := emitEvent(onEvent, StreamEvent{Type: StreamEventToolCall, ToolCall: &tc}); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func (r *Registry) embedOpenAI(
	ctx context.Context,
	provider, baseURL, apiKey, model, text string,
) ([]float32, error) {
	resolved := openAIBaseURL(provider, baseURL)
	endpoint := resolved + "/embeddings"
	payload := map[string]any{
		"model":      model,
		"input":      text,
		"dimensions": domain.EmbeddingDimensions,
	}

	resp, err := r.doJSONRequest(ctx, http.MethodPost, endpoint, openAIHeaders(apiKey), payload, apiKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode embeddings: %v", domain.ErrProviderUnavailable, err)
	}
	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("%w: empty embedding response", domain.ErrProviderUnavailable)
	}
	return NormalizeEmbedding768(parsed.Data[0].Embedding)
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
