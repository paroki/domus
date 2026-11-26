package event_test

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
	"github.com/paroki/domus/internal/event/domain"
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

type EventFeatureSuite struct {
	testutil.ApiTestSuite[any]
	client       *ent.Client
	enforcer     *authz.Enforcer
	sender       *stubOTPSender
	userFeature  *user.Feature
	refFeature   *ref.Feature
	orgFeature   *org.Feature
	eventFeature *event.Feature
}

func TestEventFeatureSuite(t *testing.T) {
	suite.Run(t, new(EventFeatureSuite))
}

func (s *EventFeatureSuite) SetupTest() {
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

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature, s.orgFeature, s.eventFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature, s.orgFeature, s.eventFeature)
}

func (s *EventFeatureSuite) loginViaOTP(email string) {
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

func (s *EventFeatureSuite) seedParokiAndOrg() (parokiID, orgID uuid.UUID) {
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
	parokiID = testutil.DecodeAs[refdomain.ParokiResponse](s).Data.ID

	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Paroki Admin",
		Role:  "admin",
	})
	s.Created()

	s.Request(fmt.Sprintf("/api/ref/paroki/%s/members", parokiID), http.MethodPost, refdomain.AddParokiMemberRequest{
		Email: "outsider@paroki.org",
		Name:  "Paroki User",
		Role:  "member",
	})
	s.Created()

	// Paroki admin creates org with orgadmin@paroki.org as initial admin, then adds orgmember@paroki.org.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	s.Request("/api/orgs", http.MethodPost, orgdomain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "omk",
		Name:              "Orang Muda Katolik",
		InitialAdminEmail: "orgadmin@paroki.org",
		InitialAdminName:  "OMK Admin",
	})
	s.Created()
	orgID = testutil.DecodeAs[orgdomain.OrganizationResponse](s).Data.ID

	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgID), http.MethodPost, orgdomain.AddOrgMemberRequest{
		Email: "orgmember@paroki.org",
		Name:  "OMK Member",
		Role:  "member",
	})
	s.Created()

	return parokiID, orgID
}

