package service

import (
	"context"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/umat/domain"
)

var e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

// ValidatePII trims and validates a NIK or Nomor KK string (4..32 UTF-8 runes, no ASCII control characters).
// Empty/whitespace-only input returns "", nil so callers can treat it as omitted or cleared.
func ValidatePII(field, raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	n := utf8.RuneCountInString(trimmed)
	if n < 4 || n > 32 {
		return "", domain.ValidationError(field, "INVALID_LENGTH", field+" must be between 4 and 32 characters.")
	}
	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			return "", domain.ValidationError(field, "INVALID_FORMAT", field+" must not contain control characters.")
		}
	}
	return trimmed, nil
}

// NormalizePhone normalizes a phone number to canonical E.164 (defaulting local ID numbers to +62).
func NormalizePhone(raw string) (string, error) {
	cleaned := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" {
		return "", nil
	}
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "+62" + cleaned[1:]
	} else if strings.HasPrefix(cleaned, "62") && !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}
	if !e164Re.MatchString(cleaned) {
		return "", domain.ValidationError("telepon", "INVALID_PHONE", "Telepon must be a valid E.164 number.")
	}
	return cleaned, nil
}

// NormalizeEmail normalizes and validates an email address.
func NormalizeEmail(raw string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))
	if trimmed == "" {
		return "", nil
	}
	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed || !strings.Contains(trimmed, "@") {
		return "", domain.ValidationError("email", "INVALID_EMAIL", "Email must be a valid email address.")
	}
	return trimmed, nil
}

// ParseDateYYYYMMDD parses a strict YYYY-MM-DD date string.
func ParseDateYYYYMMDD(field, raw string) (string, time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", time.Time{}, nil
	}
	t, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return "", time.Time{}, domain.ValidationError(field, "INVALID_DATE", field+" must be in YYYY-MM-DD format.")
	}
	return trimmed, t, nil
}

// ValidateDateOrder ensures tanggalMeninggal is not earlier than tanggalLahir.
func ValidateDateOrder(tanggalLahir, tanggalMeninggal *string) error {
	if tanggalLahir == nil || *tanggalLahir == "" || tanggalMeninggal == nil || *tanggalMeninggal == "" {
		return nil
	}
	_, birth, err := ParseDateYYYYMMDD("tanggalLahir", *tanggalLahir)
	if err != nil {
		return err
	}
	_, death, err := ParseDateYYYYMMDD("tanggalMeninggal", *tanggalMeninggal)
	if err != nil {
		return err
	}
	if death.Before(birth) {
		return domain.ValidationError("tanggalMeninggal", "INVALID_DATE_RANGE", "Tanggal meninggal cannot be earlier than tanggal lahir.")
	}
	return nil
}

func (s *Service) validateLookupField(ctx context.Context, parokiID uuid.UUID, field, category string, raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	norm := strings.ToLower(strings.TrimSpace(*raw))
	if norm == "" {
		empty := ""
		return &empty, nil
	}
	if s.refs != nil {
		ok, err := s.refs.HasActiveLookup(ctx, parokiID, category, norm)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ValidationError(field, "INVALID_LOOKUP", field+" must reference an active "+category+" lookup.")
		}
	}
	return &norm, nil
}

