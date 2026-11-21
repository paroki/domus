package repo

import (
	"context"
	"fmt"

	"github.com/paroki/domus/ent"
)

// Repository provides Ent-backed persistence for event entities.
type Repository struct {
	client *ent.Client
}

// New creates a Repository backed by the given Ent client.
func New(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// WithTx executes fn inside a database transaction.
func (r *Repository) WithTx(ctx context.Context, fn func(txRepo *Repository) error) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	txRepo := &Repository{client: tx.Client()}
	if err := fn(txRepo); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w (rollback failed: %v)", err, rerr)
		}
		return err
	}
	return tx.Commit()
}
