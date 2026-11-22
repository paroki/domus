package handler_test

import (
	"context"
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
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/org"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
	userdomain "github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type stubOTPSender struct {
	lastCode string
}

func (s *stubOTPSender) SendOTP(_ context.Context, _, _ string, code string) error {
	s.lastCode = code
	return nil
}

type AgentHandlerSuite struct {
	testutil.ApiTestSuite[any]
	client       *ent.Client
	enforcer     *authz.Enforcer
	otpSender    *stubOTPSender
	userFeature  *user.Feature
	refFeature   *ref.Feature
	orgFeature   *org.Feature
	agentFeature *agent.Feature
	settingStore *setting.Store
	llmServer    *httptest.Server
	llmHandler   http.HandlerFunc
	llmMu        sync.RWMutex
	parishID     uuid.UUID
}

func TestAgentHandlerSuite(t *testing.T) {
	suite.Run(t, new(AgentHandlerSuite))
}

func (s *AgentHandlerSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.otpSender = &stubOTPSender{}

	s.llmServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.llmMu.RLock()
		h := s.llmHandler
		s.llmMu.RUnlock()
		if h != nil {
			h(w, r)
			return
		}
		http.Error(w, "not configured", http.StatusInternalServerError)
	}))
	s.T().Cleanup(s.llmServer.Close)

	logger := slog.New(slog.DiscardHandler)
	userCfg := user.Config{
		SuperAdminEmail: "super@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     0,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.userFeature = user.New(s.client, s.enforcer, userCfg, logger, user.WithOTPSender(s.otpSender))
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
	setting.RegisterFeatures(s.settingStore, s.userFeature, s.refFeature, s.orgFeature, s.agentFeature)
	s.Require().NoError(s.settingStore.Load(ctx))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature, s.orgFeature, s.agentFeature)

	// Create a parish for tenant-scoped agent tests.
	d, err := s.client.Diocese.Create().SetCode("KAS").SetName("Keuskupan Agung").Save(ctx)
	s.Require().NoError(err)
	p, err := s.client.Parish.Create().SetCode("KTD").SetName("Paroki Katedral").SetDioceseID(d.ID).Save(ctx)
	s.Require().NoError(err)
	s.parishID = p.ID
}

func (s *AgentHandlerSuite) setLLMHandler(h http.HandlerFunc) {
	s.llmMu.Lock()
	defer s.llmMu.Unlock()
	s.llmHandler = h
}

func (s *AgentHandlerSuite) login(email string) *http.Cookie {
	if email != "super@paroki.org" {
		_, err := s.userFeature.EnsureUser(context.Background(), email, email, "", "email", email)
		s.Require().NoError(err)
	}
	s.Request("/api/auth/otp", http.MethodPost, userdomain.RequestOTPRequest{Email: email})
	s.OK()
	s.Request("/api/auth/otp/verify", http.MethodPost, userdomain.VerifyOTPRequest{
		Email: email,
		Code:  s.otpSender.lastCode,
	})
	s.OK()
	for _, c := range s.GetHttpResponse().Cookies() {
		if c.Name == "domus_session" {
			return c
		}
	}
	s.FailNow("session cookie not found")
	return nil
}

func (s *AgentHandlerSuite) joinParish(cookie *http.Cookie, email string) uuid.UUID {
	ctx := context.Background()
	u, err := s.client.User.Query().All(ctx)
	s.Require().NoError(err)
	var uid uuid.UUID
	for _, row := range u {
		if row.Email == email {
			uid = row.ID
			break
		}
	}
	s.Require().NotEqual(uuid.Nil, uid)
	_, err = s.client.ParishMember.Create().SetParishID(s.parishID).SetUserID(uid).Save(ctx)
	s.Require().NoError(err)
	s.ClearCookies()
	s.SetCookie(cookie)
	s.SetHeader("X-Parish-Id", s.parishID.String())
	return uid
}

func (s *AgentHandlerSuite) readBody() string {
	b, err := io.ReadAll(s.GetHttpResponse().Body)
	s.Require().NoError(err)
	return string(b)
}

