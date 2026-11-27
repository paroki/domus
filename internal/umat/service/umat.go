package service

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/repo"
)

// SearchUmat returns keyset-paginated masked Umat records in parokiID.
func (s *Service) SearchUmat(
	ctx context.Context,
	parokiID uuid.UUID,
	req domain.SearchUmatRequest,
	limit int,
) ([]domain.UmatListItemResponse, string, bool, error) {
	f := repo.UmatFilter{
		Q:     strings.TrimSpace(req.Q),
		Sort:  req.Sort,
		Limit: limit + 1,
	}
	if req.Filter != nil {
		f.KeluargaID = req.Filter.KeluargaID
		f.WilayahID = req.Filter.WilayahID
		f.LingkunganID = req.Filter.LingkunganID
		f.JenisKelamin = req.Filter.JenisKelamin
		f.Agama = strings.ToLower(strings.TrimSpace(req.Filter.Agama))
		f.StatusPernikahan = strings.ToLower(strings.TrimSpace(req.Filter.StatusPernikahan))
		f.Almarhum = req.Filter.Almarhum
		if tl := strings.TrimSpace(req.Filter.TanggalLahir); tl != "" {
			parsed, _, err := ParseDateYYYYMMDD("tanggalLahir", tl)
			if err != nil {
				return nil, "", false, err
			}
			f.TanggalLahir = parsed
		}
		if nik := strings.TrimSpace(req.Filter.NIK); nik != "" && s.cipher != nil {
			f.NIKHash = s.cipher.BlindIndex(nik)
		}
	}
	if req.Cursor != "" {
		if req.Sort == "nama" || req.Sort == "-nama" {
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

	rows, err := s.repo.SearchUmat(ctx, parokiID, f)
	if err != nil {
		return nil, "", false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]domain.UmatListItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, s.toUmatListItem(row))
	}
	var nextCursor string
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		if req.Sort == "nama" || req.Sort == "-nama" {
			nextCursor = encodeNameCursor(last.Nama, last.ID)
		} else {
			nextCursor = httpx.EncodeCursor(last.CreatedAt, last.ID.String())
		}
	}
	return items, nextCursor, hasMore, nil
}

// GetUmat returns a single non-deleted Umat with decrypted NIK, Keluarga summary, and Sakramen.
func (s *Service) GetUmat(ctx context.Context, parokiID, id uuid.UUID) (domain.UmatDetailResponse, error) {
	u, err := s.repo.FindUmatByID(ctx, parokiID, id)
	if ent.IsNotFound(err) {
		return domain.UmatDetailResponse{}, domain.ErrUmatNotFound
	}
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	return s.toUmatDetail(u), nil
}

