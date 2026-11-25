package repo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentmemory"
	"github.com/paroki/domus/internal/agent/domain"
)

// CreateMemory inserts a new AgentMemory and optionally stores its 768-d halfvec embedding.
func (r *Repository) CreateMemory(ctx context.Context, in domain.CreateMemoryInput, embedding []float32) (domain.MemoryResponse, error) {
	cat := agentmemory.CategoryFact
	if strings.TrimSpace(in.Category) != "" {
		cat = agentmemory.Category(strings.TrimSpace(in.Category))
	}
	b := r.client.AgentMemory.Create().
		SetParokiID(in.ParokiID).
		SetUserID(in.UserID).
		SetCategory(cat).
		SetContent(strings.TrimSpace(in.Content))
	if in.SourceSessionID != nil {
		b.SetSourceSessionID(*in.SourceSessionID)
	}
	row, err := b.Save(ctx)
	if err != nil {
		return domain.MemoryResponse{}, err
	}
	if len(embedding) == EmbeddingDimensions && r.sqlDB != nil {
		lit, err := formatHalfvecLiteral(embedding)
		if err != nil {
			return domain.MemoryResponse{}, err
		}
		if _, err := r.sqlDB.ExecContext(
			ctx,
			`UPDATE "agent_memories" SET "embedding" = $1::halfvec(768) WHERE "id" = $2 AND "paroki_id" = $3 AND "user_id" = $4`,
			lit, row.ID, in.ParokiID, in.UserID,
		); err != nil {
			return domain.MemoryResponse{}, err
		}
	}
	return toMemoryResponse(row), nil
}

// UpdateMemory updates an existing AgentMemory scoped to (parokiID, userID) and updates or clears its embedding.
func (r *Repository) UpdateMemory(
	ctx context.Context,
	parokiID, userID, memoryID uuid.UUID,
	category, content *string,
	embedding []float32,
	clearEmbedding bool,
) (domain.MemoryResponse, error) {
	existing, err := r.client.AgentMemory.Query().
		Where(agentmemory.IDEQ(memoryID), agentmemory.ParokiIDEQ(parokiID), agentmemory.UserIDEQ(userID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return domain.MemoryResponse{}, domain.ErrMemoryNotFound
	}
	if err != nil {
		return domain.MemoryResponse{}, err
	}

	upd := r.client.AgentMemory.UpdateOneID(existing.ID)
	if category != nil && strings.TrimSpace(*category) != "" {
		upd.SetCategory(agentmemory.Category(strings.TrimSpace(*category)))
	}
	if content != nil && strings.TrimSpace(*content) != "" {
		upd.SetContent(strings.TrimSpace(*content))
	}
	row, err := upd.Save(ctx)
	if err != nil {
		return domain.MemoryResponse{}, err
	}

	if r.sqlDB != nil {
		switch {
		case len(embedding) == EmbeddingDimensions:
			lit, err := formatHalfvecLiteral(embedding)
			if err != nil {
				return domain.MemoryResponse{}, err
			}
			if _, err := r.sqlDB.ExecContext(
				ctx,
				`UPDATE "agent_memories" SET "embedding" = $1::halfvec(768) WHERE "id" = $2 AND "paroki_id" = $3 AND "user_id" = $4`,
				lit, row.ID, parokiID, userID,
			); err != nil {
				return domain.MemoryResponse{}, err
			}
		case clearEmbedding:
			if _, err := r.sqlDB.ExecContext(
				ctx,
				`UPDATE "agent_memories" SET "embedding" = NULL WHERE "id" = $1 AND "paroki_id" = $2 AND "user_id" = $3`,
				row.ID, parokiID, userID,
			); err != nil {
				return domain.MemoryResponse{}, err
			}
		}
	}
	return toMemoryResponse(row), nil
}

// DeleteMemory hard-deletes an AgentMemory scoped to (parokiID, userID).
func (r *Repository) DeleteMemory(ctx context.Context, parokiID, userID, memoryID uuid.UUID) error {
	affected, err := r.client.AgentMemory.Delete().
		Where(agentmemory.IDEQ(memoryID), agentmemory.ParokiIDEQ(parokiID), agentmemory.UserIDEQ(userID)).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrMemoryNotFound
	}
	return nil
}

// ClearAllEmbeddings resets all stored halfvec embeddings to NULL.
func (r *Repository) ClearAllEmbeddings(ctx context.Context) error {
	if r.sqlDB == nil {
		return nil
	}
	_, err := r.sqlDB.ExecContext(ctx, `UPDATE "agent_memories" SET "embedding" = NULL`)
	return err
}

// SearchMemories lists or searches AgentMemory rows for (parokiID, userID).
// When q is empty, it uses deterministic keyset pagination ordered by (created_at DESC, id DESC).
// When q is non-empty, it executes ParadeDB BM25-only search, or RRF (k=60) hybrid search when len(vec) == 768.
func (r *Repository) SearchMemories(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	q, category string,
	vec []float32,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.MemoryResponse, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	trimmedQ := strings.TrimSpace(q)
	trimmedCat := strings.TrimSpace(category)
	if trimmedQ == "" {
		return r.listMemoriesKeyset(ctx, parokiID, userID, trimmedCat, cursorCreatedAt, cursorID, limit)
	}
	if len(vec) == EmbeddingDimensions {
		items, err := r.searchMemoriesRRF(ctx, parokiID, userID, trimmedQ, trimmedCat, vec, limit)
		return items, false, err
	}
	items, err := r.searchMemoriesBM25(ctx, parokiID, userID, trimmedQ, trimmedCat, limit)
	return items, false, err
}