func (s *EventFeatureSuite) TestEventCRUD_VisibilityAndRBAC() {
	parokiID, orgID := s.seedParokiAndOrg()

	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	endAt := startAt.Add(2 * time.Hour)
	deadline := startAt.Add(-2 * time.Hour)

	// Missing tenant header returns 400 TENANT_REQUIRED (when active parish is unset).
	s.loginViaOTP("super@paroki.org")
	s.SetHeader("X-Paroki-Id", "")
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Non-member ref:parish-admin CANNOT create events (403 EVENT_FORBIDDEN).
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Unauthorized Parish Admin Event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// org:member CANNOT create events (403 EVENT_FORBIDDEN).
	s.loginViaOTP("orgmember@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Unauthorized Member Event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// org:admin creates public and private events; invalid time range returns 400 EVENT_INVALID_TIME_RANGE.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Invalid Range",
		StartAt: endAt,
		EndAt:   startAt,
	})
	s.BadRequest()
	s.Equal("EVENT_INVALID_TIME_RANGE", s.GetErrorResponse().Error.Code)

	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:        &orgID,
		Title:        "Misa OMK",
		Location:     new("Gereja Katedral"),
		StartAt:      startAt,
		EndAt:        endAt,
		RsvpDeadline: &deadline,
	})
	s.Created()
	pubEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Equal("public", pubEv.Visibility)
	s.Equal("scheduled", pubEv.Status)
	s.True(pubEv.RsvpEnabled)
	s.True(pubEv.RsvpOpen)
	s.Equal("omk", pubEv.OrgCode)
	s.Require().NotNil(pubEv.MyOrgRole)
	s.Equal("admin", *pubEv.MyOrgRole)

	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:      &orgID,
		Title:      "Rapat Internal OMK",
		StartAt:    startAt.Add(24 * time.Hour),
		EndAt:      endAt.Add(24 * time.Hour),
		Visibility: "private",
	})
	s.Created()
	privEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Equal("private", privEv.Visibility)

	// Non-member parish user sees ONLY public event in search, and gets 403 on private GET.
	s.loginViaOTP("outsider@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.OK()
	outsiderList := testutil.DecodeAs[[]domain.EventResponse](s).Data
	s.Len(outsiderList, 1)
	s.Equal(pubEv.ID, outsiderList[0].ID)
	s.Nil(outsiderList[0].MyOrgRole)

	s.Request(fmt.Sprintf("/api/events/%s", privEv.ID), http.MethodGet, nil)
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Non-member ref:parish-admin ALSO sees only public event (no implicit private access).
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.OK()
	padminList := testutil.DecodeAs[[]domain.EventResponse](s).Data
	s.Len(padminList, 1)
	s.Equal(pubEv.ID, padminList[0].ID)

	// org:member sees BOTH public and private events with myOrgRole="member".
	s.loginViaOTP("orgmember@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.OK()
	memberList := testutil.DecodeAs[[]domain.EventResponse](s).Data
	s.Len(memberList, 2)

	s.Request(fmt.Sprintf("/api/events/%s", privEv.ID), http.MethodGet, nil)
	s.OK()
	gotPriv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Require().NotNil(gotPriv.MyOrgRole)
	s.Equal("member", *gotPriv.MyOrgRole)

	// org:admin updates & cancels public event, then soft-deletes private event.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	cancelled := "cancelled"
	clearDL := true
	newTitle := "Misa OMK (Cancelled)"
	s.Request(fmt.Sprintf("/api/events/%s", pubEv.ID), http.MethodPatch, domain.UpdateEventRequest{
		Title:             &newTitle,
		Status:            &cancelled,
		ClearRsvpDeadline: &clearDL,
	})
	s.OK()
	updatedPub := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Equal(newTitle, updatedPub.Title)
	s.Equal("cancelled", updatedPub.Status)
	s.Nil(updatedPub.RsvpDeadline)
	s.False(updatedPub.RsvpOpen)

	s.Request(fmt.Sprintf("/api/events/%s", privEv.ID), http.MethodDelete, nil)
	s.NoContent()

	s.Request(fmt.Sprintf("/api/events/%s", privEv.ID), http.MethodGet, nil)
	s.NotFound()
	s.Equal("EVENT_NOT_FOUND", s.GetErrorResponse().Error.Code)
}

func (s *EventFeatureSuite) TestEventRSVP_UpsertDeleteCutoffSummaryAndRoster() {
	parokiID, orgID := s.seedParokiAndOrg()

	// Add a 3rd active member so totalActiveMembers == 3.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgID), http.MethodPost, orgdomain.AddOrgMemberRequest{
		Email:    "member2@paroki.org",
		Name:     "OMK Member 2",
		Role:     "member",
		Position: new("Sekretaris"),
	})
	s.Created()

	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	endAt := startAt.Add(2 * time.Hour)

	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Latihan Koor",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Created()
	ev := testutil.DecodeAs[domain.EventResponse](s).Data

	// Non-member (outsider & ref:parish-admin) cannot RSVP, delete RSVP, or view summary/roster -> 403 EVENT_FORBIDDEN.
	for _, nonMember := range []string{"outsider@paroki.org", "padmin@paroki.org"} {
		s.loginViaOTP(nonMember)
		s.SetHeader("X-Paroki-Id", parokiID.String())
		s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
		s.Forbidden()
		s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

		s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodDelete, nil)
		s.Forbidden()

		s.Request(fmt.Sprintf("/api/events/%s/rsvps/summary", ev.ID), http.MethodGet, nil)
		s.Forbidden()

		s.Request(fmt.Sprintf("/api/events/%s/rsvps/search", ev.ID), http.MethodPost, domain.SearchEventRSVPsRequest{})
		s.Forbidden()
	}

	// org:member upserts RSVP with note, then updates status and clears note.
	s.loginViaOTP("orgmember@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodPost, domain.UpsertEventRSVPRequest{
		Status: "going",
		Note:   new("  Hadir tepat waktu  "),
	})
	s.OK()
	rsvp1 := testutil.DecodeAs[domain.EventRSVPResponse](s).Data
	s.Equal("going", rsvp1.Status)
	s.Require().NotNil(rsvp1.Note)
	s.Equal("Hadir tepat waktu", *rsvp1.Note)
	s.Equal("OMK Member", rsvp1.Name)
	s.Equal("member", rsvp1.OrgRole)

	// Idempotent update clears note when empty string.
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodPost, domain.UpsertEventRSVPRequest{
		Status: "maybe",
		Note:   new(""),
	})
	s.OK()
	rsvp1Updated := testutil.DecodeAs[domain.EventRSVPResponse](s).Data
	s.Equal(rsvp1.ID, rsvp1Updated.ID)
	s.Equal("maybe", rsvp1Updated.Status)
	s.Nil(rsvp1Updated.Note)

	// member2 RSVPs "going", orgadmin RSVPs "not_going".
	s.loginViaOTP("member2@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodPost, domain.UpsertEventRSVPRequest{
		Status: "going",
		Note:   new("Bawa gitar"),
	})
	s.OK()
	rsvp2 := testutil.DecodeAs[domain.EventRSVPResponse](s).Data
	s.Require().NotNil(rsvp2.OrgPosition)
	s.Equal("Sekretaris", *rsvp2.OrgPosition)

	// Verify GET /api/events/{id} and POST /api/events/search populate myRsvp & rsvpSummary for org member.
	s.Request(fmt.Sprintf("/api/events/%s", ev.ID), http.MethodGet, nil)
	s.OK()
	gotEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Require().NotNil(gotEv.MyRsvp)
	s.Equal("going", gotEv.MyRsvp.Status)
	s.Require().NotNil(gotEv.RsvpSummary)
	s.Equal(1, gotEv.RsvpSummary.Going)
	s.Equal(1, gotEv.RsvpSummary.Maybe)
	s.Equal(0, gotEv.RsvpSummary.NotGoing)
	s.Equal(2, gotEv.RsvpSummary.TotalResponded)
	s.Equal(3, gotEv.RsvpSummary.TotalActiveMembers)
	s.Equal(1, gotEv.RsvpSummary.Unresponded)

	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.OK()
	searchedEvs := testutil.DecodeAs[[]domain.EventResponse](s).Data
	s.Require().Len(searchedEvs, 1)
	s.Require().NotNil(searchedEvs[0].MyRsvp)
	s.Require().NotNil(searchedEvs[0].RsvpSummary)

	// Verify GET /rsvps/summary and POST /rsvps/search filtering + cursor pagination.
	s.Request(fmt.Sprintf("/api/events/%s/rsvps/summary", ev.ID), http.MethodGet, nil)
	s.OK()
	summary := testutil.DecodeAs[domain.EventRSVPSummaryResponse](s).Data
	s.Equal(2, summary.TotalResponded)
	s.Equal(3, summary.TotalActiveMembers)
	s.Equal(1, summary.Unresponded)

	s.Request(fmt.Sprintf("/api/events/%s/rsvps/search", ev.ID), http.MethodPost, domain.SearchEventRSVPsRequest{
		Limit: 1,
	})
	s.OK()
	page1 := testutil.DecodeAs[[]domain.EventRSVPResponse](s)
	s.Len(page1.Data, 1)
	s.True(page1.Meta.HasMore)
	s.NotEmpty(page1.Meta.NextCursor)

	s.Request(fmt.Sprintf("/api/events/%s/rsvps/search", ev.ID), http.MethodPost, domain.SearchEventRSVPsRequest{
		Cursor: page1.Meta.NextCursor,
		Limit:  1,
	})
	s.OK()
	page2 := testutil.DecodeAs[[]domain.EventRSVPResponse](s)
	s.Len(page2.Data, 1)
	s.False(page2.Meta.HasMore)
	s.NotEqual(page1.Data[0].ID, page2.Data[0].ID)

	// Filter by status="going" and q="gitar".
	s.Request(fmt.Sprintf("/api/events/%s/rsvps/search", ev.ID), http.MethodPost, domain.SearchEventRSVPsRequest{
		Q:      "gitar",
		Filter: &domain.EventRSVPSearchFilter{Status: "going"},
	})
	s.OK()
	filtered := testutil.DecodeAs[[]domain.EventRSVPResponse](s).Data
	s.Len(filtered, 1)
	s.Equal(rsvp2.ID, filtered[0].ID)

	// Hard delete RSVP + idempotent 204.
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodDelete, nil)
	s.NoContent()
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodDelete, nil)
	s.NoContent()

	// Cutoff tests: within grace period (startAt 10m ago, grace 30m) -> allowed!
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	graceStart := time.Now().UTC().Add(-10 * time.Minute)
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Within Grace Period",
		StartAt: graceStart,
		EndAt:   graceStart.Add(2 * time.Hour),
	})
	s.Created()
	graceEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.True(graceEv.RsvpOpen)
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", graceEv.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
	s.OK()

	// Past grace period (startAt 45m ago) -> 409 EVENT_RSVP_CLOSED.
	pastGraceStart := time.Now().UTC().Add(-45 * time.Minute)
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Past Grace Period",
		StartAt: pastGraceStart,
		EndAt:   pastGraceStart.Add(2 * time.Hour),
	})
	s.Created()
	pastGraceEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.False(pastGraceEv.RsvpOpen)
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", pastGraceEv.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
	s.Conflict()
	s.Equal("EVENT_RSVP_CLOSED", s.GetErrorResponse().Error.Code)

	// Past rsvpDeadline -> 409 EVENT_RSVP_CLOSED.
	pastDL := time.Now().UTC().Add(-5 * time.Minute)
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:        &orgID,
		Title:        "Past Deadline",
		StartAt:      startAt,
		EndAt:        endAt,
		RsvpDeadline: &pastDL,
	})
	s.Created()
	pastDLEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", pastDLEv.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
	s.Conflict()
	s.Equal("EVENT_RSVP_CLOSED", s.GetErrorResponse().Error.Code)

	// rsvpEnabled=false -> 409 EVENT_RSVP_CLOSED.
	disabled := false
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:       &orgID,
		Title:       "RSVP Disabled",
		StartAt:     startAt,
		EndAt:       endAt,
		RsvpEnabled: &disabled,
	})
	s.Created()
	disabledEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", disabledEv.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
	s.Conflict()
	s.Equal("EVENT_RSVP_CLOSED", s.GetErrorResponse().Error.Code)

	// Cancelled event -> 409 EVENT_RSVP_CLOSED on POST & DELETE, but summary/search still work!
	cancelled := "cancelled"
	s.Request(fmt.Sprintf("/api/events/%s", ev.ID), http.MethodPatch, domain.UpdateEventRequest{Status: &cancelled})
	s.OK()
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
	s.Conflict()
	s.Equal("EVENT_RSVP_CLOSED", s.GetErrorResponse().Error.Code)
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", ev.ID), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("EVENT_RSVP_CLOSED", s.GetErrorResponse().Error.Code)

	s.Request(fmt.Sprintf("/api/events/%s/rsvps/summary", ev.ID), http.MethodGet, nil)
	s.OK()
	s.Request(fmt.Sprintf("/api/events/%s/rsvps/search", ev.ID), http.MethodPost, domain.SearchEventRSVPsRequest{})
	s.OK()
}

