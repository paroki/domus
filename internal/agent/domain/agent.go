package domain

// ProviderName identifies a configured LLM provider backend.
type ProviderName string

const (
	ProviderGemini     ProviderName = "gemini"
	ProviderAnthropic  ProviderName = "anthropic"
	ProviderOpenAI     ProviderName = "openai"
	ProviderOpenRouter ProviderName = "openrouter"
	ProviderNone       ProviderName = "none"
)

// ProviderModelItem describes a discovered model from an upstream LLM provider.
type ProviderModelItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities,omitempty"`
}
