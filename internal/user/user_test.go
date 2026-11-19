package user_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/usersession"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
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
	uResp, err := s.feature.EnsureUser(ctx, "botuser@paroki.org", "Bot User", "+628123456789", "telegram", "tg-999")
	s.Require().NoError(err)
	_, err = s.feature.EnsureUser(ctx, "botuser@paroki.org", "Bot User", "+628123456789", "whatsapp", "+628123456789")
	s.Require().NoError(err)

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

	// Override takes precedence over account active_parish_id
	parish2 := uuid.Must(uuid.NewV7())
	pWa, err := s.feature.ResolveByChannel(ctx, "whatsapp", "+628123456789", parish2)
	s.Require().NoError(err)
	s.Equal("whatsapp", pWa.Channel)
	s.Equal(parish2, pWa.ParishID)

	// Disabled user -> 403 USER_DISABLED on ResolveByChannel
	_, err = s.client.User.UpdateOneID(uResp.ID).SetStatus(entuser.StatusDisabled).Save(ctx)
	s.Require().NoError(err)
	_, err = s.feature.ResolveByChannel(ctx, "telegram", "tg-999", uuid.Nil)
	s.ErrorIs(err, domain.ErrUserDisabled)
}
