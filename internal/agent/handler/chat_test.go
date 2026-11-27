package handler_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/app"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
	userdomain "github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ChatSuite struct {
	suite.Suite
	client       *ent.Client
	sqlDB        interface{ Close() error }
	enforcer     *authz.Enforcer
	otpSender    *stubOTPSender
	userFeature  *user.Feature
	refFeature   *ref.Feature
	agentFeature *agent.Feature
	settings     *setting.Store
	fiberApp     *fiber.App
	llmServer    *httptest.Server
	llmHandler   http.HandlerFunc
	llmMu        sync.RWMutex
	parokiID     uuid.UUID
	memberCookie *http.Cookie
	superCookie  *http.Cookie
}

func TestChatSuite(t *testing.T) {
	suite.Run(t, new(ChatSuite))
}

func (s *ChatSuite) SetupTest() {
	ctx := context.Background()
	db, sqlDB, _ := testutil.NewEntClientWithDB(s.T())
	s.client = db
	s.sqlDB = sqlDB
	s.Require().NoError(repo.EnsureMemorySearchSchema(ctx, sqlDB))

	enforcer, err := authz.NewEnforcer(db)
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
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.userFeature = user.New(db, enforcer, userCfg, logger, user.WithOTPSender(s.otpSender))
	s.refFeature = ref.New(db, enforcer, s.userFeature)
	s.userFeature.SetParokiMemberDirectory(s.refFeature)

	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	s.Require().NoError(err)
	s.settings = setting.NewStore(db, setting.WithCipher(cipher))

	s.agentFeature = agent.New(
		driver.NewRegistry(s.llmServer.Client()),
		logger,
		agent.WithDB(db),
		agent.WithSQLDB(sqlDB),
		agent.WithUsers(s.userFeature),
		agent.WithCipher(cipher),
		agent.WithSettings(s.settings),
	)

	s.Require().NoError(authz.RegisterFeatures(ctx, enforcer, s.userFeature, s.refFeature, s.agentFeature))
	setting.RegisterFeatures(s.settings, s.agentFeature)
	s.Require().NoError(s.settings.Load(ctx))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.fiberApp = app.GetFiber(app.Config{}, logger, s.userFeature, s.refFeature, s.agentFeature)

	keuskupan, err := db.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	s.Require().NoError(err)
	paroki, err := db.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p1").SetNama("Paroki Katedral").Save(ctx)
	s.Require().NoError(err)
	s.parokiID = paroki.ID

	memberUser, err := s.userFeature.EnsureUser(ctx, "member@paroki.org", "Member", "", "email", "member@paroki.org")
	s.Require().NoError(err)
	_, err = db.ParokiMember.Create().SetParokiID(paroki.ID).SetUserID(memberUser.ID).Save(ctx)
	s.Require().NoError(err)

	s.superCookie = s.login("super@paroki.org")
	s.memberCookie = s.login("member@paroki.org")
}

func (s *ChatSuite) setLLMHandler(h http.HandlerFunc) {
	s.llmMu.Lock()
	defer s.llmMu.Unlock()
	s.llmHandler = h
}

func (s *ChatSuite) login(email string) *http.Cookie {
	b1, _ := json.Marshal(userdomain.RequestOTPRequest{Email: email})
	r1 := httptest.NewRequest(http.MethodPost, "/api/auth/otp", bytes.NewReader(b1))
	r1.Header.Set("Content-Type", "application/json")
	resp1, err := s.fiberApp.Test(r1)
	s.Require().NoError(err)
	_ = resp1.Body.Close()

	b2, _ := json.Marshal(userdomain.VerifyOTPRequest{Email: email, Code: s.otpSender.lastCode})
	r2 := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewReader(b2))
	r2.Header.Set("Content-Type", "application/json")
	resp2, err := s.fiberApp.Test(r2)
	s.Require().NoError(err)
	defer func() { _ = resp2.Body.Close() }()
	for _, c := range resp2.Cookies() {
		if c.Name == "domus_session" {
			return c
		}
	}
	s.FailNow("session cookie not found")
	return nil
}

