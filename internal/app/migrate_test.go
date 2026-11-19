package app

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/paroki/domus/migrations"
)

func TestRunMigrations_IdempotentSQLite(t *testing.T) {
	ctx := context.Background()
	dsn := "file:migrate_idempotent?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping sqlite: %v", err)
	}

	cfg := Config{DatabaseURL: dsn}
	logger := slog.New(slog.DiscardHandler)

	if err := RunMigrations(ctx, cfg, logger); err != nil {
		t.Fatalf("first RunMigrations failed: %v", err)
	}
	// Second run must be idempotent.
	if err := RunMigrations(ctx, cfg, logger); err != nil {
		t.Fatalf("second RunMigrations failed: %v", err)
	}

	sqlFiles, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count != len(sqlFiles) {
		t.Fatalf("expected %d applied migrations, got %d", len(sqlFiles), count)
	}

	var metaTable string
	if err := db.QueryRowContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name='meta'").Scan(&metaTable); err != nil {
		t.Fatalf("expected meta table to exist: %v", err)
	}
}

func TestRunMigrations_ChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	dsn := "file:migrate_mismatch?mode=memory&cache=shared"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping sqlite: %v", err)
	}

	cfg := Config{DatabaseURL: dsn}

	fs1 := fstest.MapFS{
		"001_init.sql": &fstest.MapFile{Data: []byte("CREATE TABLE t1 (id TEXT PRIMARY KEY);")},
	}
	if err := runMigrationsFromFS(ctx, cfg, nil, fs1); err != nil {
		t.Fatalf("initial migration failed: %v", err)
	}

	fs2 := fstest.MapFS{
		"001_init.sql": &fstest.MapFile{Data: []byte("CREATE TABLE t1 (id TEXT PRIMARY KEY, extra TEXT);")},
	}
	err = runMigrationsFromFS(ctx, cfg, nil, fs2)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
}
