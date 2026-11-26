package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/repo"
)

type nameCursor struct {
	V  string `json:"v"`
	ID string `json:"id"`
}

func encodeNameCursor(v string, id uuid.UUID) string {
	b, _ := json.Marshal(nameCursor{V: v, ID: id.String()})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeNameCursor(raw string) (*string, *uuid.UUID, error) {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, nil, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	var c nameCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, nil, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	id, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, nil, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	return &c.V, &id, nil
}

func decodeTimeCursor(raw string) (*time.Time, *uuid.UUID, error) {
	cur, err := httpx.DecodeCursor(raw)
	if err != nil {
		return nil, nil, err
	}
	id, err := uuid.Parse(cur.ID)
	if err != nil {
		return nil, nil, httpx.InvalidCursor("Invalid pagination cursor.")
	}
	t := cur.CreatedAt
	return &t, &id, nil
}

// SearchKeluarga returns keyset-paginated masked Keluarga records in parokiID.
func (s *Service) SearchKeluarga(
	ctx context.Context,
	parokiID uuid.UUID,
	req domain.SearchKeluargaRequest,
	limit int,
) ([]domain.KeluargaListItemResponse, string, bool, error) {
	f := repo.KeluargaFilter{
		Q:     strings.TrimSpace(req.Q),
		Sort:  req.Sort,
		Limit: limit + 1,
	}
	if req.Filter != nil {
		f.WilayahID = req.Filter.WilayahID
		f.LingkunganID = req.Filter.LingkunganID
		f.Aktif = req.Filter.Aktif
		if kk := strings.TrimSpace(req.Filter.NomorKK); kk != "" && s.cipher != nil {
			f.NomorKKHash = s.cipher.BlindIndex(kk)
		}
	}
	if req.Cursor != "" {
		if req.Sort == "namaKeluarga" || req.Sort == "-namaKeluarga" {
			name, id, err := decodeNameCursor(req.Cursor)
			if err != nil {
				return nil, "", false, err
			}
			f.CursorName = name
			f.CursorID = id
		} else {
			t, id, err := decodeTimeCursor(req.Cursor)
			if err != nil {
				return nil, "", false, err
			}
			f.CursorTime = t
			f.CursorID = id
		}
	}

	rows, err := s.repo.SearchKeluarga(ctx, parokiID, f)
	if err != nil {
		return nil, "", false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]domain.KeluargaListItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.toKeluargaListItem(row))
	}
	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		if req.Sort == "namaKeluarga" || req.Sort == "-namaKeluarga" {
			nextCursor = encodeNameCursor(last.NamaKeluarga, last.ID)
		} else {
			nextCursor = httpx.EncodeCursor(last.CreatedAt, last.ID.String())
		}
	}
	return items, nextCursor, hasMore, nil
}

// GetKeluarga returns a single non-deleted Keluarga with decrypted NomorKK and masked Anggota list.
func (s *Service) GetKeluarga(ctx context.Context, parokiID, id uuid.UUID) (domain.KeluargaDetailResponse, error) {
	k, err := s.repo.FindKeluargaByID(ctx, parokiID, id)
	if ent.IsNotFound(err) {
		return domain.KeluargaDetailResponse{}, domain.ErrKeluargaNotFound
	}
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	return s.toKeluargaDetail(k), nil
}

