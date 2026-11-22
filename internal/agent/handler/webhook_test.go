package handler_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type outboundCall struct {
	path    string
	payload map[string]any
}

type AgentWebhookSuite struct {
	testutil.ApiTestSuite[any]
	client       *ent.Client
	enforcer     *authz.Enforcer
	userFeature  *user.Feature
	refFeature   *ref.Feature
	orgFeature   *org.Feature
	agentFeature *agent.Feature
	settingStore *setting.Store
	tgServer     *httptest.Server
	gowaServer   *httptest.Server
	llmServer    *httptest.Server
	tgCalls      []outboundCall
	gowaCalls    []outboundCall
	callsMu      sync.Mutex
	parishID     uuid.UUID
	userID       uuid.UUID
	tgSecret     string
	gowaSecret   string
}

func TestAgentWebhookSuite(t *testing.T) {
	suite.Run(t, new(AgentWebhookSuite))
}

func (s *AgentWebhookSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.tgSecret = "tg-webhook-secret"
	s.gowaSecret = "gowa-webhook-secret"

	s.tgServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		s.callsMu.Lock()
		s.tgCalls = append(s.tgCalls, outboundCall{path: r.URL.Path, payload: body})
		s.callsMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	s.T().Cleanup(s.tgServer.Close)

	s.gowaServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		s.callsMu.Lock()
		s.gowaCalls = append(s.gowaCalls, outboundCall{path: r.URL.Path, payload: body})
		s.callsMu.Unlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"SUCCESS"}`))
	}))
	s.T().Cleanup(s.gowaServer.Close)

	s.llmServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Siap Kak!\"}}]}\n\n")
			_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"stop\"}]}\n\n")
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
			return
		}
		http.NotFound(w, r)
	}))
	s.T().Cleanup(s.llmServer.Close)

	tgClient, err := telegram.New(telegram.Config{
		BotToken:      "123:ABC",
		BotUsername:   "domus_bot",
		WebhookSecret: s.tgSecret,
		BaseURL:       s.tgServer.URL,
	})
	s.Require().NoError(err)

	gowaClient, err := gowa.New(gowa.Config{
		BaseURL:       s.gowaServer.URL,
		DeviceID:      "dev-1",
		WebhookSecret: s.gowaSecret,
	})
	s.Require().NoError(err)

	logger := slog.New(slog.DiscardHandler)
	userCfg := user.Config{
		SuperAdminEmail: "super@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.userFeature = user.New(
		s.client,
		s.enforcer,
		userCfg,
		logger,
		user.WithTelegramClient(tgClient),
		user.WithGowaClient(gowaClient),
	)
	s.refFeature = ref.New(s.client, s.enforcer, s.userFeature)
	s.orgFeature = org.New(s.client, s.enforcer, s.userFeature, s.refFeature, org.Config{})
	s.userFeature.SetInviteJoiner(s.orgFeature)

	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	s.Require().NoError(err)
	s.settingStore = setting.NewStore(s.client, setting.WithCipher(cipher))

	s.agentFeature = agent.New(
		driver.NewRegistry(s.llmServer.Client()),
		logger,
		agent.WithService(s.client, nil, dialect.Postgres, s.enforcer, nil, s.userFeature, s.refFeature, s.orgFeature),
		agent.WithSettings(s.settingStore),
		agent.WithCipher(cipher),
	)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature, s.orgFeature, s.agentFeature))
	setting.RegisterFeatures(s.settingStore, s.agentFeature)
	s.Require().NoError(s.settingStore.Load(ctx))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	prov, err := s.agentFeature.CreateProvider(ctx, domain.CreateAgentProviderRequest{
		Name:     "OpenAI Webhook",
		Provider: "openai",
		APIKey:   "sk-webhook",
		BaseURL:  s.llmServer.URL,
	})
	s.Require().NoError(err)
	_, err = s.settingStore.Update(ctx, map[string]string{
		agent.SettingGabrielProviderID:   prov.ID.String(),
		agent.SettingGabrielModel:        "gpt-4o",
		agent.SettingEmbeddingProviderID: "none",
	})
	s.Require().NoError(err)

	s.WithFeatures(s.userFeature, s.refFeature, s.orgFeature, s.agentFeature)

	d, err := s.client.Diocese.Create().SetCode("KAS").SetName("Keuskupan Agung").Save(ctx)
	s.Require().NoError(err)
	p, err := s.client.Parish.Create().SetCode("ktd").SetName("Paroki Katedral").SetDioceseID(d.ID).Save(ctx)
	s.Require().NoError(err)
	s.parishID = p.ID

	u, err := s.userFeature.EnsureUser(ctx, "budi@paroki.org", "Budi", "", "telegram", "111222")
	s.Require().NoError(err)
	s.userID = u.ID
	_, err = s.client.UserAccount.Create().SetUserID(u.ID).SetProvider("whatsapp").SetSubject("+628123456789").Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.ParishMember.Create().SetParishID(s.parishID).SetUserID(u.ID).Save(ctx)
	s.Require().NoError(err)
}

// waitCalls waits until async channel turns finished sending (call count stable).
func (s *AgentWebhookSuite) waitCalls() {
	count := func() int {
		s.callsMu.Lock()
		defer s.callsMu.Unlock()
		return len(s.tgCalls) + len(s.gowaCalls)
	}
	deadline := time.Now().Add(3 * time.Second)
	last, stableSince := -1, time.Now()
	for time.Now().Before(deadline) {
		if n := count(); n != last {
			last, stableSince = n, time.Now()
		} else if n > 0 && time.Since(stableSince) > 200*time.Millisecond {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (s *AgentWebhookSuite) resetCalls() {
	s.callsMu.Lock()
	defer s.callsMu.Unlock()
	s.tgCalls = nil
	s.gowaCalls = nil
}

func (s *AgentWebhookSuite) postTelegram(payload any, secret string) {
	s.SetHeader(telegram.SecretTokenHeader, secret)
	s.Request("/api/telegram/hooks", http.MethodPost, payload)
}

func (s *AgentWebhookSuite) sendWhatsApp(payload any, secret string) {
	raw, err := json.Marshal(payload)
	s.Require().NoError(err)
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(raw)
		s.SetHeader(gowa.SignatureHeader, "sha256="+hex.EncodeToString(mac.Sum(nil)))
	} else {
		s.SetHeader(gowa.SignatureHeader, "sha256=invalid")
	}
	s.Request("/api/agent/webhooks/whatsapp", http.MethodPost, json.RawMessage(raw))
}

func (s *AgentWebhookSuite) TestTelegramWebhook_GroupMentionMediaCommandsAndTurn() {
	// 1. Group message without mention -> ignored (no outbound).
	s.resetCalls()
	s.postTelegram(map[string]any{
		"update_id": 1,
		"message": map[string]any{
			"message_id": 10,
			"from":       map[string]any{"id": 111222},
			"chat":       map[string]any{"id": -100999, "type": "supergroup"},
			"text":       "halo semua",
		},
	}, s.tgSecret)
	s.OK()
	s.Empty(s.tgCalls)

	// 2. Group media-only with mention -> text-only guard.
	s.resetCalls()
	s.postTelegram(map[string]any{
		"update_id": 2,
		"message": map[string]any{
			"message_id": 11,
			"from":       map[string]any{"id": 111222},
			"chat":       map[string]any{"id": -100999, "type": "supergroup"},
			"caption":    "@domus_bot",
			"photo":      []map[string]any{{"file_id": "p1"}},
		},
	}, s.tgSecret)
	s.OK()
	s.waitCalls()
	s.Require().Len(s.tgCalls, 1)
	s.Contains(s.tgCalls[0].payload["text"], "hanya dapat membaca pesan teks")

	// 3. Unlinked sender -> onboarding instruction.
	s.resetCalls()
	s.postTelegram(map[string]any{
		"update_id": 3,
		"message": map[string]any{
			"message_id": 12,
			"from":       map[string]any{"id": 999888},
			"chat":       map[string]any{"id": 999888, "type": "private"},
			"text":       "halo",
		},
	}, s.tgSecret)
	s.OK()
	s.waitCalls()
	s.Require().Len(s.tgCalls, 1)
	s.Contains(s.tgCalls[0].payload["text"], "/me")

	// 4. Linked user without active parish -> /parish ktd switches parish, then DM runs turn.
	s.resetCalls()
	s.postTelegram(map[string]any{
		"update_id": 4,
		"message": map[string]any{
			"message_id": 13,
			"from":       map[string]any{"id": 111222},
			"chat":       map[string]any{"id": 111222, "type": "private"},
			"text":       "/parish ktd",
		},
	}, s.tgSecret)
	s.OK()
	s.waitCalls()
	s.Require().Len(s.tgCalls, 1)
	s.Contains(s.tgCalls[0].payload["text"], "Paroki Katedral")

	// 5. /new creates fresh session.
	s.resetCalls()
	s.postTelegram(map[string]any{
		"update_id": 5,
		"message": map[string]any{
			"message_id": 14,
			"from":       map[string]any{"id": 111222},
			"chat":       map[string]any{"id": -100999, "type": "group"},
			"text":       "/new@domus_bot",
		},
	}, s.tgSecret)
	s.OK()
	s.waitCalls()
	s.Require().Len(s.tgCalls, 1)
	s.Contains(s.tgCalls[0].payload["text"], "Sesi percakapan baru")

	// 6. Normal DM turn -> sends typing + final reply.
	s.resetCalls()
	s.postTelegram(map[string]any{
		"update_id": 6,
		"message": map[string]any{
			"message_id": 15,
			"from":       map[string]any{"id": 111222},
			"chat":       map[string]any{"id": 111222, "type": "private"},
			"text":       "Halo Gabriel",
		},
	}, s.tgSecret)
	s.OK()
	s.waitCalls()
	s.Require().GreaterOrEqual(len(s.tgCalls), 2)
	s.Contains(s.tgCalls[0].path, "sendChatAction")
	s.Equal("Siap Kak!", s.tgCalls[len(s.tgCalls)-1].payload["text"])
}

func (s *AgentWebhookSuite) TestWhatsAppWebhook_SignatureGroupSelfAndTurn() {
	// 1. Invalid signature -> 401 AGENT_WEBHOOK_UNAUTHORIZED.
	s.sendWhatsApp(map[string]any{"event": "message"}, "")
	s.Unauthorized()
	s.Equal("AGENT_WEBHOOK_UNAUTHORIZED", s.GetErrorResponse().Error.Code)

	// 2. from_me -> 200 OK ignored.
	s.resetCalls()
	s.sendWhatsApp(map[string]any{
		"event":   "message",
		"from_me": true,
		"from":    "628123456789@s.whatsapp.net",
		"text":    "echo",
	}, s.gowaSecret)
	s.OK()
	s.Empty(s.gowaCalls)

	// 3. Group without mention -> ignored; group with @gabriel -> runs turn & replies to @g.us.
	s.resetCalls()
	s.sendWhatsApp(map[string]any{
		"event":     "message",
		"chat_id":   "120363000111222@g.us",
		"sender_id": "628123456789@s.whatsapp.net",
		"text":      "diskusi biasa",
	}, s.gowaSecret)
	s.OK()
	s.Empty(s.gowaCalls)

	// Set active parish via /parish ktd first.
	s.resetCalls()
	s.sendWhatsApp(map[string]any{
		"event":   "message",
		"from":    "628123456789@s.whatsapp.net",
		"chat_id": "628123456789@s.whatsapp.net",
		"text":    "/parish ktd",
	}, s.gowaSecret)
	s.OK()
	s.waitCalls()
	s.Require().Len(s.gowaCalls, 1)
	s.Contains(s.gowaCalls[0].payload["message"], "Paroki Katedral")

	// Group @gabriel turn -> replies to group JID.
	s.resetCalls()
	s.sendWhatsApp(map[string]any{
		"event": "message",
		"payload": map[string]any{
			"chat_id":   "120363000111222@g.us",
			"sender_id": "628123456789@s.whatsapp.net",
			"message":   map[string]any{"text": "@gabriel halo"},
		},
	}, s.gowaSecret)
	s.OK()
	s.waitCalls()
	s.Require().GreaterOrEqual(len(s.gowaCalls), 2)
	s.Equal("/send/chat-presence", s.gowaCalls[0].path)
	s.Equal("120363000111222@g.us", s.gowaCalls[0].payload["phone"])
	s.Equal("/send/message", s.gowaCalls[len(s.gowaCalls)-1].path)
	s.Equal("120363000111222@g.us", s.gowaCalls[len(s.gowaCalls)-1].payload["phone"])
	s.Equal("Siap Kak!", s.gowaCalls[len(s.gowaCalls)-1].payload["message"])
}
