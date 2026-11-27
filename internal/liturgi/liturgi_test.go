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
	s.eventFeature = event.New(s.client, s.enforcer, s.userFeature, s.orgFeature, eventCfg, event.WithRef(s.refFeature))
	s.liturgiFeature = liturgi.New(s.client, s.enforcer, s.userFeature, s.orgFeature, s.eventFeature)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature, s.orgFeature, s.eventFeature, s.liturgiFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature, s.orgFeature, s.eventFeature, s.liturgiFeature)
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
