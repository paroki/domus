package user_test

import (
	"bufio"
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
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
	"github.com/paroki/domus/internal/platform/pushx"
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

	// RequireTenant without paroki -> 400 TENANT_REQUIRED
	s.RequestWithToken("/api/test/tenant", http.MethodGet, nil, cookie.Value)
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Invalid X-Paroki-Id -> 400 BAD_REQUEST
	s.SetHeader("X-Paroki-Id", "not-a-uuid")
	s.RequestWithToken("/api/test/tenant", http.MethodGet, nil, cookie.Value)
	s.BadRequest()
	s.Equal("BAD_REQUEST", s.GetErrorResponse().Error.Code)

	// Valid X-Paroki-Id -> 200 with ParokiID
	parokiID := uuid.Must(uuid.NewV7())
	s.SetHeader("X-Paroki-Id", parokiID.String())
	s.RequestWithToken("/api/test/tenant", http.MethodGet, nil, cookie.Value)
	s.OK()
	pTenant := testutil.DecodeAs[authx.Principal](&s.ApiTestSuite).Data
	s.Equal(parokiID, pTenant.ParokiID)

	// Disable user -> 403 USER_DISABLED & all sessions hard-deleted
	s.Headers.Del("X-Paroki-Id")
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

func (s *UserFeatureSuite) TestChannelResolution_AndSwitchChannelParoki() {
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

	// SwitchChannelParoki updates account active_paroki_id and initializes user default_paroki_id
	paroki1 := uuid.Must(uuid.NewV7())
	s.Require().NoError(s.feature.SwitchChannelParoki(ctx, "telegram", "tg-999", paroki1))

	pTg, err := s.feature.ResolveByChannel(ctx, "telegram", "tg-999", uuid.Nil)
	s.Require().NoError(err)
	s.Equal(uResp.ID, pTg.UserID)
	s.Equal("telegram", pTg.Channel)
	s.Equal(paroki1, pTg.ParokiID)

	// Local WhatsApp subject normalizes to E.164 on SwitchChannelParoki and ResolveByChannel
	paroki2 := uuid.Must(uuid.NewV7())
	s.Require().NoError(s.feature.SwitchChannelParoki(ctx, "whatsapp", "08123456789", paroki2))
	pWa, err := s.feature.ResolveByChannel(ctx, "whatsapp", "08123456789", uuid.Nil)
	s.Require().NoError(err)
	s.Equal("whatsapp", pWa.Channel)
	s.Equal(paroki2, pWa.ParokiID)

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

	// 2. Create user + paroki + active org membership as admin
	uResp, err := s.feature.EnsureUser(ctx, "orgadmin@paroki.org", "Org Admin", "", "email", "orgadmin@paroki.org")
	s.Require().NoError(err)

	keuskupan, err := s.client.Keuskupan.Create().
		SetKode("KAJ").
		SetNama("Keuskupan Agung Jakarta").
		Save(ctx)
	s.Require().NoError(err)
	paroki1, err := s.client.Paroki.Create().
		SetKeuskupanID(keuskupan.ID).
		SetKode("P1").
		SetNama("Paroki 1").
		Save(ctx)
	s.Require().NoError(err)
	paroki2, err := s.client.Paroki.Create().
		SetKeuskupanID(keuskupan.ID).
		SetKode("P2").
		SetNama("Paroki 2").
		Save(ctx)
	s.Require().NoError(err)
	parokiID := paroki1.ID
	otherParokiID := paroki2.ID
	for _, pid := range []uuid.UUID{parokiID, otherParokiID} {
		_, err = s.client.ParokiMember.Create().
			SetParokiID(pid).
			SetUserID(uResp.ID).
			SetRole("member").
			Save(ctx)
		s.Require().NoError(err)
	}
	org, err := s.client.Organization.Create().
		SetParokiID(parokiID).
		SetKind("section").
		SetCode("OMK").
		SetName("Orang Muda Katolik").
		SetActive(true).
		Save(ctx)
	s.Require().NoError(err)

	_, err = s.client.OrgMember.Create().
		SetParokiID(parokiID).
		SetOrgID(org.ID).
		SetUserID(uResp.ID).
		SetRole(orgmember.RoleAdmin).
		SetStatus(orgmember.StatusActive).
		Save(ctx)
	s.Require().NoError(err)

	cookie := s.issueSessionFor("orgadmin@paroki.org")
	s.SetCookie(cookie)

	// Without active paroki -> only user:user
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	meNoParoki := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal([]string{user.RoleUser}, meNoParoki.Roles)
	s.Require().Len(meNoParoki.RoleDetails, 1)
	s.Equal(user.RoleUser, meNoParoki.RoleDetails[0].Code)

	// With other paroki -> no org:admin
	s.SetHeader("X-Paroki-Id", otherParokiID.String())
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	meOther := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.NotContains(meOther.Roles, "org:admin")

	// With active paroki where user is active org:admin -> includes org:admin in roles & roleDetails
	s.SetHeader("X-Paroki-Id", parokiID.String())
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

	// Partial GOWA -> validation error
	_, err := store.Update(ctx, map[string]string{
		user.SettingGowaWebhookSecret: "secret",
	})
	s.Require().Error(err)

	// Valid Telegram (Bot Token only) + GOWA -> hot-reloaded in /api/auth/providers
	_, err = store.Update(ctx, map[string]string{
		user.SettingTelegramBotToken: "123:tok",
		user.SettingGowaBaseURL:      "https://gowa.example.com",
	})
	s.Require().NoError(err)

	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()
	providers = testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite).Data
	s.Contains(providers.OTPChannels, domain.ProviderTelegram)
	s.Contains(providers.OTPChannels, domain.ProviderWhatsApp)

	// Clear Telegram + GOWA -> disabled again
	_, err = store.Update(ctx, map[string]string{
		user.SettingTelegramBotToken: "",
		user.SettingGowaBaseURL:      "",
	})
	s.Require().NoError(err)

	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()
	providers = testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite).Data
	s.Equal([]string{domain.ProviderEmail}, providers.OTPChannels)
}

