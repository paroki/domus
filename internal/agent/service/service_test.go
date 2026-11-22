package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubConfigProvider struct {
	agents map[string]domain.AgentConfig
	emb    EmbeddingConfigSnapshot
}

func (s stubConfigProvider) AgentConfig(name string) (domain.AgentConfig, bool) {
	c, ok := s.agents[name]
	return c, ok
}

func (s stubConfigProvider) EmbeddingConfig() EmbeddingConfigSnapshot {
	return s.emb
}

type stubUserReader struct{ profile *UserProfileInfo }

func (s stubUserReader) GetUser(_ context.Context, _ uuid.UUID) (*UserProfileInfo, error) {
	return s.profile, nil
}

type stubParishReader struct{ info *ParishInfo }

func (s stubParishReader) GetParish(_ context.Context, _ uuid.UUID) (*ParishInfo, error) {
	return s.info, nil
}

func (s stubParishReader) GetParishByCode(_ context.Context, _ string) (*ParishInfo, error) {
	return s.info, nil
}

type stubOrgReader struct{ orgs []OrgInfo }

func (s stubOrgReader) ListUserOrgs(_ context.Context, _, _ uuid.UUID) ([]OrgInfo, error) {
	return s.orgs, nil
}

func TestComputeAgeAndSalutation_AndBuildSystemPrompt(t *testing.T) {
	now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	pria := "pria"
	wanita := "wanita"
	young := "2002-05-10"
	adult := "1990-01-01"

	age, sal := ComputeAgeAndSalutation("Andi Wijaya", &pria, &young, now)
	require.NotNil(t, age)
	assert.Equal(t, 24, *age)
	assert.Equal(t, "Kak Andi", sal)

	age, sal = ComputeAgeAndSalutation("Budi Santoso", &pria, &adult, now)
	require.NotNil(t, age)
	assert.Equal(t, 36, *age)
	assert.Equal(t, "Bapak Budi", sal)

	_, sal = ComputeAgeAndSalutation("Siti Rahma", &wanita, &adult, now)
	assert.Equal(t, "Ibu Siti", sal)

	ageNil, salNil := ComputeAgeAndSalutation("Maria", nil, nil, now)
	assert.Nil(t, ageNil)
	assert.Equal(t, "Kak Maria", salNil)

	prompt := BuildSystemPrompt(PromptInput{
		Instruction: "Base instruction",
		Channel:     domain.ChannelWeb,
		Now:         now,
		Parish:      &ParishInfo{ID: uuid.New(), Code: "p-01", Name: "Paroki Katedral"},
		Orgs:        []OrgInfo{{ID: uuid.New(), Code: "koor", Name: "Koor St. Cecilia", Kind: "group", Role: "member"}},
		User:        &UserProfileInfo{ID: uuid.New(), Name: "Budi Santoso", Gender: &pria, BirthDate: &adult},
		PinnedMemories: []domain.Memory{
			{ID: uuid.New(), Scope: domain.MemoryScopeUser, Category: domain.MemoryCategoryProfile, Content: "Sapa Pak Budi", IsPinned: true},
		},
	})
	assert.Contains(t, prompt, "<channel>web</channel>")
	assert.Contains(t, prompt, "<default_salutation>Bapak Budi</default_salutation>")
	assert.Contains(t, prompt, "<user_memory>")
	assert.Contains(t, prompt, "Sapa Pak Budi")
}

