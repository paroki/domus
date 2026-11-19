package handler_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/paroki/domus/ent"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/useraccount"
	"github.com/paroki/domus/ent/userotp"
	"github.com/paroki/domus/ent/usersession"
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
	s.otpSender = &stubOTPSender{}
	s.cfg = user.Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 3,
		OTPCooldown:    60 * time.Second,
		SessionTTL:     24 * time.Hour,
		CookieName:     "domus_session",
		CookieSecure:   false,
		GoogleID:       "google-id",
		GoogleSecret:   "google-secret",
		GitHubID:       "github-id",
		GitHubSecret:   "github-secret",
	}
	s.feature = user.New(s.client, nil, s.cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.otpSender))
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
}

func (s *AuthHandlerSuite) TestOTP_RequestVerifyAndLogoutFlow() {
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
	u := s.createTestUser("alice@paroki.org", "Alice", entuser.StatusActive)
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

	var sessionCookie *http.Cookie
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == s.cfg.CookieName {
			sessionCookie = c
			break
		}
	}
	s.Require().NotNil(sessionCookie)
	s.True(sessionCookie.HttpOnly)
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
	u1, err := s.feature.EnsureUser(ctx, "Invite@Paroki.org", "Invited User", "+628123", "email", "")
	s.Require().NoError(err)
	s.Equal("invite@paroki.org", u1.Email)

	u2, err := s.feature.EnsureUser(ctx, "invite@paroki.org", "Invited User", "", "whatsapp", "+628123")
	s.Require().NoError(err)
	s.Equal(u1.ID, u2.ID)
}
