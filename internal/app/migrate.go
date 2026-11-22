package app

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/paroki/domus/migrations"
	_ "modernc.org/sqlite"
)

var (
	sqliteTimestamptzRe     = regexp.MustCompile(`(?i)\btimestamptz\b`)
	sqliteJsonbRe           = regexp.MustCompile(`(?i)\bjsonb\b`)
	sqliteCreateExtensionRe = regexp.MustCompile(`(?is)CREATE\s+EXTENSION\b[^;]*;`)
	sqliteHalfvecColRe      = regexp.MustCompile(`(?is)ALTER\s+TABLE\b[^;]*\bhalfvec\b[^;]*;`)
	sqlitePgUsingIndexRe    = regexp.MustCompile(`(?is)CREATE\s+INDEX\b[^;]*\bUSING\s+(?:bm25|hnsw)\b[^;]*;`)
)

func normalizeSQLiteMigration(stmt string) string {
	stmt = sqliteCreateExtensionRe.ReplaceAllString(stmt, "")
	stmt = sqliteHalfvecColRe.ReplaceAllString(stmt, "")
	stmt = sqlitePgUsingIndexRe.ReplaceAllString(stmt, "")
	stmt = sqliteTimestamptzRe.ReplaceAllString(stmt, "timestamp")
	stmt = sqliteJsonbRe.ReplaceAllString(stmt, "text")
	return stmt
}

// RunMigrations applies embedded SQL migrations in lexicographical order and tracks checksums.
func RunMigrations(ctx context.Context, cfg Config, logger *slog.Logger) error {
	return runMigrationsFromFS(ctx, cfg, logger, migrations.FS)
}

func runMigrationsFromFS(ctx context.Context, cfg Config, logger *slog.Logger, fsys fs.FS) error {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	db, err := openRawDB(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	isSQLite := strings.HasPrefix(cfg.DatabaseURL, "sqlite://") || strings.HasPrefix(cfg.DatabaseURL, "file:")
	if !isSQLite {
		if err := checkPostgresExtensions(ctx, db); err != nil {
			return err
		}
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			checksum TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}
	slices.Sort(files)

	for _, name := range files {
		raw, err := fs.ReadFile(fsys, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		sum := sha256.Sum256(raw)
		checksum := hex.EncodeToString(sum[:])

		var existing string
		err = db.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE version = $1", name).Scan(&existing)
		switch {
		case err == nil:
			if existing != checksum {
				return fmt.Errorf("migration checksum mismatch for %s: expected %s, got %s", name, existing, checksum)
			}
			continue
		case errors.Is(err, sql.ErrNoRows):
			// Not yet applied; proceed inside a transaction.
		default:
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration tx %s: %w", name, err)
		}
		stmt := string(raw)
		if isSQLite {
			stmt = normalizeSQLiteMigration(stmt)
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)", name, checksum); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
		logger.InfoContext(ctx, "applied migration", "version", name)
	}

	return nil
}

func openRawDB(ctx context.Context, cfg Config) (*sql.DB, error) {
	var driver, dsn string
	switch {
	case strings.HasPrefix(cfg.DatabaseURL, "postgres://"), strings.HasPrefix(cfg.DatabaseURL, "postgresql://"):
		driver, dsn = "pgx", cfg.DatabaseURL
	case strings.HasPrefix(cfg.DatabaseURL, "sqlite://"), strings.HasPrefix(cfg.DatabaseURL, "file:"):
		driver, dsn = "sqlite", normalizeSQLiteDSN(cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("unsupported database scheme: %q", cfg.DatabaseURL)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", driver, err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping %s: %w", driver, err)
	}
	return db, nil
}