func (s *UserFeatureSuite) TestGetUser_ReturnsActiveUserBiodata() {
	ctx := context.Background()
	u, err := s.feature.EnsureUser(ctx, "biodata@paroki.org", "Budi Santoso", "", "email", "biodata@paroki.org")
	s.Require().NoError(err)

	gender := "pria"
	birthDate := "1990-05-15"
	_, err = s.client.User.UpdateOneID(u.ID).SetGender(entuser.Gender(gender)).SetBirthDate(birthDate).Save(ctx)
	s.Require().NoError(err)

	got, err := s.feature.GetUser(ctx, u.ID)
	s.Require().NoError(err)
	s.Equal("Budi Santoso", got.Name)
	s.Require().NotNil(got.Gender)
	s.Equal("pria", *got.Gender)
	s.Require().NotNil(got.BirthDate)
	s.Equal("1990-05-15", *got.BirthDate)

	_, err = s.feature.GetUser(ctx, uuid.New())
	s.Require().Error(err)
}

func (s *UserFeatureSuite) TestUserTimezone_DefaultUpdateValidationAndReactivate() {
	ctx := context.Background()
	u, err := s.feature.EnsureUser(ctx, "tzuser@paroki.org", "TZ User", "", "email", "tzuser@paroki.org")
	s.Require().NoError(err)
	s.Equal("Asia/Makassar", u.Timezone)

	cookie := s.issueSessionFor("tzuser@paroki.org")
	s.SetCookie(cookie)

	// Default timezone on GET /api/me
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal("Asia/Makassar", me.Timezone)

	// Valid updates via PATCH /api/me
	for _, tz := range []string{"Asia/Jakarta", "Asia/Jayapura"} {
		s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Timezone: &tz})
		s.OK()
		updated := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
		s.Equal(tz, updated.Timezone)
	}

	// Invalid timezones rejected with 422 VALIDATION_ERROR
	for _, invalidTZ := range []string{"Local", "localtime", "", "   ", "Invalid/Zone"} {
		s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Timezone: &invalidTZ})
		s.UnprocessableEntity()
		errResp := s.GetErrorResponse()
		s.Require().NotNil(errResp.Error)
		s.Equal("VALIDATION_ERROR", errResp.Error.Code)
	}

	// Soft-delete and reactivate resets timezone to Asia/Makassar
	s.Require().NoError(s.client.User.UpdateOneID(u.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))
	reactivated, err := s.feature.EnsureUser(ctx, "tzuser@paroki.org", "TZ User Reactivated", "", "email", "tzuser@paroki.org")
	s.Require().NoError(err)
	s.Equal("Asia/Makassar", reactivated.Timezone)
}

