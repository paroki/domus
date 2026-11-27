package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/ent/umatdokumen"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/umat/repo"
	"github.com/paroki/domus/internal/user"
)

// PreviewVerification checks whether extracted identity data matches an existing parishioner or conflicts with another user.
func (s *Service) PreviewVerification(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	extraction user.IdentityExtraction,
) (user.VerifyPreviewResult, error) {
	var nik, nama, tanggalLahir string
	if extraction.KTP != nil {
		nik = strings.TrimSpace(extraction.KTP.NIK)
		nama = strings.TrimSpace(extraction.KTP.Nama)
		if extraction.KTP.TanggalLahir != nil {
			tanggalLahir = strings.TrimSpace(*extraction.KTP.TanggalLahir)
		}
	} else if extraction.KK != nil && len(extraction.KK.Anggota) == 1 {
		nik = strings.TrimSpace(extraction.KK.Anggota[0].NIK)
		nama = strings.TrimSpace(extraction.KK.Anggota[0].Nama)
		if extraction.KK.Anggota[0].TanggalLahir != nil {
			tanggalLahir = strings.TrimSpace(*extraction.KK.Anggota[0].TanggalLahir)
		}
	}

	matched, err := s.matchUmat(ctx, s.repo, parokiID, nik, nama, tanggalLahir)
	if err != nil {
		return user.VerifyPreviewResult{}, err
	}
	if matched != nil {
		if matched.UserID != nil && *matched.UserID != userID {
			return user.VerifyPreviewResult{MatchStatus: "conflict"}, nil
		}
		id := matched.ID
		return user.VerifyPreviewResult{MatchStatus: "matched", MatchedUmatID: &id}, nil
	}

	existing, err := s.repo.FindUmatByUserID(ctx, parokiID, userID)
	if err == nil && existing != nil {
		id := existing.ID
		return user.VerifyPreviewResult{MatchStatus: "matched", MatchedUmatID: &id}, nil
	}
	if err != nil && !ent.IsNotFound(err) {
		return user.VerifyPreviewResult{}, err
	}
	return user.VerifyPreviewResult{MatchStatus: "new"}, nil
}

