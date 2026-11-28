package event_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/app"
	"github.com/paroki/domus/internal/event"
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

type EventMCPSuite struct {
	suite.Suite
	client       *ent.Client
	enforcer     *authz.Enforcer
	sender       *stubOTPSender
	userFeature  *user.Feature
	refFeature   *ref.Feature
	orgFeature   *org.Feature
	eventFeature *event.Feature
	fiberApp     *fiber.App
}

func TestEventMCPSuite(t *testing.T) {
	suite.Run(t, new(EventMCPSuite))
}

func (s *EventMCPSuite) SetupTest() {
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

	s.fiberApp = app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler), s.userFeature, s.refFeature, s.orgFeature, s.eventFeature)
}

func (s *EventMCPSuite) loginCookies(email string) []*http.Cookie {
	s.T().Helper()
	reqOTP, _ := json.Marshal(userdomain.RequestOTPRequest{Email: email})
	r1 := httptest.NewRequest(http.MethodPost, "/api/auth/otp", bytes.NewReader(reqOTP))
	r1.Header.Set("Content-Type", "application/json")
	resp1, err := s.fiberApp.Test(r1)
	s.Require().NoError(err)
	defer func() { _ = resp1.Body.Close() }()
	s.Require().Equal(http.StatusOK, resp1.StatusCode)

	verOTP, _ := json.Marshal(userdomain.VerifyOTPRequest{Email: email, Code: s.sender.lastCode})
	r2 := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewReader(verOTP))
	r2.Header.Set("Content-Type", "application/json")
	resp2, err := s.fiberApp.Test(r2)
	s.Require().NoError(err)
	defer func() { _ = resp2.Body.Close() }()
	s.Require().Equal(http.StatusOK, resp2.StatusCode)
	return resp2.Cookies()
}

func (s *EventMCPSuite) apiJSON(method, path string, body any, cookies []*http.Cookie, parokiID string, out any) int {
	s.T().Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		s.Require().NoError(err)
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	if parokiID != "" {
		req.Header.Set("X-Paroki-Id", parokiID)
	}
	resp, err := s.fiberApp.Test(req)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	if out != nil {
		raw, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(raw, out)
	}
	return resp.StatusCode
}

func (s *EventMCPSuite) mcpPost(sessionID string, cookies []*http.Cookie, parokiID string, payload any) (*http.Response, string) {
	s.T().Helper()
	raw, err := json.Marshal(payload)
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	if parokiID != "" {
		req.Header.Set("X-Paroki-Id", parokiID)
	}
	resp, err := s.fiberApp.Test(req)
	s.Require().NoError(err)
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, string(bodyBytes)
}

func (s *EventMCPSuite) initMCPSession(cookies []*http.Cookie, parokiID string) string {
	resp, _ := s.mcpPost("", cookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "event-mcp-test", "version": "1.0.0"},
		},
	})
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	sessionID := resp.Header.Get("Mcp-Session-Id")
	s.Require().NotEmpty(sessionID)

	nResp, _ := s.mcpPost(sessionID, cookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.Require().Equal(http.StatusAccepted, nResp.StatusCode)
	return sessionID
}

