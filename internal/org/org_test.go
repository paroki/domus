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
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/org/domain"
	orgrepo "github.com/paroki/domus/internal/org/repo"
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

func (s *OrgFeatureSuite) seedParokiWithAdmin(parokiAdminEmail string) (parokiID, wilayahID, lingkunganID, parokiAdminUserID uuid.UUID) {
	s.loginViaOTP("super@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "Keuskupan Agung Jakarta",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "p-kaj-01",
		Nama:        "Paroki Katedral",
	})
	s.Created()
	paroki := testutil.DecodeAs[refdomain.ParokiResponse](s).Data
	parokiID = paroki.ID

	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: parokiAdminEmail,
		Name:  "Paroki Admin",
		Role:  "admin",
	})
	s.Created()
	pm := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data
	parokiAdminUserID = pm.UserID

	s.loginViaOTP(parokiAdminEmail)

	s.Request("/api/ref/wilayah", http.MethodPost, refdomain.CreateWilayahRequest{
		Kode: "wil-01",
		Nama: "Wilayah 1",
	})
	s.Created()
	wilayahID = testutil.DecodeAs[refdomain.WilayahResponse](s).Data.ID

	s.Request("/api/ref/lingkungan", http.MethodPost, refdomain.CreateLingkunganRequest{
		WilayahID: wilayahID,
		Kode:      "ling-01",
		Nama:      "Lingkungan 1",
	})
	s.Created()
	lingkunganID = testutil.DecodeAs[refdomain.LingkunganResponse](s).Data.ID

	return parokiID, wilayahID, lingkunganID, parokiAdminUserID
}

func (s *OrgFeatureSuite) TestOrgCRUD_SearchGetUpdateDeleteAndReactivation() {
	parokiID, wilayahID, lingkunganID, _ := s.seedParokiWithAdmin("padmin@paroki.org")

	// Regular paroki member (not ref:paroki-admin) cannot create org.
	plainUser, err := s.userFeature.EnsureUser(context.Background(), "plain@paroki.org", "Plain User", "", "email", "plain@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(context.Background(), parokiID, plainUser.ID)
	s.Require().NoError(err)
	s.loginViaOTP("plain@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Code: "sec-1",
		Name: "Seksi 1",
	})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Paroki admin creates parent org with initialAdminEmail onboarding.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:              "dpp",
		Code:              "  DPP-01  ",
		Name:              "Dewan Pastoral Paroki",
		WilayahID:         &wilayahID,
		LingkunganID:      &lingkunganID,
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
	s.SetHeader("X-Paroki-Id", parokiID.String())
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
	s.SetHeader("X-Paroki-Id", parokiID.String())
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
	parokiID, _, _, parokiAdminID := s.seedParokiWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

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
		UserID: &parokiAdminID,
		Role:   "admin",
	})
	s.Created()

	// Duplicate add by org:admin returns 409 ORG_MEMBER_ALREADY_EXISTS.
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		UserID: &parokiAdminID,
		Role:   "admin",
	})
	s.Conflict()
	s.Equal("ORG_MEMBER_ALREADY_EXISTS", s.GetErrorResponse().Error.Code)

	// Now parish admin IS org:admin -> can search roster and onboard a member with existing user name preservation.
	_, err := s.userFeature.EnsureUser(context.Background(), "member1@paroki.org", "Original Name", "", "email", "member1@paroki.org")
	s.Require().NoError(err)

	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		Email: "member1@paroki.org",
		Name:  "Ignored Override Name",
		Role:  "member",
	})
	s.Created()
	mem1 := testutil.DecodeAs[domain.OrgMemberResponse](s).Data
	s.Equal("Original Name", mem1.Name)
	s.Equal("member", mem1.Role)

	s.Request(fmt.Sprintf("/api/orgs/%s/members/search", orgRow.ID), http.MethodPost, domain.SearchOrgMembersRequest{
		Q: "Original",
	})
	s.OK()
	s.Len(testutil.DecodeAs[[]domain.OrgMemberResponse](s).Data, 1)

	// Last admin guard on PATCH (demote or deactivate) and DELETE.
	memberRole := "member"
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, parokiAdminID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Role: &memberRole,
	})
	s.Conflict()
	s.Equal("ORG_LAST_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	inactiveStatus := "inactive"
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, parokiAdminID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Status: &inactiveStatus,
	})
	s.Conflict()
	s.Equal("ORG_LAST_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, parokiAdminID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("ORG_LAST_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	// Set member1 inactive -> Casbin roles revoked.
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, mem1.UserID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Status: &inactiveStatus,
	})
	s.OK()
	updatedMem1 := testutil.DecodeAs[domain.OrgMemberResponse](s).Data
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
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgRow.ID, mem1.UserID), http.MethodDelete, nil)
	s.NoContent()
}