// VerifyAndSyncCensus re-validates the reviewed identity payload, upserts keluarga/umat/dokumen in one transaction, and links the caller.
func (s *Service) VerifyAndSyncCensus(ctx context.Context, in user.VerifyConfirmInput) (user.VerifyConfirmResult, error) {
	if s.cipher == nil {
		return user.VerifyConfirmResult{}, domain.ValidationError("cipher", "CIPHER_UNAVAILABLE", "Encryption cipher is not configured.")
	}
	if in.Payload.KTP == nil && (in.Payload.KK == nil || len(in.Payload.KK.Anggota) == 0) {
		return user.VerifyConfirmResult{}, domain.ValidationError("payload", "REQUIRED", "KTP or KK data is required.")
	}
	if in.Payload.KK != nil && len(in.Payload.KK.Anggota) > 0 {
		if in.MemberIndex < 0 || in.MemberIndex >= len(in.Payload.KK.Anggota) {
			return user.VerifyConfirmResult{}, domain.ValidationError("memberIndex", "INVALID_INDEX", "memberIndex is out of range for KK anggota.")
		}
	}

	type preparedDoc struct {
		jenis    umatdokumen.Jenis
		mimeType umatdokumen.MimeType
		size     int
		dataEnc  []byte
	}
	docs := make([]preparedDoc, 0, len(in.Documents))
	for _, d := range in.Documents {
		jenis := umatdokumen.Jenis(d.Jenis)
		switch jenis {
		case umatdokumen.JenisKtp, umatdokumen.JenisKia, umatdokumen.JenisKartuPelajar, umatdokumen.JenisKk:
		default:
			return user.VerifyConfirmResult{}, domain.ValidationError("dokumen", "INVALID_JENIS", "Unsupported document type.")
		}
		mimeType, enc, err := s.validateAndEncryptDokumen(d.Jenis, d.Data)
		if err != nil {
			return user.VerifyConfirmResult{}, err
		}
		docs = append(docs, preparedDoc{
			jenis:    jenis,
			mimeType: mimeType,
			size:     len(d.Data),
			dataEnc:  enc,
		})
	}

	wilayahID, lingkunganID, _, _, err := s.resolveTerritorial(ctx, in.ParokiID, in.WilayahID, in.LingkunganID, nil, nil)
	if err != nil {
		return user.VerifyConfirmResult{}, err
	}

	var result user.VerifyConfirmResult
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		var keluargaRow *ent.Keluarga
		if in.Payload.KK != nil {
			var txErr error
			keluargaRow, txErr = s.upsertVerifyKeluarga(ctx, txRepo, in.ParokiID, in.Payload.KK, wilayahID, lingkunganID)
			if txErr != nil {
				return txErr
			}
		}

		var callerUmat *ent.Umat
		if in.Payload.KK != nil && len(in.Payload.KK.Anggota) > 0 {
			for i, m := range in.Payload.KK.Anggota {
				isCaller := i == in.MemberIndex
				var ktpOverlay *user.ExtractedKTP
				if isCaller {
					ktpOverlay = in.Payload.KTP
				}
				var kelID *uuid.UUID
				if keluargaRow != nil {
					kelID = &keluargaRow.ID
				}
				uRow, uErr := s.upsertVerifyMember(ctx, txRepo, in.ParokiID, in.UserID, isCaller, m, ktpOverlay, kelID, wilayahID, lingkunganID)
				if uErr != nil {
					return uErr
				}
				if isCaller {
					callerUmat = uRow
				}
			}
		} else if in.Payload.KTP != nil {
			m := user.ExtractedKKMember{
				Nama:             in.Payload.KTP.Nama,
				NIK:              in.Payload.KTP.NIK,
				TempatLahir:      in.Payload.KTP.TempatLahir,
				TanggalLahir:     in.Payload.KTP.TanggalLahir,
				Agama:            in.Payload.KTP.Agama,
				Pekerjaan:        in.Payload.KTP.Pekerjaan,
				StatusPernikahan: in.Payload.KTP.StatusPernikahan,
			}
			if in.Payload.KTP.JenisKelamin != nil {
				m.JenisKelamin = *in.Payload.KTP.JenisKelamin
			}
			var kelID *uuid.UUID
			if keluargaRow != nil {
				kelID = &keluargaRow.ID
			}
			uRow, uErr := s.upsertVerifyMember(ctx, txRepo, in.ParokiID, in.UserID, true, m, in.Payload.KTP, kelID, wilayahID, lingkunganID)
			if uErr != nil {
				return uErr
			}
			callerUmat = uRow
		}

		if callerUmat == nil {
			return domain.ValidationError("payload", "REQUIRED", "Caller umat record could not be resolved.")
		}

		for _, d := range docs {
			var uID, kID *uuid.UUID
			if d.jenis == umatdokumen.JenisKk {
				if keluargaRow == nil {
					continue
				}
				kID = &keluargaRow.ID
			} else {
				uID = &callerUmat.ID
			}
			if _, err := txRepo.UpsertDokumen(ctx, in.ParokiID, uID, kID, d.jenis, d.mimeType, d.size, d.dataEnc, in.UserID); err != nil {
				return err
			}
		}

		result = user.VerifyConfirmResult{
			UmatID:       callerUmat.ID,
			NomorInduk:   callerUmat.NomorInduk,
			KeluargaID:   callerUmat.KeluargaID,
			Nama:         callerUmat.Nama,
			JenisKelamin: string(callerUmat.JenisKelamin),
			TanggalLahir: callerUmat.TanggalLahir,
		}
		return nil
	})
	if err != nil {
		return user.VerifyConfirmResult{}, err
	}
	return result, nil
}