// CreateUmat creates a new Umat in parokiID.
func (s *Service) CreateUmat(ctx context.Context, parokiID uuid.UUID, req domain.CreateUmatRequest) (domain.UmatDetailResponse, error) {
	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		return domain.UmatDetailResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama is required.")
	}

	var nikEnc, nikHash *string
	if req.NIK != nil {
		cleanNIK, err := ValidatePII("nik", *req.NIK)
		if err != nil {
			return domain.UmatDetailResponse{}, err
		}
		if cleanNIK != "" {
			if s.cipher == nil {
				return domain.UmatDetailResponse{}, domain.ValidationError("nik", "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
			}
			hash := s.cipher.BlindIndex(cleanNIK)
			exists, err := s.repo.HasActiveNIKHash(ctx, parokiID, hash, nil)
			if err != nil {
				return domain.UmatDetailResponse{}, err
			}
			if exists {
				return domain.UmatDetailResponse{}, domain.ErrNIKConflict
			}
			enc, err := s.cipher.EncryptString(cleanNIK)
			if err != nil {
				return domain.UmatDetailResponse{}, err
			}
			nikEnc = &enc
			nikHash = &hash
		}
	}

	var keluargaID *uuid.UUID
	if req.KeluargaID != nil && *req.KeluargaID != uuid.Nil {
		if _, err := s.repo.FindKeluargaByID(ctx, parokiID, *req.KeluargaID); ent.IsNotFound(err) {
			return domain.UmatDetailResponse{}, domain.ValidationError("keluargaId", "INVALID_REF", "Keluarga not found in active paroki.")
		} else if err != nil {
			return domain.UmatDetailResponse{}, err
		}
		keluargaID = req.KeluargaID
	}

	var userID *uuid.UUID
	if req.UserID != nil && *req.UserID != uuid.Nil {
		if err := s.validateUserLink(ctx, parokiID, *req.UserID, nil); err != nil {
			return domain.UmatDetailResponse{}, err
		}
		userID = req.UserID
	}

	wilayahID, lingkunganID, _, _, err := s.resolveTerritorial(ctx, parokiID, req.WilayahID, req.LingkunganID, nil, nil)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	kotaID, kecamatanID, kelurahanID, _, _, _, err := s.resolveCivilTerritory(ctx, req.KotaID, req.KecamatanID, req.KelurahanID, nil, nil)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	lookups, err := s.validateUmatLookups(ctx, parokiID, umatLookupInputs{
		Agama:            req.Agama,
		GolonganDarah:    req.GolonganDarah,
		Suku:             req.Suku,
		HubunganKeluarga: req.HubunganKeluarga,
		StatusPernikahan: req.StatusPernikahan,
		Pendidikan:       req.Pendidikan,
		Pekerjaan:        req.Pekerjaan,
		Profesi:          req.Profesi,
		Disabilitas:      req.Disabilitas,
		AktivitasSosial:  req.AktivitasSosial,
	})
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	tanggalLahir, err := parseOptDate("tanggalLahir", req.TanggalLahir)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	tanggalMeninggal, err := parseOptDate("tanggalMeninggal", req.TanggalMeninggal)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	if err := ValidateDateOrder(tanggalLahir, tanggalMeninggal); err != nil {
		return domain.UmatDetailResponse{}, err
	}

	telepon, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	email, err := normalizeOptEmail(req.Email)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	almarhum := false
	if req.Almarhum != nil {
		almarhum = *req.Almarhum
	}
	if !almarhum && req.TanggalMeninggal == nil {
		tanggalMeninggal = nil
	}

	var customNomor *string
	if req.NomorInduk != nil {
		if trimmed := strings.TrimSpace(*req.NomorInduk); trimmed != "" {
			exists, err := s.repo.HasActiveNomorInduk(ctx, parokiID, trimmed, nil)
			if err != nil {
				return domain.UmatDetailResponse{}, err
			}
			if exists {
				return domain.UmatDetailResponse{}, domain.ErrNomorIndukConflict
			}
			customNomor = &trimmed
		}
	}

	var created *ent.Umat
	const maxAttempts = 3
	for attempt := range maxAttempts {
		txErr := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
			nomorInduk := ""
			if customNomor != nil {
				nomorInduk = *customNomor
			} else {
				var gerr error
				nomorInduk, gerr = txRepo.NextNomorInduk(ctx, parokiID, time.Now().UTC().Year())
				if gerr != nil {
					return gerr
				}
			}
			row, cerr := txRepo.CreateUmat(ctx, func(c *ent.UmatCreate) *ent.UmatCreate {
				c.SetParokiID(parokiID).
					SetNomorInduk(nomorInduk).
					SetNama(nama).
					SetJenisKelamin(umat.JenisKelamin(req.JenisKelamin)).
					SetAlmarhum(almarhum).
					SetNillableKeluargaID(keluargaID).
					SetNillableWilayahID(wilayahID).
					SetNillableLingkunganID(lingkunganID).
					SetNillableUserID(userID).
					SetNillableNamaBaptis(trimOptString(req.NamaBaptis)).
					SetNillableNikEnc(nikEnc).
					SetNillableNikHash(nikHash).
					SetNillableTempatLahir(trimOptString(req.TempatLahir)).
					SetNillableTanggalLahir(nilIfEmpty(tanggalLahir)).
					SetNillableGolonganDarah(nilIfEmpty(lookups.GolonganDarah)).
					SetNillableSuku(nilIfEmpty(lookups.Suku)).
					SetNillableAgama(nilIfEmpty(lookups.Agama)).
					SetNillableKewarganegaraan(trimOptString(req.Kewarganegaraan)).
					SetNillableNegara(trimOptString(req.Negara)).
					SetNillableHubunganKeluarga(nilIfEmpty(lookups.HubunganKeluarga)).
					SetNillableStatusPernikahan(nilIfEmpty(lookups.StatusPernikahan)).
					SetNillableTanggalMeninggal(nilIfEmpty(tanggalMeninggal)).
					SetNillableAlamat(trimOptString(req.Alamat)).
					SetNillableKotaID(kotaID).
					SetNillableKecamatanID(kecamatanID).
					SetNillableKelurahanID(kelurahanID).
					SetNillableTelepon(nilIfEmpty(telepon)).
					SetNillableEmail(nilIfEmpty(email)).
					SetNillableFotoURL(trimOptString(req.FotoURL)).
					SetNillablePendidikan(nilIfEmpty(lookups.Pendidikan)).
					SetNillableJurusan(trimOptString(req.Jurusan)).
					SetNillablePekerjaan(nilIfEmpty(lookups.Pekerjaan)).
					SetNillableProfesi(nilIfEmpty(lookups.Profesi)).
					SetNillableKeterampilan(trimOptString(req.Keterampilan)).
					SetNillableDisabilitas(nilIfEmpty(lookups.Disabilitas)).
					SetNillableAktivitasSosial(nilIfEmpty(lookups.AktivitasSosial))
				return c
			})
			if cerr != nil {
				return cerr
			}
			created = row
			return nil
		})
		if txErr == nil {
			break
		}
		if ent.IsConstraintError(txErr) {
			if customNomor == nil && attempt < maxAttempts-1 {
				if nikHash != nil {
					if ok, herr := s.repo.HasActiveNIKHash(ctx, parokiID, *nikHash, nil); herr == nil && ok {
						return domain.UmatDetailResponse{}, domain.ErrNIKConflict
					}
				}
				if userID != nil {
					if ok, herr := s.repo.HasActiveUserLink(ctx, parokiID, *userID, nil); herr == nil && ok {
						return domain.UmatDetailResponse{}, domain.ErrUserAlreadyLinked
					}
				}
				continue
			}
			return domain.UmatDetailResponse{}, s.classifyUmatConstraintError(ctx, parokiID, nikHash, userID, customNomor, nil)
		}
		return domain.UmatDetailResponse{}, txErr
	}

	refetched, err := s.repo.FindUmatByID(ctx, parokiID, created.ID)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	return s.toUmatDetail(refetched), nil
}

