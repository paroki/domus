package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/paroki/domus/api/internal/config"
)

// NewCORS constructs the CORS middleware dynamically based on application configuration.
func NewCORS(cfg *config.Config) fiber.Handler {
	var allowedOrigins []string
	if cfg.Api.Cors != "" {
		for _, origin := range strings.Split(cfg.Api.Cors, ",") {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	if len(allowedOrigins) == 0 && cfg.Env == "development" {
		allowedOrigins = []string{"*"}
	}

	if len(allowedOrigins) == 0 {
		// No CORS headers will be appended, effectively blocking cross-origin requests.
		return func(c fiber.Ctx) error {
			return c.Next()
		}
	}

	hasWildcard := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			hasWildcard = true
			break
		}
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: !hasWildcard,
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",
			"Accept",
			"Origin",
			"X-Requested-With",
			"X-Request-ID",
		},
	})
}
