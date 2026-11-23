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
