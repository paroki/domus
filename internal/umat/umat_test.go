package umat_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	entumat "github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/ref"
	refdomain "github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/umat"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/user"
	userdomain "github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type stubOTPSender struct {
	lastCode string
}

func (s *stubOTPSender) SendOTP(_ context.Context, _, _, code string) error {
	s.lastCode = code
	return nil
}

type UmatSuite struct {
	testutil.ApiTestSuite[any]
	client      *ent.Client
	enforcer    *authz.Enforcer
	sender      *stubOTPSender
	userFeature *user.Feature
	refFeature  *ref.Feature
	umatFeature *umat.Feature
}

func TestUmatSuite(t *testing.T) {
	suite.Run(t, new(UmatSuite))
}

func (s *UmatSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.sender = &stubOTPSender{}

	cipher, err := crypt.New("test-secret-key-at-least-32-bytes!!")
	s.Require().NoError(err)

	userCfg := user.Config{
		SuperAdminEmail: "admin@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     60 * time.Second,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.userFeature = user.New(s.client, s.enforcer, userCfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.sender))
	s.refFeature = ref.New(s.client, s.enforcer, s.userFeature)
	s.umatFeature = umat.New(s.client, s.enforcer, s.userFeature, s.refFeature, cipher)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature, s.umatFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature, s.umatFeature)
}

func (s *UmatSuite) loginViaOTP(email string) {
	s.ClearCookies()
	s.Request("/api/auth/otp", http.MethodPost, userdomain.RequestOTPRequest{Email: email})
	s.OK()
	s.Request("/api/auth/otp/verify", http.MethodPost, userdomain.VerifyOTPRequest{
		Email: email,
		Code:  s.sender.lastCode,
	})
	s.OK()
	s.Cookies = s.HttpResponse.Cookies()
}

func (s *UmatSuite) TestPartialUniqueIndexes_IgnoreSoftDeletedAndNulls() {
	ctx := context.Background()
	parokiID := uuid.New()
	kkHash := "kk-hash-001"
	nikHash := "nik-hash-001"
	userID := uuid.New()

	// Multiple rows with NULL hashes/user_id are allowed in the same paroki.
	kNull1, err := s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNomorRegistrasi("KK-2026-000001").
		SetNamaKeluarga("Keluarga Tanpa KK 1").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNomorRegistrasi("KK-2026-000002").
		SetNamaKeluarga("Keluarga Tanpa KK 2").
		Save(ctx)
	s.Require().NoError(err)

	// Active duplicate nomor_kk_hash in same paroki fails; soft-deleted allows reuse.
	k1, err := s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNomorRegistrasi("KK-2026-000003").
		SetNamaKeluarga("Keluarga A").
		SetNomorKkHash(kkHash).
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNomorRegistrasi("KK-2026-000004").
		SetNamaKeluarga("Keluarga A Dup").
		SetNomorKkHash(kkHash).
		Save(ctx)
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))

	s.Require().NoError(s.client.Keluarga.UpdateOneID(k1.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))

	k2, err := s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNomorRegistrasi("KK-2026-000005").
		SetNamaKeluarga("Keluarga A Reused").
		SetNomorKkHash(kkHash).
		Save(ctx)
	s.Require().NoError(err)
	s.NotEqual(k1.ID, k2.ID)

	// Multiple Umat with NULL nik_hash & user_id are allowed.
	_, err = s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNomorInduk("NIU-2026-000001").
		SetKeluargaID(kNull1.ID).
		SetNama("Umat Null 1").
		SetJenisKelamin(entumat.JenisKelaminPria).
		Save(ctx)
	s.Require().NoError(err)

	// Active duplicate nik_hash and user_id fail; soft-deleting allows reuse.
	u1, err := s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNomorInduk("NIU-2026-000002").
		SetKeluargaID(k2.ID).
		SetNama("Budi").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetNikHash(nikHash).
		SetUserID(userID).
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNomorInduk("NIU-2026-000003").
		SetNama("Budi Dup NIK").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetNikHash(nikHash).
		Save(ctx)
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))

	_, err = s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNomorInduk("NIU-2026-000004").
		SetNama("Budi Dup User").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetUserID(userID).
		Save(ctx)
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))

	s.Require().NoError(s.client.Umat.UpdateOneID(u1.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))

	u2, err := s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNomorInduk("NIU-2026-000005").
		SetKeluargaID(k2.ID).
		SetNama("Budi Reused").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetNikHash(nikHash).
		SetUserID(userID).
		Save(ctx)
	s.Require().NoError(err)
	s.NotEqual(u1.ID, u2.ID)
}