func (s *EventMCPSuite) TestEventMCP_DiscoveryToolsAndRBAC() {
	superCookies := s.loginCookies("super@paroki.org")

	var dioEnv struct{ Data refdomain.KeuskupanResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/keuskupan", refdomain.CreateKeuskupanRequest{
		Kode: "kaj",
		Nama: "Keuskupan Agung Jakarta",
	}, superCookies, "", &dioEnv))

	var parEnv struct{ Data refdomain.ParokiResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki", refdomain.CreateParokiRequest{
		KeuskupanID: dioEnv.Data.ID,
		Kode:        "p-kaj-01",
		Nama:        "Paroki Katedral",
	}, superCookies, "", &parEnv))
	parokiID := parEnv.Data.ID.String()

	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Paroki Admin",
		Role:  "admin",
	}, superCookies, "", nil))

	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "outsider@paroki.org",
		Name:  "Paroki User",
		Role:  "member",
	}, superCookies, "", nil))

	pAdminCookies := s.loginCookies("padmin@paroki.org")
	var orgEnv struct {
		Data orgdomain.OrganizationResponse
	}
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/orgs", orgdomain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "omk",
		Name:              "Orang Muda Katolik",
		InitialAdminEmail: "orgadmin@paroki.org",
		InitialAdminName:  "OMK Admin",
	}, pAdminCookies, parokiID, &orgEnv))
	orgID := orgEnv.Data.ID.String()

	orgAdminCookies := s.loginCookies("orgadmin@paroki.org")
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/orgs/"+orgID+"/members", orgdomain.AddOrgMemberRequest{
		Email: "orgmember@paroki.org",
		Name:  "OMK Member",
		Role:  "member",
	}, orgAdminCookies, parokiID, nil))

	// 1. tools/list discovers all 6 event_* tools.
	adminSess := s.initMCPSession(orgAdminCookies, parokiID)
	listResp, listBody := s.mcpPost(adminSess, orgAdminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	s.Equal(http.StatusOK, listResp.StatusCode)
	for _, toolName := range []string{
		"event_list",
		"event_get",
		"event_create",
		"event_update",
		"event_rsvp",
		"event_list_rsvps",
	} {
		s.Contains(listBody, `"`+toolName+`"`)
	}

	// 2. Missing tenant (user with no paroki membership) returns TENANT_REQUIRED error on tool call.
	_, err := s.userFeature.EnsureUser(context.Background(), "notenant@paroki.org", "No Tenant", "", "email", "notenant@paroki.org")
	s.Require().NoError(err)
	noTenantCookies := s.loginCookies("notenant@paroki.org")
	noTenantSess := s.initMCPSession(noTenantCookies, "")
	_, noTenantBody := s.mcpPost(noTenantSess, noTenantCookies, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_list",
			"arguments": map[string]any{},
		},
	})
	s.Contains(noTenantBody, `"isError":true`)
	s.Contains(noTenantBody, "An active paroki tenant is required.")

	// 3. event_create (public + private) as org:admin.
	startAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	endAt := startAt.Add(2 * time.Hour)
	_, createPubBody := s.mcpPost(adminSess, orgAdminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "event_create",
			"arguments": map[string]any{
				"orgId":   orgID,
				"title":   "Misa OMK",
				"startAt": startAt.Format(time.RFC3339),
				"endAt":   endAt.Format(time.RFC3339),
			},
		},
	})
	s.Contains(createPubBody, "Misa OMK")

	_, createPrivBody := s.mcpPost(adminSess, orgAdminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "event_create",
			"arguments": map[string]any{
				"orgId":      orgID,
				"title":      "Rapat Internal OMK",
				"startAt":    startAt.Add(24 * time.Hour).Format(time.RFC3339),
				"endAt":      endAt.Add(24 * time.Hour).Format(time.RFC3339),
				"visibility": "private",
			},
		},
	})
	s.Contains(createPrivBody, "Rapat Internal OMK")

	// Extract public event ID via event_list (including date-only YYYY-MM-DD range).
	_, adminListBody := s.mcpPost(adminSess, orgAdminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      6,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "event_list",
			"arguments": map[string]any{
				"query": "Misa OMK",
				"from":  startAt.Format(time.DateOnly),
				"to":    endAt.Format(time.DateOnly),
			},
		},
	})
	s.Contains(adminListBody, "Misa OMK")

	// Fetch event ID from REST list for subsequent exact tool calls.
	var evSearchEnv struct{ Data []event.Event }
	s.Equal(http.StatusOK, s.apiJSON(http.MethodPost, "/api/events/search", map[string]any{"q": "Misa OMK"}, orgAdminCookies, parokiID, &evSearchEnv))
	s.Require().Len(evSearchEnv.Data, 1)
	pubEventID := evSearchEnv.Data[0].ID.String()

	// 4. Outsider sees only public event in event_list and gets EVENT_FORBIDDEN on event_rsvp.
	outsiderCookies := s.loginCookies("outsider@paroki.org")
	outsiderSess := s.initMCPSession(outsiderCookies, parokiID)
	_, outsiderListBody := s.mcpPost(outsiderSess, outsiderCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      7,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_list",
			"arguments": map[string]any{},
		},
	})
	s.Contains(outsiderListBody, "Misa OMK")
	s.NotContains(outsiderListBody, "Rapat Internal OMK")

	_, outsiderRSVPBody := s.mcpPost(outsiderSess, outsiderCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      8,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_rsvp",
			"arguments": map[string]any{"eventId": pubEventID, "status": "going"},
		},
	})
	s.Contains(outsiderRSVPBody, `"isError":true`)
	s.Contains(outsiderRSVPBody, "You do not have permission to perform this event action.")

	// 5. org:member submits RSVP via event_rsvp, views via event_get & event_list_rsvps.
	memberCookies := s.loginCookies("orgmember@paroki.org")
	memberSess := s.initMCPSession(memberCookies, parokiID)
	_, rsvpBody := s.mcpPost(memberSess, memberCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      9,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_rsvp",
			"arguments": map[string]any{"eventId": pubEventID, "status": "going", "note": "Hadir"},
		},
	})
	s.Contains(rsvpBody, "Hadir")

	_, getBody := s.mcpPost(memberSess, memberCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_get",
			"arguments": map[string]any{"eventId": pubEventID},
		},
	})
	s.Contains(getBody, "myRsvp")
	s.Contains(getBody, "rsvpSummary")

	_, listRSVPsBody := s.mcpPost(memberSess, memberCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_list_rsvps",
			"arguments": map[string]any{"eventId": pubEventID},
		},
	})
	s.Contains(listRSVPsBody, "OMK Member")
	s.Contains(listRSVPsBody, `"going":1`)

	// 6. org:admin cancels event via event_update; subsequent event_rsvp returns EVENT_RSVP_CLOSED.
	_, cancelBody := s.mcpPost(adminSess, orgAdminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      12,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_update",
			"arguments": map[string]any{"eventId": pubEventID, "status": "cancelled"},
		},
	})
	s.Contains(cancelBody, `"status":"cancelled"`)

	_, closedBody := s.mcpPost(memberSess, memberCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      13,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_rsvp",
			"arguments": map[string]any{"eventId": pubEventID, "status": "maybe"},
		},
	})
	s.Contains(closedBody, `"isError":true`)
	s.Contains(closedBody, "RSVP is closed for this event.")
}

