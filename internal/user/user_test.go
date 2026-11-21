package user_test

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgmember"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/usersession"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/mailx"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type stubSender struct {
	lastCode string
}

func (s *stubSender) SendOTP(_ context.Context, _, _, code string) error {
	s.lastCode = code
	return nil
}

type UserFeatureSuite struct {
	testutil.ApiTestSuite[any]
	client   *ent.Client
	enforcer *authz.Enforcer
	sender   *stubSender
	feature  *user.Feature
	cfg      user.Config
}

func TestUserFeatureSuite(t *testing.T) {
	suite.Run(t, new(UserFeatureSuite))
}

func (s *UserFeatureSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.sender = &stubSender{}
	s.cfg = user.Config{
		SuperAdminEmail: "Admin@Paroki.org",
		SuperAdminName:  "Parish Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     60 * time.Second,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.feature = user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.sender))
	s.Require().NoError(s.feature.RegisterPermissions(context.Background(), s.enforcer))
	s.WithFeatures(s.feature, &testRoutesFeature{suite: s})
}

type testRoutesFeature struct {
	suite *UserFeatureSuite
}

func (t *testRoutesFeature) Register(api fiber.Router) {
	g := api.Group("/test", t.suite.feature.AuthenticateMiddleware())
	g.Get("/whoami", t.suite.feature.RequireAuth(), func(c fiber.Ctx) error {
		p, ok := user.FromContext(c.Context())
		t.suite.True(ok)
		p2, ok2 := t.suite.feature.PrincipalResolver()(c)
		t.suite.True(ok2)
		t.suite.Equal(p.SessionID, p2.SessionID)
		return httpx.OK(c, p)
	})
	g.Get("/tenant", t.suite.feature.RequireAuth(), func(c fiber.Ctx) error {
		p, err := user.RequireTenant(c.Context())
		if err != nil {
			return err
		}
		return httpx.OK(c, p)
	})
	g.Get("/admin", t.suite.feature.Require(user.PermUsers, authz.ActionRead), func(c fiber.Ctx) error {
		return httpx.OK(c, map[string]string{"ok": "true"})
	})
}

func (s *UserFeatureSuite) issueSessionFor(email string) *http.Cookie {
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: email})
	s.OK()
	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Email: email,
		Code:  s.sender.lastCode,
	})
	s.OK()
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == s.cfg.CookieName {
			return c
		}
	}
	s.FailNow("session cookie not set")
	return nil
}

func (s *UserFeatureSuite) TestBootstrap_IdempotencyAndPermissions() {
	ctx := context.Background()
	s.Require().NoError(s.feature.Bootstrap(ctx))
	// Second run must be idempotent.
	s.Require().NoError(s.feature.Bootstrap(ctx))

	u, err := s.client.User.Query().Where(entuser.EmailEQ("admin@paroki.org")).Only(ctx)
	s.Require().NoError(err)
	s.Equal("Parish Super Admin", u.Name)

	roles, err := s.enforcer.RolesForUserInDomain(authz.UserSubject(u.ID), authz.DomainGlobal)
	s.Require().NoError(err)
	s.Contains(roles, user.RoleSuperAdmin)
	s.Contains(roles, user.RoleUser)

	// Super Admin can access /api/test/admin
	adminCookie := s.issueSessionFor("admin@paroki.org")
	s.SetCookie(adminCookie)
	s.Request("/api/test/admin", http.MethodGet, nil)
	s.OK()

	// Regular user cannot access /api/test/admin
	s.ClearCookies()
	_, err = s.feature.EnsureUser(ctx, "member@paroki.org", "Member", "", "email", "member@paroki.org")
	s.Require().NoError(err)
	memberCookie := s.issueSessionFor("member@paroki.org")
	s.SetCookie(memberCookie)
	s.Request("/api/test/admin", http.MethodGet, nil)
	s.Forbidden()
}

