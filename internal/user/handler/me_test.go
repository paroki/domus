package handler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/parokimember"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/useraccount"
	"github.com/paroki/domus/ent/userotp"
	"github.com/paroki/domus/ent/usertoken"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/telegram"
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

func (s *MeHandlerSuite) TestSwitchSessionParoki_AndMultiDeviceSessions() {
	ctx := context.Background()
	u, cookie1 := s.loginUser("multi@paroki.org", "Multi Device", "Firefox/130")
	_, cookie2 := s.loginUser("multi@paroki.org", "Multi Device", "Mobile/1.0")

	d, err := s.client.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	s.Require().NoError(err)
	p, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p-01").SetNama("Paroki 1").Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.ParokiMember.Create().SetParokiID(p.ID).SetUserID(u.ID).SetRole(parokimember.RoleMember).Save(ctx)
	s.Require().NoError(err)

	// Switch session paroki on device 1
	parokiID := p.ID
	s.SetCookie(cookie1)
	s.Request("/api/me/session", http.MethodPatch, domain.SwitchSessionParokiRequest{ActiveParokiID: parokiID})
	s.NoContent()

	// Verify User.default_paroki_id initialized
	entU, err := s.client.User.Query().Where(entuser.IDEQ(u.ID)).Only(context.Background())
	s.Require().NoError(err)
	s.Require().NotNil(entU.DefaultParokiID)
	s.Equal(parokiID, *entU.DefaultParokiID)

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

func (s *MeHandlerSuite) TestParokiMembershipEnforcement_SwitchAndHeader() {
	ctx := context.Background()
	d, err := s.client.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	s.Require().NoError(err)
	memberParoki, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p-mem").SetNama("Member Paroki").Save(ctx)
	s.Require().NoError(err)
	otherParoki, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p-oth").SetNama("Other Paroki").Save(ctx)
	s.Require().NoError(err)
	inactiveParoki, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p-ina").SetNama("Inactive Paroki").SetActive(false).Save(ctx)
	s.Require().NoError(err)

	regularUser, regularCookie := s.loginUser("regular@paroki.org", "Regular User", "Browser/1")
	_, err = s.client.ParokiMember.Create().SetParokiID(memberParoki.ID).SetUserID(regularUser.ID).SetRole(parokimember.RoleMember).Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.ParokiMember.Create().SetParokiID(inactiveParoki.ID).SetUserID(regularUser.ID).SetRole(parokimember.RoleMember).Save(ctx)
	s.Require().NoError(err)

	s.SetCookie(regularCookie)

	// Non-member paroki switch -> 403
	s.Request("/api/me/session", http.MethodPatch, domain.SwitchSessionParokiRequest{ActiveParokiID: otherParoki.ID})
	s.Forbidden()

	// Inactive paroki switch -> 403
	s.Request("/api/me/session", http.MethodPatch, domain.SwitchSessionParokiRequest{ActiveParokiID: inactiveParoki.ID})
	s.Forbidden()

	// Member of active paroki switch -> 204
	s.Request("/api/me/session", http.MethodPatch, domain.SwitchSessionParokiRequest{ActiveParokiID: memberParoki.ID})
	s.NoContent()

	// X-Paroki-Id for non-member paroki -> 403
	s.SetHeader("X-Paroki-Id", otherParoki.ID.String())
	s.Request("/api/me", http.MethodGet, nil)
	s.Forbidden()

	// X-Paroki-Id for member paroki -> 200
	s.SetHeader("X-Paroki-Id", memberParoki.ID.String())
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()

	// Super-admin can switch to any active paroki -> 204
	s.Headers.Del("X-Paroki-Id")
	s.ClearCookies()
	adminUser, adminCookie := s.loginUser("super@paroki.org", "Super Admin", "Browser/1")
	s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(adminUser.ID), domain.RoleSuperAdmin, authz.DomainGlobal))
	s.SetCookie(adminCookie)
	s.Request("/api/me/session", http.MethodPatch, domain.SwitchSessionParokiRequest{ActiveParokiID: otherParoki.ID})
	s.NoContent()
}

