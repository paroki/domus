package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/user"
)

const visionExtractionSystemPrompt = `You are an Indonesian civil identity document OCR/extraction assistant.
Extract fields from the provided Indonesian identity images (KTP, KIA, Kartu Pelajar, and/or Kartu Keluarga / KK).
Return ONLY valid JSON without markdown code fences or commentary, matching this exact schema:
{
  "ktp": {
    "nik": "string",
    "nama": "string",
    "tempatLahir": "string or null",
    "tanggalLahir": "YYYY-MM-DD or null",
    "jenisKelamin": "pria or wanita or null",
    "alamat": "string or null",
    "agama": "string or null",
    "golonganDarah": "string or null",
    "statusPernikahan": "string or null",
    "pekerjaan": "string or null"
  },
  "kk": {
    "nomorKk": "string",
    "namaKeluarga": "string",
    "alamat": "string or null",
    "anggota": [
      {
        "nama": "string",
        "nik": "string",
        "jenisKelamin": "pria or wanita",
        "tempatLahir": "string or null",
        "tanggalLahir": "YYYY-MM-DD or null",
        "agama": "string or null",
        "pendidikan": "string or null",
        "pekerjaan": "string or null",
        "statusPernikahan": "string or null",
        "hubunganKeluarga": "string or null"
      }
    ]
  }
}
Omit "ktp" if no KTP/KIA/Kartu Pelajar image was provided; omit "kk" if no KK image was provided.
Normalize dates to YYYY-MM-DD and jenisKelamin to "pria" or "wanita".`

// ExtractIdentityDocument sends identity document images to the configured vision chat model and parses strict JSON output.
func (s *Service) ExtractIdentityDocument(ctx context.Context, in user.IdentityExtractInput) (user.IdentityExtraction, error) {
	if in.KTP == nil && in.KartuPelajar == nil && in.KK == nil {
		return user.IdentityExtraction{}, domain.ValidationError("dokumen", "REQUIRED", "At least one identity document image is required.")
	}

	prov, plainKey, modelID, err := s.resolveSettingModel(ctx, settingModelVision, true)
	if err != nil {
		return user.IdentityExtraction{}, err
	}

	factory := s.chatFactory
	if factory == nil {
		factory = s.drivers.NewChatModel
	}
	cm, err := factory(ctx, string(prov.Provider), prov.BaseURL, plainKey, modelID)
	if err != nil {
		return user.IdentityExtraction{}, driver.WrapProviderError(err, plainKey)
	}

	parts := []schema.MessageInputPart{
		{
			Type: schema.ChatMessagePartTypeText,
			Text: "Extract all visible identity fields into the requested JSON structure.",
		},
	}
	appendImagePart := func(label string, img *user.IdentityImageInput) {
		if img == nil || len(img.Data) == 0 {
			return
		}
		b64 := base64.StdEncoding.EncodeToString(img.Data)
		parts = append(parts,
			schema.MessageInputPart{
				Type: schema.ChatMessagePartTypeText,
				Text: label,
			},
			schema.MessageInputPart{
				Type: schema.ChatMessagePartTypeImageURL,
				Image: &schema.MessageInputImage{
					MessagePartCommon: schema.MessagePartCommon{
						Base64Data: &b64,
						MIMEType:   img.MimeType,
					},
				},
			},
		)
	}
	appendImagePart("Dokumen: KTP", in.KTP)
	appendImagePart("Dokumen: Kartu Pelajar", in.KartuPelajar)
	appendImagePart("Dokumen: Kartu Keluarga (KK)", in.KK)

	userMsg := &schema.Message{
		Role:                  schema.User,
		UserInputMultiContent: parts,
	}

	resp, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(visionExtractionSystemPrompt),
		userMsg,
	})
	if err != nil {
		wrapped := driver.WrapProviderError(err, plainKey)
		if errors.Is(wrapped, domain.ErrProviderUnavailable) {
			lower := strings.ToLower(err.Error())
			if strings.Contains(lower, "image") || strings.Contains(lower, "vision") || strings.Contains(lower, "multimodal") || strings.Contains(lower, "unsupported") {
				return user.IdentityExtraction{}, fmt.Errorf("%w: %s", domain.ErrProviderRequestInvalid, err.Error())
			}
		}
		return user.IdentityExtraction{}, wrapped
	}

	if resp == nil {
		return user.IdentityExtraction{}, domain.ErrVerifyExtractionFailed
	}
	raw := cleanJSONFence(resp.Content)
	if raw == "" {
		return user.IdentityExtraction{}, domain.ErrVerifyExtractionFailed
	}

	var out user.IdentityExtraction
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return user.IdentityExtraction{}, domain.ErrVerifyExtractionFailed
	}
	if out.KTP == nil && out.KK == nil {
		return user.IdentityExtraction{}, domain.ErrVerifyExtractionFailed
	}
	if out.KTP != nil && strings.TrimSpace(out.KTP.Nama) == "" && strings.TrimSpace(out.KTP.NIK) == "" {
		return user.IdentityExtraction{}, domain.ErrVerifyExtractionFailed
	}
	if out.KK != nil && strings.TrimSpace(out.KK.NomorKK) == "" && len(out.KK.Anggota) == 0 {
		return user.IdentityExtraction{}, domain.ErrVerifyExtractionFailed
	}

	return out, nil
}

func cleanJSONFence(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
