package umat_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
	refdomain "github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
)

func (s *UmatSuite) TestNomorGenerationAndConflict() {
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "KAJ-N",
		Nama: "Keuskupan Nomor",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "PRK-NOMOR",
		Nama:        "Paroki Nomor",
	})
	s.Created()
	paroki := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	umatTrue := true
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email:     "umat-admin@paroki.org",
		Name:      "Admin Umat",
		Role:      "member",
		UmatAdmin: &umatTrue,
	})
	s.Created()
	s.loginViaOTP("umat-admin@paroki.org")

	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")

	year := time.Now().UTC().Year()

	// Auto-generate KK-YYYY-000001 and KK-YYYY-000002
	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga: "Keluarga Satu",
	})
	s.Created()
	k1 := testutil.DecodeAs[domain.KeluargaDetailResponse](s).Data
	s.Equal(fmt.Sprintf("KK-%04d-000001", year), k1.NomorRegistrasi)

	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga: "Keluarga Dua",
	})
	s.Created()
	k2 := testutil.DecodeAs[domain.KeluargaDetailResponse](s).Data
	s.Equal(fmt.Sprintf("KK-%04d-000002", year), k2.NomorRegistrasi)

	// Client-supplied custom nomorRegistrasi accepted, duplicate returns 409 UMAT_REGISTRASI_CONFLICT
	customReg := "KK-CUSTOM-999"
	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga:    "Keluarga Custom",
		NomorRegistrasi: &customReg,
	})
	s.Created()

	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga:    "Keluarga Dup",
		NomorRegistrasi: &customReg,
	})
	s.Conflict()
	s.Equal("UMAT_REGISTRASI_CONFLICT", s.GetErrorResponse().Error.Code)

	// Auto-generate NIU-YYYY-000001 and custom conflict 409 UMAT_NOMOR_INDUK_CONFLICT
	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Umat Satu",
		JenisKelamin: "pria",
	})
	s.Created()
	u1 := testutil.DecodeAs[domain.UmatDetailResponse](s).Data
	s.Equal(fmt.Sprintf("NIU-%04d-000001", year), u1.NomorInduk)

	customNIU := "NIU-CUSTOM-999"
	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Umat Custom",
		JenisKelamin: "wanita",
		NomorInduk:   &customNIU,
	})
	s.Created()

	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Umat Dup",
		JenisKelamin: "wanita",
		NomorInduk:   &customNIU,
	})
	s.Conflict()
	s.Equal("UMAT_NOMOR_INDUK_CONFLICT", s.GetErrorResponse().Error.Code)
}

