package ref_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/ent/umatsakramen"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/ref/domain"
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

type RefFeatureSuite struct {
	testutil.ApiTestSuite[any]
	client      *ent.Client
	enforcer    *authz.Enforcer
	sender      *stubOTPSender
	userFeature *user.Feature
	refFeature  *ref.Feature
}

func TestRefFeatureSuite(t *testing.T) {
	suite.Run(t, new(RefFeatureSuite))
}

func (s *RefFeatureSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.sender = &stubOTPSender{}

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

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature)
}

func (s *RefFeatureSuite) loginViaOTP(email string) {
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

func (s *RefFeatureSuite) TestKeuskupanCRUD_ReactivationAndDeleteGuard() {
	s.loginViaOTP("admin@paroki.org")

	// Create Keuskupan with mixed-case kode and local phone.
	telepon := "0812-3456-7890"
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode:    "  KAJ  ",
		Nama:    "Keuskupan Agung Jakarta",
		Telepon: &telepon,
	})
	s.Created()
	created := testutil.DecodeAs[domain.KeuskupanResponse](s).Data
	s.Equal("kaj", created.Kode)
	s.Require().NotNil(created.Telepon)
	s.Equal("+6281234567890", *created.Telepon)

	// Duplicate kode returns 409 REF_CODE_CONFLICT.
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "Duplicate",
	})
	s.Conflict()
	errResp := s.GetErrorResponse()
	s.Equal("REF_CODE_CONFLICT", errResp.Error.Code)

	// Search & Get.
	s.Request("/api/ref/keuskupan/search", http.MethodPost, domain.SearchKeuskupanRequest{Q: "Jakarta"})
	s.OK()
	list := testutil.DecodeAs[[]domain.KeuskupanResponse](s).Data
	s.Len(list, 1)

	s.Request("/api/ref/keuskupan/"+created.ID.String(), http.MethodGet, nil)
	s.OK()

	// Delete & Reactivate.
	s.Request("/api/ref/keuskupan/"+created.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "KAJ Reactivated",
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.KeuskupanResponse](s).Data
	s.Equal(created.ID, reactivated.ID)
	s.Equal("KAJ Reactivated", reactivated.Nama)
	s.True(reactivated.Active)

	// Create Paroki under Keuskupan -> DELETE Keuskupan blocked with 409 REF_IN_USE.
	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: reactivated.ID,
		Kode:        "p-katedral",
		Nama:        "Paroki Katedral",
	})
	s.Created()

	s.Request("/api/ref/keuskupan/"+reactivated.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Regular user can read but cannot write.
	_, err := s.userFeature.EnsureUser(context.Background(), "viewer@paroki.org", "Viewer", "", "email", "viewer@paroki.org")
	s.Require().NoError(err)
	s.loginViaOTP("viewer@paroki.org")

	s.Request("/api/ref/keuskupan/"+reactivated.ID.String(), http.MethodGet, nil)
	s.OK()
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kas",
		Nama: "Semarang",
	})
	s.Forbidden()
}

func (s *RefFeatureSuite) TestParokiCRUD_CurrentAndProfileAndMemberOnly() {
	s.loginViaOTP("admin@paroki.org")

	// Current paroki without tenant returns 400 TENANT_REQUIRED.
	s.Request("/api/ref/paroki/current", http.MethodGet, nil)
	s.BadRequest()

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "KAJ",
	})
	s.Created()
	d := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: d.ID,
		Kode:        "p-01",
		Nama:        "Paroki Satu",
	})
	s.Created()
	p1 := testutil.DecodeAs[domain.ParokiResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: d.ID,
		Kode:        "p-02",
		Nama:        "Paroki Dua",
	})
	s.Created()

	// Invite alice@paroki.org as paroki admin of p1.
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "alice@paroki.org",
		Name:  "Alice",
		Role:  "admin",
	})
	s.Created()
	aliceMember := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.Equal("admin", aliceMember.Role)
	s.Equal("alice@paroki.org", aliceMember.Email)

	// Alice logs in -> GET /api/me includes ref:paroki-admin in roles.
	s.loginViaOTP("alice@paroki.org")
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[userdomain.UserResponse](s).Data
	s.Require().NotNil(me.ActiveParokiID)
	s.Equal(p1.ID, *me.ActiveParokiID)
	s.Contains(me.Roles, ref.RoleParokiAdmin)

	// Alice can GET/PATCH /current.
	s.Request("/api/ref/paroki/current", http.MethodGet, nil)
	s.OK()

	newNama := "Paroki Satu Updated"
	s.Request("/api/ref/paroki/current", http.MethodPatch, domain.UpdateParokiProfileRequest{
		Nama: &newNama,
	})
	s.OK()
	updated := testutil.DecodeAs[domain.ParokiResponse](s).Data
	s.Equal(newNama, updated.Nama)

	// Paroki admin cannot change kode on PATCH /:id.
	newKode := "p-01-hacked"
	s.Request("/api/ref/paroki/"+p1.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{
		Kode: &newKode,
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// memberOnly search returns only p1 for alice.
	memberOnly := true
	s.Request("/api/ref/paroki/search", http.MethodPost, domain.SearchParokiRequest{
		Filter: &domain.ParokiSearchFilter{MemberOnly: &memberOnly},
	})
	s.OK()
	myParoki := testutil.DecodeAs[[]domain.ParokiResponse](s).Data
	s.Len(myParoki, 1)
	s.Equal(p1.ID, myParoki[0].ID)

	// Wilayah under p1 blocks DELETE /api/ref/paroki/:id with 409 REF_IN_USE.
	_, err := s.client.Wilayah.Create().SetParokiID(p1.ID).SetKode("z-1").SetNama("Wilayah 1").Save(context.Background())
	s.Require().NoError(err)

	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/paroki/"+p1.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)
}

func (s *RefFeatureSuite) TestParokiDelete_SucceedsWhenUnused() {
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kaj", Nama: "KAJ"})
	s.Created()
	d := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{KeuskupanID: d.ID, Kode: "p-agent", Nama: "Paroki Agent"})
	s.Created()
	p := testutil.DecodeAs[domain.ParokiResponse](s).Data

	s.Request("/api/ref/paroki/"+p.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *RefFeatureSuite) TestParokiMember_CasbinSyncAndLastAdminGuard() {
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kaj", Nama: "KAJ"})
	s.Created()
	d := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{KeuskupanID: d.ID, Kode: "p-01", Nama: "Paroki Satu"})
	s.Created()
	p1 := testutil.DecodeAs[domain.ParokiResponse](s).Data

	// Add Alice as admin, Bob as member.
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "alice@paroki.org",
		Role:  "admin",
	})
	s.Created()
	alice := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data

	// Duplicate add returns 409 REF_PAROKI_MEMBER_ALREADY_EXISTS.
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		UserID: &alice.UserID,
		Role:   "member",
	})
	s.Conflict()
	s.Equal("REF_PAROKI_MEMBER_ALREADY_EXISTS", s.GetErrorResponse().Error.Code)

	// Last admin cannot be demoted or removed.
	memberRole := "member"
	adminRole := "admin"
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+alice.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		Role: &memberRole,
	})
	s.Conflict()
	s.Equal("REF_LAST_PAROKI_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+alice.UserID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_LAST_PAROKI_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	// Alice adds Bob, promotes Bob to admin, demotes Bob, removes Bob.
	s.loginViaOTP("alice@paroki.org")
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "bob@paroki.org",
		Name:  "Bob",
		Role:  "member",
	})
	s.Created()
	bob := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data

	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		Role: &adminRole,
	})
	s.OK()
	hasAdmin, err := s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), ref.RoleParokiAdmin, authz.ParokiDomain(p1.ID))
	s.Require().NoError(err)
	s.True(hasAdmin)

	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		Role: &memberRole,
	})
	s.OK()
	hasAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), ref.RoleParokiAdmin, authz.ParokiDomain(p1.ID))
	s.Require().NoError(err)
	s.False(hasAdmin)

	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Adding an existing user by email with empty Name preserves their existing name.
	existingCarol, err := s.userFeature.EnsureUser(context.Background(), "carol@paroki.org", "Carol Original", "", "email", "carol@paroki.org")
	s.Require().NoError(err)
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "carol@paroki.org",
		Role:  "member",
	})
	s.Created()
	carolMember := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.Equal(existingCarol.ID, carolMember.UserID)
	s.Equal("Carol Original", carolMember.Name)
}

