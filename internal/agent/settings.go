package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/setting"
)

// Setting keys for AI agent model routing and persona prompts.
var (
	SettingModelGabriel         = setting.Key("agent", "model.gabriel")
	SettingModelMikael          = setting.Key("agent", "model.mikael")
	SettingModelRafael          = setting.Key("agent", "model.rafael")
	SettingModelMemoryExtractor = setting.Key("agent", "model.memory_extractor")
	SettingModelEmbedding       = setting.Key("agent", "model.embedding")

	SettingPromptGabriel         = setting.Key("agent", "prompt.gabriel")
	SettingPromptMikael          = setting.Key("agent", "prompt.mikael")
	SettingPromptRafael          = setting.Key("agent", "prompt.rafael")
	SettingPromptMemoryExtractor = setting.Key("agent", "prompt.memory_extractor")
)

var modelSettingKeys = []string{
	SettingModelGabriel,
	SettingModelMikael,
	SettingModelRafael,
	SettingModelMemoryExtractor,
	SettingModelEmbedding,
}

const (
	groupModelRouting   = "AI Agent — Model Routing"
	groupPersonaPrompts = "AI Agent — Persona Prompts"
)

// RegisterSettings registers agent model routing and persona prompt descriptors and validation.
func (f *Feature) RegisterSettings(s *setting.Store) {
	s.Register(setting.Descriptor{
		Key:         SettingModelGabriel,
		Group:       groupModelRouting,
		Label:       "Model Gabriel",
		Description: "Model chat utama untuk asisten Gabriel (format: <providerId>:<modelId>).",
		Placeholder: "<providerId>:gemini-2.5-flash",
	})
	s.Register(setting.Descriptor{
		Key:         SettingModelMikael,
		Group:       groupModelRouting,
		Label:       "Model Mikael",
		Description: "Model chat opsional untuk agen Mikael; mengikuti Gabriel jika kosong.",
		Placeholder: "<providerId>:gemini-2.5-flash",
	})
	s.Register(setting.Descriptor{
		Key:         SettingModelRafael,
		Group:       groupModelRouting,
		Label:       "Model Rafael",
		Description: "Model chat opsional untuk agen Rafael; mengikuti Gabriel jika kosong.",
		Placeholder: "<providerId>:gemini-2.5-flash",
	})
	s.Register(setting.Descriptor{
		Key:         SettingModelMemoryExtractor,
		Group:       groupModelRouting,
		Label:       "Model Memory Extractor",
		Description: "Model chat opsional untuk ekstraksi memori; mengikuti Gabriel jika kosong.",
		Placeholder: "<providerId>:gemini-2.5-flash",
	})
	s.Register(setting.Descriptor{
		Key:         SettingModelEmbedding,
		Group:       groupModelRouting,
		Label:       "Model Embedding",
		Description: "Model embedding 768-dimensi opsional (contoh: <providerId>:gemini-embedding-001); kosong berarti mode BM25 saja.",
		Placeholder: "<providerId>:gemini-embedding-001",
	})

	s.Register(setting.Descriptor{
		Key:          SettingPromptGabriel,
		Group:        groupPersonaPrompts,
		Label:        "Prompt Gabriel",
		Description:  "Instruksi sistem default untuk asisten Gabriel.",
		Multiline:    true,
		DefaultValue: strings.TrimSpace(prompts.Gabriel),
	})
	s.Register(setting.Descriptor{
		Key:          SettingPromptMikael,
		Group:        groupPersonaPrompts,
		Label:        "Prompt Mikael",
		Description:  "Instruksi sistem default untuk agen Mikael.",
		Multiline:    true,
		DefaultValue: strings.TrimSpace(prompts.Mikael),
	})
	s.Register(setting.Descriptor{
		Key:          SettingPromptRafael,
		Group:        groupPersonaPrompts,
		Label:        "Prompt Rafael",
		Description:  "Instruksi sistem default untuk agen Rafael.",
		Multiline:    true,
		DefaultValue: strings.TrimSpace(prompts.Rafael),
	})
	s.Register(setting.Descriptor{
		Key:          SettingPromptMemoryExtractor,
		Group:        groupPersonaPrompts,
		Label:        "Prompt Memory Extractor",
		Description:  "Instruksi sistem default untuk ekstraksi memori.",
		Multiline:    true,
		DefaultValue: strings.TrimSpace(prompts.MemoryExtractor),
	})

	s.OnChange(func(ctx context.Context, values map[string]string) error {
		for _, key := range modelSettingKeys {
			val := strings.TrimSpace(values[key])
			if val == "" {
				continue
			}
			providerPart, modelPart, ok := strings.Cut(val, ":")
			if !ok || strings.TrimSpace(modelPart) == "" {
				return fmt.Errorf("format %s tidak valid, gunakan <providerUUID>:<modelId>", key)
			}
			providerID, err := uuid.Parse(strings.TrimSpace(providerPart))
			if err != nil {
				return fmt.Errorf("provider UUID pada %s tidak valid", key)
			}
			if f.db != nil {
				exists, err := f.db.AgentProvider.Query().Where(agentprovider.IDEQ(providerID)).Exist(ctx)
				if err != nil {
					return err
				}
				if !exists {
					return fmt.Errorf("konfigurasi AI provider %s untuk %s tidak ditemukan", providerID, key)
				}
			}
		}
		return nil
	})
}
