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

// Register mounts /ref/keuskupan, /ref/kevikepan, /ref/kota, /ref/kecamatan, /ref/kelurahan, /ref/paroki, /ref/wilayah, /ref/lingkungan, and /ref/lookups routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireKeuskupanRead fiber.Handler,
	requireKeuskupanWrite fiber.Handler,
	requireKevikepanRead fiber.Handler,
	requireKevikepanWrite fiber.Handler,
	requireSipilRead fiber.Handler,
	requireSipilWrite fiber.Handler,
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

	kevikepan := api.Group("/ref/kevikepan", authMW, requireAuth)
	kevikepan.Post("/search", requireKevikepanRead, h.SearchKevikepan)
	kevikepan.Get("/:id", requireKevikepanRead, h.GetKevikepan)
	kevikepan.Post("", requireKevikepanWrite, h.CreateKevikepan)
	kevikepan.Patch("/:id", requireKevikepanWrite, h.UpdateKevikepan)
	kevikepan.Delete("/:id", requireKevikepanWrite, h.DeleteKevikepan)

	kota := api.Group("/ref/kota", authMW, requireAuth)
	kota.Post("/search", requireSipilRead, h.SearchKota)
	kota.Get("/:id", requireSipilRead, h.GetKota)
	kota.Post("", requireSipilWrite, h.CreateKota)
	kota.Patch("/:id", requireSipilWrite, h.UpdateKota)
	kota.Delete("/:id", requireSipilWrite, h.DeleteKota)

	kecamatan := api.Group("/ref/kecamatan", authMW, requireAuth)
	kecamatan.Post("/search", requireSipilRead, h.SearchKecamatan)
	kecamatan.Get("/:id", requireSipilRead, h.GetKecamatan)
	kecamatan.Post("", requireSipilWrite, h.CreateKecamatan)
	kecamatan.Patch("/:id", requireSipilWrite, h.UpdateKecamatan)
	kecamatan.Delete("/:id", requireSipilWrite, h.DeleteKecamatan)

	kelurahan := api.Group("/ref/kelurahan", authMW, requireAuth)
	kelurahan.Post("/search", requireSipilRead, h.SearchKelurahan)
	kelurahan.Get("/:id", requireSipilRead, h.GetKelurahan)
	kelurahan.Post("", requireSipilWrite, h.CreateKelurahan)
	kelurahan.Patch("/:id", requireSipilWrite, h.UpdateKelurahan)
	kelurahan.Delete("/:id", requireSipilWrite, h.DeleteKelurahan)

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