// UpdateUmat partially updates an Umat in parokiID.
func (s *Service) UpdateUmat(ctx context.Context, parokiID, id uuid.UUID, req domain.UpdateUmatRequest) (domain.UmatDetailResponse, error) {
	existing, err := s.repo.FindUmatByID(ctx, parokiID, id)
	if ent.IsNotFound(err) {
		return domain.UmatDetailResponse{}, domain.ErrUmatNotFound
	}
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	var nomorInduk *string
	if req.NomorInduk != nil {
		trimmed := strings.TrimSpace(*req.NomorInduk)
		if trimmed == "" {
			return domain.UmatDetailResponse{}, domain.ValidationError("nomorInduk", "REQUIRED", "Nomor induk cannot be empty.")
		}
		exists, err := s.repo.HasActiveNomorInduk(ctx, parokiID, trimmed, &id)
		if err != nil {
			return domain.UmatDetailResponse{}, err
		}
		if exists {
			return domain.UmatDetailResponse{}, domain.ErrNomorIndukConflict
		}
		nomorInduk = &trimmed
	}

	var nama *string
	if req.Nama != nil {
		trimmed := strings.TrimSpace(*req.Nama)
		if trimmed == "" {
			return domain.UmatDetailResponse{}, domain.ValidationError("nama", "REQUIRED", "Nama cannot be empty.")
		}
		nama = &trimmed
	}

	var nikEnc, nikHash *string
	var clearNIK bool
	if req.NIK != nil {
		cleanNIK, err := ValidatePII("nik", *req.NIK)
		if err != nil {
			return domain.UmatDetailResponse{}, err
		}
		if cleanNIK == "" {
			clearNIK = true
		} else {
			if s.cipher == nil {
				return domain.UmatDetailResponse{}, domain.ValidationError("nik", "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
			}
			hash := s.cipher.BlindIndex(cleanNIK)
			exists, err := s.repo.HasActiveNIKHash(ctx, parokiID, hash, &id)
			if err != nil {
				return domain.UmatDetailResponse{}, err
			}
			if exists {
				return domain.UmatDetailResponse{}, domain.ErrNIKConflict
			}
			enc, err := s.cipher.EncryptString(cleanNIK)
			if err != nil {
				return domain.UmatDetailResponse{}, err
			}
			nikEnc = &enc
			nikHash = &hash
		}
	}

	var keluargaID *uuid.UUID
	var clearKeluarga bool
	if req.KeluargaID != nil {
		if *req.KeluargaID == uuid.Nil {
			clearKeluarga = true
		} else {
			if _, err := s.repo.FindKeluargaByID(ctx, parokiID, *req.KeluargaID); ent.IsNotFound(err) {
				return domain.UmatDetailResponse{}, domain.ValidationError("keluargaId", "INVALID_REF", "Keluarga not found in active paroki.")
			} else if err != nil {
				return domain.UmatDetailResponse{}, err
			}
			keluargaID = req.KeluargaID
		}
	}

	var userID *uuid.UUID
	var clearUser bool
	if req.UserID != nil {
		if *req.UserID == uuid.Nil {
			clearUser = true
		} else {
			if err := s.validateUserLink(ctx, parokiID, *req.UserID, &id); err != nil {
				return domain.UmatDetailResponse{}, err
			}
			userID = req.UserID
		}
	}

	wilayahID, lingkunganID, clearWilayah, clearLingkungan, err := s.resolveTerritorial(
		ctx, parokiID, req.WilayahID, req.LingkunganID, existing.WilayahID, existing.LingkunganID,
	)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	kotaID, kecamatanID, kelurahanID, clearKota, clearKecamatan, clearKelurahan, err := s.resolveCivilTerritory(
		ctx, req.KotaID, req.KecamatanID, req.KelurahanID, existing.KotaID, existing.KecamatanID,
	)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	lookups, err := s.validateUmatLookups(ctx, parokiID, umatLookupInputs{
		Agama:            req.Agama,
		GolonganDarah:    req.GolonganDarah,
		Suku:             req.Suku,
		HubunganKeluarga: req.HubunganKeluarga,
		StatusPernikahan: req.StatusPernikahan,
		Pendidikan:       req.Pendidikan,
		Pekerjaan:        req.Pekerjaan,
		Profesi:          req.Profesi,
		Disabilitas:      req.Disabilitas,
		AktivitasSosial:  req.AktivitasSosial,
	})
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	tanggalLahir, err := parseOptDate("tanggalLahir", req.TanggalLahir)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	tanggalMeninggal, err := parseOptDate("tanggalMeninggal", req.TanggalMeninggal)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	clearTanggalMeninggal := req.Almarhum != nil && !*req.Almarhum && req.TanggalMeninggal == nil
	effBirth := existing.TanggalLahir
	if tanggalLahir != nil {
		if *tanggalLahir == "" {
			effBirth = nil
		} else {
			effBirth = tanggalLahir
		}
	}
	effDeath := existing.TanggalMeninggal
	if clearTanggalMeninggal {
		effDeath = nil
	} else if tanggalMeninggal != nil {
		if *tanggalMeninggal == "" {
			effDeath = nil
		} else {
			effDeath = tanggalMeninggal
		}
	}
	if err := ValidateDateOrder(effBirth, effDeath); err != nil {
		return domain.UmatDetailResponse{}, err
	}

	telepon, err := normalizeOptPhone(req.Telepon)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	email, err := normalizeOptEmail(req.Email)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}

	_, err = s.repo.UpdateUmat(ctx, id, func(u *ent.UmatUpdateOne) *ent.UmatUpdateOne {
		if nomorInduk != nil {
			u.SetNomorInduk(*nomorInduk)
		}
		if nama != nil {
			u.SetNama(*nama)
		}
		if req.JenisKelamin != nil {
			u.SetJenisKelamin(umat.JenisKelamin(*req.JenisKelamin))
		}
		if req.Almarhum != nil {
			u.SetAlmarhum(*req.Almarhum)
		}
		if clearKeluarga {
			u.ClearKeluargaID()
		} else if keluargaID != nil {
			u.SetKeluargaID(*keluargaID)
		}
		if clearUser {
			u.ClearUserID()
		} else if userID != nil {
			u.SetUserID(*userID)
		}
		if clearNIK {
			u.ClearNikEnc().ClearNikHash()
		} else if nikEnc != nil && nikHash != nil {
			u.SetNikEnc(*nikEnc).SetNikHash(*nikHash)
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
		if clearTanggalMeninggal {
			u.ClearTanggalMeninggal()
		} else {
			setOrClearUmatStr(tanggalMeninggal, u.SetTanggalMeninggal, u.ClearTanggalMeninggal)
		}
		setOrClearUmatStr(req.NamaBaptis, u.SetNamaBaptis, u.ClearNamaBaptis)
		setOrClearUmatStr(req.TempatLahir, u.SetTempatLahir, u.ClearTempatLahir)
		setOrClearUmatStr(tanggalLahir, u.SetTanggalLahir, u.ClearTanggalLahir)
		setOrClearUmatStr(lookups.GolonganDarah, u.SetGolonganDarah, u.ClearGolonganDarah)
		setOrClearUmatStr(lookups.Suku, u.SetSuku, u.ClearSuku)
		setOrClearUmatStr(lookups.Agama, u.SetAgama, u.ClearAgama)
		setOrClearUmatStr(req.Kewarganegaraan, u.SetKewarganegaraan, u.ClearKewarganegaraan)
		setOrClearUmatStr(req.Negara, u.SetNegara, u.ClearNegara)
		setOrClearUmatStr(lookups.HubunganKeluarga, u.SetHubunganKeluarga, u.ClearHubunganKeluarga)
		setOrClearUmatStr(lookups.StatusPernikahan, u.SetStatusPernikahan, u.ClearStatusPernikahan)
		setOrClearUmatStr(req.Alamat, u.SetAlamat, u.ClearAlamat)
		setOrClearUmatStr(telepon, u.SetTelepon, u.ClearTelepon)
		setOrClearUmatStr(email, u.SetEmail, u.ClearEmail)
		setOrClearUmatStr(req.FotoURL, u.SetFotoURL, u.ClearFotoURL)
		setOrClearUmatStr(lookups.Pendidikan, u.SetPendidikan, u.ClearPendidikan)
		setOrClearUmatStr(req.Jurusan, u.SetJurusan, u.ClearJurusan)
		setOrClearUmatStr(lookups.Pekerjaan, u.SetPekerjaan, u.ClearPekerjaan)
		setOrClearUmatStr(lookups.Profesi, u.SetProfesi, u.ClearProfesi)
		setOrClearUmatStr(req.Keterampilan, u.SetKeterampilan, u.ClearKeterampilan)
		setOrClearUmatStr(lookups.Disabilitas, u.SetDisabilitas, u.ClearDisabilitas)
		setOrClearUmatStr(lookups.AktivitasSosial, u.SetAktivitasSosial, u.ClearAktivitasSosial)
		return u
	})
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.UmatDetailResponse{}, s.classifyUmatConstraintError(ctx, parokiID, nikHash, userID, nomorInduk, &id)
		}
		return domain.UmatDetailResponse{}, err
	}

	refetched, err := s.repo.FindUmatByID(ctx, parokiID, id)
	if err != nil {
		return domain.UmatDetailResponse{}, err
	}
	return s.toUmatDetail(refetched), nil
}

