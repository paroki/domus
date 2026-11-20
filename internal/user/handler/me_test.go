package handler_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type MeHandlerSuite struct {
	testutil.ApiTestSuite[any]
	client    *ent.Client
	enforcer  *authz.Enforcer
	otpSender *stubOTPSender
	feature   *user.Feature
	cfg       user.Config
}

func TestMeHandlerSuite(t *testing.T) {
	suite.Run(t, new(MeHandlerSuite))
}

func (s *MeHandlerSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.otpSender = &stubOTPSender{}
	s.cfg = user.Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
		OTPCooldown:    0,
		SessionTTL:     24 * time.Hour,
		CookieName:     "domus_session",
	}
	s.feature = user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.otpSender))
	s.Require().NoError(s.feature.RegisterPermissions(context.Background(), s.enforcer))
	s.WithFeatures(s.feature)
}

func (s *MeHandlerSuite) loginUser(email, name, ua string) (*domain.UserResponse, *http.Cookie) {
	ctx := context.Background()
	u, err := s.feature.EnsureUser(ctx, email, name, "", "email", email)
	s.Require().NoError(err)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: email})
	s.OK()
	s.SetHeader("User-Agent", ua)
	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Email: email,
		Code:  s.otpSender.lastCode,
	})
	s.OK()
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == s.cfg.CookieName {
			return u, c
		}
	}
	s.FailNow("session cookie not found")
	return nil, nil
}

func (s *MeHandlerSuite) TestGetAndUpdateMe_PhoneE164AndStrictJSON() {
	_, cookie := s.loginUser("me@paroki.org", "Me User", "Browser/1")
	s.SetCookie(cookie)

	// GET /api/me
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal("me@paroki.org", me.Email)
	s.Contains(me.Roles, user.RoleUser)

	// Unknown JSON field -> 400 BAD_REQUEST
	s.Request("/api/me", http.MethodPatch, `{"name":"New","unknownField":true}`)
	s.BadRequest()
	s.Equal("BAD_REQUEST", s.GetErrorResponse().Error.Code)

	// Invalid phone -> 422 VALIDATION_ERROR with field "phone"
	badPhone := "not-a-phone"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Phone: &badPhone})
	s.UnprocessableEntity()
	errResp := s.GetErrorResponse().Error
	s.Equal("VALIDATION_ERROR", errResp.Code)
	s.Require().NotEmpty(errResp.Details)
	s.Equal("phone", errResp.Details[0].Field)

	// Valid local phone -> normalized to +6281234567890
	newName := "Updated Me"
	localPhone := "0812-3456-7890"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Name: &newName, Phone: &localPhone})
	s.OK()
	updated := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal("Updated Me", updated.Name)
	s.Require().NotNil(updated.Phone)
	s.Equal("+6281234567890", *updated.Phone)
}

func (s *MeHandlerSuite) TestSwitchSessionParish_AndMultiDeviceSessions() {
	u, cookie1 := s.loginUser("multi@paroki.org", "Multi Device", "Firefox/130")
	_, cookie2 := s.loginUser("multi@paroki.org", "Multi Device", "Mobile/1.0")

	// Switch session parish on device 1
	parishID := uuid.Must(uuid.NewV7())
	s.SetCookie(cookie1)
	s.Request("/api/me/session", http.MethodPatch, domain.SwitchSessionParishRequest{ActiveParishID: parishID})
	s.NoContent()

	// Verify User.default_parish_id initialized
	entU, err := s.client.User.Query().Where(entuser.IDEQ(u.ID)).Only(context.Background())
	s.Require().NoError(err)
	s.Require().NotNil(entU.DefaultParishID)
	s.Equal(parishID, *entU.DefaultParishID)

	// Search sessions as device 1 -> 2 sessions, only device 1 has isCurrent=true
	s.Request("/api/me/sessions/search", http.MethodPost, domain.SearchSessionsRequest{Limit: 10})
	s.OK()
	paged := testutil.DecodeAs[[]domain.UserSessionResponse](&s.ApiTestSuite)
	s.Len(paged.Data, 2)

	var currentID, otherID uuid.UUID
	for _, sess := range paged.Data {
		if sess.IsCurrent {
			currentID = sess.ID
		} else {
			otherID = sess.ID
		}
	}
	s.NotEqual(uuid.Nil, currentID)
	s.NotEqual(uuid.Nil, otherID)

	// Revoke other session -> 204
	s.Request("/api/me/sessions/"+otherID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Device 2 is now unauthorized
	s.ClearCookies()
	s.SetCookie(cookie2)
	s.Request("/api/me", http.MethodGet, nil)
	s.Unauthorized()

	// Revoke current session -> 204 and clears cookie
	s.ClearCookies()
	s.SetCookie(cookie1)
	s.Request("/api/me/sessions/"+currentID.String(), http.MethodDelete, nil)
	s.NoContent()

	// Non-existent session -> 404 USER_SESSION_NOT_FOUND
	_, freshCookie := s.loginUser("multi@paroki.org", "Multi Device", "Chrome/130")
	s.ClearCookies()
	s.SetCookie(freshCookie)
	s.Request("/api/me/sessions/"+uuid.Must(uuid.NewV7()).String(), http.MethodDelete, nil)
	s.NotFound()
	s.Equal("USER_SESSION_NOT_FOUND", s.GetErrorResponse().Error.Code)
}

func (s *MeHandlerSuite) TestSearchAndUnlinkAccounts_LastAccountGuard() {
	_, cookie := s.loginUser("acct@paroki.org", "Acct User", "Browser/1")
	s.SetCookie(cookie)

	// Unknown account ID returns 404 USER_ACCOUNT_NOT_FOUND even when user has only 1 account
	s.Request("/api/me/accounts/"+uuid.Must(uuid.NewV7()).String(), http.MethodDelete, nil)
	s.NotFound()
	s.Equal("USER_ACCOUNT_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Only 1 account -> cannot unlink (409 USER_LAST_ACCOUNT_REQUIRED)
	s.Request("/api/me/accounts/search", http.MethodPost, nil)
	s.OK()
	accounts := testutil.DecodeAs[[]domain.UserAccountResponse](&s.ApiTestSuite).Data
	s.Require().Len(accounts, 1)

	s.Request("/api/me/accounts/"+accounts[0].ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("USER_LAST_ACCOUNT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Link a second account via EnsureUser -> can unlink 1 of 2
	_, err := s.feature.EnsureUser(context.Background(), "acct@paroki.org", "Acct User", "", "google", "google-sub-99")
	s.Require().NoError(err)

	s.Request("/api/me/accounts/search", http.MethodPost, nil)
	s.OK()
	accounts2 := testutil.DecodeAs[[]domain.UserAccountResponse](&s.ApiTestSuite).Data
	s.Require().Len(accounts2, 2)

	s.Request("/api/me/accounts/"+accounts2[0].ID.String(), http.MethodDelete, nil)
	s.NoContent()
}
