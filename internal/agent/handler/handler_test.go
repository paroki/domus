package handler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
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
	sqlDB        *sql.DB
	enforcer     *authz.Enforcer
	otpSender    *stubOTPSender
	userFeature  *user.Feature
	agentFeature *agent.Feature
	settings     *setting.Store
	llmServer    *httptest.Server
	llmHandler   http.HandlerFunc
	llmMu        sync.RWMutex
}

func TestAgentHandlerSuite(t *testing.T) {
	suite.Run(t, new(AgentHandlerSuite))
}

func (s *AgentHandlerSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client, s.sqlDB, _ = testutil.NewEntClientWithDB(s.T())
	s.Require().NoError(repo.EnsureMemorySearchSchema(context.Background(), s.sqlDB))
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

	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	s.Require().NoError(err)

	s.settings = setting.NewStore(s.client, setting.WithCipher(cipher))
	s.agentFeature = agent.New(
		driver.NewRegistry(s.llmServer.Client()),
		logger,
		agent.WithDB(s.client),
		agent.WithSQLDB(s.sqlDB),
		agent.WithUsers(s.userFeature),
		agent.WithCipher(cipher),
		agent.WithSettings(s.settings),
	)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.agentFeature))
	setting.RegisterFeatures(s.settings, s.agentFeature)
	s.Require().NoError(s.settings.Load(ctx))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.agentFeature)
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

	// Delete provider succeeds
	s.Request("/api/agent/providers/"+second.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}

