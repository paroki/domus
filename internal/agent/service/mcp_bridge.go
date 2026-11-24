package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Per-agent MCP tool routing constants (never hardcoded in DB prompts).
var (
	GabrielAllowedMCPTools = map[string]bool{
		"cantus_search": true,
		"cantus_get":    true,
	}
	MikaelMCPPrefixes = []string{"event_"}
	RafaelMCPPrefixes = []string{"org_", "ref_", "user_"}
)

type mcpToolAdapter struct {
	session  *mcp.ClientSession
	mcpTool  *mcp.Tool
	toolInfo *schema.ToolInfo
}

func newMCPToolAdapter(session *mcp.ClientSession, mt *mcp.Tool) (*mcpToolAdapter, error) {
	var params *schema.ParamsOneOf
	if mt.InputSchema != nil {
		raw, err := json.Marshal(mt.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("marshal mcp schema %s: %w", mt.Name, err)
		}
		var js jsonschema.Schema
		if err := json.Unmarshal(raw, &js); err != nil {
			return nil, fmt.Errorf("unmarshal mcp jsonschema %s: %w", mt.Name, err)
		}
		params = schema.NewParamsOneOfByJSONSchema(&js)
	}
	return &mcpToolAdapter{
		session: session,
		mcpTool: mt,
		toolInfo: &schema.ToolInfo{
			Name:        mt.Name,
			Desc:        mt.Description,
			ParamsOneOf: params,
		},
	}, nil
}

func (a *mcpToolAdapter) Info(_ context.Context) (*schema.ToolInfo, error) {
	return a.toolInfo, nil
}

func (a *mcpToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var args map[string]any
	if trimmed := strings.TrimSpace(argumentsInJSON); trimmed != "" && trimmed != "null" {
		if err := json.Unmarshal([]byte(trimmed), &args); err != nil {
			return "", fmt.Errorf("invalid arguments JSON for %s: %w", a.mcpTool.Name, err)
		}
	}
	res, err := a.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      a.mcpTool.Name,
		Arguments: args,
	})
	if err != nil {
		return "", err
	}
	var parts []string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	text := strings.Join(parts, "\n")
	if res.IsError {
		return fmt.Sprintf("Tool %s error: %s", a.mcpTool.Name, text), nil
	}
	return text, nil
}

// LoadMCPTools lists tools from an open in-process MCP ClientSession and partitions them
// by Go-constant routing rules for gabriel, mikael, and rafael.
func LoadMCPTools(ctx context.Context, session *mcp.ClientSession) (gabriel, mikael, rafael []tool.BaseTool, err error) {
	if session == nil {
		return nil, nil, nil, nil
	}
	for mt, iterErr := range session.Tools(ctx, nil) {
		if iterErr != nil {
			return nil, nil, nil, iterErr
		}
		if mt == nil {
			continue
		}
		var target *[]tool.BaseTool
		switch {
		case GabrielAllowedMCPTools[mt.Name]:
			target = &gabriel
		case hasAnyPrefix(mt.Name, MikaelMCPPrefixes):
			target = &mikael
		case hasAnyPrefix(mt.Name, RafaelMCPPrefixes):
			target = &rafael
		default:
			continue
		}
		adapter, err := newMCPToolAdapter(session, mt)
		if err != nil {
			return nil, nil, nil, err
		}
		*target = append(*target, adapter)
	}
	return gabriel, mikael, rafael, nil
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}
