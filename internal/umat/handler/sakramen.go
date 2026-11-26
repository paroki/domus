package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/umat/domain"
)

// UpsertSakramen idempotently upserts a sacramental history row for an Umat.
//
//	@Summary		Upsert umat sakramen
//	@Description	Idempotently creates or updates a sacramental history record keyed by (parokiId, umatId, jenis).
//	@ID				upsertUmatSakramen
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"Umat UUID"
//	@Param			body	body		domain.UpsertUmatSakramenRequest	true	"Sakramen payload"
//	@Success		200		{object}	httpx.Envelope[domain.UmatSakramenResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		404		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/umat/{id}/sakramen [post]
func (h *Handler) UpsertSakramen(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	umatID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid umat ID.")
	}
	var req domain.UpsertUmatSakramenRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpsertSakramen(c.Context(), p.ParokiID, umatID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// UpdateSakramen partially updates an existing sacramental history row for an Umat.
//
//	@Summary		Update umat sakramen
//	@Description	Partially updates a sacramental history record belonging to the umat in the active paroki.
//	@ID				updateUmatSakramen
//	@Tags			Umat
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string							true	"Umat UUID"
//	@Param			sakramenId	path		string							true	"Sakramen UUID"
//	@Param			body		body		domain.UpdateUmatSakramenRequest	true	"Sakramen update payload"
//	@Success		200			{object}	httpx.Envelope[domain.UmatSakramenResponse]
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Failure		409			{object}	httpx.Envelope[any]
//	@Failure		422			{object}	httpx.Envelope[any]
//	@Router			/umat/{id}/sakramen/{sakramenId} [patch]
func (h *Handler) UpdateSakramen(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	umatID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid umat ID.")
	}
	sakramenID, err := uuid.Parse(c.Params("sakramenId"))
	if err != nil {
		return httpx.BadRequest("Invalid sakramen ID.")
	}
	var req domain.UpdateUmatSakramenRequest
	if err := decodeStrictJSON(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateSakramen(c.Context(), p.ParokiID, umatID, sakramenID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// DeleteSakramen physically deletes a sacramental history row for an Umat.
//
//	@Summary		Delete umat sakramen
//	@Description	Hard-deletes a sacramental history record belonging to the umat in the active paroki.
//	@ID				deleteUmatSakramen
//	@Tags			Umat
//	@Param			id			path	string	true	"Umat UUID"
//	@Param			sakramenId	path	string	true	"Sakramen UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/umat/{id}/sakramen/{sakramenId} [delete]
func (h *Handler) DeleteSakramen(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	umatID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.BadRequest("Invalid umat ID.")
	}
	sakramenID, err := uuid.Parse(c.Params("sakramenId"))
	if err != nil {
		return httpx.BadRequest("Invalid sakramen ID.")
	}
	if err := h.service.DeleteSakramen(c.Context(), p.ParokiID, umatID, sakramenID); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