// CreateKeluarga creates a new Keluarga in parokiID.
func (s *Service) CreateKeluarga(ctx context.Context, parokiID uuid.UUID, req domain.CreateKeluargaRequest) (domain.KeluargaDetailResponse, error) {
	nama := strings.TrimSpace(req.NamaKeluarga)
	if nama == "" {
		return domain.KeluargaDetailResponse{}, domain.ValidationError("namaKeluarga", "REQUIRED", "Nama keluarga is required.")
	}

	var kkEnc, kkHash *string
	if req.NomorKK != nil {
		cleanKK, err := ValidatePII("nomorKk", *req.NomorKK)
		if err != nil {
			return domain.KeluargaDetailResponse{}, err
		}
		if cleanKK != "" {
			if s.cipher == nil {
				return domain.KeluargaDetailResponse{}, domain.ValidationError("nomorKk", "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
			}
			hash := s.cipher.BlindIndex(cleanKK)
			exists, err := s.repo.HasActiveNomorKKHash(ctx, parokiID, hash, nil)
			if err != nil {
				return domain.KeluargaDetailResponse{}, err
			}
			if exists {
				return domain.KeluargaDetailResponse{}, domain.ErrKKConflict
			}
			enc, err := s.cipher.EncryptString(cleanKK)
			if err != nil {
				return domain.KeluargaDetailResponse{}, err
			}
			kkEnc = &enc
			kkHash = &hash
		}
	}

	wilayahID, lingkunganID, _, _, err := s.resolveTerritorial(ctx, parokiID, req.WilayahID, req.LingkunganID, nil, nil)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	kotaID, kecamatanID, kelurahanID, _, _, _, err := s.resolveCivilTerritory(ctx, req.KotaID, req.KecamatanID, req.KelurahanID, nil, nil)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	statusRumah, err := s.validateLookupField(ctx, parokiID, "statusRumah", "status_rumah", req.StatusRumah)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	statusEkonomi, err := s.validateLookupField(ctx, parokiID, "statusEkonomi", "status_ekonomi", req.StatusEkonomi)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}

	aktif := true
	if req.Aktif != nil {
		aktif = *req.Aktif
	}

	created, err := s.repo.CreateKeluarga(ctx, func(c *ent.KeluargaCreate) *ent.KeluargaCreate {
		c.SetParokiID(parokiID).
			SetNamaKeluarga(nama).
			SetAktif(aktif).
			SetNillableWilayahID(wilayahID).
			SetNillableLingkunganID(lingkunganID).
			SetNillableNomorKkEnc(kkEnc).
			SetNillableNomorKkHash(kkHash).
			SetNillableAlamat(trimOptString(req.Alamat)).
			SetNillableKotaID(kotaID).
			SetNillableKecamatanID(kecamatanID).
			SetNillableKelurahanID(kelurahanID).
			SetNillableKodePos(trimOptString(req.KodePos)).
			SetNillableStatusRumah(nilIfEmpty(statusRumah)).
			SetNillableStatusEkonomi(nilIfEmpty(statusEkonomi)).
			SetNillableCatatan(trimOptString(req.Catatan))
		return c
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KeluargaDetailResponse{}, domain.ErrKKConflict
		}
		return domain.KeluargaDetailResponse{}, err
	}
	return s.toKeluargaDetail(created), nil
}

