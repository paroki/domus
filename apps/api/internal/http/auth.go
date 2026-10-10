package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/shared/authz"
)

func RequirePermission(obj authz.Resource, act authz.Action) fiber.Handler {
	return func(c fiber.Ctx) error {
		user := core.UserFromContext(c)
		if authz.Can(user, obj, act) {
			return c.Next()
		}
		return core.ErrForbidden
	}
}