func (s *RefFeatureSuite) TestFacadeHelpers() {
	ctx := context.Background()
	d, err := s.client.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	s.Require().NoError(err)
	p1, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p1").SetNama("P1").Save(ctx)
	s.Require().NoError(err)
	p2, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p2").SetNama("P2").Save(ctx)
	s.Require().NoError(err)

	z1, err := s.client.Wilayah.Create().SetParokiID(p1.ID).SetKode("z1").SetNama("Z1").Save(ctx)
	s.Require().NoError(err)
	w1, err := s.client.Lingkungan.Create().SetParokiID(p1.ID).SetWilayahID(z1.ID).SetKode("w1").SetNama("W1").Save(ctx)
	s.Require().NoError(err)

	gotP, err := s.refFeature.GetParoki(ctx, p1.ID)
	s.Require().NoError(err)
	s.Equal(p1.ID, gotP.ID)

	gotZ, err := s.refFeature.ValidateWilayahInParoki(ctx, p1.ID, z1.ID)
	s.Require().NoError(err)
	s.Equal(z1.ID, gotZ.ID)

	_, err = s.refFeature.ValidateWilayahInParoki(ctx, p2.ID, z1.ID)
	s.ErrorIs(err, domain.ErrWilayahNotFound)

	gotW, err := s.refFeature.ValidateLingkunganInParoki(ctx, p1.ID, w1.ID)
	s.Require().NoError(err)
	s.Equal(w1.ID, gotW.ID)

	u, err := s.userFeature.EnsureUser(ctx, "member@paroki.org", "Member", "", "email", "member@paroki.org")
	s.Require().NoError(err)
	pm1, err := s.refFeature.EnsureParokiMember(ctx, p1.ID, u.ID)
	s.Require().NoError(err)
	pm2, err := s.refFeature.EnsureParokiMember(ctx, p1.ID, u.ID)
	s.Require().NoError(err)
	s.Equal(pm1.ID, pm2.ID)
	s.NotEqual(uuid.Nil, pm1.ID)
}

func (s *RefFeatureSuite) setupTwoParokiWithAliceAdmin() (domain.ParokiResponse, domain.ParokiResponse) {
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kaj", Nama: "KAJ"})
	s.Created()
	d := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{KeuskupanID: d.ID, Kode: "p-01", Nama: "Paroki 1"})
	s.Created()
	p1 := testutil.DecodeAs[domain.ParokiResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{KeuskupanID: d.ID, Kode: "p-02", Nama: "Paroki 2"})
	s.Created()
	p2 := testutil.DecodeAs[domain.ParokiResponse](s).Data

	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "alice@paroki.org",
		Name:  "Alice",
		Role:  "admin",
	})
	s.Created()
	return p1, p2
}

