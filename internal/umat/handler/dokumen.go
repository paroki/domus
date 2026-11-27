package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

// GetDokumen decrypts and streams an identity document image for an authorized caller.
//
//	@Summary		Get umat dokumen
//	@Description	Decrypts and returns an identity document image (owner, head of keluarga, or umat:admin).
//	@ID				getUmatDokumen
//	@Tags			Umat
//	@Produce		image/jpeg,image/png,image/webp
//	@Param			dokumenId	path		string	true	"Dokumen UUID"
//	@Success		200			{file}		binary
//	@Failure		400			{object}	httpx.Envelope[any]
//	@Failure		401			{object}	httpx.Envelope[any]
//	@Failure		403			{object}	httpx.Envelope[any]
//	@Failure		404			{object}	httpx.Envelope[any]
//	@Router			/umat/dokumen/{dokumenId} [get]
func (h *Handler) GetDokumen(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("dokumenId"))
	if err != nil {
		return httpx.BadRequest("Invalid dokumen ID.")
	}
	mimeType, data, err := h.service.GetDokumen(c.Context(), p.ParokiID, p.UserID, id)
	if err != nil {
		return err
	}
	c.Set("Cache-Control", "no-store, private")
	c.Set("Content-Type", mimeType)
	return c.Send(data)
}

// DeleteDokumen hard-deletes an identity document for an authorized caller.
//
//	@Summary		Delete umat dokumen
//	@Description	Hard-deletes an identity document (owner, head of keluarga, or umat:admin).
//	@ID				deleteUmatDokumen
//	@Tags			Umat
//	@Param			dokumenId	path	string	true	"Dokumen UUID"
//	@Success		204
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Failure		404	{object}	httpx.Envelope[any]
//	@Router			/umat/dokumen/{dokumenId} [delete]
func (h *Handler) DeleteDokumen(c fiber.Ctx) error {
	p, ok := authx.UserContext(c.Context())
	if !ok {
		return httpx.Unauthorized("Authentication is required.")
	}
	id, err := uuid.Parse(c.Params("dokumenId"))
	if err != nil {
		return httpx.BadRequest("Invalid dokumen ID.")
	}
	if err := h.service.DeleteDokumen(c.Context(), p.ParokiID, p.UserID, id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}
