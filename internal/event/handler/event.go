package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// SearchEvents searches visible events in the active parish.
//
//	@Summary		Search events
//	@Description	Searches visible events (public parish events and private events for caller's organizations) with filters and cursor pagination.
//	@ID				searchEvents
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.SearchEventsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.EventResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Router			/events/search [post]
func (h *Handler) SearchEvents(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.SearchEventsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchEvents(c.Context(), p, req, cursorT, cursorID, limit)
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

// GetEvent fetches a single Event in the active parish by ID.
//
//	@Summary		Get event
//	@Description	Returns a single non-deleted event in the active parish if visible to the caller.
//	@ID				getEvent
//	@Tags			Event
//	@Produce		json
//	@Param			id	path		string	true	"Event UUID"
//	@Success		200	{object}	httpx.Envelope[domain.EventResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/events/{id} [get]
func (h *Handler) GetEvent(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	res, err := h.service.GetEvent(c.Context(), p, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// CreateEvent creates a new Event owned by an organization in the active parish.
//
//	@Summary		Create event
//	@Description	Creates an organization-owned event in the active parish (requires org:admin of the owning organization).
//	@ID				createEvent
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateEventRequest	true	"Event creation payload"
//	@Success		201		{object}	httpx.Envelope[domain.EventResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/events [post]
func (h *Handler) CreateEvent(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	var req domain.CreateEventRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateEvent(c.Context(), p, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateEvent updates an existing Event in the active parish.
//
//	@Summary		Update event
//	@Description	Updates or cancels an event in the active parish (requires org:admin of the owning organization).
//	@ID				updateEvent
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Event UUID"
//	@Param			body	body		domain.UpdateEventRequest	true	"Event update payload"
//	@Success		200		{object}	httpx.Envelope[domain.EventResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/events/{id} [patch]
func (h *Handler) UpdateEvent(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	var req domain.UpdateEventRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateEvent(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteEvent soft-deletes an Event in the active parish.
//
//	@Summary		Delete event
//	@Description	Soft-deletes an event in the active parish (requires org:admin of the owning organization).
//	@ID				deleteEvent
//	@Tags			Event
//	@Produce		json
//	@Param			id	path	string	true	"Event UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/events/{id} [delete]
func (h *Handler) DeleteEvent(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	if err := h.service.DeleteEvent(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