func TestService_RunTurn_MemoryAndDelegateSpecialist(t *testing.T) {
	ctx := context.Background()
	client, sqlDB, d := testutil.NewEntClientWithDB(t)
	r := repo.New(client, sqlDB, d)

	enforcer, err := authz.NewMemoryEnforcer()
	require.NoError(t, err)

	parishID := uuid.New()
	userID := uuid.New()
	principal := authx.Principal{
		UserID:   userID,
		ParishID: parishID,
	}

	// Register an event_list tool on MCP server to verify in-memory propagation.
	mcpSrv := mcpx.NewServer("test-agent")
	var mcpPrincipalUserID string
	mcp.AddTool(mcpSrv, &mcp.Tool{
		Name:        "event_list",
		Description: "List events",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, map[string]any, error) {
		if p, ok := authx.UserContext(ctx); ok {
			mcpPrincipalUserID = p.UserID.String()
		}
		return nil, map[string]any{"events": []string{"Misa Natal"}}, nil
	})

	var turnStep atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		step := turnStep.Add(1)
		switch step {
		case 1:
			// Gabriel calls memory_save (parish scope -> forbidden for plain user)
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"tc_1\",\"function\":{\"name\":\"memory_save\",\"arguments\":\"{\\\"content\\\":\\\"FAQ\\\",\\\"category\\\":\\\"parish_faq\\\",\\\"scope\\\":\\\"parish\\\"}\"}}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		case 2:
			// Gabriel delegates to Mikael
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"tc_2\",\"function\":{\"name\":\"delegate_specialist\",\"arguments\":\"{\\\"agent_name\\\":\\\"mikael\\\",\\\"task\\\":\\\"Cek acara\\\"}\"}}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		case 3:
			// Mikael calls event_list via MCP
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"thinking\":\"\",\"reasoning_content\":\"Cek event\",\"tool_calls\":[{\"index\":0,\"id\":\"tc_3\",\"function\":{\"name\":\"event_list\",\"arguments\":\"{}\"}}]}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		case 4:
			// Mikael returns final answer (streamed directly to user; Gabriel does not reply again)
			_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Ada Misa Natal.\"}}]}\n\n")
			_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		default:
			t.Fatalf("unexpected extra LLM call step %d after delegation handoff", step)
		}
	}))
	defer srv.Close()

	cfg := stubConfigProvider{
		agents: map[string]domain.AgentConfig{
			"gabriel": {
				Name:        "gabriel",
				Instruction: "Gabriel",
				Model:       domain.ModelConfig{Provider: "openai", BaseURL: srv.URL, APIKey: "k", Model: "gpt-4o"},
				Tools:       []string{"delegate_specialist", "memory_*"},
				MaxSteps:    domain.MaxSteps,
			},
			"mikael": {
				Name:        "mikael",
				Instruction: "Mikael",
				Model:       domain.ModelConfig{Provider: "openai", BaseURL: srv.URL, APIKey: "k", Model: "gpt-4o"},
				Tools:       []string{"event_*"},
				MaxSteps:    domain.MaxSteps,
			},
		},
	}

	svc := New(
		r,
		cfg,
		driver.NewRegistry(srv.Client()),
		enforcer,
		mcpSrv,
		stubUserReader{profile: &UserProfileInfo{ID: userID, Name: "Budi"}},
		stubParishReader{info: &ParishInfo{ID: parishID, Code: "p-1", Name: "Katedral"}},
		stubOrgReader{},
		slog.New(slog.DiscardHandler),
	)

	var events []domain.AgentSSEEvent
	sess, answer, err := svc.RunTurn(ctx, ChatTurnInput{
		Principal: principal,
		Channel:   domain.ChannelWeb,
		Message:   "Ada acara apa?",
		OnStep: func(ev domain.AgentSSEEvent) error {
			events = append(events, ev)
			return nil
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Ada Misa Natal.", answer)
	assert.Equal(t, int32(4), turnStep.Load())
	assert.Equal(t, userID.String(), mcpPrincipalUserID)

	// Verify parish-scope RBAC rejection was returned as structured tool_result and Mikael delta streamed directly
	var sawForbidden, sawMikaelDelta, sawMikaelToolResult, sawDone bool
	for _, ev := range events {
		if ev.Event == "tool_result" && ev.ToolName == "memory_save" && ev.AgentName == "gabriel" {
			var payload map[string]map[string]string
			if json.Unmarshal([]byte(ev.Content), &payload) == nil && payload["error"]["code"] == "AGENT_FORBIDDEN" {
				sawForbidden = true
			}
		}
		if ev.Event == "delta" && ev.AgentName == "mikael" {
			sawMikaelDelta = true
		}
		if ev.Event == "tool_result" && ev.AgentName == "mikael" && ev.ToolName == "event_list" {
			sawMikaelToolResult = true
		}
		if ev.Event == "done" {
			sawDone = true
		}
	}
	assert.True(t, sawForbidden)
	assert.True(t, sawMikaelDelta)
	assert.True(t, sawMikaelToolResult)
	assert.True(t, sawDone)

	// Grant ref:parish-admin and verify parish-scope memory_save succeeds
	require.NoError(t, enforcer.AssignRole(ctx, principal.Subject(), roleParishAdmin, authz.ParishDomain(parishID)))
	resStr, err := svc.execMemorySave(ctx, principal, sess.ID, memorySaveArgs{
		Content:  "Misa jam 07:00",
		Category: "parish_faq",
		Scope:    "parish",
	})
	require.NoError(t, err)
	assert.Contains(t, resStr, `"scope":"parish"`)
}

func TestService_SessionBusy_AndMaxStepsExceeded(t *testing.T) {
	ctx := context.Background()
	client, sqlDB, d := testutil.NewEntClientWithDB(t)
	r := repo.New(client, sqlDB, d)

	parishID := uuid.New()
	userID := uuid.New()
	principal := authx.Principal{UserID: userID, ParishID: parishID}

	sess, err := r.CreateSession(ctx, domain.CreateSessionInput{
		ParishID: parishID,
		UserID:   userID,
		Channel:  domain.ChannelWeb,
		Title:    "Busy Test",
	})
	require.NoError(t, err)

	// Always return a tool call so MaxSteps = 8 is exceeded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"tc_loop\",\"function\":{\"name\":\"memory_recall\",\"arguments\":\"{\\\"query\\\":\\\"x\\\"}\"}}]}}]}\n\n")
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	cfg := stubConfigProvider{
		agents: map[string]domain.AgentConfig{
			"gabriel": {
				Name:        "gabriel",
				Instruction: "Gabriel",
				Model:       domain.ModelConfig{Provider: "openai", BaseURL: srv.URL, APIKey: "k", Model: "gpt-4o"},
				Tools:       []string{"memory_*"},
				MaxSteps:    domain.MaxSteps,
			},
		},
	}

	svc := New(r, cfg, driver.NewRegistry(srv.Client()), nil, nil, nil, nil, nil, nil)

	// Empty message -> ErrEmptyMessage
	_, _, err = svc.RunTurn(ctx, ChatTurnInput{Principal: principal, Message: "   "})
	assert.ErrorIs(t, err, domain.ErrEmptyMessage)

	// Busy lock -> ErrSessionBusy
	require.True(t, svc.locker.tryLock(sess.ID))
	_, _, err = svc.RunTurn(ctx, ChatTurnInput{Principal: principal, SessionID: &sess.ID, Message: "Halo"})
	assert.ErrorIs(t, err, domain.ErrSessionBusy)
	svc.locker.unlock(sess.ID)

	// MaxSteps exceeded -> ErrMaxStepsExceeded
	_, _, err = svc.RunTurn(ctx, ChatTurnInput{Principal: principal, SessionID: &sess.ID, Message: "Loop"})
	assert.ErrorIs(t, err, domain.ErrMaxStepsExceeded)
}

func TestService_BuildPromptForTurn_InjectsUserRoles(t *testing.T) {
	ctx := context.Background()
	client, sqlDB, d := testutil.NewEntClientWithDB(t)
	r := repo.New(client, sqlDB, d)
	enforcer, err := authz.NewEnforcer(client)
	require.NoError(t, err)

	parishID := uuid.New()
	userID := uuid.New()
	principal := authx.Principal{UserID: userID, ParishID: parishID}

	require.NoError(t, enforcer.AssignRole(ctx, principal.Subject(), "user:super-admin", authz.DomainGlobal))
	require.NoError(t, enforcer.AssignRole(ctx, principal.Subject(), roleParishAdmin, authz.ParishDomain(parishID)))

	svc := New(
		r,
		stubConfigProvider{},
		driver.NewRegistry(nil),
		enforcer,
		nil,
		stubUserReader{profile: &UserProfileInfo{ID: userID, Name: "Anthonius"}},
		stubParishReader{info: &ParishInfo{ID: parishID, Code: "p-1", Name: "Katedral"}},
		stubOrgReader{},
		slog.New(slog.DiscardHandler),
	)

	prompt, err := svc.buildPromptForTurn(ctx, principal, domain.ChannelWeb, "Gabriel")
	require.NoError(t, err)
	assert.Contains(t, prompt, "<user_roles>\n    <role>ref:parish-admin</role>\n    <role>user:super-admin</role>\n  </user_roles>")
}
