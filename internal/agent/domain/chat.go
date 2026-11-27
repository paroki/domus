package domain

import "github.com/google/uuid"

// PageContextInput describes the active UI route the user is viewing.
type PageContextInput struct {
	Path  string `json:"path"`
	Title string `json:"title,omitempty"`
}

// ChatRequest is the payload for POST /api/agent/chat.
type ChatRequest struct {
	SessionID   *uuid.UUID        `json:"sessionId,omitempty"`
	Message     string            `json:"message" validate:"required"`
	Retry       bool              `json:"retry,omitempty"`
	PageContext *PageContextInput `json:"pageContext,omitempty"`
}

// QuestionOption represents a selectable choice in an agent_ask_user prompt.
type QuestionOption struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// QuestionPayload is emitted via SSE event "question" and persisted in message meta.
type QuestionPayload struct {
	Question    string           `json:"question"`
	Options     []QuestionOption `json:"options"`
	MultiSelect bool             `json:"multiSelect"`
	AllowCustom bool             `json:"allowCustom"`
}

// NavigationPayload is emitted via SSE event "navigate" and persisted in message meta.
type NavigationPayload struct {
	Path         string `json:"path"`
	Label        string `json:"label"`
	Description  string `json:"description,omitempty"`
	AutoRedirect bool   `json:"autoRedirect"`
}
