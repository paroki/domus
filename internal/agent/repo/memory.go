package repo

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentmemory"
	"github.com/paroki/domus/internal/agent/domain"
)

// ListPinnedMemories returns Tier 1 always-injected pinned memories for (parishID, userID).
func (r *Repository) ListPinnedMemories(ctx context.Context, parishID, userID uuid.UUID, limit int) ([]*ent.AgentMemory, error) {
	if limit <= 0 {
		limit = domain.MaxPinnedMemories
	}
	return r.client.AgentMemory.Query().
		Where(
			agentmemory.ParishIDEQ(parishID),
			agentmemory.ScopeEQ(agentmemory.ScopeUser),
			agentmemory.UserIDEQ(userID),
			agentmemory.IsPinnedEQ(true),
			agentmemory.DeletedAtIsNil(),
		).
		Order(agentmemory.ByUpdatedAt(sql.OrderAsc()), agentmemory.ByID(sql.OrderAsc())).
		Limit(limit).
		All(ctx)
}

// formatHalfvec formats a 768-dimension float32 slice as a pgvector literal "[0.1,0.2,...]".
func formatHalfvec(vec []float32) (string, error) {
	if len(vec) == 0 {
		return "", nil
	}
	if len(vec) != domain.EmbeddingDimensions {
		return "", fmt.Errorf("expected %d embedding dimensions, got %d", domain.EmbeddingDimensions, len(vec))
	}
	var b strings.Builder
	b.Grow(len(vec) * 8)
	b.WriteByte('[')
	for i, v := range vec {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String(), nil
}

// SaveMemory inserts a long-term memory fact, optionally superseding an existing fact and enforcing quotas.
func (r *Repository) SaveMemory(ctx context.Context, in domain.SaveMemoryInput) (*ent.AgentMemory, error) {
	vecStr, err := formatHalfvec(in.Embedding)
	if err != nil {
		return nil, err
	}

	var saved *ent.AgentMemory
	err = r.WithTx(ctx, func(txRepo *Repository) error {
		now := time.Now().UTC()
		if in.SupersedeID != nil {
			target, err := txRepo.client.AgentMemory.Query().
				Where(
					agentmemory.IDEQ(*in.SupersedeID),
					agentmemory.ParishIDEQ(in.ParishID),
					agentmemory.DeletedAtIsNil(),
				).
				Only(ctx)
			if ent.IsNotFound(err) {
				return domain.ErrMemoryNotFound
			}
			if err != nil {
				return err
			}
			if target.Scope != agentmemory.Scope(in.Scope) {
				return domain.ErrForbidden
			}
			if in.Scope == domain.MemoryScopeUser {
				if target.UserID == nil || in.UserID == nil || *target.UserID != *in.UserID {
					return domain.ErrForbidden
				}
			}
			if err := txRepo.client.AgentMemory.UpdateOneID(target.ID).
				SetDeletedAt(now).
				SetUpdatedAt(now).
				Exec(ctx); err != nil {
				return err
			}
		}

		if err := txRepo.checkMemoryQuota(ctx, in.ParishID, in.UserID, in.Scope, true, in.IsPinned); err != nil {
			return err
		}

		cre := txRepo.client.AgentMemory.Create().
			SetParishID(in.ParishID).
			SetScope(agentmemory.Scope(in.Scope)).
			SetCategory(agentmemory.Category(in.Category)).
			SetContent(strings.TrimSpace(in.Content)).
			SetIsPinned(in.IsPinned)
		if in.UserID != nil {
			cre.SetUserID(*in.UserID)
		}
		if in.SourceSessionID != nil {
			cre.SetSourceSessionID(*in.SourceSessionID)
		}
		saved, err = cre.Save(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	if r.dialect == dialect.Postgres && r.db != nil && vecStr != "" {
		if _, err := r.db.ExecContext(
			ctx,
			`UPDATE "agent_memories" SET "embedding" = $1::halfvec(768) WHERE "id" = $2`,
			vecStr, saved.ID,
		); err != nil {
			return nil, err
		}
	}
	return saved, nil
}

// UpdateMemory updates mutable fields on an existing non-deleted memory fact.
func (r *Repository) UpdateMemory(ctx context.Context, in domain.UpdateMemoryInput) (*ent.AgentMemory, error) {
	vecStr, err := formatHalfvec(in.Embedding)
	if err != nil {
		return nil, err
	}

	var updated *ent.AgentMemory
	err = r.WithTx(ctx, func(txRepo *Repository) error {
		row, err := txRepo.client.AgentMemory.Query().
			Where(
				agentmemory.IDEQ(in.ID),
				agentmemory.ParishIDEQ(in.ParishID),
				agentmemory.DeletedAtIsNil(),
			).
			Only(ctx)
		if ent.IsNotFound(err) {
			return domain.ErrMemoryNotFound
		}
		if err != nil {
			return err
		}
		if row.Scope == agentmemory.ScopeUser {
			if row.UserID == nil || *row.UserID != in.UserID {
				return domain.ErrForbidden
			}
		}

		if in.IsPinned != nil && *in.IsPinned && !row.IsPinned {
			if err := txRepo.checkMemoryQuota(ctx, in.ParishID, row.UserID, domain.MemoryScope(row.Scope), false, true); err != nil {
				return err
			}
		}

		upd := txRepo.client.AgentMemory.UpdateOneID(row.ID)
		if in.Category != nil {
			upd.SetCategory(agentmemory.Category(*in.Category))
		}
		if in.Content != nil {
			upd.SetContent(strings.TrimSpace(*in.Content))
		}
		if in.IsPinned != nil {
			upd.SetIsPinned(*in.IsPinned)
		}
		updated, err = upd.Save(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	if r.dialect == dialect.Postgres && r.db != nil {
		if in.ClearEmbedding {
			if _, err := r.db.ExecContext(ctx, `UPDATE "agent_memories" SET "embedding" = NULL WHERE "id" = $1`, updated.ID); err != nil {
				return nil, err
			}
		} else if vecStr != "" {
			if _, err := r.db.ExecContext(ctx, `UPDATE "agent_memories" SET "embedding" = $1::halfvec(768) WHERE "id" = $2`, vecStr, updated.ID); err != nil {
				return nil, err
			}
		}
	}
	return updated, nil
}

// SoftDeleteMemory soft-deletes a memory fact scoped to (parishID, userID).
func (r *Repository) SoftDeleteMemory(ctx context.Context, parishID, userID, memoryID uuid.UUID) error {
	row, err := r.FindMemoryByID(ctx, parishID, memoryID)
	if err != nil {
		return err
	}
	if row.Scope == agentmemory.ScopeUser {
		if row.UserID == nil || *row.UserID != userID {
			return domain.ErrForbidden
		}
	}
	return r.client.AgentMemory.UpdateOneID(row.ID).
		SetDeletedAt(time.Now().UTC()).
		Exec(ctx)
}

// FindMemoryByID returns a non-deleted memory fact in parishID.
func (r *Repository) FindMemoryByID(ctx context.Context, parishID, memoryID uuid.UUID) (*ent.AgentMemory, error) {
	row, err := r.client.AgentMemory.Query().
		Where(
			agentmemory.IDEQ(memoryID),
			agentmemory.ParishIDEQ(parishID),
			agentmemory.DeletedAtIsNil(),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, domain.ErrMemoryNotFound
	}
	return row, err
}

func (r *Repository) checkMemoryQuota(
	ctx context.Context,
	parishID uuid.UUID,
	userID *uuid.UUID,
	scope domain.MemoryScope,
	checkActive, checkPinned bool,
) error {
	baseQuery := func() *ent.AgentMemoryQuery {
		q := r.client.AgentMemory.Query().
			Where(
				agentmemory.ParishIDEQ(parishID),
				agentmemory.ScopeEQ(agentmemory.Scope(scope)),
				agentmemory.DeletedAtIsNil(),
			)
		if scope == domain.MemoryScopeUser && userID != nil {
			q = q.Where(agentmemory.UserIDEQ(*userID))
		}
		return q
	}

	if checkActive {
		count, err := baseQuery().Count(ctx)
		if err != nil {
			return err
		}
		if count >= domain.MaxActiveMemories {
			return domain.ErrMemoryLimitExceeded
		}
	}
	if checkPinned {
		pinnedCount, err := baseQuery().Where(agentmemory.IsPinnedEQ(true)).Count(ctx)
		if err != nil {
			return err
		}
		if pinnedCount >= domain.MaxPinnedMemories {
			return domain.ErrMemoryLimitExceeded
		}
	}
	return nil
}

// RecallMemories retrieves Tier 3 long-term facts using Hybrid RRF (BM25 + halfvec), BM25-only, or Ent fallback.
func (r *Repository) RecallMemories(ctx context.Context, in domain.RecallMemoriesInput) ([]domain.Memory, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	q := strings.TrimSpace(in.Query)

	if r.dialect == dialect.Postgres && r.db != nil && q != "" {
		if len(in.QueryEmbedding) == domain.EmbeddingDimensions {
			vecStr, err := formatHalfvec(in.QueryEmbedding)
			if err != nil {
				return nil, err
			}
			rows, err := r.recallHybridRRFPostgres(ctx, in, q, vecStr, limit)
			if err == nil {
				return rows, nil
			}
			rows, fbErr := r.recallHybridRRFILIKEPostgres(ctx, in, q, vecStr, limit)
			if fbErr == nil {
				return rows, nil
			}
		} else if len(in.QueryEmbedding) == 0 {
			rows, err := r.recallBM25Postgres(ctx, in, q, limit)
			if err == nil && len(rows) > 0 {
				return rows, nil
			}
		}
	}

	query := r.client.AgentMemory.Query().
		Where(
			agentmemory.ParishIDEQ(in.ParishID),
			agentmemory.DeletedAtIsNil(),
			agentmemory.Or(
				agentmemory.And(
					agentmemory.ScopeEQ(agentmemory.ScopeUser),
					agentmemory.UserIDEQ(in.UserID),
				),
				agentmemory.ScopeEQ(agentmemory.ScopeParish),
			),
		)
	if in.Category != nil {
		query = query.Where(agentmemory.CategoryEQ(agentmemory.Category(*in.Category)))
	}
	if q != "" {
		query = query.Where(agentmemory.ContentContainsFold(q))
	}
	entRows, err := query.
		Order(agentmemory.ByIsPinned(sql.OrderDesc()), agentmemory.ByUpdatedAt(sql.OrderDesc()), agentmemory.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainMemories(entRows), nil
}

func (r *Repository) recallHybridRRFPostgres(
	ctx context.Context,
	in domain.RecallMemoriesInput,
	q, vecStr string,
	limit int,
) ([]domain.Memory, error) {
	tx, err := r.db.BeginTx(ctx, &stdsql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SET LOCAL hnsw.iterative_scan = 'relaxed_order'`); err != nil {
		return nil, err
	}

	args := []any{q, in.ParishID, in.UserID, vecStr}
	catFilter := ""
	nextParam := 5
	if in.Category != nil {
		catFilter = fmt.Sprintf(" AND category = $%d", nextParam)
		args = append(args, string(*in.Category))
		nextParam++
	}
	limitParam := fmt.Sprintf("$%d", nextParam)
	args = append(args, limit)

	rawSQL := `
		WITH lexical AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY paradedb.score(id) DESC) AS rank
			FROM agent_memories
			WHERE parish_id = $2
			  AND deleted_at IS NULL
			  AND ((scope = 'user' AND user_id = $3) OR scope = 'parish')` + catFilter + `
			  AND content @@@ $1
			LIMIT 30
		),
		semantic AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY embedding <=> $4::halfvec) AS rank
			FROM agent_memories
			WHERE parish_id = $2
			  AND deleted_at IS NULL
			  AND ((scope = 'user' AND user_id = $3) OR scope = 'parish')` + catFilter + `
			  AND embedding IS NOT NULL
			LIMIT 30
		),
		rrf AS (
			SELECT COALESCE(l.id, s.id) AS id,
			       COALESCE(1.0 / (60 + l.rank), 0.0) + COALESCE(1.0 / (60 + s.rank), 0.0) AS score
			FROM lexical l
			FULL OUTER JOIN semantic s ON l.id = s.id
		)
		SELECT m.id, m.parish_id, m.user_id, m.scope, m.category, m.content, m.is_pinned,
		       m.source_session_id, rrf.score, m.created_at, m.updated_at
		FROM rrf
		JOIN agent_memories m ON m.id = rrf.id
		ORDER BY rrf.score DESC, m.is_pinned DESC, m.updated_at DESC, m.id DESC
		LIMIT ` + limitParam

	rows, err := tx.QueryContext(ctx, rawSQL, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDomainMemories(rows, true)
}

func (r *Repository) recallHybridRRFILIKEPostgres(
	ctx context.Context,
	in domain.RecallMemoriesInput,
	q, vecStr string,
	limit int,
) ([]domain.Memory, error) {
	tx, err := r.db.BeginTx(ctx, &stdsql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `SET LOCAL hnsw.iterative_scan = 'relaxed_order'`); err != nil {
		return nil, err
	}

	args := []any{"%" + q + "%", in.ParishID, in.UserID, vecStr}
	catFilter := ""
	nextParam := 5
	if in.Category != nil {
		catFilter = fmt.Sprintf(" AND category = $%d", nextParam)
		args = append(args, string(*in.Category))
		nextParam++
	}
	limitParam := fmt.Sprintf("$%d", nextParam)
	args = append(args, limit)

	rawSQL := `
		WITH lexical AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY is_pinned DESC, updated_at DESC, id DESC) AS rank
			FROM agent_memories
			WHERE parish_id = $2
			  AND deleted_at IS NULL
			  AND ((scope = 'user' AND user_id = $3) OR scope = 'parish')` + catFilter + `
			  AND content ILIKE $1
			LIMIT 30
		),
		semantic AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY embedding <=> $4::halfvec) AS rank
			FROM agent_memories
			WHERE parish_id = $2
			  AND deleted_at IS NULL
			  AND ((scope = 'user' AND user_id = $3) OR scope = 'parish')` + catFilter + `
			  AND embedding IS NOT NULL
			LIMIT 30
		),
		rrf AS (
			SELECT COALESCE(l.id, s.id) AS id,
			       COALESCE(1.0 / (60 + l.rank), 0.0) + COALESCE(1.0 / (60 + s.rank), 0.0) AS score
			FROM lexical l
			FULL OUTER JOIN semantic s ON l.id = s.id
		)
		SELECT m.id, m.parish_id, m.user_id, m.scope, m.category, m.content, m.is_pinned,
		       m.source_session_id, rrf.score, m.created_at, m.updated_at
		FROM rrf
		JOIN agent_memories m ON m.id = rrf.id
		ORDER BY rrf.score DESC, m.is_pinned DESC, m.updated_at DESC, m.id DESC
		LIMIT ` + limitParam

	rows, err := tx.QueryContext(ctx, rawSQL, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDomainMemories(rows, true)
}

func (r *Repository) recallBM25Postgres(
	ctx context.Context,
	in domain.RecallMemoriesInput,
	q string,
	limit int,
) ([]domain.Memory, error) {
	args := []any{q, in.ParishID, in.UserID}
	catFilter := ""
	nextParam := 4
	if in.Category != nil {
		catFilter = fmt.Sprintf(" AND category = $%d", nextParam)
		args = append(args, string(*in.Category))
		nextParam++
	}
	limitParam := fmt.Sprintf("$%d", nextParam)
	args = append(args, limit)

	rawSQL := `
		SELECT id, parish_id, user_id, scope, category, content, is_pinned,
		       source_session_id, paradedb.score(id) AS score, created_at, updated_at
		FROM agent_memories
		WHERE parish_id = $2
		  AND deleted_at IS NULL
		  AND ((scope = 'user' AND user_id = $3) OR scope = 'parish')` + catFilter + `
		  AND content @@@ $1
		ORDER BY paradedb.score(id) DESC, is_pinned DESC, updated_at DESC, id DESC
		LIMIT ` + limitParam

	rows, err := r.db.QueryContext(ctx, rawSQL, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanDomainMemories(rows, true)
}

// SearchMemories returns paginated memories for the Memory Drawer UI.
func (r *Repository) SearchMemories(ctx context.Context, in domain.SearchMemoriesInput) ([]*ent.AgentMemory, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = domain.HistoryWindow
	}
	q := strings.TrimSpace(in.Q)

	if q != "" && r.dialect == dialect.Postgres && r.db != nil {
		rows, err := r.searchMemoriesBM25Postgres(ctx, in, q, limit)
		if err == nil {
			return rows, nil
		}
	}

	query := r.client.AgentMemory.Query().
		Where(
			agentmemory.ParishIDEQ(in.ParishID),
			agentmemory.DeletedAtIsNil(),
			agentmemory.Or(
				agentmemory.And(
					agentmemory.ScopeEQ(agentmemory.ScopeUser),
					agentmemory.UserIDEQ(in.UserID),
				),
				agentmemory.ScopeEQ(agentmemory.ScopeParish),
			),
		)
	if in.Scope != "" {
		query = query.Where(agentmemory.ScopeEQ(agentmemory.Scope(in.Scope)))
	}
	if in.Category != "" {
		query = query.Where(agentmemory.CategoryEQ(agentmemory.Category(in.Category)))
	}
	if in.IsPinned != nil {
		query = query.Where(agentmemory.IsPinnedEQ(*in.IsPinned))
	}
	if q != "" {
		query = query.Where(agentmemory.ContentContainsFold(q))
	}
	if in.CursorAt != nil && in.CursorID != nil {
		query = query.Where(agentmemory.Or(
			agentmemory.CreatedAtLT(*in.CursorAt),
			agentmemory.And(
				agentmemory.CreatedAtEQ(*in.CursorAt),
				agentmemory.IDLT(*in.CursorID),
			),
		))
	}
	return query.
		Order(agentmemory.ByCreatedAt(sql.OrderDesc()), agentmemory.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

func (r *Repository) searchMemoriesBM25Postgres(
	ctx context.Context,
	in domain.SearchMemoriesInput,
	q string,
	limit int,
) ([]*ent.AgentMemory, error) {
	args := []any{q, in.ParishID, in.UserID}
	var extra strings.Builder
	nextParam := 4
	if in.Scope != "" {
		fmt.Fprintf(&extra, " AND scope = $%d", nextParam)
		args = append(args, in.Scope)
		nextParam++
	}
	if in.Category != "" {
		fmt.Fprintf(&extra, " AND category = $%d", nextParam)
		args = append(args, in.Category)
		nextParam++
	}
	if in.IsPinned != nil {
		fmt.Fprintf(&extra, " AND is_pinned = $%d", nextParam)
		args = append(args, *in.IsPinned)
		nextParam++
	}
	limitParam := fmt.Sprintf("$%d", nextParam)
	args = append(args, limit)

	rawSQL := `
		SELECT id, parish_id, user_id, scope, category, content, is_pinned,
		       source_session_id, created_at, updated_at
		FROM agent_memories
		WHERE parish_id = $2
		  AND deleted_at IS NULL
		  AND ((scope = 'user' AND user_id = $3) OR scope = 'parish')` + extra.String() + `
		  AND content @@@ $1
		ORDER BY paradedb.score(id) DESC, created_at DESC, id DESC
		LIMIT ` + limitParam

	rows, err := r.db.QueryContext(ctx, rawSQL, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []*ent.AgentMemory
	for rows.Next() {
		var m ent.AgentMemory
		var scope, cat string
		if err := rows.Scan(
			&m.ID, &m.ParishID, &m.UserID, &scope, &cat, &m.Content, &m.IsPinned,
			&m.SourceSessionID, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		m.Scope = agentmemory.Scope(scope)
		m.Category = agentmemory.Category(cat)
		out = append(out, &m)
	}
	return out, rows.Err()
}

func scanDomainMemories(rows *stdsql.Rows, withScore bool) ([]domain.Memory, error) {
	var out []domain.Memory
	for rows.Next() {
		var m domain.Memory
		var scope, cat string
		if withScore {
			if err := rows.Scan(
				&m.ID, &m.ParishID, &m.UserID, &scope, &cat, &m.Content, &m.IsPinned,
				&m.SourceSessionID, &m.Score, &m.CreatedAt, &m.UpdatedAt,
			); err != nil {
				return nil, err
			}
		} else {
			if err := rows.Scan(
				&m.ID, &m.ParishID, &m.UserID, &scope, &cat, &m.Content, &m.IsPinned,
				&m.SourceSessionID, &m.CreatedAt, &m.UpdatedAt,
			); err != nil {
				return nil, err
			}
		}
		m.Scope = domain.MemoryScope(scope)
		m.Category = domain.MemoryCategory(cat)
		out = append(out, m)
	}
	return out, rows.Err()
}

func toDomainMemories(rows []*ent.AgentMemory) []domain.Memory {
	out := make([]domain.Memory, len(rows))
	for i, r := range rows {
		out[i] = domain.Memory{
			ID:              r.ID,
			ParishID:        r.ParishID,
			UserID:          r.UserID,
			Scope:           domain.MemoryScope(r.Scope),
			Category:        domain.MemoryCategory(r.Category),
			Content:         r.Content,
			IsPinned:        r.IsPinned,
			SourceSessionID: r.SourceSessionID,
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
		}
	}
	return out
}
