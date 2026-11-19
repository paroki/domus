package httpx

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
)

// Envelope is the standard Domus HTTP JSON response shape { data, meta, error }.
type Envelope[T any] struct {
	Data  T             `json:"data,omitempty"`
	Meta  *Meta         `json:"meta,omitempty"`
	Error *ErrorPayload `json:"error,omitempty"`
}

// MarshalJSON serializes Envelope[T] with mutually exclusive error and data/meta keys.
func (e Envelope[T]) MarshalJSON() ([]byte, error) {
	if e.Error != nil {
		return json.Marshal(struct {
			Error *ErrorPayload `json:"error"`
		}{
			Error: e.Error,
		})
	}
	return json.Marshal(struct {
		Data T     `json:"data"`
		Meta *Meta `json:"meta,omitempty"`
	}{
		Data: e.Data,
		Meta: e.Meta,
	})
}

// Meta holds cursor pagination metadata for list responses.
type Meta struct {
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	Limit      int    `json:"limit"`
	Total      *int   `json:"total,omitempty"`
}

// ErrorPayload represents the structured error object in Envelope.Error.
type ErrorPayload struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"requestId,omitempty"`
	Details   []FieldError `json:"details,omitempty"`
}

// FieldError represents a single field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 JSON envelope with the given data payload.
func OK[T any](c fiber.Ctx, data T) error {
	return c.Status(fiber.StatusOK).JSON(Envelope[T]{Data: data})
}

// Created writes a 201 JSON envelope with the given data payload.
func Created[T any](c fiber.Ctx, data T) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope[T]{Data: data})
}

// List writes a 200 JSON envelope with items (never null) and pagination meta.
func List[T any](c fiber.Ctx, items []T, meta Meta) error {
	if items == nil {
		items = []T{}
	}
	return c.Status(fiber.StatusOK).JSON(Envelope[[]T]{
		Data: items,
		Meta: &meta,
	})
}

// NoContent writes a 204 No Content response with an empty body.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
