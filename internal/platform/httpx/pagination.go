package httpx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	// DefaultLimit is the default page size when limit is omitted or zero.
	DefaultLimit = 20
	// MaxLimit is the maximum allowed page size.
	MaxLimit = 100
)

// PageRequest holds cursor pagination query parameters.
type PageRequest struct {
	Cursor string `query:"cursor"`
	Limit  int    `query:"limit"`
}

// Cursor represents an opaque keyset pagination cursor.
type Cursor struct {
	CreatedAt time.Time `json:"t"`
	ID        string    `json:"id"`
}

// ParsePageRequest binds and validates cursor pagination query parameters.
func ParsePageRequest(c fiber.Ctx) (PageRequest, error) {
	var req PageRequest
	if err := c.Bind().Query(&req); err != nil {
		return PageRequest{}, ValidationError(FieldError{
			Field:   "limit",
			Code:    "INVALID",
			Message: "limit must be a valid integer",
		})
	}
	if req.Limit == 0 {
		req.Limit = DefaultLimit
	}
	if req.Limit < 0 || req.Limit > MaxLimit {
		return PageRequest{}, ValidationError(FieldError{
			Field:   "limit",
			Code:    "RANGE",
			Message: fmt.Sprintf("limit must be between 1 and %d", MaxLimit),
		})
	}
	return req, nil
}

// EncodeCursor encodes a (createdAt, id) keyset into a URL-safe base64 cursor.
func EncodeCursor(createdAt time.Time, id string) string {
	raw, _ := json.Marshal(Cursor{
		CreatedAt: createdAt.UTC(),
		ID:        id,
	})
	return base64.RawURLEncoding.EncodeToString(raw)
}

// DecodeCursor decodes and validates a URL-safe base64 pagination cursor.
func DecodeCursor(raw string) (Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, BadRequest("Invalid pagination cursor.")
	}
	var cur Cursor
	if err := json.Unmarshal(decoded, &cur); err != nil {
		return Cursor{}, BadRequest("Invalid pagination cursor.")
	}
	if cur.ID == "" || cur.CreatedAt.IsZero() {
		return Cursor{}, BadRequest("Invalid pagination cursor.")
	}
	return cur, nil
}
