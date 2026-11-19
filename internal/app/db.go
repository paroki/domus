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
	_ "modernc.org/sqlite"
)

// OpenDB opens a database connection pool for PostgreSQL or SQLite and returns an Ent client.
func OpenDB(ctx context.Context, cfg Config) (*ent.Client, error) {
	switch {
	case strings.HasPrefix(cfg.DatabaseURL, "postgres://"), strings.HasPrefix(cfg.DatabaseURL, "postgresql://"):
		db, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		if cfg.DBMaxOpenConns > 0 {
			db.SetMaxOpenConns(cfg.DBMaxOpenConns)
		}
		if cfg.DBMaxIdleConns > 0 {
			db.SetMaxIdleConns(cfg.DBMaxIdleConns)
		}
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ping postgres: %w", err)
		}
		drv := entsql.OpenDB(dialect.Postgres, db)
		return ent.NewClient(ent.Driver(drv)), nil

	case strings.HasPrefix(cfg.DatabaseURL, "sqlite://"), strings.HasPrefix(cfg.DatabaseURL, "file:"):
		dsn := normalizeSQLiteDSN(cfg.DatabaseURL)
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("ping sqlite: %w", err)
		}
		drv := entsql.OpenDB(dialect.SQLite, db)
		return ent.NewClient(ent.Driver(drv)), nil

	default:
		return nil, fmt.Errorf("unsupported database scheme: %q", cfg.DatabaseURL)
	}
}

func normalizeSQLiteDSN(raw string) string {
	dsn := strings.TrimPrefix(raw, "sqlite://")
	if !strings.Contains(dsn, "foreign_keys") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "_pragma=foreign_keys(1)"
	}
	return dsn
}
