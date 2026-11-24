package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/setting"
)

const (
	settingModelGabriel   = "agent.model.gabriel"
	settingModelEmbedding = "agent.model.embedding"
)

// Service implements business logic for agent sessions, messages, and long-term memories.
type Service struct {
	repo     *repo.Repository
	db       *ent.Client
	drivers  *driver.Registry
	cipher   *crypt.Cipher
	settings *setting.Store
	logger   *slog.Logger
}

// New constructs an agent Service.
func New(
	r *repo.Repository,
	db *ent.Client,
	drivers *driver.Registry,
	cipher *crypt.Cipher,
	settings *setting.Store,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Service{
		repo:     r,
		db:       db,
		drivers:  drivers,
		cipher:   cipher,
		settings: settings,
		logger:   logger,
	}
}

func (s *Service) embedTextBestEffort(ctx context.Context, text string) []float32 {
	if s.settings == nil || s.db == nil || s.drivers == nil {
		return nil
	}
	raw := strings.TrimSpace(s.settings.Get(settingModelEmbedding))
	if raw == "" {
		return nil
	}
	provPart, modelID, ok := strings.Cut(raw, ":")
	if !ok || strings.TrimSpace(modelID) == "" {
		return nil
	}
	provID, err := uuid.Parse(strings.TrimSpace(provPart))
	if err != nil {
		return nil
	}
	row, err := s.db.AgentProvider.Query().Where(agentprovider.IDEQ(provID)).Only(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "embedding provider lookup failed; falling back to BM25-only", "providerId", provID, "err", err)
		return nil
	}
	plainKey := row.APIKey
	if s.cipher != nil && strings.HasPrefix(plainKey, crypt.StringPrefixV1) {
		decrypted, err := s.cipher.DecryptString(plainKey)
		if err != nil {
			s.logger.WarnContext(ctx, "embedding API key decrypt failed; falling back to BM25-only", "providerId", provID, "err", err)
			return nil
		}
		plainKey = decrypted
	}
	return s.drivers.EmbedTextBestEffort(ctx, string(row.Provider), row.BaseURL, plainKey, strings.TrimSpace(modelID), text, s.logger)
}

func (s *Service) resolveDefaultGabrielModel(ctx context.Context) (*uuid.UUID, string) {
	if s.settings == nil || s.db == nil {
		return nil, ""
	}
	raw := strings.TrimSpace(s.settings.Get(settingModelGabriel))
	if raw == "" {
		return nil, ""
	}
	provPart, modelID, ok := strings.Cut(raw, ":")
	if !ok || strings.TrimSpace(modelID) == "" {
		return nil, ""
	}
	provID, err := uuid.Parse(strings.TrimSpace(provPart))
	if err != nil {
		return nil, ""
	}
	exists, err := s.db.AgentProvider.Query().Where(agentprovider.IDEQ(provID)).Exist(ctx)
	if err != nil || !exists {
		return nil, ""
	}
	return &provID, strings.TrimSpace(modelID)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		return httpx.MaxLimit
	}
	return limit
}
