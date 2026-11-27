package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// CreateLagu adds a liturgical song entry to a parish Mass schedule.
//
//	@Summary		Add liturgy song
//	@Description	Adds a liturgical song entry (optionally linked to cantusId with auto-filled metadata) to a Mass schedule.
//	@ID				createLiturgiLagu
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId	path		string							true	"Mass schedule event UUID"
//	@Param			body	body		domain.CreateLiturgiLaguRequest	true	"Song entry payload"
//	@Success		201		{object}	httpx.Envelope[domain.LiturgiLaguResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/lagu [post]
func (h *Handler) CreateLagu(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	var req domain.CreateLiturgiLaguRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateLagu(c.Context(), p, eventID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdateLagu updates a liturgical song entry on a parish Mass schedule.
//
//	@Summary		Update liturgy song
//	@Description	Updates a liturgical song entry on a Mass schedule.
//	@ID				updateLiturgiLagu
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId	path		string							true	"Mass schedule event UUID"
//	@Param			laguId	path		string							true	"Song entry UUID"
//	@Param			body	body		domain.UpdateLiturgiLaguRequest	true	"Song entry update payload"
//	@Success		200		{object}	httpx.Envelope[domain.LiturgiLaguResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/lagu/{laguId} [patch]
func (h *Handler) UpdateLagu(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	laguID, err := uuid.Parse(c.Params("laguId"))
	if err != nil {
		return httpx.BadRequest("Invalid lagu ID.")
	}
	var req domain.UpdateLiturgiLaguRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateLagu(c.Context(), p, eventID, laguID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteLagu removes a liturgical song entry from a parish Mass schedule.
//
//	@Summary		Delete liturgy song
//	@Description	Removes a liturgical song entry from a Mass schedule.
//	@ID				deleteLiturgiLagu
//	@Tags			Liturgi
//	@Produce		json
//	@Param			eventId	path	string	true	"Mass schedule event UUID"
//	@Param			laguId	path	string	true	"Song entry UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/lagu/{laguId} [delete]
func (h *Handler) DeleteLagu(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	laguID, err := uuid.Parse(c.Params("laguId"))
	if err != nil {
		return httpx.BadRequest("Invalid lagu ID.")
	}
	if err := h.service.DeleteLagu(c.Context(), p, eventID, laguID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
