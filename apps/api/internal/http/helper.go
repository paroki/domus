package http

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
	"github.com/paroki/domus/api/internal/model"
)

func GetID(c fiber.Ctx) (*uuid.UUID, error) {
	strID := c.Params("id", "undefined")
	if strID == "undefined" {
		return nil, core.ErrInvalidID
	}

	id, err := uuid.Parse(strID)
	if err != nil {
		return nil, core.ErrInvalidID
	}
	return &id, nil
}

func NewMeta(c fiber.Ctx, cursor string) model.Meta {
	return model.Meta{
		RequestID: requestid.FromContext(c),
		Timestamp: time.Now(),
		Cursor:    cursor,
	}
}

func Success[T any](c fiber.Ctx, status int, data T) error {
	return c.Status(status).JSON(model.WebResponse[T]{
		Data: data,
		Meta: NewMeta(c, ""),
	})
}

func WithCursor[T any](c fiber.Ctx, data T, cursor string) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse[T]{
		Data: data,
		Meta: NewMeta(c, cursor),
	})
}

func OK[T any](c fiber.Ctx, data T) error {
	return Success(c, fiber.StatusOK, data)
}

func Created[T any](c fiber.Ctx, data T) error {
	return Success(c, fiber.StatusCreated, data)
}