func (s *UmatSuite) TestUmatAdminRBAC_LeastPrivilegeAndGetMeEnrichment() {
	// RoleDescriptor registered with paroki scope.
	var found bool
	for _, r := range s.enforcer.Roles() {
		if r.Code == umat.RoleAdmin {
			found = true
			s.Equal(authz.RoleScopeParoki, r.Scope)
		}
	}
	s.True(found)

	// Least privilege: super-admin and paroki-admin do NOT have implicit umat permissions.
	parokiID := uuid.New()
	dom := authz.ParokiDomain(parokiID)
	superSub := authz.UserSubject(uuid.New())
	parokiAdminSub := authz.UserSubject(uuid.New())
	umatAdminSub := authz.UserSubject(uuid.New())
	ctx := context.Background()

	s.Require().NoError(s.enforcer.AssignRole(ctx, superSub, user.RoleSuperAdmin, authz.DomainGlobal))
	s.Require().NoError(s.enforcer.AssignRole(ctx, parokiAdminSub, ref.RoleParokiAdmin, dom))
	s.Require().NoError(s.enforcer.AssignRole(ctx, umatAdminSub, umat.RoleAdmin, dom))

	for _, perm := range []string{umat.PermKeluarga, umat.PermUmat} {
		for _, act := range []authz.Action{authz.ActionRead, authz.ActionWrite} {
			ok, err := s.enforcer.Enforce(superSub, dom, perm, act)
			s.Require().NoError(err)
			s.False(ok, "super-admin should not implicitly have %s:%s", perm, act)

			ok, err = s.enforcer.Enforce(parokiAdminSub, dom, perm, act)
			s.Require().NoError(err)
			s.False(ok, "paroki-admin should not implicitly have %s:%s", perm, act)

			ok, err = s.enforcer.Enforce(umatAdminSub, dom, perm, act)
			s.Require().NoError(err)
			s.True(ok, "umat-admin should have %s:%s", perm, act)
		}
	}

	// GET /api/me enriches umat:admin when assigned via ParokiMember.
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "Keuskupan Agung Jakarta",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "katedral",
		Nama:        "Paroki Katedral",
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
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[userdomain.UserResponse](s).Data
	s.Contains(me.Roles, umat.RoleAdmin)
	s.NotContains(me.Roles, ref.RoleParokiAdmin)
}

