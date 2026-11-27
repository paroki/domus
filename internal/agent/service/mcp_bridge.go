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
	GabrielAllowedMCPTools = map[string]bool{}

	MikaelMCPPrefixes    = []string{"event_"}
	RafaelMCPPrefixes    = []string{"org_", "ref_", "user_"}
	YosefMCPPrefixes     = []string{"umat_"}
	YosefSharedMCPTools  = map[string]bool{"ref_get_paroki": true, "ref_list_wilayah": true, "ref_list_lingkungan": true}
	GregoriusMCPPrefixes = []string{"liturgi_"}
	GregoriusExcluded    = map[string]bool{"liturgi_set_lagu": true}
	SesiliaMCPPrefixes   = []string{"cantus_"}
	SesiliaSharedTools   = map[string]bool{
		"liturgi_list_schedules": true,
		"liturgi_get_schedule":   true,
		"liturgi_set_lagu":       true,
	}
)

// MCPToolSet holds partitioned MCP tools for each specialist sub-agent.
type MCPToolSet struct {
	Mikael    []tool.BaseTool
	Rafael    []tool.BaseTool
	Yosef     []tool.BaseTool
	Gregorius []tool.BaseTool
	Sesilia   []tool.BaseTool
}

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
// by Go-constant routing rules across Mikael, Rafael, Yosef, Gregorius, and Sesilia.
func LoadMCPTools(ctx context.Context, session *mcp.ClientSession) (MCPToolSet, error) {
	var set MCPToolSet
	if session == nil {
		return set, nil
	}
	for mt, iterErr := range session.Tools(ctx, nil) {
		if iterErr != nil {
			return MCPToolSet{}, iterErr
		}
		if mt == nil {
			continue
		}

		toMikael := hasAnyPrefix(mt.Name, MikaelMCPPrefixes)
		toRafael := hasAnyPrefix(mt.Name, RafaelMCPPrefixes)
		toYosef := hasAnyPrefix(mt.Name, YosefMCPPrefixes) || YosefSharedMCPTools[mt.Name]
		toGregorius := hasAnyPrefix(mt.Name, GregoriusMCPPrefixes) && !GregoriusExcluded[mt.Name]
		toSesilia := hasAnyPrefix(mt.Name, SesiliaMCPPrefixes) || SesiliaSharedTools[mt.Name]

		if !toMikael && !toRafael && !toYosef && !toGregorius && !toSesilia {
			continue
		}

		adapter, err := newMCPToolAdapter(session, mt)
		if err != nil {
			return MCPToolSet{}, err
		}
		if toMikael {
			set.Mikael = append(set.Mikael, adapter)
		}
		if toRafael {
			set.Rafael = append(set.Rafael, adapter)
		}
		if toYosef {
			set.Yosef = append(set.Yosef, adapter)
		}
		if toGregorius {
			set.Gregorius = append(set.Gregorius, adapter)
		}
		if toSesilia {
			set.Sesilia = append(set.Sesilia, adapter)
		}
	}
	return set, nil
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}
