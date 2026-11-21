package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// UpsertEventRSVP creates or updates the caller's RSVP for an Event.
//
//	@Summary		Upsert event RSVP
//	@Description	Creates or updates the authenticated organization member's RSVP for an event.
//	@ID				upsertEventRSVP
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Event UUID"
//	@Param			body	body		domain.UpsertEventRSVPRequest	true	"RSVP payload"
//	@Success		200		{object}	httpx.Envelope[domain.EventRSVPResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		409		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/events/{id}/rsvp [post]
func (h *Handler) UpsertEventRSVP(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	var req domain.UpsertEventRSVPRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpsertRSVP(c.Context(), p, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteEventRSVP removes the caller's RSVP for an Event.
//
//	@Summary		Delete event RSVP
//	@Description	Removes the authenticated organization member's RSVP for an event (idempotent).
//	@ID				deleteEventRSVP
//	@Tags			Event
//	@Produce		json
//	@Param			id	path	string	true	"Event UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Failure		409	{object}	httpx.Envelope[any]
//	@Router			/events/{id}/rsvp [delete]
func (h *Handler) DeleteEventRSVP(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	if err := h.service.DeleteRSVP(c.Context(), p, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

// GetEventRSVPSummary returns RSVP attendance summary counts for an Event.
//
//	@Summary		Get event RSVP summary
//	@Description	Returns RSVP attendance counts and unresponded member count for an event (requires owning org membership).
//	@ID				getEventRSVPSummary
//	@Tags			Event
//	@Produce		json
//	@Param			id	path		string	true	"Event UUID"
//	@Success		200	{object}	httpx.Envelope[domain.EventRSVPSummaryResponse]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/events/{id}/rsvps/summary [get]
func (h *Handler) GetEventRSVPSummary(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	res, err := h.service.GetRSVPSummary(c.Context(), p, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// SearchEventRSVPs searches the attendee RSVP roster for an Event.
//
//	@Summary		Search event RSVPs
//	@Description	Searches attendee RSVPs for an event with filters and cursor pagination (requires owning org membership).
//	@ID				searchEventRSVPs
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Event UUID"
//	@Param			body	body		domain.SearchEventRSVPsRequest	false	"Search parameters"
//	@Success		200		{object}	httpx.Envelope[[]domain.EventRSVPResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Router			/events/{id}/rsvps/search [post]
func (h *Handler) SearchEventRSVPs(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	var req domain.SearchEventRSVPsRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	cursorT, cursorID, limit, err := parseCursorAndLimit(req.Cursor, req.Limit)
	if err != nil {
		return err
	}
	items, hasMore, err := h.service.SearchRSVPs(c.Context(), p, id, req, cursorT, cursorID, limit)
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
