package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Session is the domain representation of an AI agent conversation session.
type Session struct {
	ID            uuid.UUID
	ParishID      uuid.UUID
	UserID        uuid.UUID
	Channel       Channel
	Title         string
	LastMessageAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ToolCall represents a serialized LLM tool invocation on a Message.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
	AgentName string
}

// Message is the domain representation of a single turn in a Session.
type Message struct {
	ID         uuid.UUID
	ParishID   uuid.UUID
	UserID     uuid.UUID
	SessionID  uuid.UUID
	AgentName  string
	Role       Role
	Content    string
	Thinking   *string
	ToolCalls  []ToolCall
	ToolCallID *string
	ToolName   *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Memory is the domain representation of a long-term user or parish fact.
type Memory struct {
	ID              uuid.UUID
	ParishID        uuid.UUID
	UserID          *uuid.UUID
	Scope           MemoryScope
	Category        MemoryCategory
	Content         string
	IsPinned        bool
	SourceSessionID *uuid.UUID
	Score           float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CreateSessionInput holds fields for creating an AgentSession.
type CreateSessionInput struct {
	ParishID uuid.UUID
	UserID   uuid.UUID
	Channel  Channel
	Title    string
}

// SearchSessionsInput holds filters and cursor pagination for session search.
type SearchSessionsInput struct {
	ParishID uuid.UUID
	UserID   uuid.UUID
	Channel  string
	CursorAt *time.Time
	CursorID *uuid.UUID
	Limit    int
}

// CreateMessageInput holds fields for appending a message turn to a session.
type CreateMessageInput struct {
	ParishID   uuid.UUID
	UserID     uuid.UUID
	SessionID  uuid.UUID
	AgentName  string
	Role       Role
	Content    string
	Thinking   *string
	ToolCalls  []ToolCall
	ToolCallID *string
	ToolName   *string
}

// SearchSessionMessagesInput holds filters and cursor pagination for a single session.
type SearchSessionMessagesInput struct {
	ParishID  uuid.UUID
	UserID    uuid.UUID
	SessionID uuid.UUID
	Q         string
	CursorAt  *time.Time
	CursorID  *uuid.UUID
	Limit     int
}

// SearchUserHistoryInput holds filters and cursor pagination across a user's sessions.
type SearchUserHistoryInput struct {
	ParishID uuid.UUID
	UserID   uuid.UUID
	Q        string
	CursorAt *time.Time
	CursorID *uuid.UUID
	Limit    int
}

// SaveMemoryInput holds fields for creating (and optionally superseding) an AgentMemory.
type SaveMemoryInput struct {
	ParishID        uuid.UUID
	UserID          *uuid.UUID
	Scope           MemoryScope
	Category        MemoryCategory
	Content         string
	IsPinned        bool
	SourceSessionID *uuid.UUID
	SupersedeID     *uuid.UUID
	Embedding       []float32
}

// UpdateMemoryInput holds fields for updating an existing AgentMemory.
type UpdateMemoryInput struct {
	ID             uuid.UUID
	ParishID       uuid.UUID
	UserID         uuid.UUID
	Category       *MemoryCategory
	Content        *string
	IsPinned       *bool
	Embedding      []float32
	ClearEmbedding bool
}

// RecallMemoriesInput holds parameters for Tier 3 hybrid memory recall.
type RecallMemoriesInput struct {
	ParishID       uuid.UUID
	UserID         uuid.UUID
	Query          string
	QueryEmbedding []float32
	Category       *MemoryCategory
	Limit          int
}

// SearchMemoriesInput holds filters and cursor pagination for the Memory Drawer.
type SearchMemoriesInput struct {
	ParishID uuid.UUID
	UserID   uuid.UUID
	Scope    string
	Category string
	IsPinned *bool
	Q        string
	CursorAt *time.Time
	CursorID *uuid.UUID
	Limit    int
}

var xmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
)

// FormatUserMemoryXML renders pinned or recalled memories as a safe XML block for prompt injection.
func FormatUserMemoryXML(memories []Memory) string {
	if len(memories) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<user_memory>\n")
	for _, m := range memories {
		fmt.Fprintf(
			&b,
			`  <fact id="%s" category="%s" scope="%s" pinned="%s">%s</fact>`+"\n",
			xmlEscaper.Replace(m.ID.String()),
			xmlEscaper.Replace(string(m.Category)),
			xmlEscaper.Replace(string(m.Scope)),
			strconv.FormatBool(m.IsPinned),
			xmlEscaper.Replace(m.Content),
		)
	}
	b.WriteString("</user_memory>")
	return b.String()
}
