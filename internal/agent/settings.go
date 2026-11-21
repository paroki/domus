package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/setting"
)

// Exported setting keys for AI Agent providers, personas, and embeddings (26 keys).
const (
	SettingProviderGeminiAPIKey  = "agent.provider.gemini.api_key"
	SettingProviderGeminiBaseURL = "agent.provider.gemini.base_url"
	SettingProviderGeminiModels  = "agent.provider.gemini.models"

	SettingProviderAnthropicAPIKey  = "agent.provider.anthropic.api_key"
	SettingProviderAnthropicBaseURL = "agent.provider.anthropic.base_url"
	SettingProviderAnthropicModels  = "agent.provider.anthropic.models"

	SettingProviderOpenAIAPIKey  = "agent.provider.openai.api_key"
	SettingProviderOpenAIBaseURL = "agent.provider.openai.base_url"
	SettingProviderOpenAIModels  = "agent.provider.openai.models"

	SettingProviderOpenRouterAPIKey  = "agent.provider.openrouter.api_key"
	SettingProviderOpenRouterBaseURL = "agent.provider.openrouter.base_url"
	SettingProviderOpenRouterModels  = "agent.provider.openrouter.models"

	SettingGabrielProvider    = "agent.gabriel.provider"
	SettingGabrielModel       = "agent.gabriel.model"
	SettingGabrielEffort      = "agent.gabriel.effort"
	SettingGabrielInstruction = "agent.gabriel.instruction"

	SettingMikaelProvider    = "agent.mikael.provider"
	SettingMikaelModel       = "agent.mikael.model"
	SettingMikaelEffort      = "agent.mikael.effort"
	SettingMikaelInstruction = "agent.mikael.instruction"

	SettingRafaelProvider    = "agent.rafael.provider"
	SettingRafaelModel       = "agent.rafael.model"
	SettingRafaelEffort      = "agent.rafael.effort"
	SettingRafaelInstruction = "agent.rafael.instruction"

	SettingEmbeddingProvider = "agent.embedding.provider"
	SettingEmbeddingModel    = "agent.embedding.model"
)

// EmbeddingRuntimeConfig holds the resolved runtime settings for text embeddings.
type EmbeddingRuntimeConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

// ConfigStore holds thread-safe hot-reloaded AI agent, embedding, and provider model settings.
type ConfigStore struct {
	mu             sync.RWMutex
	agents         map[string]domain.AgentConfig
	embedding      EmbeddingRuntimeConfig
	providerModels map[string][]domain.ProviderModelItem
	prevAPIKeys    map[string]string
	initialized    bool
	drivers        *driver.Registry
	logger         *slog.Logger
}

// NewConfigStore creates a ConfigStore backed by drivers and logger.
func NewConfigStore(drivers *driver.Registry, logger *slog.Logger) *ConfigStore {
	if drivers == nil {
		drivers = driver.NewRegistry(nil)
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &ConfigStore{
		agents:         make(map[string]domain.AgentConfig),
		providerModels: make(map[string][]domain.ProviderModelItem),
		prevAPIKeys:    make(map[string]string),
		drivers:        drivers,
		logger:         logger,
	}
}

// AgentConfig returns a snapshot of the runtime configuration for the named agent.
func (c *ConfigStore) AgentConfig(name string) (domain.AgentConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cfg, ok := c.agents[name]
	return cfg, ok
}

// EmbeddingConfig returns a snapshot of the active embedding configuration.
func (c *ConfigStore) EmbeddingConfig() EmbeddingRuntimeConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.embedding
}

// ProviderModels returns cached discovered models for a provider.
func (c *ConfigStore) ProviderModels(provider string) []domain.ProviderModelItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Clone(c.providerModels[provider])
}

// SetProviderModels updates the in-memory discovered models cache for a provider.
func (c *ConfigStore) SetProviderModels(provider string, models []domain.ProviderModelItem) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providerModels[provider] = slices.Clone(models)
}

type providerDef struct {
	name           string
	label          string
	group          string
	apiKeyKey      string
	baseURLKey     string
	defaultBaseURL string
	modelsKey      string
}

