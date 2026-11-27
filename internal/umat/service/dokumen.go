package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umatdokumen"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/umat/domain"
)

// MaxDokumenSizeBytes is the maximum allowed identity document image size (2 MiB).
const MaxDokumenSizeBytes = 2 * 1024 * 1024

// SniffImageMimeType validates size and sniffs the image MIME type from raw bytes.
func SniffImageMimeType(field string, data []byte) (umatdokumen.MimeType, error) {
	if len(data) == 0 {
		return "", domain.ValidationError(field, "REQUIRED", field+" file is required.")
	}
	if len(data) > MaxDokumenSizeBytes {
		return "", domain.ValidationError(field, "FILE_TOO_LARGE", field+" exceeds 2 MiB maximum size.")
	}
	detected := http.DetectContentType(data)
	switch umatdokumen.MimeType(detected) {
	case umatdokumen.MimeTypeWebP, umatdokumen.MimeTypeJPEG, umatdokumen.MimeTypePNG:
		return umatdokumen.MimeType(detected), nil
	default:
		return "", domain.ValidationError(field, "INVALID_MIME", field+" must be image/jpeg, image/png, or image/webp.")
	}
}

func (s *Service) validateAndEncryptDokumen(field string, data []byte) (umatdokumen.MimeType, []byte, error) {
	mimeType, err := SniffImageMimeType(field, data)
	if err != nil {
		return "", nil, err
	}
	if s.cipher == nil {
		return "", nil, domain.ValidationError(field, "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
	}
	enc, err := s.cipher.EncryptBytes(data)
	if err != nil {
		return "", nil, err
	}
	return mimeType, enc, nil
}

// GetDokumen returns the decrypted bytes and MIME type of an identity document if callerUserID is authorized.
func (s *Service) GetDokumen(ctx context.Context, parokiID, callerUserID, dokumenID uuid.UUID) (string, []byte, error) {
	doc, err := s.repo.FindDokumenByID(ctx, parokiID, dokumenID)
	if ent.IsNotFound(err) {
		return "", nil, domain.ErrDokumenNotFound
	}
	if err != nil {
		return "", nil, err
	}
	if err := s.authorizeDokumenAccess(ctx, parokiID, callerUserID, doc); err != nil {
		return "", nil, err
	}
	if s.cipher == nil {
		return "", nil, domain.ValidationError("dokumen", "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
	}
	plain, err := s.cipher.DecryptBytes(doc.DataEnc)
	if err != nil {
		return "", nil, err
	}
	return string(doc.MimeType), plain, nil
}

// DeleteDokumen hard-deletes an identity document if callerUserID is authorized.
func (s *Service) DeleteDokumen(ctx context.Context, parokiID, callerUserID, dokumenID uuid.UUID) error {
	doc, err := s.repo.FindDokumenByID(ctx, parokiID, dokumenID)
	if ent.IsNotFound(err) {
		return domain.ErrDokumenNotFound
	}
	if err != nil {
		return err
	}
	if err := s.authorizeDokumenAccess(ctx, parokiID, callerUserID, doc); err != nil {
		return err
	}
	return s.repo.DeleteDokumen(ctx, parokiID, dokumenID)
}

func (s *Service) authorizeDokumenAccess(ctx context.Context, parokiID, callerUserID uuid.UUID, doc *ent.UmatDokumen) error {
	if doc.Edges.Umat != nil && doc.Edges.Umat.UserID != nil && *doc.Edges.Umat.UserID == callerUserID {
		return nil
	}

	callerUmat, err := s.repo.FindUmatByUserID(ctx, parokiID, callerUserID)
	if err == nil && callerUmat != nil && callerUmat.KeluargaID != nil &&
		callerUmat.HubunganKeluarga != nil && strings.EqualFold(*callerUmat.HubunganKeluarga, "kepala-keluarga") {
		var targetKeluargaID *uuid.UUID
		if doc.KeluargaID != nil {
			targetKeluargaID = doc.KeluargaID
		} else if doc.Edges.Umat != nil && doc.Edges.Umat.KeluargaID != nil {
			targetKeluargaID = doc.Edges.Umat.KeluargaID
		}
		if targetKeluargaID != nil && *targetKeluargaID == *callerUmat.KeluargaID {
			return nil
		}
	} else if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if s.enforcer != nil {
		allowed, err := s.enforcer.HasRoleInDomain(authz.UserSubject(callerUserID), domain.RoleAdmin, authz.ParokiDomain(parokiID))
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
	}

	return domain.ErrForbidden
}