func (s *UserFeatureSuite) TestMiddleware_CookieBearerTenantAndDisabled() {
	ctx := context.Background()
	s.Require().NoError(s.feature.Bootstrap(ctx))
	cookie := s.issueSessionFor("admin@paroki.org")

	// Unauthenticated -> 401
	s.ClearCookies()
	s.Request("/api/test/whoami", http.MethodGet, nil)
	s.Unauthorized()

	// Cookie auth -> 200 with SessionID & Roles
	s.SetCookie(cookie)
	s.Request("/api/test/whoami", http.MethodGet, nil)
	s.OK()
	p := testutil.DecodeAs[authx.Principal](&s.ApiTestSuite).Data
	s.NotEqual(uuid.Nil, p.SessionID)
	s.Equal("web", p.Channel)
	s.Contains(p.Roles, user.RoleSuperAdmin)

	// Bearer auth -> 200
	s.ClearCookies()
	s.RequestWithToken("/api/test/whoami", http.MethodGet, nil, cookie.Value)
	s.OK()

	// RequireTenant without parish -> 400 TENANT_REQUIRED
	s.RequestWithToken("/api/test/tenant", http.MethodGet, nil, cookie.Value)
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Invalid X-Parish-Id -> 400 BAD_REQUEST
	s.SetHeader("X-Parish-Id", "not-a-uuid")
	s.RequestWithToken("/api/test/tenant", http.MethodGet, nil, cookie.Value)
	s.BadRequest()
	s.Equal("BAD_REQUEST", s.GetErrorResponse().Error.Code)

	// Valid X-Parish-Id -> 200 with ParishID
	parishID := uuid.Must(uuid.NewV7())
	s.SetHeader("X-Parish-Id", parishID.String())
	s.RequestWithToken("/api/test/tenant", http.MethodGet, nil, cookie.Value)
	s.OK()
	pTenant := testutil.DecodeAs[authx.Principal](&s.ApiTestSuite).Data
	s.Equal(parishID, pTenant.ParishID)

	// Disable user -> 403 USER_DISABLED & all sessions hard-deleted
	s.Headers.Del("X-Parish-Id")
	_, err := s.client.User.Update().
		Where(entuser.EmailEQ("admin@paroki.org")).
		SetStatus(entuser.StatusDisabled).
		Save(ctx)
	s.Require().NoError(err)

	s.RequestWithToken("/api/test/whoami", http.MethodGet, nil, cookie.Value)
	s.Forbidden()
	s.Equal("USER_DISABLED", s.GetErrorResponse().Error.Code)

	sum := sha256.Sum256([]byte(cookie.Value))
	sessCount, err := s.client.UserSession.Query().
		Where(usersession.TokenHashEQ(hex.EncodeToString(sum[:]))).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(0, sessCount)
}

func (s *UserFeatureSuite) TestChannelResolution_AndSwitchChannelParish() {
	ctx := context.Background()
	uResp, err := s.feature.EnsureUser(ctx, "botuser@paroki.org", "Bot User", "08123456789", "telegram", "tg-999")
	s.Require().NoError(err)
	s.Require().NotNil(uResp.Phone)
	s.Equal("+628123456789", *uResp.Phone)

	_, err = s.feature.EnsureUser(ctx, "botuser@paroki.org", "Bot User", "08123456789", "whatsapp", "08123456789")
	s.Require().NoError(err)

	// Invalid phone -> 422 VALIDATION_ERROR
	_, err = s.feature.EnsureUser(ctx, "badphone@paroki.org", "Bad Phone", "not-a-phone", "whatsapp", "")
	var appErr *errx.Error
	s.Require().ErrorAs(err, &appErr)
	s.Equal(errx.KindValidation, appErr.Kind)
	s.Equal("VALIDATION_ERROR", appErr.Code)

	// Unknown channel -> 404 USER_NOT_FOUND
	_, err = s.feature.ResolveByChannel(ctx, "telegram", "unknown-id", uuid.Nil)
	s.ErrorIs(err, domain.ErrUserNotFound)

	// SwitchChannelParish updates account active_parish_id and initializes user default_parish_id
	parish1 := uuid.Must(uuid.NewV7())
	s.Require().NoError(s.feature.SwitchChannelParish(ctx, "telegram", "tg-999", parish1))

	pTg, err := s.feature.ResolveByChannel(ctx, "telegram", "tg-999", uuid.Nil)
	s.Require().NoError(err)
	s.Equal(uResp.ID, pTg.UserID)
	s.Equal("telegram", pTg.Channel)
	s.Equal(parish1, pTg.ParishID)

	// Local WhatsApp subject normalizes to E.164 on SwitchChannelParish and ResolveByChannel
	parish2 := uuid.Must(uuid.NewV7())
	s.Require().NoError(s.feature.SwitchChannelParish(ctx, "whatsapp", "08123456789", parish2))
	pWa, err := s.feature.ResolveByChannel(ctx, "whatsapp", "08123456789", uuid.Nil)
	s.Require().NoError(err)
	s.Equal("whatsapp", pWa.Channel)
	s.Equal(parish2, pWa.ParishID)

	// Disabled user -> 403 USER_DISABLED on ResolveByChannel
	_, err = s.client.User.UpdateOneID(uResp.ID).SetStatus(entuser.StatusDisabled).Save(ctx)
	s.Require().NoError(err)
	_, err = s.feature.ResolveByChannel(ctx, "telegram", "tg-999", uuid.Nil)
	s.ErrorIs(err, domain.ErrUserDisabled)
}