type stubNotificationSender struct {
	calls []struct {
		target  string
		payload domain.NotificationPayload
	}
}

func (n *stubNotificationSender) SendNotification(_ context.Context, target string, payload domain.NotificationPayload) error {
	n.calls = append(n.calls, struct {
		target  string
		payload domain.NotificationPayload
	}{target: target, payload: payload})
	return nil
}

func generateTestWebPushKeys(t *testing.T) (p256dh, auth string) {
	t.Helper()
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	authBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, authBytes); err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes()), base64.RawURLEncoding.EncodeToString(authBytes)
}

func (s *UserFeatureSuite) TestVAPIDBootstrap_SubscribeUnsubscribeAndSessionCleanup() {
	ctx := context.Background()
	store := setting.NewStore(s.client)
	setting.RegisterFeatures(store, s.feature)
	s.Require().NoError(store.Load(ctx))
	s.Require().NoError(s.feature.Bootstrap(ctx))

	pub1 := store.Get(user.SettingPushVAPIDPublicKey)
	priv1 := store.Get(user.SettingPushVAPIDPrivateKey)
	s.NotEmpty(pub1)
	s.NotEmpty(priv1)

	// Idempotent on second Bootstrap
	s.Require().NoError(s.feature.Bootstrap(ctx))
	s.Equal(pub1, store.Get(user.SettingPushVAPIDPublicKey))

	// GET /api/auth/push/vapid-public-key
	s.Request("/api/auth/push/vapid-public-key", http.MethodGet, nil)
	s.OK()
	vapidResp := testutil.DecodeAs[domain.VAPIDPublicKeyResponse](&s.ApiTestSuite).Data
	s.Equal(pub1, vapidResp.PublicKey)

	u, err := s.feature.EnsureUser(ctx, "pushuser@paroki.org", "Push User", "", "email", "pushuser@paroki.org")
	s.Require().NoError(err)
	s.True(u.NotifyPush)
	s.True(u.NotifyWhatsapp)
	s.True(u.NotifyTelegram)

	cookie := s.issueSessionFor("pushuser@paroki.org")
	s.SetCookie(cookie)

	p256dh, auth := generateTestWebPushKeys(s.T())

	// SSRF and key validation rejections (422)
	invalidReqs := []domain.SubscribePushRequest{
		{Endpoint: "http://example.com/push", Keys: domain.PushSubscriptionKeys{P256DH: p256dh, Auth: auth}},
		{Endpoint: "https://10.0.0.1/push", Keys: domain.PushSubscriptionKeys{P256DH: p256dh, Auth: auth}},
		{Endpoint: "https://fcm.googleapis.com/fcm/send/abc", Keys: domain.PushSubscriptionKeys{P256DH: "bad", Auth: auth}},
		{Endpoint: "https://fcm.googleapis.com/fcm/send/abc", Keys: domain.PushSubscriptionKeys{P256DH: p256dh, Auth: "bad"}},
	}
	for _, req := range invalidReqs {
		s.Request("/api/me/push/subscribe", http.MethodPost, req)
		s.UnprocessableEntity()
		errResp := s.GetErrorResponse()
		s.Require().NotNil(errResp.Error)
		s.Equal("VALIDATION_ERROR", errResp.Error.Code)
	}

	// Idempotent upsert on POST /api/me/push/subscribe
	subReq := domain.SubscribePushRequest{
		Endpoint: "https://fcm.googleapis.com/fcm/send/sub-1",
		Keys:     domain.PushSubscriptionKeys{P256DH: p256dh, Auth: auth},
	}
	s.Request("/api/me/push/subscribe", http.MethodPost, subReq)
	s.NoContent()
	s.Request("/api/me/push/subscribe", http.MethodPost, subReq)
	s.NoContent()

	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal(1, me.PushSubscriptionCount)

	// PATCH /api/me updates notification preferences
	f := false
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{
		NotifyPush:     &f,
		NotifyWhatsapp: &f,
		NotifyTelegram: &f,
	})
	s.OK()
	me = testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.False(me.NotifyPush)
	s.False(me.NotifyWhatsapp)
	s.False(me.NotifyTelegram)

	// POST /api/me/push/unsubscribe
	s.Request("/api/me/push/unsubscribe", http.MethodPost, domain.UnsubscribePushRequest{Endpoint: subReq.Endpoint})
	s.NoContent()
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me = testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal(0, me.PushSubscriptionCount)

	// Re-subscribe then logout -> session-bound subscription cleaned up
	s.Request("/api/me/push/subscribe", http.MethodPost, subReq)
	s.NoContent()
	s.Request("/api/auth/logout", http.MethodPost, nil)
	s.NoContent()
	count, err := s.client.UserPushSubscription.Query().Count(ctx)
	s.Require().NoError(err)
	s.Equal(0, count)

	// Soft-delete + ReactivateUser resets notification prefs to true
	s.Require().NoError(s.client.User.UpdateOneID(u.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))
	reactivated, err := s.feature.EnsureUser(ctx, "pushuser@paroki.org", "Push User", "", "email", "pushuser@paroki.org")
	s.Require().NoError(err)
	s.True(reactivated.NotifyPush)
	s.True(reactivated.NotifyWhatsapp)
	s.True(reactivated.NotifyTelegram)
}

