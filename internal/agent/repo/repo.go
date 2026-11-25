package repo

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/paroki/domus/ent"
)

// EmbeddingDimensions is the fixed vector width stored in agent_memories.embedding.
const EmbeddingDimensions = 768

// Repository provides persistence and ParadeDB hybrid RAG search for agent sessions, messages, and memories.
type Repository struct {
	client *ent.Client
	sqlDB  *sql.DB
}

// New constructs an agent Repository backed by Ent and raw *sql.DB for ParadeDB/halfvec queries.
func New(client *ent.Client, sqlDB *sql.DB) *Repository {
	return &Repository{
		client: client,
		sqlDB:  sqlDB,
	}
}

// EnsureMemorySearchSchema idempotently adds the halfvec(768) embedding column and
// ParadeDB BM25 + HNSW indexes on agent_memories (used by ParadeDB integration tests).
func EnsureMemorySearchSchema(ctx context.Context, sqlDB *sql.DB) error {
	if sqlDB == nil {
		return nil
	}
	stmts := []string{
		`ALTER TABLE "agent_memories" ADD COLUMN IF NOT EXISTS "embedding" halfvec(768) NULL`,
		`CREATE INDEX IF NOT EXISTS "agent_memories_bm25_idx" ON "agent_memories" USING bm25 ("id", "content", "category", "paroki_id", "user_id") WITH (key_field = 'id')`,
		`CREATE INDEX IF NOT EXISTS "agent_memories_embedding_hnsw_idx" ON "agent_memories" USING hnsw ("embedding" halfvec_cosine_ops) WITH (m = 16, ef_construction = 64) WHERE "embedding" IS NOT NULL`,
	}
	for _, stmt := range stmts {
		if _, err := sqlDB.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func formatHalfvecLiteral(vec []float32) (string, error) {
	if len(vec) != EmbeddingDimensions {
		return "", fmt.Errorf("expected %d embedding dimensions, got %d", EmbeddingDimensions, len(vec))
	}
	var b strings.Builder
	b.Grow(len(vec) * 8)
	b.WriteByte('[')
	for i, v := range vec {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return "", fmt.Errorf("invalid embedding float at index %d", i)
		}
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String(), nil
}
