package ref_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
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
	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+alice.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		Role: "member",
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
		Role: "admin",
	})
	s.OK()
	hasAdmin, err := s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), ref.RoleParokiAdmin, authz.ParokiDomain(p1.ID))
	s.Require().NoError(err)
	s.True(hasAdmin)

	s.Request("/api/ref/paroki/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParokiMemberRequest{
		Role: "member",
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
	per1, err := s.client.Period.Create().SetParokiID(p1.ID).SetCode("fy26").SetName("FY26").SetStartDate("2026-01-01").SetEndDate("2026-12-31").Save(ctx)
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

	gotPer, err := s.refFeature.ValidatePeriodInParoki(ctx, p1.ID, per1.ID)
	s.Require().NoError(err)
	s.Equal(per1.ID, gotPer.ID)

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
	p1, p2 := s.setupTwoParokiWithAliceAdmin()

	// Missing tenant returns 400 TENANT_REQUIRED.
	s.Request("/api/ref/wilayah/search", http.MethodPost, domain.SearchWilayahRequest{})
	s.BadRequest()

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

func (s *RefFeatureSuite) TestPeriod_OverlapValidationAndIsCurrentRotation() {
	s.setupTwoParokiWithAliceAdmin()
	s.loginViaOTP("alice@paroki.org")

	// Invalid date range returns 422.
	s.Request("/api/ref/periods", http.MethodPost, domain.CreatePeriodRequest{
		Code:      "bad",
		Name:      "Bad",
		StartDate: "2026-12-31",
		EndDate:   "2026-01-01",
	})
	s.UnprocessableEntity()

	// Create FY2026 as current.
	s.Request("/api/ref/periods", http.MethodPost, domain.CreatePeriodRequest{
		Type:      "fiscal",
		Code:      "FY2026",
		Name:      "Fiscal 2026",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
		IsCurrent: true,
	})
	s.Created()
	fy26 := testutil.DecodeAs[domain.PeriodResponse](s).Data
	s.True(fy26.IsCurrent)

	// Overlapping fiscal period returns 409 REF_PERIOD_OVERLAP.
	s.Request("/api/ref/periods", http.MethodPost, domain.CreatePeriodRequest{
		Type:      "fiscal",
		Code:      "FY2026-H2",
		Name:      "Overlap",
		StartDate: "2026-06-01",
		EndDate:   "2027-05-31",
	})
	s.Conflict()
	s.Equal("REF_PERIOD_OVERLAP", s.GetErrorResponse().Error.Code)

	// Same dates with type=liturgical succeeds.
	s.Request("/api/ref/periods", http.MethodPost, domain.CreatePeriodRequest{
		Type:      "liturgical",
		Code:      "LIT-2026",
		Name:      "Liturgical 2026",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
		IsCurrent: true,
	})
	s.Created()

	// Create FY2027 with isCurrent=true -> clears FY2026.isCurrent.
	s.Request("/api/ref/periods", http.MethodPost, domain.CreatePeriodRequest{
		Type:      "fiscal",
		Code:      "FY2027",
		Name:      "Fiscal 2027",
		StartDate: "2027-01-01",
		EndDate:   "2027-12-31",
		IsCurrent: true,
	})
	s.Created()
	fy27 := testutil.DecodeAs[domain.PeriodResponse](s).Data
	s.True(fy27.IsCurrent)

	s.Request("/api/ref/periods/"+fy26.ID.String(), http.MethodGet, nil)
	s.OK()
	got26 := testutil.DecodeAs[domain.PeriodResponse](s).Data
	s.False(got26.IsCurrent)

	// PATCH FY2026 isCurrent=true -> clears FY2027.isCurrent.
	tVal := true
	s.Request("/api/ref/periods/"+fy26.ID.String(), http.MethodPatch, domain.UpdatePeriodRequest{IsCurrent: &tVal})
	s.OK()
	s.Request("/api/ref/periods/"+fy27.ID.String(), http.MethodGet, nil)
	s.OK()
	got27 := testutil.DecodeAs[domain.PeriodResponse](s).Data
	s.False(got27.IsCurrent)
}
