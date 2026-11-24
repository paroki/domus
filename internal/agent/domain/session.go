package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/schema"
)

// SessionResponse represents a persisted AI conversation session.
type SessionResponse struct {
	ID            uuid.UUID  `json:"id"`
	ParishID      uuid.UUID  `json:"parishId"`
	UserID        uuid.UUID  `json:"userId"`
	Title         string     `json:"title"`
	ProviderID    *uuid.UUID `json:"providerId,omitempty"`
	ModelID       *string    `json:"modelId,omitempty"`
	LastMessageAt time.Time  `json:"lastMessageAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// MessageResponse represents a persisted message in an AgentSession.
type MessageResponse struct {
	ID         uuid.UUID               `json:"id"`
	SessionID  uuid.UUID               `json:"sessionId"`
	ParishID   uuid.UUID               `json:"parishId"`
	UserID     uuid.UUID               `json:"userId"`
	AgentName  string                  `json:"agentName"`
	Role       string                  `json:"role"`
	Content    string                  `json:"content"`
	ToolCalls  []schema.ToolCallRecord `json:"toolCalls,omitempty"`
	ToolCallID *string                 `json:"toolCallId,omitempty"`
	ToolName   *string                 `json:"toolName,omitempty"`
	Meta       map[string]any          `json:"meta,omitempty"`
	CreatedAt  time.Time               `json:"createdAt"`
	UpdatedAt  time.Time               `json:"updatedAt"`
}

// AppendMessageInput describes a message to append to an AgentSession.
type AppendMessageInput struct {
	AgentName  string
	Role       string
	Content    string
	ToolCalls  []schema.ToolCallRecord
	ToolCallID *string
	ToolName   *string
	Meta       map[string]any
}
