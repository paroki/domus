package handler_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/paroki/domus/ent"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/useraccount"
	"github.com/paroki/domus/ent/userotp"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type TelegramHandlerSuite struct {
	testutil.ApiTestSuite[any]
	client    *ent.Client
	enforcer  *authz.Enforcer
	otpSender *stubOTPSender
	tgSrv     *httptest.Server
	tgCalls   []map[string]any
	tgClient  *telegram.Client
	feature   *user.Feature
	cfg       user.Config
}

func TestTelegramHandlerSuite(t *testing.T) {
	suite.Run(t, new(TelegramHandlerSuite))
}

func (s *TelegramHandlerSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.otpSender = &stubOTPSender{}
	s.tgCalls = nil

	s.tgSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		s.Require().NoError(json.NewDecoder(r.Body).Decode(&body))
		s.tgCalls = append(s.tgCalls, body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	s.T().Cleanup(s.tgSrv.Close)

	s.tgClient, err = telegram.New(telegram.Config{
		BotToken:      "123:TG_SECRET",
		BotUsername:   "domus_test_bot",
		WebhookSecret: "webhook-secret-123",
		BaseURL:       s.tgSrv.URL,
	})
	s.Require().NoError(err)

	s.cfg = user.Config{
		OTPTTL:         5 * time.Minute,
		OTPMaxAttempts: 5,
		OTPCooldown:    0,
		SessionTTL:     24 * time.Hour,
		CookieName:     "domus_session",
	}
	s.feature = user.New(
		s.client,
		s.enforcer,
		s.cfg,
		slog.New(slog.DiscardHandler),
		user.WithOTPSender(s.otpSender),
		user.WithTelegramClient(s.tgClient),
	)
	s.Require().NoError(s.feature.RegisterPermissions(context.Background(), s.enforcer))
	s.WithFeatures(s.feature)
}

func (s *TelegramHandlerSuite) loginUser(email, name string) (*domain.UserResponse, *http.Cookie) {
	u, err := s.feature.EnsureUser(context.Background(), email, name, "", "email", email)
	s.Require().NoError(err)

	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: email})
	s.OK()
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

func (s *TelegramHandlerSuite) extractStartToken(deepLinkURL string) string {
	u, err := url.Parse(deepLinkURL)
	s.Require().NoError(err)
	tok := u.Query().Get("start")
	s.Require().NotEmpty(tok)
	return tok
}

func (s *TelegramHandlerSuite) TestTelegramWebhook_SecretValidationAndMalformedJSON() {
	// Missing header -> 401 UNAUTHORIZED
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{UpdateID: 1})
	s.Unauthorized()

	// Wrong secret -> 401 UNAUTHORIZED
	s.SetHeader(telegram.SecretTokenHeader, "wrong-secret")
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{UpdateID: 1})
	s.Unauthorized()

	// Empty configured webhook secret (fail-closed) -> 401 UNAUTHORIZED
	noSecretClient, err := telegram.New(telegram.Config{
		BotToken:    "123:TG_SECRET",
		BotUsername: "domus_test_bot",
		BaseURL:     s.tgSrv.URL,
	})
	s.Require().NoError(err)
	noSecretFeature := user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler), user.WithTelegramClient(noSecretClient))
	s.WithFeatures(noSecretFeature)
	s.SetHeader(telegram.SecretTokenHeader, "")
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{UpdateID: 1})
	s.Unauthorized()

	// Valid secret + malformed JSON -> 400 BAD_REQUEST
	s.WithFeatures(s.feature)
	s.SetHeader(telegram.SecretTokenHeader, "webhook-secret-123")
	s.Request("/api/telegram/hooks", http.MethodPost, "{bad-json")
	s.BadRequest()
}

