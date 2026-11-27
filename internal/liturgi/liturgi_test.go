package liturgi_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/usernotificationlog"
	"github.com/paroki/domus/internal/cantus"
	cantusdomain "github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/event"
	eventdomain "github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/liturgi"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/org"
	orgdomain "github.com/paroki/domus/internal/org/domain"
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

type LiturgiFeatureSuite struct {
	testutil.ApiTestSuite[any]
	client         *ent.Client
	enforcer       *authz.Enforcer
	sender         *stubOTPSender
	userFeature    *user.Feature
	refFeature     *ref.Feature
	orgFeature     *org.Feature
	cantusFeature  *cantus.Feature
	eventFeature   *event.Feature
	liturgiFeature *liturgi.Feature
}

func TestLiturgiFeatureSuite(t *testing.T) {
	suite.Run(t, new(LiturgiFeatureSuite))
}

func (s *LiturgiFeatureSuite) SetupTest() {
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
	eventCfg := event.Config{
		DefaultVisibility: "public",
		RSVPGracePeriod:   30 * time.Minute,
	}

	s.userFeature = user.New(s.client, s.enforcer, userCfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.sender))
	s.refFeature = ref.New(s.client, s.enforcer, s.userFeature)
	s.orgFeature = org.New(s.client, s.enforcer, s.userFeature, s.refFeature, orgCfg)
	s.userFeature.SetInviteJoiner(s.orgFeature)
	s.cantusFeature = cantus.New(s.client, s.enforcer, s.userFeature)
	s.eventFeature = event.New(s.client, s.enforcer, s.userFeature, s.orgFeature, eventCfg, event.WithRef(s.refFeature))
	s.liturgiFeature = liturgi.New(s.client, s.enforcer, s.userFeature, s.orgFeature, s.eventFeature, liturgi.WithCantus(s.cantusFeature), liturgi.WithRef(s.refFeature))
	s.cantusFeature.SetInUseChecker(s.liturgiFeature.IsCantusInUse)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature, s.orgFeature, s.cantusFeature, s.eventFeature, s.liturgiFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature, s.orgFeature, s.cantusFeature, s.eventFeature, s.liturgiFeature)
}

func (s *LiturgiFeatureSuite) loginViaOTP(email string) {
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

type seededParish struct {
	parokiID       uuid.UUID
	orgID          uuid.UUID
	liturgiAdminID uuid.UUID
	leaderID       uuid.UUID
	orgAdminID     uuid.UUID
	memberID       uuid.UUID
}

func (s *LiturgiFeatureSuite) seedParish() seededParish {
	s.loginViaOTP("super@paroki.org")

	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "kaj-liturgi",
		Nama: "Keuskupan Agung Jakarta",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "p-lit-01",
		Nama:        "Paroki Katedral",
	})
	s.Created()
	parokiID := testutil.DecodeAs[refdomain.ParokiResponse](s).Data.ID

	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Paroki Admin",
		Role:  "admin",
	})
	s.Created()

	litTrue := true
	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email:        "litadmin@paroki.org",
		Name:         "Liturgi Admin",
		Role:         "member",
		LiturgiAdmin: &litTrue,
	})
	s.Created()
	liturgiAdminID := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data.UserID

	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "leader@paroki.org",
		Name:  "Koordinator Lektor",
		Role:  "member",
	})
	s.Created()
	leaderID := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data.UserID

	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "member@paroki.org",
		Name:  "Umat Biasa",
		Role:  "member",
	})
	s.Created()
	memberID := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data.UserID

	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	s.Request("/api/orgs", http.MethodPost, orgdomain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "sie-lit-kor",
		Name:              "Seksi Paduan Suara",
		InitialAdminEmail: "orgadmin@paroki.org",
		InitialAdminName:  "Org Admin Kor",
	})
	s.Created()
	orgID := testutil.DecodeAs[orgdomain.OrganizationResponse](s).Data.ID

	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members/search", orgID), http.MethodPost, nil)
	s.OK()
	members := testutil.DecodeAs[[]orgdomain.OrgMemberResponse](s).Data
	s.Require().Len(members, 1)
	orgAdminID := members[0].UserID

	return seededParish{
		parokiID:       parokiID,
		orgID:          orgID,
		liturgiAdminID: liturgiAdminID,
		leaderID:       leaderID,
		orgAdminID:     orgAdminID,
		memberID:       memberID,
	}
}

