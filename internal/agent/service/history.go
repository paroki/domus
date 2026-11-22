package service

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
)

// maxToolResultBytes caps a single tool result fed back to the LLM, DB and SSE.
const maxToolResultBytes = 12 * 1024

func truncateToolResult(s string) string {
	if len(s) <= maxToolResultBytes {
		return s
	}
	cut := maxToolResultBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "…[truncated]"
}

type delegateDepthKey struct{}

func inDelegation(ctx context.Context) bool {
	v, _ := ctx.Value(delegateDepthKey{}).(bool)
	return v
}

func withDelegation(ctx context.Context) context.Context {
	return context.WithValue(ctx, delegateDepthKey{}, true)
}

// sanitizeHistory drops tool messages without a preceding matching assistant tool call and
// assistant tool-call messages whose results are missing (window cut or aborted turn).
// Providers reject both shapes.
func sanitizeHistory(in []driver.ChatMessage) []driver.ChatMessage {
	out := make([]driver.ChatMessage, 0, len(in))
	for i := 0; i < len(in); i++ {
		m := in[i]
		switch {
		case m.Role == domain.RoleTool:
			continue // orphan: valid ones are consumed with their assistant message below
		case m.Role == domain.RoleAssistant && len(m.ToolCalls) > 0:
			want := make(map[string]struct{}, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				want[tc.ID] = struct{}{}
			}
			j := i + 1
			var results []driver.ChatMessage
			for ; j < len(in) && in[j].Role == domain.RoleTool; j++ {
				if _, ok := want[in[j].ToolCallID]; ok {
					delete(want, in[j].ToolCallID)
					results = append(results, in[j])
				}
			}
			i = j - 1
			if len(want) > 0 {
				if strings.TrimSpace(m.Content) != "" {
					m.ToolCalls = nil
					out = append(out, m)
				}
				continue
			}
			out = append(out, m)
			out = append(out, results...)
		default:
			out = append(out, m)
		}
	}
	return out
}

// closeDelegatedToolTurns inserts a synthetic assistant message after a
// delegate_specialist tool result when the next message is from the user,
// keeping the [assistant(tool_calls), tool, assistant, user] sequence valid
// across LLM providers on subsequent turns.
func closeDelegatedToolTurns(in []driver.ChatMessage) []driver.ChatMessage {
	out := make([]driver.ChatMessage, 0, len(in))
	for i, m := range in {
		out = append(out, m)
		if m.Role != domain.RoleTool || m.ToolName != "delegate_specialist" {
			continue
		}
		if i+1 >= len(in) || in[i+1].Role != domain.RoleUser {
			continue
		}
		var res struct {
			Answer string `json:"answer"`
		}
		_ = json.Unmarshal([]byte(m.Content), &res)
		if strings.TrimSpace(res.Answer) != "" {
			out = append(out, driver.ChatMessage{
				Role:    domain.RoleAssistant,
				Content: res.Answer,
			})
		}
	}
	return out
}
