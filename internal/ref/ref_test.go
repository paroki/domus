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