func (s *MeHandlerSuite) TestOnboardedFlag_UpdateAndReactivateReset() {
	ctx := context.Background()
	u, cookie := s.loginUser("onboard@paroki.org", "Onboard User", "Browser/1")
	s.SetCookie(cookie)

	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.False(me.Onboarded)
	s.Equal([]string{"email"}, me.LinkedProviders)

	onboardedTrue := true
	// Missing gender -> 422 VALIDATION_ERROR
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Onboarded: &onboardedTrue})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("gender", s.GetErrorResponse().Error.Details[0].Field)

	// Gender provided, missing birthDate -> 422 VALIDATION_ERROR
	gender := "pria"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Gender: &gender, Onboarded: &onboardedTrue})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("birthDate", s.GetErrorResponse().Error.Details[0].Field)

	// Both gender and birthDate provided -> 200 OK
	birthDate := "1995-05-10"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{
		Gender:    &gender,
		BirthDate: &birthDate,
		Onboarded: &onboardedTrue,
	})
	s.OK()
	updated := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.True(updated.Onboarded)
	s.Require().NotNil(updated.Gender)
	s.Equal("pria", *updated.Gender)
	s.Require().NotNil(updated.BirthDate)
	s.Equal("1995-05-10", *updated.BirthDate)
	s.Equal([]string{"email"}, updated.LinkedProviders)

	// Soft-deleted user reactivated via EnsureUser resets onboarded, gender, and birthDate.
	_, err := s.client.User.UpdateOneID(u.ID).SetDeletedAt(time.Now().UTC()).Save(ctx)
	s.Require().NoError(err)
	reactivated, err := s.feature.EnsureUser(ctx, "onboard@paroki.org", "Onboard User", "", "email", "onboard@paroki.org")
	s.Require().NoError(err)
	s.False(reactivated.Onboarded)
	s.Nil(reactivated.Gender)
	s.Nil(reactivated.BirthDate)
}

func (s *MeHandlerSuite) TestUpdateMe_GenderAndBirthDateValidation() {
	_, cookie := s.loginUser("biodata@paroki.org", "Biodata User", "Browser/1")
	s.SetCookie(cookie)

	empty := "  "
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Gender: &empty})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("gender", s.GetErrorResponse().Error.Details[0].Field)

	badGender := "other"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Gender: &badGender})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("gender", s.GetErrorResponse().Error.Details[0].Field)

	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{BirthDate: &empty})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("birthDate", s.GetErrorResponse().Error.Details[0].Field)

	badFormat := "10-05-1995"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{BirthDate: &badFormat})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("birthDate", s.GetErrorResponse().Error.Details[0].Field)

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tooYoung := today.AddDate(-12, 0, 0).Format("2006-01-02")
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{BirthDate: &tooYoung})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("birthDate", s.GetErrorResponse().Error.Details[0].Field)

	tooOld := today.AddDate(-121, 0, 0).Format("2006-01-02")
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{BirthDate: &tooOld})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	s.Equal("birthDate", s.GetErrorResponse().Error.Details[0].Field)

	minValid := today.AddDate(-120, 0, 0).Format("2006-01-02")
	maxValid := today.AddDate(-13, 0, 0).Format("2006-01-02")
	wanita := "wanita"
	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{Gender: &wanita, BirthDate: &minValid})
	s.OK()
	resMin := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Require().NotNil(resMin.Gender)
	s.Equal("wanita", *resMin.Gender)
	s.Require().NotNil(resMin.BirthDate)
	s.Equal(minValid, *resMin.BirthDate)

	s.Request("/api/me", http.MethodPatch, domain.UpdateMeRequest{BirthDate: &maxValid})
	s.OK()
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	fetched := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Require().NotNil(fetched.Gender)
	s.Equal("wanita", *fetched.Gender)
	s.Require().NotNil(fetched.BirthDate)
	s.Equal(maxValid, *fetched.BirthDate)
}

