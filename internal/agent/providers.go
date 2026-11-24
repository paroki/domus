package agent

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/httpx"
)

// ListProviders returns all configured LLM provider entries ordered by creation time.
func (f *Feature) ListProviders(ctx context.Context) ([]domain.AgentProviderResponse, error) {
	if f.db == nil {
		return nil, domain.ErrProviderUnavailable
	}
	rows, err := f.db.AgentProvider.Query().
		Order(agentprovider.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AgentProviderResponse, len(rows))
	for i, row := range rows {
		plainKey, err := decryptAPIKey(f.cipher, row.APIKey)
		if err != nil {
			return nil, err
		}
		out[i] = toProviderResponse(row, plainKey)
	}
	return out, nil
}

// CreateProvider stores a new LLM provider config and auto-syncs models best-effort.
func (f *Feature) CreateProvider(ctx context.Context, req domain.CreateAgentProviderRequest) (domain.AgentProviderResponse, error) {
	if f.db == nil {
		return domain.AgentProviderResponse{}, domain.ErrProviderUnavailable
	}
	name := strings.TrimSpace(req.Name)
	desc := strings.TrimSpace(req.Description)
	baseURL := strings.TrimSpace(req.BaseURL)
	if baseURL != "" {
		if err := validateHTTPURL(baseURL); err != nil {
			return domain.AgentProviderResponse{}, httpx.BadRequest(err.Error())
		}
	}
	plainKey := strings.TrimSpace(req.APIKey)
	if plainKey == "" {
		return domain.AgentProviderResponse{}, domain.ValidationError("apiKey", "REQUIRED", "API key is required.")
	}
	encKey, err := encryptAPIKey(f.cipher, plainKey)
	if err != nil {
		return domain.AgentProviderResponse{}, err
	}

	var schemaModels []schema.AgentProviderModel
	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	discovered, syncErr := f.drivers.DiscoverModels(syncCtx, req.Provider, baseURL, plainKey)
	cancel()
	if syncErr == nil {
		schemaModels = toSchemaModels(discovered)
	} else {
		f.logger.WarnContext(ctx, "auto-sync provider models on create failed", "provider", req.Provider, "err", syncErr)
	}

	row, err := f.db.AgentProvider.Create().
		SetName(name).
		SetDescription(desc).
		SetProvider(agentprovider.Provider(req.Provider)).
		SetAPIKey(encKey).
		SetBaseURL(baseURL).
		SetModels(schemaModels).
		Save(ctx)
	if err != nil {
		return domain.AgentProviderResponse{}, err
	}
	return toProviderResponse(row, plainKey), nil
}

// UpdateProvider updates an existing LLM provider config and re-syncs models if credentials changed.
func (f *Feature) UpdateProvider(ctx context.Context, id uuid.UUID, req domain.UpdateAgentProviderRequest) (domain.AgentProviderResponse, error) {
	if f.db == nil {
		return domain.AgentProviderResponse{}, domain.ErrProviderUnavailable
	}
	row, err := f.db.AgentProvider.Get(ctx, id)
	if ent.IsNotFound(err) {
		return domain.AgentProviderResponse{}, domain.ErrProviderNotFound
	}
	if err != nil {
		return domain.AgentProviderResponse{}, err
	}

	oldPlainKey, err := decryptAPIKey(f.cipher, row.APIKey)
	if err != nil {
		return domain.AgentProviderResponse{}, err
	}

	upd := f.db.AgentProvider.UpdateOneID(id)
	nextProvider := string(row.Provider)
	nextBaseURL := row.BaseURL
	nextPlainKey := oldPlainKey
	credentialsChanged := false

	if req.Name != nil {
		upd.SetName(strings.TrimSpace(*req.Name))
	}
	if req.Description != nil {
		upd.SetDescription(strings.TrimSpace(*req.Description))
	}
	if req.Provider != nil && *req.Provider != nextProvider {
		nextProvider = *req.Provider
		upd.SetProvider(agentprovider.Provider(nextProvider))
		credentialsChanged = true
	}
	if req.BaseURL != nil {
		b := strings.TrimSpace(*req.BaseURL)
		if b != "" {
			if err := validateHTTPURL(b); err != nil {
				return domain.AgentProviderResponse{}, httpx.BadRequest(err.Error())
			}
		}
		if b != nextBaseURL {
			nextBaseURL = b
			upd.SetBaseURL(b)
			credentialsChanged = true
		}
	}
	if req.APIKey != nil && strings.TrimSpace(*req.APIKey) != "" {
		k := strings.TrimSpace(*req.APIKey)
		if k != oldPlainKey {
			enc, err := encryptAPIKey(f.cipher, k)
			if err != nil {
				return domain.AgentProviderResponse{}, err
			}
			nextPlainKey = k
			upd.SetAPIKey(enc)
			credentialsChanged = true
		}
	}

	if credentialsChanged {
		syncCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		discovered, syncErr := f.drivers.DiscoverModels(syncCtx, nextProvider, nextBaseURL, nextPlainKey)
		cancel()
		if syncErr == nil {
			upd.SetModels(toSchemaModels(discovered))
		} else {
			f.logger.WarnContext(ctx, "auto-sync provider models on update failed", "provider", nextProvider, "err", syncErr)
		}
	}

	saved, err := upd.Save(ctx)
	if err != nil {
		return domain.AgentProviderResponse{}, err
	}
	return toProviderResponse(saved, nextPlainKey), nil
}

