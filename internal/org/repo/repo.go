package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
)

// Repository provides Ent-backed persistence for org entities.
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

// FindUserByID returns a non-deleted user by ID.
func (r *Repository) FindUserByID(ctx context.Context, id uuid.UUID) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.IDEQ(id), user.DeletedAtIsNil()).
		Only(ctx)
}

// FindUserByEmail returns a non-deleted user by email.
func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.EmailEQ(email), user.DeletedAtIsNil()).
		Only(ctx)
}
