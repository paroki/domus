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

func (r *Registry) streamAnthropic(
	ctx context.Context,
	modelCfg domain.ModelConfig,
	req ChatRequest,
	onEvent StreamCallback,
) (*ChatResponse, error) {
	baseURL := resolveBaseURL(modelCfg.BaseURL, defaultAnthropicBaseURL)
	endpoint := baseURL + "/messages"

	msgs := make([]map[string]any, 0, len(req.Messages))
	for _, msg := range req.Messages {
		switch msg.Role {
		case domain.RoleUser:
			msgs = append(msgs, map[string]any{
				"role":    "user",
				"content": msg.Content,
			})
		case domain.RoleAssistant:
			blocks := make([]map[string]any, 0, 1+len(msg.ToolCalls))
			if msg.Content != "" {
				blocks = append(blocks, map[string]any{
					"type": "text",
					"text": msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				var input any = map[string]any{}
				if tc.Arguments != "" {
					_ = json.Unmarshal([]byte(tc.Arguments), &input)
				}
				blocks = append(blocks, map[string]any{
					"type":  "tool_use",
					"id":    tc.ID,
					"name":  tc.Name,
					"input": input,
				})
			}
			msgs = append(msgs, map[string]any{
				"role":    "assistant",
				"content": blocks,
			})
		case domain.RoleTool:
			msgs = append(msgs, map[string]any{
				"role": "user",
				"content": []map[string]any{
					{
						"type":        "tool_result",
						"tool_use_id": msg.ToolCallID,
						"content":     msg.Content,
					},
				},
			})
		}
	}

	budget := anthropicThinkingBudget(modelCfg.ReasoningEffort)
	maxTokens := 4096
	if budget > 0 {
		maxTokens = budget + 4096
	}

	payload := map[string]any{
		"model":      modelCfg.Model,
		"max_tokens": maxTokens,
		"stream":     true,
		"messages":   msgs,
	}
	if strings.TrimSpace(req.SystemPrompt) != "" {
		payload["system"] = req.SystemPrompt
	}
	if budget > 0 {
		payload["thinking"] = map[string]any{
			"type":          "enabled",
			"budget_tokens": budget,
		}
	} else {
		temp := modelCfg.Temperature
		if temp <= 0 {
			temp = domain.DefaultTemperature
		}
		payload["temperature"] = temp
	}
	if len(req.Tools) > 0 {
		tools := make([]map[string]any, 0, len(req.Tools))
		for _, t := range req.Tools {
			var schema any = map[string]any{"type": "object"}
			if len(t.Parameters) > 0 {
				_ = json.Unmarshal(t.Parameters, &schema)
			}
			tools = append(tools, map[string]any{
				"name":         t.Name,
				"description":  t.Description,
				"input_schema": schema,
			})
		}
		payload["tools"] = tools
	}

	resp, err := r.doJSONRequest(ctx, http.MethodPost, endpoint, anthropicHeaders(modelCfg.APIKey), payload, modelCfg.APIKey)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	type pendingTool struct {
		id   string
		name string
		args strings.Builder
	}
	pendingTools := make(map[int]*pendingTool)
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
		var ev struct {
			Type         string `json:"type"`
			Index        int    `json:"index"`
			ContentBlock *struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
			Delta *struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				Thinking    string `json:"thinking"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
		}
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "content_block_start":
			if ev.ContentBlock != nil && ev.ContentBlock.Type == "tool_use" {
				pendingTools[ev.Index] = &pendingTool{
					id:   ev.ContentBlock.ID,
					name: ev.ContentBlock.Name,
				}
			}
		case "content_block_delta":
			if ev.Delta == nil {
				continue
			}
			switch ev.Delta.Type {
			case "thinking_delta":
				if ev.Delta.Thinking != "" {
					out.Thinking += ev.Delta.Thinking
					if err := emitEvent(onEvent, StreamEvent{Type: StreamEventThinking, Delta: ev.Delta.Thinking}); err != nil {
						return nil, err
					}
				}
			case "text_delta":
				if ev.Delta.Text != "" {
					out.Content += ev.Delta.Text
					if err := emitEvent(onEvent, StreamEvent{Type: StreamEventDelta, Delta: ev.Delta.Text}); err != nil {
						return nil, err
					}
				}
			case "input_json_delta":
				if pt, ok := pendingTools[ev.Index]; ok {
					pt.args.WriteString(ev.Delta.PartialJSON)
				}
			}
		case "content_block_stop":
			if pt, ok := pendingTools[ev.Index]; ok {
				argsStr := strings.TrimSpace(pt.args.String())
				if argsStr == "" {
					argsStr = "{}"
				}
				tc := domain.ToolCall{
					ID:        pt.id,
					Name:      pt.name,
					Arguments: argsStr,
				}
				out.ToolCalls = append(out.ToolCalls, tc)
				delete(pendingTools, ev.Index)
				if err := emitEvent(onEvent, StreamEvent{Type: StreamEventToolCall, ToolCall: &tc}); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%w: read anthropic stream: %v", domain.ErrProviderUnavailable, err)
	}
	return out, nil
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
