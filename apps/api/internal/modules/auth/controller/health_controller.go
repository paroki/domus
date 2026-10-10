package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/http"
	_ "github.com/paroki/domus/api/internal/model"
)

type HealthController struct {
}

// Ping godoc
//
//	@Summary		Health check ping
//	@Description	Returns current authenticated user information from context
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.WebResponse[core.AuthenticatedUser]
//	@Failure		401	{object}	model.ErrorResponse
//	@Router			/ping [get]
func (hc HealthController) Ping(c fiber.Ctx) error {
	user := core.UserFromContext(c)
	return http.OK(c, user)
}
