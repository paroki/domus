package org_test

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
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/app"
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

type OrgMCPSuite struct {
	suite.Suite
	client      *ent.Client
	enforcer    *authz.Enforcer
	sender      *stubOTPSender
	userFeature *user.Feature
	refFeature  *ref.Feature
	orgFeature  *org.Feature
	fiberApp    *fiber.App
}

func TestOrgMCPSuite(t *testing.T) {
	suite.Run(t, new(OrgMCPSuite))
}

func (s *OrgMCPSuite) SetupTest() {
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

	s.fiberApp = app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler), s.userFeature, s.refFeature, s.orgFeature)
}

func (s *OrgMCPSuite) loginCookies(email string) []*http.Cookie {
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

func (s *OrgMCPSuite) apiJSON(method, path string, body any, cookies []*http.Cookie, parishID string, out any) int {
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
	if parishID != "" {
		req.Header.Set("X-Parish-Id", parishID)
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

func (s *OrgMCPSuite) mcpPost(sessionID string, cookies []*http.Cookie, parishID string, payload any) (*http.Response, string) {
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
	if parishID != "" {
		req.Header.Set("X-Parish-Id", parishID)
	}
	resp, err := s.fiberApp.Test(req)
	s.Require().NoError(err)
	bodyBytes, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, string(bodyBytes)
}

func (s *OrgMCPSuite) initMCPSession(cookies []*http.Cookie, parishID string) string {
	resp, _ := s.mcpPost("", cookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "org-mcp-test", "version": "1.0.0"},
		},
	})
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	sessionID := resp.Header.Get("Mcp-Session-Id")
	s.Require().NotEmpty(sessionID)

	nResp, _ := s.mcpPost(sessionID, cookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.Require().Equal(http.StatusAccepted, nResp.StatusCode)
	return sessionID
}

func (s *OrgMCPSuite) TestOrgMCP_DiscoveryToolsAndRBAC() {
	superCookies := s.loginCookies("super@paroki.org")

	var dioEnv struct{ Data refdomain.DioceseResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/dioceses", refdomain.CreateDioceseRequest{
		Code: "kas",
		Name: "Keuskupan Agung Semarang",
	}, superCookies, "", &dioEnv))

	var parEnv struct{ Data refdomain.ParishResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/parishes", refdomain.CreateParishRequest{
		DioceseID: dioEnv.Data.ID,
		Code:      "p-kas-01",
		Name:      "Paroki Kotabaru",
	}, superCookies, "", &parEnv))
	parishID := parEnv.Data.ID.String()

	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/ref/parishes/"+parishID+"/members", refdomain.AddParishMemberRequest{
		Email: "padmin@paroki.org",
		Name:  "Parish Admin",
		Role:  "admin",
	}, superCookies, "", nil))

	pAdminCookies := s.loginCookies("padmin@paroki.org")
	var orgEnv struct{ Data domain.OrganizationResponse }
	s.Equal(http.StatusCreated, s.apiJSON(http.MethodPost, "/api/orgs", domain.CreateOrganizationRequest{
		Kind:              "section",
		Code:              "sek-katekese",
		Name:              "Seksi Katekese",
		InitialAdminEmail: "orgadmin@paroki.org",
		InitialAdminName:  "Org Admin",
	}, pAdminCookies, parishID, &orgEnv))
	orgID := orgEnv.Data.ID.String()

	orgAdminCookies := s.loginCookies("orgadmin@paroki.org")
	sessionID := s.initMCPSession(orgAdminCookies, parishID)

	// 1. tools/list discovers all 7 org_* tools.
	listResp, listBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	s.Equal(http.StatusOK, listResp.StatusCode)
	for _, toolName := range []string{
		"org_list",
		"org_get",
		"org_list_members",
		"org_update_member",
		"org_create_invite",
		"org_revoke_invite",
		"org_redeem_invite",
	} {
		s.Contains(listBody, `"`+toolName+`"`)
	}

	// 2. org_list & org_get.
	_, callBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_list",
			"arguments": map[string]any{"myOrgsOnly": true},
		},
	})
	s.Contains(callBody, "sek-katekese")

	_, getBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_get",
			"arguments": map[string]any{"orgId": orgID},
		},
	})
	s.Contains(getBody, "Seksi Katekese")

	// 3. org_update_member (onboard new member) & org_list_members.
	_, addMemBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "org_update_member",
			"arguments": map[string]any{
				"orgId":    orgID,
				"email":    "katekis@paroki.org",
				"name":     "Katekis 1",
				"role":     "member",
				"position": "Pengajar",
			},
		},
	})
	s.Contains(addMemBody, "katekis@paroki.org")

	_, listMemBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      6,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_list_members",
			"arguments": map[string]any{"orgId": orgID},
		},
	})
	s.Contains(listMemBody, "katekis@paroki.org")

	// 4. org_create_invite.
	_, invBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      7,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_create_invite",
			"arguments": map[string]any{"orgId": orgID, "role": "member"},
		},
	})
	var rpcInv struct {
		Result struct {
			StructuredContent org.MCPOrgInvite `json:"structuredContent"`
		} `json:"result"`
	}
	s.Require().NoError(json.Unmarshal([]byte(invBody), &rpcInv))
	inv := rpcInv.Result.StructuredContent
	s.Len(inv.Code, 12)

	// 5. Guest user with NO active parish redeems invite via org_redeem_invite.
	_, err := s.userFeature.EnsureUser(context.Background(), "guestmcp@paroki.org", "Guest MCP", "", "email", "guestmcp@paroki.org")
	s.Require().NoError(err)
	guestCookies := s.loginCookies("guestmcp@paroki.org")
	guestSessionID := s.initMCPSession(guestCookies, "")

	_, redeemBody := s.mcpPost(guestSessionID, guestCookies, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      8,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_redeem_invite",
			"arguments": map[string]any{"code": inv.Code},
		},
	})
	s.Contains(redeemBody, orgID)

	// 6. RBAC rejection: non-admin member cannot revoke invite.
	guestParishSessionID := s.initMCPSession(guestCookies, parishID)
	_, rbacBody := s.mcpPost(guestParishSessionID, guestCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      9,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_revoke_invite",
			"arguments": map[string]any{"orgId": orgID, "inviteId": inv.ID},
		},
	})
	s.Contains(rbacBody, `"isError":true`)
	s.Contains(rbacBody, domain.ErrForbidden.Error())

	// 7. org:admin revokes invite via org_revoke_invite.
	_, revBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      10,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_revoke_invite",
			"arguments": map[string]any{"orgId": orgID, "inviteId": inv.ID},
		},
	})
	s.Contains(revBody, `"revoked":true`)

	// 8. org_create: non-parish-admin rejected, ref:parish-admin succeeds.
	_, createDeniedBody := s.mcpPost(sessionID, orgAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_create",
			"arguments": map[string]any{"kind": "ward", "code": "ling-st-yusuf", "name": "Lingkungan St. Yusuf"},
		},
	})
	s.Contains(createDeniedBody, `"isError":true`)
	s.Contains(createDeniedBody, domain.ErrForbidden.Error())

	pAdminSessionID := s.initMCPSession(pAdminCookies, parishID)
	_, createOKBody := s.mcpPost(pAdminSessionID, pAdminCookies, parishID, map[string]any{
		"jsonrpc": "2.0",
		"id":      12,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "org_create",
			"arguments": map[string]any{"kind": "ward", "code": "ling-st-yusuf", "name": "Lingkungan St. Yusuf"},
		},
	})
	s.NotContains(createOKBody, `"isError":true`)
	s.Contains(createOKBody, `"code":"ling-st-yusuf"`)

	_ = uuid.Nil
}
