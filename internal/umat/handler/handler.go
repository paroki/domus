package handler

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/service"
)

// Handler exposes HTTP endpoints for the umat feature.
type Handler struct {
	service *service.Service
}

// New constructs a umat Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /umat routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireKeluargaRead fiber.Handler,
	requireKeluargaWrite fiber.Handler,
	requireUmatRead fiber.Handler,
	requireUmatWrite fiber.Handler,
) {
	u := api.Group("/umat", authMW, requireAuth, requireTenant)

	// Static /keluarga routes MUST be registered before parameterized /:id routes.
	keluarga := u.Group("/keluarga")
	keluarga.Post("/search", requireKeluargaRead, h.SearchKeluarga)
	keluarga.Get("/:id", requireKeluargaRead, h.GetKeluarga)
	keluarga.Post("", requireKeluargaWrite, h.CreateKeluarga)
	keluarga.Patch("/:id", requireKeluargaWrite, h.UpdateKeluarga)
	keluarga.Delete("/:id", requireKeluargaWrite, h.DeleteKeluarga)

	u.Post("/search", requireUmatRead, h.SearchUmat)
	u.Post("/lookup", h.LookupUmat)
	u.Get("/:id", requireUmatRead, h.GetUmat)
	u.Post("", requireUmatWrite, h.CreateUmat)
	u.Patch("/:id", requireUmatWrite, h.UpdateUmat)
	u.Delete("/:id", requireUmatWrite, h.DeleteUmat)

	u.Post("/:id/sakramen", requireUmatWrite, h.UpsertSakramen)
	u.Patch("/:id/sakramen/:sakramenId", requireUmatWrite, h.UpdateSakramen)
	u.Delete("/:id/sakramen/:sakramenId", requireUmatWrite, h.DeleteSakramen)
}