// DeleteProvider removes an LLM provider config if no agent.model.* setting references it.
func (f *Feature) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	if f.db == nil {
		return domain.ErrProviderUnavailable
	}
	exists, err := f.db.AgentProvider.Query().Where(agentprovider.IDEQ(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrProviderNotFound
	}
	if f.settings != nil {
		prefix := id.String() + ":"
		for _, key := range modelSettingKeys {
			if strings.HasPrefix(f.settings.Get(key), prefix) {
				return domain.ErrProviderInUse
			}
		}
	}
	err = f.db.AgentProvider.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return domain.ErrProviderNotFound
	}
	return err
}

// SyncProviderModels discovers available models from the upstream provider and persists the catalog.
func (f *Feature) SyncProviderModels(ctx context.Context, id uuid.UUID) (domain.SyncProviderModelsResponse, error) {
	if f.db == nil {
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderUnavailable
	}
	row, err := f.db.AgentProvider.Get(ctx, id)
	if ent.IsNotFound(err) {
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderNotFound
	}
	if err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}

	plainKey, err := decryptAPIKey(f.cipher, row.APIKey)
	if err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}
	if strings.TrimSpace(plainKey) == "" {
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderUnavailable
	}

	discovered, err := f.drivers.DiscoverModels(ctx, string(row.Provider), row.BaseURL, plainKey)
	if err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}

	saved, err := f.db.AgentProvider.UpdateOneID(id).
		SetModels(toSchemaModels(discovered)).
		Save(ctx)
	if err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}

	dtos := make([]domain.ProviderModelDTO, len(discovered))
	for i, m := range discovered {
		dtos[i] = domain.ProviderModelDTO(m)
	}
	return domain.SyncProviderModelsResponse{
		ID:       saved.ID,
		Provider: string(saved.Provider),
		Models:   dtos,
	}, nil
}

func toSchemaModels(items []domain.ProviderModelItem) []schema.AgentProviderModel {
	out := make([]schema.AgentProviderModel, len(items))
	for i, m := range items {
		out[i] = schema.AgentProviderModel{
			ID:           m.ID,
			Name:         m.Name,
			Capabilities: slices.Clone(m.Capabilities),
		}
	}
	return out
}

func toProviderResponse(row *ent.AgentProvider, plainKey string) domain.AgentProviderResponse {
	models := make([]domain.ProviderModelDTO, len(row.Models))
	for i, m := range row.Models {
		models[i] = domain.ProviderModelDTO{
			ID:           m.ID,
			Name:         m.Name,
			Capabilities: slices.Clone(m.Capabilities),
		}
	}
	return domain.AgentProviderResponse{
		ID:           row.ID,
		Name:         row.Name,
		Description:  row.Description,
		Provider:     string(row.Provider),
		BaseURL:      row.BaseURL,
		HasAPIKey:    strings.TrimSpace(plainKey) != "",
		APIKeyMasked: maskAPIKey(plainKey),
		Models:       models,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("URL harus diawali http:// atau https:// dan memiliki host yang valid")
	}
	return nil
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
