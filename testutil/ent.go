package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/enttest"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var entDBCounter atomic.Uint64

// NewEntClient creates an isolated Ent client with migrated schema.
// When TEST_DSN is set, it provisions a per-test Postgres schema; otherwise it
// falls back to an in-memory SQLite database.
func NewEntClient(t testing.TB) *ent.Client {
	t.Helper()
	client, _, _ := NewEntClientWithDB(t)
	return client
}

// NewEntClientWithDB creates an isolated Ent client and returns its underlying *sql.DB and dialect name.
func NewEntClientWithDB(t testing.TB) (*ent.Client, *sql.DB, string) {
	t.Helper()
	if dsn := os.Getenv("TEST_DSN"); dsn != "" {
		return newPostgresEntClient(t, dsn)
	}
	return newSQLiteEntClient(t)
}

func newSQLiteEntClient(t testing.TB) (*ent.Client, *sql.DB, string) {
	t.Helper()
	id := entDBCounter.Add(1)
	dsn := fmt.Sprintf("file:ent_%d?mode=memory&cache=shared&_pragma=foreign_keys(1)", id)
	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "open in-memory sqlite")

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(ent.Driver(drv)))
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client, db, dialect.SQLite
}

func newPostgresEntClient(t testing.TB, dsn string) (*ent.Client, *sql.DB, string) {
	t.Helper()
	schema := fmt.Sprintf("test_%d_%d", os.Getpid(), entDBCounter.Add(1))

	adminDB, err := sql.Open("pgx", dsn)
	require.NoError(t, err, "open postgres admin connection")

	ctx := context.Background()
	_, err = adminDB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	require.NoError(t, err, "enable vector extension")
	_, err = adminDB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pg_search CASCADE")
	require.NoError(t, err, "enable pg_search extension")

	_, err = adminDB.ExecContext(ctx, "CREATE SCHEMA "+schema)
	require.NoError(t, err, "create isolated test schema")

	u, err := url.Parse(dsn)
	require.NoError(t, err, "parse TEST_DSN")
	q := u.Query()
	q.Set("search_path", schema+",public")
	u.RawQuery = q.Encode()

	db, err := sql.Open("pgx", u.String())
	require.NoError(t, err, "open postgres schema connection")

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := enttest.NewClient(t, enttest.WithOptions(ent.Driver(drv)))

	for _, stmt := range []string{
		`ALTER TABLE "agent_memories" ADD COLUMN IF NOT EXISTS "embedding" halfvec(768) NULL`,
		`CREATE INDEX IF NOT EXISTS "agent_messages_bm25_idx" ON "agent_messages" USING bm25 ("id", "content", "parish_id", "user_id", "session_id", "role") WITH (key_field = 'id')`,
		`CREATE INDEX IF NOT EXISTS "agent_memories_bm25_idx" ON "agent_memories" USING bm25 ("id", "content", "category", "scope", "parish_id", "user_id") WITH (key_field = 'id')`,
		`CREATE INDEX IF NOT EXISTS "agent_memories_embedding_hnsw_idx" ON "agent_memories" USING hnsw ("embedding" halfvec_cosine_ops) WITH (m = 16, ef_construction = 64) WHERE "deleted_at" IS NULL AND "embedding" IS NOT NULL`,
	} {
		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err, "apply postgres agent DDL: %s", stmt)
	}

	t.Cleanup(func() {
		_ = client.Close()
		_, _ = adminDB.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
		_ = adminDB.Close()
	})
	return client, db, dialect.Postgres
}
