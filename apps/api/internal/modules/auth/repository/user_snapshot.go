package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
	"github.com/paroki/domus/api/internal/platform/database"
)

type UserSnapshotRepository struct {
	entcli *ent.Client
	log    *slog.Logger
}

func NewUserSnapshotRepository(entcli *ent.Client, log *slog.Logger) UserSnapshotRepository {
	return UserSnapshotRepository{entcli, log}
}

func (r UserSnapshotRepository) GetByID(c context.Context, id uuid.UUID) (*model.UserSnapshotResponse, error) {
	var response model.UserSnapshotResponse

	user, err := r.entcli.User.Get(c, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, core.ErrItemNotFound
		}
		return nil, err
	}

	core.ToValue(user, &response)

	return &response, nil
}

func (r UserSnapshotRepository) Create(c context.Context, req model.CreateUserSnapshotRequest) (*model.UserSnapshotResponse, error) {
	var res model.UserSnapshotResponse
	var user *ent.User
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		var err error
		user, err = tx.User.Create().
			SetID(req.ID).
			SetName(req.Name).
			SetEmail(req.Email).
			SetNillableAvatar(nilIfEmpty(req.Avatar)).
			Save(c)

		return err
	})

	if err != nil {
		return nil, err
	}

	core.ToValue(user, &res)

	return &res, nil
}

func (r UserSnapshotRepository) Update(c context.Context, id uuid.UUID, req model.UpdateUserSnapshotRequest) (*model.UserSnapshotResponse, error) {
	var res model.UserSnapshotResponse
	var user *ent.User
	err := database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		var err error
		user, err = tx.User.UpdateOneID(id).
			SetName(req.Name).
			SetEmail(req.Email).
			SetNillableAvatar(nilIfEmpty(req.Avatar)).
			Save(c)

		return err
	})

	if err != nil {
		return nil, err
	}

	core.ToValue(user, &res)

	return &res, nil
}

// nilIfEmpty maps "" to nil so empty optional fields are stored as NULL.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r UserSnapshotRepository) Delete(c context.Context, id uuid.UUID) error {
	return database.WithTx(c, r.entcli, func(tx *ent.Tx) error {
		return tx.User.DeleteOneID(id).Exec(c)
	})
}
