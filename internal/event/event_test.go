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
	s.eventFeature = event.New(s.client, s.enforcer, s.userFeature, s.orgFeature, eventCfg)

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

func (s *EventFeatureSuite) seedParishAndOrg() (parishID, orgID uuid.UUID) {
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
	parishID = testutil.DecodeAs[refdomain.ParishResponse](s).Data.ID

	s.Request(fmt.Sprintf("/api/ref/parishes/%s/members", parishID), http.MethodPost, refdomain.AddParishMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Parish Admin",
		Role:  "admin",
	})
	s.Created()

	s.Request(fmt.Sprintf("/api/ref/parishes/%s/members", parishID), http.MethodPost, refdomain.AddParishMemberRequest{
		Email: "outsider@paroki.org",
		Name:  "Parish User",
		Role:  "member",
	})
	s.Created()

	// Parish admin creates org with orgadmin@paroki.org as initial admin, then adds orgmember@paroki.org.
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())

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
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request(fmt.Sprintf("/api/orgs/%s/members", orgID), http.MethodPost, orgdomain.AddOrgMemberRequest{
		Email: "orgmember@paroki.org",
		Name:  "OMK Member",
		Role:  "member",
	})
	s.Created()

	return parishID, orgID
}

func (s *EventFeatureSuite) TestEventCRUD_VisibilityAndRBAC() {
	parishID, orgID := s.seedParishAndOrg()

	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	endAt := startAt.Add(2 * time.Hour)
	deadline := startAt.Add(-2 * time.Hour)

	// Missing tenant header returns 400 TENANT_REQUIRED (when active parish is unset).
	s.loginViaOTP("super@paroki.org")
	s.SetHeader("X-Parish-Id", "")
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Non-member ref:parish-admin CANNOT create events (403 EVENT_FORBIDDEN).
	s.loginViaOTP("padmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   orgID,
		Title:   "Unauthorized Parish Admin Event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// org:member CANNOT create events (403 EVENT_FORBIDDEN).
	s.loginViaOTP("orgmember@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   orgID,
		Title:   "Unauthorized Member Event",
		StartAt: startAt,
		EndAt:   endAt,
	})
	s.Forbidden()
	s.Equal("EVENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// org:admin creates public and private events; invalid time range returns 400 EVENT_INVALID_TIME_RANGE.
	s.loginViaOTP("orgadmin@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())

	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:   orgID,
		Title:   "Invalid Range",
		StartAt: endAt,
		EndAt:   startAt,
	})
	s.BadRequest()
	s.Equal("EVENT_INVALID_TIME_RANGE", s.GetErrorResponse().Error.Code)

	s.Request("/api/events", http.MethodPost, domain.CreateEventRequest{
		OrgID:        orgID,
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
		OrgID:      orgID,
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
	s.SetHeader("X-Parish-Id", parishID.String())
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
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request("/api/events/search", http.MethodPost, domain.SearchEventsRequest{})
	s.OK()
	padminList := testutil.DecodeAs[[]domain.EventResponse](s).Data
	s.Len(padminList, 1)
	s.Equal(pubEv.ID, padminList[0].ID)

	// org:member sees BOTH public and private events with myOrgRole="member".
	s.loginViaOTP("orgmember@paroki.org")
	s.SetHeader("X-Parish-Id", parishID.String())
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
	s.SetHeader("X-Parish-Id", parishID.String())
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