func (s *EventFeatureSuite) TestParishLevelEventsAndOpenRSVP() {
	parokiID, orgID := s.seedParokiAndOrg()
	startAt := time.Now().UTC().Add(24 * time.Hour)
	endAt := startAt.Add(2 * time.Hour)

	// 1. org:admin CANNOT create parish-level event (orgId nil or uuid.Nil) -> 403 EVENT_FORBIDDEN.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		Title:   "Unauthorized Parish Event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	nilOrg := uuid.Nil
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &nilOrg,
		Title:   "Unauthorized Nil UUID Parish Event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// 2. ref:paroki-admin CANNOT create private parish-level event -> 422 VALIDATION_ERROR.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		Title:      "Private Parish Event",
		StartAt:    startAt,
		EndAt:      endAt,
		Visibility: "private",
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// 3. ref:paroki-admin creates public parish-level event.
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		Title:   "Misa Syukur Paroki",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Created()
	parishEv := testutil.DecodeAs[domain.EventResponse](s).Data
	s.Nil(parishEv.OrgID)
	s.Equal("public", parishEv.Visibility)
	s.True(parishEv.CanManage)
	s.True(parishEv.CanRsvp)

	// 4. Cannot update parish-level event to private -> 422 VALIDATION_ERROR.
	privVis := "private"
	s.Request(fmt.Sprintf("/api/events/%s", parishEv.ID), http.MethodPatch, domain.UpdateEventRequest{
		Visibility: &privVis,
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// 5. org:admin CANNOT update or delete parish-level event -> 403 EVENT_FORBIDDEN.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	newTitle := "Hijack"
	s.Request(fmt.Sprintf("/api/events/%s", parishEv.ID), http.MethodPatch, domain.UpdateEventRequest{Title: &newTitle})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)
	s.Request(fmt.Sprintf("/api/events/%s", parishEv.ID), http.MethodDelete, nil)
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Create an org event as well to test parokiWideOnly filter.
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   &orgID,
		Title:   "Rapat OMK",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Created()

	// 6. Regular parish member (outsider@paroki.org) can see, RSVP, view summary & roster on parish-level event.
	s.loginViaOTP("outsider@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())

	parokiWide := true
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{
		Filter: &domain.EventSearchFilter{ParokiWideOnly: &parokiWide},
	})
	s.OK()
	wideList := testutil.DecodeAs[[]domain.EventResponse](s).Data
	s.Require().Len(wideList, 1)
	s.Equal(parishEv.ID, wideList[0].ID)
	s.False(wideList[0].CanManage)
	s.True(wideList[0].CanRsvp)

	s.Request(fmt.Sprintf("/api/events/%s/rsvp", parishEv.ID), http.MethodPost, domain.UpsertEventRSVPRequest{
		Status: "going",
		Note:   new("Hadir bersama keluarga"),
	})
	s.OK()
	rsvp := testutil.DecodeAs[domain.EventRSVPResponse](s).Data
	s.Equal("going", rsvp.Status)
	s.Empty(rsvp.OrgRole)
	s.Nil(rsvp.OrgPosition)

	s.Request(fmt.Sprintf("/api/events/%s/rsvps/summary", parishEv.ID), http.MethodGet, nil)
	s.OK()
	summary := testutil.DecodeAs[domain.EventRSVPSummaryResponse](s).Data
	s.Equal(1, summary.Going)
	// seedParokiAndOrg adds 4 ParokiMembers: padmin, outsider, orgadmin, orgmember.
	s.Equal(4, summary.TotalActiveMembers)
	s.Equal(3, summary.Unresponded)

	s.Request(fmt.Sprintf("/api/events/%s/rsvps/search", parishEv.ID), http.MethodPost, domain.SearchEventRSVPsRequest{})
	s.OK()
	roster := testutil.DecodeAs[[]domain.EventRSVPResponse](s).Data
	s.Require().Len(roster, 1)
	s.Equal("outsider@paroki.org", roster[0].Email)

	// 7. Non-member of parish (super@paroki.org has no ParokiMember row) gets 403 on RSVP endpoints.
	s.loginViaOTP("super@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.Request(fmt.Sprintf("/api/events/%s/rsvp", parishEv.ID), http.MethodPost, domain.UpsertEventRSVPRequest{Status: "going"})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// 8. ref:paroki-admin can update and delete the parish-level event.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Paroki-Id", parokiID.String())
	updatedTitle := "Misa Syukur Agung Paroki"
	s.Request(fmt.Sprintf("/api/events/%s", parishEv.ID), http.MethodPatch, domain.UpdateEventRequest{Title: &updatedTitle})
	s.OK()
	s.Equal(updatedTitle, testutil.DecodeAs[domain.EventResponse](s).Data.Title)

	s.Request(fmt.Sprintf("/api/events/%s", parishEv.ID), http.MethodDelete, nil)
	s.NoContent()
}
