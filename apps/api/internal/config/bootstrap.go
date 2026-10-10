package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/api/ent"
)

type State struct {
	Ent      *ent.Client
	Config   Config
	FiberApp *fiber.App
	Log      *slog.Logger
}

func Bootstrap(state State) {
	initCore(state)
}
