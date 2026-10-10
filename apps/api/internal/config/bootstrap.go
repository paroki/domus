package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/modules/auth"
)

type State struct {
	Ent      *ent.Client
	Config   Config
	FiberApp *fiber.App
	Log      *slog.Logger
}

func initAuth(state State) {
	auth := auth.New(auth.Config{
		RestIssuer:   state.Config.AuthUrl,
		RestAudience: state.Config.AuthUrl,
		EntCli:       state.Ent,
		Log:          state.Log,
	})
	auth.InitMiddleware(state.FiberApp)
	auth.InitRoutes(state.FiberApp)
}

func Bootstrap(state State) {
	initAuth(state)
}
