package domain

// Core operational limits and defaults for the AI agent feature.
const (
	MaxSteps            = 8
	HistoryWindow       = 20
	MaxPinnedMemories   = 5
	MaxActiveMemories   = 100
	MaxMemoryContentLen = 500
	EmbeddingDimensions = 768
	RRFConstantK        = 60
	DefaultTemperature  = 0.1
)

// Channel identifies where an agent conversation takes place.
type Channel string

const (
	ChannelWeb      Channel = "web"
	ChannelTelegram Channel = "telegram"
	ChannelWhatsApp Channel = "whatsapp"
)

// Role identifies the author of an AgentMessage turn.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// MemoryScope distinguishes user-private facts from shared parish FAQs.
type MemoryScope string

const (
	MemoryScopeUser   MemoryScope = "user"
	MemoryScopeParish MemoryScope = "parish"
)

// MemoryCategory classifies stored long-term agent memories.
type MemoryCategory string

const (
	MemoryCategoryProfile    MemoryCategory = "profile"
	MemoryCategoryPreference MemoryCategory = "preference"
	MemoryCategoryMinistry   MemoryCategory = "ministry"
	MemoryCategoryParishFAQ  MemoryCategory = "parish_faq"
)

// ReasoningEffort controls model thinking budget when supported.
type ReasoningEffort string

const (
	ReasoningNone   ReasoningEffort = "none"
	ReasoningLow    ReasoningEffort = "low"
	ReasoningMedium ReasoningEffort = "medium"
	ReasoningHigh   ReasoningEffort = "high"
)

// ProviderName identifies a configured LLM provider backend.
type ProviderName string

const (
	ProviderGemini     ProviderName = "gemini"
	ProviderAnthropic  ProviderName = "anthropic"
	ProviderOpenAI     ProviderName = "openai"
	ProviderOpenRouter ProviderName = "openrouter"
	ProviderNone       ProviderName = "none"
)

// ModelConfig configures an LLM driver instance.
type ModelConfig struct {
	Provider        string
	BaseURL         string
	APIKey          string
	Model           string
	ReasoningEffort ReasoningEffort
	Temperature     float64
}

// AgentConfig defines a runnable AI agent persona and its toolset.
type AgentConfig struct {
	Name        string
	DisplayName string
	Description string
	Instruction string
	Model       ModelConfig
	Tools       []string
	MaxSteps    int
}

// ProviderModelItem describes a discovered model from an upstream LLM provider.
type ProviderModelItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities,omitempty"`
}