func (s *MeHandlerSuite) TestLinkWhatsAppAndTelegramToken() {
	ctx := context.Background()

	// Before configuring Gowa/Telegram, linking endpoints return 400 AUTH_CHANNEL_DISABLED.
	_, disabledCookie := s.loginUser("disabledlink@paroki.org", "Disabled Link", "Browser/1")
	s.SetCookie(disabledCookie)
	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "whatsapp",
		Target:  "081234567890",
	})
	s.BadRequest()
	s.Equal("AUTH_CHANNEL_DISABLED", s.GetErrorResponse().Error.Code)

	s.Request("/api/me/accounts/telegram/token", http.MethodPost, nil)
	s.BadRequest()
	s.Equal("AUTH_CHANNEL_DISABLED", s.GetErrorResponse().Error.Code)

	// Configure GOWA + Telegram clients and cooldown for linking tests.
	var lastGowaMsg map[string]string
	gowaSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Require().NoError(json.NewDecoder(r.Body).Decode(&lastGowaMsg))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"SUCCESS"}`))
	}))
	defer gowaSrv.Close()

	gowaClient, err := gowa.New(gowa.Config{BaseURL: gowaSrv.URL})
	s.Require().NoError(err)
	tgClient, err := telegram.New(telegram.Config{
		BotToken:      "123:SECRET",
		BotUsername:   "domus_link_bot",
		WebhookSecret: "hook-secret",
		BaseURL:       gowaSrv.URL,
	})
	s.Require().NoError(err)

	linkCfg := s.cfg
	linkCfg.OTPCooldown = 60 * time.Second
	linkFeature := user.New(
		s.client,
		s.enforcer,
		linkCfg,
		slog.New(slog.DiscardHandler),
		user.WithOTPSender(s.otpSender),
		user.WithGowaClient(gowaClient),
		user.WithTelegramClient(tgClient),
	)
	s.WithFeatures(linkFeature)

	s.ClearCookies()
	user1, cookie1 := s.loginUser("wa1@paroki.org", "WA One", "Browser/1")
	s.ClearCookies()
	_, cookie2 := s.loginUser("wa2@paroki.org", "WA Two", "Browser/1")

	// User 1 validates strict JSON, channel enum, and invalid phone.
	s.SetCookie(cookie1)
	s.Request("/api/me/accounts/otp", http.MethodPost, `{"channel":"whatsapp","target":"081234567890","extra":1}`)
	s.BadRequest()
	s.Equal("BAD_REQUEST", s.GetErrorResponse().Error.Code)

	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "email",
		Target:  "wa1@paroki.org",
	})
	s.UnprocessableEntity()
	s.Equal("channel", s.GetErrorResponse().Error.Details[0].Field)

	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "whatsapp",
		Target:  "not-a-phone",
	})
	s.UnprocessableEntity()
	s.Equal("target", s.GetErrorResponse().Error.Details[0].Field)

	// User 1 requests WhatsApp verification OTP.
	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "whatsapp",
		Target:  "0812-3456-7890",
	})
	s.OK()
	s.Equal("6281234567890@s.whatsapp.net", lastGowaMsg["phone"])
	s.Contains(lastGowaMsg["message"], "Kode verifikasi WhatsApp Domus")

	// Cooldown blocks immediate second OTP request for same number.
	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "whatsapp",
		Target:  "+6281234567890",
	})
	s.Equal(http.StatusTooManyRequests, s.GetHttpResponse().StatusCode)
	s.Equal("AUTH_OTP_RATE_LIMITED", s.GetErrorResponse().Error.Code)

	otpRe := regexp.MustCompile(`\b[0-9]{6}\b`)
	code1 := otpRe.FindString(lastGowaMsg["message"])
	s.Require().Len(code1, 6)

	// Wrong code -> 401 AUTH_OTP_INVALID.
	s.Request("/api/me/accounts/verify", http.MethodPost, domain.VerifyLinkAccountRequest{
		Channel: "whatsapp",
		Target:  "+6281234567890",
		Code:    "000000",
	})
	s.Unauthorized()
	s.Equal("AUTH_OTP_INVALID", s.GetErrorResponse().Error.Code)

	// Valid code -> links WhatsApp account, updates User.phone, updates linkedProviders.
	s.Request("/api/me/accounts/verify", http.MethodPost, domain.VerifyLinkAccountRequest{
		Channel: "whatsapp",
		Target:  "081234567890",
		Code:    code1,
	})
	s.OK()
	waAcc := testutil.DecodeAs[domain.UserAccountResponse](&s.ApiTestSuite).Data
	s.Equal("whatsapp", waAcc.Provider)
	s.Equal("+6281234567890", waAcc.Subject)

	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me1 := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Require().NotNil(me1.Phone)
	s.Equal("+6281234567890", *me1.Phone)
	s.Equal([]string{"email", "whatsapp"}, me1.LinkedProviders)

	// User 2 cannot request OTP for a number already linked to User 1 -> 409 USER_ACCOUNT_CONFLICT.
	s.ClearCookies()
	s.SetCookie(cookie2)
	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "whatsapp",
		Target:  "081234567890",
	})
	s.Conflict()
	s.Equal("USER_ACCOUNT_CONFLICT", s.GetErrorResponse().Error.Code)

	// User 1 can replace their own WhatsApp number with a new one.
	s.ClearCookies()
	s.SetCookie(cookie1)
	s.Request("/api/me/accounts/otp", http.MethodPost, domain.LinkAccountOTPRequest{
		Channel: "whatsapp",
		Target:  "081999888777",
	})
	s.OK()
	code2 := otpRe.FindString(lastGowaMsg["message"])
	s.Require().Len(code2, 6)

	s.Request("/api/me/accounts/verify", http.MethodPost, domain.VerifyLinkAccountRequest{
		Channel: "whatsapp",
		Target:  "+6281999888777",
		Code:    code2,
	})
	s.OK()
	waAcc2 := testutil.DecodeAs[domain.UserAccountResponse](&s.ApiTestSuite).Data
	s.Equal("+6281999888777", waAcc2.Subject)

	waCount, err := s.client.UserAccount.Query().
		Where(useraccount.UserIDEQ(user1.ID), useraccount.ProviderEQ(domain.ProviderWhatsApp)).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, waCount)

	// Unlinking WhatsApp account leaves User.phone untouched.
	s.Request("/api/me/accounts/"+waAcc2.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	entU1, err := s.client.User.Get(ctx, user1.ID)
	s.Require().NoError(err)
	s.Require().NotNil(entU1.Phone)
	s.Equal("+6281999888777", *entU1.Phone)

	// Telegram token endpoint works without cooldown even when OTPCooldown is 60s.
	s.Request("/api/me/accounts/telegram/token", http.MethodPost, nil)
	s.OK()
	tg1 := testutil.DecodeAs[domain.TelegramLinkTokenResponse](&s.ApiTestSuite).Data
	s.Equal("domus_link_bot", tg1.BotUsername)
	s.Contains(tg1.DeepLinkURL, "https://t.me/domus_link_bot?start=")
	s.Equal(300, tg1.ExpiresInSeconds)

	s.Request("/api/me/accounts/telegram/token", http.MethodPost, nil)
	s.OK()
	tg2 := testutil.DecodeAs[domain.TelegramLinkTokenResponse](&s.ApiTestSuite).Data
	s.NotEqual(tg1.DeepLinkURL, tg2.DeepLinkURL)

	tgRows, err := s.client.UserOTP.Query().
		Where(userotp.ChannelEQ(domain.ProviderTelegram), userotp.TargetEQ(user1.ID.String())).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, tgRows)
}

func (s *MeHandlerSuite) TestPersonalAccessTokens_LifecycleAndBearerAuth() {
	ctx := context.Background()
	u, cookie := s.loginUser("pat-user@paroki.org", "PAT User", "Browser/1")
	otherUser, otherCookie := s.loginUser("pat-other@paroki.org", "Other User", "Browser/2")

	// Setup paroki membership for X-Paroki-Id tenant resolution test.
	d, err := s.client.Keuskupan.Create().SetKode("kaj-pat").SetNama("KAJ PAT").Save(ctx)
	s.Require().NoError(err)
	p, err := s.client.Paroki.Create().SetKeuskupanID(d.ID).SetKode("p-pat").SetNama("Paroki PAT").Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.ParokiMember.Create().SetParokiID(p.ID).SetUserID(u.ID).SetRole(parokimember.RoleMember).Save(ctx)
	s.Require().NoError(err)

	// Whitespace-only name returns 422 VALIDATION_ERROR.
	s.SetCookie(cookie)
	s.Request("/api/me/tokens", http.MethodPost, domain.CreateUserTokenRequest{Name: "   "})
	s.UnprocessableEntity()

	// Create PAT with 30-day expiry.
	days := 30
	s.Request("/api/me/tokens", http.MethodPost, domain.CreateUserTokenRequest{
		Name:          "CLI Exporter",
		ExpiresInDays: &days,
	})
	s.Created()
	created := testutil.DecodeAs[domain.CreateUserTokenResponse](&s.ApiTestSuite).Data
	s.Equal("CLI Exporter", created.Name)
	s.True(len(created.Token) == 47 && created.Token[:4] == "dms_")
	s.Equal(created.Token[:12], created.TokenPrefix)
	s.NotNil(created.ExpiresAt)
	s.Nil(created.LastUsedAt)

	// Search tokens -> returns prefix, never plaintext.
	s.Request("/api/me/tokens/search", http.MethodPost, domain.SearchTokensRequest{Limit: 10})
	s.OK()
	listed := testutil.DecodeAs[[]domain.UserTokenResponse](&s.ApiTestSuite)
	s.Require().Len(listed.Data, 1)
	s.Equal(created.ID, listed.Data[0].ID)
	s.Equal(created.TokenPrefix, listed.Data[0].TokenPrefix)

	// Authenticate via Authorization: Bearer <token> + X-Paroki-Id on /api/me.
	s.ClearCookies()
	s.SetHeader("X-Paroki-Id", p.ID.String())
	s.RequestWithToken("/api/me", http.MethodGet, nil, created.Token)
	s.OK()
	me := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal(u.ID, me.ID)
	s.Require().NotNil(me.ActiveParokiID)
	s.Equal(p.ID, *me.ActiveParokiID)

	// Verify last_used_at was updated and expires_at was NOT extended.
	tokRow, err := s.client.UserToken.Get(ctx, created.ID)
	s.Require().NoError(err)
	s.NotNil(tokRow.LastUsedAt)
	s.Equal(created.ExpiresAt.Unix(), tokRow.ExpiresAt.Unix())

	// Verify PrincipalResolver resolves Bearer PAT (for /mcp).
	resolver := s.feature.PrincipalResolver()
	s.App.Post("/mcp-check", func(c fiber.Ctx) error {
		resolved, ok := resolver(c)
		s.Require().True(ok)
		s.Equal(u.ID, resolved.UserID)
		s.Equal(p.ID, resolved.ParokiID)
		s.Equal("api", resolved.Channel)
		s.Equal(uuid.Nil, resolved.SessionID)
		return c.SendStatus(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/mcp-check", nil)
	req.Header.Set("Authorization", "Bearer "+created.Token)
	req.Header.Set("X-Paroki-Id", p.ID.String())
	resp, err := s.App.Test(req)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// Cross-user revoke returns 404 USER_TOKEN_NOT_FOUND.
	s.SetHeader("X-Paroki-Id", "")
	s.SetCookie(otherCookie)
	s.Request("/api/me/tokens/"+created.ID.String(), http.MethodDelete, nil)
	s.NotFound()
	s.Equal("USER_TOKEN_NOT_FOUND", s.GetErrorResponse().Error.Code)
	_ = otherUser

	// Expired token is rejected with 401 and deleted.
	_, err = s.client.UserToken.UpdateOneID(created.ID).
		SetExpiresAt(time.Now().UTC().Add(-time.Minute)).
		Save(ctx)
	s.Require().NoError(err)
	s.ClearCookies()
	s.RequestWithToken("/api/me", http.MethodGet, nil, created.Token)
	s.Unauthorized()
	exists, err := s.client.UserToken.Query().Where(usertoken.IDEQ(created.ID)).Exist(ctx)
	s.Require().NoError(err)
	s.False(exists)

	// Create never-expiring token, then disable user -> token revoked & 403 USER_DISABLED.
	s.SetCookie(cookie)
	s.Request("/api/me/tokens", http.MethodPost, domain.CreateUserTokenRequest{Name: "MCP Token"})
	s.Created()
	neverExp := testutil.DecodeAs[domain.CreateUserTokenResponse](&s.ApiTestSuite).Data
	s.Nil(neverExp.ExpiresAt)

	_, err = s.client.User.UpdateOneID(u.ID).SetStatus(entuser.StatusDisabled).Save(ctx)
	s.Require().NoError(err)
	s.ClearCookies()
	s.RequestWithToken("/api/me", http.MethodGet, nil, neverExp.Token)
	s.Forbidden()
	s.Equal("USER_DISABLED", s.GetErrorResponse().Error.Code)
	countAfterDisable, err := s.client.UserToken.Query().Where(usertoken.UserIDEQ(u.ID)).Count(ctx)
	s.Require().NoError(err)
	s.Equal(0, countAfterDisable)

	// Re-enable user, create token, and revoke it via DELETE /api/me/tokens/:id.
	_, err = s.client.User.UpdateOneID(u.ID).SetStatus(entuser.StatusActive).Save(ctx)
	s.Require().NoError(err)
	_, freshCookie := s.loginUser("pat-user@paroki.org", "PAT User", "Browser/3")
	s.SetCookie(freshCookie)
	s.Request("/api/me/tokens", http.MethodPost, domain.CreateUserTokenRequest{Name: "Temp Token"})
	s.Created()
	tempTok := testutil.DecodeAs[domain.CreateUserTokenResponse](&s.ApiTestSuite).Data

	s.Request("/api/me/tokens/"+tempTok.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	s.ClearCookies()
	s.RequestWithToken("/api/me", http.MethodGet, nil, tempTok.Token)
	s.Unauthorized()
}

func (s *MeHandlerSuite) TestGetMe_LazyHealsStaleOrDeactivatedParoki() {
	ctx := context.Background()
	keuskupan, err := s.client.Keuskupan.Create().
		SetKode("KAJ").
		SetNama("Keuskupan Agung Jakarta").
		Save(ctx)
	s.Require().NoError(err)

	p1, err := s.client.Paroki.Create().
		SetKeuskupanID(keuskupan.ID).
		SetKode("P1").
		SetNama("Paroki 1").
		SetActive(true).
		SetCreatedAt(time.Now().UTC().Add(-2 * time.Hour)).
		Save(ctx)
	s.Require().NoError(err)

	p2, err := s.client.Paroki.Create().
		SetKeuskupanID(keuskupan.ID).
		SetKode("P2").
		SetNama("Paroki 2").
		SetActive(true).
		SetCreatedAt(time.Now().UTC().Add(-1 * time.Hour)).
		Save(ctx)
	s.Require().NoError(err)

	u, err := s.feature.EnsureUser(ctx, "heal@paroki.org", "Heal User", "", "email", "heal@paroki.org")
	s.Require().NoError(err)
	for _, pid := range []uuid.UUID{p1.ID, p2.ID} {
		_, err = s.client.ParokiMember.Create().
			SetParokiID(pid).
			SetUserID(u.ID).
			SetRole("member").
			Save(ctx)
		s.Require().NoError(err)
	}

	// Login -> auto-selects p1
	_, cookie := s.loginUser("heal@paroki.org", "Heal User", "Browser/1")
	s.SetCookie(cookie)
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me1 := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Require().NotNil(me1.ActiveParokiID)
	s.Equal(p1.ID, *me1.ActiveParokiID)

	// Soft-delete p1 -> GET /api/me lazily heals session + user default to p2
	_, err = s.client.Paroki.UpdateOneID(p1.ID).SetDeletedAt(time.Now().UTC()).Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me2 := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Require().NotNil(me2.ActiveParokiID)
	s.Equal(p2.ID, *me2.ActiveParokiID)
	s.Require().NotNil(me2.DefaultParokiID)
	s.Equal(p2.ID, *me2.DefaultParokiID)

	uEnt, err := s.client.User.Get(ctx, u.ID)
	s.Require().NoError(err)
	s.Require().NotNil(uEnt.DefaultParokiID)
	s.Equal(p2.ID, *uEnt.DefaultParokiID)
}
