package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestRootCmd_SubcommandsAndMigrateCLI(t *testing.T) {
	root := newRootCmd()
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	if !names["serve"] || !names["migrate"] {
		t.Fatalf("expected serve and migrate subcommands, got %v", names)
	}

	dbPath := filepath.Join(t.TempDir(), "cli.db")
	t.Setenv("APP_DB_URL", "sqlite://"+dbPath)

	root.SetArgs([]string{"migrate"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("domus migrate failed: %v", err)
	}
}

func TestServeCmd_AutoMigrateToggle(t *testing.T) {
	t.Run("AutoMigrate=false skips migrations", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "skip.db")
		t.Setenv("AUTO_MIGRATE", "false")
		t.Setenv("APP_DB_URL", "sqlite://"+dbPath)
		t.Setenv("HOST", "127.0.0.1")
		t.Setenv("PORT", "0")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		root := newRootCmd()
		root.SetArgs([]string{"serve"})
		_ = root.ExecuteContext(ctx)

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		defer func() { _ = db.Close() }()

		var name string
		err = db.QueryRowContext(context.Background(), "SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&name)
		if err == nil {
			t.Fatal("expected schema_migrations table not to exist when AUTO_MIGRATE=false")
		}
	})
}
