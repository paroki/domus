package config

import (
	"context"
	"fmt"
	"time"

	"github.com/paroki/domus/api/ent"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseConfig holds database configuration fields.
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// GetDB initializes the ent client and runs auto-migration.
func GetDB(cfg *Config) (*ent.Client, error) {
	client, err := ent.Open(cfg.DB.Driver, cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// Run auto-migration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("run auto-migration: %w", err)
	}

	return client, nil
}