// UpdateKeluarga partially updates a Keluarga in parokiID.
func (s *Service) UpdateKeluarga(ctx context.Context, parokiID, id uuid.UUID, req domain.UpdateKeluargaRequest) (domain.KeluargaDetailResponse, error) {
	existing, err := s.repo.FindKeluargaByID(ctx, parokiID, id)
	if ent.IsNotFound(err) {
		return domain.KeluargaDetailResponse{}, domain.ErrKeluargaNotFound
	}
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}

	var nama *string
	if req.NamaKeluarga != nil {
		trimmed := strings.TrimSpace(*req.NamaKeluarga)
		if trimmed == "" {
			return domain.KeluargaDetailResponse{}, domain.ValidationError("namaKeluarga", "REQUIRED", "Nama keluarga cannot be empty.")
		}
		nama = &trimmed
	}

	var kkEnc, kkHash *string
	var clearKK bool
	if req.NomorKK != nil {
		cleanKK, err := ValidatePII("nomorKk", *req.NomorKK)
		if err != nil {
			return domain.KeluargaDetailResponse{}, err
		}
		if cleanKK == "" {
			clearKK = true
		} else {
			if s.cipher == nil {
				return domain.KeluargaDetailResponse{}, domain.ValidationError("nomorKk", "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
			}
			hash := s.cipher.BlindIndex(cleanKK)
			exists, err := s.repo.HasActiveNomorKKHash(ctx, parokiID, hash, &id)
			if err != nil {
				return domain.KeluargaDetailResponse{}, err
			}
			if exists {
				return domain.KeluargaDetailResponse{}, domain.ErrKKConflict
			}
			enc, err := s.cipher.EncryptString(cleanKK)
			if err != nil {
				return domain.KeluargaDetailResponse{}, err
			}
			kkEnc = &enc
			kkHash = &hash
		}
	}

	wilayahID, lingkunganID, clearWilayah, clearLingkungan, err := s.resolveTerritorial(
		ctx, parokiID, req.WilayahID, req.LingkunganID, existing.WilayahID, existing.LingkunganID,
	)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	kotaID, kecamatanID, kelurahanID, clearKota, clearKecamatan, clearKelurahan, err := s.resolveCivilTerritory(
		ctx, req.KotaID, req.KecamatanID, req.KelurahanID, existing.KotaID, existing.KecamatanID,
	)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	statusRumah, err := s.validateLookupField(ctx, parokiID, "statusRumah", "status_rumah", req.StatusRumah)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	statusEkonomi, err := s.validateLookupField(ctx, parokiID, "statusEkonomi", "status_ekonomi", req.StatusEkonomi)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}

	_, err = s.repo.UpdateKeluarga(ctx, id, func(u *ent.KeluargaUpdateOne) *ent.KeluargaUpdateOne {
		if nama != nil {
			u.SetNamaKeluarga(*nama)
		}
		if clearKK {
			u.ClearNomorKkEnc().ClearNomorKkHash()
		} else if kkEnc != nil && kkHash != nil {
			u.SetNomorKkEnc(*kkEnc).SetNomorKkHash(*kkHash)
		}
		if clearWilayah {
			u.ClearWilayahID()
		} else if wilayahID != nil {
			u.SetWilayahID(*wilayahID)
		}
		if clearLingkungan {
			u.ClearLingkunganID()
		} else if lingkunganID != nil {
			u.SetLingkunganID(*lingkunganID)
		}
		if clearKota {
			u.ClearKotaID()
		} else if kotaID != nil {
			u.SetKotaID(*kotaID)
		}
		if clearKecamatan {
			u.ClearKecamatanID()
		} else if kecamatanID != nil {
			u.SetKecamatanID(*kecamatanID)
		}
		if clearKelurahan {
			u.ClearKelurahanID()
		} else if kelurahanID != nil {
			u.SetKelurahanID(*kelurahanID)
		}
		setOrClearKeluargaStr(req.Alamat, u.SetAlamat, u.ClearAlamat)
		setOrClearKeluargaStr(req.KodePos, u.SetKodePos, u.ClearKodePos)
		setOrClearKeluargaStr(statusRumah, u.SetStatusRumah, u.ClearStatusRumah)
		setOrClearKeluargaStr(statusEkonomi, u.SetStatusEkonomi, u.ClearStatusEkonomi)
		setOrClearKeluargaStr(req.Catatan, u.SetCatatan, u.ClearCatatan)
		if req.Aktif != nil {
			u.SetAktif(*req.Aktif)
		}
		return u
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.KeluargaDetailResponse{}, domain.ErrKKConflict
		}
		return domain.KeluargaDetailResponse{}, err
	}

	refetched, err := s.repo.FindKeluargaByID(ctx, parokiID, id)
	if err != nil {
		return domain.KeluargaDetailResponse{}, err
	}
	return s.toKeluargaDetail(refetched), nil
}

