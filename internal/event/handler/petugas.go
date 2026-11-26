package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// CreateEventPetugas assigns an officer/minister to an Event.
//
//	@Summary		Create event petugas
//	@Description	Assigns an officer/minister to an event (requires event manager: org:admin or ref:paroki-admin).
//	@ID				createEventPetugas
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Event UUID"
//	@Param			body	body		domain.CreateEventPetugasRequest	true	"Officer assignment payload"
//	@Success		201		{object}	httpx.Envelope[domain.EventPetugasResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/events/{id}/petugas [post]
func (h *Handler) CreateEventPetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	var req domain.CreateEventPetugasRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateEventPetugas(c.Context(), p, eventID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateEventPetugas updates an officer assignment or confirms/declines officer duty.
//
//	@Summary		Update event petugas
//	@Description	Updates an officer assignment (event manager) or confirms/declines duty (assigned user or org:admin).
//	@ID				updateEventPetugas
//	@Tags			Event
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string								true	"Event UUID"
//	@Param			petugasId	path		string								true	"Petugas UUID"
//	@Param			body		body		domain.UpdateEventPetugasRequest	true	"Officer update payload"
//	@Success		200			{object}	httpx.Envelope[domain.EventPetugasResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/events/{id}/petugas/{petugasId} [patch]
func (h *Handler) UpdateEventPetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	petugasID, err := uuid.Parse(c.Params("petugasId"))
	if err != nil {
		return httpx.BadRequest("Invalid petugas ID.")
	}
	var req domain.UpdateEventPetugasRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateEventPetugas(c.Context(), p, eventID, petugasID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteEventPetugas removes an officer assignment from an Event.
//
//	@Summary		Delete event petugas
//	@Description	Removes an officer assignment from an event (requires event manager).
//	@ID				deleteEventPetugas
//	@Tags			Event
//	@Produce		json
//	@Param			id			path	string	true	"Event UUID"
//	@Param			petugasId	path	string	true	"Petugas UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/events/{id}/petugas/{petugasId} [delete]
func (h *Handler) DeleteEventPetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid event ID.")
	}
	petugasID, err := uuid.Parse(c.Params("petugasId"))
	if err != nil {
		return httpx.BadRequest("Invalid petugas ID.")
	}
	if err := h.service.DeleteEventPetugas(c.Context(), p, eventID, petugasID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