func (s *UserFeatureSuite) TestPushTestEndpoint_ChannelsAndAutoPrune410() {
	ctx := context.Background()
	waStub := &stubNotificationSender{}
	tgStub := &stubNotificationSender{}

	var pushHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushHits++
		if strings.HasSuffix(r.URL.Path, "/expired") {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	cfg := user.Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
		SessionTTL:     24 * time.Hour,
		CookieName:     "domus_session",
	}
	otpStub := &stubSender{}
	f := user.New(
		s.client,
		s.enforcer,
		cfg,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		user.WithOTPSender(otpStub),
		user.WithChannelNotification("whatsapp", waStub),
		user.WithChannelNotification("telegram", tgStub),
	)
	store := setting.NewStore(s.client)
	setting.RegisterFeatures(store, f)
	s.Require().NoError(store.Load(ctx))
	s.Require().NoError(f.Bootstrap(ctx))
	s.Require().NotEmpty(pushx.ErrSubscriptionExpired)

	suite := &testutil.ApiTestSuite[any]{}
	suite.SetT(s.T())
	suite.WithFeatures(f)

	phone := "+6281234567890"
	u, err := f.EnsureUser(ctx, "notifyall@paroki.org", "Notify All", phone, "email", "notifyall@paroki.org")
	s.Require().NoError(err)
	_, err = s.client.UserAccount.Create().
		SetUserID(u.ID).
		SetProvider(domain.ProviderTelegram).
		SetSubject("987654321").
		SetEmail("notifyall@paroki.org").
		Save(ctx)
	s.Require().NoError(err)

	suite.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "notifyall@paroki.org"})
	suite.OK()
	suite.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Email: "notifyall@paroki.org",
		Code:  otpStub.lastCode,
	})
	suite.OK()
	var sessCookie *http.Cookie
	for _, c := range suite.GetHttpResponse().Cookies() {
		if c.Name == "domus_session" {
			sessCookie = c
			break
		}
	}
	s.Require().NotNil(sessCookie)
	suite.SetCookie(sessCookie)

	p256dh, auth := generateTestWebPushKeys(s.T())
	suite.Request("/api/me/push/subscribe", http.MethodPost, domain.SubscribePushRequest{
		Endpoint: srv.URL + "/active",
		Keys:     domain.PushSubscriptionKeys{P256DH: p256dh, Auth: auth},
	})
	suite.NoContent()
	suite.Request("/api/me/push/subscribe", http.MethodPost, domain.SubscribePushRequest{
		Endpoint: srv.URL + "/expired",
		Keys:     domain.PushSubscriptionKeys{P256DH: p256dh, Auth: auth},
	})
	suite.NoContent()

	suite.Request("/api/me/push/test", http.MethodPost, nil)
	suite.NoContent()

	s.Equal(2, pushHits)
	s.Len(waStub.calls, 1)
	s.Equal(phone, waStub.calls[0].target)
	s.Len(tgStub.calls, 1)
	s.Equal("987654321", tgStub.calls[0].target)

	// Expired 410 Gone subscription was auto-pruned; only /active remains
	suite.Request("/api/me", http.MethodGet, nil)
	suite.OK()
	me := testutil.DecodeAs[domain.UserResponse](suite).Data
	s.Equal(1, me.PushSubscriptionCount)
}
