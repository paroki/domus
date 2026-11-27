package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/liturgi/service"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// Handler exposes HTTP endpoints for parish liturgical ministries and Mass schedules.
type Handler struct {
	service *service.Service
}

// New constructs a liturgi Handler backed by svc.
func New(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Register mounts /liturgi routes onto api (/api).
func (h *Handler) Register(
	api fiber.Router,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireTenant fiber.Handler,
	requireMinistriesRead fiber.Handler,
	requireSchedulesRead fiber.Handler,
) {
	liturgi := api.Group("/liturgi", authMW, requireAuth, requireTenant)

	liturgi.Get("/ministries", requireMinistriesRead, h.ListMinistries)
	liturgi.Patch("/ministries/:roleSlug", h.UpdateMinistry)

	liturgi.Post("/schedules/search", requireSchedulesRead, h.SearchSchedules)
	liturgi.Get("/schedules/:eventId", requireSchedulesRead, h.GetSchedule)
	liturgi.Post("/schedules", h.CreateSchedule)
	liturgi.Patch("/schedules/:eventId", h.UpdateSchedule)
	liturgi.Delete("/schedules/:eventId", h.DeleteSchedule)
}

// ListMinistries returns all 6 liturgical ministry configurations for the active parish.
//
//	@Summary		List liturgical ministries
//	@Description	Returns the 6 canonical parish liturgical ministries (lektor, mazmur, pembagi-komuni, misdinar, organis, kor) in fixed order with org/leader metadata and canManage flags.
//	@ID				listLiturgiMinistries
//	@Tags			Liturgi
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[[]domain.LiturgiMinistryResponse]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Router			/liturgi/ministries [get]
func (h *Handler) ListMinistries(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	items, err := h.service.ListMinistries(c.Context(), p)
	if err != nil {
		return err
	}
	return httpx.OK(c, items)
}

// UpdateMinistry updates the organization and/or leader user for a liturgical ministry role.
//
//	@Summary		Update liturgical ministry
//	@Description	Configures the assigned organization and/or leader user for one of the 6 parish liturgical ministry roles (requires liturgi:admin or ref:paroki-admin).
//	@ID				updateLiturgiMinistry
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			roleSlug	path		string								true	"Ministry role slug (lektor, mazmur, pembagi-komuni, misdinar, organis, kor)"
//	@Param			body		body		domain.UpdateLiturgiMinistryRequest	true	"Ministry update payload"
//	@Success		200			{object}	httpx.Envelope[domain.LiturgiMinistryResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/liturgi/ministries/{roleSlug} [patch]
func (h *Handler) UpdateMinistry(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	roleSlug := c.Params("roleSlug")
	var req domain.UpdateLiturgiMinistryRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateMinistry(c.Context(), p, roleSlug, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// SearchSchedules searches parish Mass schedules in the active parish.
//
//	@Summary		Search Mass schedules
//	@Description	Searches parish-level Mass schedules (events with org_id=NULL and jenis=misa) with petugas and lagu lists.
//	@ID				searchLiturgiSchedules
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchLiturgiSchedulesRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.LiturgiScheduleResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/search [post]
func (h *Handler) SearchSchedules(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchLiturgiSchedulesRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchSchedules(c.Context(), p, req, cursorT, cursorID, limit)
	if err != nil {
		return err
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = httpx.EncodeCursor(last.CreatedAt, last.ID.String())
	}
	return httpx.List(c, items, httpx.Meta{
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	})
}

// GetSchedule returns a single parish Mass schedule by eventId.
//
//	@Summary		Get Mass schedule
//	@Description	Returns a single parish-level Mass schedule with petugas, lagu, RSVP summary, and ministry permissions.
//	@ID				getLiturgiSchedule
//	@Tags			Liturgi
//	@Produce		json
//	@Param			eventId	path		string	true	"Mass schedule event UUID"
//	@Success		200		{object}	httpx.Envelope[domain.LiturgiScheduleResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId} [get]
func (h *Handler) GetSchedule(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	res, err := h.service.GetSchedule(c.Context(), p, eventID)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateSchedule creates a new parish-level Mass schedule.
//
//	@Summary		Create Mass schedule
//	@Description	Creates a parish-level Mass schedule in events (org_id=NULL, jenis=misa, visibility=public, rsvp_enabled=true default).
//	@ID				createLiturgiSchedule
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateLiturgiScheduleRequest	true	"Mass schedule payload"
//	@Success		201		{object}	httpx.Envelope[domain.LiturgiScheduleResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules [post]
func (h *Handler) CreateSchedule(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateLiturgiScheduleRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateSchedule(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateSchedule updates an existing parish-level Mass schedule.
//
//	@Summary		Update Mass schedule
//	@Description	Updates an existing parish-level Mass schedule (requires liturgi:admin or ref:paroki-admin).
//	@ID				updateLiturgiSchedule
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId	path		string								true	"Mass schedule event UUID"
//	@Param			body	body		domain.UpdateLiturgiScheduleRequest	true	"Mass schedule update payload"
//	@Success		200		{object}	httpx.Envelope[domain.LiturgiScheduleResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId} [patch]
func (h *Handler) UpdateSchedule(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	var req domain.UpdateLiturgiScheduleRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateSchedule(c.Context(), p, eventID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteSchedule soft-deletes a parish-level Mass schedule.
//
//	@Summary		Delete Mass schedule
//	@Description	Soft-deletes a parish-level Mass schedule (requires liturgi:admin or ref:paroki-admin).
//	@ID				deleteLiturgiSchedule
//	@Tags			Liturgi
//	@Produce		json
//	@Param			eventId	path	string	true	"Mass schedule event UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId} [delete]
func (h *Handler) DeleteSchedule(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	if err := h.service.DeleteSchedule(c.Context(), p, eventID); err != nil {
		return err
	}
	return httpx.NoContent(c)
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