func (s *Service) resolveTerritorial(
	ctx context.Context,
	parokiID uuid.UUID,
	reqWilayahID, reqLingkunganID *uuid.UUID,
	curWilayahID, curLingkunganID *uuid.UUID,
) (wilayahID *uuid.UUID, lingkunganID *uuid.UUID, clearWilayah bool, clearLingkungan bool, err error) {
	effWilayah := curWilayahID
	if reqWilayahID != nil {
		if *reqWilayahID == uuid.Nil {
			effWilayah = nil
			clearWilayah = true
			if reqLingkunganID == nil {
				clearLingkungan = true
			}
		} else {
			effWilayah = reqWilayahID
			wilayahID = reqWilayahID
		}
	}

	if reqLingkunganID != nil {
		if *reqLingkunganID == uuid.Nil {
			clearLingkungan = true
		} else {
			if s.refs != nil {
				ling, err := s.refs.ValidateLingkunganInParoki(ctx, parokiID, *reqLingkunganID)
				if err != nil {
					return nil, nil, false, false, domain.ValidationError("lingkunganId", "INVALID_REF", "Lingkungan not found in active paroki.")
				}
				if effWilayah == nil {
					wID := ling.WilayahID
					wilayahID = &wID
					clearWilayah = false
				} else if *effWilayah != ling.WilayahID {
					return nil, nil, false, false, domain.ValidationError("lingkunganId", "INVALID_REF", "Lingkungan does not belong to the specified wilayah.")
				}
			}
			lingkunganID = reqLingkunganID
			clearLingkungan = false
		}
	}

	if wilayahID != nil && *wilayahID != uuid.Nil && s.refs != nil {
		if _, err := s.refs.ValidateWilayahInParoki(ctx, parokiID, *wilayahID); err != nil {
			return nil, nil, false, false, domain.ValidationError("wilayahId", "INVALID_REF", "Wilayah not found in active paroki.")
		}
		// If wilayah changed without a new lingkunganId, clear stale lingkunganId if it belongs to old wilayah.
		if reqLingkunganID == nil && curLingkunganID != nil && curWilayahID != nil && *wilayahID != *curWilayahID {
			clearLingkungan = true
		}
	}

	return wilayahID, lingkunganID, clearWilayah, clearLingkungan, nil
}

func (s *Service) resolveCivilTerritory(
	ctx context.Context,
	reqKotaID, reqKecamatanID, reqKelurahanID *uuid.UUID,
	curKotaID, curKecamatanID *uuid.UUID,
) (kotaID, kecamatanID, kelurahanID *uuid.UUID, clearKota, clearKecamatan, clearKelurahan bool, err error) {
	effKota := curKotaID
	if reqKotaID != nil {
		if *reqKotaID == uuid.Nil {
			effKota = nil
			clearKota = true
			if reqKecamatanID == nil {
				clearKecamatan = true
			}
			if reqKelurahanID == nil {
				clearKelurahan = true
			}
		} else {
			if s.refs != nil {
				if _, err := s.refs.ValidateKota(ctx, *reqKotaID); err != nil {
					return nil, nil, nil, false, false, false, domain.ValidationError("kotaId", "INVALID_REF", "Kota not found.")
				}
			}
			effKota = reqKotaID
			kotaID = reqKotaID
			if reqKecamatanID == nil && curKotaID != nil && *reqKotaID != *curKotaID {
				clearKecamatan = true
				clearKelurahan = true
			}
		}
	}

	effKec := curKecamatanID
	if clearKecamatan {
		effKec = nil
	}
	if reqKecamatanID != nil {
		if *reqKecamatanID == uuid.Nil {
			effKec = nil
			clearKecamatan = true
			if reqKelurahanID == nil {
				clearKelurahan = true
			}
		} else {
			if effKota == nil {
				return nil, nil, nil, false, false, false, domain.ValidationError("kecamatanId", "INVALID_REF", "KotaId is required when kecamatanId is provided.")
			}
			if s.refs != nil {
				if _, err := s.refs.ValidateKecamatan(ctx, *effKota, *reqKecamatanID); err != nil {
					return nil, nil, nil, false, false, false, domain.ValidationError("kecamatanId", "INVALID_REF", "Kecamatan does not belong to kota.")
				}
			}
			effKec = reqKecamatanID
			kecamatanID = reqKecamatanID
			clearKecamatan = false
			if reqKelurahanID == nil && curKecamatanID != nil && *reqKecamatanID != *curKecamatanID {
				clearKelurahan = true
			}
		}
	}

	if reqKelurahanID != nil {
		if *reqKelurahanID == uuid.Nil {
			clearKelurahan = true
		} else {
			if effKec == nil {
				return nil, nil, nil, false, false, false, domain.ValidationError("kelurahanId", "INVALID_REF", "KecamatanId is required when kelurahanId is provided.")
			}
			if s.refs != nil {
				if _, err := s.refs.ValidateKelurahan(ctx, *effKec, *reqKelurahanID); err != nil {
					return nil, nil, nil, false, false, false, domain.ValidationError("kelurahanId", "INVALID_REF", "Kelurahan does not belong to kecamatan.")
				}
			}
			kelurahanID = reqKelurahanID
			clearKelurahan = false
		}
	}

	return kotaID, kecamatanID, kelurahanID, clearKota, clearKecamatan, clearKelurahan, nil
}