func (s *ChatSuite) TestPreStreamValidationErrors() {
	// 1. Unauthenticated -> 401
	req := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.fiberApp.Test(req)
	s.Require().NoError(err)
	_ = resp.Body.Close()
	s.Equal(http.StatusUnauthorized, resp.StatusCode)

	// 2. Missing X-Paroki-Id -> 400 TENANT_REQUIRED
	req = httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"hi"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(s.memberCookie)
	resp, err = s.fiberApp.Test(req)
	s.Require().NoError(err)
	_ = resp.Body.Close()
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	// 3. Empty or > 16000 runes -> 422 VALIDATION_ERROR
	for _, msg := range []string{"   ", strings.Repeat("a", 16001)} {
		body, _ := json.Marshal(domain.ChatRequest{Message: msg})
		req = httptest.NewRequest(http.MethodPost, "/api/agent/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(s.memberCookie)
		req.Header.Set("X-Paroki-Id", s.parokiID.String())
		resp, err = s.fiberApp.Test(req)
		s.Require().NoError(err)
		_ = resp.Body.Close()
		s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	}

	// 4. Unconfigured agent.model.gabriel -> 400 AGENT_PROVIDER_UNAVAILABLE
	req = httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"halo"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(s.memberCookie)
	req.Header.Set("X-Paroki-Id", s.parokiID.String())
	resp, err = s.fiberApp.Test(req)
	s.Require().NoError(err)
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	s.Equal(http.StatusBadRequest, resp.StatusCode)
	s.Contains(string(raw), "AGENT_PROVIDER_UNAVAILABLE")

	// 5. Unknown sessionId -> 404 AGENT_SESSION_NOT_FOUND
	body, _ := json.Marshal(domain.ChatRequest{
		SessionID: func() *uuid.UUID { id := uuid.New(); return &id }(),
		Message:   "halo",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/agent/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(s.memberCookie)
	req.Header.Set("X-Paroki-Id", s.parokiID.String())
	resp, err = s.fiberApp.Test(req)
	s.Require().NoError(err)
	_ = resp.Body.Close()
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *ChatSuite) TestStreamingDelegationBusyLockAndCancel() {
	ctx := context.Background()
	var callNum atomic.Int32
	slowGate := make(chan struct{})

	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": "gpt-4o", "name": "GPT-4o"}},
			})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			reqBody, _ := io.ReadAll(r.Body)
			// Background memory extractor call
			if strings.Contains(string(reqBody), "Memory Extractor") || strings.Contains(string(reqBody), "Return ONLY a JSON array") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id": "ext-1", "object": "chat.completion", "created": 1, "model": "gpt-4o",
					"choices": []map[string]any{
						{"index": 0, "message": map[string]any{"role": "assistant", "content": "[]"}, "finish_reason": "stop"},
					},
				})
				return
			}

			n := callNum.Add(1)
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)

			switch n {
			case 1:
				// Gabriel delegates to rafael
				writeOpenAIToolChunk(w, "call_raf", "rafael", `{"request":"cek nama paroki aktif"}`)
				if flusher != nil {
					flusher.Flush()
				}
			case 2:
				// Rafael calls ref_get_paroki
				writeOpenAIToolChunk(w, "call_ref", "ref_get_paroki", `{}`)
				if flusher != nil {
					flusher.Flush()
				}
			case 3:
				// Rafael finishes
				writeOpenAITextChunk(w, "Paroki aktif adalah Paroki Katedral.")
				if flusher != nil {
					flusher.Flush()
				}
			case 4:
				// Gabriel final answer
				writeOpenAITextChunk(w, "Paroki aktif Anda adalah Paroki Katedral.")
				if flusher != nil {
					flusher.Flush()
				}
			default:
				// Busy/cancel test turn: wait on slowGate before streaming
				<-slowGate
				writeOpenAITextChunk(w, "Selesai.")
				if flusher != nil {
					flusher.Flush()
				}
			}
			return
		}
		http.NotFound(w, r)
	})

	// Create provider + set agent.model.gabriel and prompt with {curly_braces}
	provBody, _ := json.Marshal(domain.CreateAgentProviderRequest{
		Name:     "OpenAI",
		Provider: "openai",
		APIKey:   "sk-test",
		BaseURL:  s.llmServer.URL,
	})
	pReq := httptest.NewRequest(http.MethodPost, "/api/agent/providers", bytes.NewReader(provBody))
	pReq.Header.Set("Content-Type", "application/json")
	pReq.AddCookie(s.superCookie)
	pResp, err := s.fiberApp.Test(pReq)
	s.Require().NoError(err)
	var provEnv struct {
		Data domain.AgentProviderResponse `json:"data"`
	}
	s.Require().NoError(json.NewDecoder(pResp.Body).Decode(&provEnv))
	_ = pResp.Body.Close()

	_, err = s.settings.Update(ctx, map[string]string{
		agent.SettingModelGabriel:  provEnv.Data.ID.String() + ":gpt-4o",
		agent.SettingPromptGabriel: "Gabriel prompt with {curly_braces} test.",
	})
	s.Require().NoError(err)

	// Start real HTTP server over Fiber listener? Fiber v3 app.Test buffers response, so let's test via s.fiberApp.Test for full stream AND net.Listener if needed.
	// 1. Streaming delegation turn
	chatReq := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"Paroki apa yang aktif?"}`))
	chatReq.Header.Set("Content-Type", "application/json")
	chatReq.AddCookie(s.memberCookie)
	chatReq.Header.Set("X-Paroki-Id", s.parokiID.String())
	chatResp, err := s.fiberApp.Test(chatReq, fiber.TestConfig{Timeout: 10 * time.Second})
	s.Require().NoError(err)
	defer func() { _ = chatResp.Body.Close() }()
	s.Equal(http.StatusOK, chatResp.StatusCode)

	events := parseSSEEvents(s.T(), chatResp.Body)
	var eventNames []string
	var sessionID string
	var streamedText strings.Builder
	for _, ev := range events {
		eventNames = append(eventNames, ev.name)
		if ev.name == "session" {
			var payload map[string]any
			_ = json.Unmarshal([]byte(ev.data), &payload)
			sessionID, _ = payload["sessionId"].(string)
		}
		if ev.name == "token" {
			var payload map[string]string
			_ = json.Unmarshal([]byte(ev.data), &payload)
			s.Equal("gabriel", payload["agent"])
			streamedText.WriteString(payload["delta"])
		}
	}
	s.Require().NotEmpty(sessionID)
	s.Contains(eventNames, "session")
	s.Contains(eventNames, "agent_transfer")
	s.Contains(eventNames, "tool_start")
	s.Contains(eventNames, "tool_end")
	s.Contains(eventNames, "token")
	s.Contains(eventNames, "done")
	s.Equal("Paroki aktif Anda adalah Paroki Katedral.", streamedText.String())

	// Verify persisted messages in DB
	sid, err := uuid.Parse(sessionID)
	s.Require().NoError(err)
	msgs, err := s.client.AgentMessage.Query().All(ctx)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(msgs), 2)
	_ = sid

	// 2. Concurrent request on same sessionId -> 409 AGENT_SESSION_BUSY
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b, _ := json.Marshal(domain.ChatRequest{SessionID: &sid, Message: "Turn 2"})
		r := httptest.NewRequest(http.MethodPost, "/api/agent/chat", bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.AddCookie(s.memberCookie)
		r.Header.Set("X-Paroki-Id", s.parokiID.String())
		resp, err := s.fiberApp.Test(r, fiber.TestConfig{Timeout: 10 * time.Second})
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	// Wait briefly for Turn 2 to acquire the session lock and block on slowGate
	time.Sleep(80 * time.Millisecond)
	bBusy, _ := json.Marshal(domain.ChatRequest{SessionID: &sid, Message: "Concurrent"})
	rBusy := httptest.NewRequest(http.MethodPost, "/api/agent/chat", bytes.NewReader(bBusy))
	rBusy.Header.Set("Content-Type", "application/json")
	rBusy.AddCookie(s.memberCookie)
	rBusy.Header.Set("X-Paroki-Id", s.parokiID.String())
	respBusy, err := s.fiberApp.Test(rBusy)
	s.Require().NoError(err)
	_ = respBusy.Body.Close()
	s.Equal(http.StatusConflict, respBusy.StatusCode)

	close(slowGate)
	wg.Wait()
}

func (s *ChatSuite) TestAbortMidStreamCancelsAndPersistsCompletedTools() {
	ctx := context.Background()
	var callNum atomic.Int32
	abortGate := make(chan struct{})
	s.T().Cleanup(func() {
		select {
		case <-abortGate:
		default:
			close(abortGate)
		}
	})

	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": "gpt-4o", "name": "GPT-4o"}},
			})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			n := callNum.Add(1)
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher, _ := w.(http.Flusher)

			switch n {
			case 1:
				writeOpenAIToolChunk(w, "call_rem", "agent_remember", `{"content":"Suka misa pagi","category":"preference"}`)
				if flusher != nil {
					flusher.Flush()
				}
			case 2:
				// Emit partial chunk then keep streaming periodically so Flush() detects client abort
				for i := 0; i < 50; i++ {
					text := "more"
					if i == 0 {
						text = "Partial text that must be discarded"
					}
					chunk := map[string]any{
						"id": "chatcmpl-2", "object": "chat.completion.chunk", "created": 1, "model": "gpt-4o",
						"choices": []map[string]any{
							{"index": 0, "delta": map[string]any{"role": "assistant", "content": text}},
						},
					}
					b, _ := json.Marshal(chunk)
					_, _ = fmt.Fprintf(w, "data: %s\n\n", b)
					if flusher != nil {
						flusher.Flush()
					}
					select {
					case <-r.Context().Done():
						return
					case <-abortGate:
						return
					case <-time.After(30 * time.Millisecond):
					}
				}
			default:
				writeOpenAITextChunk(w, "Turn 2 ok")
				if flusher != nil {
					flusher.Flush()
				}
			}
			return
		}
		http.NotFound(w, r)
	})

	provBody, _ := json.Marshal(domain.CreateAgentProviderRequest{
		Name:     "OpenAI",
		Provider: "openai",
		APIKey:   "sk-test",
		BaseURL:  s.llmServer.URL,
	})
	pReq := httptest.NewRequest(http.MethodPost, "/api/agent/providers", bytes.NewReader(provBody))
	pReq.Header.Set("Content-Type", "application/json")
	pReq.AddCookie(s.superCookie)
	pResp, err := s.fiberApp.Test(pReq)
	s.Require().NoError(err)
	var provEnv struct {
		Data domain.AgentProviderResponse `json:"data"`
	}
	s.Require().NoError(json.NewDecoder(pResp.Body).Decode(&provEnv))
	_ = pResp.Body.Close()

	_, err = s.settings.Update(ctx, map[string]string{
		agent.SettingModelGabriel: provEnv.Data.ID.String() + ":gpt-4o",
	})
	s.Require().NoError(err)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	go func() {
		_ = s.fiberApp.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true})
	}()
	s.T().Cleanup(func() {
		_ = s.fiberApp.Shutdown()
	})

	abortCtx, cancelAbort := context.WithCancel(ctx)
	defer cancelAbort()
	req, err := http.NewRequestWithContext(abortCtx, http.MethodPost, "http://"+ln.Addr().String()+"/api/agent/chat", strings.NewReader(`{"message":"Catat preferensi saya"}`))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(s.memberCookie)
	req.Header.Set("X-Paroki-Id", s.parokiID.String())

	resp, err := (&http.Client{}).Do(req)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	var sessionID string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if strings.Contains(data, "sessionId") && sessionID == "" {
				var p map[string]any
				_ = json.Unmarshal([]byte(data), &p)
				sessionID, _ = p["sessionId"].(string)
			}
			if strings.Contains(data, "Partial text") {
				cancelAbort()
				break
			}
		}
	}
	s.Require().NotEmpty(sessionID)

	// Wait for server-side cancel + persistence + lock release
	sid, err := uuid.Parse(sessionID)
	s.Require().NoError(err)
	s.Eventually(func() bool {
		msgs, err := s.client.AgentMessage.Query().All(ctx)
		if err != nil {
			return false
		}
		hasTool := false
		for _, m := range msgs {
			if m.Role == "tool" {
				hasTool = true
			}
			s.NotContains(m.Content, "Partial text that must be discarded")
		}
		return hasTool
	}, 3*time.Second, 50*time.Millisecond)

	// Verify lock was released by sending a follow-up request on the same sessionId
	followReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+ln.Addr().String()+"/api/agent/chat", strings.NewReader(fmt.Sprintf(`{"sessionId":%q,"message":"Lanjut"}`, sid.String())))
	followReq.Header.Set("Content-Type", "application/json")
	followReq.AddCookie(s.memberCookie)
	followReq.Header.Set("X-Paroki-Id", s.parokiID.String())
	followResp, err := (&http.Client{}).Do(followReq)
	s.Require().NoError(err)
	_ = followResp.Body.Close()
	s.Equal(http.StatusOK, followResp.StatusCode)
}

func (s *ChatSuite) TestChatRetryDoesNotDuplicateUserMessage() {
	ctx := context.Background()
	var callNum atomic.Int32

	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{{"id": "gpt-4o", "name": "GPT-4o"}},
			})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			n := callNum.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"upstream failed"}`))
				return
			}
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			writeOpenAITextChunk(w, "Berhasil setelah retry")
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			return
		}
		http.NotFound(w, r)
	})

	provBody, _ := json.Marshal(domain.CreateAgentProviderRequest{
		Name:     "OpenAI",
		Provider: "openai",
		APIKey:   "sk-test",
		BaseURL:  s.llmServer.URL,
	})
	pReq := httptest.NewRequest(http.MethodPost, "/api/agent/providers", bytes.NewReader(provBody))
	pReq.Header.Set("Content-Type", "application/json")
	pReq.AddCookie(s.superCookie)
	pResp, err := s.fiberApp.Test(pReq)
	s.Require().NoError(err)
	var provEnv struct {
		Data domain.AgentProviderResponse `json:"data"`
	}
	s.Require().NoError(json.NewDecoder(pResp.Body).Decode(&provEnv))
	_ = pResp.Body.Close()

	_, err = s.settings.Update(ctx, map[string]string{
		"agent.model.gabriel": provEnv.Data.ID.String() + ":gpt-4o",
	})
	s.Require().NoError(err)

	// First attempt fails mid-stream after persisting the user message
	req1 := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"Cek jadwal misa"}`))
	req1.Header.Set("Content-Type", "application/json")
	req1.AddCookie(s.memberCookie)
	req1.Header.Set("X-Paroki-Id", s.parokiID.String())
	resp1, err := s.fiberApp.Test(req1, fiber.TestConfig{Timeout: 10 * time.Second})
	s.Require().NoError(err)
	events1 := parseSSEEvents(s.T(), resp1.Body)
	_ = resp1.Body.Close()

	var sessionID string
	hasErr := false
	for _, ev := range events1 {
		if ev.name == "session" {
			var p map[string]any
			_ = json.Unmarshal([]byte(ev.data), &p)
			sessionID, _ = p["sessionId"].(string)
		}
		if ev.name == "error" {
			hasErr = true
		}
	}
	s.Require().NotEmpty(sessionID)
	s.True(hasErr)

	// Retry with retry: true should succeed without duplicating the user message
	retryBody := fmt.Sprintf(`{"sessionId":%q,"message":"Cek jadwal misa","retry":true}`, sessionID)
	req2 := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(retryBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.AddCookie(s.memberCookie)
	req2.Header.Set("X-Paroki-Id", s.parokiID.String())
	resp2, err := s.fiberApp.Test(req2, fiber.TestConfig{Timeout: 10 * time.Second})
	s.Require().NoError(err)
	events2 := parseSSEEvents(s.T(), resp2.Body)
	_ = resp2.Body.Close()

	hasDone := false
	for _, ev := range events2 {
		if ev.name == "done" {
			hasDone = true
		}
	}
	s.True(hasDone)

	sid, err := uuid.Parse(sessionID)
	s.Require().NoError(err)
	msgs, err := s.client.AgentMessage.Query().All(ctx)
	s.Require().NoError(err)
	userCount := 0
	assistantCount := 0
	for _, m := range msgs {
		if m.SessionID != sid {
			continue
		}
		switch m.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}
	s.Equal(1, userCount)
	s.Equal(1, assistantCount)
}

func (s *ChatSuite) TestChat_RuntimeContextAndPageContextResolution() {
	ctx := context.Background()
	err := s.client.Paroki.UpdateOneID(s.parokiID).SetTimezone("Asia/Jakarta").Exec(ctx)
	s.Require().NoError(err)

	orgRow, err := s.client.Organization.Create().
		SetParokiID(s.parokiID).
		SetCode("omk-yosef").
		SetName("OMK St. Yosef").
		Save(ctx)
	s.Require().NoError(err)

	provBody, _ := json.Marshal(map[string]any{
		"name":     "Mock OpenAI",
		"provider": "openai",
		"baseUrl":  s.llmServer.URL,
		"apiKey":   "sk-test-secret",
	})
	provReq := httptest.NewRequest(http.MethodPost, "/api/agent/providers", bytes.NewReader(provBody))
	provReq.Header.Set("Content-Type", "application/json")
	provReq.AddCookie(s.superCookie)
	provResp, err := s.fiberApp.Test(provReq)
	s.Require().NoError(err)
	defer func() { _ = provResp.Body.Close() }()
	var provEnv struct {
		Data domain.AgentProviderResponse `json:"data"`
	}
	s.Require().NoError(json.NewDecoder(provResp.Body).Decode(&provEnv))

	_, err = s.settings.Update(ctx, map[string]string{
		"agent.model.gabriel": provEnv.Data.ID.String() + ":gpt-4o",
	})
	s.Require().NoError(err)

	var capturedSystem string
	s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		for _, m := range body.Messages {
			if m.Role == "system" {
				capturedSystem = m.Content
			}
		}
		w.Header().Set("Content-Type", "text/event-stream")
		writeOpenAITextChunk(w, "Halo dari Gabriel!")
	})

	chatBody := fmt.Sprintf(`{"message":"Info organisasi ini?","pageContext":{"path":"/org/%s?tab=members","title":"Detail Organisasi"}}`, orgRow.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(chatBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(s.memberCookie)
	req.Header.Set("X-Paroki-Id", s.parokiID.String())

	resp, err := s.fiberApp.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	s.Require().NoError(err)
	_ = parseSSEEvents(s.T(), resp.Body)
	_ = resp.Body.Close()

	s.Contains(capturedSystem, "- User Timezone: Asia/Makassar")
	s.Contains(capturedSystem, "- Paroki Timezone: Asia/Jakarta")
	s.Contains(capturedSystem, "+07:00")
	s.Contains(capturedSystem, fmt.Sprintf("- Active Page: /org/%s?tab=members (Detail Organisasi)", orgRow.ID))
	s.Contains(capturedSystem, fmt.Sprintf("- Active Organization: OMK St. Yosef (orgId: %s)", orgRow.ID))
	s.Contains(capturedSystem, "- Language Rule: Always reply to the user in Bahasa Indonesia.")
}

func (s *ChatSuite) TestChat_SubAgentDelegation_YosefGregoriusSesilia() {
	ctx := context.Background()
	provBody, _ := json.Marshal(map[string]any{
		"name":     "Mock OpenAI",
		"provider": "openai",
		"baseUrl":  s.llmServer.URL,
		"apiKey":   "sk-test-secret",
	})
	provReq := httptest.NewRequest(http.MethodPost, "/api/agent/providers", bytes.NewReader(provBody))
	provReq.Header.Set("Content-Type", "application/json")
	provReq.AddCookie(s.superCookie)
	provResp, err := s.fiberApp.Test(provReq)
	s.Require().NoError(err)
	defer func() { _ = provResp.Body.Close() }()
	var provEnv struct {
		Data domain.AgentProviderResponse `json:"data"`
	}
	s.Require().NoError(json.NewDecoder(provResp.Body).Decode(&provEnv))

	_, err = s.settings.Update(ctx, map[string]string{
		"agent.model.gabriel": provEnv.Data.ID.String() + ":gpt-4o",
	})
	s.Require().NoError(err)

	for _, target := range []string{"yosef", "gregorius", "sesilia"} {
		s.Run(target, func() {
			var turn atomic.Int32
			s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
				rawBody, _ := io.ReadAll(r.Body)
				if strings.Contains(string(rawBody), "Memory Extractor") {
					w.Header().Set("Content-Type", "text/event-stream")
					writeOpenAITextChunk(w, "[]")
					return
				}
				step := turn.Add(1)
				w.Header().Set("Content-Type", "text/event-stream")
				switch step {
				case 1:
					writeOpenAIToolChunk(w, "call_sub", target, `{"request":"Cek data"}`)
				case 2:
					writeOpenAITextChunk(w, "Hasil dari "+target)
				default:
					writeOpenAITextChunk(w, "Selesai dari "+target)
				}
			})

			req := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"Tolong cek"}`))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(s.memberCookie)
			req.Header.Set("X-Paroki-Id", s.parokiID.String())

			resp, err := s.fiberApp.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
			s.Require().NoError(err)
			events := parseSSEEvents(s.T(), resp.Body)
			_ = resp.Body.Close()

			foundTransfer := false
			for _, ev := range events {
				if ev.name == "agent_transfer" && strings.Contains(ev.data, `"`+target+`"`) {
					foundTransfer = true
				}
			}
			s.True(foundTransfer, "expected agent_transfer to %s, got events: %+v", target, events)
		})
	}
}

