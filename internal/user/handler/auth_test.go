package handler_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/useraccount"
	"github.com/paroki/domus/ent/userotp"
	"github.com/paroki/domus/ent/usersession"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type stubOTPSender struct {
	lastChannel string
	lastTarget  string
	lastCode    string
}

func (s *stubOTPSender) SendOTP(_ context.Context, channel, target, code string) error {
	s.lastChannel = channel
	s.lastTarget = target
	s.lastCode = code
	return nil
}

type stubOAuthProvider struct {
	name string
	info domain.OAuthUserInfo
	err  error
}

func (p *stubOAuthProvider) Name() string { return p.name }

func (p *stubOAuthProvider) AuthCodeURL(state, redirectURI string) string {
	v := url.Values{}
	v.Set("state", state)
	v.Set("redirect_uri", redirectURI)
	return "https://oauth.example.test/authorize?" + v.Encode()
}

func (p *stubOAuthProvider) Exchange(_ context.Context, _, _ string) (domain.OAuthUserInfo, error) {
	return p.info, p.err
}

type AuthHandlerSuite struct {
	testutil.ApiTestSuite[any]
	client    *ent.Client
	enforcer  *authz.Enforcer
	otpSender *stubOTPSender
	feature   *user.Feature
	cfg       user.Config
}

func TestAuthHandlerSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerSuite))
}

func (s *AuthHandlerSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.otpSender = &stubOTPSender{}
	s.cfg = user.Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 3,
		OTPCooldown:    60 * time.Second,
		SessionTTL:     24 * time.Hour,
		CookieName:     "domus_session",
		GoogleID:       "google-id",
		GoogleSecret:   "google-secret",
		GitHubID:       "github-id",
		GitHubSecret:   "github-secret",
	}
	s.feature = user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.otpSender))
	s.WithFeatures(s.feature)
}

func (s *AuthHandlerSuite) createTestUser(email, name string, status entuser.Status) *ent.User {
	u, err := s.client.User.Create().
		SetEmail(email).
		SetName(name).
		SetStatus(status).
		Save(context.Background())
	s.Require().NoError(err)
	return u
}

func (s *AuthHandlerSuite) TestGetProviders_DeterministicOrder() {
	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()

	res := testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite)
	s.Equal([]string{"google", "github"}, res.Data.Providers)
	s.Equal([]string{"email"}, res.Data.OTPChannels)
}

func (s *AuthHandlerSuite) TestMultiChannelOTP_DiscoveryValidationAndDisabled() {
	// Disabled whatsapp/telegram channels return 400 AUTH_CHANNEL_DISABLED.
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "whatsapp",
		Target:  "+6281234567890",
	})
	s.BadRequest()
	s.Equal("AUTH_CHANNEL_DISABLED", s.GetErrorResponse().Error.Code)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "telegram",
		Target:  "alice@paroki.org",
	})
	s.BadRequest()
	s.Equal("AUTH_CHANNEL_DISABLED", s.GetErrorResponse().Error.Code)

	// Invalid channel -> 422 VALIDATION_ERROR on "channel".
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "sms",
		Target:  "+6281234567890",
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Require().NotEmpty(s.GetErrorResponse().Error.Details)
	s.Equal("channel", s.GetErrorResponse().Error.Details[0].Field)

	// Missing both target and email -> 422 VALIDATION_ERROR REQUIRED.
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{})
	s.UnprocessableEntity()
	s.Equal("REQUIRED", s.GetErrorResponse().Error.Details[0].Code)

	// Differing target and email -> 422 VALIDATION_ERROR on "target".
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Target: "a@paroki.org",
		Email:  "b@paroki.org",
	})
	s.UnprocessableEntity()
	s.Equal("target", s.GetErrorResponse().Error.Details[0].Field)

	// Invalid email syntax -> 422 VALIDATION_ERROR EMAIL.
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Email: "not-an-email",
	})
	s.UnprocessableEntity()
	s.Equal("EMAIL", s.GetErrorResponse().Error.Details[0].Code)

	// Reconfigure feature with WhatsApp + Telegram channels enabled.
	waSender := &stubOTPSender{}
	tgSender := &stubOTPSender{}
	multiCfg := s.cfg
	multiCfg.OTPCooldown = 0
	multiFeature := user.New(
		s.client,
		s.enforcer,
		multiCfg,
		slog.New(slog.DiscardHandler),
		user.WithOTPSender(s.otpSender),
		user.WithChannelOTPSender("telegram", tgSender),
		user.WithChannelOTPSender("whatsapp", waSender),
	)
	s.WithFeatures(multiFeature)

	s.Request("/api/auth/providers", http.MethodGet, nil)
	s.OK()
	res := testutil.DecodeAs[domain.AuthProvidersResponse](&s.ApiTestSuite)
	s.Equal([]string{"email", "whatsapp", "telegram"}, res.Data.OTPChannels)

	// inviteCode rejected on non-email channel -> 422 on "channel".
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel:    "whatsapp",
		Target:     "+6281234567890",
		InviteCode: "INV-123",
	})
	s.UnprocessableEntity()
	s.Equal("channel", s.GetErrorResponse().Error.Details[0].Field)
}

