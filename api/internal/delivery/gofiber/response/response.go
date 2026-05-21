package response

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// Envelope is the top-level wrapper for all API responses.
type Envelope[T any] struct {
	Success bool       `json:"success"`
	Data    T          `json:"data"`
	Meta    *Meta      `json:"meta"`
	Error   *ErrorBody `json:"error"`
}

// Meta carries auxiliary metadata included in every response.
type Meta struct {
	RequestID  string      `json:"request_id"`
	Timestamp  time.Time   `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination carries paging information for collection responses.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// ErrorBody carries structured error information.
type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details"`
}

// FieldError represents a single field-level validation failure.
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

func newMeta(c fiber.Ctx) *Meta {
	return &Meta{
		RequestID: requestid.FromContext(c),
		Timestamp: time.Now().UTC(),
	}
}

// OK writes a 200 success response.
func OK[T any](c fiber.Ctx, data T) error {
	return c.Status(fiber.StatusOK).JSON(Envelope[T]{
		Success: true,
		Data:    data,
		Meta:    newMeta(c),
		Error:   nil,
	})
}

// Created writes a 201 success response.
func Created[T any](c fiber.Ctx, data T) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope[T]{
		Success: true,
		Data:    data,
		Meta:    newMeta(c),
		Error:   nil,
	})
}

// NoContent writes a 204 response with no body.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// PaginatedOK writes a 200 paginated collection response.
func PaginatedOK[T any](c fiber.Ctx, data []T, p Pagination) error {
	m := newMeta(c)
	m.Pagination = &p
	return c.Status(fiber.StatusOK).JSON(Envelope[[]T]{
		Success: true,
		Data:    data,
		Meta:    m,
		Error:   nil,
	})
}

// Fail writes a structured error response.
func Fail(c fiber.Ctx, status int, code, message string, details []FieldError) error {
	return c.Status(status).JSON(Envelope[any]{
		Success: false,
		Data:    nil,
		Meta:    newMeta(c),
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