func (s *UserFeatureSuite) TestMCP_PrincipalResolverCookieAndBearer() {
	ctx := context.Background()
	s.Require().NoError(s.feature.Bootstrap(ctx))
	admin, err := s.client.User.Query().Where(entuser.EmailEQ("admin@paroki.org")).Only(ctx)
	s.Require().NoError(err)
	cookie := s.issueSessionFor("admin@paroki.org")

	initAndPing := func(withCookie bool) string {
		s.ClearCookies()
		if withCookie {
			s.SetCookie(cookie)
		}
		s.SetHeader("Accept", "application/json, text/event-stream")
		initPayload := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]any{
				"protocolVersion": "2025-06-18",
				"capabilities":    map[string]any{},
				"clientInfo":      map[string]any{"name": "test", "version": "1.0.0"},
			},
		}
		if withCookie {
			s.Request("/mcp", http.MethodPost, initPayload)
		} else {
			s.RequestWithToken("/mcp", http.MethodPost, initPayload, cookie.Value)
		}
		s.OK()
		mcpSessionID := s.GetHttpResponse().Header.Get("Mcp-Session-Id")
		s.Require().NotEmpty(mcpSessionID)

		s.SetHeader("Mcp-Session-Id", mcpSessionID)
		notifPayload := map[string]any{
			"jsonrpc": "2.0",
			"method":  "notifications/initialized",
		}
		if withCookie {
			s.Request("/mcp", http.MethodPost, notifPayload)
		} else {
			s.RequestWithToken("/mcp", http.MethodPost, notifPayload, cookie.Value)
		}
		s.Require().Less(s.GetHttpResponse().StatusCode, 300)

		pingPayload := map[string]any{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/call",
			"params": map[string]any{
				"name":      "system_ping",
				"arguments": map[string]any{},
			},
		}
		if withCookie {
			s.Request("/mcp", http.MethodPost, pingPayload)
		} else {
			s.RequestWithToken("/mcp", http.MethodPost, pingPayload, cookie.Value)
		}
		s.OK()
		bodyBytes, err := io.ReadAll(s.GetHttpResponse().Body)
		s.Require().NoError(err)
		return string(bodyBytes)
	}

	// Cookie auth on /mcp -> PrincipalResolver populates userId
	s.Contains(initAndPing(true), admin.ID.String())

	// Bearer token auth on /mcp -> PrincipalResolver populates userId
	s.Headers.Del("Mcp-Session-Id")
	s.Contains(initAndPing(false), admin.ID.String())
}

