package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProviderModelDTO describes an available model returned by provider model sync.
type ProviderModelDTO struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// SyncProviderModelsResponse returns discovered models for a provider config.
type SyncProviderModelsResponse struct {
	ID       uuid.UUID          `json:"id"`
	Provider string             `json:"provider"`
	Models   []ProviderModelDTO `json:"models"`
}

// AgentProviderResponse is the public API representation of an AgentProvider config.
type AgentProviderResponse struct {
	ID           uuid.UUID          `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Provider     string             `json:"provider"`
	BaseURL      string             `json:"baseUrl"`
	HasAPIKey    bool               `json:"hasApiKey"`
	APIKeyMasked string             `json:"apiKeyMasked"`
	Models       []ProviderModelDTO `json:"models"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// CreateAgentProviderRequest creates a new LLM provider configuration.
type CreateAgentProviderRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
	Provider    string `json:"provider" validate:"required,oneof=gemini anthropic openai openrouter"`
	APIKey      string `json:"apiKey" validate:"required,min=1"`
	BaseURL     string `json:"baseUrl,omitempty" validate:"omitempty,http_url"`
}

// UpdateAgentProviderRequest updates an existing LLM provider configuration.
type UpdateAgentProviderRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Provider    *string `json:"provider,omitempty" validate:"omitempty,oneof=gemini anthropic openai openrouter"`
	APIKey      *string `json:"apiKey,omitempty"`
	BaseURL     *string `json:"baseUrl,omitempty"`
}

// AgentSessionResponse is the public API representation of an AgentSession.
type AgentSessionResponse struct {
	ID            uuid.UUID `json:"id"`
	ParishID      uuid.UUID `json:"parishId"`
	UserID        uuid.UUID `json:"userId"`
	Channel       string    `json:"channel"`
	Title         string    `json:"title"`
	LastMessageAt time.Time `json:"lastMessageAt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// AgentSessionSearchFilter filters agent session search queries.
type AgentSessionSearchFilter struct {
	Channel string `json:"channel,omitempty" validate:"omitempty,oneof=web telegram whatsapp"`
}

// SearchAgentSessionsRequest searches the caller's agent sessions.
type SearchAgentSessionsRequest struct {
	Filter *AgentSessionSearchFilter `json:"filter,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
	Limit  int                       `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateAgentSessionRequest creates a new web agent conversation session.
type CreateAgentSessionRequest struct {
	Title *string `json:"title,omitempty" validate:"omitempty,min=1,max=120"`
}

// UpdateAgentSessionRequest renames an existing agent conversation session.
type UpdateAgentSessionRequest struct {
	Title string `json:"title" validate:"required,min=1,max=120"`
}

// AgentToolCallDTO represents a serialized tool call on an AgentMessageResponse.
type AgentToolCallDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	AgentName string `json:"agentName,omitempty"`
}

// AgentMessageResponse is the public API representation of an AgentMessage.
type AgentMessageResponse struct {
	ID         uuid.UUID          `json:"id"`
	SessionID  uuid.UUID          `json:"sessionId"`
	AgentName  string             `json:"agentName"`
	Role       string             `json:"role"`
	Content    string             `json:"content"`
	Thinking   *string            `json:"thinking,omitempty"`
	ToolCalls  []AgentToolCallDTO `json:"toolCalls,omitempty"`
	ToolCallID *string            `json:"toolCallId,omitempty"`
	ToolName   *string            `json:"toolName,omitempty"`
	CreatedAt  time.Time          `json:"createdAt"`
}

// SearchAgentMessagesRequest searches messages within a session.
type SearchAgentMessagesRequest struct {
	Q      string `json:"q,omitempty"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// AgentMemoryResponse is the public API representation of an AgentMemory.
type AgentMemoryResponse struct {
	ID              uuid.UUID  `json:"id"`
	ParishID        uuid.UUID  `json:"parishId"`
	UserID          *uuid.UUID `json:"userId,omitempty"`
	Scope           string     `json:"scope"`
	Category        string     `json:"category"`
	Content         string     `json:"content"`
	IsPinned        bool       `json:"isPinned"`
	SourceSessionID *uuid.UUID `json:"sourceSessionId,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// AgentMemorySearchFilter filters memory drawer search queries.
type AgentMemorySearchFilter struct {
	Scope    string `json:"scope,omitempty" validate:"omitempty,oneof=user parish"`
	Category string `json:"category,omitempty" validate:"omitempty,oneof=profile preference ministry parish_faq"`
	IsPinned *bool  `json:"isPinned,omitempty"`
}

// SearchAgentMemoriesRequest searches long-term agent memories.
type SearchAgentMemoriesRequest struct {
	Q      string                   `json:"q,omitempty"`
	Filter *AgentMemorySearchFilter `json:"filter,omitempty"`
	Cursor string                   `json:"cursor,omitempty"`
	Limit  int                      `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateAgentMemoryRequest creates a new long-term agent memory fact.
type CreateAgentMemoryRequest struct {
	Scope       string     `json:"scope" validate:"required,oneof=user parish"`
	Category    string     `json:"category" validate:"required,oneof=profile preference ministry parish_faq"`
	Content     string     `json:"content" validate:"required,min=1,max=500"`
	IsPinned    *bool      `json:"isPinned,omitempty"`
	SupersedeID *uuid.UUID `json:"supersedeId,omitempty"`
}

// UpdateAgentMemoryRequest updates an existing long-term agent memory fact.
type UpdateAgentMemoryRequest struct {
	Category *string `json:"category,omitempty" validate:"omitempty,oneof=profile preference ministry parish_faq"`
	Content  *string `json:"content,omitempty" validate:"omitempty,min=1,max=500"`
	IsPinned *bool   `json:"isPinned,omitempty"`
}

// AgentChatRequest sends a user message to the AI agent chat SSE endpoint.
type AgentChatRequest struct {
	SessionID *uuid.UUID `json:"sessionId,omitempty"`
	Message   string     `json:"message" validate:"required,min=1"`
}

// AgentSSEEvent represents a single Server-Sent Event frame emitted during agent chat.
type AgentSSEEvent struct {
	Event      string     `json:"event"`
	SessionID  *uuid.UUID `json:"sessionId,omitempty"`
	AgentName  string     `json:"agentName,omitempty"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"toolCallId,omitempty"`
	ToolName   string     `json:"toolName,omitempty"`
	Arguments  string     `json:"arguments,omitempty"`
	Code       string     `json:"code,omitempty"`
	Message    string     `json:"message,omitempty"`
}