func (s *AgentHandlerSuite) TestSessionsAndMemoriesHTTP() {
	ctx := context.Background()
	keuskupan, err := s.client.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	s.Require().NoError(err)
	paroki1, err := s.client.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p1").SetNama("Paroki 1").Save(ctx)
	s.Require().NoError(err)
	paroki2, err := s.client.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p2").SetNama("Paroki 2").Save(ctx)
	s.Require().NoError(err)

	superCookie := s.login("super@paroki.org")
	s.SetCookie(superCookie)

	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/models") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": "gpt-4o", "name": "GPT-4o"}},
			})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/embeddings") {
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "fallback-trigger") {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			vec := make([]float64, driver.EmbeddingDimensions)
			vec[0] = 0.5
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list",
				"data":   []map[string]any{{"object": "embedding", "index": 0, "embedding": vec}},
			})
			return
		}
		http.NotFound(w, r)
	})

	s.Request("/api/agent/providers", http.MethodPost, domain.CreateAgentProviderRequest{
		Name:     "OpenAI",
		Provider: "openai",
		APIKey:   "sk-123",
		BaseURL:  s.llmServer.URL,
	})
	s.Created()
	prov := testutil.DecodeAs[domain.AgentProviderResponse](&s.ApiTestSuite).Data

	s.Request("/api/agent/providers", http.MethodPost, domain.CreateAgentProviderRequest{
		Name:     "OpenAI Extra",
		Provider: "openai",
		APIKey:   "sk-456",
		BaseURL:  s.llmServer.URL,
	})
	s.Created()
	extraProv := testutil.DecodeAs[domain.AgentProviderResponse](&s.ApiTestSuite).Data

	_, err = s.settings.Update(ctx, map[string]string{
		agent.SettingModelGabriel:   prov.ID.String() + ":gpt-4o",
		agent.SettingModelEmbedding: prov.ID.String() + ":text-embedding-3-small",
	})
	s.Require().NoError(err)

	memberCookie := s.login("member@paroki.org")
	otherCookie := s.login("other@paroki.org")

	memberUser, err := s.userFeature.EnsureUser(ctx, "member@paroki.org", "Member", "", "email", "member@paroki.org")
	s.Require().NoError(err)
	otherUser, err := s.userFeature.EnsureUser(ctx, "other@paroki.org", "Other", "", "email", "other@paroki.org")
	s.Require().NoError(err)

	_, err = s.client.ParokiMember.Create().SetParokiID(paroki1.ID).SetUserID(memberUser.ID).Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.ParokiMember.Create().SetParokiID(paroki2.ID).SetUserID(memberUser.ID).Save(ctx)
	s.Require().NoError(err)
	_, err = s.client.ParokiMember.Create().SetParokiID(paroki1.ID).SetUserID(otherUser.ID).Save(ctx)
	s.Require().NoError(err)

	// Without X-Paroki-Id -> 400 TENANT_REQUIRED
	s.ClearCookies()
	s.SetCookie(memberCookie)
	s.Request("/api/agent/sessions/search", http.MethodPost, domain.SearchSessionsRequest{})
	s.BadRequest()
	s.Equal("TENANT_REQUIRED", s.GetErrorResponse().Error.Code)

	// Seed session directly for member in paroki1
	r := repo.New(s.client, s.sqlDB)
	sess, err := r.CreateSession(ctx, paroki1.ID, memberUser.ID, "Sesi Awal", nil, nil)
	s.Require().NoError(err)
	_, err = r.AppendMessages(ctx, paroki1.ID, memberUser.ID, sess.ID, []domain.AppendMessageInput{
		{Role: "user", Content: "Halo"},
	})
	s.Require().NoError(err)

	// Cross-user isolation -> 404
	s.ClearCookies()
	s.SetCookie(otherCookie)
	s.SetHeader("X-Paroki-Id", paroki1.ID.String())
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodGet, nil)
	s.NotFound()

	// Cross-parish isolation -> 404
	s.ClearCookies()
	s.SetCookie(memberCookie)
	s.SetHeader("X-Paroki-Id", paroki2.ID.String())
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodGet, nil)
	s.NotFound()

	// Owner in paroki1 -> 200
	s.SetHeader("X-Paroki-Id", paroki1.ID.String())
	s.Request("/api/agent/sessions/search", http.MethodPost, domain.SearchSessionsRequest{Q: "awal"})
	s.OK()
	sessList := testutil.DecodeAs[[]domain.SessionResponse](&s.ApiTestSuite).Data
	s.Len(sessList, 1)

	// Set override on session -> then clear with "" -> then verify deleted provider fallback
	newTitle := "Sesi Baru"
	pidStr := extraProv.ID.String()
	midStr := "gpt-4o-mini"
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodPatch, domain.UpdateSessionRequest{
		Title:      &newTitle,
		ProviderID: &pidStr,
		ModelID:    &midStr,
	})
	s.OK()
	updSess := testutil.DecodeAs[domain.SessionResponse](&s.ApiTestSuite).Data
	s.Equal("Sesi Baru", updSess.Title)
	s.Require().NotNil(updSess.ProviderID)

	emptyStr := ""
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodPatch, domain.UpdateSessionRequest{
		ProviderID: &emptyStr,
	})
	s.OK()
	clearedSess := testutil.DecodeAs[domain.SessionResponse](&s.ApiTestSuite).Data
	s.Nil(clearedSess.ProviderID)
	s.Nil(clearedSess.ModelID)

	// Re-bind to extraProv, delete extraProv, GET session falls back cleanly to Gabriel
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodPatch, domain.UpdateSessionRequest{
		ProviderID: &pidStr,
		ModelID:    &midStr,
	})
	s.OK()
	s.Require().NoError(s.client.AgentProvider.DeleteOneID(extraProv.ID).Exec(ctx))
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodGet, nil)
	s.OK()
	detail := testutil.DecodeAs[domain.SessionDetailResponse](&s.ApiTestSuite).Data
	s.Require().NotNil(detail.EffectiveProviderID)
	s.Equal(prov.ID, *detail.EffectiveProviderID)
	s.Equal("gpt-4o", detail.EffectiveModelID)
	s.Len(detail.Messages, 1)

	// Delete session -> 204
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodDelete, nil)
	s.NoContent()
	s.Request("/api/agent/sessions/"+sess.ID.String(), http.MethodGet, nil)
	s.NotFound()

	// Memories: > 2000 chars rejected with 422
	s.Request("/api/agent/memories", http.MethodPost, domain.CreateMemoryRequest{
		Content: strings.Repeat("a", 2001),
	})
	s.UnprocessableEntity()

	// Valid memory creation + hybrid search + BM25 fallback
	s.Request("/api/agent/memories", http.MethodPost, domain.CreateMemoryRequest{
		Content:  "Jadwal misa pagi jam 6",
		Category: "fact",
	})
	s.Created()
	mem := testutil.DecodeAs[domain.MemoryResponse](&s.ApiTestSuite).Data

	s.Request("/api/agent/memories/search", http.MethodPost, domain.SearchMemoriesRequest{
		Q: "jadwal misa",
	})
	s.OK()
	memHits := testutil.DecodeAs[[]domain.MemoryResponse](&s.ApiTestSuite).Data
	s.NotEmpty(memHits)

	// Fallback to BM25 when embedding returns 503
	s.Request("/api/agent/memories/search", http.MethodPost, domain.SearchMemoriesRequest{
		Q: "misa fallback-trigger",
	})
	s.OK()
	bm25Hits := testutil.DecodeAs[[]domain.MemoryResponse](&s.ApiTestSuite).Data
	s.NotEmpty(bm25Hits)

	// Patch & Delete memory
	updContent := "Jadwal misa pagi pukul 06:00"
	s.Request("/api/agent/memories/"+mem.ID.String(), http.MethodPatch, domain.UpdateMemoryRequest{
		Content: &updContent,
	})
	s.OK()
	s.Request("/api/agent/memories/"+mem.ID.String(), http.MethodDelete, nil)
	s.NoContent()
}
