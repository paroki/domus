package umat_test

import (
	"context"
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
		SetNamaKeluarga("Keluarga Tanpa KK 1").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNamaKeluarga("Keluarga Tanpa KK 2").
		Save(ctx)
	s.Require().NoError(err)

	// Active duplicate nomor_kk_hash in same paroki fails; soft-deleted allows reuse.
	k1, err := s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNamaKeluarga("Keluarga A").
		SetNomorKkHash(kkHash).
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNamaKeluarga("Keluarga A Dup").
		SetNomorKkHash(kkHash).
		Save(ctx)
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))

	s.Require().NoError(s.client.Keluarga.UpdateOneID(k1.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))

	k2, err := s.client.Keluarga.Create().
		SetParokiID(parokiID).
		SetNamaKeluarga("Keluarga A Reused").
		SetNomorKkHash(kkHash).
		Save(ctx)
	s.Require().NoError(err)
	s.NotEqual(k1.ID, k2.ID)

	// Multiple Umat with NULL nik_hash & user_id are allowed.
	_, err = s.client.Umat.Create().
		SetParokiID(parokiID).
		SetKeluargaID(kNull1.ID).
		SetNama("Umat Null 1").
		SetJenisKelamin(entumat.JenisKelaminPria).
		Save(ctx)
	s.Require().NoError(err)

	// Active duplicate nik_hash and user_id fail; soft-deleting allows reuse.
	u1, err := s.client.Umat.Create().
		SetParokiID(parokiID).
		SetKeluargaID(k2.ID).
		SetNama("Budi").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetNikHash(nikHash).
		SetUserID(userID).
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNama("Budi Dup NIK").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetNikHash(nikHash).
		Save(ctx)
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))

	_, err = s.client.Umat.Create().
		SetParokiID(parokiID).
		SetNama("Budi Dup User").
		SetJenisKelamin(entumat.JenisKelaminPria).
		SetUserID(userID).
		Save(ctx)
	s.Require().Error(err)
	s.True(ent.IsConstraintError(err))

	s.Require().NoError(s.client.Umat.UpdateOneID(u1.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))

	u2, err := s.client.Umat.Create().
		SetParokiID(parokiID).
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

	s.Request("/api/ref/lookups", http.MethodPost, refdomain.CreateLookupRequest{Category: "status_rumah", Code: "milik_sendiri", Label: "Milik Sendiri"})
	s.Created()
	s.Request("/api/ref/lookups", http.MethodPost, refdomain.CreateLookupRequest{Category: "agama", Code: "katolik", Label: "Katolik"})
	s.Created()

	// Paroki-admin WITHOUT umat:admin -> 403 on /api/umat/keluarga/search.
	s.Request("/api/umat/keluarga/search", http.MethodPost, domain.SearchKeluargaRequest{})
	s.Forbidden()

	// Super-admin WITHOUT umat:admin -> 403 on /api/umat/search (even with X-Paroki-Id), 400 TENANT_REQUIRED without X-Paroki-Id.
	s.loginViaOTP("admin@paroki.org")
	s.SetHeader("X-Paroki-Id", "")
	s.Request("/api/umat/search", http.MethodPost, domain.SearchUmatRequest{})
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	s.Request("/api/umat/search", http.MethodPost, domain.SearchUmatRequest{})
	s.Forbidden()

	// Login as umat:admin.
	s.loginViaOTP("uadmin@paroki.org")

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
