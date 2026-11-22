package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paroki/domus/ent"
)

// OpenDB opens a database connection pool for PostgreSQL and returns an Ent client.
func OpenDB(ctx context.Context, cfg Config) (*ent.Client, error) {
	client, _, _, err := OpenDBPool(ctx, cfg)
	return client, err
}

// OpenDBPool opens a database connection pool and returns the Ent client, underlying *sql.DB, and dialect name.
func OpenDBPool(ctx context.Context, cfg Config) (*ent.Client, *sql.DB, string, error) {
	switch {
	case strings.HasPrefix(cfg.DatabaseURL, "postgres://"), strings.HasPrefix(cfg.DatabaseURL, "postgresql://"):
		db, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return nil, nil, "", fmt.Errorf("open postgres: %w", err)
		}
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, nil, "", fmt.Errorf("ping postgres: %w", err)
		}
		if err := checkPostgresExtensions(ctx, db); err != nil {
			_ = db.Close()
			return nil, nil, "", err
		}
		drv := entsql.OpenDB(dialect.Postgres, db)
		return ent.NewClient(ent.Driver(drv)), db, dialect.Postgres, nil

	default:
		return nil, nil, "", fmt.Errorf("unsupported database scheme: %q", cfg.DatabaseURL)
	}
}

var requiredPostgresExtensions = []string{"vector", "pg_search"}

// checkPostgresExtensions verifies that all required PostgreSQL extensions are installed.
func checkPostgresExtensions(ctx context.Context, db *sql.DB, required ...string) error {
	if len(required) == 0 {
		required = requiredPostgresExtensions
	}
	rows, err := db.QueryContext(ctx, "SELECT extname FROM pg_extension WHERE extname = ANY($1)", required)
	if err != nil {
		return fmt.Errorf("check postgres extensions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	installed := make(map[string]bool, len(required))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan postgres extension: %w", err)
		}
		installed[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate postgres extensions: %w", err)
	}

	var missing []string
	for _, ext := range required {
		if !installed[ext] {
			missing = append(missing, ext)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required postgres extension(s): %s", strings.Join(missing, ", "))
	}
	return nil
}
