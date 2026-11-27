package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// CreatePetugas assigns an officer/minister to a parish Mass schedule.
//
//	@Summary		Assign liturgy officer
//	@Description	Creates a liturgical officer assignment for a Mass schedule (requires liturgi:admin, ref:paroki-admin, or ministry leader/org:admin for the role).
//	@ID				createLiturgiPetugas
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId	path		string								true	"Mass schedule event UUID"
//	@Param			body	body		domain.CreateLiturgiPetugasRequest	true	"Officer assignment payload"
//	@Success		201		{object}	httpx.Envelope[domain.LiturgiPetugasResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/petugas [post]
func (h *Handler) CreatePetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	var req domain.CreateLiturgiPetugasRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreatePetugas(c.Context(), p, eventID, req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// UpdatePetugas updates an existing officer assignment on a parish Mass schedule.
//
//	@Summary		Update liturgy officer
//	@Description	Updates an officer assignment on a Mass schedule (requires permission to manage both existing and target roles).
//	@ID				updateLiturgiPetugas
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId		path		string								true	"Mass schedule event UUID"
//	@Param			petugasId	path		string								true	"Officer assignment UUID"
//	@Param			body		body		domain.UpdateLiturgiPetugasRequest	true	"Officer update payload"
//	@Success		200			{object}	httpx.Envelope[domain.LiturgiPetugasResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/petugas/{petugasId} [patch]
func (h *Handler) UpdatePetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	petugasID, err := uuid.Parse(c.Params("petugasId"))
	if err != nil {
		return httpx.BadRequest("Invalid petugas ID.")
	}
	var req domain.UpdateLiturgiPetugasRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdatePetugas(c.Context(), p, eventID, petugasID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// RespondPetugas records an officer confirmation or decline with an optional substitute proposal.
//
//	@Summary		Respond to liturgy assignment
//	@Description	Records officer confirmation or decline (optionally proposing a substitute) for an assigned officer, assigned org admin, or ministry leader.
//	@ID				respondLiturgiPetugas
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId		path		string								true	"Mass schedule event UUID"
//	@Param			petugasId	path		string								true	"Officer assignment UUID"
//	@Param			body		body		domain.RespondLiturgiPetugasRequest	true	"Officer response payload"
//	@Success		200			{object}	httpx.Envelope[domain.LiturgiPetugasResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/petugas/{petugasId}/respond [patch]
func (h *Handler) RespondPetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	petugasID, err := uuid.Parse(c.Params("petugasId"))
	if err != nil {
		return httpx.BadRequest("Invalid petugas ID.")
	}
	var req domain.RespondLiturgiPetugasRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.RespondPetugas(c.Context(), p, eventID, petugasID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// ApproveSubstitute atomically promotes a proposed substitute onto the active assignment.
//
//	@Summary		Approve proposed substitute
//	@Description	Atomically promotes the proposed substitute onto the active officer assignment (requires ministry management permission).
//	@ID				approveLiturgiSubstitute
//	@Tags			Liturgi
//	@Accept			json
//	@Produce		json
//	@Param			eventId		path		string									true	"Mass schedule event UUID"
//	@Param			petugasId	path		string									true	"Officer assignment UUID"
//	@Param			body		body		domain.ApproveLiturgiSubstituteRequest	false	"Substitute approval options"
//	@Success		200			{object}	httpx.Envelope[domain.LiturgiPetugasResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/petugas/{petugasId}/approve-substitute [post]
func (h *Handler) ApproveSubstitute(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	petugasID, err := uuid.Parse(c.Params("petugasId"))
	if err != nil {
		return httpx.BadRequest("Invalid petugas ID.")
	}
	var req domain.ApproveLiturgiSubstituteRequest
	if err := decodeEmptyOrStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.ApproveSubstitute(c.Context(), p, eventID, petugasID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeletePetugas hard-deletes an officer assignment from a parish Mass schedule.
//
//	@Summary		Delete liturgy officer
//	@Description	Removes an officer assignment from a Mass schedule (requires ministry management permission).
//	@ID				deleteLiturgiPetugas
//	@Tags			Liturgi
//	@Produce		json
//	@Param			eventId		path	string	true	"Mass schedule event UUID"
//	@Param			petugasId	path	string	true	"Officer assignment UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/liturgi/schedules/{eventId}/petugas/{petugasId} [delete]
func (h *Handler) DeletePetugas(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	eventID, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return httpx.BadRequest("Invalid schedule ID.")
	}
	petugasID, err := uuid.Parse(c.Params("petugasId"))
	if err != nil {
		return httpx.BadRequest("Invalid petugas ID.")
	}
	if err := h.service.DeletePetugas(c.Context(), p, eventID, petugasID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
