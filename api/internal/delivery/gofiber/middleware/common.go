package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Setup(app *fiber.App) {
	app.Use(recover.New())
	app.Use(cors.New())
}