func (r *Repository) listMemoriesKeyset(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	category string,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.MemoryResponse, bool, error) {
	q := r.client.AgentMemory.Query().
		Where(agentmemory.ParokiIDEQ(parokiID), agentmemory.UserIDEQ(userID))
	if category != "" {
		q = q.Where(agentmemory.CategoryEQ(agentmemory.Category(category)))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		q = q.Where(agentmemory.Or(
			agentmemory.CreatedAtLT(*cursorCreatedAt),
			agentmemory.And(
				agentmemory.CreatedAtEQ(*cursorCreatedAt),
				agentmemory.IDLT(*cursorID),
			),
		))
	}
	rows, err := q.
		Order(ent.Desc(agentmemory.FieldCreatedAt), ent.Desc(agentmemory.FieldID)).
		Limit(limit + 1).
		All(ctx)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	out := make([]domain.MemoryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toMemoryResponse(row))
	}
	return out, hasMore, nil
}

func (r *Repository) searchMemoriesBM25(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	q, category string,
	limit int,
) ([]domain.MemoryResponse, error) {
	const query = `
		SELECT id, paroki_id, user_id, category, content, source_session_id,
		       paradedb.score(id)::double precision AS score, created_at, updated_at
		FROM agent_memories
		WHERE paroki_id = $1
		  AND user_id = $2
		  AND ($3 = '' OR category = $3)
		  AND id @@@ paradedb.boolean(should => ARRAY[
		    paradedb.match('content', $4, conjunction_mode => false),
		    paradedb.match('category', $4, conjunction_mode => false)
		  ])
		ORDER BY score DESC, created_at DESC, id DESC
		LIMIT $5
	`
	rows, err := r.sqlDB.QueryContext(ctx, query, parokiID, userID, category, q, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanMemoryRows(rows)
}

func (r *Repository) searchMemoriesRRF(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	q, category string,
	vec []float32,
	limit int,
) ([]domain.MemoryResponse, error) {
	lit, err := formatHalfvecLiteral(vec)
	if err != nil {
		return nil, err
	}
	candidateLimit := max(limit*3, 50)
	const query = `
		WITH bm25 AS (
			SELECT id,
			       ROW_NUMBER() OVER (ORDER BY paradedb.score(id) DESC, created_at DESC, id DESC) AS r
			FROM agent_memories
			WHERE paroki_id = $1
			  AND user_id = $2
			  AND ($3 = '' OR category = $3)
			  AND id @@@ paradedb.boolean(should => ARRAY[
			    paradedb.match('content', $4, conjunction_mode => false),
			    paradedb.match('category', $4, conjunction_mode => false)
			  ])
			ORDER BY paradedb.score(id) DESC, created_at DESC, id DESC
			LIMIT $5
		),
		vec AS (
			SELECT id,
			       ROW_NUMBER() OVER (ORDER BY embedding <=> $6::halfvec(768) ASC, created_at DESC, id DESC) AS r
			FROM agent_memories
			WHERE paroki_id = $1
			  AND user_id = $2
			  AND ($3 = '' OR category = $3)
			  AND embedding IS NOT NULL
			ORDER BY embedding <=> $6::halfvec(768) ASC, created_at DESC, id DESC
			LIMIT $5
		),
		rrf AS (
			SELECT COALESCE(bm25.id, vec.id) AS id,
			       COALESCE(1.0 / (60.0 + bm25.r), 0.0) + COALESCE(1.0 / (60.0 + vec.r), 0.0) AS score
			FROM bm25
			FULL OUTER JOIN vec ON bm25.id = vec.id
		)
		SELECT m.id, m.paroki_id, m.user_id, m.category, m.content, m.source_session_id,
		       rrf.score::double precision AS score, m.created_at, m.updated_at
		FROM rrf
		JOIN agent_memories m ON m.id = rrf.id
		ORDER BY rrf.score DESC, m.created_at DESC, m.id DESC
		LIMIT $7
	`
	rows, err := r.sqlDB.QueryContext(ctx, query, parokiID, userID, category, q, candidateLimit, lit, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanMemoryRows(rows)
}

func scanMemoryRows(rows *sql.Rows) ([]domain.MemoryResponse, error) {
	var out []domain.MemoryResponse
	for rows.Next() {
		var m domain.MemoryResponse
		var sourceSessionID sql.NullString
		if err := rows.Scan(
			&m.ID,
			&m.ParokiID,
			&m.UserID,
			&m.Category,
			&m.Content,
			&sourceSessionID,
			&m.Score,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if sourceSessionID.Valid {
			id, err := uuid.Parse(sourceSessionID.String)
			if err != nil {
				return nil, err
			}
			m.SourceSessionID = &id
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func toMemoryResponse(row *ent.AgentMemory) domain.MemoryResponse {
	return domain.MemoryResponse{
		ID:              row.ID,
		ParokiID:        row.ParokiID,
		UserID:          row.UserID,
		Category:        string(row.Category),
		Content:         row.Content,
		SourceSessionID: row.SourceSessionID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
