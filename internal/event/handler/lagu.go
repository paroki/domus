package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// CreateEventLagu adds a liturgical song entry to an Event.
//
//	@Summary		Create event lagu
//	@Description	Adds a liturgical song entry to an event (requires event manager: org:admin or ref:paroki-admin).
//	@ID				createEventLagu
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Event UUID"
//	@Param			body	body		domain.CreateEventLaguRequest	true	"Song entry payload"
//	@Success		201		{object}	httpx.Envelope[domain.EventLaguResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/events/{id}/lagu [post]
func (h *Handler) CreateEventLagu(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	var req domain.CreateEventLaguRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateEventLagu(c.Context(), p, eventID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateEventLagu updates a liturgical song entry in an Event.
//
//	@Summary		Update event lagu
//	@Description	Updates a liturgical song entry in an event (requires event manager).
//	@ID				updateEventLagu
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Event UUID"
//	@Param			laguId	path		string							true	"Lagu UUID"
//	@Param			body	body		domain.UpdateEventLaguRequest	true	"Song update payload"
//	@Success		200		{object}	httpx.Envelope[domain.EventLaguResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/events/{id}/lagu/{laguId} [patch]
func (h *Handler) UpdateEventLagu(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	laguID, err := uuid.Parse(c.Params("laguId"))
	if err != nil {
		return httpx.BadRequest("Invalid lagu ID.")
	}
	var req domain.UpdateEventLaguRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateEventLagu(c.Context(), p, eventID, laguID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteEventLagu removes a liturgical song entry from an Event.
//
//	@Summary		Delete event lagu
//	@Description	Removes a liturgical song entry from an event (requires event manager).
//	@ID				deleteEventLagu
//	@Tags			Event
//	@Produce		json
//	@Param			id		path	string	true	"Event UUID"
//	@Param			laguId	path	string	true	"Lagu UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/events/{id}/lagu/{laguId} [delete]
func (h *Handler) DeleteEventLagu(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	laguID, err := uuid.Parse(c.Params("laguId"))
	if err != nil {
		return httpx.BadRequest("Invalid lagu ID.")
	}
	if err := h.service.DeleteEventLagu(c.Context(), p, eventID, laguID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