func (s *LiturgiFeatureSuite) TestMinistries_ListOrder_PatchRBAC_AndCanManage() {
	seed := s.seedParish()

	// 1. Any member sees all 6 canonical ministries in fixed order even when unconfigured.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/ministries", http.MethodGet, nil)
	s.OK()
	items := testutil.DecodeAs[[]domain.LiturgiMinistryResponse](s).Data
	s.Require().Len(items, 6)
	s.Equal(domain.MinistryRoles, []string{
		items[0].RoleSlug,
		items[1].RoleSlug,
		items[2].RoleSlug,
		items[3].RoleSlug,
		items[4].RoleSlug,
		items[5].RoleSlug,
	})
	for _, it := range items {
		s.False(it.CanManage)
		s.Nil(it.OrgID)
		s.Nil(it.LeaderUserID)
	}

	// 2. Regular member or orgadmin CANNOT PATCH ministry -> 403 LITURGI_FORBIDDEN.
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		LeaderUserID: &seed.leaderID,
	})
	s.Forbidden()
	s.Equal("LITURGI_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// 3. liturgi:admin validates roleSlug, orgId, leaderUserId -> 422 VALIDATION_ERROR.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())

	s.Request("/api/liturgi/ministries/invalid-role", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	badID := uuid.New()
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		OrgID: &badID,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		LeaderUserID: &badID,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// 4. liturgi:admin assigns leaderUserId to lektor and orgId to kor.
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		LeaderUserID: &seed.leaderID,
	})
	s.OK()
	lektorRes := testutil.DecodeAs[domain.LiturgiMinistryResponse](s).Data
	s.Require().NotNil(lektorRes.LeaderUserID)
	s.Equal(seed.leaderID, *lektorRes.LeaderUserID)
	s.Require().NotNil(lektorRes.LeaderName)
	s.Equal("Koordinator Lektor", *lektorRes.LeaderName)
	s.True(lektorRes.CanManage)

	s.Request("/api/liturgi/ministries/kor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		OrgID: &seed.orgID,
	})
	s.OK()
	korRes := testutil.DecodeAs[domain.LiturgiMinistryResponse](s).Data
	s.Require().NotNil(korRes.OrgID)
	s.Equal(seed.orgID, *korRes.OrgID)
	s.Require().NotNil(korRes.OrgName)
	s.Equal("Seksi Paduan Suara", *korRes.OrgName)

	// 5. leader@paroki.org sees canManage=true only for lektor.
	s.loginViaOTP("leader@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/ministries", http.MethodGet, nil)
	s.OK()
	leaderList := testutil.DecodeAs[[]domain.LiturgiMinistryResponse](s).Data
	s.True(leaderList[0].CanManage)  // lektor
	s.False(leaderList[5].CanManage) // kor

	// 6. orgadmin@paroki.org sees canManage=true only for kor.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/ministries", http.MethodGet, nil)
	s.OK()
	orgAdminList := testutil.DecodeAs[[]domain.LiturgiMinistryResponse](s).Data
	s.False(orgAdminList[0].CanManage) // lektor
	s.True(orgAdminList[5].CanManage)  // kor

	// 7. liturgi:admin clears lektor leaderUserId.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	clearTrue := true
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		ClearLeaderUserID: &clearTrue,
	})
	s.OK()
	cleared := testutil.DecodeAs[domain.LiturgiMinistryResponse](s).Data
	s.Nil(cleared.LeaderUserID)
	s.Nil(cleared.LeaderName)
}