// DeleteKeluarga soft-deletes a Keluarga if no active Umat references it.
func (s *Service) DeleteKeluarga(ctx context.Context, parokiID, id uuid.UUID) error {
	if _, err := s.repo.FindKeluargaByID(ctx, parokiID, id); ent.IsNotFound(err) {
		return domain.ErrKeluargaNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasActiveUmatByKeluarga(ctx, parokiID, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteKeluarga(ctx, id, time.Now().UTC())
}

func (s *Service) decryptAndMask(enc *string) (plain *string, masked *string) {
	if enc == nil || *enc == "" || s.cipher == nil {
		return nil, nil
	}
	p, err := s.cipher.DecryptString(*enc)
	if err != nil || p == "" {
		return nil, nil
	}
	m := domain.MaskPII(p)
	return &p, &m
}

func (s *Service) toKeluargaListItem(k *ent.Keluarga) domain.KeluargaListItemResponse {
	_, masked := s.decryptAndMask(k.NomorKkEnc)
	return domain.KeluargaListItemResponse{
		ID:            k.ID,
		ParokiID:      k.ParokiID,
		WilayahID:     k.WilayahID,
		LingkunganID:  k.LingkunganID,
		NomorKKMasked: masked,
		NamaKeluarga:  k.NamaKeluarga,
		Alamat:        k.Alamat,
		KotaID:        k.KotaID,
		KecamatanID:   k.KecamatanID,
		KelurahanID:   k.KelurahanID,
		KodePos:       k.KodePos,
		StatusRumah:   k.StatusRumah,
		StatusEkonomi: k.StatusEkonomi,
		Aktif:         k.Aktif,
		Catatan:       k.Catatan,
		CreatedAt:     k.CreatedAt,
		UpdatedAt:     k.UpdatedAt,
	}
}

func (s *Service) toKeluargaDetail(k *ent.Keluarga) domain.KeluargaDetailResponse {
	plain, masked := s.decryptAndMask(k.NomorKkEnc)
	anggota := make([]domain.KeluargaAnggotaSummary, 0, len(k.Edges.Anggota))
	for _, a := range k.Edges.Anggota {
		_, nikMasked := s.decryptAndMask(a.NikEnc)
		anggota = append(anggota, domain.KeluargaAnggotaSummary{
			ID:               a.ID,
			Nama:             a.Nama,
			NamaBaptis:       a.NamaBaptis,
			NIKMasked:        nikMasked,
			JenisKelamin:     string(a.JenisKelamin),
			HubunganKeluarga: a.HubunganKeluarga,
			Almarhum:         a.Almarhum,
		})
	}
	return domain.KeluargaDetailResponse{
		ID:            k.ID,
		ParokiID:      k.ParokiID,
		WilayahID:     k.WilayahID,
		LingkunganID:  k.LingkunganID,
		NomorKK:       plain,
		NomorKKMasked: masked,
		NamaKeluarga:  k.NamaKeluarga,
		Alamat:        k.Alamat,
		KotaID:        k.KotaID,
		KecamatanID:   k.KecamatanID,
		KelurahanID:   k.KelurahanID,
		KodePos:       k.KodePos,
		StatusRumah:   k.StatusRumah,
		StatusEkonomi: k.StatusEkonomi,
		Aktif:         k.Aktif,
		Catatan:       k.Catatan,
		Anggota:       anggota,
		CreatedAt:     k.CreatedAt,
		UpdatedAt:     k.UpdatedAt,
	}
}

func trimOptString(val *string) *string {
	if val == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func nilIfEmpty(val *string) *string {
	if val == nil || *val == "" {
		return nil
	}
	return val
}

func setOrClearKeluargaStr(val *string, setFn func(string) *ent.KeluargaUpdateOne, clearFn func() *ent.KeluargaUpdateOne) {
	if val == nil {
		return
	}
	trimmed := strings.TrimSpace(*val)
	if trimmed == "" {
		clearFn()
		return
	}
	setFn(trimmed)
}
