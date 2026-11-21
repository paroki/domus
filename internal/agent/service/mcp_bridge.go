package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/platform/errx"
)

func discoverMCPTools(
	ctx context.Context,
	session *mcp.ClientSession,
	patterns []string,
) ([]driver.ToolDefinition, map[string]struct{}, error) {
	if session == nil || len(patterns) == 0 {
		return nil, map[string]struct{}{}, nil
	}

	var defs []driver.ToolDefinition
	allowed := make(map[string]struct{})
	var cursor string
	for {
		res, err := session.ListTools(ctx, &mcp.ListToolsParams{Cursor: cursor})
		if err != nil {
			return nil, nil, err
		}
		for _, t := range res.Tools {
			if !matchesAnyPattern(patterns, t.Name) {
				continue
			}
			allowed[t.Name] = struct{}{}
			params := json.RawMessage(`{"type":"object"}`)
			if t.InputSchema != nil {
				if raw, merr := json.Marshal(t.InputSchema); merr == nil && len(raw) > 0 {
					params = raw
				}
			}
			defs = append(defs, driver.ToolDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			})
		}
		if res.NextCursor == "" {
			break
		}
		cursor = res.NextCursor
	}
	return defs, allowed, nil
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
	argsJSON string,
) (string, error) {
	if session == nil {
		return "", errx.Internal("MCP_UNAVAILABLE", "MCP session is not connected.", nil)
	}
	args := map[string]any{}
	if trimmed := strings.TrimSpace(argsJSON); trimmed != "" {
		if err := json.Unmarshal([]byte(trimmed), &args); err != nil {
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
