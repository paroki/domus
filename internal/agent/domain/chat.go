package domain

import "github.com/google/uuid"

// ChatRequest is the payload for POST /api/agent/chat.
type ChatRequest struct {
	SessionID *uuid.UUID `json:"sessionId,omitempty"`
	Message   string     `json:"message" validate:"required"`
	Retry     bool       `json:"retry,omitempty"`
}
