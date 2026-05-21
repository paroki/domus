package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func Setup(app *fiber.App) {
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(cors.New())
}