var providerDefs = []providerDef{
	{
		name:           "gemini",
		label:          "Gemini",
		group:          "AI Provider: Gemini",
		apiKeyKey:      SettingProviderGeminiAPIKey,
		baseURLKey:     SettingProviderGeminiBaseURL,
		defaultBaseURL: "",
		modelsKey:      SettingProviderGeminiModels,
	},
	{
		name:           "anthropic",
		label:          "Anthropic",
		group:          "AI Provider: Anthropic",
		apiKeyKey:      SettingProviderAnthropicAPIKey,
		baseURLKey:     SettingProviderAnthropicBaseURL,
		defaultBaseURL: "",
		modelsKey:      SettingProviderAnthropicModels,
	},
	{
		name:           "openai",
		label:          "OpenAI",
		group:          "AI Provider: OpenAI",
		apiKeyKey:      SettingProviderOpenAIAPIKey,
		baseURLKey:     SettingProviderOpenAIBaseURL,
		defaultBaseURL: "https://api.openai.com/v1",
		modelsKey:      SettingProviderOpenAIModels,
	},
	{
		name:           "openrouter",
		label:          "OpenRouter",
		group:          "AI Provider: OpenRouter",
		apiKeyKey:      SettingProviderOpenRouterAPIKey,
		baseURLKey:     SettingProviderOpenRouterBaseURL,
		defaultBaseURL: "https://openrouter.ai/api/v1",
		modelsKey:      SettingProviderOpenRouterModels,
	},
}

type agentDef struct {
	name               string
	displayName        string
	description        string
	group              string
	providerKey        string
	modelKey           string
	effortKey          string
	instructionKey     string
	defaultInstruction string
	tools              []string
}

var agentDefs = []agentDef{
	{
		name:               "gabriel",
		displayName:        "Gabriel",
		description:        "Asisten AI utama dan orkestrator Domus.",
		group:              "AI Agent: Gabriel",
		providerKey:        SettingGabrielProvider,
		modelKey:           SettingGabrielModel,
		effortKey:          SettingGabrielEffort,
		instructionKey:     SettingGabrielInstruction,
		defaultInstruction: prompts.GabrielInstruction,
		tools:              []string{"memory_save", "memory_recall", "memory_forget", "memory_search_history", "delegate_specialist"},
	},
	{
		name:               "mikael",
		displayName:        "Mikael",
		description:        "Spesialis kegiatan, jadwal acara, dan RSVP paroki.",
		group:              "AI Agent: Mikael",
		providerKey:        SettingMikaelProvider,
		modelKey:           SettingMikaelModel,
		effortKey:          SettingMikaelEffort,
		instructionKey:     SettingMikaelInstruction,
		defaultInstruction: prompts.MikaelInstruction,
		tools:              []string{"event_list", "event_get", "event_create", "event_update", "event_rsvp", "event_list_rsvps"},
	},
	{
		name:               "rafael",
		displayName:        "Rafael",
		description:        "Spesialis organisasi, keanggotaan, dan tautan undangan.",
		group:              "AI Agent: Rafael",
		providerKey:        SettingRafaelProvider,
		modelKey:           SettingRafaelModel,
		effortKey:          SettingRafaelEffort,
		instructionKey:     SettingRafaelInstruction,
		defaultInstruction: prompts.RafaelInstruction,
		tools:              []string{"org_list", "org_get", "org_list_members", "org_update_member", "org_create_invite", "org_revoke_invite", "org_redeem_invite"},
	},
}