// GetSelfProfile returns the caller's linked umat, keluarga (with decrypted NIK/nomorKk), members, and document metadata.
func (s *Service) GetSelfProfile(ctx context.Context, parokiID, userID uuid.UUID) (*user.SelfUmatProfile, error) {
	u, err := s.repo.FindUmatByUserID(ctx, parokiID, userID)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	selfInfo := s.toSelfUmatInfo(u)
	profile := &user.SelfUmatProfile{
		Umat:    &selfInfo,
		Members: []user.SelfUmatInfo{},
		Dokumen: []user.SelfDokumenMeta{},
	}

	isHead := u.HubunganKeluarga != nil && strings.EqualFold(*u.HubunganKeluarga, "kepala-keluarga")
	var memberIDs []uuid.UUID
	memberIDs = append(memberIDs, u.ID)

	if u.KeluargaID != nil {
		kel, err := s.repo.FindKeluargaByID(ctx, parokiID, *u.KeluargaID)
		if err == nil && kel != nil {
			kkPlain, _ := s.decryptAndMask(kel.NomorKkEnc)
			profile.Keluarga = &user.SelfKeluargaInfo{
				ID:              kel.ID,
				NomorRegistrasi: kel.NomorRegistrasi,
				NomorKK:         kkPlain,
				NamaKeluarga:    kel.NamaKeluarga,
				Alamat:          kel.Alamat,
				WilayahID:       kel.WilayahID,
				LingkunganID:    kel.LingkunganID,
			}
			for _, a := range kel.Edges.Anggota {
				profile.Members = append(profile.Members, s.toSelfUmatInfo(a))
				if isHead && a.ID != u.ID {
					memberIDs = append(memberIDs, a.ID)
				}
			}
		} else if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
	}

	seenDocs := make(map[uuid.UUID]struct{})
	umatDocs, err := s.repo.ListDokumenByUmatIDs(ctx, parokiID, memberIDs)
	if err != nil {
		return nil, err
	}
	for _, d := range umatDocs {
		seenDocs[d.ID] = struct{}{}
		profile.Dokumen = append(profile.Dokumen, toSelfDokumenMeta(d))
	}

	if isHead && u.KeluargaID != nil {
		kelDocs, err := s.repo.ListDokumenByKeluargaID(ctx, parokiID, *u.KeluargaID)
		if err != nil {
			return nil, err
		}
		for _, d := range kelDocs {
			if _, ok := seenDocs[d.ID]; ok {
				continue
			}
			profile.Dokumen = append(profile.Dokumen, toSelfDokumenMeta(d))
		}
	}

	return profile, nil
}

