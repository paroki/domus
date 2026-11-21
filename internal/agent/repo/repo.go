package repo

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/paroki/domus/ent"
)

// Repository provides Ent and ParadeDB-backed persistence for the AI agent feature.
type Repository struct {
	client  *ent.Client
	db      *sql.DB
	dialect string
}

// New creates a Repository backed by the given Ent client, optional *sql.DB, and dialect name.
func New(client *ent.Client, db *sql.DB, dialectName string) *Repository {
	if dialectName == "" {
		dialectName = dialect.Postgres
	}
	return &Repository{
		client:  client,
		db:      db,
		dialect: dialectName,
	}
}

// WithTx executes fn inside an Ent database transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(txRepo *Repository) error) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	txRepo := &Repository{
		client:  tx.Client(),
		db:      r.db,
		dialect: r.dialect,
	}
	if err := fn(txRepo); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rerr)
		}
		return err
	}
	return tx.Commit()
}
