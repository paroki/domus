package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
)

type UserSnapshotService interface {
	Validate(c context.Context, user core.AuthenticatedUser) error
}

func Snapshot(svc UserSnapshotService) fiber.Handler {
	return func(c fiber.Ctx) error {
		user := core.UserFromContext(c)

		if err := svc.Validate(c, user); err != nil {
			return err
		}

		return c.Next()
	}
}
