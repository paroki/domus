package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/modules/auth/model"
)

type UserSnapshotRepository interface {
	GetByID(c context.Context, id uuid.UUID) (*model.UserSnapshotResponse, error)
	Create(c context.Context, req model.CreateUserSnapshotRequest) (*model.UserSnapshotResponse, error)
	Update(c context.Context, id uuid.UUID, req model.UpdateUserSnapshotRequest) (*model.UserSnapshotResponse, error)
	Delete(c context.Context, id uuid.UUID) error
}

type UserSnapshotService struct {
	users UserSnapshotRepository
	log   *slog.Logger
}

func NewUserSnapshotService(users UserSnapshotRepository, log *slog.Logger) UserSnapshotService {
	return UserSnapshotService{users, log}
}

func (s UserSnapshotService) Create(c context.Context, req model.CreateUserSnapshotRequest) error {
	if _, err := s.users.Create(c, req); err != nil {
		return err
	}

	return nil
}

// Validate ensure user snapshot exists
func (s UserSnapshotService) Validate(c context.Context, user core.AuthenticatedUser) error {
	_, err := s.users.GetByID(c, user.ID)

	if err != nil {
		if core.IsNotFound(err) {
			return s.Create(c, model.CreateUserSnapshotRequest{
				ID:     user.ID,
				Name:   user.Name,
				Email:  user.Email,
				Avatar: user.Avatar,
			})
		}
		return err
	}

	return nil
}
