package liturgi_test

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
	"github.com/paroki/domus/internal/cantus"
	cantusdomain "github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/event"
	"github.com/paroki/domus/internal/liturgi"
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

type LiturgiMCPSuite struct {
	suite.Suite
	client         *ent.Client
	enforcer       *authz.Enforcer
	sender         *stubOTPSender
	userFeature    *user.Feature
	refFeature     *ref.Feature
	orgFeature     *org.Feature
	cantusFeature  *cantus.Feature
	eventFeature   *event.Feature
	liturgiFeature *liturgi.Feature
	fiberApp       *fiber.App
}

func TestLiturgiMCPSuite(t *testing.T) {
	suite.Run(t, new(LiturgiMCPSuite))
}

func (s *LiturgiMCPSuite) SetupTest() {
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

	s.fiberApp = app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler), s.userFeature, s.refFeature, s.orgFeature, s.cantusFeature, s.eventFeature, s.liturgiFeature)
}

func (s *LiturgiMCPSuite) loginCookies(email string) []*http.Cookie {
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

func (s *LiturgiMCPSuite) apiJSON(method, path string, body any, cookies []*http.Cookie, parokiID string, out any) int {
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

func (s *LiturgiMCPSuite) mcpPost(sessionID string, cookies []*http.Cookie, parokiID string, payload any) (*http.Response, string) {
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

func (s *LiturgiMCPSuite) initMCPSession(cookies []*http.Cookie, parokiID string) string {
	resp, _ := s.mcpPost("", cookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "liturgi-mcp-test", "version": "1.0.0"},
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

func (s *LiturgiMCPSuite) callTool(sessionID string, cookies []*http.Cookie, parokiID string, id int, name string, args map[string]any, out any) {
	s.T().Helper()
	resp, raw := s.mcpPost(sessionID, cookies, parokiID, map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	})
	s.Require().Equal(http.StatusOK, resp.StatusCode, raw)
	var rpc struct {
		Result struct {
			IsError           bool            `json:"isError"`
			StructuredContent json.RawMessage `json:"structuredContent"`
			Content           []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"result"`
	}
	s.Require().NoError(json.Unmarshal([]byte(raw), &rpc), raw)
	s.Require().False(rpc.Result.IsError, raw)
	if out != nil {
		if len(rpc.Result.StructuredContent) > 0 {
			s.Require().NoError(json.Unmarshal(rpc.Result.StructuredContent, out), string(rpc.Result.StructuredContent))
		} else if len(rpc.Result.Content) > 0 {
			s.Require().NoError(json.Unmarshal([]byte(rpc.Result.Content[0].Text), out), rpc.Result.Content[0].Text)
		}
	}
}

func (s *LiturgiMCPSuite) TestLiturgiMCP_All9ToolsAndTimezoneNormalization() {
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
		Timezone:    new("Asia/Jakarta"),
	}, superCookies, "", &parEnv))
	parokiID := parEnv.Data.ID.String()

	var adminMemberEnv struct {
		Data refdomain.ParokiMemberResponse
	}
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Paroki Admin",
		Role:  "admin",
	}, superCookies, "", &adminMemberEnv))

	var lektorEnv struct {
		Data refdomain.ParokiMemberResponse
	}
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "lektor@paroki.org",
		Name:  "Lektor Satu",
		Role:  "member",
	}, superCookies, "", &lektorEnv))

	var subEnv struct {
		Data refdomain.ParokiMemberResponse
	}
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/paroki/"+parokiID+"/members", refdomain.AddParokiMemberRequest{
		Email: "sub@paroki.org",
		Name:  "Lektor Pengganti",
		Role:  "member",
	}, superCookies, "", &subEnv))

	ctx := context.Background()
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(adminMemberEnv.Data.UserID), cantus.RoleAdmin, authz.DomainGlobal))
	adminCookies := s.loginCookies("padmin@paroki.org")

	var orgEnv struct {
		Data orgdomain.OrganizationResponse
	}
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/orgs", orgdomain.CreateOrganizationRequest{
		Code:              "kor-katedral",
		Name:              "Paduan Suara Katedral",
		Kind:              "section",
		InitialAdminEmail: "koradmin@paroki.org",
		InitialAdminName:  "Kor Admin",
	}, adminCookies, parokiID, &orgEnv))

	// Create a Cantus entry for liturgi_set_lagu cantusId auto-fill.

	var cantusEnv struct{ Data cantusdomain.CantusResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/cantus", cantusdomain.CreateCantusRequest{
		Slug:      "ps-565-datanglah",
		Jenis:     "lagu",
		Judul:     "Datanglah Ya Roh Pencipta",
		Sumber:    new("Puji Syukur"),
		Nomor:     new("565"),
		NadaDasar: new("D"),
	}, adminCookies, "", &cantusEnv))

	adminSession := s.initMCPSession(adminCookies, parokiID)

	// 1. liturgi_set_ministry
	var minOut liturgi.MCPLiturgiMinistry
	s.callTool(adminSession, adminCookies, parokiID, 10, "liturgi_set_ministry", map[string]any{
		"roleSlug":     "lektor",
		"orgId":        orgEnv.Data.ID.String(),
		"leaderUserId": lektorEnv.Data.UserID.String(),
	}, &minOut)
	s.Equal("lektor", minOut.RoleSlug)
	s.Require().NotNil(minOut.LeaderUserID)
	s.Equal(lektorEnv.Data.UserID.String(), *minOut.LeaderUserID)

	// 2. liturgi_create_schedule (wall-clock in Asia/Jakarta -> +07:00)
	var sched liturgi.MCPLiturgiSchedule
	s.callTool(adminSession, adminCookies, parokiID, 11, "liturgi_create_schedule", map[string]any{
		"title":   "Misa Minggu Pagi",
		"startAt": "2030-07-07T08:00:00",
		"endAt":   "2030-07-07T09:30:00",
	}, &sched)
	s.Equal("Asia/Jakarta", sched.Timezone)
	s.Equal("2030-07-07T08:00:00+07:00", sched.StartAt)

	// 3. liturgi_update_schedule
	var updatedSched liturgi.MCPLiturgiSchedule
	s.callTool(adminSession, adminCookies, parokiID, 12, "liturgi_update_schedule", map[string]any{
		"eventId": sched.ID,
		"title":   "Misa Minggu Pagi II",
	}, &updatedSched)
	s.Equal("Misa Minggu Pagi II", updatedSched.Title)

	// 4. liturgi_list_schedules
	var listOut liturgi.LiturgiListSchedulesOutput
	s.callTool(adminSession, adminCookies, parokiID, 13, "liturgi_list_schedules", map[string]any{
		"query": "Pagi II",
	}, &listOut)
	s.Require().Len(listOut.Schedules, 1)
	s.Equal(sched.ID, listOut.Schedules[0].ID)

	// 5. liturgi_set_petugas (set create)
	var petOut liturgi.MCPLiturgiPetugas
	s.callTool(adminSession, adminCookies, parokiID, 14, "liturgi_set_petugas", map[string]any{
		"action":  "set",
		"eventId": sched.ID,
		"role":    "lektor",
		"userId":  lektorEnv.Data.UserID.String(),
	}, &petOut)
	s.Equal("Lektor Satu", petOut.Name)

	// 6. liturgi_respond_petugas (lektor declines with substitute)
	lektorCookies := s.loginCookies("lektor@paroki.org")
	lektorSession := s.initMCPSession(lektorCookies, parokiID)
	var respOut liturgi.MCPLiturgiPetugas
	s.callTool(lektorSession, lektorCookies, parokiID, 15, "liturgi_respond_petugas", map[string]any{
		"eventId":          sched.ID,
		"petugasId":        petOut.ID,
		"status":           "declined",
		"substituteUserId": subEnv.Data.UserID.String(),
	}, &respOut)
	s.Equal("declined", respOut.Status)
	s.Require().NotNil(respOut.SubstituteName)
	s.Equal("Lektor Pengganti", *respOut.SubstituteName)

	// 7. liturgi_approve_substitute
	var appOut liturgi.MCPLiturgiPetugas
	s.callTool(adminSession, adminCookies, parokiID, 16, "liturgi_approve_substitute", map[string]any{
		"eventId":       sched.ID,
		"petugasId":     petOut.ID,
		"markConfirmed": true,
	}, &appOut)
	s.Equal("confirmed", appOut.Status)
	s.Equal("Lektor Pengganti", appOut.Name)

	// 8. liturgi_set_lagu (set with cantusId auto-fill)
	var laguOut liturgi.MCPLiturgiLagu
	s.callTool(adminSession, adminCookies, parokiID, 17, "liturgi_set_lagu", map[string]any{
		"action":   "set",
		"eventId":  sched.ID,
		"bagian":   "pembuka",
		"cantusId": cantusEnv.Data.ID.String(),
	}, &laguOut)
	s.Equal("Datanglah Ya Roh Pencipta", laguOut.Judul)

	// 9. liturgi_get_schedule
	var getOut liturgi.MCPLiturgiSchedule
	s.callTool(adminSession, adminCookies, parokiID, 18, "liturgi_get_schedule", map[string]any{
		"eventId": sched.ID,
	}, &getOut)
	s.Require().Len(getOut.Petugas, 1)
	s.Require().Len(getOut.Lagu, 1)

	// Remove lagu & petugas via action="remove"
	var delOut liturgi.LiturgiDeleteOutput
	s.callTool(adminSession, adminCookies, parokiID, 19, "liturgi_set_lagu", map[string]any{
		"action":  "remove",
		"eventId": sched.ID,
		"laguId":  laguOut.ID,
	}, &delOut)
	s.True(delOut.Deleted)

	s.callTool(adminSession, adminCookies, parokiID, 20, "liturgi_set_petugas", map[string]any{
		"action":    "remove",
		"eventId":   sched.ID,
		"petugasId": petOut.ID,
	}, &delOut)
	s.True(delOut.Deleted)
}
