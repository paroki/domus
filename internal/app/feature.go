package app

import "github.com/gofiber/fiber/v3"

// Feature is satisfied by every internal/<feature>/<feature>.go facade.
// Features never import this package; Go structural typing checks them at the
// mountFeatures call site.
type Feature interface {
	// Register mounts the feature's route groups under /api (e.g. api.Group("/users")).
	Register(api fiber.Router)
}

func mountFeatures(api fiber.Router, features ...Feature) {
	for _, f := range features {
		f.Register(api)
	}
}