func (s *OrgFeatureSuite) TestOrgFacadeHelpersAndRefDeleteGuards() {
	parokiID, wilayahID, lingkunganID, parokiAdminID := s.seedParokiWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:               "lingkungan",
		Code:               "org-ling-01",
		Name:               "Pengurus Lingkungan 1",
		WilayahID:          &wilayahID,
		LingkunganID:       &lingkunganID,
		InitialAdminUserID: &parokiAdminID,
	})
	s.Created()
	orgRow := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	ctx := context.Background()
	got, err := s.orgFeature.GetOrganization(ctx, parokiID, orgRow.ID)
	s.Require().NoError(err)
	s.Equal(orgRow.ID, got.ID)

	isMem, err := s.orgFeature.IsMember(ctx, orgRow.ID, parokiAdminID)
	s.Require().NoError(err)
	s.True(isMem)

	isAdm, err := s.orgFeature.IsAdmin(ctx, orgRow.ID, parokiAdminID)
	s.Require().NoError(err)
	s.True(isAdm)

	list, err := s.orgFeature.ListUserOrgs(ctx, parokiID, parokiAdminID)
	s.Require().NoError(err)
	s.Len(list, 1)
	s.Equal(orgRow.ID, list[0].ID)

	// Lingkungan & Wilayah cannot be deleted while referenced by non-deleted Organization.
	s.Request(fmt.Sprintf("/api/ref/lingkungan/%s", lingkunganID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("REF_IN_USE", s.GetErrorResponse().Error.Code)

	s.Request(fmt.Sprintf("/api/ref/wilayah/%s", wilayahID), http.MethodDelete, nil)
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
	parokiID, _, _, parokiAdminID := s.seedParokiWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

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
	s.SetHeader("X-Paroki-Id", parokiID.String())

	// TTL exceeding InviteMaxTTL (168h) returns 422 VALIDATION_ERROR.
	tooLongTTL := 169 * 3600
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		TTLSeconds: &tooLongTTL,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Create single-use member invite.
	maxOne := 1
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgRow.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		Role:    "member",
		MaxUses: &maxOne,
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
	s.SetHeader("X-Paroki-Id", "")
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
	s.Equal(parokiID, actPreview.ParokiID)

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
	s.Require().NotNil(otpSess.ActiveParokiID)
	s.Equal(parokiID, *otpSess.ActiveParokiID)
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
	s.SetHeader("X-Paroki-Id", parokiID.String())
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
	s.Equal(parokiID, redeemRes.ParokiID)
	s.Equal("member", redeemRes.Membership.Role)

	s.loginViaOTP("omkadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/invites/search", orgRow.ID), http.MethodPost, domain.SearchOrgInvitesRequest{})
	s.OK()
	for _, it := range testutil.DecodeAs[[]domain.OrgInviteResponse](s).Data {
		if it.ID == invMulti.ID {
			s.Equal(0, it.UsedCount)
		}
	}

	// Expired invite returns 410 ORG_INVITE_EXPIRED.
	expiredInv, err := s.client.OrgInvite.Create().
		SetParokiID(parokiID).
		SetOrgID(orgRow.ID).
		SetCode("EXPIRED12345").
		SetRole(orginvite.RoleMember).
		SetExpiresAt(time.Now().UTC().Add(-1 * time.Hour)).
		SetCreatedBy(parokiAdminID).
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
	s.SetHeader("X-Paroki-Id", parokiID.String())
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
	s.SetHeader("X-Paroki-Id", "")
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

func (s *OrgFeatureSuite) TestIssue94_ReactivationInactiveMemberCycleMaxUsesAndOTPPhone() {
	parokiID, _, _, parokiAdminID := s.seedParokiWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	ctx := context.Background()

	// Pre-create an existing user with a distinct display name.
	existingUser, err := s.userFeature.EnsureUser(ctx, "budi@paroki.org", "Budi Santoso", "", "email", "budi@paroki.org")
	s.Require().NoError(err)

	// Create org with InitialAdminEmail (omitting InitialAdminName) -> preserves Budi's name.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "org-cycle-a",
		Name:              "Org A",
		InitialAdminEmail: "budi@paroki.org",
	})
	s.Created()
	orgA := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	budiAfterCreate, err := s.client.User.Get(ctx, existingUser.ID)
	s.Require().NoError(err)
	s.Equal("Budi Santoso", budiAfterCreate.Name)

	// Demote & deactivate Budi's OrgMember row, then soft-delete Org A.
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgA.ID), http.MethodPost, domain.AddOrgMemberRequest{
		UserID: &parokiAdminID,
		Role:   "admin",
	})
	s.Created()
	memberRole := "member"
	inactiveStatus := "inactive"
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgA.ID, existingUser.ID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Role:   &memberRole,
		Status: &inactiveStatus,
	})
	s.OK()
	s.Request(fmt.Sprintf("/api/orgs/%s", orgA.ID), http.MethodDelete, nil)
	s.NoContent()

	// Reactivate Org A with InitialAdminEmail -> existing OrgMember row must become active admin.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "org-cycle-a",
		Name:              "Org A Reactivated",
		InitialAdminEmail: "budi@paroki.org",
	})
	s.Created()
	reactivatedA := testutil.DecodeAs[domain.OrganizationResponse](s).Data
	s.Equal(orgA.ID, reactivatedA.ID)

	budiMember, err := s.client.OrgMember.Query().
		Where(orgmember.OrgIDEQ(orgA.ID), orgmember.UserIDEQ(existingUser.ID)).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal(orgmember.RoleAdmin, budiMember.Role)
	s.Equal(orgmember.StatusActive, budiMember.Status)

	// Add a member, deactivate them, then re-add via Email (no Name) -> reactivates & preserves name.
	memUser, err := s.userFeature.EnsureUser(ctx, "siti@paroki.org", "Siti Rahayu", "", "email", "siti@paroki.org")
	s.Require().NoError(err)
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgA.ID), http.MethodPost, domain.AddOrgMemberRequest{
		UserID: &memUser.ID,
		Role:   "member",
	})
	s.Created()
	s.Request(fmt.Sprintf("/api/orgs/%s/members/%s", orgA.ID, memUser.ID), http.MethodPatch, domain.UpdateOrgMemberRequest{
		Status: &inactiveStatus,
	})
	s.OK()

	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgA.ID), http.MethodPost, domain.AddOrgMemberRequest{
		Email: "siti@paroki.org",
		Role:  "member",
	})
	s.Created()
	readded := testutil.DecodeAs[domain.OrgMemberResponse](s).Data
	s.Equal("active", readded.Status)
	sitiAfter, err := s.client.User.Get(ctx, memUser.ID)
	s.Require().NoError(err)
	s.Equal("Siti Rahayu", sitiAfter.Name)

	// Cyclic parentId rejection: A -> B -> C, then patching A.parentId = C fails with 422 INVALID_PARENT.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:     "section",
		Code:     "org-cycle-b",
		Name:     "Org B",
		ParentID: &orgA.ID,
	})
	s.Created()
	orgB := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:     "section",
		Code:     "org-cycle-c",
		Name:     "Org C",
		ParentID: &orgB.ID,
	})
	s.Created()
	orgC := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	s.Request(fmt.Sprintf("/api/orgs/%s", orgA.ID), http.MethodPatch, domain.UpdateOrganizationRequest{
		ParentID: &orgC.ID,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Atomic IncrementOrgInviteUsedCount max_uses boundary check.
	maxOne := 1
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgA.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		Role:    "member",
		MaxUses: &maxOne,
	})
	s.Created()
	inv := testutil.DecodeAs[domain.OrgInviteResponse](s).Data

	r := orgrepo.New(s.client)
	s.Require().NoError(r.IncrementOrgInviteUsedCount(ctx, inv.ID, &maxOne))
	err = r.IncrementOrgInviteUsedCount(ctx, inv.ID, &maxOne)
	s.Require().ErrorIs(err, domain.ErrInviteMaxUsesReached)

	// VerifyOTP guest invite onboarding normalizes and validates req.Phone.
	s.Request(fmt.Sprintf("/api/orgs/%s/invites", orgA.ID), http.MethodPost, domain.CreateOrgInviteRequest{
		Role: "member",
	})
	s.Created()
	openInv := testutil.DecodeAs[domain.OrgInviteResponse](s).Data

	s.ClearCookies()
	s.SetHeader("X-Paroki-Id", "")
	s.Request("/api/auth/otp", http.MethodPost, userdomain.RequestOTPRequest{
		Email:      "phoneguest@paroki.org",
		InviteCode: openInv.Code,
	})
	s.OK()

	badPhone := "not-a-phone"
	s.Request("/api/auth/otp/verify", http.MethodPost, userdomain.VerifyOTPRequest{
		Email:      "phoneguest@paroki.org",
		Code:       s.sender.lastCode,
		InviteCode: openInv.Code,
		Name:       "Phone Guest",
		Phone:      &badPhone,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	localPhone := "081234567890"
	s.Request("/api/auth/otp/verify", http.MethodPost, userdomain.VerifyOTPRequest{
		Email:      "phoneguest@paroki.org",
		Code:       s.sender.lastCode,
		InviteCode: openInv.Code,
		Name:       "Phone Guest",
		Phone:      &localPhone,
	})
	s.OK()
	sess := testutil.DecodeAs[userdomain.AuthSessionResponse](s).Data
	s.Require().NotNil(sess.User.Phone)
	s.Equal("+6281234567890", *sess.User.Phone)
}

func (s *OrgFeatureSuite) TestDeleteOrganization_InUseWithEvent() {
	parokiID, _, _, parokiAdminID := s.seedParokiWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind: "section",
		Code: "sek-evt",
		Name: "Seksi Acara",
	})
	s.Created()
	orgRow := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	ctx := context.Background()
	startAt := time.Now().UTC().Add(24 * time.Hour)
	ev, err := s.client.Event.Create().
		SetParokiID(parokiID).
		SetOrgID(orgRow.ID).
		SetCreatedBy(parokiAdminID).
		SetTitle("Rapat Persiapan").
		SetStartAt(startAt).
		SetEndAt(startAt.Add(time.Hour)).
		Save(ctx)
	s.Require().NoError(err)

	// Active Event blocks organization deletion with 409 ORG_IN_USE.
	s.Request(fmt.Sprintf("/api/orgs/%s", orgRow.ID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("ORG_IN_USE", s.GetErrorResponse().Error.Code)

	// Soft-deleted Event does not block organization deletion.
	err = s.client.Event.UpdateOneID(ev.ID).
		SetDeletedAt(time.Now().UTC()).
		Exec(ctx)
	s.Require().NoError(err)

	// Active EventPetugas referencing orgRow.ID on a non-deleted Event blocks deletion with 409 ORG_IN_USE.
	parishEv, err := s.client.Event.Create().
		SetParokiID(parokiID).
		SetCreatedBy(parokiAdminID).
		SetTitle("Misa Paroki").
		SetStartAt(startAt).
		SetEndAt(startAt.Add(time.Hour)).
		Save(ctx)
	s.Require().NoError(err)
	petugasRow, err := s.client.EventPetugas.Create().
		SetParokiID(parokiID).
		SetEvent(parishEv).
		SetRole("Kor").
		SetOrgID(orgRow.ID).
		SetName(orgRow.Name).
		Save(ctx)
	s.Require().NoError(err)

	s.Request(fmt.Sprintf("/api/orgs/%s", orgRow.ID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("ORG_IN_USE", s.GetErrorResponse().Error.Code)

	err = s.client.EventPetugas.DeleteOneID(petugasRow.ID).Exec(ctx)
	s.Require().NoError(err)

	s.Request(fmt.Sprintf("/api/orgs/%s", orgRow.ID), http.MethodDelete, nil)
	s.NoContent()
}

func (s *OrgFeatureSuite) TestOrgPeriode_CRUD_Overlap_IsCurrent_Reactivation_RBAC_AndInUse() {
	parokiID, _, _, _ := s.seedParokiWithAdmin("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	// Create org with separate initial org:admin.
	s.Request("/api/orgs", http.MethodPost, domain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "sek-liturgi",
		Name:              "Seksi Liturgi",
		InitialAdminEmail: "liturgiadmin@paroki.org",
		InitialAdminName:  "Liturgi Admin",
	})
	s.Created()
	orgRow := testutil.DecodeAs[domain.OrganizationResponse](s).Data

	// Create a non-member paroki user and an org:member.
	outsider, err := s.userFeature.EnsureUser(context.Background(), "outsider@paroki.org", "Outsider", "", "email", "outsider@paroki.org")
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(context.Background(), parokiID, outsider.ID)
	s.Require().NoError(err)

	s.loginViaOTP("liturgiadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgRow.ID), http.MethodPost, domain.AddOrgMemberRequest{
		Email: "member@paroki.org",
		Name:  "Liturgi Member",
		Role:  "member",
	})
	s.Created()

	// Invalid date range returns 422 VALIDATION_ERROR.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "bad-range",
		Nama:           "Bad Range",
		TanggalMulai:   "2026-12-31",
		TanggalSelesai: "2026-01-01",
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Create first period (2025-2026) as isCurrent=true.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "2025-2026",
		Nama:           "Periode 2025-2026",
		TanggalMulai:   "2025-01-01",
		TanggalSelesai: "2026-12-31",
		IsCurrent:      true,
	})
	s.Created()
	p1 := testutil.DecodeAs[domain.OrgPeriodeResponse](s).Data
	s.True(p1.IsCurrent)
	s.Equal("open", p1.Status)

	// Duplicate kode returns 409 ORG_CODE_CONFLICT.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "2025-2026",
		Nama:           "Duplicate",
		TanggalMulai:   "2027-01-01",
		TanggalSelesai: "2027-12-31",
	})
	s.Conflict()
	s.Equal("ORG_CODE_CONFLICT", s.GetErrorResponse().Error.Code)

	// Overlapping date range returns 409 ORG_PERIODE_OVERLAP.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "overlap-2026",
		Nama:           "Overlap",
		TanggalMulai:   "2026-06-01",
		TanggalSelesai: "2027-05-31",
	})
	s.Conflict()
	s.Equal("ORG_PERIODE_OVERLAP", s.GetErrorResponse().Error.Code)

	// Create second period (2027-2028) with isCurrent=true -> atomically clears p1.isCurrent.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "2027-2028",
		Nama:           "Periode 2027-2028",
		TanggalMulai:   "2027-01-01",
		TanggalSelesai: "2028-12-31",
		IsCurrent:      true,
	})
	s.Created()
	p2 := testutil.DecodeAs[domain.OrgPeriodeResponse](s).Data
	s.True(p2.IsCurrent)

	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p1.ID), http.MethodGet, nil)
	s.OK()
	s.False(testutil.DecodeAs[domain.OrgPeriodeResponse](s).Data.IsCurrent)

	// PATCH p1 isCurrent=true -> atomically clears p2.isCurrent; PATCH overlapping dates -> 409 ORG_PERIODE_OVERLAP.
	tVal := true
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p1.ID), http.MethodPatch, domain.UpdateOrgPeriodeRequest{
		IsCurrent: &tVal,
	})
	s.OK()
	s.True(testutil.DecodeAs[domain.OrgPeriodeResponse](s).Data.IsCurrent)

	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p2.ID), http.MethodGet, nil)
	s.OK()
	s.False(testutil.DecodeAs[domain.OrgPeriodeResponse](s).Data.IsCurrent)

	overlapEnd := "2027-06-01"
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p1.ID), http.MethodPatch, domain.UpdateOrgPeriodeRequest{
		TanggalSelesai: &overlapEnd,
	})
	s.Conflict()
	s.Equal("ORG_PERIODE_OVERLAP", s.GetErrorResponse().Error.Code)

	// Search ordered by tanggal_mulai DESC, id DESC.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/search", orgRow.ID), http.MethodPost, domain.SearchOrgPeriodeRequest{})
	s.OK()
	listed := testutil.DecodeAs[[]domain.OrgPeriodeResponse](s).Data
	s.Require().Len(listed, 2)
	s.Equal(p2.ID, listed[0].ID)
	s.Equal(p1.ID, listed[1].ID)

	// RBAC: org:member can read (search/get) but cannot write (403 ORG_FORBIDDEN).
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/search", orgRow.ID), http.MethodPost, domain.SearchOrgPeriodeRequest{})
	s.OK()
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p1.ID), http.MethodGet, nil)
	s.OK()
	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "2029-2030",
		Nama:           "Denied",
		TanggalMulai:   "2029-01-01",
		TanggalSelesai: "2030-12-31",
	})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// RBAC: non-member paroki user gets 403 ORG_FORBIDDEN on read & write.
	s.loginViaOTP("outsider@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/search", orgRow.ID), http.MethodPost, domain.SearchOrgPeriodeRequest{})
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p1.ID), http.MethodGet, nil)
	s.Forbidden()
	s.Equal("ORG_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Soft-delete p1 and reactivate with the same kode; ref:paroki-admin can also manage & delete org.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	// Organization with active OrgPeriode cannot be deleted (409 ORG_IN_USE).
	s.Request(fmt.Sprintf("/api/orgs/%s", orgRow.ID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("ORG_IN_USE", s.GetErrorResponse().Error.Code)

	// Soft-delete p1, then reactivate with same kode.
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p1.ID), http.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/api/orgs/%s/periode", orgRow.ID), http.MethodPost, domain.CreateOrgPeriodeRequest{
		Kode:           "2025-2026",
		Nama:           "Periode 2025-2026 Reactivated",
		TanggalMulai:   "2025-01-01",
		TanggalSelesai: "2026-12-31",
		Status:         "closed",
	})
	s.Created()
	reactivated := testutil.DecodeAs[domain.OrgPeriodeResponse](s).Data
	s.Equal(p1.ID, reactivated.ID)
	s.Equal("closed", reactivated.Status)
	s.False(reactivated.IsCurrent)

	// Delete both periods -> organization delete succeeds (204).
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, reactivated.ID), http.MethodDelete, nil)
	s.NoContent()
	s.Request(fmt.Sprintf("/api/orgs/%s/periode/%s", orgRow.ID, p2.ID), http.MethodDelete, nil)
	s.NoContent()
	s.Request(fmt.Sprintf("/api/orgs/%s", orgRow.ID), http.MethodDelete, nil)
	s.NoContent()
}
