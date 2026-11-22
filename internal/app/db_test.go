package app

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
)

func TestOpenDB_SQLite(t *testing.T) {
	ctx := context.Background()
	for _, dsn := range []string{"sqlite://:memory:", "file:app_db_test?mode=memory&cache=shared"} {
		client, err := OpenDB(ctx, Config{DatabaseURL: dsn})
		if err != nil {
			t.Fatalf("OpenDB(%q) failed: %v", dsn, err)
		}
		if err := client.Schema.Create(ctx); err != nil {
			_ = client.Close()
			t.Fatalf("Schema.Create(%q) failed: %v", dsn, err)
		}
		if _, err := client.Setting.Create().SetKey("k1").SetValue("v1").Save(ctx); err != nil {
			_ = client.Close()
			t.Fatalf("Setting.Create(%q) failed: %v", dsn, err)
		}
		_ = client.Close()
	}
}

func TestGetConfig_DBDefaults(t *testing.T) {
	cfg, err := GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if !strings.HasPrefix(cfg.DatabaseURL, "postgres://") {
		t.Fatalf("expected postgres:// default DatabaseURL, got %q", cfg.DatabaseURL)
	}
}

func TestCheckPostgresExtensions(t *testing.T) {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		t.Skip("TEST_DSN not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := checkPostgresExtensions(ctx, db); err != nil {
		t.Fatalf("expected required extensions to be installed, got: %v", err)
	}

	err = checkPostgresExtensions(ctx, db, "vector", "missing_ext_test")
	if err == nil || !strings.Contains(err.Error(), "missing_ext_test") {
		t.Fatalf("expected missing_ext_test error, got: %v", err)
	}
}
