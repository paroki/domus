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

func (s *RefFeatureSuite) TestDioceseCRUD_ReactivationAndDeleteGuard() {
	s.loginViaOTP("admin@paroki.org")

	// Create Diocese with mixed-case code and local phone.
	phone := "0812-3456-7890"
	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{
		Code:  "  KAJ  ",
		Name:  "Keuskupan Agung Jakarta",
		Phone: &phone,
	})
	s.Created()
	created := testutil.DecodeAs[domain.DioceseResponse](s).Data
	s.Equal("kaj", created.Code)
	s.Require().NotNil(created.Phone)
	s.Equal("+6281234567890", *created.Phone)

	// Duplicate code returns 409 REF_CODE_CONFLICT.
	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{
		Code: "kaj",
		Name: "Duplicate",
	})
	s.Conflict()
	errResp := s.GetErrorResponse()
	s.Equal("REF_CODE_CONFLICT", errResp.Error.Code)

	// Search & Get.
	s.Request("/api/ref/dioceses/search", http.MethodPost, domain.SearchDiocesesRequest{Q: "Jakarta"})
	s.OK()
	list := testutil.DecodeAs[[]domain.DioceseResponse](s).Data
	s.Len(list, 1)

	s.Request("/api/ref/dioceses/"+created.ID.String(), http.MethodGet, nil)
	s.OK()

	// Delete & Reactivate.
	s.Request("/api/ref/dioceses/"+created.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{
		Code: "kaj",
		Name: "KAJ Reactivated",
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.DioceseResponse](s).Data
	s.Equal(created.ID, reactivated.ID)
	s.Equal("KAJ Reactivated", reactivated.Name)
	s.True(reactivated.Active)

	// Create Parish under Diocese -> DELETE Diocese blocked with 409 REF_IN_USE.
	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{
		DioceseID: reactivated.ID,
		Code:      "p-katedral",
		Name:      "Paroki Katedral",
	})
	s.Created()

	s.Request("/api/ref/dioceses/"+reactivated.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Regular user can read but cannot write.
	_, err := s.userFeature.EnsureUser(context.Background(), "viewer@paroki.org", "Viewer", "", "email", "viewer@paroki.org")
	s.Require().NoError(err)
	s.loginViaOTP("viewer@paroki.org")

	s.Request("/api/ref/dioceses/"+reactivated.ID.String(), http.MethodGet, nil)
	s.OK()
	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{
		Code: "kas",
		Name: "Semarang",
	})
	s.Forbidden()
}

func (s *RefFeatureSuite) TestParishCRUD_CurrentAndProfileAndMemberOnly() {
	s.loginViaOTP("admin@paroki.org")

	// Current parish without tenant returns 400 TENANT_REQUIRED.
	s.Request("/api/ref/parishes/current", http.MethodGet, nil)
	s.BadRequest()

	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{
		Code: "kaj",
		Name: "KAJ",
	})
	s.Created()
	d := testutil.DecodeAs[domain.DioceseResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{
		DioceseID: d.ID,
		Code:      "p-01",
		Name:      "Paroki Satu",
	})
	s.Created()
	p1 := testutil.DecodeAs[domain.ParishResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{
		DioceseID: d.ID,
		Code:      "p-02",
		Name:      "Paroki Dua",
	})
	s.Created()

	// Invite alice@paroki.org as parish admin of p1.
	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParishMemberRequest{
		Email: "alice@paroki.org",
		Name:  "Alice",
		Role:  "admin",
	})
	s.Created()
	aliceMember := testutil.DecodeAs[domain.ParishMemberResponse](s).Data
	s.Equal("admin", aliceMember.Role)
	s.Equal("alice@paroki.org", aliceMember.Email)

	// Alice logs in -> GET /api/me includes ref:parish-admin in roles.
	s.loginViaOTP("alice@paroki.org")
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[userdomain.UserResponse](s).Data
	s.Require().NotNil(me.ActiveParishID)
	s.Equal(p1.ID, *me.ActiveParishID)
	s.Contains(me.Roles, ref.RoleParishAdmin)

	// Alice can GET/PATCH /current.
	s.Request("/api/ref/parishes/current", http.MethodGet, nil)
	s.OK()

	newName := "Paroki Satu Updated"
	s.Request("/api/ref/parishes/current", http.MethodPatch, domain.UpdateParishProfileRequest{
		Name: &newName,
	})
	s.OK()
	updated := testutil.DecodeAs[domain.ParishResponse](s).Data
	s.Equal(newName, updated.Name)

	// Parish admin cannot change code on PATCH /:id.
	newCode := "p-01-hacked"
	s.Request("/api/ref/parishes/"+p1.ID.String(), http.MethodPatch, domain.UpdateParishRequest{
		Code: &newCode,
	})
	s.Forbidden()
	s.Equal("REF_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// memberOnly search returns only p1 for alice.
	memberOnly := true
	s.Request("/api/ref/parishes/search", http.MethodPost, domain.SearchParishesRequest{
		Filter: &domain.ParishSearchFilter{MemberOnly: &memberOnly},
	})
	s.OK()
	myParishes := testutil.DecodeAs[[]domain.ParishResponse](s).Data
	s.Len(myParishes, 1)
	s.Equal(p1.ID, myParishes[0].ID)

	// Zone under p1 blocks DELETE /api/ref/parishes/:id with 409 REF_IN_USE.
	_, err := s.client.Zone.Create().SetParishID(p1.ID).SetCode("z-1").SetName("Wilayah 1").Save(context.Background())
	s.Require().NoError(err)

	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/parishes/"+p1.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)
}

