package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
)

// Exported setting keys for AI Agent persona and embeddings (6 keys).
const (
	SettingGabrielProviderID  = "agent.gabriel.provider_id"
	SettingGabrielModel       = "agent.gabriel.model"
	SettingGabrielEffort      = "agent.gabriel.effort"
	SettingGabrielInstruction = "agent.gabriel.instruction"

	SettingEmbeddingProviderID = "agent.embedding.provider_id"
	SettingEmbeddingModel      = "agent.embedding.model"
)

// ProviderRuntimeConfig holds a decrypted LLM provider configuration in memory.
type ProviderRuntimeConfig struct {
	ID       uuid.UUID
	Name     string
	Provider string
	BaseURL  string
	APIKey   string
	Models   []domain.ProviderModelItem
}

// EmbeddingRuntimeConfig holds the resolved runtime settings for text embeddings.
type EmbeddingRuntimeConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

// ConfigStore holds thread-safe hot-reloaded AI agent, embedding, and provider configurations.
type ConfigStore struct {
	mu                  sync.RWMutex
	db                  *ent.Client
	cipher              *crypt.Cipher
	providers           map[uuid.UUID]ProviderRuntimeConfig
	providersLoaded     bool
	agentProviderIDs    map[string]uuid.UUID
	embeddingProviderID uuid.UUID
	agents              map[string]domain.AgentConfig
	embedding           EmbeddingRuntimeConfig
	drivers             *driver.Registry
	logger              *slog.Logger
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
		providers:        make(map[uuid.UUID]ProviderRuntimeConfig),
		agentProviderIDs: make(map[string]uuid.UUID),
		agents:           make(map[string]domain.AgentConfig),
		drivers:          drivers,
		logger:           logger,
	}
}

// BindStorage attaches the Ent client and optional cipher used to load provider credentials.
func (c *ConfigStore) BindStorage(db *ent.Client, cipher *crypt.Cipher) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db = db
	c.cipher = cipher
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

// Provider returns a snapshot of the runtime provider configuration for id.
func (c *ConfigStore) Provider(id uuid.UUID) (ProviderRuntimeConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.providers[id]
	if ok {
		p.Models = slices.Clone(p.Models)
	}
	return p, ok
}

// ProviderModels returns cached discovered models for a provider config UUID.
func (c *ConfigStore) ProviderModels(id uuid.UUID) []domain.ProviderModelItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.providers[id]
	if !ok {
		return nil
	}
	return slices.Clone(p.Models)
}

// IsProviderInUse reports whether any agent or embedding config currently references id.
func (c *ConfigStore) IsProviderInUse(id uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.embeddingProviderID == id {
		return true
	}
	for _, boundID := range c.agentProviderIDs {
		if boundID == id {
			return true
		}
	}
	return false
}

// IsProviderUsedByEmbedding reports whether embedding currently references id.
func (c *ConfigStore) IsProviderUsedByEmbedding(id uuid.UUID) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.embeddingProviderID == id
}

// UpsertProvider updates the in-memory provider cache and refreshes bound agent/embedding snapshots.
func (c *ConfigStore) UpsertProvider(p ProviderRuntimeConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	p.Models = slices.Clone(p.Models)
	c.providers[p.ID] = p
	c.providersLoaded = true

	for agentName, boundID := range c.agentProviderIDs {
		if boundID == p.ID {
			if cfg, ok := c.agents[agentName]; ok {
				cfg.Model.Provider = p.Provider
				cfg.Model.BaseURL = p.BaseURL
				cfg.Model.APIKey = p.APIKey
				c.agents[agentName] = cfg
			}
		}
	}
	if c.embeddingProviderID == p.ID {
		c.embedding.Provider = p.Provider
		c.embedding.BaseURL = p.BaseURL
		c.embedding.APIKey = p.APIKey
	}
}

// DeleteProvider removes id from the in-memory provider cache.
func (c *ConfigStore) DeleteProvider(id uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.providers, id)
}

// ReloadProviders loads all agent_providers rows from DB into the in-memory cache.
func (c *ConfigStore) ReloadProviders(ctx context.Context) error {
	c.mu.RLock()
	db := c.db
	cipher := c.cipher
	c.mu.RUnlock()
	if db == nil {
		return nil
	}

	rows, err := db.AgentProvider.Query().
		Order(agentprovider.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return err
	}

	next := make(map[uuid.UUID]ProviderRuntimeConfig, len(rows))
	for _, row := range rows {
		apiKey, err := decryptAPIKey(cipher, row.APIKey)
		if err != nil {
			return fmt.Errorf("agent_provider %s: %w", row.ID, err)
		}
		models := make([]domain.ProviderModelItem, len(row.Models))
		for i, m := range row.Models {
			models[i] = domain.ProviderModelItem{
				ID:           m.ID,
				Name:         m.Name,
				Capabilities: slices.Clone(m.Capabilities),
			}
		}
		next[row.ID] = ProviderRuntimeConfig{
			ID:       row.ID,
			Name:     row.Name,
			Provider: string(row.Provider),
			BaseURL:  row.BaseURL,
			APIKey:   apiKey,
			Models:   models,
		}
	}

	c.mu.Lock()
	c.providers = next
	c.providersLoaded = true
	c.mu.Unlock()
	return nil
}

