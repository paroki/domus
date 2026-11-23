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
)

const defaultTestDSN = "postgres://domus:domus@localhost:5432/domus_test?sslmode=disable"

var entDBCounter atomic.Uint64

// NewEntClient creates an isolated Ent client with migrated schema on ParadeDB.
func NewEntClient(t testing.TB) *ent.Client {
	t.Helper()
	client, _, _ := NewEntClientWithDB(t)
	return client
}

// NewEntClientWithDB creates an isolated Ent client on ParadeDB and returns its underlying *sql.DB and dialect name.
func NewEntClientWithDB(t testing.TB) (*ent.Client, *sql.DB, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = defaultTestDSN
	}
	return newPostgresEntClient(t, dsn)
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

	t.Cleanup(func() {
		_ = client.Close()
		_, _ = adminDB.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
		_ = adminDB.Close()
	})
	return client, db, dialect.Postgres
}