func (s *Service) matchUmat(ctx context.Context, r *repo.Repository, parokiID uuid.UUID, nik, nama, tanggalLahir string) (*ent.Umat, error) {
	if cleanNIK := strings.TrimSpace(nik); cleanNIK != "" && s.cipher != nil {
		row, err := r.FindUmatByNIKHash(ctx, parokiID, s.cipher.BlindIndex(cleanNIK))
		if err == nil {
			return row, nil
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
	}
	if strings.TrimSpace(nama) != "" && strings.TrimSpace(tanggalLahir) != "" {
		row, err := r.FindUmatByNameAndBirthDate(ctx, parokiID, nama, tanggalLahir)
		if err == nil {
			return row, nil
		}
		if !ent.IsNotFound(err) {
			return nil, err
		}
	}
	return nil, nil
}

func (s *Service) upsertVerifyKeluarga(
	ctx context.Context,
	txRepo *repo.Repository,
	parokiID uuid.UUID,
	kk *user.ExtractedKK,
	wilayahID, lingkunganID *uuid.UUID,
) (*ent.Keluarga, error) {
	cleanKK, err := ValidatePII("nomorKk", kk.NomorKK)
	if err != nil {
		return nil, err
	}
	if cleanKK == "" {
		return nil, domain.ValidationError("nomorKk", "REQUIRED", "Nomor KK is required.")
	}
	namaKeluarga := strings.TrimSpace(kk.NamaKeluarga)
	if namaKeluarga == "" {
		return nil, domain.ValidationError("namaKeluarga", "REQUIRED", "Nama keluarga is required.")
	}
	kkHash := s.cipher.BlindIndex(cleanKK)
	kkEnc, err := s.cipher.EncryptString(cleanKK)
	if err != nil {
		return nil, err
	}

	existing, err := txRepo.FindKeluargaByNomorKKHash(ctx, parokiID, kkHash)
	if err == nil && existing != nil {
		return txRepo.UpdateKeluarga(ctx, existing.ID, func(u *ent.KeluargaUpdateOne) *ent.KeluargaUpdateOne {
			u.SetNamaKeluarga(namaKeluarga).
				SetNomorKkEnc(kkEnc).
				SetNomorKkHash(kkHash)
			if wilayahID != nil {
				u.SetWilayahID(*wilayahID)
			}
			if lingkunganID != nil {
				u.SetLingkunganID(*lingkunganID)
			}
			if kk.Alamat != nil && strings.TrimSpace(*kk.Alamat) != "" {
				u.SetAlamat(strings.TrimSpace(*kk.Alamat))
			}
			return u
		})
	}
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	nomorReg, err := txRepo.NextNomorRegistrasi(ctx, parokiID, time.Now().UTC().Year())
	if err != nil {
		return nil, err
	}
	return txRepo.CreateKeluarga(ctx, func(c *ent.KeluargaCreate) *ent.KeluargaCreate {
		c.SetParokiID(parokiID).
			SetNomorRegistrasi(nomorReg).
			SetNamaKeluarga(namaKeluarga).
			SetNomorKkEnc(kkEnc).
			SetNomorKkHash(kkHash).
			SetAktif(true).
			SetNillableWilayahID(wilayahID).
			SetNillableLingkunganID(lingkunganID).
			SetNillableAlamat(trimOptString(kk.Alamat))
		return c
	})
}

func (s *Service) upsertVerifyMember(
	ctx context.Context,
	txRepo *repo.Repository,
	parokiID, userID uuid.UUID,
	isCaller bool,
	m user.ExtractedKKMember,
	ktp *user.ExtractedKTP,
	keluargaID, wilayahID, lingkunganID *uuid.UUID,
) (*ent.Umat, error) {
	nama := strings.TrimSpace(m.Nama)
	nikRaw := strings.TrimSpace(m.NIK)
	jenisKelaminRaw := strings.TrimSpace(m.JenisKelamin)
	tempatLahir := m.TempatLahir
	tanggalLahir := m.TanggalLahir
	agama := m.Agama
	pekerjaan := m.Pekerjaan
	statusPernikahan := m.StatusPernikahan
	hubunganKeluarga := m.HubunganKeluarga
	var alamat, golonganDarah *string

	if ktp != nil {
		if n := strings.TrimSpace(ktp.Nama); n != "" {
			nama = n
		}
		if n := strings.TrimSpace(ktp.NIK); n != "" {
			nikRaw = n
		}
		if ktp.JenisKelamin != nil && strings.TrimSpace(*ktp.JenisKelamin) != "" {
			jenisKelaminRaw = strings.TrimSpace(*ktp.JenisKelamin)
		}
		if ktp.TempatLahir != nil && strings.TrimSpace(*ktp.TempatLahir) != "" {
			tempatLahir = ktp.TempatLahir
		}
		if ktp.TanggalLahir != nil && strings.TrimSpace(*ktp.TanggalLahir) != "" {
			tanggalLahir = ktp.TanggalLahir
		}
		if ktp.Agama != nil && strings.TrimSpace(*ktp.Agama) != "" {
			agama = ktp.Agama
		}
		if ktp.Pekerjaan != nil && strings.TrimSpace(*ktp.Pekerjaan) != "" {
			pekerjaan = ktp.Pekerjaan
		}
		if ktp.StatusPernikahan != nil && strings.TrimSpace(*ktp.StatusPernikahan) != "" {
			statusPernikahan = ktp.StatusPernikahan
		}
		if ktp.Alamat != nil && strings.TrimSpace(*ktp.Alamat) != "" {
			alamat = ktp.Alamat
		}
		if ktp.GolonganDarah != nil && strings.TrimSpace(*ktp.GolonganDarah) != "" {
			golonganDarah = ktp.GolonganDarah
		}
	}

	if nama == "" {
		return nil, domain.ValidationError("nama", "REQUIRED", "Nama is required.")
	}
	jk, err := normalizeVerifyGender(jenisKelaminRaw)
	if err != nil {
		return nil, err
	}

	var parsedBirth *string
	if tanggalLahir != nil && strings.TrimSpace(*tanggalLahir) != "" {
		p, _, err := ParseDateYYYYMMDD("tanggalLahir", *tanggalLahir)
		if err != nil {
			return nil, err
		}
		parsedBirth = &p
	}

	var nikEnc, nikHash *string
	if nikRaw != "" {
		cleanNIK, err := ValidatePII("nik", nikRaw)
		if err != nil {
			return nil, err
		}
		if cleanNIK != "" {
			hash := s.cipher.BlindIndex(cleanNIK)
			enc, err := s.cipher.EncryptString(cleanNIK)
			if err != nil {
				return nil, err
			}
			nikHash = &hash
			nikEnc = &enc
			nikRaw = cleanNIK
		}
	}

	var birthVal string
	if parsedBirth != nil {
		birthVal = *parsedBirth
	}
	matched, err := s.matchUmat(ctx, txRepo, parokiID, nikRaw, nama, birthVal)
	if err != nil {
		return nil, err
	}
	if matched == nil && isCaller {
		existingByUser, uerr := txRepo.FindUmatByUserID(ctx, parokiID, userID)
		if uerr == nil {
			matched = existingByUser
		} else if !ent.IsNotFound(uerr) {
			return nil, uerr
		}
	}

	if matched != nil {
		if isCaller {
			if matched.UserID != nil && *matched.UserID != userID {
				return nil, domain.ErrUserAlreadyLinked
			}
			linkedOther, lerr := txRepo.HasActiveUserLink(ctx, parokiID, userID, &matched.ID)
			if lerr != nil {
				return nil, lerr
			}
			if linkedOther {
				return nil, domain.ErrUserAlreadyLinked
			}
		}
		return txRepo.UpdateUmat(ctx, matched.ID, func(u *ent.UmatUpdateOne) *ent.UmatUpdateOne {
			u.SetNama(nama).SetJenisKelamin(jk)
			if isCaller {
				u.SetUserID(userID)
			}
			if keluargaID != nil {
				u.SetKeluargaID(*keluargaID)
			}
			if wilayahID != nil {
				u.SetWilayahID(*wilayahID)
			}
			if lingkunganID != nil {
				u.SetLingkunganID(*lingkunganID)
			}
			if nikEnc != nil && nikHash != nil {
				u.SetNikEnc(*nikEnc).SetNikHash(*nikHash)
			}
			if v := trimOptString(tempatLahir); v != nil {
				u.SetTempatLahir(*v)
			}
			if parsedBirth != nil {
				u.SetTanggalLahir(*parsedBirth)
			}
			if v := trimOptString(agama); v != nil {
				u.SetAgama(strings.ToLower(*v))
			}
			if v := trimOptString(golonganDarah); v != nil {
				u.SetGolonganDarah(*v)
			}
			if v := trimOptString(hubunganKeluarga); v != nil {
				u.SetHubunganKeluarga(strings.ToLower(*v))
			}
			if v := trimOptString(statusPernikahan); v != nil {
				u.SetStatusPernikahan(strings.ToLower(*v))
			}
			if v := trimOptString(pekerjaan); v != nil {
				u.SetPekerjaan(strings.ToLower(*v))
			}
			if v := trimOptString(alamat); v != nil {
				u.SetAlamat(*v)
			}
			return u
		})
	}

	if isCaller {
		linkedOther, lerr := txRepo.HasActiveUserLink(ctx, parokiID, userID, nil)
		if lerr != nil {
			return nil, lerr
		}
		if linkedOther {
			return nil, domain.ErrUserAlreadyLinked
		}
	}

	nomorInduk, err := txRepo.NextNomorInduk(ctx, parokiID, time.Now().UTC().Year())
	if err != nil {
		return nil, err
	}
	return txRepo.CreateUmat(ctx, func(c *ent.UmatCreate) *ent.UmatCreate {
		c.SetParokiID(parokiID).
			SetNomorInduk(nomorInduk).
			SetNama(nama).
			SetJenisKelamin(jk).
			SetAlmarhum(false).
			SetNillableKeluargaID(keluargaID).
			SetNillableWilayahID(wilayahID).
			SetNillableLingkunganID(lingkunganID).
			SetNillableNikEnc(nikEnc).
			SetNillableNikHash(nikHash).
			SetNillableTempatLahir(trimOptString(tempatLahir)).
			SetNillableTanggalLahir(parsedBirth).
			SetNillableAgama(lowerOptString(agama)).
			SetNillableGolonganDarah(trimOptString(golonganDarah)).
			SetNillableHubunganKeluarga(lowerOptString(hubunganKeluarga)).
			SetNillableStatusPernikahan(lowerOptString(statusPernikahan)).
			SetNillablePekerjaan(lowerOptString(pekerjaan)).
			SetNillableAlamat(trimOptString(alamat))
		if isCaller {
			c.SetUserID(userID)
		}
		return c
	})
}

func normalizeVerifyGender(raw string) (umat.JenisKelamin, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pria", "laki-laki", "laki laki", "l":
		return umat.JenisKelaminPria, nil
	case "wanita", "perempuan", "p":
		return umat.JenisKelaminWanita, nil
	default:
		return "", domain.ValidationError("jenisKelamin", "INVALID_GENDER", "jenisKelamin must be pria or wanita.")
	}
}