func (s *LiturgiFeatureSuite) TestSchedules_CRUD_SearchFilters_AndEventRSVPCompat() {
	seed := s.seedParish()
	ctx := context.Background()
	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	endAt := startAt.Add(90 * time.Minute)

	// Configure kor org + lektor leader for manageableRoles & canManageLagu assertions.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		LeaderUserID: &seed.leaderID,
	})
	s.OK()
	s.Request("/api/liturgi/ministries/kor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		OrgID: &seed.orgID,
	})
	s.OK()

	// 1. Regular member CANNOT create schedule -> 403 LITURGI_FORBIDDEN.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules", http.MethodPost, domain.CreateLiturgiScheduleRequest{
		Title:   "Unauthorized Mass",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("LITURGI_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// 2. liturgi:admin invalid time range -> 400 LITURGI_INVALID_TIME_RANGE.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules", http.MethodPost, domain.CreateLiturgiScheduleRequest{
		Title:   "Bad Mass",
		StartAt: endAt,
		EndAt:   startAt,
	})
	s.BadRequest()
	s.Equal("LITURGI_INVALID_TIME_RANGE", s.GetErrorResponse().Error.Code)

	// 3. liturgi:admin creates parish Mass schedule -> stored in events (org_id=NULL, jenis="misa", public, rsvp_enabled=true).
	s.Request("/api/liturgi/schedules", http.MethodPost, domain.CreateLiturgiScheduleRequest{
		Title:       "Misa Minggu Pagi",
		Description: new("Misa pagi bahasa Indonesia"),
		Location:    new("Gereja Katedral"),
		StartAt:     startAt,
		EndAt:       endAt,
	})
	s.Created()
	sched := testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data
	s.Equal("Misa Minggu Pagi", sched.Title)
	s.Equal("public", sched.Visibility)
	s.True(sched.RsvpEnabled)
	s.Equal("scheduled", sched.Status)
	s.True(sched.CanManageSchedule)
	s.True(sched.CanManageLagu)
	s.Len(sched.ManageableRoles, 6)

	// 4. Non-misa parish event is NOT accessible via /api/liturgi/schedules/:eventId -> 404 LITURGI_SCHEDULE_NOT_FOUND.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/events", http.MethodPost, eventdomain.CreateEventRequest{
		Title:   "Rapat Dewan Paroki",
		Jenis:   new("rapat"),
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Created()
	rapatEv := testutil.DecodeAs[eventdomain.EventResponse](s).Data

	s.Request("/api/liturgi/schedules/"+rapatEv.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Equal("LITURGI_SCHEDULE_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// 5. Mass schedule appears in /api/events/search and supports /api/events/:id/rsvp.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/events/search", http.MethodPost, eventdomain.SearchEventsRequest{
		Filter: &eventdomain.EventSearchFilter{Jenis: "misa"},
	})
	s.OK()
	evList := testutil.DecodeAs[[]eventdomain.EventResponse](s).Data
	s.Require().Len(evList, 1)
	s.Equal(sched.ID, evList[0].ID)

	s.Request(fmt.Sprintf("/api/events/%s/rsvp", sched.ID), http.MethodPost, eventdomain.UpsertEventRSVPRequest{
		Status: "going",
	})
	s.OK()

	// 6. Seed petugas & lagu rows directly to verify GetSchedule & SearchSchedules enrichment + filters.
	_, err := s.client.LiturgiPetugas.Create().
		SetParokiID(seed.parokiID).
		SetEventID(sched.ID).
		SetRole("lektor").
		SetUserID(seed.memberID).
		SetName("Umat Biasa").
		SetUrutan(1).
		SetStatus("declined").
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.LiturgiPetugas.Create().
		SetParokiID(seed.parokiID).
		SetEventID(sched.ID).
		SetRole("kor").
		SetOrgID(seed.orgID).
		SetName("Seksi Paduan Suara").
		SetUrutan(1).
		SetStatus("assigned").
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.LiturgiLagu.Create().
		SetParokiID(seed.parokiID).
		SetEventID(sched.ID).
		SetUrutan(1).
		SetBagian("Pembuka").
		SetJudul("Mari Kita Bersukacita").
		Save(ctx)
	s.Require().NoError(err)

	// Member sees myRsvp, rsvpSummary, petugas[0].canConfirm=true, lagu, and canManageSchedule=false.
	s.Request("/api/liturgi/schedules/"+sched.ID.String(), http.MethodGet, nil)
	s.OK()
	detail := testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data
	s.False(detail.CanManageSchedule)
	s.False(detail.CanManageLagu)
	s.Empty(detail.ManageableRoles)
	s.Require().NotNil(detail.MyRsvp)
	s.Equal("going", detail.MyRsvp.Status)
	s.Require().NotNil(detail.RsvpSummary)
	s.Equal(1, detail.RsvpSummary.Going)
	s.Require().Len(detail.Petugas, 2)
	s.Equal("kor", detail.Petugas[0].Role)
	s.False(detail.Petugas[0].CanConfirm)
	s.Equal("lektor", detail.Petugas[1].Role)
	s.True(detail.Petugas[1].CanConfirm)
	s.Require().Len(detail.Lagu, 1)

	// orgadmin@paroki.org sees manageableRoles=["kor"] and canManageLagu=true.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+sched.ID.String(), http.MethodGet, nil)
	s.OK()
	orgView := testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data
	s.Equal([]string{"kor"}, orgView.ManageableRoles)
	s.True(orgView.CanManageLagu)
	s.True(orgView.Petugas[0].CanConfirm) // kor org admin

	// 7. SearchSchedules filters: assignedToMe, assignedOrgId, hasDeclined, role.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	tTrue := true
	s.Request("/api/liturgi/schedules/search", http.MethodPost, domain.SearchLiturgiSchedulesRequest{
		Filter: &domain.LiturgiScheduleSearchFilter{
			AssignedToMe: &tTrue,
			HasDeclined:  &tTrue,
			Role:         "lektor",
		},
	})
	s.OK()
	searchRes := testutil.DecodeAs[[]domain.LiturgiScheduleResponse](s).Data
	s.Require().Len(searchRes, 1)
	s.Equal(sched.ID, searchRes[0].ID)

	s.Request("/api/liturgi/schedules/search", http.MethodPost, domain.SearchLiturgiSchedulesRequest{
		Filter: &domain.LiturgiScheduleSearchFilter{
			AssignedOrgID: &seed.orgID,
			Role:          "mazmur",
		},
	})
	s.OK()
	s.Empty(testutil.DecodeAs[[]domain.LiturgiScheduleResponse](s).Data)

	// 8. liturgi:admin updates and deletes schedule.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+sched.ID.String(), http.MethodPatch, domain.UpdateLiturgiScheduleRequest{
		Title: new("Misa Minggu Sore"),
	})
	s.OK()
	s.Equal("Misa Minggu Sore", testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data.Title)

	s.Request("/api/liturgi/schedules/"+sched.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/liturgi/schedules/"+sched.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Equal("LITURGI_SCHEDULE_NOT_FOUND", s.GetErrorResponse().Error.Code)
}

