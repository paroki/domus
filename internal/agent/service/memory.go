package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

const maxMemoryRunes = 2000

// SearchMemories lists (q == "") or searches (q != "") caller's memories via RRF hybrid or BM25-only search.
func (s *Service) SearchMemories(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchMemoriesRequest,
) ([]domain.MemoryResponse, *string, bool, error) {
	limit := clampLimit(req.Limit)
	var cat string
	if req.Filter != nil {
		cat = strings.TrimSpace(req.Filter.Category)
	}

	trimmedQ := strings.TrimSpace(req.Q)
	if trimmedQ == "" {
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
		items, hasMore, err := s.repo.SearchMemories(ctx, p.ParokiID, p.UserID, "", cat, nil, cursorAt, cursorID, limit)
		if err != nil {
			return nil, nil, false, err
		}
		var nextCursor *string
		if hasMore && len(items) > 0 {
			last := items[len(items)-1]
			c := httpx.EncodeCursor(last.CreatedAt, last.ID.String())
			nextCursor = &c
		}
		return items, nextCursor, hasMore, nil
	}

	vec := s.embedTextBestEffort(ctx, trimmedQ)
	items, _, err := s.repo.SearchMemories(ctx, p.ParokiID, p.UserID, trimmedQ, cat, vec, nil, nil, limit)
	if err != nil {
		return nil, nil, false, err
	}
	return items, nil, false, nil
}

// CreateMemory creates a new AgentMemory and best-effort computes its 768-dim embedding.
func (s *Service) CreateMemory(
	ctx context.Context,
	p authx.Principal,
	req domain.CreateMemoryRequest,
) (domain.MemoryResponse, error) {
	content := strings.TrimSpace(req.Content)
	if err := validateMemoryContent(content); err != nil {
		return domain.MemoryResponse{}, err
	}
	cat := strings.TrimSpace(req.Category)
	if cat == "" {
		cat = domain.MemoryCategoryFact
	}
	if !isValidMemoryCategory(cat) {
		return domain.MemoryResponse{}, domain.ValidationError("category", "INVALID_CATEGORY", "Invalid memory category.")
	}

	vec := s.embedTextBestEffort(ctx, content)
	return s.repo.CreateMemory(ctx, domain.CreateMemoryInput{
		ParokiID: p.ParokiID,
		UserID:   p.UserID,
		Category: cat,
		Content:  content,
	}, vec)
}

// UpdateMemory updates an AgentMemory's content and/or category and re-computes its embedding when content changes.
func (s *Service) UpdateMemory(
	ctx context.Context,
	p authx.Principal,
	memoryID uuid.UUID,
	req domain.UpdateMemoryRequest,
) (domain.MemoryResponse, error) {
	var contentPtr *string
	var vec []float32
	clearEmbedding := false

	if req.Content != nil {
		trimmed := strings.TrimSpace(*req.Content)
		if err := validateMemoryContent(trimmed); err != nil {
			return domain.MemoryResponse{}, err
		}
		contentPtr = &trimmed
		vec = s.embedTextBestEffort(ctx, trimmed)
		clearEmbedding = len(vec) == 0
	}

	var catPtr *string
	if req.Category != nil {
		trimmed := strings.TrimSpace(*req.Category)
		if !isValidMemoryCategory(trimmed) {
			return domain.MemoryResponse{}, domain.ValidationError("category", "INVALID_CATEGORY", "Invalid memory category.")
		}
		catPtr = &trimmed
	}

	return s.repo.UpdateMemory(ctx, p.ParokiID, p.UserID, memoryID, catPtr, contentPtr, vec, clearEmbedding)
}

// DeleteMemory hard-deletes an AgentMemory owned by (p.ParokiID, p.UserID).
func (s *Service) DeleteMemory(ctx context.Context, p authx.Principal, memoryID uuid.UUID) error {
	return s.repo.DeleteMemory(ctx, p.ParokiID, p.UserID, memoryID)
}

func validateMemoryContent(content string) error {
	n := utf8.RuneCountInString(content)
	if n == 0 {
		return domain.ValidationError("content", "REQUIRED", "Memory content is required.")
	}
	if n > maxMemoryRunes {
		return domain.ValidationError("content", "MAX_LENGTH", "Memory content must be at most 2000 characters.")
	}
	return nil
}

func isValidMemoryCategory(cat string) bool {
	return cat == domain.MemoryCategoryPreference || cat == domain.MemoryCategoryFact || cat == domain.MemoryCategoryContext
}