func (s *RefFeatureSuite) TestParishDelete_BlockedByAgentSessionOrMemory() {
	s.loginViaOTP("admin@paroki.org")
	ctx := context.Background()

	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{Code: "kaj", Name: "KAJ"})
	s.Created()
	d := testutil.DecodeAs[domain.DioceseResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{DioceseID: d.ID, Code: "p-agent", Name: "Paroki Agent"})
	s.Created()
	p := testutil.DecodeAs[domain.ParishResponse](s).Data

	userID := uuid.New()
	sess, err := s.client.AgentSession.Create().SetParishID(p.ID).SetUserID(userID).Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/ref/parishes/"+p.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	_, err = s.client.AgentSession.UpdateOneID(sess.ID).SetDeletedAt(time.Now().UTC()).Save(ctx)
	s.Require().NoError(err)

	mem, err := s.client.AgentMemory.Create().SetParishID(p.ID).SetUserID(userID).SetContent("FAQ").Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/ref/parishes/"+p.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	_, err = s.client.AgentMemory.UpdateOneID(mem.ID).SetDeletedAt(time.Now().UTC()).Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/ref/parishes/"+p.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *RefFeatureSuite) TestParishMember_CasbinSyncAndLastAdminGuard() {
	s.loginViaOTP("admin@paroki.org")

	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{Code: "kaj", Name: "KAJ"})
	s.Created()
	d := testutil.DecodeAs[domain.DioceseResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{DioceseID: d.ID, Code: "p-01", Name: "Paroki Satu"})
	s.Created()
	p1 := testutil.DecodeAs[domain.ParishResponse](s).Data

	// Add Alice as admin, Bob as member.
	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParishMemberRequest{
		Email: "alice@paroki.org",
		Role:  "admin",
	})
	s.Created()
	alice := testutil.DecodeAs[domain.ParishMemberResponse](s).Data

	// Duplicate add returns 409 REF_PARISH_MEMBER_ALREADY_EXISTS.
	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParishMemberRequest{
		UserID: &alice.UserID,
		Role:   "member",
	})
	s.Conflict()
	s.Equal("REF_PARISH_MEMBER_ALREADY_EXISTS", s.GetErrorResponse().Error.Code)

	// Last admin cannot be demoted or removed.
	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members/"+alice.UserID.String(), http.MethodPatch, domain.UpdateParishMemberRequest{
		Role: "member",
	})
	s.Conflict()
	s.Equal("REF_LAST_PARISH_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members/"+alice.UserID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_LAST_PARISH_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	// Alice adds Bob, promotes Bob to admin, demotes Bob, removes Bob.
	s.loginViaOTP("alice@paroki.org")
	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParishMemberRequest{
		Email: "bob@paroki.org",
		Name:  "Bob",
		Role:  "member",
	})
	s.Created()
	bob := testutil.DecodeAs[domain.ParishMemberResponse](s).Data

	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParishMemberRequest{
		Role: "admin",
	})
	s.OK()
	hasAdmin, err := s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), ref.RoleParishAdmin, authz.ParishDomain(p1.ID))
	s.Require().NoError(err)
	s.True(hasAdmin)

	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodPatch, domain.UpdateParishMemberRequest{
		Role: "member",
	})
	s.OK()
	hasAdmin, err = s.enforcer.HasRoleInDomain(authz.UserSubject(bob.UserID), ref.RoleParishAdmin, authz.ParishDomain(p1.ID))
	s.Require().NoError(err)
	s.False(hasAdmin)

	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members/"+bob.UserID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Adding an existing user by email with empty Name preserves their existing name.
	existingCarol, err := s.userFeature.EnsureUser(context.Background(), "carol@paroki.org", "Carol Original", "", "email", "carol@paroki.org")
	s.Require().NoError(err)
	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParishMemberRequest{
		Email: "carol@paroki.org",
		Role:  "member",
	})
	s.Created()
	carolMember := testutil.DecodeAs[domain.ParishMemberResponse](s).Data
	s.Equal(existingCarol.ID, carolMember.UserID)
	s.Equal("Carol Original", carolMember.Name)
}

