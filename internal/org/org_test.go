package org_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orginvite"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref"
	refdomain "github.com/paroki/domus/internal/ref/domain"
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

type OrgFeatureSuite struct {
	testutil.ApiTestSuite[any]
	client      *ent.Client
	enforcer    *authz.Enforcer
	sender      *stubOTPSender
	userFeature *user.Feature
	refFeature  *ref.Feature
	orgFeature  *org.Feature
}

func TestOrgFeatureSuite(t *testing.T) {
	suite.Run(t, new(OrgFeatureSuite))
}

func (s *OrgFeatureSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.sender = &stubOTPSender{}

	userCfg := user.Config{
		SuperAdminEmail: "super@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     60 * time.Second,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	orgCfg := org.Config{
		InviteDefaultTTL: 24 * time.Hour,
		InviteMaxTTL:     168 * time.Hour,
	}

	s.userFeature = user.New(s.client, s.enforcer, userCfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.sender))
	s.refFeature = ref.New(s.client, s.enforcer, s.userFeature)
	s.orgFeature = org.New(s.client, s.enforcer, s.userFeature, s.refFeature, orgCfg)
	s.userFeature.SetInviteJoiner(s.orgFeature)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature, s.orgFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature, s.orgFeature)
}

func (s *OrgFeatureSuite) loginViaOTP(email string) {
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

func (s *OrgFeatureSuite) seedParishWithAdmin(parishAdminEmail string) (parishID, zoneID, wardID, periodID, parishAdminUserID uuid.UUID) {
	s.loginViaOTP("super@paroki.org")

	s.Request("/api/ref/dioceses", http.MethodPost, refdomain.CreateDioceseRequest{
		Code: "kaj",
		Name: "Keuskupan Agung Jakarta",
	})
	s.Created()
	diocese := testutil.DecodeAs[refdomain.DioceseResponse](s).Data

	s.Request("/api/ref/parishes", http.MethodPost, refdomain.CreateParishRequest{
		DioceseID: diocese.ID,
		Code:      "p-kaj-01",
		Name:      "Paroki Katedral",
	})
	s.Created()
	parish := testutil.DecodeAs[refdomain.ParishResponse](s).Data
	parishID = parish.ID

	s.Request(fmt.Sprintf("/api/ref/parishes/%s/members", parishID), http.MethodPost, refdomain.AddParishMemberRequest{
		Email: parishAdminEmail,
		Name:  "Parish Admin",
		Role:  "admin",
	})
	s.Created()
	pm := testutil.DecodeAs[refdomain.ParishMemberResponse](s).Data
	parishAdminUserID = pm.UserID

	s.loginViaOTP(parishAdminEmail)

	s.Request("/api/ref/zones", http.MethodPost, refdomain.CreateZoneRequest{
		Code: "wil-01",
		Name: "Wilayah 1",
	})
	s.Created()
	zoneID = testutil.DecodeAs[refdomain.ZoneResponse](s).Data.ID

	s.Request("/api/ref/wards", http.MethodPost, refdomain.CreateWardRequest{
		ZoneID: zoneID,
		Code:   "ling-01",
		Name:   "Lingkungan 1",
	})
	s.Created()
	wardID = testutil.DecodeAs[refdomain.WardResponse](s).Data.ID

	s.Request("/api/ref/periods", http.MethodPost, refdomain.CreatePeriodRequest{
		Type:      "fiscal",
		Code:      "2026",
		Name:      "Tahun 2026",
		StartDate: "2026-01-01",
		EndDate:   "2026-12-31",
	})
	s.Created()
	periodID = testutil.DecodeAs[refdomain.PeriodResponse](s).Data.ID

	return parishID, zoneID, wardID, periodID, parishAdminUserID
}

func (s *OrgFeatureSuite) TestOrgCRUD_SearchGetUpdateDeleteAndReactivation() {
	parishID, zoneID, wardID, _, _ := s.seedParishWithAdmin("padmin@paroki.org")

	// Regular parish member (not ref:parish-admin) cannot create org.
	plainUser, err := s.userFeature.EnsureUser(context.Background(), "plain@paroki.org", "Plain User", "", "email", "plain@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParishMember(context.Background(), parishID, plainUser.ID)
	s.Require().NoError(err)
	s.loginViaOTP("plain@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Code: "sec-1",
		Name: "Seksi 1",
	})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Parish admin creates parent org with initialAdminEmail onboarding.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:              "dpp",
		Code:              "  DPP-01  ",
		Name:              "Dewan Pastoral Paroki",
		ZoneID:            &zoneID,
		WardID:            &wardID,
		InitialAdminEmail: "orgadmin@paroki.org",
		InitialAdminName:  "Org Admin",
	})
	s.Created()
	parent := testutil.DecodeAs[domain.OrganizationResponse](s).Data
	s.Equal("dpp-01", parent.Code)
	s.Equal("dpp", parent.Kind)
	s.Nil(parent.MyRole)

	// Duplicate active code returns 409 ORG_CODE_CONFLICT.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Code: "dpp-01",
		Name: "Dup",
	})
	s.Conflict()
	s.Equal("ORG_CODE_CONFLICT", s.GetErrorResponse().Error.Code)

	// Create child org under parent.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		ParentID: &parent.ID,
		Kind:     "section",
		Code:     "sek-lit",
		Name:     "Seksi Liturgi",
	})
	s.Created()
	child := testutil.DecodeAs[domain.OrganizationResponse](s).Data
	s.Require().NotNil(child.ParentID)
	s.Equal(parent.ID, *child.ParentID)

	// Deleting parent with active child returns 409 ORG_IN_USE.
	s.Request(fmt.Sprintf("/api/orgs/%s", parent.ID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("ORG_IN_USE", s.GetErrorResponse().Error.Code)

	// org:admin (not parish-admin) can GET (with myRole="admin"), PATCH, and search with myOrgsOnly=true.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s", parent.ID), http.MethodGet, nil)
	s.OK()
	gotParent := testutil.DecodeAs[domain.OrganizationResponse](s).Data
	s.Require().NotNil(gotParent.MyRole)
	s.Equal("admin", *gotParent.MyRole)

	newName := "DPP Inti"
	s.Request(fmt.Sprintf("/api/orgs/%s", parent.ID), http.MethodPatch, domain.UpdateOrganizationRequest{
		Name: &newName,
	})
	s.OK()
	s.Equal("DPP Inti", testutil.DecodeAs[domain.OrganizationResponse](s).Data.Name)

	myOnly := true
	s.Request("/api/orgs/search", http.MethodPost, domain.SearchOrganizationsRequest{
		Filter: &domain.OrganizationSearchFilter{MyOrgsOnly: &myOnly},
	})
	s.OK()
	searchRes := testutil.DecodeAs[[]domain.OrganizationResponse](s).Data
	s.Len(searchRes, 1)
	s.Equal(parent.ID, searchRes[0].ID)

	// Parish admin deletes child then parent -> 204, then reactivates code "sek-lit".
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s", child.ID), http.MethodDelete, nil)
	s.NoContent()
	s.Request(fmt.Sprintf("/api/orgs/%s", parent.ID), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind: "section",
		Code: "sek-lit",
		Name: "Seksi Liturgi Baru",
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.OrganizationResponse](s).Data
	s.Equal(child.ID, reactivated.ID)
	s.Equal("Seksi Liturgi Baru", reactivated.Name)
}

func (s *OrgFeatureSuite) TestOrgMemberCRUD_LeastPrivilegeCasbinAndLastAdminGuard() {
	parishID, _, _, periodID, parishAdminID := s.seedParishWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())

	// Create org without initial admin.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind: "section",
		Code: "sek-kom",
		Name: "Seksi Komsos",
	})
	s.Created()
	orgRow := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	// Parish admin is NOT an org member yet -> cannot view roster (403 ORG_FORBIDDEN).
	s.Request(fmt.Sprintf("/api/orgs/%s/members/search", orgRow.ID), http.MethodPost, domain.SearchOrgMembersRequest{})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Parish admin cannot appoint "member" without being org:admin (403 ORG_FORBIDDEN).
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		Email: "member1@paroki.org",
		Name:  "Member 1",
		Role:  "member",
	})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Parish admin self-assigns as org:admin -> 201.
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		UserID: &parishAdminID,
		Role:   "admin",
	})
	s.Created()

	// Duplicate add by org:admin returns 409 ORG_MEMBER_ALREADY_EXISTS.
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		UserID: &parishAdminID,
		Role:   "admin",
	})
	s.Conflict()
	s.Equal("ORG_MEMBER_ALREADY_EXISTS", s.GetErrorResponse().Error.Code)

	// Now parish admin IS org:admin -> can search roster and onboard a member with existing user name preservation.
	_, err := s.userFeature.EnsureUser(context.Background(), "member1@paroki.org", "Original Name", "", "email", "member1@paroki.org")
	s.Require().NoError(err)

	pos := "Sekretaris"
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		Email:    "member1@paroki.org",
		Name:     "Ignored Override Name",
		Role:     "member",
		Position: &pos,
		PeriodID: &periodID,
	})
	s.Created()
	mem1 := testutil.DecodeAs[domain.OrgMemberResponse](s).Data
	s.Equal("Original Name", mem1.Name)
	s.Equal("member", mem1.Role)

	s.Request(fmt.Sprintf("/api/orgs/%s/members/search", orgRow.ID), http.MethodPost, domain.SearchOrgMembersRequest{
		Q: "Sekretaris",
	})
	s.OK()
	s.Len(testutil.DecodeAs[[]domain.OrgMemberResponse](s).Data, 1)

	// Last admin guard on PATCH (demote or deactivate) and DELETE.
	memberRole := "member"
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, parishAdminID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Role: &memberRole,
	})
	s.Conflict()
	s.Equal("ORG_LAST_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	inactiveStatus := "inactive"
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, parishAdminID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Status: &inactiveStatus,
	})
	s.Conflict()
	s.Equal("ORG_LAST_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, parishAdminID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("ORG_LAST_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	// Clear position & periodId, and set member1 inactive -> Casbin roles revoked.
	emptyPos := ""
	nilPeriod := uuid.Nil
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, mem1.UserID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Position: &emptyPos,
		PeriodID: &nilPeriod,
		Status:   &inactiveStatus,
	})
	s.OK()
	updatedMem1 := testutil.DecodeAs[domain.OrgMemberResponse](s).Data
	s.Nil(updatedMem1.Position)
	s.Nil(updatedMem1.PeriodID)
	s.Equal("inactive", updatedMem1.Status)

	hasMemRole, err := s.enforcer.HasRoleInDomain(authz.UserSubject(mem1.UserID), org.RoleMember, authz.OrgDomain(orgRow.ID))
	s.Require().NoError(err)
	s.False(hasMemRole)

	// Reactivate member1 and test self-leave.
	activeStatus := "active"
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, mem1.UserID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Status: &activeStatus,
	})
	s.OK()

	s.loginViaOTP("member1@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, mem1.UserID), http.MethodDelete, nil)
	s.NoContent()
}