func (s *ChatSuite) TestChat_AgentAskUserAndNavigate_SSEAndMetaPersistence() {
	ctx := context.Background()
	provBody, _ := json.Marshal(map[string]any{
		"name":     "Mock OpenAI",
		"provider": "openai",
		"baseUrl":  s.llmServer.URL,
		"apiKey":   "sk-test-secret",
	})
	provReq := httptest.NewRequest(http.MethodPost, "/api/agent/providers", bytes.NewReader(provBody))
	provReq.Header.Set("Content-Type", "application/json")
	provReq.AddCookie(s.superCookie)
	provResp, err := s.fiberApp.Test(provReq)
	s.Require().NoError(err)
	defer func() { _ = provResp.Body.Close() }()
	var provEnv struct {
		Data domain.AgentProviderResponse `json:"data"`
	}
	s.Require().NoError(json.NewDecoder(provResp.Body).Decode(&provEnv))

	_, err = s.settings.Update(ctx, map[string]string{
		"agent.model.gabriel": provEnv.Data.ID.String() + ":gpt-4o",
	})
	s.Require().NoError(err)

	s.Run("agent_ask_user", func() {
		var turn atomic.Int32
		s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
			rawBody, _ := io.ReadAll(r.Body)
			if strings.Contains(string(rawBody), "Memory Extractor") {
				w.Header().Set("Content-Type", "text/event-stream")
				writeOpenAITextChunk(w, "[]")
				return
			}
			step := turn.Add(1)
			w.Header().Set("Content-Type", "text/event-stream")
			if step == 1 {
				writeOpenAIToolChunk(w, "call_ask", "agent_ask_user", `{"question":"Pilih misa?","options":[{"label":"Misa Pagi"},{"label":"Misa Sore"}]}`)
				return
			}
			writeOpenAITextChunk(w, "")
		})

		req := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"Jadwalkan misa"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(s.memberCookie)
		req.Header.Set("X-Paroki-Id", s.parokiID.String())

		resp, err := s.fiberApp.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		s.Require().NoError(err)
		events := parseSSEEvents(s.T(), resp.Body)
		_ = resp.Body.Close()

		var questionEv, doneEv sseEvent
		for _, ev := range events {
			if ev.name == "question" {
				questionEv = ev
			}
			if ev.name == "done" {
				doneEv = ev
			}
		}
		s.NotEmpty(questionEv.data)
		var qPayload domain.QuestionPayload
		s.Require().NoError(json.Unmarshal([]byte(questionEv.data), &qPayload))
		s.Equal("Pilih misa?", qPayload.Question)
		s.True(qPayload.AllowCustom)
		s.False(qPayload.MultiSelect)
		s.Equal("Misa Pagi", qPayload.Options[0].Value)

		var doneData map[string]string
		s.Require().NoError(json.Unmarshal([]byte(doneEv.data), &doneData))
		msgID, err := uuid.Parse(doneData["messageId"])
		s.Require().NoError(err)
		msgRow, err := s.client.AgentMessage.Get(ctx, msgID)
		s.Require().NoError(err)
		s.Equal("Pilih misa?", msgRow.Content)
		s.NotNil(msgRow.Meta["question"])
	})

	s.Run("agent_navigate", func() {
		var turn atomic.Int32
		s.setLLMHandler(func(w http.ResponseWriter, r *http.Request) {
			rawBody, _ := io.ReadAll(r.Body)
			if strings.Contains(string(rawBody), "Memory Extractor") {
				w.Header().Set("Content-Type", "text/event-stream")
				writeOpenAITextChunk(w, "[]")
				return
			}
			step := turn.Add(1)
			w.Header().Set("Content-Type", "text/event-stream")
			if step == 1 {
				writeOpenAIToolChunk(w, "call_nav", "agent_navigate", `{"path":"/umat","label":"Buka Daftar Umat","autoRedirect":true}`)
				return
			}
			writeOpenAITextChunk(w, "Membuka halaman daftar umat.")
		})

		req := httptest.NewRequest(http.MethodPost, "/api/agent/chat", strings.NewReader(`{"message":"Buka halaman umat"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(s.memberCookie)
		req.Header.Set("X-Paroki-Id", s.parokiID.String())

		resp, err := s.fiberApp.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		s.Require().NoError(err)
		events := parseSSEEvents(s.T(), resp.Body)
		_ = resp.Body.Close()

		var navEv, doneEv sseEvent
		for _, ev := range events {
			if ev.name == "navigate" {
				navEv = ev
			}
			if ev.name == "done" {
				doneEv = ev
			}
		}
		s.NotEmpty(navEv.data)
		var navPayload domain.NavigationPayload
		s.Require().NoError(json.Unmarshal([]byte(navEv.data), &navPayload))
		s.Equal("/umat", navPayload.Path)
		s.True(navPayload.AutoRedirect)

		var doneData map[string]string
		s.Require().NoError(json.Unmarshal([]byte(doneEv.data), &doneData))
		msgID, err := uuid.Parse(doneData["messageId"])
		s.Require().NoError(err)
		msgRow, err := s.client.AgentMessage.Get(ctx, msgID)
		s.Require().NoError(err)
		s.NotNil(msgRow.Meta["navigation"])
	})
}

type sseEvent struct {
	name string
	data string
}

func parseSSEEvents(t *testing.T, r io.Reader) []sseEvent {
	t.Helper()
	var events []sseEvent
	scanner := bufio.NewScanner(r)
	var cur sseEvent
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if cur.name != "" || cur.data != "" {
				events = append(events, cur)
				cur = sseEvent{}
			}
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			cur.name = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			cur.data = strings.TrimPrefix(line, "data: ")
		}
	}
	require.NoError(t, scanner.Err())
	return events
}

func writeOpenAIToolChunk(w io.Writer, callID, name, args string) {
	mid := len(args) / 2
	part1, part2 := args[:mid], args[mid:]
	chunks := []map[string]any{
		{
			"id": "chatcmpl-1", "object": "chat.completion.chunk", "created": 1, "model": "gpt-4o",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"role": "assistant",
						"tool_calls": []map[string]any{
							{
								"index": 0,
								"id":    callID,
								"type":  "function",
								"function": map[string]any{
									"name":      name,
									"arguments": part1,
								},
							},
						},
					},
				},
			},
		},
		{
			"id": "chatcmpl-1", "object": "chat.completion.chunk", "created": 1, "model": "gpt-4o",
			"choices": []map[string]any{
				{
					"index": 0,
					"delta": map[string]any{
						"tool_calls": []map[string]any{
							{
								"index": 0,
								"function": map[string]any{
									"arguments": part2,
								},
							},
						},
					},
				},
			},
		},
	}
	for _, chunk := range chunks {
		b, _ := json.Marshal(chunk)
		_, _ = fmt.Fprintf(w, "data: %s\n\n", b)
	}
	_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
}

func writeOpenAITextChunk(w io.Writer, text string) {
	chunk := map[string]any{
		"id": "chatcmpl-1", "object": "chat.completion.chunk", "created": 1, "model": "gpt-4o",
		"choices": []map[string]any{
			{
				"index":         0,
				"delta":         map[string]any{"role": "assistant", "content": text},
				"finish_reason": "stop",
			},
		},
	}
	b, _ := json.Marshal(chunk)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", b)
	_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
}
