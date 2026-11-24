package service

import (
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/internal/agent/domain"
)

const (
	maxHistoryMessages = 40
	maxToolNoteRunes   = 300
)

// ProjectHistory converts up to the last 40 persisted AgentMessage rows into clean
// cross-turn user/assistant messages, folding tool results into compact assistant notes
// and coalescing consecutive user messages with "\n\n".
func ProjectHistory(rows []domain.MessageResponse) []*schema.Message {
	if len(rows) > maxHistoryMessages {
		rows = rows[len(rows)-maxHistoryMessages:]
	}

	var out []*schema.Message
	for _, row := range rows {
		switch row.Role {
		case "user":
			text := strings.TrimSpace(row.Content)
			if text == "" {
				continue
			}
			if n := len(out); n > 0 && out[n-1].Role == schema.User {
				out[n-1].Content += "\n\n" + text
			} else {
				out = append(out, schema.UserMessage(text))
			}
		case "assistant":
			text := strings.TrimSpace(row.Content)
			if text == "" {
				continue
			}
			out = append(out, schema.AssistantMessage(text, nil))
		case "tool":
			name := "tool"
			if row.ToolName != nil && strings.TrimSpace(*row.ToolName) != "" {
				name = strings.TrimSpace(*row.ToolName)
			}
			note := fmt.Sprintf("[Tool %s]: %s", name, truncateRunes(strings.TrimSpace(row.Content), maxToolNoteRunes))
			if n := len(out); n > 0 && out[n-1].Role == schema.Assistant {
				out[n-1].Content += "\n" + note
			} else {
				out = append(out, schema.AssistantMessage(note, nil))
			}
		}
	}
	return out
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