func (s *RefFeatureSuite) TestWilayahAndLingkunganCRUD_ReactivationIsolationAndInUse() {
	// Missing tenant returns 400 TENANT_REQUIRED when no active paroki exists.
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/wilayah/search", http.MethodPost, domain.SearchWilayahRequest{})
	s.BadRequest()

	p1, p2 := s.setupTwoParokiWithAliceAdmin()

	// Super-admin without ref:paroki-admin in p1 gets 403 on write.
	s.SetHeader("X-Paroki-Id", p1.ID.String())
	s.Request("/api/ref/wilayah", http.MethodPost, domain.CreateWilayahRequest{Kode: "z-super", Nama: "Denied"})
	s.Forbidden()

	// Alice (paroki-admin of p1) creates Wilayah with E.164 phone normalization.
	s.loginViaOTP("alice@paroki.org")
	coordPhone := "0811-2233-4455"
	s.Request("/api/ref/wilayah", http.MethodPost, domain.CreateWilayahRequest{
		Kode:               "WIL-01",
		Nama:               "Wilayah 1",
		TeleponKoordinator: &coordPhone,
	})
	s.Created()
	z1 := testutil.DecodeAs[domain.WilayahResponse](s).Data
	s.Equal("wil-01", z1.Kode)
	s.Equal("wilayah", z1.Jenis)
	s.Require().NotNil(z1.TeleponKoordinator)
	s.Equal("+6281122334455", *z1.TeleponKoordinator)

	// Delete & reactivate Wilayah.
	s.Request("/api/ref/wilayah/"+z1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/wilayah", http.MethodPost, domain.CreateWilayahRequest{
		Jenis: "stasi",
		Kode:  "wil-01",
		Nama:  "Stasi 1",
	})
	s.Created()
	zReactivated := testutil.DecodeAs[domain.WilayahResponse](s).Data
	s.Equal(z1.ID, zReactivated.ID)
	s.Equal("stasi", zReactivated.Jenis)

	// Create Lingkungan under Wilayah -> Delete Wilayah blocked with 409 REF_IN_USE.
	s.Request("/api/ref/lingkungan", http.MethodPost, domain.CreateLingkunganRequest{
		WilayahID: z1.ID,
		Kode:      "LING-01",
		Nama:      "Lingkungan 1",
	})
	s.Created()
	w1 := testutil.DecodeAs[domain.LingkunganResponse](s).Data

	s.Request("/api/ref/wilayah/"+z1.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Delete Lingkungan -> Delete Wilayah succeeds.
	s.Request("/api/ref/lingkungan/"+w1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/wilayah/"+z1.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Cross-tenant isolation: wilayah in p2 is 404 when active paroki is p1.
	z2, err := s.client.Wilayah.Create().SetParokiID(p2.ID).SetKode("z-p2").SetNama("Wilayah P2").Save(context.Background())
	s.Require().NoError(err)
	s.Request("/api/ref/wilayah/"+z2.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Request("/api/ref/lingkungan", http.MethodPost, domain.CreateLingkunganRequest{
		WilayahID: z2.ID,
		Kode:      "w-x",
		Nama:      "Cross Tenant",
	})
	s.NotFound()
}

func (s *RefFeatureSuite) TestLookup_HybridGlobalAndParokiOverride() {
	p1, _ := s.setupTwoParokiWithAliceAdmin()

	// SuperAdmin creates global lookup.
	s.Request("/api/ref/lookups", http.MethodPost, domain.CreateLookupRequest{
		Category:  "SACRAMENT",
		Code:      "BAPTISM",
		Label:     "Baptism",
		SortOrder: 1,
		Global:    true,
	})
	s.Created()
	globalRow := testutil.DecodeAs[domain.LookupResponse](s).Data
	s.Nil(globalRow.ParokiID)
	s.Equal("sacrament", globalRow.Category)
	s.Equal("baptism", globalRow.Code)

	// Alice cannot create or update global lookup.
	s.loginViaOTP("alice@paroki.org")
	s.Request("/api/ref/lookups", http.MethodPost, domain.CreateLookupRequest{
		Category: "sacrament",
		Code:     "confirmation",
		Label:    "Confirmation",
		Global:   true,
	})
	s.Forbidden()
	newLabel := "Hacked"
	s.Request("/api/ref/lookups/"+globalRow.ID.String(), http.MethodPatch, domain.UpdateLookupRequest{Label: &newLabel})
	s.Forbidden()

	// Alice creates paroki override for (sacrament, baptism).
	s.Request("/api/ref/lookups", http.MethodPost, domain.CreateLookupRequest{
		Category:  "sacrament",
		Code:      "baptism",
		Label:     "Baptis Paroki",
		SortOrder: 1,
	})
	s.Created()
	parokiRow := testutil.DecodeAs[domain.LookupResponse](s).Data
	s.Require().NotNil(parokiRow.ParokiID)
	s.Equal(p1.ID, *parokiRow.ParokiID)

	// Search scope=all merges and overrides global with paroki row.
	s.Request("/api/ref/lookups/search", http.MethodPost, domain.SearchLookupsRequest{
		Filter: &domain.LookupSearchFilter{Category: "sacrament", Scope: "all"},
	})
	s.OK()
	allRows := testutil.DecodeAs[[]domain.LookupResponse](s).Data
	s.Len(allRows, 1)
	s.Equal("Baptis Paroki", allRows[0].Label)

	// Search scope=global returns global row.
	s.Request("/api/ref/lookups/search", http.MethodPost, domain.SearchLookupsRequest{
		Filter: &domain.LookupSearchFilter{Category: "sacrament", Scope: "global"},
	})
	s.OK()
	globalList := testutil.DecodeAs[[]domain.LookupResponse](s).Data
	s.Len(globalList, 1)
	s.Equal("Baptism", globalList[0].Label)

	// SuperAdmin deletes and recreates global lookup -> reactivates cleanly.
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/lookups/"+globalRow.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/lookups", http.MethodPost, domain.CreateLookupRequest{
		Category: "sacrament",
		Code:     "baptism",
		Label:    "Baptism v2",
		Global:   true,
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.LookupResponse](s).Data
	s.Equal(globalRow.ID, reactivated.ID)
	s.Equal("Baptism v2", reactivated.Label)
}

func (s *RefFeatureSuite) TestLookup_SearchSortOrder() {
	s.setupTwoParokiWithAliceAdmin()
	s.loginViaOTP("alice@paroki.org")

	for _, tc := range []struct {
		code  string
		order int
	}{
		{code: "third", order: 30},
		{code: "first", order: 10},
		{code: "second", order: 20},
	} {
		s.Request("/api/ref/lookups", http.MethodPost, domain.CreateLookupRequest{
			Category:  "status",
			Code:      tc.code,
			Label:     tc.code,
			SortOrder: tc.order,
		})
		s.Created()
	}

	s.Request("/api/ref/lookups/search", http.MethodPost, domain.SearchLookupsRequest{
		Filter: &domain.LookupSearchFilter{Category: "status"},
	})
	s.OK()
	rows := testutil.DecodeAs[[]domain.LookupResponse](s).Data
	s.Require().Len(rows, 3)
	s.Equal("first", rows[0].Code)
	s.Equal("second", rows[1].Code)
	s.Equal("third", rows[2].Code)
}

func (s *RefFeatureSuite) TestKevikepanCRUD_SearchPaginationReactivationAndRBAC() {
	s.loginViaOTP("admin@paroki.org")

	// Create two Keuskupan.
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "Keuskupan Agung Jakarta",
	})
	s.Created()
	kaj := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kas",
		Nama: "Keuskupan Agung Semarang",
	})
	s.Created()
	kas := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	// Facade GetKeuskupan.
	gotKaj, err := s.refFeature.GetKeuskupan(context.Background(), kaj.ID)
	s.Require().NoError(err)
	s.Equal(kaj.ID, gotKaj.ID)
	_, err = s.refFeature.GetKeuskupan(context.Background(), uuid.New())
	s.Require().ErrorIs(err, domain.ErrKeuskupanNotFound)

	// Invalid keuskupanId -> 404 REF_KEUSKUPAN_NOT_FOUND.
	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: uuid.New(),
		Kode:        "pusat",
		Nama:        "Kevikepan Pusat",
	})
	s.NotFound()
	s.Equal("REF_KEUSKUPAN_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Invalid kode -> 422.
	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kaj.ID,
		Kode:        "A",
		Nama:        "Kevikepan Pusat",
	})
	s.UnprocessableEntity()

	// Create two Kevikepan under KAJ and one under KAS.
	vikep := "RD. Anton"
	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kaj.ID,
		Kode:        "  PUSAT ",
		Nama:        " Kevikepan Jakarta Pusat ",
		Vikep:       &vikep,
	})
	s.Created()
	pusat := testutil.DecodeAs[domain.KevikepanResponse](s).Data
	s.Equal("pusat", pusat.Kode)
	s.Equal("Kevikepan Jakarta Pusat", pusat.Nama)
	s.Require().NotNil(pusat.Vikep)
	s.Equal("RD. Anton", *pusat.Vikep)

	// Duplicate (keuskupanId, kode) -> 409 REF_CODE_CONFLICT.
	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kaj.ID,
		Kode:        "pusat",
		Nama:        "Duplicate",
	})
	s.Conflict()
	s.Equal("REF_CODE_CONFLICT", s.GetErrorResponse().Error.Code)

	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kaj.ID,
		Kode:        "barat",
		Nama:        "Kevikepan Jakarta Barat",
	})
	s.Created()
	barat := testutil.DecodeAs[domain.KevikepanResponse](s).Data

	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kas.ID,
		Kode:        "pusat",
		Nama:        "Kevikepan Semarang",
	})
	s.Created()

	// Search with keuskupanId filter + keyset pagination.
	s.Request("/api/ref/kevikepan/search", http.MethodPost, domain.SearchKevikepanRequest{
		Filter: &domain.KevikepanSearchFilter{KeuskupanID: &kaj.ID},
		Limit:  1,
	})
	s.OK()
	p1 := s.PagedResponse()
	s.True(p1.Meta.HasMore)
	s.Require().NotEmpty(p1.Meta.NextCursor)

	s.Request("/api/ref/kevikepan/search", http.MethodPost, domain.SearchKevikepanRequest{
		Filter: &domain.KevikepanSearchFilter{KeuskupanID: &kaj.ID},
		Cursor: p1.Meta.NextCursor,
		Limit:  1,
	})
	s.OK()
	p2 := s.PagedResponse()
	s.False(p2.Meta.HasMore)

	// PATCH clear vikep with "".
	empty := ""
	s.Request("/api/ref/kevikepan/"+pusat.ID.String(), http.MethodPatch, domain.UpdateKevikepanRequest{
		Vikep: &empty,
	})
	s.OK()
	updated := testutil.DecodeAs[domain.KevikepanResponse](s).Data
	s.Nil(updated.Vikep)

	// PATCH kode collision -> 409 REF_CODE_CONFLICT.
	dupKode := "barat"
	s.Request("/api/ref/kevikepan/"+pusat.ID.String(), http.MethodPatch, domain.UpdateKevikepanRequest{
		Kode: &dupKode,
	})
	s.Conflict()
	s.Equal("REF_CODE_CONFLICT", s.GetErrorResponse().Error.Code)

	// Delete Keuskupan blocked by Kevikepan -> 409 REF_IN_USE.
	s.Request("/api/ref/keuskupan/"+kaj.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Delete Kevikepan -> 204, then reactivate on Create.
	s.Request("/api/ref/kevikepan/"+barat.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/ref/kevikepan/"+barat.ID.String(), http.MethodGet, nil)
	s.NotFound()

	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kaj.ID,
		Kode:        "barat",
		Nama:        "Kevikepan Barat Reactivated",
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.KevikepanResponse](s).Data
	s.Equal(barat.ID, reactivated.ID)
	s.Equal("Kevikepan Barat Reactivated", reactivated.Nama)
	s.True(reactivated.Active)

	// Regular user can read but cannot write.
	_, err = s.userFeature.EnsureUser(context.Background(), "viewer@paroki.org", "Viewer", "", "email", "viewer@paroki.org")
	s.Require().NoError(err)
	s.loginViaOTP("viewer@paroki.org")

	s.Request("/api/ref/kevikepan/"+pusat.ID.String(), http.MethodGet, nil)
	s.OK()
	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{
		KeuskupanID: kaj.ID,
		Kode:        "timur",
		Nama:        "Timur",
	})
	s.Forbidden()
}