func (s *UserFeatureSuite) TestMailOTPSender_AndLogFallback() {
	// Default LogOTPSender fallback (no WithOTPSender) succeeds.
	fallbackFeature := user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler))
	s.Require().NoError(fallbackFeature.Bootstrap(context.Background()))
	s.WithFeatures(fallbackFeature)
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "admin@paroki.org"})
	s.OK()

	// MailOTPSender delivers 6-digit OTP over SMTP and rejects non-email channels.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer func() { _ = ln.Close() }()

	received := make(chan string, 1)
	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		writeLine := func(line string) {
			_, _ = rw.WriteString(line + "\r\n")
			_ = rw.Flush()
		}
		writeLine("220 localhost ESMTP")
		var full strings.Builder
		for {
			line, readErr := rw.ReadString('\n')
			if readErr != nil {
				return
			}
			full.WriteString(line)
			trimmed := strings.TrimRight(line, "\r\n")
			upper := strings.ToUpper(trimmed)
			switch {
			case strings.HasPrefix(upper, "EHLO "), strings.HasPrefix(upper, "HELO "):
				writeLine("250 OK")
			case strings.HasPrefix(upper, "MAIL FROM:"), strings.HasPrefix(upper, "RCPT TO:"):
				writeLine("250 OK")
			case upper == "DATA":
				writeLine("354 End")
				for {
					dline, derr := rw.ReadString('\n')
					if derr != nil {
						return
					}
					full.WriteString(dline)
					if strings.TrimRight(dline, "\r\n") == "." {
						break
					}
				}
				writeLine("250 Queued")
			case upper == "QUIT":
				writeLine("221 Bye")
				received <- full.String()
				return
			default:
				writeLine("500 Unknown")
			}
		}
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	mailer, err := mailx.New(mailx.Config{
		Host: "127.0.0.1",
		Port: port,
		From: "Domus <no-reply@example.com>",
	})
	s.Require().NoError(err)

	otpSender := user.NewMailOTPSender(mailer, 5*time.Minute)
	s.Error(otpSender.SendOTP(context.Background(), "whatsapp", "+6281234567890", "123456"))

	mailFeature := user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(otpSender))
	_, err = mailFeature.EnsureUser(context.Background(), "mailtest@paroki.org", "Mail Test", "", "email", "mailtest@paroki.org")
	s.Require().NoError(err)
	s.WithFeatures(mailFeature)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "mailtest@paroki.org"})
	s.OK()

	select {
	case rawSMTP := <-received:
		s.Contains(rawSMTP, "Kode Verifikasi Login Domus")
		s.Contains(rawSMTP, "5 menit")
	case <-time.After(5 * time.Second):
		s.FailNow("timed out waiting for SMTP OTP delivery")
	}
}