func (s *RefFeatureSuite) TestFacadeHelpers() {
	ctx := context.Background()
	d, err := s.client.Diocese.Create().SetCode("kaj").SetName("KAJ").Save(ctx)
	s.Require().NoError(err)
	p1, err := s.client.Parish.Create().SetDioceseID(d.ID).SetCode("p1").SetName("P1").Save(ctx)
	s.Require().NoError(err)
	p2, err := s.client.Parish.Create().SetDioceseID(d.ID).SetCode("p2").SetName("P2").Save(ctx)
	s.Require().NoError(err)

	z1, err := s.client.Zone.Create().SetParishID(p1.ID).SetCode("z1").SetName("Z1").Save(ctx)
	s.Require().NoError(err)
	w1, err := s.client.Ward.Create().SetParishID(p1.ID).SetZoneID(z1.ID).SetCode("w1").SetName("W1").Save(ctx)
	s.Require().NoError(err)
	per1, err := s.client.Period.Create().SetParishID(p1.ID).SetCode("fy26").SetName("FY26").SetStartDate("2026-01-01").SetEndDate("2026-12-31").Save(ctx)
	s.Require().NoError(err)

	gotP, err := s.refFeature.GetParish(ctx, p1.ID)
	s.Require().NoError(err)
	s.Equal(p1.ID, gotP.ID)

	gotZ, err := s.refFeature.ValidateZoneInParish(ctx, p1.ID, z1.ID)
	s.Require().NoError(err)
	s.Equal(z1.ID, gotZ.ID)

	_, err = s.refFeature.ValidateZoneInParish(ctx, p2.ID, z1.ID)
	s.ErrorIs(err, domain.ErrZoneNotFound)

	gotW, err := s.refFeature.ValidateWardInParish(ctx, p1.ID, w1.ID)
	s.Require().NoError(err)
	s.Equal(w1.ID, gotW.ID)

	gotPer, err := s.refFeature.ValidatePeriodInParish(ctx, p1.ID, per1.ID)
	s.Require().NoError(err)
	s.Equal(per1.ID, gotPer.ID)

	u, err := s.userFeature.EnsureUser(ctx, "member@paroki.org", "Member", "", "email", "member@paroki.org")
	s.Require().NoError(err)
	pm1, err := s.refFeature.EnsureParishMember(ctx, p1.ID, u.ID)
	s.Require().NoError(err)
	pm2, err := s.refFeature.EnsureParishMember(ctx, p1.ID, u.ID)
	s.Require().NoError(err)
	s.Equal(pm1.ID, pm2.ID)
	s.NotEqual(uuid.Nil, pm1.ID)
}

func (s *RefFeatureSuite) setupTwoParishesWithAliceAdmin() (domain.ParishResponse, domain.ParishResponse) {
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/dioceses", http.MethodPost, domain.CreateDioceseRequest{Code: "kaj", Name: "KAJ"})
	s.Created()
	d := testutil.DecodeAs[domain.DioceseResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{DioceseID: d.ID, Code: "p-01", Name: "Paroki 1"})
	s.Created()
	p1 := testutil.DecodeAs[domain.ParishResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, domain.CreateParishRequest{DioceseID: d.ID, Code: "p-02", Name: "Paroki 2"})
	s.Created()
	p2 := testutil.DecodeAs[domain.ParishResponse](s).Data

	s.Request("/api/ref/parishes/"+p1.ID.String()+"/members", http.MethodPost, domain.AddParishMemberRequest{
		Email: "alice@paroki.org",
		Name:  "Alice",
		Role:  "admin",
	})
	s.Created()
	return p1, p2
}