func (s *LiturgiFeatureSuite) TestPetugasLifecycleAndPermissions() {
	seed := s.seedParish()

	// Create a second member in paroki for substitution testing.
	s.loginViaOTP("super@paroki.org")
	s.Request("/api/ref/paroki/"+seed.parokiID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "sub@paroki.org",
		Name:  "Pengganti Lektor",
		Role:  "member",
	})
	s.Created()
	subID := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data.UserID

	// Configure lektor ministry leader = leader@paroki.org.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		LeaderUserID: &seed.leaderID,
	})
	s.OK()

	// Create Mass schedule.
	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	endAt := startAt.Add(90 * time.Minute)
	s.Request("/api/liturgi/schedules", http.MethodPost, domain.CreateLiturgiScheduleRequest{
		Title:   "Misa Harian",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Created()
	schedID := testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data.ID

	// Lektor leader cannot manage misdinar -> 403 LITURGI_FORBIDDEN.
	s.loginViaOTP("leader@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role: "misdinar",
		Name: "Misdinar Tamu",
	})
	s.Forbidden()
	s.Equal("LITURGI_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Mutual exclusion: userId + orgId -> 422 VALIDATION_ERROR.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role:   "lektor",
		UserID: &seed.memberID,
		OrgID:  &seed.orgID,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// 1. User mode: auto-fills name from user.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role:   "lektor",
		UserID: &seed.memberID,
	})
	s.Created()
	pUser := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Equal("Umat Biasa", pUser.Name)
	s.Equal("assigned", pUser.Status)
	s.True(pUser.CanEdit)
	s.True(pUser.CanRespond)
	s.False(pUser.CanApproveSubstitute)

	// 2. Org mode: auto-fills name from org.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role:  "lektor",
		OrgID: &seed.orgID,
	})
	s.Created()
	pOrg := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Equal("Seksi Paduan Suara", pOrg.Name)

	// 3. Manual name mode.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role: "lektor",
		Name: "Lektor Tamu",
	})
	s.Created()
	pManual := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Equal("Lektor Tamu", pManual.Name)

	// Unrelated user cannot respond -> 403 LITURGI_FORBIDDEN.
	s.loginViaOTP("sub@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String()+"/respond", http.MethodPatch, domain.RespondLiturgiPetugasRequest{
		Status: "confirmed",
	})
	s.Forbidden()
	s.Equal("LITURGI_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Assigned user confirms -> respondedAt & respondedBy populated.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String()+"/respond", http.MethodPatch, domain.RespondLiturgiPetugasRequest{
		Status: "confirmed",
	})
	s.OK()
	confirmed := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Equal("confirmed", confirmed.Status)
	s.Require().NotNil(confirmed.RespondedAt)
	s.Require().NotNil(confirmed.RespondedBy)
	s.Equal(seed.memberID, *confirmed.RespondedBy)

	// Assigned user declines with mutually exclusive substituteUserId + substituteOrgId -> 422 MUTUALLY_EXCLUSIVE.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String()+"/respond", http.MethodPatch, domain.RespondLiturgiPetugasRequest{
		Status:           "declined",
		SubstituteUserID: &subID,
		SubstituteOrgID:  &seed.orgID,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Assigned user declines with substituteUserId -> auto-fills substituteName.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String()+"/respond", http.MethodPatch, domain.RespondLiturgiPetugasRequest{
		Status:           "declined",
		SubstituteUserID: &subID,
		Note:             new("Sedang di luar kota"),
	})
	s.OK()
	declined := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Equal("declined", declined.Status)
	s.Require().NotNil(declined.SubstituteUserID)
	s.Equal(subID, *declined.SubstituteUserID)
	s.Require().NotNil(declined.SubstituteName)
	s.Equal("Pengganti Lektor", *declined.SubstituteName)

	// Leader approves substitute with markConfirmed=true.
	s.loginViaOTP("leader@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String()+"/approve-substitute", http.MethodPost, domain.ApproveLiturgiSubstituteRequest{
		MarkConfirmed: true,
	})
	s.OK()
	approved := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Require().NotNil(approved.UserID)
	s.Equal(subID, *approved.UserID)
	s.Equal("Pengganti Lektor", approved.Name)
	s.Equal("confirmed", approved.Status)
	s.Nil(approved.SubstituteUserID)
	s.Nil(approved.SubstituteName)

	// Approve substitute when none exists -> 422 SUBSTITUTE_REQUIRED.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String()+"/approve-substitute", http.MethodPost, domain.ApproveLiturgiSubstituteRequest{})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Reassign via PATCH resets respondedAt/respondedBy and status to assigned.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pUser.ID.String(), http.MethodPatch, domain.UpdateLiturgiPetugasRequest{
		UserID: &seed.memberID,
	})
	s.OK()
	reassigned := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data
	s.Equal("assigned", reassigned.Status)
	s.Nil(reassigned.RespondedAt)
	s.Nil(reassigned.RespondedBy)
	s.Equal("Umat Biasa", reassigned.Name)

	// Delete manual assignment.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+pManual.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *LiturgiFeatureSuite) TestLaguCantusIntegrationAnd409InUse() {
	seed := s.seedParish()

	// Grant liturgiAdmin cantus:admin and create a Cantus entry.
	ctx := context.Background()
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(seed.liturgiAdminID), cantus.RoleAdmin, authz.DomainGlobal))
	s.loginViaOTP("litadmin@paroki.org")
	s.Request("/api/cantus", http.MethodPost, cantusdomain.CreateCantusRequest{
		Slug:      "ps-351-tuhan-kasihanilah",
		Jenis:     "lagu",
		Judul:     "Tuhan Kasihanilah Kami",
		Sumber:    new("Puji Syukur"),
		Nomor:     new("351"),
		NadaDasar: new("F"),
	})
	s.Created()
	cantusID := testutil.DecodeAs[cantusdomain.CantusResponse](s).Data.ID

	// Create schedule + assign member@paroki.org as lektor and another member as organis.
	s.loginViaOTP("super@paroki.org")
	s.Request("/api/ref/paroki/"+seed.parokiID.String()+"/members", http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "organis@paroki.org",
		Name:  "Organis Paroki",
		Role:  "member",
	})
	s.Created()
	organisID := testutil.DecodeAs[refdomain.ParokiMemberResponse](s).Data.UserID

	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	s.Request("/api/liturgi/schedules", http.MethodPost, domain.CreateLiturgiScheduleRequest{
		Title:   "Misa Minggu",
		StartAt: startAt,
		EndAt:   startAt.Add(time.Hour),
	})
	s.Created()
	schedID := testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data.ID

	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role:   "lektor",
		UserID: &seed.memberID,
	})
	s.Created()
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role:   "organis",
		UserID: &organisID,
	})
	s.Created()

	// Lektor officer cannot manage lagu -> 403 LITURGI_FORBIDDEN.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/lagu", http.MethodPost, domain.CreateLiturgiLaguRequest{
		Bagian:   "kyrie",
		CantusID: &cantusID,
	})
	s.Forbidden()
	s.Equal("LITURGI_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Assigned organis officer can add lagu with cantusId auto-fill.
	s.loginViaOTP("organis@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/lagu", http.MethodPost, domain.CreateLiturgiLaguRequest{
		Bagian:   "kyrie",
		CantusID: &cantusID,
		Urutan:   new(2),
	})
	s.Created()
	lagu := testutil.DecodeAs[domain.LiturgiLaguResponse](s).Data
	s.Equal("Tuhan Kasihanilah Kami", lagu.Judul)
	s.Require().NotNil(lagu.Sumber)
	s.Equal("Puji Syukur", *lagu.Sumber)
	s.Require().NotNil(lagu.Nomor)
	s.Equal("351", *lagu.Nomor)
	s.Require().NotNil(lagu.NadaDasar)
	s.Equal("F", *lagu.NadaDasar)

	// Deleting Cantus while referenced by active Mass schedule -> 409 CANTUS_IN_USE.
	s.loginViaOTP("litadmin@paroki.org")
	s.Request("/api/cantus/"+cantusID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("CANTUS_IN_USE", s.GetErrorResponse().Error.Code)

	// Delete the lagu entry, then Cantus delete succeeds.
	s.loginViaOTP("organis@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/lagu/"+lagu.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.loginViaOTP("litadmin@paroki.org")
	s.Request("/api/cantus/"+cantusID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *LiturgiFeatureSuite) TestLiturgiNotificationsAndReminders() {
	seed := s.seedParish()
	ctx := context.Background()

	// Configure lektor leader = leader@paroki.org.
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/ministries/lektor", http.MethodPatch, domain.UpdateLiturgiMinistryRequest{
		LeaderUserID: &seed.leaderID,
	})
	s.OK()

	jktLoc, err := time.LoadLocation("Asia/Jakarta")
	s.Require().NoError(err)
	// Mass starts at 10:00 WIB -> h-24h is 10:00 WIB (outside quiet hours), h-1h is 09:00 WIB (outside quiet hours).
	startAt := time.Date(2030, 6, 15, 10, 0, 0, 0, jktLoc).UTC()
	s.Request("/api/liturgi/schedules", http.MethodPost, domain.CreateLiturgiScheduleRequest{
		Title:   "Misa Kudus",
		StartAt: startAt,
		EndAt:   startAt.Add(time.Hour),
	})
	s.Created()
	schedID := testutil.DecodeAs[domain.LiturgiScheduleResponse](s).Data.ID

	// Assign member@paroki.org -> triggers liturgi_assigned.
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas", http.MethodPost, domain.CreateLiturgiPetugasRequest{
		Role:   "lektor",
		UserID: &seed.memberID,
	})
	s.Created()
	petugasID := testutil.DecodeAs[domain.LiturgiPetugasResponse](s).Data.ID

	assignedCount, err := s.client.UserNotificationLog.Query().
		Where(
			usernotificationlog.UserIDEQ(seed.memberID),
			usernotificationlog.KindEQ("liturgi_assigned"),
			usernotificationlog.RefIDEQ(schedID),
		).Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, assignedCount)

	// Member declines with substitute -> triggers liturgi_substitute_proposed to leader.
	s.loginViaOTP("member@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+petugasID.String()+"/respond", http.MethodPatch, domain.RespondLiturgiPetugasRequest{
		Status:           "declined",
		SubstituteUserID: &seed.orgAdminID,
	})
	s.OK()

	proposedCount, err := s.client.UserNotificationLog.Query().
		Where(
			usernotificationlog.UserIDEQ(seed.leaderID),
			usernotificationlog.KindEQ("liturgi_substitute_proposed"),
			usernotificationlog.RefIDEQ(schedID),
		).Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, proposedCount)

	// Leader approves substitute -> triggers liturgi_substitute_approved to orgAdminID and memberID.
	s.loginViaOTP("leader@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	s.Request("/api/liturgi/schedules/"+schedID.String()+"/petugas/"+petugasID.String()+"/approve-substitute", http.MethodPost, domain.ApproveLiturgiSubstituteRequest{
		MarkConfirmed: true,
	})
	s.OK()

	approvedCount, err := s.client.UserNotificationLog.Query().
		Where(
			usernotificationlog.UserIDEQ(seed.orgAdminID),
			usernotificationlog.KindEQ("liturgi_substitute_approved"),
			usernotificationlog.RefIDEQ(schedID),
		).Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, approvedCount)

	// RunReminderTick at h-24h -> dispatches liturgi_reminder for orgAdminID.
	s.Require().NoError(s.liturgiFeature.RunReminderTick(ctx, startAt.Add(-23*time.Hour)))
	rem24Count, err := s.client.UserNotificationLog.Query().
		Where(
			usernotificationlog.UserIDEQ(seed.orgAdminID),
			usernotificationlog.KindEQ("liturgi_reminder"),
			usernotificationlog.RefIDEQ(schedID),
			usernotificationlog.WindowEQ("h-24h"),
		).Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, rem24Count)

	// Updating Mass schedule start time triggers event_updated to active liturgi_petugas (orgAdminID).
	s.loginViaOTP("litadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", seed.parokiID.String())
	newStart := startAt.Add(30 * time.Minute)
	s.Request("/api/liturgi/schedules/"+schedID.String(), http.MethodPatch, domain.UpdateLiturgiScheduleRequest{
		StartAt: &newStart,
		EndAt:   new(newStart.Add(time.Hour)),
	})
	s.OK()

	evUpdatedCount, err := s.client.UserNotificationLog.Query().
		Where(
			usernotificationlog.UserIDEQ(seed.orgAdminID),
			usernotificationlog.KindEQ("event_updated"),
			usernotificationlog.RefIDEQ(schedID),
		).Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, evUpdatedCount)
}