func lowerOptString(val *string) *string {
	t := trimOptString(val)
	if t == nil {
		return nil
	}
	l := strings.ToLower(*t)
	return &l
}

func (s *Service) toSelfUmatInfo(u *ent.Umat) user.SelfUmatInfo {
	nikPlain, _ := s.decryptAndMask(u.NikEnc)
	return user.SelfUmatInfo{
		ID:               u.ID,
		NomorInduk:       u.NomorInduk,
		Nama:             u.Nama,
		NIK:              nikPlain,
		JenisKelamin:     string(u.JenisKelamin),
		TempatLahir:      u.TempatLahir,
		TanggalLahir:     u.TanggalLahir,
		Agama:            u.Agama,
		GolonganDarah:    u.GolonganDarah,
		HubunganKeluarga: u.HubunganKeluarga,
		StatusPernikahan: u.StatusPernikahan,
		Pekerjaan:        u.Pekerjaan,
		Alamat:           u.Alamat,
		WilayahID:        u.WilayahID,
		LingkunganID:     u.LingkunganID,
		UserID:           u.UserID,
	}
}

func toSelfDokumenMeta(d *ent.UmatDokumen) user.SelfDokumenMeta {
	return user.SelfDokumenMeta{
		ID:         d.ID,
		UmatID:     d.UmatID,
		KeluargaID: d.KeluargaID,
		Jenis:      string(d.Jenis),
		MimeType:   string(d.MimeType),
		SizeBytes:  d.SizeBytes,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
}
