package app

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
)

const defaultTestDSN = "postgres://domus:domus@localhost:5432/domus_test?sslmode=disable"

func testDSN() string {
	if dsn := os.Getenv("TEST_DSN"); dsn != "" {
		return dsn
	}
	return defaultTestDSN
}

func TestOpenDB_UnsupportedScheme(t *testing.T) {
	_, err := OpenDB(context.Background(), Config{DatabaseURL: "mysql://localhost:3306/domus"})
	if err == nil || !strings.Contains(err.Error(), "unsupported database scheme") {
		t.Fatalf("expected unsupported database scheme error, got %v", err)
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
	ctx := context.Background()
	db, err := sql.Open("pgx", testDSN())
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