func (s *AgentHandlerSuite) TestProvidersCRUD_RBAC_AndSyncModels() {
	memberCookie := s.login("member@paroki.org")
	s.SetCookie(memberCookie)
	s.Request("/api/agent/providers", http.MethodGet, nil)
	s.Forbidden()
	s.Request("/api/agent/providers/"+uuid.NewString()+"/sync-models", http.MethodPost, nil)
	s.Forbidden()

	superCookie := s.login("super@paroki.org")
	s.ClearCookies()
	s.SetCookie(superCookie)

	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"id": "gpt-4o", "name": "GPT-4o"},
				},
			})
			return
		}
		http.NotFound(w, r)
	})

	// Create first OpenAI config
	s.Request("/api/agent/providers", http.MethodPost, domain.CreateAgentProviderRequest{
		Name:        "OpenAI 1",
		Description: "Main OpenAI key",
		Provider:    "openai",
		APIKey:      "sk-test-1234",
		BaseURL:     s.llmServer.URL,
	})
	s.Created()
	created := testutil.DecodeAs[domain.AgentProviderResponse](&s.ApiTestSuite).Data
	s.Equal("OpenAI 1", created.Name)
	s.Equal("openai", created.Provider)
	s.True(created.HasAPIKey)
	s.NotContains(created.APIKeyMasked, "sk-test-1234")
	s.Require().Len(created.Models, 1)

	// Create second OpenAI config (multiple configs per provider allowed)
	s.Request("/api/agent/providers", http.MethodPost, domain.CreateAgentProviderRequest{
		Name:     "OpenAI 2",
		Provider: "openai",
		APIKey:   "sk-second-5678",
		BaseURL:  s.llmServer.URL,
	})
	s.Created()
	second := testutil.DecodeAs[domain.AgentProviderResponse](&s.ApiTestSuite).Data

	// List returns both
	s.Request("/api/agent/providers", http.MethodGet, nil)
	s.OK()
	listed := testutil.DecodeAs[[]domain.AgentProviderResponse](&s.ApiTestSuite).Data
	s.Len(listed, 2)

	// Patch with empty apiKey preserves old key
	newName := "OpenAI Primary"
	emptyKey := ""
	s.Request("/api/agent/providers/"+created.ID.String(), http.MethodPatch, domain.UpdateAgentProviderRequest{
		Name:   &newName,
		APIKey: &emptyKey,
	})
	s.OK()
	updated := testutil.DecodeAs[domain.AgentProviderResponse](&s.ApiTestSuite).Data
	s.Equal("OpenAI Primary", updated.Name)
	s.True(updated.HasAPIKey)

	// Sync models on UUID
	s.Request("/api/agent/providers/"+created.ID.String()+"/sync-models", http.MethodPost, nil)
	s.OK()
	res := testutil.DecodeAs[domain.SyncProviderModelsResponse](&s.ApiTestSuite).Data
	s.Equal(created.ID, res.ID)
	s.Equal("openai", res.Provider)
	s.Require().Len(res.Models, 1)
	s.Equal("gpt-4o", res.Models[0].ID)
	s.NotEmpty(s.agentFeature.Config().ProviderModels(created.ID))

	// Bind Gabriel to created.ID -> delete must fail with 409 Conflict
	_, err := s.settingStore.Update(context.Background(), map[string]string{
		agent.SettingGabrielProviderID: created.ID.String(),
	})
	s.Require().NoError(err)

	s.Request("/api/agent/providers/"+created.ID.String(), http.MethodDelete, nil)
	s.AssertStatus(http.StatusConflict)
	s.Equal("AGENT_PROVIDER_IN_USE", s.GetErrorResponse().Error.Code)

	// Patch with empty baseUrl clears it; invalid baseUrl returns 400
	emptyBaseURL := ""
	s.Request("/api/agent/providers/"+second.ID.String(), http.MethodPatch, domain.UpdateAgentProviderRequest{
		BaseURL: &emptyBaseURL,
	})
	s.OK()
	cleared := testutil.DecodeAs[domain.AgentProviderResponse](&s.ApiTestSuite).Data
	s.Equal("", cleared.BaseURL)

	badBaseURL := "not-a-url"
	s.Request("/api/agent/providers/"+second.ID.String(), http.MethodPatch, domain.UpdateAgentProviderRequest{
		BaseURL: &badBaseURL,
	})
	s.BadRequest()

	// Unused provider can be deleted
	s.Request("/api/agent/providers/"+second.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *AgentHandlerSuite) TestSessionsAndMessagesCRUD() {
	aliceCookie := s.login("alice@paroki.org")
	s.joinParish(aliceCookie, "alice@paroki.org")

	// Create session with default title.
	s.Request("/api/agent/sessions", http.MethodPost, nil)
	s.Created()
	sess1 := testutil.DecodeAs[domain.AgentSessionResponse](&s.ApiTestSuite).Data
	s.Equal("Percakapan Baru", sess1.Title)
	s.Equal("web", sess1.Channel)

	// Create second session with explicit title.
	s.Request("/api/agent/sessions", http.MethodPost, map[string]any{"title": "Jadwal Misa"})
	s.Created()
	sess2 := testutil.DecodeAs[domain.AgentSessionResponse](&s.ApiTestSuite).Data
	s.Equal("Jadwal Misa", sess2.Title)

	// Search sessions with limit=1 for cursor pagination.
	s.Request("/api/agent/sessions/search", http.MethodPost, map[string]any{"limit": 1})
	s.OK()
	env1 := testutil.DecodeAs[[]domain.AgentSessionResponse](&s.ApiTestSuite)
	s.Len(env1.Data, 1)
	s.Require().NotNil(env1.Meta)
	s.True(env1.Meta.HasMore)
	s.NotEmpty(env1.Meta.NextCursor)

	// Second page via cursor.
	s.Request("/api/agent/sessions/search", http.MethodPost, map[string]any{
		"cursor": env1.Meta.NextCursor,
		"limit":  1,
	})
	s.OK()
	list2 := testutil.DecodeAs[[]domain.AgentSessionResponse](&s.ApiTestSuite).Data
	s.Len(list2, 1)
	s.NotEqual(env1.Data[0].ID, list2[0].ID)

	// Get & Patch session.
	s.Request("/api/agent/sessions/"+sess1.ID.String(), http.MethodGet, nil)
	s.OK()
	s.Request("/api/agent/sessions/"+sess1.ID.String(), http.MethodPatch, map[string]any{"title": "Renamed"})
	s.OK()
	s.Equal("Renamed", testutil.DecodeAs[domain.AgentSessionResponse](&s.ApiTestSuite).Data.Title)

	// Cross-user isolation: bob cannot view/search alice's session.
	bobCookie := s.login("bob@paroki.org")
	s.joinParish(bobCookie, "bob@paroki.org")
	s.Request("/api/agent/sessions/"+sess1.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Request("/api/agent/sessions/"+sess1.ID.String()+"/messages/search", http.MethodPost, nil)
	s.NotFound()

	// Alice deletes session.
	s.ClearCookies()
	s.SetCookie(aliceCookie)
	s.Request("/api/agent/sessions/"+sess1.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/agent/sessions/"+sess1.ID.String(), http.MethodGet, nil)
	s.NotFound()
}

func (s *AgentHandlerSuite) TestMemoriesCRUD_PermissionsAndQuota() {
	memberCookie := s.login("memuser@paroki.org")
	s.joinParish(memberCookie, "memuser@paroki.org")

	// Member cannot create parish-scoped memory.
	s.Request("/api/agent/memories", http.MethodPost, map[string]any{
		"scope":    "parish",
		"category": "parish_faq",
		"content":  "Misa Minggu pukul 07:00",
	})
	s.Forbidden()
	s.Equal("AGENT_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Member creates personal memory (default scope = "user").
	s.Request("/api/agent/memories", http.MethodPost, map[string]any{
		"category": "preference",
		"content":  "Suka bahasa Indonesia santai",
		"isPinned": true,
	})
	s.Created()
	mem := testutil.DecodeAs[domain.AgentMemoryResponse](&s.ApiTestSuite).Data
	s.Equal("user", mem.Scope)
	s.True(mem.IsPinned)

	// Search memories.
	s.Request("/api/agent/memories/search", http.MethodPost, map[string]any{"q": "santai"})
	s.OK()
	found := testutil.DecodeAs[[]domain.AgentMemoryResponse](&s.ApiTestSuite).Data
	s.Len(found, 1)

	// Pinning 5 more hits 422 AGENT_MEMORY_LIMIT_EXCEEDED on the 6th.
	for i := 0; i < 4; i++ {
		s.Request("/api/agent/memories", http.MethodPost, map[string]any{
			"category": "profile",
			"content":  "Fact",
			"isPinned": true,
		})
		s.Created()
	}
	s.Request("/api/agent/memories", http.MethodPost, map[string]any{
		"category": "profile",
		"content":  "Sixth pinned",
		"isPinned": true,
	})
	s.UnprocessableEntity()
	s.Equal("AGENT_MEMORY_LIMIT_EXCEEDED", s.GetErrorResponse().Error.Code)

	// Super admin (has ref:parish-admin via bootstrap/role) can create/update/delete parish memory.
	superCookie := s.login("super@paroki.org")
	superID := s.joinParish(superCookie, "super@paroki.org")
	s.Require().NoError(s.enforcer.AssignRole(context.Background(), authx.Principal{UserID: superID}.Subject(), ref.RoleParishAdmin, authz.ParishDomain(s.parishID)))

	s.Request("/api/agent/memories", http.MethodPost, map[string]any{
		"scope":    "parish",
		"category": "parish_faq",
		"content":  "Misa harian pukul 06:00",
	})
	s.Created()
	parishMem := testutil.DecodeAs[domain.AgentMemoryResponse](&s.ApiTestSuite).Data
	s.Equal("parish", parishMem.Scope)

	s.Request("/api/agent/memories/"+parishMem.ID.String(), http.MethodPatch, map[string]any{
		"content": "Misa harian pukul 05:30",
	})
	s.OK()
	s.Request("/api/agent/memories/"+parishMem.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *AgentHandlerSuite) TestChatSSE_PreStreamErrorsAndStreaming() {
	aliceCookie := s.login("chatalice@paroki.org")
	s.joinParish(aliceCookie, "chatalice@paroki.org")

	// 1. Empty message -> 400 AGENT_EMPTY_MESSAGE.
	s.Request("/api/agent/chat", http.MethodPost, map[string]any{"message": "   "})
	s.BadRequest()
	s.Equal("AGENT_EMPTY_MESSAGE", s.GetErrorResponse().Error.Code)

	// 2. Non-existent sessionId -> 404 AGENT_SESSION_NOT_FOUND.
	s.Request("/api/agent/chat", http.MethodPost, map[string]any{
		"sessionId": uuid.New(),
		"message":   "Halo",
	})
	s.NotFound()
	s.Equal("AGENT_SESSION_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// 3. Unconfigured provider -> 503 AGENT_PROVIDER_UNAVAILABLE.
	s.Request("/api/agent/chat", http.MethodPost, map[string]any{"message": "Halo"})
	s.AssertStatus(http.StatusServiceUnavailable)
	s.Equal("AGENT_PROVIDER_UNAVAILABLE", s.GetErrorResponse().Error.Code)

	// Configure Gabriel on OpenAI mock server.
	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Halo \"}}]}\n\n")
			_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Kak!\"},\"finish_reason\":\"stop\"}]}\n\n")
			_, _ = io.WriteString(w, "data: [DONE]\n\n")
			return
		}
		http.NotFound(w, r)
	})
	prov, err := s.agentFeature.CreateProvider(context.Background(), domain.CreateAgentProviderRequest{
		Name:     "OpenAI Chat",
		Provider: "openai",
		APIKey:   "sk-chat",
		BaseURL:  s.llmServer.URL,
	})
	s.Require().NoError(err)
	_, err = s.settingStore.Update(context.Background(), map[string]string{
		agent.SettingGabrielProviderID:   prov.ID.String(),
		agent.SettingGabrielModel:        "gpt-4o",
		agent.SettingEmbeddingProviderID: "none",
	})
	s.Require().NoError(err)

	// Create session and stream SSE turn.
	s.Request("/api/agent/sessions", http.MethodPost, map[string]any{"title": "Web Chat"})
	s.Created()
	sess := testutil.DecodeAs[domain.AgentSessionResponse](&s.ApiTestSuite).Data

	s.Request("/api/agent/chat", http.MethodPost, map[string]any{
		"sessionId": sess.ID,
		"message":   "Halo Gabriel",
	})
	s.OK()
	s.Contains(s.GetHttpResponse().Header.Get("Content-Type"), "text/event-stream")
	body := s.readBody()
	s.Contains(body, "event: session\n")
	s.Contains(body, "event: delta\n")
	s.Contains(body, "Halo ")
	s.Contains(body, "Kak!")
	s.Contains(body, "event: done\n")

	// Verify messages persisted & searchable.
	s.Request("/api/agent/sessions/"+sess.ID.String()+"/messages/search", http.MethodPost, nil)
	s.OK()
	msgs := testutil.DecodeAs[[]domain.AgentMessageResponse](&s.ApiTestSuite).Data
	s.Len(msgs, 2)

	// 4. Busy session lock -> 409 AGENT_SESSION_BUSY.
	aliceID := sess.UserID
	prep, err := s.agentFeature.Service().PrepareTurn(context.Background(), service.ChatTurnInput{
		Principal: authx.Principal{UserID: aliceID, ParishID: s.parishID},
		Channel:   domain.ChannelWeb,
		SessionID: &sess.ID,
		Message:   "Lock holder",
	})
	s.Require().NoError(err)

	s.Request("/api/agent/chat", http.MethodPost, map[string]any{
		"sessionId": sess.ID,
		"message":   "Concurrent turn",
	})
	s.Conflict()
	s.Equal("AGENT_SESSION_BUSY", s.GetErrorResponse().Error.Code)
	prep.Release()

	// 5. Mid-stream upstream error emits named SSE error frame.
	s.setLLMHandler(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream failure", http.StatusBadGateway)
	})
	s.Request("/api/agent/chat", http.MethodPost, map[string]any{
		"sessionId": sess.ID,
		"message":   "Trigger stream error",
	})
	s.OK()
	s.Contains(s.readBody(), "event: error\n")
}