// DeleteUmat soft-deletes an Umat in parokiID if no active OrgJabatan references it.
func (s *Service) DeleteUmat(ctx context.Context, parokiID, id uuid.UUID) error {
	if _, err := s.repo.FindUmatByID(ctx, parokiID, id); ent.IsNotFound(err) {
		return domain.ErrUmatNotFound
	} else if err != nil {
		return err
	}
	inUse, err := s.repo.HasOrgJabatanByUmat(ctx, parokiID, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteUmat(ctx, id, time.Now().UTC())
}

// LookupUmat returns PII-free parishioner summaries for umat:admin, ref:paroki-admin, or active org:admin callers.
func (s *Service) LookupUmat(ctx context.Context, p authx.Principal, req domain.LookupUmatRequest) ([]domain.UmatLookupItemResponse, error) {
	allowed, err := s.canLookupUmat(ctx, p)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	rows, err := s.repo.LookupUmat(ctx, p.ParokiID, strings.TrimSpace(req.Q), req.WilayahID, req.LingkunganID, limit)
	if err != nil {
		return nil, err
	}
	items := make([]domain.UmatLookupItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUmatBasic(row))
	}
	return items, nil
}

// GetBasic returns a PII-free Umat summary by ID within parokiID.
func (s *Service) GetBasic(ctx context.Context, parokiID, umatID uuid.UUID) (*domain.UmatLookupItemResponse, error) {
	row, err := s.repo.FindUmatBasicByID(ctx, parokiID, umatID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrUmatNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toUmatBasic(row)
	return &res, nil
}

// GetByUserID returns a PII-free Umat summary linked to userID within parokiID, or (nil, nil) if none exists.
func (s *Service) GetByUserID(ctx context.Context, parokiID, userID uuid.UUID) (*domain.UmatLookupItemResponse, error) {
	row, err := s.repo.FindUmatByUserID(ctx, parokiID, userID)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	res := toUmatBasic(row)
	return &res, nil
}

func (s *Service) canLookupUmat(ctx context.Context, p authx.Principal) (bool, error) {
	if s.enforcer != nil {
		sub := authz.UserSubject(p.UserID)
		dom := authz.ParokiDomain(p.ParokiID)
		isUmatAdm, err := s.enforcer.HasRoleInDomain(sub, domain.RoleAdmin, dom)
		if err != nil {
			return false, err
		}
		if isUmatAdm {
			return true, nil
		}
		isParokiAdm, err := s.enforcer.HasRoleInDomain(sub, ref.RoleParokiAdmin, dom)
		if err != nil {
			return false, err
		}
		if isParokiAdm {
			return true, nil
		}
	}
	return s.repo.HasActiveOrgAdminInParoki(ctx, p.ParokiID, p.UserID)
}

func toUmatBasic(u *ent.Umat) domain.UmatLookupItemResponse {
	return domain.UmatLookupItemResponse{
		ID:           u.ID,
		ParokiID:     u.ParokiID,
		NomorInduk:   u.NomorInduk,
		Nama:         u.Nama,
		NamaBaptis:   u.NamaBaptis,
		WilayahID:    u.WilayahID,
		LingkunganID: u.LingkunganID,
		UserID:       u.UserID,
	}
}

func (s *Service) validateUserLink(ctx context.Context, parokiID, userID uuid.UUID, excludeUmatID *uuid.UUID) error {
	if s.refs != nil {
		ok, err := s.refs.HasParokiMember(ctx, parokiID, userID)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ValidationError("userId", "INVALID_REF", "User is not a member of active paroki.")
		}
	}
	linked, err := s.repo.HasActiveUserLink(ctx, parokiID, userID, excludeUmatID)
	if err != nil {
		return err
	}
	if linked {
		return domain.ErrUserAlreadyLinked
	}
	return nil
}