func (s *EventMCPSuite) TestEventMCPTools_ParishEventAndTimezone() {
	s.loginCookies("super@paroki.org")
	superCookies := s.loginCookies("super@paroki.org")

	var keuEnv struct{ Data refdomain.KeuskupanResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/keuskupan", refdomain.CreateKeuskupanRequest{
		Kode: "kam",
		Nama: "Keuskupan Agung Makassar",
	}, superCookies, "", &keuEnv))

	var parEnv struct{ Data refdomain.ParokiResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki", refdomain.CreateParokiRequest{
		KeuskupanID: keuEnv.Data.ID,
		Kode:        "p-kam-01",
		Nama:        "Paroki Hati Kudus",
		Timezone:    new("Asia/Makassar"),
	}, superCookies, "", &parEnv))
	parokiID := parEnv.Data.ID.String()

	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "padmin-tz@paroki.org", Name: "Paroki Admin", Role: "admin",
	}, superCookies, "", nil))
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "umat-tz@paroki.org", Name: "Umat Paroki", Role: "member",
	}, superCookies, "", nil))

	// Paroki admin creates a parish-level event with local wall-clock YYYY-MM-DDTHH:MM:SS (Asia/Makassar = +08:00).
	padminCookies := s.loginCookies("padmin-tz@paroki.org")
	padminSess := s.initMCPSession(padminCookies, parokiID)

	_, createBody := s.mcpPost(padminSess, padminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "event_create",
			"arguments": map[string]any{
				"title":   "Misa Pagi Paroki",
				"startAt": "2026-12-01T08:00:00",
				"endAt":   "2026-12-01T09:30:00",
			},
		},
	})
	s.Contains(createBody, "Misa Pagi Paroki")
	s.Contains(createBody, `"timezone":"Asia/Makassar"`)
	s.Contains(createBody, "2026-12-01T08:00:00+08:00")
	s.Contains(createBody, "2026-12-01T09:30:00+08:00")

	// Verify HTTP API stores UTC (08:00 WITA == 00:00 UTC).
	var evSearchEnv struct{ Data []event.Event }
	s.Equal(http.StatusOK, s.apiJSON(http.MethodPost, "/api/events/search", map[string]any{
		"filter": map[string]any{"parokiWideOnly": true},
	}, padminCookies, parokiID, &evSearchEnv))
	s.Require().Len(evSearchEnv.Data, 1)
	s.Nil(evSearchEnv.Data[0].OrgID)
	s.Equal("2026-12-01T00:00:00Z", evSearchEnv.Data[0].StartAt.UTC().Format(time.RFC3339))

	// Parish member RSVPs via MCP.
	umatCookies := s.loginCookies("umat-tz@paroki.org")
	umatSess := s.initMCPSession(umatCookies, parokiID)
	_, rsvpBody := s.mcpPost(umatSess, umatCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_rsvp",
			"arguments": map[string]any{"eventId": evSearchEnv.Data[0].ID.String(), "status": "going"},
		},
	})
	s.Contains(rsvpBody, `"status":"going"`)

	// MCP event_create, event_list, event_update with jenis.
	_, createJenisBody := s.mcpPost(padminSess, padminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "event_create",
			"arguments": map[string]any{
				"title":   "Ibadat Sabda Paroki",
				"jenis":   "ibadat",
				"startAt": "2026-12-02T18:00:00",
				"endAt":   "2026-12-02T19:00:00",
			},
		},
	})
	s.Contains(createJenisBody, `"jenis":"ibadat"`)

	_, listJenisBody := s.mcpPost(padminSess, padminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "event_list",
			"arguments": map[string]any{"jenis": "ibadat"},
		},
	})
	s.Contains(listJenisBody, "Ibadat Sabda Paroki")
	s.NotContains(listJenisBody, "Misa Pagi Paroki")

	var ibadatSearch struct{ Data []event.Event }
	s.Equal(http.StatusOK, s.apiJSON(http.MethodPost, "/api/events/search", map[string]any{
		"filter": map[string]any{"jenis": "ibadat"},
	}, padminCookies, parokiID, &ibadatSearch))
	s.Require().Len(ibadatSearch.Data, 1)

	_, clearJenisBody := s.mcpPost(padminSess, padminCookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "event_update",
			"arguments": map[string]any{
				"eventId": ibadatSearch.Data[0].ID.String(),
				"jenis":   "",
			},
		},
	})
	s.NotContains(clearJenisBody, `"jenis":"ibadat"`)
}
