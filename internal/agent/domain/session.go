package domain

import (
	"time"

	"github.com/google/uuid"
)

// ToolCallRecord represents a tool invocation recorded on an AgentMessage.
type ToolCallRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments,omitempty"`
}

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
	ID         uuid.UUID        `json:"id"`
	SessionID  uuid.UUID        `json:"sessionId"`
	ParishID   uuid.UUID        `json:"parishId"`
	UserID     uuid.UUID        `json:"userId"`
	AgentName  string           `json:"agentName"`
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []ToolCallRecord `json:"toolCalls,omitempty"`
	ToolCallID *string          `json:"toolCallId,omitempty"`
	ToolName   *string          `json:"toolName,omitempty"`
	Meta       map[string]any   `json:"meta,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
	UpdatedAt  time.Time        `json:"updatedAt"`
}

// AppendMessageInput describes a message to append to an AgentSession.
type AppendMessageInput struct {
	AgentName  string
	Role       string
	Content    string
	ToolCalls  []ToolCallRecord
	ToolCallID *string
	ToolName   *string
	Meta       map[string]any
}

// SearchSessionsRequest searches caller's active AI sessions.
type SearchSessionsRequest struct {
	Q      string `json:"q,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// SessionDetailResponse returns a session with its chronological messages and resolved effective model.
type SessionDetailResponse struct {
	ID                  uuid.UUID         `json:"id"`
	ParishID            uuid.UUID         `json:"parishId"`
	UserID              uuid.UUID         `json:"userId"`
	Title               string            `json:"title"`
	ProviderID          *uuid.UUID        `json:"providerId,omitempty"`
	ModelID             *string           `json:"modelId,omitempty"`
	EffectiveProviderID *uuid.UUID        `json:"effectiveProviderId,omitempty"`
	EffectiveModelID    string            `json:"effectiveModelId,omitempty"`
	LastMessageAt       time.Time         `json:"lastMessageAt"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
	Messages            []MessageResponse `json:"messages"`
}

// UpdateSessionRequest updates session title and/or model override.
// Sending "" on providerId or modelId clears both provider_id and model_id back to NULL.
type UpdateSessionRequest struct {
	Title      *string `json:"title,omitempty"`
	ProviderID *string `json:"providerId,omitempty"`
	ModelID    *string `json:"modelId,omitempty"`
}
