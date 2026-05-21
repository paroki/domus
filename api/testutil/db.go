package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/paroki/domus/api/ent"

	_ "github.com/mattn/go-sqlite3"
)

// CloseFunc defines a function signature to clean up DB connections.
type CloseFunc func()

// NewTestDB returns an isolated *ent.Client for use in integration tests.
// Uses SQLite in-memory by default; can be configured via TEST_DB_DSN env var.
// Automatically runs schema migration.
func NewTestDB(t *testing.T) (*ent.Client, CloseFunc) {
	driver := "sqlite3"
	dsn := "file:ent?mode=memory&cache=shared&_fk=1"

	if envDsn := os.Getenv("TEST_DB_DSN"); envDsn != "" {
		driver = "postgres"
		dsn = envDsn
	}

	client, err := ent.Open(driver, dsn)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		t.Fatalf("failed to run test schema auto-migration: %v", err)
	}

	closeFn := func() {
		if err := client.Close(); err != nil {
			t.Errorf("failed to close test database: %v", err)
		}
	}

	return client, closeFn
}
