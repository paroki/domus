package domain

import (
	"time"

	"github.com/google/uuid"
)

// Memory category constants matching AgentMemory.category enum.
const (
	MemoryCategoryPreference = "preference"
	MemoryCategoryFact       = "fact"
	MemoryCategoryContext    = "context"
)

// MemoryResponse represents a persisted long-term AI memory item.
type MemoryResponse struct {
	ID              uuid.UUID  `json:"id"`
	ParokiID        uuid.UUID  `json:"parokiId"`
	UserID          uuid.UUID  `json:"userId"`
	Category        string     `json:"category"`
	Content         string     `json:"content"`
	SourceSessionID *uuid.UUID `json:"sourceSessionId,omitempty"`
	Score           float64    `json:"score,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// CreateMemoryInput describes a new AgentMemory record.
type CreateMemoryInput struct {
	ParokiID        uuid.UUID
	UserID          uuid.UUID
	Category        string
	Content         string
	SourceSessionID *uuid.UUID
}

// MemorySearchFilter filters memory search results.
type MemorySearchFilter struct {
	Category string `json:"category,omitempty" validate:"omitempty,oneof=preference fact context"`
}

// SearchMemoriesRequest searches or lists caller's long-term AI memories.
type SearchMemoriesRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *MemorySearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateMemoryRequest creates a new AgentMemory item.
type CreateMemoryRequest struct {
	Content  string `json:"content" validate:"required"`
	Category string `json:"category,omitempty" validate:"omitempty,oneof=preference fact context"`
}

// UpdateMemoryRequest updates an existing AgentMemory item.
type UpdateMemoryRequest struct {
	Content  *string `json:"content,omitempty"`
	Category *string `json:"category,omitempty" validate:"omitempty,oneof=preference fact context"`
}