func (s *Service) classifyUmatConstraintError(ctx context.Context, parokiID uuid.UUID, nikHash *string, userID *uuid.UUID, nomorInduk *string, excludeID *uuid.UUID) error {
	if nomorInduk != nil {
		if ok, err := s.repo.HasActiveNomorInduk(ctx, parokiID, *nomorInduk, excludeID); err == nil && ok {
			return domain.ErrNomorIndukConflict
		}
	}
	if nikHash != nil {
		if ok, err := s.repo.HasActiveNIKHash(ctx, parokiID, *nikHash, excludeID); err == nil && ok {
			return domain.ErrNIKConflict
		}
	}
	if userID != nil {
		if ok, err := s.repo.HasActiveUserLink(ctx, parokiID, *userID, excludeID); err == nil && ok {
			return domain.ErrUserAlreadyLinked
		}
	}
	return domain.ErrNIKConflict
}

type umatLookupInputs struct {
	Agama            *string
	GolonganDarah    *string
	Suku             *string
	HubunganKeluarga *string
	StatusPernikahan *string
	Pendidikan       *string
	Pekerjaan        *string
	Profesi          *string
	Disabilitas      *string
	AktivitasSosial  *string
}

func (s *Service) validateUmatLookups(ctx context.Context, parokiID uuid.UUID, in umatLookupInputs) (umatLookupInputs, error) {
	var out umatLookupInputs
	var err error
	if out.Agama, err = s.validateLookupField(ctx, parokiID, "agama", "agama", in.Agama); err != nil {
		return out, err
	}
	if out.GolonganDarah, err = s.validateLookupField(ctx, parokiID, "golonganDarah", "golongan_darah", in.GolonganDarah); err != nil {
		return out, err
	}
	if out.Suku, err = s.validateLookupField(ctx, parokiID, "suku", "suku", in.Suku); err != nil {
		return out, err
	}
	if out.HubunganKeluarga, err = s.validateLookupField(ctx, parokiID, "hubunganKeluarga", "hubungan_keluarga", in.HubunganKeluarga); err != nil {
		return out, err
	}
	if out.StatusPernikahan, err = s.validateLookupField(ctx, parokiID, "statusPernikahan", "status_pernikahan", in.StatusPernikahan); err != nil {
		return out, err
	}
	if out.Pendidikan, err = s.validateLookupField(ctx, parokiID, "pendidikan", "pendidikan", in.Pendidikan); err != nil {
		return out, err
	}
	if out.Pekerjaan, err = s.validateLookupField(ctx, parokiID, "pekerjaan", "pekerjaan", in.Pekerjaan); err != nil {
		return out, err
	}
	if out.Profesi, err = s.validateLookupField(ctx, parokiID, "profesi", "profesi", in.Profesi); err != nil {
		return out, err
	}
	if out.Disabilitas, err = s.validateLookupField(ctx, parokiID, "disabilitas", "disabilitas", in.Disabilitas); err != nil {
		return out, err
	}
	if out.AktivitasSosial, err = s.validateLookupField(ctx, parokiID, "aktivitasSosial", "aktivitas_sosial", in.AktivitasSosial); err != nil {
		return out, err
	}
	return out, nil
}

