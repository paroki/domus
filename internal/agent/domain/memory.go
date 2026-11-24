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
	ParishID        uuid.UUID  `json:"parishId"`
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
	ParishID        uuid.UUID
	UserID          uuid.UUID
	Category        string
	Content         string
	SourceSessionID *uuid.UUID
}