func (s *AuthHandlerSuite) TestMultiChannelOTP_WhatsAppAndTelegramLogin() {
	ctx := context.Background()
	var gotGowaBody map[string]string
	gowaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Require().NoError(json.NewDecoder(r.Body).Decode(&gotGowaBody))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"SUCCESS"}`))
	}))
	defer gowaSrv.Close()

	var gotTgPayload map[string]any
	tgSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Require().NoError(json.NewDecoder(r.Body).Decode(&gotTgPayload))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer tgSrv.Close()

	gowaClient, err := gowa.New(gowa.Config{BaseURL: gowaSrv.URL})
	s.Require().NoError(err)
	tgClient, err := telegram.New(telegram.Config{
		BotToken:      "123:SECRET",
		BotUsername:   "domus_test_bot",
		WebhookSecret: "hook-secret",
		BaseURL:       tgSrv.URL,
	})
	s.Require().NoError(err)

	cfg := s.cfg
	cfg.OTPCooldown = 0
	multiFeature := user.New(
		s.client,
		s.enforcer,
		cfg,
		slog.New(slog.DiscardHandler),
		user.WithOTPSender(s.otpSender),
		user.WithGowaClient(gowaClient),
		user.WithTelegramClient(tgClient),
	)
	s.WithFeatures(multiFeature)

	// User with ONLY unverified User.phone (no whatsapp UserAccount) -> rejected with 404.
	unverifiedPhone := "+628111111111"
	unverifiedUser, err := s.client.User.Create().
		SetEmail("unverified@paroki.org").
		SetName("Unverified Phone").
		SetPhone(unverifiedPhone).
		SetStatus(entuser.StatusActive).
		Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "whatsapp",
		Target:  "08111111111",
	})
	s.NotFound()
	s.Equal("USER_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Linking Telegram without verified WhatsApp still fails WhatsApp lookup -> 404.
	_, err = s.client.UserAccount.Create().
		SetUserID(unverifiedUser.ID).
		SetProvider(domain.ProviderTelegram).
		SetSubject("5550001").
		Save(ctx)
	s.Require().NoError(err)
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "telegram",
		Target:  unverifiedPhone,
	})
	s.NotFound()
	s.Equal("USER_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Invalid phone format -> 422 VALIDATION_ERROR on "target".
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "whatsapp",
		Target:  "abc",
	})
	s.UnprocessableEntity()
	s.Equal("target", s.GetErrorResponse().Error.Details[0].Field)

	// Verified WhatsApp & Telegram accounts on active user.
	verifiedUser := s.createTestUser("multi-otp@paroki.org", "Multi OTP", entuser.StatusActive)
	waAcc, err := s.client.UserAccount.Create().
		SetUserID(verifiedUser.ID).
		SetProvider(domain.ProviderWhatsApp).
		SetSubject("+6281234567890").
		Save(ctx)
	s.Require().NoError(err)
	tgAcc, err := s.client.UserAccount.Create().
		SetUserID(verifiedUser.ID).
		SetProvider(domain.ProviderTelegram).
		SetSubject("987654321").
		Save(ctx)
	s.Require().NoError(err)

	// 1) WhatsApp OTP login via local 08... target.
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "whatsapp",
		Target:  "0812-3456-7890",
	})
	s.OK()
	s.Equal("6281234567890@s.whatsapp.net", gotGowaBody["phone"])
	s.Contains(gotGowaBody["message"], "Kode OTP login Domus")

	// Extract 6-digit code from Indonesian message.
	var waCode string
	for _, field := range bytes.Fields([]byte(gotGowaBody["message"])) {
		if len(field) == 6 {
			waCode = string(field)
			break
		}
	}
	s.Require().Len(waCode, 6)

	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Channel: "whatsapp",
		Target:  "+6281234567890",
		Code:    waCode,
	})
	s.OK()
	waSess := testutil.DecodeAs[domain.AuthSessionResponse](&s.ApiTestSuite).Data
	s.Equal(verifiedUser.ID, waSess.User.ID)

	// Verify email account was NOT created and WhatsApp LastUsedAt was updated.
	emailAccCount, err := s.client.UserAccount.Query().
		Where(useraccount.UserIDEQ(verifiedUser.ID), useraccount.ProviderEQ(domain.ProviderEmail)).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(0, emailAccCount)
	updatedWA, err := s.client.UserAccount.Get(ctx, waAcc.ID)
	s.Require().NoError(err)
	s.NotNil(updatedWA.LastUsedAt)

	// 2) Telegram OTP login via email lookup.
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "telegram",
		Target:  "MULTI-OTP@paroki.org",
	})
	s.OK()
	s.Equal(float64(987654321), gotTgPayload["chat_id"])
	tgMsg, _ := gotTgPayload["text"].(string)
	s.Contains(tgMsg, "Kode OTP login Domus")

	var tgCode1 string
	for _, field := range bytes.Fields([]byte(tgMsg)) {
		if len(field) == 6 {
			tgCode1 = string(field)
			break
		}
	}
	s.Require().Len(tgCode1, 6)

	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Channel: "telegram",
		Target:  "multi-otp@paroki.org",
		Code:    tgCode1,
	})
	s.OK()

	// 3) Telegram OTP login via verified WhatsApp phone lookup.
	gotTgPayload = nil
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "telegram",
		Target:  "081234567890",
	})
	s.OK()
	tgMsg2, _ := gotTgPayload["text"].(string)
	var tgCode2 string
	for _, field := range bytes.Fields([]byte(tgMsg2)) {
		if len(field) == 6 {
			tgCode2 = string(field)
			break
		}
	}
	s.Require().Len(tgCode2, 6)

	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Channel: "telegram",
		Target:  "+6281234567890",
		Code:    tgCode2,
	})
	s.OK()

	updatedTG, err := s.client.UserAccount.Get(ctx, tgAcc.ID)
	s.Require().NoError(err)
	s.NotNil(updatedTG.LastUsedAt)

	// Active user without telegram account -> 404 USER_NOT_FOUND.
	s.createTestUser("notg@paroki.org", "No TG", entuser.StatusActive)
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{
		Channel: "telegram",
		Target:  "notg@paroki.org",
	})
	s.NotFound()
	s.Equal("USER_NOT_FOUND", s.GetErrorResponse().Error.Code)
}

func (s *AuthHandlerSuite) TestOTP_RequestVerifyAndLogoutFlow() {
	// Unknown JSON field on /api/auth/otp -> 400 BAD_REQUEST
	s.Request("/api/auth/otp", http.MethodPost, map[string]any{
		"email":   "missing@paroki.org",
		"unknown": "field",
	})
	s.BadRequest()
	s.Equal("BAD_REQUEST", s.GetErrorResponse().Error.Code)

	// Unregistered user -> 404 USER_NOT_FOUND
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "missing@paroki.org"})
	s.NotFound()
	s.Equal("USER_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Disabled user -> 403 USER_DISABLED
	s.createTestUser("disabled@paroki.org", "Disabled", entuser.StatusDisabled)
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "disabled@paroki.org"})
	s.Forbidden()
	s.Equal("USER_DISABLED", s.GetErrorResponse().Error.Code)

	// Active user -> 200 and hashed OTP stored
	parishID := uuid.New()
	u := s.createTestUser("alice@paroki.org", "Alice", entuser.StatusActive)
	u, err := s.client.User.UpdateOneID(u.ID).SetDefaultParishID(parishID).Save(context.Background())
	s.Require().NoError(err)
	err = s.enforcer.AssignRole(context.Background(), authz.UserSubject(u.ID), user.RoleUser, authz.DomainGlobal)
	s.Require().NoError(err)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "ALICE@paroki.org"})
	s.OK()

	otpRes := testutil.DecodeAs[domain.RequestOTPResponse](&s.ApiTestSuite)
	s.Equal(300, otpRes.Data.ExpiresInSeconds)
	s.Equal("email", s.otpSender.lastChannel)
	s.Equal("alice@paroki.org", s.otpSender.lastTarget)
	s.Len(s.otpSender.lastCode, 6)

	otpRow, err := s.client.UserOTP.Query().Where(userotp.TargetEQ("alice@paroki.org")).Only(context.Background())
	s.Require().NoError(err)
	s.NotEqual(s.otpSender.lastCode, otpRow.CodeHash)

	// Cooldown -> 429 AUTH_OTP_RATE_LIMITED
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "alice@paroki.org"})
	s.Equal(http.StatusTooManyRequests, s.GetHttpResponse().StatusCode)
	s.Equal("AUTH_OTP_RATE_LIMITED", s.GetErrorResponse().Error.Code)

	// Unknown JSON field on /api/auth/otp/verify -> 400 BAD_REQUEST
	s.Request("/api/auth/otp/verify", http.MethodPost, map[string]any{
		"email":   "alice@paroki.org",
		"code":    "000000",
		"unknown": "field",
	})
	s.BadRequest()
	s.Equal("BAD_REQUEST", s.GetErrorResponse().Error.Code)

	// Wrong code -> 401 AUTH_OTP_INVALID, increments attempts
	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{Email: "alice@paroki.org", Code: "000000"})
	s.Unauthorized()
	s.Equal("AUTH_OTP_INVALID", s.GetErrorResponse().Error.Code)

	// Valid code -> 200, sets HttpOnly session cookie, deletes OTP, links account
	s.SetHeader("User-Agent", "TestBrowser/1.0")
	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Email: "alice@paroki.org",
		Code:  s.otpSender.lastCode,
	})
	s.OK()

	sessionRes := testutil.DecodeAs[domain.AuthSessionResponse](&s.ApiTestSuite)
	s.Equal(u.ID, sessionRes.Data.User.ID)
	s.NotNil(sessionRes.Data.User.LastLoginAt)
	s.Contains(sessionRes.Data.User.Roles, user.RoleUser)
	s.Require().NotNil(sessionRes.Data.User.ActiveParishID)
	s.Equal(parishID, *sessionRes.Data.User.ActiveParishID)

	var sessionCookie *http.Cookie
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == s.cfg.CookieName {
			sessionCookie = c
			break
		}
	}
	s.Require().NotNil(sessionCookie)
	s.True(sessionCookie.HttpOnly)
	s.False(sessionCookie.Secure)
	s.Equal(http.SameSiteLaxMode, sessionCookie.SameSite)
	s.Equal("/", sessionCookie.Path)

	sum := sha256.Sum256([]byte(sessionCookie.Value))
	tokenHash := hex.EncodeToString(sum[:])
	sessRow, err := s.client.UserSession.Query().Where(usersession.TokenHashEQ(tokenHash)).Only(context.Background())
	s.Require().NoError(err)
	s.Equal(u.ID, sessRow.UserID)
	s.Equal("TestBrowser/1.0", sessRow.UserAgent)

	// OTP hard-deleted after verify
	otpCount, err := s.client.UserOTP.Query().Where(userotp.TargetEQ("alice@paroki.org")).Count(context.Background())
	s.Require().NoError(err)
	s.Equal(0, otpCount)

	// UserAccount linked for email
	accCount, err := s.client.UserAccount.Query().
		Where(useraccount.ProviderEQ("email"), useraccount.SubjectEQ("alice@paroki.org")).
		Count(context.Background())
	s.Require().NoError(err)
	s.Equal(1, accCount)

	// Logout -> 204, clears cookie and deletes session row
	s.SetCookie(sessionCookie)
	s.Request("/api/auth/logout", http.MethodPost, nil)
	s.NoContent()

	sessCount, err := s.client.UserSession.Query().Where(usersession.TokenHashEQ(tokenHash)).Count(context.Background())
	s.Require().NoError(err)
	s.Equal(0, sessCount)
}

func (s *AuthHandlerSuite) TestOTP_ExpiredAndMaxAttempts() {
	s.createTestUser("bob@paroki.org", "Bob", entuser.StatusActive)

	// Expired OTP -> 401 AUTH_OTP_EXPIRED & hard-deleted
	_, err := s.client.UserOTP.Create().
		SetChannel("email").
		SetTarget("bob@paroki.org").
		SetCodeHash("deadbeef").
		SetExpiresAt(time.Now().UTC().Add(-1 * time.Minute)).
		Save(context.Background())
	s.Require().NoError(err)

	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{Email: "bob@paroki.org", Code: "123456"})
	s.Unauthorized()
	s.Equal("AUTH_OTP_EXPIRED", s.GetErrorResponse().Error.Code)

	count, err := s.client.UserOTP.Query().Where(userotp.TargetEQ("bob@paroki.org")).Count(context.Background())
	s.Require().NoError(err)
	s.Equal(0, count)

	// Max attempts -> deleted after 3rd failed attempt
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "bob@paroki.org"})
	s.OK()

	for range 3 {
		s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{Email: "bob@paroki.org", Code: "000000"})
		s.Unauthorized()
		s.Equal("AUTH_OTP_INVALID", s.GetErrorResponse().Error.Code)
	}

	count, err = s.client.UserOTP.Query().Where(userotp.TargetEQ("bob@paroki.org")).Count(context.Background())
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *AuthHandlerSuite) TestOAuth_RedirectAndCallbackFlows() {
	// Unknown provider -> 400 AUTH_OAUTH_PROVIDER_UNKNOWN
	s.Request("/api/auth/unknown", http.MethodGet, nil)
	s.BadRequest()
	s.Equal("AUTH_OAUTH_PROVIDER_UNKNOWN", s.GetErrorResponse().Error.Code)

	// Inject stub google provider
	stub := &stubOAuthProvider{
		name: "google",
		info: domain.OAuthUserInfo{
			Subject:       "google-sub-1",
			Email:         "carol@paroki.org",
			EmailVerified: true,
			Name:          "Carol",
			AvatarURL:     "https://img.example.test/carol.png",
		},
	}
	s.feature.RegisterOAuthProvider(stub)

	// Redirect sets state + sanitized redirect cookies
	s.Request("/api/auth/google?redirect=//evil.example.com", http.MethodGet, nil)
	s.Equal(http.StatusFound, s.GetHttpResponse().StatusCode)

	loc, err := url.Parse(s.GetHttpResponse().Header.Get("Location"))
	s.Require().NoError(err)
	state := loc.Query().Get("state")
	s.NotEmpty(state)
	s.Contains(loc.Query().Get("redirect_uri"), "/api/auth/google/callback")

	var stateCookie, redirectCookie *http.Cookie
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == "domus_oauth_state" {
			stateCookie = c
		}
		if c.Name == "domus_oauth_redirect" {
			redirectCookie = c
		}
	}
	s.Require().NotNil(stateCookie)
	s.Require().NotNil(redirectCookie)
	s.Equal("/", redirectCookie.Value)

	// Mismatched state -> 400 AUTH_OAUTH_STATE_INVALID
	s.SetCookie(stateCookie)
	s.Request("/api/auth/google/callback?code=abc&state=wrong", http.MethodGet, nil)
	s.BadRequest()
	s.Equal("AUTH_OAUTH_STATE_INVALID", s.GetErrorResponse().Error.Code)

	// Unregistered user -> 404 USER_NOT_FOUND
	s.ClearCookies()
	s.SetCookie(stateCookie)
	s.SetCookie(redirectCookie)
	s.Request("/api/auth/google/callback?code=abc&state="+url.QueryEscape(state), http.MethodGet, nil)
	s.NotFound()
	s.Equal("USER_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Unverified email -> 403 AUTH_OAUTH_EMAIL_UNVERIFIED
	s.createTestUser("carol@paroki.org", "Carol", entuser.StatusActive)
	stub.info.EmailVerified = false
	s.ClearCookies()
	s.SetCookie(stateCookie)
	s.Request("/api/auth/google/callback?code=abc&state="+url.QueryEscape(state), http.MethodGet, nil)
	s.Forbidden()
	s.Equal("AUTH_OAUTH_EMAIL_UNVERIFIED", s.GetErrorResponse().Error.Code)

	// Verified email + active user -> 302 to "/" + session cookie + account linked
	stub.info.EmailVerified = true
	s.ClearCookies()
	s.SetCookie(stateCookie)
	s.SetCookie(&http.Cookie{Name: "domus_oauth_redirect", Value: "/dashboard"})
	s.Request("/api/auth/google/callback?code=abc&state="+url.QueryEscape(state), http.MethodGet, nil)
	s.Equal(http.StatusFound, s.GetHttpResponse().StatusCode)
	s.Equal("/dashboard", s.GetHttpResponse().Header.Get("Location"))

	var sessionCookie *http.Cookie
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == s.cfg.CookieName {
			sessionCookie = c
		}
	}
	s.Require().NotNil(sessionCookie)
}

func (s *AuthHandlerSuite) TestEnsureUser_CreatesAndLinksIdempotently() {
	ctx := context.Background()
	u1, err := s.feature.EnsureUser(ctx, "Invite@Paroki.org", "Invited User", "08123456789", "email", "")
	s.Require().NoError(err)
	s.Equal("invite@paroki.org", u1.Email)
	s.Require().NotNil(u1.Phone)
	s.Equal("+628123456789", *u1.Phone)

	u2, err := s.feature.EnsureUser(ctx, "invite@paroki.org", "Invited User", "", "whatsapp", "08123456789")
	s.Require().NoError(err)
	s.Equal(u1.ID, u2.ID)
}

func (s *AuthHandlerSuite) TestCookieSecure_AutoDetectsHTTPS() {
	s.createTestUser("https@paroki.org", "HTTPS User", entuser.StatusActive)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "https@paroki.org"})
	s.OK()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.App.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true})
	}()
	s.T().Cleanup(func() {
		_ = s.App.Shutdown()
		s.NoError(<-errCh)
	})

	payload, err := json.Marshal(domain.VerifyOTPRequest{
		Email: "https@paroki.org",
		Code:  s.otpSender.lastCode,
	})
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodPost, "http://"+ln.Addr().String()+"/api/auth/otp/verify", bytes.NewReader(payload))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	s.Equal(http.StatusOK, resp.StatusCode)

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == s.cfg.CookieName {
			sessionCookie = c
			break
		}
	}
	s.Require().NotNil(sessionCookie)
	s.True(sessionCookie.Secure)
}