type agentDef struct {
	name               string
	displayName        string
	description        string
	group              string
	providerIDKey      string
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
		description:        "Asisten AI utama Domus untuk kegiatan, organisasi, dan memori paroki.",
		group:              "AI Agent: Gabriel",
		providerIDKey:      SettingGabrielProviderID,
		modelKey:           SettingGabrielModel,
		effortKey:          SettingGabrielEffort,
		instructionKey:     SettingGabrielInstruction,
		defaultInstruction: prompts.GabrielInstruction,
		tools: []string{
			"memory_save", "memory_recall", "memory_forget", "memory_search_history",
			"event_list", "event_get", "event_create", "event_update", "event_rsvp", "event_list_rsvps",
			"org_list", "org_get", "org_create", "org_list_members", "org_update_member", "org_create_invite", "org_revoke_invite", "org_redeem_invite",
		},
	},
}

// RegisterSettings registers all 6 AI agent settings descriptors and a hot-reload listener on s.
func (c *ConfigStore) RegisterSettings(s *setting.Store) {
	for _, a := range agentDefs {
		s.Register(setting.Descriptor{
			Key:          a.providerIDKey,
			Group:        a.group,
			Label:        "Provider Config",
			Description:  "UUID konfigurasi penyedia LLM untuk agen " + a.displayName + ".",
			DefaultValue: "",
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
		Key:          SettingEmbeddingProviderID,
		Group:        "AI Embedding",
		Label:        "Embedding Provider Config",
		Description:  "UUID konfigurasi penyedia LLM untuk embedding (atau kosong / none).",
		DefaultValue: "",
	})
	s.Register(setting.Descriptor{
		Key:          SettingEmbeddingModel,
		Group:        "AI Embedding",
		Label:        "Embedding Model",
		Description:  "Nama model embedding 768-dimensi.",
		DefaultValue: "text-embedding-004",
	})

	s.OnChange(func(ctx context.Context, values map[string]string) error {
		return c.applySettings(ctx, values)
	})
}

func (c *ConfigStore) applySettings(ctx context.Context, values map[string]string) error {
	c.mu.RLock()
	loaded := c.providersLoaded
	hasDB := c.db != nil
	c.mu.RUnlock()
	if !loaded && hasDB {
		if err := c.ReloadProviders(ctx); err != nil {
			return err
		}
	}

	c.mu.RLock()
	providers := c.providers
	c.mu.RUnlock()

	nextAgents := make(map[string]domain.AgentConfig, len(agentDefs))
	nextAgentIDs := make(map[string]uuid.UUID, len(agentDefs))
	for _, a := range agentDefs {
		rawID := strings.TrimSpace(values[a.providerIDKey])
		var boundID uuid.UUID
		var provCfg ProviderRuntimeConfig
		if rawID != "" {
			parsed, err := uuid.Parse(rawID)
			if err != nil {
				return fmt.Errorf("%s tidak valid: harus berupa UUID", a.providerIDKey)
			}
			p, ok := providers[parsed]
			if !ok {
				return fmt.Errorf("%s tidak ditemukan: %q", a.providerIDKey, rawID)
			}
			boundID = parsed
			provCfg = p
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

		if boundID != uuid.Nil {
			nextAgentIDs[a.name] = boundID
		}
		nextAgents[a.name] = domain.AgentConfig{
			Name:        a.name,
			DisplayName: a.displayName,
			Description: a.description,
			Instruction: instr,
			Model: domain.ModelConfig{
				Provider:        provCfg.Provider,
				BaseURL:         provCfg.BaseURL,
				APIKey:          provCfg.APIKey,
				Model:           model,
				ReasoningEffort: effort,
				Temperature:     domain.DefaultTemperature,
			},
			Tools:    slices.Clone(a.tools),
			MaxSteps: domain.MaxSteps,
		}
	}

	rawEmbID := strings.TrimSpace(values[SettingEmbeddingProviderID])
	embModel := strings.TrimSpace(values[SettingEmbeddingModel])
	if embModel == "" {
		return fmt.Errorf("%s wajib diisi", SettingEmbeddingModel)
	}

	var nextEmbID uuid.UUID
	nextEmbedding := EmbeddingRuntimeConfig{
		Provider: string(domain.ProviderNone),
		Model:    embModel,
	}
	if rawEmbID != "" && rawEmbID != string(domain.ProviderNone) {
		parsed, err := uuid.Parse(rawEmbID)
		if err != nil {
			return fmt.Errorf("%s tidak valid: harus berupa UUID", SettingEmbeddingProviderID)
		}
		p, ok := providers[parsed]
		if !ok {
			return fmt.Errorf("%s tidak ditemukan: %q", SettingEmbeddingProviderID, rawEmbID)
		}
		if !isValidEmbeddingProvider(p.Provider) {
			return fmt.Errorf("%s tidak mendukung embedding: %q", SettingEmbeddingProviderID, p.Provider)
		}
		nextEmbID = parsed
		nextEmbedding = EmbeddingRuntimeConfig{
			Provider: p.Provider,
			BaseURL:  p.BaseURL,
			APIKey:   p.APIKey,
			Model:    embModel,
		}
	}

	c.mu.Lock()
	c.agents = nextAgents
	c.agentProviderIDs = nextAgentIDs
	c.embedding = nextEmbedding
	c.embeddingProviderID = nextEmbID
	c.mu.Unlock()
	return nil
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("URL harus diawali http:// atau https:// dan memiliki host yang valid")
	}
	return nil
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

func encryptAPIKey(cipher *crypt.Cipher, plain string) (string, error) {
	if cipher == nil {
		return plain, nil
	}
	return cipher.EncryptString(plain)
}

func decryptAPIKey(cipher *crypt.Cipher, stored string) (string, error) {
	if cipher == nil || !strings.HasPrefix(stored, crypt.StringPrefixV1) {
		return stored, nil
	}
	return cipher.DecryptString(stored)
}

func maskAPIKey(plain string) string {
	trimmed := strings.TrimSpace(plain)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 8 {
		return "****"
	}
	return trimmed[:4] + "..." + trimmed[len(trimmed)-4:]
}
