package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/service"
)

// Handler exposes HTTP endpoints for reference master data (dioceses, parishes, members).
type Handler struct {
	service *service.Service
}

// New constructs a ref Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /ref/dioceses and /ref/parishes routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireDioceseRead fiber.Handler,
	requireDioceseWrite fiber.Handler,
	requireParishRead fiber.Handler,
	requireParishWrite fiber.Handler,
) {
	dioceses := api.Group("/ref/dioceses", authMW, requireAuth)
	dioceses.Post("/search", requireDioceseRead, h.SearchDioceses)
	dioceses.Get("/:id", requireDioceseRead, h.GetDiocese)
	dioceses.Post("", requireDioceseWrite, h.CreateDiocese)
	dioceses.Patch("/:id", requireDioceseWrite, h.UpdateDiocese)
	dioceses.Delete("/:id", requireDioceseWrite, h.DeleteDiocese)

	parishes := api.Group("/ref/parishes", authMW, requireAuth)
	parishes.Post("/search", requireParishRead, h.SearchParishes)
	// Static /current routes MUST be registered before parameterized /:id routes.
	parishes.Get("/current", requireParishRead, requireTenant, h.GetCurrentParish)
	parishes.Patch("/current", requireTenant, h.UpdateCurrentParishProfile)
	parishes.Get("/:id", requireParishRead, h.GetParish)
	parishes.Post("", requireParishWrite, h.CreateParish)
	parishes.Patch("/:id", h.UpdateParish)
	parishes.Delete("/:id", requireParishWrite, h.DeleteParish)

	parishes.Post("/:id/members/search", h.SearchParishMembers)
	parishes.Post("/:id/members", h.AddParishMember)
	parishes.Patch("/:id/members/:userId", h.UpdateParishMember)
	parishes.Delete("/:id/members/:userId", h.RemoveParishMember)
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
	t := cur.CreatedAt
	return &t, &id, limit, nil
}
