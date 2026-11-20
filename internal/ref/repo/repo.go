package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/ward"
	"github.com/paroki/domus/ent/zone"
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

// SetUserDefaultParishIfNil sets User.default_parish_id only when currently unset.
func (r *Repository) SetUserDefaultParishIfNil(ctx context.Context, userID, parishID uuid.UUID) error {
	_, err := r.client.User.Update().
		Where(user.IDEQ(userID), user.DefaultParishIDIsNil()).
		SetDefaultParishID(parishID).
		Save(ctx)
	return err
}

// FindZoneInParish returns a non-deleted Zone belonging to parishID.
func (r *Repository) FindZoneInParish(ctx context.Context, parishID, zoneID uuid.UUID) (*ent.Zone, error) {
	return r.client.Zone.Query().
		Where(zone.IDEQ(zoneID), zone.ParishIDEQ(parishID), zone.DeletedAtIsNil()).
		Only(ctx)
}

// FindWardInParish returns a non-deleted Ward belonging to parishID.
func (r *Repository) FindWardInParish(ctx context.Context, parishID, wardID uuid.UUID) (*ent.Ward, error) {
	return r.client.Ward.Query().
		Where(ward.IDEQ(wardID), ward.ParishIDEQ(parishID), ward.DeletedAtIsNil()).
		Only(ctx)
}

// FindPeriodInParish returns a non-deleted Period belonging to parishID.
func (r *Repository) FindPeriodInParish(ctx context.Context, parishID, periodID uuid.UUID) (*ent.Period, error) {
	return r.client.Period.Query().
		Where(period.IDEQ(periodID), period.ParishIDEQ(parishID), period.DeletedAtIsNil()).
		Only(ctx)
}