// RegisterSettings registers all 26 AI agent descriptors and a hot-reload listener on s.
func (c *ConfigStore) RegisterSettings(s *setting.Store) {
	for _, p := range providerDefs {
		s.Register(setting.Descriptor{
			Key:         p.apiKeyKey,
			Group:       p.group,
			Label:       p.label + " API Key",
			Description: "Kunci API untuk penyedia LLM " + p.label + ".",
			Secret:      true,
			Placeholder: "sk-...",
		})
		s.Register(setting.Descriptor{
			Key:          p.baseURLKey,
			Group:        p.group,
			Label:        p.label + " Base URL",
			Description:  "URL dasar endpoint API " + p.label + " (opsional jika menggunakan endpoint standar).",
			DefaultValue: p.defaultBaseURL,
			Placeholder:  p.defaultBaseURL,
		})
		s.Register(setting.Descriptor{
			Key:          p.modelsKey,
			Group:        p.group,
			Label:        p.label + " Models",
			Description:  "Daftar model tersinkronisasi dalam format array JSON.",
			DefaultValue: "[]",
		})
	}

	for _, a := range agentDefs {
		s.Register(setting.Descriptor{
			Key:          a.providerKey,
			Group:        a.group,
			Label:        "Provider",
			Description:  "Penyedia LLM aktif untuk agen " + a.displayName + ".",
			DefaultValue: "gemini",
		})
		s.Register(setting.Descriptor{
			Key:          a.modelKey,
			Group:        a.group,
			Label:        "Model",
			Description:  "Nama model LLM untuk agen " + a.displayName + ".",
			DefaultValue: "gemini-2.5-flash",
		})
		s.Register(setting.Descriptor{
			Key:          a.effortKey,
			Group:        a.group,
			Label:        "Reasoning Effort",
			Description:  "Tingkat penalaran model (none, low, medium, high).",
			DefaultValue: "low",
		})
		s.Register(setting.Descriptor{
			Key:          a.instructionKey,
			Group:        a.group,
			Label:        "Instruction",
			Description:  "System prompt persona untuk agen " + a.displayName + ".",
			DefaultValue: a.defaultInstruction,
		})
	}

	s.Register(setting.Descriptor{
		Key:          SettingEmbeddingProvider,
		Group:        "AI Embedding",
		Label:        "Embedding Provider",
		Description:  "Penyedia model embedding (gemini, openai, openrouter, none).",
		DefaultValue: "gemini",
	})
	s.Register(setting.Descriptor{
		Key:          SettingEmbeddingModel,
		Group:        "AI Embedding",
		Label:        "Embedding Model",
		Description:  "Nama model embedding 768-dimensi.",
		DefaultValue: "text-embedding-004",
	})

	s.OnChange(func(ctx context.Context, values map[string]string) error {
		return c.applySettings(ctx, s, values)
	})
}

