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