func (s *UserFeatureSuite) TestAuthRolesAndMeOrgAdminResolution() {
	ctx := context.Background()
	s.enforcer.RegisterRole(authz.RoleDescriptor{
		Code:        "org:admin",
		Name:        "Admin Organisasi",
		Scope:       authz.RoleScopeOrg,
		Description: "Org admin",
	})

	// 1. GET /api/auth/roles returns registered descriptors
	s.Request("/api/auth/roles", http.MethodGet, nil)
	s.OK()
	catalog := testutil.DecodeAs[[]authz.RoleDescriptor](&s.ApiTestSuite).Data
	s.Require().Len(catalog, 3)
	s.Equal(user.RoleSuperAdmin, catalog[0].Code)
	s.Equal(user.RoleUser, catalog[1].Code)
	s.Equal("org:admin", catalog[2].Code)

	// 2. Create user + parish + active org membership as admin
	uResp, err := s.feature.EnsureUser(ctx, "orgadmin@paroki.org", "Org Admin", "", "email", "orgadmin@paroki.org")
	s.Require().NoError(err)

	diocese, err := s.client.Diocese.Create().
		SetCode("KAJ").
		SetName("Keuskupan Agung Jakarta").
		Save(ctx)
	s.Require().NoError(err)
	parish1, err := s.client.Parish.Create().
		SetDioceseID(diocese.ID).
		SetCode("P1").
		SetName("Paroki 1").
		Save(ctx)
	s.Require().NoError(err)
	parish2, err := s.client.Parish.Create().
		SetDioceseID(diocese.ID).
		SetCode("P2").
		SetName("Paroki 2").
		Save(ctx)
	s.Require().NoError(err)
	parishID := parish1.ID
	otherParishID := parish2.ID
	for _, pid := range []uuid.UUID{parishID, otherParishID} {
		_, err = s.client.ParishMember.Create().
			SetParishID(pid).
			SetUserID(uResp.ID).
			SetRole("member").
			Save(ctx)
		s.Require().NoError(err)
	}
	org, err := s.client.Organization.Create().
		SetParishID(parishID).
		SetKind("section").
		SetCode("OMK").
		SetName("Orang Muda Katolik").
		SetActive(true).
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.OrgMember.Create().
		SetParishID(parishID).
		SetOrgID(org.ID).
		SetUserID(uResp.ID).
		SetRole(orgmember.RoleAdmin).
		SetStatus(orgmember.StatusActive).
		Save(ctx)
	s.Require().NoError(err)

	cookie := s.issueSessionFor("orgadmin@paroki.org")
	s.SetCookie(cookie)

	// Without active parish -> only user:user
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	meNoParish := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal([]string{user.RoleUser}, meNoParish.Roles)
	s.Require().Len(meNoParish.RoleDetails, 1)
	s.Equal(user.RoleUser, meNoParish.RoleDetails[0].Code)

	// With other parish -> no org:admin
	s.SetHeader("X-Parish-Id", otherParishID.String())
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	meOther := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.NotContains(meOther.Roles, "org:admin")

	// With active parish where user is active org:admin -> includes org:admin in roles & roleDetails
	s.SetHeader("X-Parish-Id", parishID.String())
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	meActive := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Contains(meActive.Roles, user.RoleUser)
	s.Contains(meActive.Roles, "org:admin")
	s.Require().Len(meActive.RoleDetails, 2)
	s.Equal("org:admin", meActive.RoleDetails[1].Code)
}

func (s *UserFeatureSuite) TestRegisterSettings_HotReloadAndValidation() {
	ctx := context.Background()
	store := setting.NewStore(s.client)
	setting.RegisterFeatures(store, s.feature)
	s.Require().NoError(store.Load(ctx))

	// Initially only email OTP is enabled
	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()
	providers := testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite).Data
	s.Equal([]string{domain.ProviderEmail}, providers.OTPChannels)

	// Partial Telegram -> validation error
	_, err := store.Update(ctx, map[string]string{
		user.SettingTelegramBotToken: "123:tok",
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "Konfigurasi Telegram harus mencakup Bot Token, Bot Username, dan Webhook Secret.")

	// Partial GOWA -> validation error
	_, err = store.Update(ctx, map[string]string{
		user.SettingGowaBaseURL:   "https://gowa.example.com",
		user.SettingGowaBasicUser: "gowa",
	})
	s.Require().Error(err)

	// Valid Telegram + GOWA -> hot-reloaded in /api/auth/providers
	_, err = store.Update(ctx, map[string]string{
		user.SettingTelegramBotToken:      "123:tok",
		user.SettingTelegramBotUsername:   "domus_bot",
		user.SettingTelegramWebhookSecret: "hook-secret",
		user.SettingGowaBaseURL:           "https://gowa.example.com",
	})
	s.Require().NoError(err)

	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()
	providers = testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite).Data
	s.Contains(providers.OTPChannels, domain.ProviderTelegram)
	s.Contains(providers.OTPChannels, domain.ProviderWhatsApp)

	// Clear Telegram + GOWA -> disabled again
	_, err = store.Update(ctx, map[string]string{
		user.SettingTelegramBotToken:      "",
		user.SettingTelegramBotUsername:   "",
		user.SettingTelegramWebhookSecret: "",
		user.SettingGowaBaseURL:           "",
	})
	s.Require().NoError(err)

	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()
	providers = testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite).Data
	s.Equal([]string{domain.ProviderEmail}, providers.OTPChannels)
}
