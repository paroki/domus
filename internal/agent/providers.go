package agent

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/internal/agent/domain"
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

// CreateProvider stores a new LLM provider config, auto-syncs models best-effort, and updates ConfigStore.
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
	var domainModels []domain.ProviderModelItem
	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	discovered, syncErr := f.drivers.DiscoverModels(syncCtx, req.Provider, baseURL, plainKey)
	cancel()
	if syncErr == nil {
		schemaModels, domainModels = toModelsPair(discovered)
	} else {
		f.config.logger.WarnContext(ctx, "auto-sync provider models on create failed", "provider", req.Provider, "err", syncErr)
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

	f.config.UpsertProvider(ProviderRuntimeConfig{
		ID:       row.ID,
		Name:     row.Name,
		Provider: string(row.Provider),
		BaseURL:  row.BaseURL,
		APIKey:   plainKey,
		Models:   domainModels,
	})
	return toProviderResponse(row, plainKey), nil
}

// UpdateProvider updates an existing LLM provider config and refreshes ConfigStore.
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
		if *req.Provider == string(domain.ProviderAnthropic) && f.config.IsProviderUsedByEmbedding(id) {
			return domain.AgentProviderResponse{}, domain.ErrProviderEmbeddingUnsupported
		}
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
			schemaModels, _ := toModelsPair(discovered)
			upd.SetModels(schemaModels)
		} else {
			f.config.logger.WarnContext(ctx, "auto-sync provider models on update failed", "provider", nextProvider, "err", syncErr)
		}
	}

	saved, err := upd.Save(ctx)
	if err != nil {
		return domain.AgentProviderResponse{}, err
	}

	domainModels := make([]domain.ProviderModelItem, len(saved.Models))
	for i, m := range saved.Models {
		domainModels[i] = domain.ProviderModelItem{
			ID:           m.ID,
			Name:         m.Name,
			Capabilities: slices.Clone(m.Capabilities),
		}
	}
	f.config.UpsertProvider(ProviderRuntimeConfig{
		ID:       saved.ID,
		Name:     saved.Name,
		Provider: string(saved.Provider),
		BaseURL:  saved.BaseURL,
		APIKey:   nextPlainKey,
		Models:   domainModels,
	})
	return toProviderResponse(saved, nextPlainKey), nil
}

// DeleteProvider removes an unused LLM provider config; returns ErrProviderInUse (409) if bound.
func (f *Feature) DeleteProvider(ctx context.Context, id uuid.UUID) error {
	if f.db == nil {
		return domain.ErrProviderUnavailable
	}
	if f.config.IsProviderInUse(id) {
		return domain.ErrProviderInUse
	}
	err := f.db.AgentProvider.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return domain.ErrProviderNotFound
	}
	if err != nil {
		return err
	}
	f.config.DeleteProvider(id)
	return nil
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
		return domain.SyncProviderModelsResponse{}, domain.ErrProviderUnavailable
	}

	schemaModels, domainModels := toModelsPair(discovered)
	saved, err := f.db.AgentProvider.UpdateOneID(id).
		SetModels(schemaModels).
		Save(ctx)
	if err != nil {
		return domain.SyncProviderModelsResponse{}, err
	}

	f.config.UpsertProvider(ProviderRuntimeConfig{
		ID:       saved.ID,
		Name:     saved.Name,
		Provider: string(saved.Provider),
		BaseURL:  saved.BaseURL,
		APIKey:   plainKey,
		Models:   domainModels,
	})

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

func toModelsPair(items []domain.ProviderModelItem) ([]schema.AgentProviderModel, []domain.ProviderModelItem) {
	schemaModels := make([]schema.AgentProviderModel, len(items))
	domainModels := make([]domain.ProviderModelItem, len(items))
	for i, m := range items {
		caps := slices.Clone(m.Capabilities)
		schemaModels[i] = schema.AgentProviderModel{
			ID:           m.ID,
			Name:         m.Name,
			Capabilities: caps,
		}
		domainModels[i] = domain.ProviderModelItem{
			ID:           m.ID,
			Name:         m.Name,
			Capabilities: slices.Clone(caps),
		}
	}
	return schemaModels, domainModels
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
