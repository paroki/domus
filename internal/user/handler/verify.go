package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user/domain"
)

const maxMultipartImageBytes = 2 * 1024 * 1024

func readMultipartImage(c fiber.Ctx, field string) (*domain.IdentityImageInput, error) {
	fh, err := c.FormFile(field)
	if err != nil || fh == nil {
		return nil, nil
	}
	if fh.Size > maxMultipartImageBytes {
		return nil, domain.ValidationError(field, "FILE_TOO_LARGE", field+" exceeds 2 MiB maximum size.")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, httpx.BadRequest("Failed to read uploaded file " + field + ".")
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(io.LimitReader(f, maxMultipartImageBytes+1))
	if err != nil {
		return nil, httpx.BadRequest("Failed to read uploaded file " + field + ".")
	}
	if len(data) == 0 {
		return nil, domain.ValidationError(field, "REQUIRED", field+" file is empty.")
	}
	if len(data) > maxMultipartImageBytes {
		return nil, domain.ValidationError(field, "FILE_TOO_LARGE", field+" exceeds 2 MiB maximum size.")
	}
	detected := http.DetectContentType(data)
	switch detected {
	case "image/jpeg", "image/png", "image/webp":
		return &domain.IdentityImageInput{MimeType: detected, Data: data}, nil
	default:
		return nil, domain.ValidationError(field, "INVALID_MIME", field+" must be image/jpeg, image/png, or image/webp.")
	}
}

// VerifyExtract extracts structured identity fields from uploaded KTP/Kartu Pelajar and KK images via AI Vision.
//
//	@Summary		Extract identity documents
//	@Description	Extracts structured identity data from KTP/Kartu Pelajar and KK photos using AI Vision and previews census match status.
//	@ID				verifyExtract
//	@Tags			Users
//	@Accept			mpfd
//	@Produce		json
//	@Param			ktp				formData	file	false	"KTP image (JPEG/PNG/WebP, max 2 MiB)"
//	@Param			kartuPelajar	formData	file	false	"Kartu Pelajar / KIA image (JPEG/PNG/WebP, max 2 MiB)"
//	@Param			kk				formData	file	false	"Kartu Keluarga image (JPEG/PNG/WebP, max 2 MiB; required when ktp is omitted)"
//	@Success		200				{object}	httpx.Envelope[domain.VerifyExtractResponse]
//	@Failure		400				{object}	httpx.Envelope[any]
//	@Failure		401				{object}	httpx.Envelope[any]
//	@Failure		422				{object}	httpx.Envelope[any]
//	@Failure		429				{object}	httpx.Envelope[any]
//	@Router			/me/verify/extract [post]
func (h *Handler) VerifyExtract(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	ktpImg, err := readMultipartImage(c, "ktp")
	if err != nil {
		return err
	}
	kpImg, err := readMultipartImage(c, "kartuPelajar")
	if err != nil {
		return err
	}
	kkImg, err := readMultipartImage(c, "kk")
	if err != nil {
		return err
	}

	res, err := h.service.ExtractIdentity(c.Context(), p, domain.IdentityExtractInput{
		KTP:          ktpImg,
		KartuPelajar: kpImg,
		KK:           kkImg,
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// VerifyConfirm confirms reviewed identity data, syncs parish census records, stores encrypted documents, and marks the user verified.
//
//	@Summary		Confirm identity verification
//	@Description	Re-validates reviewed identity data and documents, syncs keluarga and umat records, and marks the caller verified.
//	@ID				verifyConfirm
//	@Tags			Users
//	@Accept			mpfd
//	@Produce		json
//	@Param			payload			formData	string	true	"Reviewed IdentityExtraction JSON"
//	@Param			memberIndex		formData	int		false	"0-based index of caller in kk.anggota"
//	@Param			wilayahId		formData	string	false	"Optional Wilayah UUID"
//	@Param			lingkunganId	formData	string	false	"Optional Lingkungan UUID"
//	@Param			ktp				formData	file	false	"KTP image (JPEG/PNG/WebP, max 2 MiB)"
//	@Param			kartuPelajar	formData	file	false	"Kartu Pelajar image (JPEG/PNG/WebP, max 2 MiB)"
//	@Param			kia				formData	file	false	"KIA image (JPEG/PNG/WebP, max 2 MiB)"
//	@Param			kk				formData	file	false	"Kartu Keluarga image (JPEG/PNG/WebP, max 2 MiB)"
//	@Success		200				{object}	httpx.Envelope[domain.VerifyConfirmResult]
//	@Failure		400				{object}	httpx.Envelope[any]
//	@Failure		401				{object}	httpx.Envelope[any]
//	@Failure		409				{object}	httpx.Envelope[any]
//	@Failure		422				{object}	httpx.Envelope[any]
//	@Router			/me/verify/confirm [post]
func (h *Handler) VerifyConfirm(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}

	rawPayload := strings.TrimSpace(c.FormValue("payload"))
	if rawPayload == "" {
		return domain.ValidationError("payload", "REQUIRED", "payload is required.")
	}
	var payload domain.IdentityExtraction
	if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
		return httpx.BadRequest("Invalid payload JSON.")
	}

	memberIndex := 0
	if rawIdx := strings.TrimSpace(c.FormValue("memberIndex")); rawIdx != "" {
		idx, err := strconv.Atoi(rawIdx)
		if err != nil {
			return domain.ValidationError("memberIndex", "INVALID_FORMAT", "memberIndex must be an integer.")
		}
		memberIndex = idx
	}

	var wilayahID, lingkunganID *uuid.UUID
	if rawW := strings.TrimSpace(c.FormValue("wilayahId")); rawW != "" {
		id, err := uuid.Parse(rawW)
		if err != nil {
			return domain.ValidationError("wilayahId", "INVALID_UUID", "Invalid wilayahId.")
		}
		wilayahID = &id
	}
	if rawL := strings.TrimSpace(c.FormValue("lingkunganId")); rawL != "" {
		id, err := uuid.Parse(rawL)
		if err != nil {
			return domain.ValidationError("lingkunganId", "INVALID_UUID", "Invalid lingkunganId.")
		}
		lingkunganID = &id
	}

	var docs []domain.VerifyConfirmDocument
	docFields := []struct {
		field string
		jenis string
	}{
		{"ktp", "ktp"},
		{"kia", "kia"},
		{"kartuPelajar", "kartu_pelajar"},
		{"kk", "kk"},
	}
	for _, df := range docFields {
		img, err := readMultipartImage(c, df.field)
		if err != nil {
			return err
		}
		if img != nil {
			docs = append(docs, domain.VerifyConfirmDocument{
				Jenis:    df.jenis,
				MimeType: img.MimeType,
				Data:     img.Data,
			})
		}
	}

	res, err := h.service.ConfirmVerification(c.Context(), p, domain.VerifyConfirmInput{
		Payload:      payload,
		MemberIndex:  memberIndex,
		WilayahID:    wilayahID,
		LingkunganID: lingkunganID,
		Documents:    docs,
	})
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// GetMyUmat returns the caller's linked umat, keluarga, members, and document metadata in the active paroki.
//
//	@Summary		Get my umat profile
//	@Description	Returns the caller's linked umat, keluarga (with decrypted NIK and nomorKk), family members, and document metadata.
//	@ID				getMyUmat
//	@Tags			Users
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[domain.SelfUmatProfile]
//	@Failure		400	{object}	httpx.Envelope[any]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Router			/me/umat [get]
func (h *Handler) GetMyUmat(c fiber.Ctx) error {
	p, err := authx.RequireTenant(c.Context())
	if err != nil {
		return err
	}
	res, err := h.service.GetMyUmat(c.Context(), p)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}