func (s *RefFeatureSuite) TestZoneAndWardCRUD_ReactivationIsolationAndInUse() {
	p1, p2 := s.setupTwoParishesWithAliceAdmin()

	// Missing tenant returns 400 TENANT_REQUIRED.
	s.Request("/api/ref/zones/search", http.MethodPost, domain.SearchZonesRequest{})
	s.BadRequest()

	// Super-admin without ref:parish-admin in p1 gets 403 on write.
	s.SetHeader("X-Parish-Id", p1.ID.String())
	s.Request("/api/ref/zones", http.MethodPost, domain.CreateZoneRequest{Code: "z-super", Name: "Denied"})
	s.Forbidden()

	// Alice (parish-admin of p1) creates Zone with E.164 phone normalization.
	s.loginViaOTP("alice@paroki.org")
	coordPhone := "0811-2233-4455"
	s.Request("/api/ref/zones", http.MethodPost, domain.CreateZoneRequest{
		Code:             "WIL-01",
		Name:             "Wilayah 1",
		CoordinatorPhone: &coordPhone,
	})
	s.Created()
	z1 := testutil.DecodeAs[domain.ZoneResponse](s).Data
	s.Equal("wil-01", z1.Code)
	s.Equal("zone", z1.Type)
	s.Require().NotNil(z1.CoordinatorPhone)
	s.Equal("+6281122334455", *z1.CoordinatorPhone)

	// Delete & reactivate Zone.
	s.Request("/api/ref/zones/"+z1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/zones", http.MethodPost, domain.CreateZoneRequest{
		Type: "station",
		Code: "wil-01",
		Name: "Stasi 1",
	})
	s.Created()
	zReactivated := testutil.DecodeAs[domain.ZoneResponse](s).Data
	s.Equal(z1.ID, zReactivated.ID)
	s.Equal("station", zReactivated.Type)

	// Create Ward under Zone -> Delete Zone blocked with 409 REF_IN_USE.
	s.Request("/api/ref/wards", http.MethodPost, domain.CreateWardRequest{
		ZoneID: z1.ID,
		Code:   "LING-01",
		Name:   "Lingkungan 1",
	})
	s.Created()
	w1 := testutil.DecodeAs[domain.WardResponse](s).Data

	s.Request("/api/ref/zones/"+z1.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	// Delete Ward -> Delete Zone succeeds.
	s.Request("/api/ref/wards/"+w1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/ref/zones/"+z1.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Cross-tenant isolation: zone in p2 is 404 when active parish is p1.
	z2, err := s.client.Zone.Create().SetParishID(p2.ID).SetCode("z-p2").SetName("Zone P2").Save(context.Background())
	s.Require().NoError(err)
	s.Request("/api/ref/zones/"+z2.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Request("/api/ref/wards", http.MethodPost, domain.CreateWardRequest{
		ZoneID: z2.ID,
		Code:   "w-x",
		Name:   "Cross Tenant",
	})
	s.NotFound()
}

func (s *RefFeatureSuite) TestLookup_HybridGlobalAndParishOverride() {
	p1, _ := s.setupTwoParishesWithAliceAdmin()

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
	s.Nil(globalRow.ParishID)
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

	// Alice creates parish override for (sacrament, baptism).
	s.Request("/api/ref/lookups", http.MethodPost, domain.CreateLookupRequest{
		Category:  "sacrament",
		Code:      "baptism",
		Label:     "Baptis Paroki",
		SortOrder: 1,
	})
	s.Created()
	parishRow := testutil.DecodeAs[domain.LookupResponse](s).Data
	s.Require().NotNil(parishRow.ParishID)
	s.Equal(p1.ID, *parishRow.ParishID)

	// Search scope=all merges and overrides global with parish row.
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
	s.setupTwoParishesWithAliceAdmin()
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
	s.setupTwoParishesWithAliceAdmin()
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
