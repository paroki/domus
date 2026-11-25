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

// Handler exposes HTTP endpoints for reference master data (keuskupan, paroki, members).
type Handler struct {
	service *service.Service
}

// New constructs a ref Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /ref/keuskupan, /ref/paroki, /ref/wilayah, /ref/lingkungan, /ref/lookups, and /ref/periods routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireKeuskupanRead fiber.Handler,
	requireKeuskupanWrite fiber.Handler,
	requireParokiRead fiber.Handler,
	requireParokiWrite fiber.Handler,
	requireMasterDataRead fiber.Handler,
	requireMasterDataWrite fiber.Handler,
) {
	keuskupan := api.Group("/ref/keuskupan", authMW, requireAuth)
	keuskupan.Post("/search", requireKeuskupanRead, h.SearchKeuskupan)
	keuskupan.Get("/:id", requireKeuskupanRead, h.GetKeuskupan)
	keuskupan.Post("", requireKeuskupanWrite, h.CreateKeuskupan)
	keuskupan.Patch("/:id", requireKeuskupanWrite, h.UpdateKeuskupan)
	keuskupan.Delete("/:id", requireKeuskupanWrite, h.DeleteKeuskupan)

	paroki := api.Group("/ref/paroki", authMW, requireAuth)
	paroki.Post("/search", requireParokiRead, h.SearchParoki)
	// Static /current routes MUST be registered before parameterized /:id routes.
	paroki.Get("/current", requireParokiRead, requireTenant, h.GetCurrentParoki)
	paroki.Patch("/current", requireTenant, h.UpdateCurrentParokiProfile)
	paroki.Get("/:id", requireParokiRead, h.GetParoki)
	paroki.Post("", requireParokiWrite, h.CreateParoki)
	paroki.Patch("/:id", h.UpdateParoki)
	paroki.Delete("/:id", requireParokiWrite, h.DeleteParoki)

	paroki.Post("/:id/members/search", h.SearchParokiMembers)
	paroki.Post("/:id/members", h.AddParokiMember)
	paroki.Patch("/:id/members/:userId", h.UpdateParokiMember)
	paroki.Delete("/:id/members/:userId", h.RemoveParokiMember)

	wilayah := api.Group("/ref/wilayah", authMW, requireAuth, requireTenant)
	wilayah.Post("/search", requireMasterDataRead, h.SearchWilayah)
	wilayah.Get("/:id", requireMasterDataRead, h.GetWilayah)
	wilayah.Post("", requireMasterDataWrite, h.CreateWilayah)
	wilayah.Patch("/:id", requireMasterDataWrite, h.UpdateWilayah)
	wilayah.Delete("/:id", requireMasterDataWrite, h.DeleteWilayah)

	lingkungan := api.Group("/ref/lingkungan", authMW, requireAuth, requireTenant)
	lingkungan.Post("/search", requireMasterDataRead, h.SearchLingkungan)
	lingkungan.Get("/:id", requireMasterDataRead, h.GetLingkungan)
	lingkungan.Post("", requireMasterDataWrite, h.CreateLingkungan)
	lingkungan.Patch("/:id", requireMasterDataWrite, h.UpdateLingkungan)
	lingkungan.Delete("/:id", requireMasterDataWrite, h.DeleteLingkungan)

	lookups := api.Group("/ref/lookups", authMW, requireAuth)
	lookups.Post("/search", requireMasterDataRead, h.SearchLookups)
	lookups.Get("/:id", requireMasterDataRead, h.GetLookup)
	lookups.Post("", h.CreateLookup)
	lookups.Patch("/:id", h.UpdateLookup)
	lookups.Delete("/:id", h.DeleteLookup)

	periods := api.Group("/ref/periods", authMW, requireAuth, requireTenant)
	periods.Post("/search", requireMasterDataRead, h.SearchPeriods)
	periods.Get("/:id", requireMasterDataRead, h.GetPeriod)
	periods.Post("", requireMasterDataWrite, h.CreatePeriod)
	periods.Patch("/:id", requireMasterDataWrite, h.UpdatePeriod)
	periods.Delete("/:id", requireMasterDataWrite, h.DeletePeriod)
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
