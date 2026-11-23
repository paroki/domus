package app

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/migrations"
)

var migrateSchemaCounter atomic.Uint64

func newTestPostgresSchemaDSN(t *testing.T) (string, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	dsn := testDSN()

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

	schema := fmt.Sprintf("migrate_test_%d_%d", os.Getpid(), migrateSchemaCounter.Add(1))
	if _, err := adminDB.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}

	u, err := url.Parse(dsn)
	if err != nil {
		_ = adminDB.Close()
		t.Fatalf("parse dsn: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema+",public")
	u.RawQuery = q.Encode()
	schemaDSN := u.String()

	db, err := sql.Open("pgx", schemaDSN)
	if err != nil {
		_ = adminDB.Close()
		t.Fatalf("open postgres schema: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		_ = adminDB.Close()
		t.Fatalf("ping postgres schema: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		_, _ = adminDB.ExecContext(context.Background(), "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
		_ = adminDB.Close()
	})
	return schemaDSN, db
}

func TestRunMigrations_Idempotent(t *testing.T) {
	ctx := context.Background()
	dsn, db := newTestPostgresSchemaDSN(t)

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

	for _, tbl := range []string{"settings", "agent_providers"} {
		var found string
		if err := db.QueryRowContext(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1", tbl).Scan(&found); err != nil {
			t.Fatalf("expected %s table to exist: %v", tbl, err)
		}
	}

	client, err := OpenDB(ctx, cfg)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer func() { _ = client.Close() }()

	created, err := client.User.Create().
		SetEmail("migrate-test@domus.test").
		SetName("Migrate Test").
		Save(ctx)
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	got, err := client.User.Query().Where(user.IDEQ(created.ID)).Only(ctx)
	if err != nil {
		t.Fatalf("query user failed: %v", err)
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected non-zero created_at")
	}
}

func TestRunMigrations_ChecksumMismatch(t *testing.T) {
	ctx := context.Background()
	dsn, _ := newTestPostgresSchemaDSN(t)

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
	err := runMigrationsFromFS(ctx, cfg, nil, fs2)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("expected checksum mismatch error, got %v", err)
	}
}
