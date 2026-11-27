package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/service"
	"github.com/paroki/domus/internal/platform/httpx"
)

// Handler exposes HTTP endpoints for parish and organizational events.
type Handler struct {
	service *service.Service
}

// New constructs an event Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /events routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireEventRead fiber.Handler,
) {
	events := api.Group("/events", authMW, requireAuth, requireTenant)
	events.Post("/search", requireEventRead, h.SearchEvents)
	events.Get("/:id", requireEventRead, h.GetEvent)
	events.Post("", h.CreateEvent)
	events.Patch("/:id", h.UpdateEvent)
	events.Delete("/:id", h.DeleteEvent)

	events.Post("/:id/rsvp", h.UpsertEventRSVP)
	events.Delete("/:id/rsvp", h.DeleteEventRSVP)
	events.Get("/:id/rsvps/summary", h.GetEventRSVPSummary)
	events.Post("/:id/rsvps/search", h.SearchEventRSVPs)
}

func decodeStrictJSON(c fiber.Ctx, dst any) error {
	body := bytes.TrimSpace(c.Body())
	if len(body) > 0 {
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		if err := dec.Decode(dst); err != nil {
			return httpx.BadRequest("Invalid request body.")
		}
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			return httpx.BadRequest("Invalid request body.")
		}
	}
	if v := c.App().Config().StructValidator; v != nil {
		if err := v.Validate(dst); err != nil {
			return err
		}
	}
	return nil
}

func decodeEmptyOrStrictJSON(c fiber.Ctx, dst any) error {
	if len(bytes.TrimSpace(c.Body())) == 0 {
		return nil
	}
	return decodeStrictJSON(c, dst)
}

func parseCursorAndLimit(rawCursor string, rawLimit int) (*time.Time, *uuid.UUID, int, error) {
	limit := rawLimit
	if limit <= 0 {
		limit = httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		limit = httpx.MaxLimit
	}
	if rawCursor == "" {
		return nil, nil, limit, nil
	}
	cur, err := httpx.DecodeCursor(rawCursor)
	if err != nil {
		return nil, nil, 0, err
	}
	id, err := uuid.Parse(cur.ID)
	if err != nil {
		return nil, nil, 0, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	t := cur.CreatedAt.Local()
	return &t, &id, limit, nil
}
