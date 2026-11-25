package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const defaultTestDSN = "postgres://domus:domus@localhost:5432/domus_test?sslmode=disable"

func TestRootCmd_SubcommandsAndMigrateCLI(t *testing.T) {
	root := newRootCmd()
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	if !names["serve"] || !names["migrate"] {
		t.Fatalf("expected serve and migrate subcommands, got %v", names)
	}

	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = defaultTestDSN
	}

	ctx := context.Background()
	adminDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres admin: %v", err)
	}
	if _, err := adminDB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		_ = adminDB.Close()
		t.Fatalf("enable vector extension: %v", err)
	}
	if _, err := adminDB.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS pg_search CASCADE"); err != nil {
		_ = adminDB.Close()
		t.Fatalf("enable pg_search extension: %v", err)
	}

	schema := fmt.Sprintf("cli_migrate_test_%d", os.Getpid())
	if _, err := adminDB.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
		_ = adminDB.Close()
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema+",public")
	u.RawQuery = q.Encode()

	t.Setenv("APP_DB_URL", u.String())

	// Default (APP_AUTO_MIGRATE=false) is a no-op when migrations.FS has no SQL files.
	t.Setenv("APP_AUTO_MIGRATE", "")
	root.SetArgs([]string{"migrate"})
	if err := root.ExecuteContext(ctx); err != nil {
		t.Fatalf("domus migrate (default) failed: %v", err)
	}

	// Opt-in auto-migrate creates Ent tables + agent memory search schema idempotently.
	t.Setenv("APP_AUTO_MIGRATE", "true")
	for i := range 2 {
		root.SetArgs([]string{"migrate"})
		if err := root.ExecuteContext(ctx); err != nil {
			t.Fatalf("domus migrate (auto-migrate run %d) failed: %v", i+1, err)
		}
	}

	schemaDB, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("open schema db: %v", err)
	}
	defer func() { _ = schemaDB.Close() }()

	for _, tbl := range []string{"users", "settings", "agent_providers", "agent_memories"} {
		var found string
		if err := schemaDB.QueryRowContext(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1", tbl).Scan(&found); err != nil {
			t.Fatalf("expected %s table to exist: %v", tbl, err)
		}
	}

	var colName string
	if err := schemaDB.QueryRowContext(ctx, "SELECT column_name FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'agent_memories' AND column_name = 'embedding'").Scan(&colName); err != nil {
		t.Fatalf("expected agent_memories.embedding column to exist: %v", err)
	}

	for _, idx := range []string{"agent_memories_bm25_idx", "agent_memories_embedding_hnsw_idx"} {
		var found string
		if err := schemaDB.QueryRowContext(ctx, "SELECT indexname FROM pg_indexes WHERE schemaname = current_schema() AND tablename = 'agent_memories' AND indexname = $1", idx).Scan(&found); err != nil {
			t.Fatalf("expected index %s to exist: %v", idx, err)
		}
	}
}
