package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchSessions returns keyset-paginated active sessions for the caller in their active paroki.
func (s *Service) SearchSessions(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchSessionsRequest,
) ([]domain.SessionResponse, *string, bool, error) {
	limit := clampLimit(req.Limit)
	var cursorAt *time.Time
	var cursorID *uuid.UUID
	if strings.TrimSpace(req.Cursor) != "" {
		cur, err := httpx.DecodeCursor(req.Cursor)
		if err != nil {
			return nil, nil, false, err
		}
		id, err := uuid.Parse(cur.ID)
		if err != nil {
			return nil, nil, false, httpx.InvalidCursor("Invalid pagination cursor.")
		}
		cursorAt = &cur.CreatedAt
		cursorID = &id
	}

	items, hasMore, err := s.repo.SearchSessions(ctx, p.ParokiID, p.UserID, req.Q, cursorAt, cursorID, limit)
	if err != nil {
		return nil, nil, false, err
	}
	var nextCursor *string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		c := httpx.EncodeCursor(last.LastMessageAt, last.ID.String())
		nextCursor = &c
	}
	return items, nextCursor, hasMore, nil
}

// GetSession returns session metadata, chronological messages, and the effective model (falling back cleanly to Gabriel if provider was deleted).
func (s *Service) GetSession(ctx context.Context, p authx.Principal, sessionID uuid.UUID) (domain.SessionDetailResponse, error) {
	sess, err := s.repo.GetSession(ctx, p.ParokiID, p.UserID, sessionID)
	if err != nil {
		return domain.SessionDetailResponse{}, err
	}
	msgs, err := s.repo.ListMessages(ctx, p.ParokiID, p.UserID, sessionID, 0)
	if err != nil {
		return domain.SessionDetailResponse{}, err
	}

	var effProvID *uuid.UUID
	var effModelID string
	if sess.ProviderID != nil && sess.ModelID != nil && strings.TrimSpace(*sess.ModelID) != "" && s.db != nil {
		exists, err := s.db.AgentProvider.Query().Where(agentprovider.IDEQ(*sess.ProviderID)).Exist(ctx)
		if err == nil && exists {
			effProvID = sess.ProviderID
			effModelID = strings.TrimSpace(*sess.ModelID)
		}
	}
	if effProvID == nil {
		effProvID, effModelID = s.resolveDefaultGabrielModel(ctx)
	}

	return domain.SessionDetailResponse{
		ID:                  sess.ID,
		ParokiID:            sess.ParokiID,
		UserID:              sess.UserID,
		Title:               sess.Title,
		ProviderID:          sess.ProviderID,
		ModelID:             sess.ModelID,
		EffectiveProviderID: effProvID,
		EffectiveModelID:    effModelID,
		LastMessageAt:       sess.LastMessageAt,
		CreatedAt:           sess.CreatedAt,
		UpdatedAt:           sess.UpdatedAt,
		Messages:            msgs,
	}, nil
}

// UpdateSession updates a session's title and/or model override.
// Sending "" on providerId or modelId clears both provider_id and model_id back to NULL.
func (s *Service) UpdateSession(
	ctx context.Context,
	p authx.Principal,
	sessionID uuid.UUID,
	req domain.UpdateSessionRequest,
) (domain.SessionResponse, error) {
	existing, err := s.repo.GetSession(ctx, p.ParokiID, p.UserID, sessionID)
	if err != nil {
		return domain.SessionResponse{}, err
	}

	var titlePtr *string
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return domain.SessionResponse{}, domain.ValidationError("title", "REQUIRED", "Title cannot be empty.")
		}
		titlePtr = &trimmed
	}

	clearModel := false
	var nextProvID *uuid.UUID
	var nextModelID *string

	if (req.ProviderID != nil && strings.TrimSpace(*req.ProviderID) == "") ||
		(req.ModelID != nil && strings.TrimSpace(*req.ModelID) == "") {
		clearModel = true
	} else if req.ProviderID != nil || req.ModelID != nil {
		provID := existing.ProviderID
		if req.ProviderID != nil {
			parsed, err := uuid.Parse(strings.TrimSpace(*req.ProviderID))
			if err != nil {
				return domain.SessionResponse{}, domain.ValidationError("providerId", "INVALID_UUID", "Invalid providerId.")
			}
			provID = &parsed
		}
		if provID == nil {
			return domain.SessionResponse{}, domain.ValidationError("providerId", "REQUIRED", "providerId is required.")
		}
		if s.db != nil {
			exists, err := s.db.AgentProvider.Query().Where(agentprovider.IDEQ(*provID)).Exist(ctx)
			if err != nil {
				return domain.SessionResponse{}, err
			}
			if !exists {
				return domain.SessionResponse{}, domain.ErrProviderNotFound
			}
		}

		modelID := ""
		if existing.ModelID != nil {
			modelID = strings.TrimSpace(*existing.ModelID)
		}
		if req.ModelID != nil {
			modelID = strings.TrimSpace(*req.ModelID)
		}
		if modelID == "" {
			return domain.SessionResponse{}, domain.ValidationError("modelId", "REQUIRED", "modelId is required.")
		}
		nextProvID = provID
		nextModelID = &modelID
	}

	return s.repo.UpdateSession(ctx, p.ParokiID, p.UserID, sessionID, titlePtr, clearModel, nextProvID, nextModelID)
}

// DeleteSession soft-deletes an active session owned by (p.ParokiID, p.UserID).
func (s *Service) DeleteSession(ctx context.Context, p authx.Principal, sessionID uuid.UUID) error {
	return s.repo.SoftDeleteSession(ctx, p.ParokiID, p.UserID, sessionID)
}
