package main

import (
	"context"
	"path/filepath"
	"strings"
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
	t.Run("AutoMigrate=true aborts on migration failure", func(t *testing.T) {
		t.Setenv("AUTO_MIGRATE", "true")
		t.Setenv("APP_DB_URL", "mysql://invalid")

		root := newRootCmd()
		root.SetArgs([]string{"serve"})
		err := root.ExecuteContext(context.Background())
		if err == nil || !strings.Contains(err.Error(), "unsupported database scheme") {
			t.Fatalf("expected unsupported database scheme error, got %v", err)
		}
	})

	t.Run("AutoMigrate=false skips migrations", func(t *testing.T) {
		t.Setenv("AUTO_MIGRATE", "false")
		t.Setenv("APP_DB_URL", "mysql://invalid")
		t.Setenv("HOST", "127.0.0.1")
		t.Setenv("PORT", "0")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		root := newRootCmd()
		root.SetArgs([]string{"serve"})
		err := root.ExecuteContext(ctx)
		if err != nil && strings.Contains(err.Error(), "unsupported database scheme") {
			t.Fatalf("expected migrations to be skipped when AUTO_MIGRATE=false, got %v", err)
		}
	})
}