func (s *UmatSuite) TestDokumenAuthorizationMatrixAndCacheControl() {
	ctx := context.Background()
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "KAJ-D",
		Nama: "Keuskupan Dokumen",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "PRK-DOK",
		Nama:        "Paroki Dokumen",
	})
	s.Created()
	paroki := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	// Create 3 parish members: owner, head of same keluarga, unrelated user
	ownerUser, err := s.userFeature.EnsureUser(ctx, "owner@paroki.org", "Owner", "", "email", "owner@paroki.org")
	s.Require().NoError(err)
	headUser, err := s.userFeature.EnsureUser(ctx, "head@paroki.org", "Head", "", "email", "head@paroki.org")
	s.Require().NoError(err)
	otherUser, err := s.userFeature.EnsureUser(ctx, "other@paroki.org", "Other", "", "email", "other@paroki.org")
	s.Require().NoError(err)
	umatAdminUser, err := s.userFeature.EnsureUser(ctx, "umatadmin@paroki.org", "Umat Admin", "", "email", "umatadmin@paroki.org")
	s.Require().NoError(err)

	for _, uid := range []uuid.UUID{ownerUser.ID, headUser.ID, otherUser.ID, umatAdminUser.ID} {
		_, err = s.refFeature.EnsureParokiMember(ctx, paroki.ID, uid)
		s.Require().NoError(err)
	}
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(umatAdminUser.ID), "umat:admin", authz.ParokiDomain(paroki.ID)))

	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x02, 0x03}
	res, err := s.umatFeature.VerifyAndSyncCensus(ctx, user.VerifyConfirmInput{
		ParokiID:    paroki.ID,
		UserID:      ownerUser.ID,
		MemberIndex: 1,
		Payload: user.IdentityExtraction{
			KK: &user.ExtractedKK{
				NomorKK:      "3174010101900088",
				NamaKeluarga: "Keluarga Dok",
				Anggota: []user.ExtractedKKMember{
					{
						Nama:             "Head",
						NIK:              "3174010101900001",
						JenisKelamin:     "pria",
						HubunganKeluarga: strPtr("kepala-keluarga"),
					},
					{
						Nama:             "Owner",
						NIK:              "3174010101900002",
						JenisKelamin:     "wanita",
						HubunganKeluarga: strPtr("anak"),
					},
				},
			},
		},
		Documents: []user.VerifyConfirmDocument{
			{Jenis: "ktp", MimeType: "image/jpeg", Data: jpegBytes},
		},
	})
	s.Require().NoError(err)

	// Link headUser to the head umat row
	kel, err := s.client.Keluarga.Get(ctx, *res.KeluargaID)
	s.Require().NoError(err)
	headUmat, err := kel.QueryAnggota().Where().All(ctx)
	s.Require().NoError(err)
	for _, m := range headUmat {
		if m.Nama == "Head" {
			_, err = s.client.Umat.UpdateOneID(m.ID).SetUserID(headUser.ID).Save(ctx)
			s.Require().NoError(err)
		}
	}

	profile, err := s.umatFeature.GetSelfProfile(ctx, paroki.ID, ownerUser.ID)
	s.Require().NoError(err)
	s.Require().Len(profile.Dokumen, 1)
	docID := profile.Dokumen[0].ID

	// 1. Unrelated user -> 403 UMAT_FORBIDDEN
	s.loginViaOTP("other@paroki.org")
	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")
	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodGet, nil)
	s.Forbidden()
	s.Equal("UMAT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodDelete, nil)
	s.Forbidden()

	// 2. Owner -> 200 + Cache-Control: no-store, private + decrypted bytes
	s.loginViaOTP("owner@paroki.org")
	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodGet, nil)
	s.OK()
	s.Equal("no-store, private", s.HttpResponse.Header.Get("Cache-Control"))
	s.Equal("image/jpeg", s.HttpResponse.Header.Get("Content-Type"))
	body, err := io.ReadAll(s.HttpResponse.Body)
	s.Require().NoError(err)
	s.Equal(jpegBytes, body)

	// 3. Head of family -> 200
	s.loginViaOTP("head@paroki.org")
	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodGet, nil)
	s.OK()

	// 4. umat:admin -> 200 and DELETE -> 204
	s.loginViaOTP("umatadmin@paroki.org")
	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodGet, nil)
	s.OK()

	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/umat/dokumen/"+docID.String(), http.MethodGet, nil)
	s.NotFound()

	// 5. VerifyAndSyncCensus conflict when another user tries to claim an already-linked NIK
	preview, err := s.umatFeature.PreviewVerification(ctx, paroki.ID, otherUser.ID, user.IdentityExtraction{
		KTP: &user.ExtractedKTP{NIK: "3174010101900002", Nama: "Owner"},
	})
	s.Require().NoError(err)
	s.Equal("conflict", preview.MatchStatus)

	_, err = s.umatFeature.VerifyAndSyncCensus(ctx, user.VerifyConfirmInput{
		ParokiID: paroki.ID,
		UserID:   otherUser.ID,
		Payload: user.IdentityExtraction{
			KTP: &user.ExtractedKTP{
				NIK:          "3174010101900002",
				Nama:         "Owner",
				JenisKelamin: strPtr("wanita"),
			},
		},
	})
	s.Require().ErrorIs(err, domain.ErrUserAlreadyLinked)
}

func strPtr(v string) *string {
	return &v
}