func (s *Service) toUmatListItem(u *ent.Umat) domain.UmatListItemResponse {
	_, nikMasked := s.decryptAndMask(u.NikEnc)
	return domain.UmatListItemResponse{
		ID:               u.ID,
		ParokiID:         u.ParokiID,
		KeluargaID:       u.KeluargaID,
		WilayahID:        u.WilayahID,
		LingkunganID:     u.LingkunganID,
		UserID:           u.UserID,
		NomorInduk:       u.NomorInduk,
		Nama:             u.Nama,
		NamaBaptis:       u.NamaBaptis,
		NIKMasked:        nikMasked,
		JenisKelamin:     string(u.JenisKelamin),
		TempatLahir:      u.TempatLahir,
		TanggalLahir:     u.TanggalLahir,
		GolonganDarah:    u.GolonganDarah,
		Suku:             u.Suku,
		Agama:            u.Agama,
		Kewarganegaraan:  u.Kewarganegaraan,
		Negara:           u.Negara,
		HubunganKeluarga: u.HubunganKeluarga,
		StatusPernikahan: u.StatusPernikahan,
		Almarhum:         u.Almarhum,
		TanggalMeninggal: u.TanggalMeninggal,
		Alamat:           u.Alamat,
		KotaID:           u.KotaID,
		KecamatanID:      u.KecamatanID,
		KelurahanID:      u.KelurahanID,
		Telepon:          u.Telepon,
		Email:            u.Email,
		FotoURL:          u.FotoURL,
		Pendidikan:       u.Pendidikan,
		Jurusan:          u.Jurusan,
		Pekerjaan:        u.Pekerjaan,
		Profesi:          u.Profesi,
		Keterampilan:     u.Keterampilan,
		Disabilitas:      u.Disabilitas,
		AktivitasSosial:  u.AktivitasSosial,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}

func (s *Service) toUmatDetail(u *ent.Umat) domain.UmatDetailResponse {
	nikPlain, nikMasked := s.decryptAndMask(u.NikEnc)
	var kelSummary *domain.UmatKeluargaSummary
	if u.Edges.Keluarga != nil {
		_, kkMasked := s.decryptAndMask(u.Edges.Keluarga.NomorKkEnc)
		kelSummary = &domain.UmatKeluargaSummary{
			ID:              u.Edges.Keluarga.ID,
			NomorRegistrasi: u.Edges.Keluarga.NomorRegistrasi,
			NamaKeluarga:    u.Edges.Keluarga.NamaKeluarga,
			NomorKKMasked:   kkMasked,
		}
	}
	sakramen := make([]domain.UmatSakramenResponse, 0, len(u.Edges.Sakramen))
	for _, sk := range u.Edges.Sakramen {
		sakramen = append(sakramen, toSakramenResponse(sk))
	}
	slices.SortFunc(sakramen, func(a, b domain.UmatSakramenResponse) int {
		ra, rb := sakramenLifecycleRank(a.Jenis), sakramenLifecycleRank(b.Jenis)
		if ra != rb {
			if ra < rb {
				return -1
			}
			return 1
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	return domain.UmatDetailResponse{
		ID:               u.ID,
		ParokiID:         u.ParokiID,
		KeluargaID:       u.KeluargaID,
		WilayahID:        u.WilayahID,
		LingkunganID:     u.LingkunganID,
		UserID:           u.UserID,
		NomorInduk:       u.NomorInduk,
		Nama:             u.Nama,
		NamaBaptis:       u.NamaBaptis,
		NIK:              nikPlain,
		NIKMasked:        nikMasked,
		JenisKelamin:     string(u.JenisKelamin),
		TempatLahir:      u.TempatLahir,
		TanggalLahir:     u.TanggalLahir,
		GolonganDarah:    u.GolonganDarah,
		Suku:             u.Suku,
		Agama:            u.Agama,
		Kewarganegaraan:  u.Kewarganegaraan,
		Negara:           u.Negara,
		HubunganKeluarga: u.HubunganKeluarga,
		StatusPernikahan: u.StatusPernikahan,
		Almarhum:         u.Almarhum,
		TanggalMeninggal: u.TanggalMeninggal,
		Alamat:           u.Alamat,
		KotaID:           u.KotaID,
		KecamatanID:      u.KecamatanID,
		KelurahanID:      u.KelurahanID,
		Telepon:          u.Telepon,
		Email:            u.Email,
		FotoURL:          u.FotoURL,
		Pendidikan:       u.Pendidikan,
		Jurusan:          u.Jurusan,
		Pekerjaan:        u.Pekerjaan,
		Profesi:          u.Profesi,
		Keterampilan:     u.Keterampilan,
		Disabilitas:      u.Disabilitas,
		AktivitasSosial:  u.AktivitasSosial,
		Keluarga:         kelSummary,
		Sakramen:         sakramen,
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
	}
}

func parseOptDate(field string, raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	parsed, _, err := ParseDateYYYYMMDD(field, *raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func normalizeOptPhone(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	norm, err := NormalizePhone(*raw)
	if err != nil {
		return nil, err
	}
	return &norm, nil
}

func normalizeOptEmail(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	norm, err := NormalizeEmail(*raw)
	if err != nil {
		return nil, err
	}
	return &norm, nil
}

func setOrClearUmatStr(val *string, setFn func(string) *ent.UmatUpdateOne, clearFn func() *ent.UmatUpdateOne) {
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
