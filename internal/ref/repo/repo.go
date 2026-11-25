package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/lingkungan"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/wilayah"
)

// Repository provides Ent-backed persistence for ref entities.
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

// SetUserDefaultParokiIfNil sets User.default_paroki_id only when currently unset.
func (r *Repository) SetUserDefaultParokiIfNil(ctx context.Context, userID, parokiID uuid.UUID) error {
	_, err := r.client.User.Update().
		Where(user.IDEQ(userID), user.DefaultParokiIDIsNil()).
		SetDefaultParokiID(parokiID).
		Save(ctx)
	return err
}

// FindWilayahInParoki returns a non-deleted Wilayah belonging to parokiID.
func (r *Repository) FindWilayahInParoki(ctx context.Context, parokiID, wilayahID uuid.UUID) (*ent.Wilayah, error) {
	return r.client.Wilayah.Query().
		Where(wilayah.IDEQ(wilayahID), wilayah.ParokiIDEQ(parokiID), wilayah.DeletedAtIsNil()).
		Only(ctx)
}

// FindLingkunganInParoki returns a non-deleted Lingkungan belonging to parokiID.
func (r *Repository) FindLingkunganInParoki(ctx context.Context, parokiID, lingkunganID uuid.UUID) (*ent.Lingkungan, error) {
	return r.client.Lingkungan.Query().
		Where(lingkungan.IDEQ(lingkunganID), lingkungan.ParokiIDEQ(parokiID), lingkungan.DeletedAtIsNil()).
		Only(ctx)
}

// FindPeriodInParoki returns a non-deleted Period belonging to parokiID.
func (r *Repository) FindPeriodInParoki(ctx context.Context, parokiID, periodID uuid.UUID) (*ent.Period, error) {
	return r.client.Period.Query().
		Where(period.IDEQ(periodID), period.ParokiIDEQ(parokiID), period.DeletedAtIsNil()).
		Only(ctx)
}