func (c *ConfigStore) applySettings(ctx context.Context, store *setting.Store, values map[string]string) error {
	providerAPIKeys := make(map[string]string, len(providerDefs))
	providerBaseURLs := make(map[string]string, len(providerDefs))
	nextModels := make(map[string][]domain.ProviderModelItem, len(providerDefs))

	for _, p := range providerDefs {
		apiKey := strings.TrimSpace(values[p.apiKeyKey])
		baseURL := strings.TrimSpace(values[p.baseURLKey])
		if baseURL != "" {
			if err := validateHTTPURL(baseURL); err != nil {
				return fmt.Errorf("%s: %w", p.baseURLKey, err)
			}
		}
		modelsJSON := strings.TrimSpace(values[p.modelsKey])
		if modelsJSON == "" {
			modelsJSON = "[]"
		}
		var parsedModels []domain.ProviderModelItem
		if err := json.Unmarshal([]byte(modelsJSON), &parsedModels); err != nil {
			return fmt.Errorf("%s harus berupa array JSON yang valid", p.modelsKey)
		}
		providerAPIKeys[p.name] = apiKey
		providerBaseURLs[p.name] = baseURL
		nextModels[p.name] = parsedModels
	}

	nextAgents := make(map[string]domain.AgentConfig, len(agentDefs))
	for _, a := range agentDefs {
		prov := strings.TrimSpace(values[a.providerKey])
		if !isValidAgentProvider(prov) {
			return fmt.Errorf("%s tidak valid: %q", a.providerKey, prov)
		}
		model := strings.TrimSpace(values[a.modelKey])
		if model == "" {
			return fmt.Errorf("%s wajib diisi", a.modelKey)
		}
		effort := domain.ReasoningEffort(strings.TrimSpace(values[a.effortKey]))
		if !isValidEffort(effort) {
			return fmt.Errorf("%s tidak valid: %q", a.effortKey, effort)
		}
		instr := strings.TrimSpace(values[a.instructionKey])
		if instr == "" {
			return fmt.Errorf("%s wajib diisi", a.instructionKey)
		}

		nextAgents[a.name] = domain.AgentConfig{
			Name:        a.name,
			DisplayName: a.displayName,
			Description: a.description,
			Instruction: instr,
			Model: domain.ModelConfig{
				Provider:        prov,
				BaseURL:         providerBaseURLs[prov],
				APIKey:          providerAPIKeys[prov],
				Model:           model,
				ReasoningEffort: effort,
				Temperature:     domain.DefaultTemperature,
			},
			Tools:    slices.Clone(a.tools),
			MaxSteps: domain.MaxSteps,
		}
	}

	embProv := strings.TrimSpace(values[SettingEmbeddingProvider])
	if !isValidEmbeddingProvider(embProv) {
		return fmt.Errorf("%s tidak valid: %q", SettingEmbeddingProvider, embProv)
	}
	embModel := strings.TrimSpace(values[SettingEmbeddingModel])
	if embModel == "" {
		return fmt.Errorf("%s wajib diisi", SettingEmbeddingModel)
	}
	nextEmbedding := EmbeddingRuntimeConfig{
		Provider: embProv,
		BaseURL:  providerBaseURLs[embProv],
		APIKey:   providerAPIKeys[embProv],
		Model:    embModel,
	}

	c.mu.Lock()
	wasInitialized := c.initialized
	prevKeys := maps.Clone(c.prevAPIKeys)
	c.agents = nextAgents
	c.embedding = nextEmbedding
	c.providerModels = nextModels
	c.prevAPIKeys = maps.Clone(providerAPIKeys)
	c.initialized = true
	c.mu.Unlock()

	if !wasInitialized {
		return nil
	}

	// Best-effort /models auto-sync on api_key transitions after initial Load.
	for _, p := range providerDefs {
		newKey := providerAPIKeys[p.name]
		oldKey := prevKeys[p.name]
		if newKey == oldKey {
			continue
		}
		if newKey == "" {
			if err := store.PersistValue(ctx, p.modelsKey, "[]"); err != nil {
				c.logger.WarnContext(ctx, "clear provider models failed", "provider", p.name, "err", err)
			}
			c.mu.Lock()
			c.providerModels[p.name] = nil
			c.mu.Unlock()
			continue
		}
		syncCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		discovered, err := c.drivers.DiscoverModels(syncCtx, p.name, providerBaseURLs[p.name], newKey)
		cancel()
		if err != nil {
			c.logger.WarnContext(ctx, "auto-sync provider models failed", "provider", p.name, "err", err)
			continue
		}
		b, err := json.Marshal(discovered)
		if err != nil {
			c.logger.WarnContext(ctx, "marshal discovered models failed", "provider", p.name, "err", err)
			continue
		}
		if err := store.PersistValue(ctx, p.modelsKey, string(b)); err != nil {
			c.logger.WarnContext(ctx, "persist discovered models failed", "provider", p.name, "err", err)
			continue
		}
		c.mu.Lock()
		c.providerModels[p.name] = discovered
		c.mu.Unlock()
	}
	return nil
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("URL harus diawali http:// atau https:// dan memiliki host yang valid")
	}
	return nil
}

func isValidAgentProvider(p string) bool {
	switch domain.ProviderName(p) {
	case domain.ProviderGemini, domain.ProviderAnthropic, domain.ProviderOpenAI, domain.ProviderOpenRouter:
		return true
	default:
		return false
	}
}

func isValidEmbeddingProvider(p string) bool {
	switch domain.ProviderName(p) {
	case domain.ProviderGemini, domain.ProviderOpenAI, domain.ProviderOpenRouter, domain.ProviderNone:
		return true
	default:
		return false
	}
}

func isValidEffort(e domain.ReasoningEffort) bool {
	switch e {
	case domain.ReasoningNone, domain.ReasoningLow, domain.ReasoningMedium, domain.ReasoningHigh:
		return true
	default:
		return false
	}
}