func (s *RefFeatureSuite) TestSipilCRUD_KotaKecamatanKelurahanHierarchyAndFacade() {
	s.loginViaOTP("admin@paroki.org")

	// Create Kota.
	s.Request("/api/ref/kota", http.MethodPost, domain.CreateKotaRequest{
		Kode: "3171",
		Nama: "Kota Jakarta Selatan",
	})
	s.Created()
	jaksel := testutil.DecodeAs[domain.KotaResponse](s).Data

	s.Request("/api/ref/kota", http.MethodPost, domain.CreateKotaRequest{
		Kode: "3172",
		Nama: "Kota Jakarta Timur",
	})
	s.Created()
	jaktim := testutil.DecodeAs[domain.KotaResponse](s).Data

	// Create Kecamatan under jaksel.
	s.Request("/api/ref/kecamatan", http.MethodPost, domain.CreateKecamatanRequest{
		KotaID: jaksel.ID,
		Kode:   "317101",
		Nama:   "Tebet",
	})
	s.Created()
	tebet := testutil.DecodeAs[domain.KecamatanResponse](s).Data

	// Create Kelurahan under tebet with kodePos validation.
	tooLongPos := "12345678901"
	s.Request("/api/ref/kelurahan", http.MethodPost, domain.CreateKelurahanRequest{
		KecamatanID: tebet.ID,
		Kode:        "31710101",
		Nama:        "Tebet Barat",
		KodePos:     &tooLongPos,
	})
	s.UnprocessableEntity()

	kodePos := "12810"
	s.Request("/api/ref/kelurahan", http.MethodPost, domain.CreateKelurahanRequest{
		KecamatanID: tebet.ID,
		Kode:        "31710101",
		Nama:        "Tebet Barat",
		KodePos:     &kodePos,
	})
	s.Created()
	tebetBarat := testutil.DecodeAs[domain.KelurahanResponse](s).Data
	s.Require().NotNil(tebetBarat.KodePos)
	s.Equal("12810", *tebetBarat.KodePos)

	// Clear kodePos via PATCH "".
	empty := ""
	s.Request("/api/ref/kelurahan/"+tebetBarat.ID.String(), http.MethodPatch, domain.UpdateKelurahanRequest{
		KodePos: &empty,
	})
	s.OK()
	cleared := testutil.DecodeAs[domain.KelurahanResponse](s).Data
	s.Nil(cleared.KodePos)

	// Facade hierarchy validation.
	ctx := context.Background()
	_, err := s.refFeature.ValidateKota(ctx, jaksel.ID)
	s.Require().NoError(err)
	_, err = s.refFeature.ValidateKecamatan(ctx, jaksel.ID, tebet.ID)
	s.Require().NoError(err)
	_, err = s.refFeature.ValidateKecamatan(ctx, jaktim.ID, tebet.ID)
	s.Require().ErrorIs(err, domain.ErrKecamatanNotFound)
	_, err = s.refFeature.ValidateKelurahan(ctx, tebet.ID, tebetBarat.ID)
	s.Require().NoError(err)
	_, err = s.refFeature.ValidateKelurahan(ctx, uuid.New(), tebetBarat.ID)
	s.Require().ErrorIs(err, domain.ErrKelurahanNotFound)

	// 409 REF_IN_USE guards.
	s.Request("/api/ref/kota/"+jaksel.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	s.Request("/api/ref/kecamatan/"+tebet.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Delete bottom-up succeeds.
	s.Request("/api/ref/kelurahan/"+tebetBarat.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/kecamatan/"+tebet.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/kota/"+jaksel.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Reactivate Kota.
	s.Request("/api/ref/kota", http.MethodPost, domain.CreateKotaRequest{
		Kode: "3171",
		Nama: "Jakarta Selatan Reactivated",
	})
	s.Created()
	jakselReactivated := testutil.DecodeAs[domain.KotaResponse](s).Data
	s.Equal(jaksel.ID, jakselReactivated.ID)

	// Regular user can search/get sipil, cannot write.
	_, err = s.userFeature.EnsureUser(context.Background(), "sipil-viewer@paroki.org", "Sipil Viewer", "", "email", "sipil-viewer@paroki.org")
	s.Require().NoError(err)
	s.loginViaOTP("sipil-viewer@paroki.org")

	s.Request("/api/ref/kota/search", http.MethodPost, domain.SearchKotaRequest{Q: "Selatan"})
	s.OK()
	s.Request("/api/ref/kota", http.MethodPost, domain.CreateKotaRequest{
		Kode: "3173",
		Nama: "Jakarta Barat",
	})
	s.Forbidden()
}

func (s *RefFeatureSuite) TestParokiEnrichment_TimezoneHierarchyClearingAndDeleteGuards() {
	s.loginViaOTP("admin@paroki.org")

	// Setup Keuskupan, Kevikepan, and Sipil hierarchy.
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kaj", Nama: "KAJ"})
	s.Created()
	kaj := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kas", Nama: "KAS"})
	s.Created()
	kas := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{KeuskupanID: kaj.ID, Kode: "pusat", Nama: "Pusat"})
	s.Created()
	kevKaj := testutil.DecodeAs[domain.KevikepanResponse](s).Data

	s.Request("/api/ref/kevikepan", http.MethodPost, domain.CreateKevikepanRequest{KeuskupanID: kas.ID, Kode: "semarang", Nama: "Semarang"})
	s.Created()
	kevKas := testutil.DecodeAs[domain.KevikepanResponse](s).Data

	s.Request("/api/ref/kota", http.MethodPost, domain.CreateKotaRequest{Kode: "3171", Nama: "Jakarta Pusat"})
	s.Created()
	kota := testutil.DecodeAs[domain.KotaResponse](s).Data

	s.Request("/api/ref/kecamatan", http.MethodPost, domain.CreateKecamatanRequest{KotaID: kota.ID, Kode: "317101", Nama: "Gambir"})
	s.Created()
	kec := testutil.DecodeAs[domain.KecamatanResponse](s).Data

	s.Request("/api/ref/kelurahan", http.MethodPost, domain.CreateKelurahanRequest{KecamatanID: kec.ID, Kode: "31710101", Nama: "Gambir"})
	s.Created()
	kel := testutil.DecodeAs[domain.KelurahanResponse](s).Data

	// Invalid timezone -> 422.
	badTZ := "Invalid/Zone"
	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: kaj.ID,
		Kode:        "p-bad-tz",
		Nama:        "Bad TZ",
		Timezone:    &badTZ,
	})
	s.UnprocessableEntity()

	// Kevikepan from different Keuskupan -> 422.
	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: kaj.ID,
		KevikepanID: &kevKas.ID,
		Kode:        "p-bad-kev",
		Nama:        "Bad Kev",
	})
	s.UnprocessableEntity()

	// Create parent Paroki with default timezone ("Asia/Makassar").
	lat := -6.1695
	lng := 106.8331
	kodeResmi := "KAJ-001"
	status := "paroki"
	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID:  kaj.ID,
		KevikepanID:  &kevKaj.ID,
		Kode:         "p-katedral",
		KodeResmi:    &kodeResmi,
		Nama:         "Paroki Katedral",
		KotaID:       &kota.ID,
		KecamatanID:  &kec.ID,
		KelurahanID:  &kel.ID,
		Latitude:     &lat,
		Longitude:    &lng,
		StatusParoki: &status,
	})
	s.Created()
	parent := testutil.DecodeAs[domain.ParokiResponse](s).Data
	s.Equal("Asia/Makassar", parent.Timezone)
	s.Require().NotNil(parent.KevikepanID)
	s.Equal(kevKaj.ID, *parent.KevikepanID)
	s.Require().NotNil(parent.KodeResmi)
	s.Equal("KAJ-001", *parent.KodeResmi)

	// Search by KevikepanID filter.
	s.Request("/api/ref/paroki/search", http.MethodPost, domain.SearchParokiRequest{
		Filter: &domain.ParokiSearchFilter{KevikepanID: &kevKaj.ID},
	})
	s.OK()
	s.Len(testutil.DecodeAs[[]domain.ParokiResponse](s).Data, 1)

	// Delete guards: Kevikepan, Kota, Kecamatan, Kelurahan in use -> 409 REF_IN_USE.
	s.Request("/api/ref/kevikepan/"+kevKaj.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	s.Request("/api/ref/kelurahan/"+kel.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Create child Paroki under parent.
	wib := "Asia/Jakarta"
	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: kaj.ID,
		IndukID:     &parent.ID,
		Kode:        "p-stasi",
		Nama:        "Stasi Katedral",
		Timezone:    &wib,
	})
	s.Created()
	child := testutil.DecodeAs[domain.ParokiResponse](s).Data
	s.Equal("Asia/Jakarta", child.Timezone)

	// 1-level induk_id guard: grandchild -> 422, parent becoming child -> 422, self -> 422.
	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: kaj.ID,
		IndukID:     &child.ID,
		Kode:        "p-grandchild",
		Nama:        "Grandchild",
	})
	s.UnprocessableEntity()

	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{
		IndukID: &child.ID,
	})
	s.UnprocessableEntity()

	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{
		IndukID: &parent.ID,
	})
	s.UnprocessableEntity()

	// Delete parent blocked by child Paroki -> 409 REF_IN_USE.
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// PATCH invalid timezone ("", "Local") -> 422.
	empty := ""
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{Timezone: &empty})
	s.UnprocessableEntity()
	localTZ := "Local"
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{Timezone: &localTZ})
	s.UnprocessableEntity()

	// Coordinate validation & clearCoordinates.
	badLat := 91.0
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{Latitude: &badLat})
	s.UnprocessableEntity()
	clearTrue := true
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{
		ClearCoordinates: &clearTrue,
		Latitude:         &lat,
	})
	s.UnprocessableEntity()

	// Cascade clear kotaId via uuid.Nil clears kecamatanId & kelurahanId, plus clearCoordinates.
	nilID := uuid.Nil
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{
		KotaID:           &nilID,
		ClearCoordinates: &clearTrue,
	})
	s.OK()
	cleared := testutil.DecodeAs[domain.ParokiResponse](s).Data
	s.Nil(cleared.KotaID)
	s.Nil(cleared.KecamatanID)
	s.Nil(cleared.KelurahanID)
	s.Nil(cleared.Latitude)
	s.Nil(cleared.Longitude)

	// Role boundary: paroki-admin cannot edit structural fields via PATCH /paroki/:id -> 403 REF_FORBIDDEN.
	alice, err := s.userFeature.EnsureUser(context.Background(), "alice-200@paroki.org", "Alice", "", "email", "alice-200@paroki.org")
	s.Require().NoError(err)
	s.Request("/api/ref/paroki/"+parent.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		UserID: &alice.ID,
		Role:   "admin",
	})
	s.Created()

	s.loginViaOTP("alice-200@paroki.org")
	s.Request("/api/ref/paroki/"+parent.ID.String(), http.MethodPatch, domain.UpdateParokiRequest{
		KodeResmi: &kodeResmi,
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Paroki admin CAN update profile & domicile fields via PATCH /paroki/current.
	s.SetHeader("X-Paroki-Id", parent.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")
	fax := "+62-21-3456789"
	wit := "Asia/Jayapura"
	s.Request("/api/ref/paroki/current", http.MethodPatch, domain.UpdateParokiProfileRequest{
		KotaID:      &kota.ID,
		KecamatanID: &kec.ID,
		KelurahanID: &kel.ID,
		Timezone:    &wit,
		Fax:         &fax,
	})
	s.OK()
	prof := testutil.DecodeAs[domain.ParokiResponse](s).Data
	s.Equal("Asia/Jayapura", prof.Timezone)
	s.Require().NotNil(prof.Fax)
	s.Equal("+62-21-3456789", *prof.Fax)
	s.Require().NotNil(prof.KelurahanID)
	s.Equal(kel.ID, *prof.KelurahanID)
}

func (s *RefFeatureSuite) TestParokiMember_UmatAdminDelegationAndDualRoles() {
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{
		Kode: "kaj-del",
		Nama: "KAJ Delegation",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "paroki-del",
		Nama:        "Paroki Delegation",
	})
	s.Created()
	paroki := testutil.DecodeAs[domain.ParokiResponse](s).Data

	// Add Alice as paroki-admin AND umat-admin (dual role).
	umatTrue := true
	umatFalse := false
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email:     "alice-del@paroki.org",
		Name:      "Alice",
		Role:      "admin",
		UmatAdmin: &umatTrue,
	})
	s.Created()
	alice := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.True(alice.UmatAdmin)
	s.Equal("admin", alice.Role)

	hasParokiAdmin, err := s.enforcer.HasRoleInDomain(authz.UserSubject(alice.UserID), ref.RoleParokiAdmin, authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.True(hasParokiAdmin)
	hasUmatAdmin, err := s.enforcer.HasRoleInDomain(authz.UserSubject(alice.UserID), "umat:admin", authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.True(hasUmatAdmin)

	// Toggle umatAdmin off on Alice -> preserves ref:paroki-admin.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+alice.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		UmatAdmin: &umatFalse,
	})
	s.OK()
	aliceUpdated := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.False(aliceUpdated.UmatAdmin)
	s.Equal("admin", aliceUpdated.Role)
	hasParokiAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(alice.UserID), ref.RoleParokiAdmin, authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.True(hasParokiAdmin)
	hasUmatAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(alice.UserID), "umat:admin", authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.False(hasUmatAdmin)

	// Empty PATCH (both Role and UmatAdmin nil) -> 422 VALIDATION_ERROR.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+alice.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{})
	s.UnprocessableEntity()

	// Add Bob as member + umatAdmin=true.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email:     "bob-del@paroki.org",
		Name:      "Bob",
		Role:      "member",
		UmatAdmin: &umatTrue,
	})
	s.Created()
	bob := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.True(bob.UmatAdmin)

	// Bob (umat:admin only) can search, add member with role=member & umatAdmin=true, and toggle umatAdmin.
	s.loginViaOTP("bob-del@paroki.org")
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/search", http.MethodPost, domain.SearchParokiMembersRequest{
		Filter: &domain.ParokiMemberSearchFilter{UmatAdmin: &umatTrue},
	})
	s.OK()
	filtered := testutil.DecodeAs[[]domain.ParokiMemberResponse](s).Data
	s.Len(filtered, 1)
	s.Equal(bob.UserID, filtered[0].UserID)

	// Bob cannot add a member with role=admin -> 403 REF_FORBIDDEN.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "carol-del@paroki.org",
		Name:  "Carol",
		Role:  "admin",
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Bob CAN add Carol as member with umatAdmin=true.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email:     "carol-del@paroki.org",
		Name:      "Carol",
		Role:      "member",
		UmatAdmin: &umatTrue,
	})
	s.Created()
	carol := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.True(carol.UmatAdmin)

	// Bob cannot promote Carol to role=admin -> 403 REF_FORBIDDEN.
	adminRole := "admin"
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+carol.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		Role: &adminRole,
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Bob CAN revoke Carol's umatAdmin.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+carol.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		UmatAdmin: &umatFalse,
	})
	s.OK()
	carolUpdated := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.False(carolUpdated.UmatAdmin)

	// Bob cannot remove members -> 403 REF_FORBIDDEN.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+carol.UserID.String(), http.MethodDelete, nil)
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Alice (paroki-admin) removes Bob -> revokes umat:admin in Casbin.
	s.loginViaOTP("alice-del@paroki.org")
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+bob.UserID.String(), http.MethodDelete, nil)
	s.NoContent()
	hasUmatAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), "umat:admin", authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.False(hasUmatAdmin)

	// HasParokiMember check.
	isMember, err := s.refFeature.HasParokiMember(context.Background(), paroki.ID, alice.UserID)
	s.Require().NoError(err)
	s.True(isMember)
	isMember, err = s.refFeature.HasParokiMember(context.Background(), paroki.ID, bob.UserID)
	s.Require().NoError(err)
	s.False(isMember)
}

