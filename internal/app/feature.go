package app

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// Feature is satisfied by every internal/<feature>/<feature>.go facade.
// Features never import this package; Go structural typing checks them at the
// mountFeatures call site.
type Feature interface {
	// Prefix is the path owned by the feature under /api, e.g. "/parishioners".
	Prefix() string
	// Register mounts routes relative to Prefix.
	Register(r fiber.Router)
}

// mountFeatures fails fast on duplicate prefixes because Fiber silently lets
// the first registered route win.
func mountFeatures(api fiber.Router, features ...Feature) error {
	seen := make(map[string]struct{}, len(features))
	for _, f := range features {
		p := f.Prefix()
		if _, dup := seen[p]; dup {
			return fmt.Errorf("duplicate feature prefix %q", p)
		}
		seen[p] = struct{}{}
		f.Register(api.Group(p))
	}
	return nil
}
