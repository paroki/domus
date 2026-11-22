package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/errx"
)

func discoverMCPTools(
	ctx context.Context,
	session *mcp.ClientSession,
	patterns []string,
) ([]ai.Tool, error) {
	if session == nil || len(patterns) == 0 {
		return nil, nil
	}

	var tools []ai.Tool
	var cursor string
	for {
		res, err := session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return nil, err
		}
		for _, t := range res.Tools {
			if !matchesAnyPattern(patterns, t.Name) {
				continue
			}
			schema := map[string]any{"type": "object"}
			if t.InputSchema != nil {
				if raw, merr := json.Marshal(t.InputSchema); merr == nil && len(raw) > 0 {
					var m map[string]any
					if json.Unmarshal(raw, &m) == nil && len(m) > 0 {
						schema = m
					}
				}
			}
			toolName := t.Name
			tools = append(tools, ai.NewTool(toolName, t.Description,
				func(tc *ai.ToolContext, input any) (string, error) {
					return callMCPTool(tc.Context, session, toolName, input)
				},
				ai.WithInputSchema(schema),
			))
		}
		if res.NextCursor == "" {
			break
		}
		cursor = res.NextCursor
	}
	return tools, nil
}

func matchesAnyPattern(patterns []string, name string) bool {
	for _, pat := range patterns {
		if ok, err := path.Match(pat, name); err == nil && ok {
			return true
		}
	}
	return false
}

func callMCPTool(
	ctx context.Context,
	session *mcp.ClientSession,
	name string,
	input any,
) (string, error) {
	if session == nil {
		return "", errx.Internal("MCP_UNAVAILABLE", "MCP session is not connected.", nil)
	}
	args := map[string]any{}
	switch v := input.(type) {
	case map[string]any:
		if v != nil {
			args = v
		}
	case string:
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			if err := json.Unmarshal([]byte(trimmed), &args); err != nil {
				return "", errx.Invalid("INVALID_TOOL_ARGS", fmt.Sprintf("Invalid JSON arguments for %s: %v", name, err))
			}
		}
	case nil:
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return "", errx.Invalid("INVALID_TOOL_ARGS", fmt.Sprintf("Invalid arguments for %s: %v", name, err))
		}
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", errx.Invalid("INVALID_TOOL_ARGS", fmt.Sprintf("Invalid JSON arguments for %s: %v", name, err))
		}
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		return "", err
	}

	text := extractMCPResultText(res)
	if res.IsError {
		if text == "" {
			text = fmt.Sprintf("Tool %s failed.", name)
		}
		return "", errx.Invalid("TOOL_EXECUTION_ERROR", text)
	}
	return text, nil
}

func extractMCPResultText(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	var parts []string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok && tc.Text != "" {
			parts = append(parts, tc.Text)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}
	if res.StructuredContent != nil {
		if b, err := json.Marshal(res.StructuredContent); err == nil {
			return string(b)
		}
	}
	return "{}"
}
