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

	root.SetArgs([]string{"migrate"})
	if err := root.ExecuteContext(ctx); err != nil {
		t.Fatalf("domus migrate failed: %v", err)
	}
}