func (s *UmatSuite) TestKeluargaAndUmatCRUD_EncryptionMaskingAndValidations() {
	// Setup keuskupan, kota, kecamatan, kelurahan, paroki, wilayah, lingkungan, lookup, umat:admin & paroki:admin.
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{Kode: "kaj-u", Nama: "KAJ Umat"})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/kota", http.MethodPost, refdomain.CreateKotaRequest{Kode: "jkt-u", Nama: "Jakarta Umat"})
	s.Created()
	kota := testutil.DecodeAs[refdomain.KotaResponse](s).Data

	s.Request("/api/ref/kecamatan", http.MethodPost, refdomain.CreateKecamatanRequest{KotaID: kota.ID, Kode: "kec-u", Nama: "Kecamatan Umat"})
	s.Created()
	kec := testutil.DecodeAs[refdomain.KecamatanResponse](s).Data

	s.Request("/api/ref/kelurahan", http.MethodPost, refdomain.CreateKelurahanRequest{KecamatanID: kec.ID, Kode: "kel-u", Nama: "Kelurahan Umat"})
	s.Created()
	kel := testutil.DecodeAs[refdomain.KelurahanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "paroki-u",
		Nama:        "Paroki Umat",
	})
	s.Created()
	paroki := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	// Add paroki-admin (without umat:admin) and umat-admin (without paroki:admin) + regular member.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Paroki Admin",
		Role:  "admin",
	})
	s.Created()

	umatTrue := true
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email:     "uadmin@paroki.org",
		Name:      "Umat Admin",
		Role:      "member",
		UmatAdmin: &umatTrue,
	})
	s.Created()

	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "member@paroki.org",
		Name:  "Member Umat",
		Role:  "member",
	})
	s.Created()
	member := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data

	// Create wilayah, lingkungan, and active lookups as paroki-admin.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")

	s.Request("/api/ref/wilayah", http.MethodPost, refdomain.CreateWilayahRequest{Jenis: "wilayah", Kode: "w1", Nama: "Wilayah 1"})
	s.Created()
	wil := testutil.DecodeAs[refdomain.WilayahResponse](s).Data

	s.Request("/api/ref/lingkungan", http.MethodPost, refdomain.CreateLingkunganRequest{WilayahID: wil.ID, Kode: "l1", Nama: "Lingkungan 1"})
	s.Created()
	ling := testutil.DecodeAs[refdomain.LingkunganResponse](s).Data

	// Paroki-admin WITHOUT umat:admin -> 403 on /api/umat/keluarga/search.
	s.Request("/api/umat/keluarga/search", http.MethodPost, domain.SearchKeluargaRequest{})
	s.Forbidden()

	// User without paroki membership and without X-Paroki-Id -> 400 TENANT_REQUIRED.
	_, err := s.userFeature.EnsureUser(context.Background(), "notenant@paroki.org", "No Tenant", "", "email", "notenant@paroki.org")
	s.Require().NoError(err)
	s.loginViaOTP("notenant@paroki.org")
	s.SetHeader("X-Paroki-Id", "")
	s.Request("/api/umat/search", http.MethodPost, domain.SearchUmatRequest{})
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Super-admin WITHOUT umat:admin -> 403 on /api/umat/search (both auto-resolved and explicit X-Paroki-Id).
	s.loginViaOTP("admin@paroki.org")
	s.SetHeader("X-Paroki-Id", "")
	s.Request("/api/umat/search", http.MethodPost, domain.SearchUmatRequest{})
	s.Forbidden()

	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	s.Request("/api/umat/search", http.MethodPost, domain.SearchUmatRequest{})
	s.Forbidden()

	// Login as umat:admin and create active umat referensi entries.
	s.loginViaOTP("uadmin@paroki.org")
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{Category: "status_rumah", Code: "milik_sendiri", Label: "Milik Sendiri"})
	s.Created()
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{Category: "agama", Code: "katolik", Label: "Katolik"})
	s.Created()

	// Create Keluarga with auto-resolved wilayahId from lingkunganId, encrypted nomorKk, civil territory, and lookup.
	kk1 := "  3174010101900001  "
	statusRumah := "MILIK_SENDIRI"
	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		LingkunganID: &ling.ID,
		NomorKK:      &kk1,
		NamaKeluarga: "Keluarga Santoso",
		KotaID:       &kota.ID,
		KecamatanID:  &kec.ID,
		KelurahanID:  &kel.ID,
		StatusRumah:  &statusRumah,
	})
	s.Created()
	kel1 := testutil.DecodeAs[domain.KeluargaDetailResponse](s).Data
	s.Require().NotNil(kel1.WilayahID)
	s.Equal(wil.ID, *kel1.WilayahID)
	s.Require().NotNil(kel1.NomorKK)
	s.Equal("3174010101900001", *kel1.NomorKK)
	s.Require().NotNil(kel1.NomorKKMasked)
	s.Equal("************0001", *kel1.NomorKKMasked)
	s.Require().NotNil(kel1.StatusRumah)
	s.Equal("milik_sendiri", *kel1.StatusRumah)
	s.Empty(kel1.Anggota)

	// Duplicate nomorKk in same paroki -> 409 UMAT_KK_CONFLICT.
	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga: "Keluarga Dup",
		NomorKK:      &kk1,
	})
	s.Conflict()
	s.Equal("UMAT_KK_CONFLICT", s.GetErrorResponse().Error.Code)

	// Invalid lookup -> 422 VALIDATION_ERROR.
	badLookup := "tidak_ada"
	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga: "Keluarga Bad Lookup",
		StatusRumah:  &badLookup,
	})
	s.UnprocessableEntity()

	// Search Keluarga by nomorKk blind index -> returns masked only (never plaintext nomorKk).
	s.Request("/api/umat/keluarga/search", http.MethodPost, domain.SearchKeluargaRequest{
		Filter: &domain.KeluargaSearchFilter{NomorKK: "3174010101900001"},
	})
	s.OK()
	kelList := testutil.DecodeAs[[]domain.KeluargaListItemResponse](s).Data
	s.Require().Len(kelList, 1)
	s.Require().NotNil(kelList[0].NomorKKMasked)
	s.Equal("************0001", *kelList[0].NomorKKMasked)
	// Verify raw JSON does not contain plaintext nomorKk.
	rawRows := testutil.DecodeAs[[]map[string]any](s).Data
	s.Require().Len(rawRows, 1)
	_, hasPlainKK := rawRows[0]["nomorKk"]
	s.False(hasPlainKK, "search response must not expose plaintext nomorKk")

	// Create Umat with NIK, phone normalization, email normalization, lookup, and manual userId link.
	nik1 := "3174010101909999"
	agama := "Katolik"
	tl := "1990-05-12"
	telp := "0812-3456-7890"
	email := "  Budi@Paroki.ORG "
	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		KeluargaID:   &kel1.ID,
		LingkunganID: &ling.ID,
		UserID:       &member.UserID,
		Nama:         "Budi Santoso",
		NIK:          &nik1,
		JenisKelamin: "pria",
		TanggalLahir: &tl,
		Agama:        &agama,
		Telepon:      &telp,
		Email:        &email,
	})
	s.Created()
	umat1 := testutil.DecodeAs[domain.UmatDetailResponse](s).Data
	s.Require().NotNil(umat1.NIK)
	s.Equal("3174010101909999", *umat1.NIK)
	s.Require().NotNil(umat1.NIKMasked)
	s.Equal("************9999", *umat1.NIKMasked)
	s.Require().NotNil(umat1.Telepon)
	s.Equal("+6281234567890", *umat1.Telepon)
	s.Require().NotNil(umat1.Email)
	s.Equal("budi@paroki.org", *umat1.Email)
	s.Require().NotNil(umat1.Keluarga)
	s.Equal("************0001", *umat1.Keluarga.NomorKKMasked)
	s.Empty(umat1.Sakramen)

	// Duplicate NIK -> 409 UMAT_NIK_CONFLICT; Duplicate userId -> 409 UMAT_USER_ALREADY_LINKED.
	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Dup NIK",
		NIK:          &nik1,
		JenisKelamin: "pria",
	})
	s.Conflict()
	s.Equal("UMAT_NIK_CONFLICT", s.GetErrorResponse().Error.Code)

	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Dup User",
		UserID:       &member.UserID,
		JenisKelamin: "pria",
	})
	s.Conflict()
	s.Equal("UMAT_USER_ALREADY_LINKED", s.GetErrorResponse().Error.Code)

	// Non-member userId -> 422 VALIDATION_ERROR.
	nonMemberID := uuid.New()
	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Non Member",
		UserID:       &nonMemberID,
		JenisKelamin: "pria",
	})
	s.UnprocessableEntity()

	// Search Umat by NIK blind index -> masked only.
	s.Request("/api/umat/search", http.MethodPost, domain.SearchUmatRequest{
		Filter: &domain.UmatSearchFilter{NIK: nik1},
	})
	s.OK()
	umatList := testutil.DecodeAs[[]domain.UmatListItemResponse](s).Data
	s.Require().Len(umatList, 1)
	s.Equal("************9999", *umatList[0].NIKMasked)

	// GET Keluarga detail includes masked anggota.
	s.Request("/api/umat/keluarga/"+kel1.ID.String(), http.MethodGet, nil)
	s.OK()
	kelDetail := testutil.DecodeAs[domain.KeluargaDetailResponse](s).Data
	s.Require().Len(kelDetail.Anggota, 1)
	s.Equal("************9999", *kelDetail.Anggota[0].NIKMasked)

	// Delete Keluarga while active Umat references it -> 409 UMAT_IN_USE.
	s.Request("/api/umat/keluarga/"+kel1.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("UMAT_IN_USE", s.GetErrorResponse().Error.Code)

	// PATCH Umat: cascade clear kotaId via uuid.Nil clears kecamatanId/kelurahanId; clear userId & nik.
	nilID := uuid.Nil
	emptyStr := ""
	s.Request("/api/umat/"+umat1.ID.String(), http.MethodPatch, domain.UpdateUmatRequest{
		KotaID: &nilID,
		UserID: &nilID,
		NIK:    &emptyStr,
	})
	s.OK()
	umatPatched := testutil.DecodeAs[domain.UmatDetailResponse](s).Data
	s.Nil(umatPatched.KotaID)
	s.Nil(umatPatched.UserID)
	s.Nil(umatPatched.NIK)

	// Soft-delete Umat, then soft-delete Keluarga -> both 204.
	s.Request("/api/umat/"+umat1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/umat/keluarga/"+kel1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *UmatSuite) TestUmatSakramen_UpsertUpdateDeleteAndCanonicalOrdering() {
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{Kode: "kaj-sak", Nama: "KAJ Sakramen"})
	s.Created()
	k1 := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{Kode: "kbg-sak", Nama: "Keuskupan Bogor"})
	s.Created()
	k2 := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: k1.ID,
		Kode:        "paroki-domicile",
		Nama:        "Paroki Domicile",
	})
	s.Created()
	pDomicile := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: k2.ID,
		Kode:        "paroki-baptis",
		Nama:        "Paroki Baptis Bogor",
	})
	s.Created()
	pOther := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	umatTrue := true
	s.Request("/api/ref/paroki/"+pDomicile.ID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email:     "uadmin-sak@paroki.org",
		Name:      "Umat Admin Sakramen",
		Role:      "member",
		UmatAdmin: &umatTrue,
	})
	s.Created()

	s.loginViaOTP("uadmin-sak@paroki.org")
	s.SetHeader("X-Paroki-Id", pDomicile.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")

	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Maria Sakramen",
		JenisKelamin: "wanita",
	})
	s.Created()
	u := testutil.DecodeAs[domain.UmatDetailResponse](s).Data

	// 1. Insert krisma first, then baptis with cross-paroki parokiSakramenId (auto-populates keuskupanId=k2.ID).
	tglKrisma := "2015-06-20"
	s.Request("/api/umat/"+u.ID.String()+"/sakramen", http.MethodPost, domain.UpsertUmatSakramenRequest{
		Jenis:   "krisma",
		Tanggal: &tglKrisma,
	})
	s.OK()
	krisma := testutil.DecodeAs[domain.UmatSakramenResponse](s).Data

	tglBaptis := "2000-01-15"
	buku := "B-01"
	s.Request("/api/umat/"+u.ID.String()+"/sakramen", http.MethodPost, domain.UpsertUmatSakramenRequest{
		Jenis:            "baptis",
		Tanggal:          &tglBaptis,
		ParokiSakramenID: &pOther.ID,
		Buku:             &buku,
	})
	s.OK()
	baptis := testutil.DecodeAs[domain.UmatSakramenResponse](s).Data
	s.Require().NotNil(baptis.KeuskupanID)
	s.Equal(k2.ID, *baptis.KeuskupanID)

	// 2. Idempotent upsert on baptis updates in place and keeps same ID.
	nomor := "123/2000"
	s.Request("/api/umat/"+u.ID.String()+"/sakramen", http.MethodPost, domain.UpsertUmatSakramenRequest{
		Jenis:            "baptis",
		ParokiSakramenID: &pOther.ID,
		Nomor:            &nomor,
	})
	s.OK()
	baptisUpdated := testutil.DecodeAs[domain.UmatSakramenResponse](s).Data
	s.Equal(baptis.ID, baptisUpdated.ID)
	s.Require().NotNil(baptisUpdated.Nomor)
	s.Equal("123/2000", *baptisUpdated.Nomor)

	// 3. Mismatched keuskupanId & parokiSakramenId -> 422 VALIDATION_ERROR.
	s.Request("/api/umat/"+u.ID.String()+"/sakramen", http.MethodPost, domain.UpsertUmatSakramenRequest{
		Jenis:            "komuni_pertama",
		KeuskupanID:      &k1.ID,
		ParokiSakramenID: &pOther.ID,
	})
	s.UnprocessableEntity()

	// 4. GET /api/umat/:id returns canonical lifecycle order: baptis before krisma.
	s.Request("/api/umat/"+u.ID.String(), http.MethodGet, nil)
	s.OK()
	detail := testutil.DecodeAs[domain.UmatDetailResponse](s).Data
	s.Require().Len(detail.Sakramen, 2)
	s.Equal("baptis", detail.Sakramen[0].Jenis)
	s.Equal("krisma", detail.Sakramen[1].Jenis)

	// 5. PATCH krisma to jenis=baptis -> 409 UMAT_SAKRAMEN_CONFLICT.
	baptisJenis := "baptis"
	s.Request("/api/umat/"+u.ID.String()+"/sakramen/"+krisma.ID.String(), http.MethodPatch, domain.UpdateUmatSakramenRequest{
		Jenis: &baptisJenis,
	})
	s.Conflict()
	s.Equal("UMAT_SAKRAMEN_CONFLICT", s.GetErrorResponse().Error.Code)

	// 6. PATCH baptis clearing parokiSakramenId via uuid.Nil and buku via "".
	nilID := uuid.Nil
	emptyStr := ""
	s.Request("/api/umat/"+u.ID.String()+"/sakramen/"+baptis.ID.String(), http.MethodPatch, domain.UpdateUmatSakramenRequest{
		ParokiSakramenID: &nilID,
		Buku:             &emptyStr,
	})
	s.OK()
	baptisCleared := testutil.DecodeAs[domain.UmatSakramenResponse](s).Data
	s.Nil(baptisCleared.ParokiSakramenID)
	s.Nil(baptisCleared.Buku)

	// 7. Physical hard-delete -> 204, then second delete -> 404 UMAT_SAKRAMEN_NOT_FOUND.
	s.Request("/api/umat/"+u.ID.String()+"/sakramen/"+krisma.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/umat/"+u.ID.String()+"/sakramen/"+krisma.ID.String(), http.MethodDelete, nil)
	s.NotFound()
	s.Equal("UMAT_SAKRAMEN_NOT_FOUND", s.GetErrorResponse().Error.Code)
}

