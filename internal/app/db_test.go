package app

import (
	"context"
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
		if _, err := client.Meta.Create().SetKey("k1").SetValue("v1").Save(ctx); err != nil {
			_ = client.Close()
			t.Fatalf("Meta.Create(%q) failed: %v", dsn, err)
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