// SearchKeluarga searches Keluarga records in the active paroki.
//
//	@Summary		Search keluarga
//	@Description	Searches household (Kartu Keluarga) records in the active paroki with masked nomorKk and cursor pagination.
//	@ID				searchKeluarga
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchKeluargaRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.KeluargaListItemResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/umat/keluarga/search [post]
func (h *Handler) SearchKeluarga(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchKeluargaRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	limit := normalizeLimit(req.Limit)
	items, nextCursor, hasMore, err := h.service.SearchKeluarga(c.Context(), p.ParokiID, req, limit)
	if err != nil {
		return err
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// GetKeluarga fetches a single Keluarga in the active paroki.
//
//	@Summary		Get keluarga
//	@Description	Returns a single non-deleted keluarga with decrypted nomorKk and masked anggota list.
//	@ID				getKeluarga
//	@Tags			Umat
//	@Produce		json
//	@Param			id	path		string	true	"Keluarga UUID"
//	@Success		200	{object}	httpx.Envelope[domain.KeluargaDetailResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/umat/keluarga/{id} [get]
func (h *Handler) GetKeluarga(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid keluarga ID.")
	}
	res, err := h.service.GetKeluarga(c.Context(), p.ParokiID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateKeluarga creates a new Keluarga in the active paroki.
//
//	@Summary		Create keluarga
//	@Description	Creates a household (Kartu Keluarga) record in the active paroki.
//	@ID				createKeluarga
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateKeluargaRequest	true	"Keluarga payload"
//	@Success		201		{object}	httpx.Envelope[domain.KeluargaDetailResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/keluarga [post]
func (h *Handler) CreateKeluarga(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateKeluargaRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateKeluarga(c.Context(), p.ParokiID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateKeluarga partially updates a Keluarga in the active paroki.
//
//	@Summary		Update keluarga
//	@Description	Partially updates a household (Kartu Keluarga) record in the active paroki.
//	@ID				updateKeluarga
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Keluarga UUID"
//	@Param			body	body		domain.UpdateKeluargaRequest	true	"Keluarga update payload"
//	@Success		200		{object}	httpx.Envelope[domain.KeluargaDetailResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/keluarga/{id} [patch]
func (h *Handler) UpdateKeluarga(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid keluarga ID.")
	}
	var req domain.UpdateKeluargaRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateKeluarga(c.Context(), p.ParokiID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteKeluarga soft-deletes a Keluarga in the active paroki.
//
//	@Summary		Delete keluarga
//	@Description	Soft-deletes a household (Kartu Keluarga) if no active umat references it.
//	@ID				deleteKeluarga
//	@Tags			Umat
//	@Param			id	path	string	true	"Keluarga UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/umat/keluarga/{id} [delete]
func (h *Handler) DeleteKeluarga(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid keluarga ID.")
	}
	if err := h.service.DeleteKeluarga(c.Context(), p.ParokiID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// SearchUmat searches Umat records in the active paroki.
//
//	@Summary		Search umat
//	@Description	Searches individual parishioner profiles in the active paroki with masked nik and cursor pagination.
//	@ID				searchUmat
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchUmatRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.UmatListItemResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/umat/search [post]
func (h *Handler) SearchUmat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchUmatRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	limit := normalizeLimit(req.Limit)
	items, nextCursor, hasMore, err := h.service.SearchUmat(c.Context(), p.ParokiID, req, limit)
	if err != nil {
		return err
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// GetUmat fetches a single Umat in the active paroki.
//
//	@Summary		Get umat
//	@Description	Returns a single non-deleted umat with decrypted nik, keluarga summary, and sakramen list.
//	@ID				getUmat
//	@Tags			Umat
//	@Produce		json
//	@Param			id	path		string	true	"Umat UUID"
//	@Success		200	{object}	httpx.Envelope[domain.UmatDetailResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/umat/{id} [get]
func (h *Handler) GetUmat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid umat ID.")
	}
	res, err := h.service.GetUmat(c.Context(), p.ParokiID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateUmat creates a new Umat in the active paroki.
//
//	@Summary		Create umat
//	@Description	Creates an individual parishioner profile in the active paroki.
//	@ID				createUmat
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateUmatRequest	true	"Umat payload"
//	@Success		201		{object}	httpx.Envelope[domain.UmatDetailResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat [post]
func (h *Handler) CreateUmat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateUmatRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateUmat(c.Context(), p.ParokiID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateUmat partially updates an Umat in the active paroki.
//
//	@Summary		Update umat
//	@Description	Partially updates an individual parishioner profile in the active paroki.
//	@ID				updateUmat
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Umat UUID"
//	@Param			body	body		domain.UpdateUmatRequest	true	"Umat update payload"
//	@Success		200		{object}	httpx.Envelope[domain.UmatDetailResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/{id} [patch]
func (h *Handler) UpdateUmat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid umat ID.")
	}
	var req domain.UpdateUmatRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateUmat(c.Context(), p.ParokiID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteUmat soft-deletes an Umat in the active paroki.
//
//	@Summary		Delete umat
//	@Description	Soft-deletes an individual parishioner profile in the active paroki.
//	@ID				deleteUmat
//	@Tags			Umat
//	@Param			id	path	string	true	"Umat UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/umat/{id} [delete]
func (h *Handler) DeleteUmat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid umat ID.")
	}
	if err := h.service.DeleteUmat(c.Context(), p.ParokiID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// LookupUmat returns PII-free parishioner summaries for organization position assignment.
//
//	@Summary		Lookup umat (PII-free)
//	@Description	Searches active, non-deceased parishioners in the active paroki without PII fields. Allowed for umat:admin, ref:paroki-admin, or active org:admin.
//	@ID				lookupUmat
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.LookupUmatRequest	false	"Lookup parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.UmatLookupItemResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/umat/lookup [post]
func (h *Handler) LookupUmat(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.LookupUmatRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	items, err := h.service.LookupUmat(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, items)
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

func normalizeLimit(raw int) int {
	if raw <= 0 {
		return httpx.DefaultLimit
	}
	if raw > httpx.MaxLimit {
		return httpx.MaxLimit
	}
	return raw
}
