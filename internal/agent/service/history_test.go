package service

import (
	"strings"
	"testing"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
)

func TestSanitizeHistory(t *testing.T) {
	call := func(id string) []domain.ToolCall { return []domain.ToolCall{{ID: id, Name: "x"}} }
	in := []driver.ChatMessage{
		{Role: domain.RoleTool, ToolCallID: "a"}, // orphan at window start
		{Role: domain.RoleUser, Content: "hi"},
		{Role: domain.RoleAssistant, ToolCalls: call("b")}, // aborted: no result
		{Role: domain.RoleUser, Content: "again"},
		{Role: domain.RoleAssistant, ToolCalls: call("c")},
		{Role: domain.RoleTool, ToolCallID: "c"},
		{Role: domain.RoleAssistant, Content: "done"},
	}
	out := sanitizeHistory(in)
	var roles []domain.Role
	for _, m := range out {
		roles = append(roles, m.Role)
	}
	want := []domain.Role{domain.RoleUser, domain.RoleUser, domain.RoleAssistant, domain.RoleTool, domain.RoleAssistant}
	if len(roles) != len(want) {
		t.Fatalf("roles=%v want %v", roles, want)
	}
	for i := range want {
		if roles[i] != want[i] {
			t.Fatalf("roles=%v want %v", roles, want)
		}
	}
}

func TestTruncateToolResult(t *testing.T) {
	long := strings.Repeat("é", maxToolResultBytes)
	got := truncateToolResult(long)
	if len(got) > maxToolResultBytes+len("…[truncated]") || !strings.HasSuffix(got, "[truncated]") {
		t.Fatalf("bad truncation len=%d", len(got))
	}
	if truncateToolResult("ok") != "ok" {
		t.Fatal("short result changed")
	}
}
