package handler_test

import (
	"context"
	"encoding/json"
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
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
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
	agentFeature *agent.Feature
	llmServer    *httptest.Server
	llmHandler   http.HandlerFunc
	llmMu        sync.RWMutex
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

	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	s.Require().NoError(err)

	s.agentFeature = agent.New(
		driver.NewRegistry(s.llmServer.Client()),
		logger,
		agent.WithDB(s.client),
		agent.WithUsers(s.userFeature),
		agent.WithCipher(cipher),
	)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.agentFeature))
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