func (s *TelegramHandlerSuite) TestTelegramWebhook_LinkLifecycleAndConflicts() {
	ctx := context.Background()
	u1, cookie1 := s.loginUser("tg1@paroki.org", "TG User 1")
	s.ClearCookies()
	u2, cookie2 := s.loginUser("tg2@paroki.org", "TG User 2")

	// User 1 generates deep-link token.
	s.ClearCookies()
	s.SetCookie(cookie1)
	s.Request("/api/me/accounts/telegram/token", http.MethodPost, nil)
	s.OK()
	tok1 := s.extractStartToken(testutil.DecodeAs[domain.TelegramLinkTokenResponse](&s.ApiTestSuite).Data.DeepLinkURL)

	// Telegram sends /start <token> update with unknown extra Telegram fields -> 200 OK.
	s.ClearCookies()
	s.SetHeader(telegram.SecretTokenHeader, "webhook-secret-123")
	s.Request("/api/telegram/hooks", http.MethodPost, map[string]any{
		"update_id": 1001,
		"message": map[string]any{
			"message_id": 10,
			"date":       1700000000,
			"chat":       map[string]any{"id": 777001, "type": "private"},
			"from":       map[string]any{"id": 777001, "first_name": "Budi", "last_name": "Santoso", "username": "budis"},
			"text":       "/start " + tok1,
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 1)
	s.Equal(float64(777001), s.tgCalls[0]["chat_id"])
	s.Equal("✅ Telegram berhasil terhubung ke akun Domus Anda!", s.tgCalls[0]["text"])

	// Verify account persisted and exposed in /api/me.
	s.Headers.Del(telegram.SecretTokenHeader)
	s.SetCookie(cookie1)
	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me1 := testutil.DecodeAs[domain.UserResponse](&s.ApiTestSuite).Data
	s.Equal([]string{"email", "telegram"}, me1.LinkedProviders)

	tgAcc, err := s.client.UserAccount.Query().
		Where(useraccount.UserIDEQ(u1.ID), useraccount.ProviderEQ(domain.ProviderTelegram)).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal("777001", tgAcc.Subject)
	s.Require().NotNil(tgAcc.DisplayName)
	s.Equal("Budi Santoso", *tgAcc.DisplayName)

	// Reusing consumed token -> 200 OK with invalid/expired Indonesian reply.
	s.ClearCookies()
	s.SetHeader(telegram.SecretTokenHeader, "webhook-secret-123")
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 1002,
		Message: &telegram.Message{
			MessageID: 11,
			Chat:      telegram.Chat{ID: 777001, Type: "private"},
			Text:      "/start " + tok1,
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 2)
	s.Contains(s.tgCalls[1]["text"], "tidak valid atau sudah kedaluwarsa")

	// User 2 creates token, but User 1's chatID tries to claim it -> 200 OK + conflict warning.
	s.Headers.Del(telegram.SecretTokenHeader)
	s.SetCookie(cookie2)
	s.Request("/api/me/accounts/telegram/token", http.MethodPost, nil)
	s.OK()
	tok2 := s.extractStartToken(testutil.DecodeAs[domain.TelegramLinkTokenResponse](&s.ApiTestSuite).Data.DeepLinkURL)

	s.ClearCookies()
	s.SetHeader(telegram.SecretTokenHeader, "webhook-secret-123")
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 1003,
		Message: &telegram.Message{
			MessageID: 12,
			Chat:      telegram.Chat{ID: 777001, Type: "private"},
			Text:      "/start " + tok2,
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 3)
	s.Equal("⚠️ Akun Telegram ini sudah terhubung ke akun Domus lain.", s.tgCalls[2]["text"])

	// User 2 links with a different chatID -> succeeds.
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 1004,
		Message: &telegram.Message{
			MessageID: 13,
			Chat:      telegram.Chat{ID: 777002, Type: "private"},
			From:      &telegram.User{ID: 777002, Username: "user2tg"},
			Text:      "/start " + tok2,
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 4)
	s.Equal("✅ Telegram berhasil terhubung ke akun Domus Anda!", s.tgCalls[3]["text"])
	tgAcc2, err := s.client.UserAccount.Query().
		Where(useraccount.UserIDEQ(u2.ID), useraccount.ProviderEQ(domain.ProviderTelegram)).
		Only(ctx)
	s.Require().NoError(err)
	s.Require().NotNil(tgAcc2.DisplayName)
	s.Equal("@user2tg", *tgAcc2.DisplayName)

	// Expired token -> 200 OK + invalid/expired warning and OTP row deleted.
	expiredTok := "expired-tok-abc"
	sum := sha256.Sum256([]byte(expiredTok))
	_, err = s.client.UserOTP.Create().
		SetChannel(domain.ProviderTelegram).
		SetTarget(u1.ID.String()).
		SetCodeHash(hex.EncodeToString(sum[:])).
		SetExpiresAt(time.Now().UTC().Add(-1 * time.Minute)).
		Save(ctx)
	s.Require().NoError(err)

	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 1005,
		Message: &telegram.Message{
			MessageID: 14,
			Chat:      telegram.Chat{ID: 777003, Type: "private"},
			Text:      "/start " + expiredTok,
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 5)
	s.Contains(s.tgCalls[4]["text"], "tidak valid atau sudah kedaluwarsa")
	expCount, err := s.client.UserOTP.Query().Where(userotp.ChannelEQ(domain.ProviderTelegram), userotp.CodeHashEQ(hex.EncodeToString(sum[:]))).Count(ctx)
	s.Require().NoError(err)
	s.Equal(0, expCount)

	// Disabled user token -> 200 OK + disabled warning.
	s.Headers.Del(telegram.SecretTokenHeader)
	s.SetCookie(cookie1)
	s.Request("/api/me/accounts/telegram/token", http.MethodPost, nil)
	s.OK()
	tokDisabled := s.extractStartToken(testutil.DecodeAs[domain.TelegramLinkTokenResponse](&s.ApiTestSuite).Data.DeepLinkURL)

	_, err = s.client.User.UpdateOneID(u1.ID).SetStatus(entuser.StatusDisabled).Save(ctx)
	s.Require().NoError(err)

	s.ClearCookies()
	s.SetHeader(telegram.SecretTokenHeader, "webhook-secret-123")
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 1006,
		Message: &telegram.Message{
			MessageID: 15,
			Chat:      telegram.Chat{ID: 777001, Type: "private"},
			Text:      "/start " + tokDisabled,
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 6)
	s.Equal("⚠️ Akun Domus Anda sedang dinonaktifkan.", s.tgCalls[5]["text"])
}

func (s *TelegramHandlerSuite) TestTelegramWebhook_DefaultAndCustomFallback() {
	s.SetHeader(telegram.SecretTokenHeader, "webhook-secret-123")

	// Default non-/start fallback sends Indonesian guidance to /me.
	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 2001,
		Message: &telegram.Message{
			MessageID: 21,
			Chat:      telegram.Chat{ID: 888001, Type: "private"},
			Text:      "halo bot",
		},
	})
	s.OK()
	s.Require().Len(s.tgCalls, 1)
	s.Contains(s.tgCalls[0]["text"], "/me")

	// Custom fallback handler receives update and webhook still returns 200 even if fallback errors.
	var gotText string
	s.feature.SetTelegramMessageHandler(func(_ context.Context, upd telegram.Update) error {
		if upd.Message != nil {
			gotText = upd.Message.Text
		}
		return errors.New("simulated downstream error")
	})

	s.Request("/api/telegram/hooks", http.MethodPost, telegram.Update{
		UpdateID: 2002,
		Message: &telegram.Message{
			MessageID: 22,
			Chat:      telegram.Chat{ID: 888001, Type: "private"},
			Text:      "daftar acara",
		},
	})
	s.OK()
	s.Equal("daftar acara", gotText)
	s.Len(s.tgCalls, 1)
}