func (s *OrgFeatureSuite) TestOrgFacadeHelpersAndRefDeleteGuards() {
	parishID, zoneID, wardID, _, parishAdminID := s.seedParishWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())

	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:               "ward",
		Code:               "org-ling-01",
		Name:               "Pengurus Lingkungan 1",
		ZoneID:             &zoneID,
		WardID:             &wardID,
		InitialAdminUserID: &parishAdminID,
	})
	s.Created()
	orgRow := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	ctx := context.Background()
	got, err := s.orgFeature.GetOrganization(ctx, parishID, orgRow.ID)
	s.Require().NoError(err)
	s.Equal(orgRow.ID, got.ID)

	isMem, err := s.orgFeature.IsMember(ctx, orgRow.ID, parishAdminID)
	s.Require().NoError(err)
	s.True(isMem)

	isAdm, err := s.orgFeature.IsAdmin(ctx, orgRow.ID, parishAdminID)
	s.Require().NoError(err)
	s.True(isAdm)

	list, err := s.orgFeature.ListUserOrgs(ctx, parishID, parishAdminID)
	s.Require().NoError(err)
	s.Len(list, 1)
	s.Equal(orgRow.ID, list[0].ID)

	// Ward & Zone cannot be deleted while referenced by non-deleted Organization.
	s.Request(fmt.Sprintf("/api/ref/wards/%s", wardID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	s.Request(fmt.Sprintf("/api/ref/zones/%s", zoneID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)
}

type stubOAuthProvider struct {
	info userdomain.OAuthUserInfo
}

func (p *stubOAuthProvider) Name() string { return "stub" }

func (p *stubOAuthProvider) AuthCodeURL(state, redirectURI string) string {
	return fmt.Sprintf("https://stub.example/auth?state=%s&redirect_uri=%s", url.QueryEscape(state), url.QueryEscape(redirectURI))
}

func (p *stubOAuthProvider) Exchange(_ context.Context, _, _ string) (userdomain.OAuthUserInfo, error) {
	return p.info, nil
}

func (s *OrgFeatureSuite) TestOrgInvites_LifecycleRedeemAndGuestOnboarding() {
	parishID, _, _, _, parishAdminID := s.seedParishWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())

	// Create org with separate initial org:admin.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "sek-omk",
		Name:              "Seksi OMK",
		InitialAdminEmail: "omkadmin@paroki.org",
		InitialAdminName:  "OMK Admin",
	})
	s.Created()
	orgRow := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	// Non-member ref:parish-admin CANNOT create, search, or revoke invites (403 ORG_FORBIDDEN).
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	s.Request(fmt.Sprintf("/api/orgs/%s/invites/search", orgRow.ID), http.MethodPost, domain.SearchOrgInvitesRequest{})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Login as org:admin.
	s.loginViaOTP("omkadmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())

	// TTL exceeding InviteMaxTTL (168h) returns 422 VALIDATION_ERROR.
	tooLongTTL := 169 * 3600
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		TTLSeconds: &tooLongTTL,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Create single-use member invite.
	maxOne := 1
	pos := "Anggota"
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		Role:            "member",
		DefaultPosition: &pos,
		MaxUses:         &maxOne,
	})
	s.Created()
	inv1 := testutil.DecodeAs[domain.OrgInviteResponse](s).Data
	s.Len(inv1.Code, 12)
	s.Equal("member", inv1.Role)
	s.Equal(0, inv1.UsedCount)

	// Create revocable invite & revoke it.
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{})
	s.Created()
	invRevoked := testutil.DecodeAs[domain.OrgInviteResponse](s).Data
	s.Request(fmt.Sprintf("/api/orgs/%s/invites/%s", orgRow.ID, invRevoked.ID), http.MethodDelete, nil)
	s.NoContent()

	// Search invites filtered by status.
	s.Request(fmt.Sprintf("/api/orgs/%s/invites/search", orgRow.ID), http.MethodPost, domain.SearchOrgInvitesRequest{
		Filter: &domain.OrgInviteSearchFilter{Status: "active"},
	})
	s.OK()
	s.Len(testutil.DecodeAs[[]domain.OrgInviteResponse](s).Data, 1)

	// Public preview works without auth for active and revoked invites.
	s.ClearCookies()
	s.SetHeader("X-Parish-Id", "")
	s.Request(fmt.Sprintf("/api/orgs/invites/%s", invRevoked.Code), http.MethodGet, nil)
	s.OK()
	revPreview := testutil.DecodeAs[domain.OrgInvitePreviewResponse](s).Data
	s.False(revPreview.Valid)
	s.Equal("revoked", revPreview.Status)

	s.Request(fmt.Sprintf("/api/orgs/invites/%s", inv1.Code), http.MethodGet, nil)
	s.OK()
	actPreview := testutil.DecodeAs[domain.OrgInvitePreviewResponse](s).Data
	s.True(actPreview.Valid)
	s.Equal("active", actPreview.Status)
	s.Equal(orgRow.ID, actPreview.OrgID)
	s.Equal(parishID, actPreview.ParishID)

	// Unauthenticated guest OTP onboarding with inv1 (unregistered user).
	s.Request("/api/auth/otp", http.MethodPost, userdomain.RequestOTPRequest{
		Email:      "guest1@paroki.org",
		InviteCode: inv1.Code,
	})
	s.OK()
	s.Request("/api/auth/otp/verify", http.MethodPost, userdomain.VerifyOTPRequest{
		Email:      "guest1@paroki.org",
		Code:       s.sender.lastCode,
		InviteCode: inv1.Code,
		Name:       "Guest Satu",
	})
	s.OK()
	otpSess := testutil.DecodeAs[userdomain.AuthSessionResponse](s).Data
	s.Require().NotNil(otpSess.ActiveParishID)
	s.Equal(parishID, *otpSess.ActiveParishID)
	s.Cookies = s.HttpResponse.Cookies()

	// Second user trying to redeem exhausted inv1 gets 409 ORG_INVITE_MAX_USES_REACHED.
	s.loginViaOTP("padmin@paroki.org")
	s.Request(fmt.Sprintf("/api/orgs/invites/%s/redeem", inv1.Code), http.MethodPost, nil)
	s.Conflict()
	s.Equal("ORG_INVITE_MAX_USES_REACHED", s.GetErrorResponse().Error.Code)

	// Preview of inv1 now shows exhausted & valid=false.
	s.ClearCookies()
	s.Request(fmt.Sprintf("/api/orgs/invites/%s", inv1.Code), http.MethodGet, nil)
	s.OK()
	exhPreview := testutil.DecodeAs[domain.OrgInvitePreviewResponse](s).Data
	s.False(exhPreview.Valid)
	s.Equal("exhausted", exhPreview.Status)

	// Idempotent redeem by an already-active member does NOT increment used_count.
	s.loginViaOTP("omkadmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	maxTwo := 2
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		Role:    "member",
		MaxUses: &maxTwo,
	})
	s.Created()
	invMulti := testutil.DecodeAs[domain.OrgInviteResponse](s).Data

	s.loginViaOTP("guest1@paroki.org")
	s.Request(fmt.Sprintf("/api/orgs/invites/%s/redeem", invMulti.Code), http.MethodPost, nil)
	s.OK()
	redeemRes := testutil.DecodeAs[domain.RedeemOrgInviteResponse](s).Data
	s.Equal(parishID, redeemRes.ParishID)
	s.Equal("member", redeemRes.Membership.Role)

	s.loginViaOTP("omkadmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/invites/search", orgRow.ID), http.MethodPost, domain.SearchOrgInvitesRequest{})
	s.OK()
	for _, it := range testutil.DecodeAs[[]domain.OrgInviteResponse](s).Data {
		if it.ID == invMulti.ID {
			s.Equal(0, it.UsedCount)
		}
	}

	// Expired invite returns 410 ORG_INVITE_EXPIRED.
	expiredInv, err := s.client.OrgInvite.Create().
		SetParishID(parishID).
		SetOrgID(orgRow.ID).
		SetCode("EXPIRED12345").
		SetRole(orginvite.RoleMember).
		SetExpiresAt(time.Now().UTC().Add(-1 * time.Hour)).
		SetCreatedBy(parishAdminID).
		Save(context.Background())
	s.Require().NoError(err)

	s.Request("/api/auth/otp", http.MethodPost, userdomain.RequestOTPRequest{
		Email:      "guest2@paroki.org",
		InviteCode: expiredInv.Code,
	})
	s.Equal(http.StatusGone, s.HttpResponse.StatusCode)
	s.Equal("ORG_INVITE_EXPIRED", s.GetErrorResponse().Error.Code)

	// OAuth guest onboarding with inviteCode.
	s.loginViaOTP("omkadmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		Role: "admin",
	})
	s.Created()
	invOAuth := testutil.DecodeAs[domain.OrgInviteResponse](s).Data

	s.userFeature.RegisterOAuthProvider(&stubOAuthProvider{
		info: userdomain.OAuthUserInfo{
			Subject:       "stub-sub-1",
			Email:         "oauthguest@paroki.org",
			EmailVerified: true,
			Name:          "OAuth Guest",
		},
	})

	s.ClearCookies()
	s.SetHeader("X-Parish-Id", "")
	s.Request(fmt.Sprintf("/api/auth/stub?redirect=/org/join/%s&inviteCode=%s", invOAuth.Code, invOAuth.Code), http.MethodGet, nil)
	s.Equal(http.StatusFound, s.HttpResponse.StatusCode)
	s.Cookies = s.HttpResponse.Cookies()

	var stateVal string
	for _, ck := range s.Cookies {
		if ck.Name == "domus_oauth_state" {
			stateVal = ck.Value
		}
	}
	s.Require().NotEmpty(stateVal)

	s.Request(fmt.Sprintf("/api/auth/stub/callback?code=authcode&state=%s", url.QueryEscape(stateVal)), http.MethodGet, nil)
	s.Equal(http.StatusFound, s.HttpResponse.StatusCode)
	s.Cookies = s.HttpResponse.Cookies()

	// Verify oauthguest is now an active org:admin.
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[userdomain.UserResponse](s).Data
	isAdm, err := s.orgFeature.IsAdmin(context.Background(), orgRow.ID, me.ID)
	s.Require().NoError(err)
	s.True(isAdm)
}