func (s *RefFeatureSuite) TestParokiMember_LiturgiAdminDelegationAndRefInUse() {
	s.loginViaOTP("admin@paroki.org")
	ctx := context.Background()

	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kaj-lit", Nama: "KAJ Liturgi"})
	s.Created()
	keuskupan := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "paroki-lit",
		Nama:        "Paroki Liturgi",
	})
	s.Created()
	paroki := testutil.DecodeAs[domain.ParokiResponse](s).Data

	// Add Alice as paroki member with umatAdmin=true (not paroki-admin).
	umatTrue := true
	litTrue := true
	litFalse := false
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email:     "alice-umat@paroki.org",
		Name:      "Alice Umat",
		Role:      "member",
		UmatAdmin: &umatTrue,
	})
	s.Created()
	alice := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.False(alice.LiturgiAdmin)

	// Super-admin adds Bob with liturgiAdmin=true -> Casbin liturgi:admin granted.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email:        "bob-lit@paroki.org",
		Name:         "Bob Liturgi",
		Role:         "member",
		LiturgiAdmin: &litTrue,
	})
	s.Created()
	bob := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.True(bob.LiturgiAdmin)

	hasLiturgiAdmin, err := s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), "liturgi:admin", authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.True(hasLiturgiAdmin)

	// Filter search by liturgiAdmin=true returns Bob only.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/search", http.MethodPost, domain.SearchParokiMembersRequest{
		Filter: &domain.ParokiMemberSearchFilter{LiturgiAdmin: &litTrue},
	})
	s.OK()
	filtered := testutil.DecodeAs[[]domain.ParokiMemberResponse](s).Data
	s.Len(filtered, 1)
	s.Equal(bob.UserID, filtered[0].UserID)

	// UmatAdmin-only caller (Alice) CANNOT grant or revoke liturgiAdmin -> 403 REF_FORBIDDEN.
	s.loginViaOTP("alice-umat@paroki.org")
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email:        "carol-lit@paroki.org",
		Name:         "Carol Liturgi",
		Role:         "member",
		LiturgiAdmin: &litTrue,
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		LiturgiAdmin: &litFalse,
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Super-admin revokes Bob's liturgiAdmin -> Casbin liturgi:admin revoked.
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		LiturgiAdmin: &litFalse,
	})
	s.OK()
	bobUpdated := testutil.DecodeAs[domain.ParokiMemberResponse](s).Data
	s.False(bobUpdated.LiturgiAdmin)

	hasLiturgiAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), "liturgi:admin", authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.False(hasLiturgiAdmin)

	// Re-grant liturgiAdmin and remove Bob -> Casbin liturgi:admin revoked.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		LiturgiAdmin: &litTrue,
	})
	s.OK()
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members/"+bob.UserID.String(), http.MethodDelete, nil)
	s.NoContent()
	hasLiturgiAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), "liturgi:admin", authz.ParokiDomain(paroki.ID))
	s.Require().NoError(err)
	s.False(hasLiturgiAdmin)

	// HasParokiChildren blocks deleting paroki when liturgi_ministries row exists -> 409 REF_IN_USE.
	ministry, err := s.client.LiturgiMinistry.Create().
		SetParokiID(paroki.ID).
		SetRoleSlug("lektor").
		Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/ref/paroki/"+paroki.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	s.Require().NoError(s.client.LiturgiMinistry.DeleteOneID(ministry.ID).Exec(ctx))

	// HasParokiChildren also blocks deleting paroki when umat_referensi row exists -> 409 REF_IN_USE.
	umatRef, err := s.client.UmatReferensi.Create().
		SetParokiID(paroki.ID).
		SetCategory("agama").
		SetCode("katolik").
		SetLabel("Katolik").
		Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/ref/paroki/"+paroki.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	s.Require().NoError(s.client.UmatReferensi.DeleteOneID(umatRef.ID).Exec(ctx))
	s.Request("/api/ref/paroki/"+paroki.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *RefFeatureSuite) TestRefDeleteGuards_UmatReferences() {
	s.loginViaOTP("admin@paroki.org")
	ctx := context.Background()

	// Setup keuskupan, kota, kecamatan, kelurahan, paroki, wilayah, lingkungan.
	s.Request("/api/ref/keuskupan", http.MethodPost, domain.CreateKeuskupanRequest{Kode: "kaj-guard", Nama: "KAJ Guard"})
	s.Created()
	keuskupan := testutil.DecodeAs[domain.KeuskupanResponse](s).Data

	s.Request("/api/ref/kota", http.MethodPost, domain.CreateKotaRequest{Kode: "jkt-guard", Nama: "Jakarta Guard"})
	s.Created()
	kota := testutil.DecodeAs[domain.KotaResponse](s).Data

	s.Request("/api/ref/kecamatan", http.MethodPost, domain.CreateKecamatanRequest{KotaID: kota.ID, Kode: "kec-guard", Nama: "Kecamatan Guard"})
	s.Created()
	kec := testutil.DecodeAs[domain.KecamatanResponse](s).Data

	s.Request("/api/ref/kelurahan", http.MethodPost, domain.CreateKelurahanRequest{KecamatanID: kec.ID, Kode: "kel-guard", Nama: "Kelurahan Guard"})
	s.Created()
	kel := testutil.DecodeAs[domain.KelurahanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, domain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "paroki-guard",
		Nama:        "Paroki Guard",
	})
	s.Created()
	paroki := testutil.DecodeAs[domain.ParokiResponse](s).Data

	// Add a paroki-admin to manage wilayah/lingkungan.
	s.Request("/api/ref/paroki/"+paroki.ID.String()+"/members", http.MethodPost, domain.AddParokiMemberRequest{
		Email: "padmin-guard@paroki.org",
		Name:  "Paroki Admin Guard",
		Role:  "admin",
	})
	s.Created()

	s.loginViaOTP("padmin-guard@paroki.org")
	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	defer s.SetHeader("X-Paroki-Id", "")

	s.Request("/api/ref/wilayah", http.MethodPost, domain.CreateWilayahRequest{Jenis: "wilayah", Kode: "w1", Nama: "Wilayah 1"})
	s.Created()
	wil := testutil.DecodeAs[domain.WilayahResponse](s).Data

	s.Request("/api/ref/lingkungan", http.MethodPost, domain.CreateLingkunganRequest{WilayahID: wil.ID, Kode: "l1", Nama: "Lingkungan 1"})
	s.Created()
	ling := testutil.DecodeAs[domain.LingkunganResponse](s).Data

	// Create Keluarga + Umat + UmatSakramen referencing these ref entities.
	keluargaRow, err := s.client.Keluarga.Create().
		SetParokiID(paroki.ID).
		SetNomorRegistrasi("KK-2026-000001").
		SetWilayahID(wil.ID).
		SetLingkunganID(ling.ID).
		SetKotaID(kota.ID).
		SetKecamatanID(kec.ID).
		SetKelurahanID(kel.ID).
		SetNamaKeluarga("Keluarga Guard").
		Save(ctx)
	s.Require().NoError(err)

	umatRow, err := s.client.Umat.Create().
		SetParokiID(paroki.ID).
		SetNomorInduk("NIU-2026-000001").
		SetKeluargaID(keluargaRow.ID).
		SetWilayahID(wil.ID).
		SetLingkunganID(ling.ID).
		SetKotaID(kota.ID).
		SetKecamatanID(kec.ID).
		SetKelurahanID(kel.ID).
		SetNama("Umat Guard").
		SetJenisKelamin(umat.JenisKelaminPria).
		Save(ctx)
	s.Require().NoError(err)

	sakramenRow, err := s.client.UmatSakramen.Create().
		SetParokiID(paroki.ID).
		SetUmatID(umatRow.ID).
		SetJenis(umatsakramen.JenisBaptis).
		SetKeuskupanID(keuskupan.ID).
		SetParokiSakramenID(paroki.ID).
		Save(ctx)
	s.Require().NoError(err)

	// All referenced entities must fail delete with 409 REF_IN_USE.
	for _, path := range []string{
		"/api/ref/lingkungan/" + ling.ID.String(),
		"/api/ref/wilayah/" + wil.ID.String(),
	} {
		s.Request(path, http.MethodDelete, nil)
		s.Conflict()
		s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)
	}
	s.loginViaOTP("admin@paroki.org")
	for _, path := range []string{
		"/api/ref/kelurahan/" + kel.ID.String(),
		"/api/ref/kecamatan/" + kec.ID.String(),
		"/api/ref/kota/" + kota.ID.String(),
		"/api/ref/paroki/" + paroki.ID.String(),
		"/api/ref/keuskupan/" + keuskupan.ID.String(),
	} {
		s.Request(path, http.MethodDelete, nil)
		s.Conflict()
		s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)
	}

	// Clean up UmatSakramen and soft-delete Umat + Keluarga -> deletes succeed.
	s.Require().NoError(s.client.UmatSakramen.DeleteOneID(sakramenRow.ID).Exec(ctx))
	now := time.Now().UTC()
	s.Require().NoError(s.client.Umat.UpdateOneID(umatRow.ID).SetDeletedAt(now).Exec(ctx))
	s.Require().NoError(s.client.Keluarga.UpdateOneID(keluargaRow.ID).SetDeletedAt(now).Exec(ctx))

	s.loginViaOTP("padmin-guard@paroki.org")
	for _, path := range []string{
		"/api/ref/lingkungan/" + ling.ID.String(),
		"/api/ref/wilayah/" + wil.ID.String(),
	} {
		s.Request(path, http.MethodDelete, nil)
		s.NoContent()
	}
	s.loginViaOTP("admin@paroki.org")
	for _, path := range []string{
		"/api/ref/kelurahan/" + kel.ID.String(),
		"/api/ref/kecamatan/" + kec.ID.String(),
		"/api/ref/kota/" + kota.ID.String(),
		"/api/ref/paroki/" + paroki.ID.String(),
		"/api/ref/keuskupan/" + keuskupan.ID.String(),
	} {
		s.Request(path, http.MethodDelete, nil)
		s.NoContent()
	}
}
