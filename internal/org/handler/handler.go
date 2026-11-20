package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/org/service"
	"github.com/paroki/domus/internal/platform/httpx"
)

// Handler exposes HTTP endpoints for organizations and organization memberships.
type Handler struct {
	service *service.Service
}

// New constructs an org Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /orgs routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireOrgRead fiber.Handler,
) {
	orgs := api.Group("/orgs", authMW, requireAuth, requireTenant)
	orgs.Post("/search", requireOrgRead, h.SearchOrganizations)
	orgs.Get("/:id", requireOrgRead, h.GetOrganization)
	orgs.Post("", h.CreateOrganization)
	orgs.Patch("/:id", h.UpdateOrganization)
	orgs.Delete("/:id", h.DeleteOrganization)

	orgs.Post("/:id/members/search", h.SearchOrgMembers)
	orgs.Post("/:id/members", h.AddOrgMember)
	orgs.Patch("/:id/members/:userId", h.UpdateOrgMember)
	orgs.Delete("/:id/members/:userId", h.RemoveOrgMember)
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
