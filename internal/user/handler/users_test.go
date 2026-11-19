package handler_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type UsersHandlerSuite struct {
	testutil.ApiTestSuite[any]
	client    *ent.Client
	enforcer  *authz.Enforcer
	otpSender *stubOTPSender
	feature   *user.Feature
	cfg       user.Config
}

func TestUsersHandlerSuite(t *testing.T) {
	suite.Run(t, new(UsersHandlerSuite))
}

func (s *UsersHandlerSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.otpSender = &stubOTPSender{}
	s.cfg = user.Config{
		SuperAdminEmail: "super@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     0,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.feature = user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.otpSender))
	s.Require().NoError(s.feature.RegisterPermissions(context.Background(), s.enforcer))
	s.Require().NoError(s.feature.Bootstrap(context.Background()))
	s.WithFeatures(s.feature)
}

func (s *UsersHandlerSuite) login(email string) *http.Cookie {
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: email})
	s.OK()
	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Email: email,
		Code:  s.otpSender.lastCode,
	})
	s.OK()
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == s.cfg.CookieName {
			return c
		}
	}
	s.FailNow("session cookie not found")
	return nil
}

func (s *UsersHandlerSuite) TestAdminUsers_RBACAndCRUDLifecycle() {
	ctx := context.Background()

	// Unauthenticated -> 401
	s.Request("/api/users/search", http.MethodPost, nil)
	s.Unauthorized()

	// Non-super-admin -> 403
	_, err := s.feature.EnsureUser(ctx, "regular@paroki.org", "Regular", "", "email", "regular@paroki.org")
	s.Require().NoError(err)
	regularCookie := s.login("regular@paroki.org")
	s.SetCookie(regularCookie)
	s.Request("/api/users/search", http.MethodPost, nil)
	s.Forbidden()

	// Login as Super Admin
	s.ClearCookies()
	adminCookie := s.login("super@paroki.org")
	s.SetCookie(adminCookie)

	// Search finds super + regular
	s.Request("/api/users/search", http.MethodPost, domain.SearchUsersRequest{Q: "regular"})
	s.OK()
	searchRes := testutil.DecodeAs[[]domain.UserResponse](&s.ApiTestSuite)
	s.Require().Len(searchRes.Data, 1)
	s.Equal("regular@paroki.org", searchRes.Data[0].Email)

	// Create new Super Admin
	phone := "081122334455"
	s.Request("/api/users", http.MethodPost, domain.AdminCreateUserRequest{
		Email:        "second-admin@paroki.org",
		Name:         "Second Admin",
		Phone:        &phone,
		IsSuperAdmin: true,
	})
	s.Created()
	created := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal("second-admin@paroki.org", created.Email)
	s.Contains(created.Roles, user.RoleSuperAdmin)
	s.Require().NotNil(created.Phone)
	s.Equal("+6281122334455", *created.Phone)

	// Duplicate email -> 409 USER_EMAIL_CONFLICT
	s.Request("/api/users", http.MethodPost, domain.AdminCreateUserRequest{
		Email: "second-admin@paroki.org",
		Name:  "Dup",
	})
	s.Conflict()
	s.Equal("USER_EMAIL_CONFLICT", s.GetErrorResponse().Error.Code)

	// GET /api/users/{id}
	s.Request("/api/users/"+created.ID.String(), http.MethodGet, nil)
	s.OK()
	got := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal(created.ID, got.ID)

	// Demote second admin (still 1 active super admin left) -> 200
	demote := false
	s.Request("/api/users/"+created.ID.String(), http.MethodPatch, domain.AdminUpdateUserRequest{
		IsSuperAdmin: &demote,
	})
	s.OK()
	demoted := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.NotContains(demoted.Roles, user.RoleSuperAdmin)

	// Now only 1 active Super Admin ("super@paroki.org") remains:
	// Demoting, disabling, or deleting the last active Super Admin -> 409 USER_LAST_SUPER_ADMIN_REQUIRED
	s.Request("/api/users/search", http.MethodPost, domain.SearchUsersRequest{Q: "super@paroki.org"})
	s.OK()
	superID := testutil.DecodeAs[[]domain.UserResponse](&s.ApiTestSuite).Data[0].ID

	s.Request("/api/users/"+superID.String(), http.MethodPatch, domain.AdminUpdateUserRequest{IsSuperAdmin: &demote})
	s.Conflict()
	s.Equal("USER_LAST_SUPER_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	disabled := "disabled"
	s.Request("/api/users/"+superID.String(), http.MethodPatch, domain.AdminUpdateUserRequest{Status: &disabled})
	s.Conflict()
	s.Equal("USER_LAST_SUPER_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	s.Request("/api/users/"+superID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("USER_LAST_SUPER_ADMIN_REQUIRED", s.GetErrorResponse().Error.Code)

	// Disabling regular user revokes their session immediately (401 on old cookie, 403 USER_DISABLED on login)
	s.Request("/api/users/"+searchRes.Data[0].ID.String(), http.MethodPatch, domain.AdminUpdateUserRequest{Status: &disabled})
	s.OK()

	s.ClearCookies()
	s.SetCookie(regularCookie)
	s.Request("/api/me", http.MethodGet, nil)
	s.Unauthorized()

	s.ClearCookies()
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: "regular@paroki.org"})
	s.Forbidden()
	s.Equal("USER_DISABLED", s.GetErrorResponse().Error.Code)

	// Soft-deleting second-admin excludes them from GET and search
	s.ClearCookies()
	s.SetCookie(adminCookie)
	s.Request("/api/users/"+created.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.Request("/api/users/"+created.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Equal("USER_NOT_FOUND", s.GetErrorResponse().Error.Code)

	s.Request("/api/users/"+uuid.Must(uuid.NewV7()).String(), http.MethodGet, nil)
	s.NotFound()
}