func (s *UmatSuite) TestUmatLookup_RBAC_PIIFree_AndDeleteInUseGuard() {
	ctx := context.Background()
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "kaj-lookup",
		Nama: "Keuskupan Lookup",
	})
	s.Created()
	dio := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: dio.ID,
		Kode:        "p-lookup",
		Nama:        "Paroki Lookup",
	})
	s.Created()
	par := testutil.DecodeAs[refdomain.ParokiResponse](s).Data
	s.SetHeader("X-Paroki-Id", par.ID.String())

	// Seed users: paroki-admin, umat-admin, org-admin, plain paroki member.
	pAdmin, err := s.userFeature.EnsureUser(ctx, "padmin@paroki.org", "Paroki Admin", "", "email", "padmin@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(ctx, par.ID, pAdmin.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(pAdmin.ID), ref.RoleParokiAdmin, authz.ParokiDomain(par.ID)))

	uAdmin, err := s.userFeature.EnsureUser(ctx, "umatadmin@paroki.org", "Umat Admin", "", "email", "umatadmin@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(ctx, par.ID, uAdmin.ID)
	s.Require().NoError(err)
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(uAdmin.ID), umat.RoleAdmin, authz.ParokiDomain(par.ID)))

	orgAdmin, err := s.userFeature.EnsureUser(ctx, "orgadmin@paroki.org", "Org Admin", "", "email", "orgadmin@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(ctx, par.ID, orgAdmin.ID)
	s.Require().NoError(err)

	plainUser, err := s.userFeature.EnsureUser(ctx, "plain@paroki.org", "Plain Member", "", "email", "plain@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(ctx, par.ID, plainUser.ID)
	s.Require().NoError(err)

	orgRow, err := s.client.Organization.Create().
		SetParokiID(par.ID).
		SetKind("section").
		SetCode("sek-lookup").
		SetName("Seksi Lookup").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.OrgMember.Create().
		SetParokiID(par.ID).
		SetOrgID(orgRow.ID).
		SetUserID(orgAdmin.ID).
		SetRole("admin").
		SetStatus("active").
		Save(ctx)
	s.Require().NoError(err)

	// Create one active Umat (with PII) and one deceased Umat.
	s.loginViaOTP("umatadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", par.ID.String())
	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Budi Santoso",
		NamaBaptis:   new("Yohanes"),
		JenisKelamin: "pria",
		NIK:          new("3273010101900001"),
		Telepon:      new("+628123456789"),
		Email:        new("budi.pii@example.com"),
		UserID:       &orgAdmin.ID,
	})
	s.Created()
	budi := testutil.DecodeAs[domain.UmatDetailResponse](s).Data

	s.Request("/api/umat", http.MethodPost, domain.CreateUmatRequest{
		Nama:         "Almarhum Test",
		JenisKelamin: "pria",
		Almarhum:     new(true),
	})
	s.Created()

	// Verify umat:admin, ref:paroki-admin, and active org:admin can call POST /api/umat/lookup without PII.
	for _, email := range []string{"umatadmin@paroki.org", "padmin@paroki.org", "orgadmin@paroki.org"} {
		s.loginViaOTP(email)
		s.SetHeader("X-Paroki-Id", par.ID.String())
		s.Request("/api/umat/lookup", http.MethodPost, domain.LookupUmatRequest{Q: "Budi"})
		s.OK()
		rawBytes, readErr := io.ReadAll(s.HttpResponse.Body)
		s.Require().NoError(readErr)
		rawBody := string(rawBytes)
		s.NotContains(rawBody, "3273010101900001")
		s.NotContains(rawBody, "+628123456789")
		s.NotContains(rawBody, "budi.pii@example.com")

		s.Request("/api/umat/lookup", http.MethodPost, domain.LookupUmatRequest{Q: "Budi"})
		s.OK()
		items := testutil.DecodeAs[[]domain.UmatLookupItemResponse](s).Data
		s.Require().Len(items, 1)
		s.Equal(budi.ID, items[0].ID)
		s.Equal("Budi Santoso", items[0].Nama)
		s.Require().NotNil(items[0].UserID)
		s.Equal(orgAdmin.ID, *items[0].UserID)
	}

	// Regular member gets 403 UMAT_FORBIDDEN.
	s.loginViaOTP("plain@paroki.org")
	s.SetHeader("X-Paroki-Id", par.ID.String())
	s.Request("/api/umat/lookup", http.MethodPost, domain.LookupUmatRequest{})
	s.Forbidden()
	s.Equal("UMAT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Facade helpers GetBasic & GetByUserID.
	basic, err := s.umatFeature.GetBasic(ctx, par.ID, budi.ID)
	s.Require().NoError(err)
	s.Equal("Budi Santoso", basic.Nama)
	byUser, err := s.umatFeature.GetByUserID(ctx, par.ID, orgAdmin.ID)
	s.Require().NoError(err)
	s.Require().NotNil(byUser)
	s.Equal(budi.ID, byUser.ID)

	// 409 UMAT_IN_USE when referenced by an active OrgJabatan.
	perRow, err := s.client.OrgPeriode.Create().
		SetParokiID(par.ID).
		SetOrgID(orgRow.ID).
		SetKode("2026").
		SetNama("2026").
		SetTanggalMulai("2026-01-01").
		SetTanggalSelesai("2026-12-31").
		Save(ctx)
	s.Require().NoError(err)
	jabRow, err := s.client.OrgJabatan.Create().
		SetParokiID(par.ID).
		SetOrgID(orgRow.ID).
		SetPeriodeID(perRow.ID).
		SetJabatan("Ketua").
		SetUmatID(budi.ID).
		Save(ctx)
	s.Require().NoError(err)

	s.loginViaOTP("umatadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", par.ID.String())
	s.Request("/api/umat/"+budi.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("UMAT_IN_USE", s.GetErrorResponse().Error.Code)

	// Soft-delete OrgJabatan -> DeleteUmat succeeds.
	_, err = s.client.OrgJabatan.UpdateOneID(jabRow.ID).SetDeletedAt(time.Now().UTC()).Save(ctx)
	s.Require().NoError(err)
	s.Request("/api/umat/"+budi.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *UmatSuite) TestUmatReferensi_CRUD_Merge_RBAC_AndInUseGuard() {
	ctx := context.Background()

	// Setup keuskupan + paroki + users.
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{Kode: "kaj-ref", Nama: "KAJ Ref"})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "p-ref",
		Nama:        "Paroki Ref",
	})
	s.Created()
	paroki := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	uAdminUser, err := s.userFeature.EnsureUser(ctx, "umatadmin-ref@paroki.org", "Umat Admin Ref", "", "email", "umatadmin-ref@paroki.org")
	s.Require().NoError(err)
	pAdminUser, err := s.userFeature.EnsureUser(ctx, "parokiadmin-ref@paroki.org", "Paroki Admin Ref", "", "email", "parokiadmin-ref@paroki.org")
	s.Require().NoError(err)
	regularUser, err := s.userFeature.EnsureUser(ctx, "user-ref@paroki.org", "Regular User Ref", "", "email", "user-ref@paroki.org")
	s.Require().NoError(err)

	for _, uid := range []uuid.UUID{uAdminUser.ID, pAdminUser.ID, regularUser.ID} {
		id := uid
		s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{UserID: &id})
		s.Created()
	}

	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(uAdminUser.ID), umat.RoleAdmin, authz.ParokiDomain(paroki.ID)))
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(pAdminUser.ID), ref.RoleParokiAdmin, authz.ParokiDomain(paroki.ID)))

	// 1. Super-admin creates global referensi; super-admin cannot write parish-scoped referensi (G8).
	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")

	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category:  "agama",
		Code:      "katolik",
		Label:     "Katolik Global",
		SortOrder: 1,
		Global:    true,
	})
	s.Created()
	globalAgama := testutil.DecodeAs[domain.UmatReferensiResponse](s).Data
	s.Nil(globalAgama.ParokiID)

	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "agama",
		Code:     "kristen",
		Label:    "Kristen Parish",
		Global:   false,
	})
	s.Forbidden()

	// Invalid category -> 422 VALIDATION_ERROR.
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "status_paroki",
		Code:     "aktif",
		Label:    "Aktif",
		Global:   true,
	})
	s.UnprocessableEntity()

	// 2. Paroki-admin without umat:admin -> 403 on write.
	s.loginViaOTP("parokiadmin-ref@paroki.org")
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "agama",
		Code:     "hindu",
		Label:    "Hindu",
	})
	s.Forbidden()

	// 3. Regular user (user:user) -> 200 on search & get, 403 on write.
	s.loginViaOTP("user-ref@paroki.org")
	s.Request("/api/umat/referensi/search", http.MethodPost, domain.SearchUmatReferensiRequest{
		Filter: &domain.UmatReferensiSearchFilter{Category: "agama"},
	})
	s.OK()
	searchRes := testutil.DecodeAs[[]domain.UmatReferensiResponse](s).Data
	s.Len(searchRes, 1)
	s.Equal("Katolik Global", searchRes[0].Label)

	s.Request("/api/umat/referensi/"+globalAgama.ID.String(), http.MethodGet, nil)
	s.OK()

	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "agama",
		Code:     "buddha",
		Label:    "Buddha",
	})
	s.Forbidden()
	s.Request("/api/umat/referensi/"+globalAgama.ID.String(), http.MethodPatch, domain.UpdateUmatReferensiRequest{Label: ptrStr("X")})
	s.Forbidden()
	s.Request("/api/umat/referensi/"+globalAgama.ID.String(), http.MethodDelete, nil)
	s.Forbidden()

	// 4. umat:admin cannot write global, CAN write parish-scoped & override global.
	s.loginViaOTP("umatadmin-ref@paroki.org")
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "agama",
		Code:     "buddha",
		Label:    "Buddha",
		Global:   true,
	})
	s.Forbidden()

	// Parish override for katolik + new parish-only status_rumah.
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category:  "agama",
		Code:      "katolik",
		Label:     "Katolik Paroki",
		SortOrder: 1,
	})
	s.Created()
	parishAgama := testutil.DecodeAs[domain.UmatReferensiResponse](s).Data
	s.Require().NotNil(parishAgama.ParokiID)
	s.Equal(paroki.ID, *parishAgama.ParokiID)

	// Duplicate in same scope -> 409 UMAT_REFERENSI_CONFLICT.
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "agama",
		Code:     "katolik",
		Label:    "Duplicate",
	})
	s.Conflict()
	s.Equal("UMAT_REFERENSI_CONFLICT", s.GetErrorResponse().Error.Code)

	// Search merge (scope=all) returns parish override; scope=global returns global; scope=paroki returns parish.
	s.Request("/api/umat/referensi/search", http.MethodPost, domain.SearchUmatReferensiRequest{
		Filter: &domain.UmatReferensiSearchFilter{Category: "agama", Scope: "all"},
	})
	s.OK()
	merged := testutil.DecodeAs[[]domain.UmatReferensiResponse](s).Data
	s.Len(merged, 1)
	s.Equal("Katolik Paroki", merged[0].Label)

	s.Request("/api/umat/referensi/search", http.MethodPost, domain.SearchUmatReferensiRequest{
		Filter: &domain.UmatReferensiSearchFilter{Category: "agama", Scope: "global"},
	})
	s.OK()
	globList := testutil.DecodeAs[[]domain.UmatReferensiResponse](s).Data
	s.Len(globList, 1)
	s.Equal("Katolik Global", globList[0].Label)

	// Invalid category in search -> 422 VALIDATION_ERROR.
	s.Request("/api/umat/referensi/search", http.MethodPost, domain.SearchUmatReferensiRequest{
		Filter: &domain.UmatReferensiSearchFilter{Category: "bad_cat"},
	})
	s.UnprocessableEntity()

	// Create status_rumah, use it in Keluarga, verify UMAT_IN_USE on delete, then delete Keluarga and verify soft-delete + reactivation.
	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "status_rumah",
		Code:     "kontrak",
		Label:    "Kontrak",
	})
	s.Created()
	rumahRef := testutil.DecodeAs[domain.UmatReferensiResponse](s).Data

	// Create Keluarga with unknown status_rumah -> 422 VALIDATION_ERROR.
	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga: "Kel Unknown",
		StatusRumah:  ptrStr("unknown_code"),
	})
	s.UnprocessableEntity()

	s.Request("/api/umat/keluarga", http.MethodPost, domain.CreateKeluargaRequest{
		NamaKeluarga: "Kel Kontrak",
		StatusRumah:  ptrStr("kontrak"),
	})
	s.Created()
	kelKontrak := testutil.DecodeAs[domain.KeluargaDetailResponse](s).Data

	// Delete status_rumah while in use -> 409 UMAT_IN_USE.
	s.Request("/api/umat/referensi/"+rumahRef.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("UMAT_IN_USE", s.GetErrorResponse().Error.Code)

	// Delete Keluarga -> Delete status_rumah succeeds -> Re-create reactivates same ID.
	s.Request("/api/umat/keluarga/"+kelKontrak.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/umat/referensi/"+rumahRef.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/umat/referensi", http.MethodPost, domain.CreateUmatReferensiRequest{
		Category: "status_rumah",
		Code:     "kontrak",
		Label:    "Kontrak Reactivated",
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.UmatReferensiResponse](s).Data
	s.Equal(rumahRef.ID, reactivated.ID)
	s.Equal("Kontrak Reactivated", reactivated.Label)
}

func ptrStr(v string) *string { return &v }
