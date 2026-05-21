package handler

import "github.com/gofiber/fiber/v3"

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SendSuccess(c fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Response{
		Success: false,
		Message: message,
	})
}
