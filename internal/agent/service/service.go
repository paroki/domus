package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/gowa"
	"github.com/paroki/domus/internal/platform/mcpx"
)

// StepCallback receives real-time orchestration events during an agent turn.
type StepCallback func(event domain.AgentSSEEvent) error

// ConfigProvider exposes hot-reloaded agent and embedding snapshots.
type ConfigProvider interface {
	AgentConfig(name string) (domain.AgentConfig, bool)
	EmbeddingConfig() EmbeddingConfigSnapshot
}

// EmbeddingConfigSnapshot holds resolved embedding provider settings.
type EmbeddingConfigSnapshot struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

type sessionLocker struct {
	mu     sync.Mutex
	active map[uuid.UUID]struct{}
}

func newSessionLocker() *sessionLocker {
	return &sessionLocker{active: make(map[uuid.UUID]struct{})}
}

func (l *sessionLocker) tryLock(id uuid.UUID) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, busy := l.active[id]; busy {
		return false
	}
	l.active[id] = struct{}{}
	return true
}

func (l *sessionLocker) unlock(id uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.active, id)
}

// Service orchestrates AI agent conversations, 3-tier memory, and MCP tool execution.
type Service struct {
	repo            *repo.Repository
	config          ConfigProvider
	drivers         *driver.Registry
	enforcer        *authz.Enforcer
	mcpServer       *mcp.Server
	users           UserReader
	parishes        ParishReader
	orgs            OrgReader
	channelResolver ChannelUserResolver
	locker          *sessionLocker
	dedup           dedupCache
	logger          *slog.Logger
	nowFn           func() time.Time
}

// New constructs an agent Service.
func New(
	r *repo.Repository,
	cfg ConfigProvider,
	drivers *driver.Registry,
	enforcer *authz.Enforcer,
	mcpServer *mcp.Server,
	users UserReader,
	parishes ParishReader,
	orgs OrgReader,
	logger *slog.Logger,
) *Service {
	if drivers == nil {
		drivers = driver.NewRegistry(nil)
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Service{
		repo:      r,
		config:    cfg,
		drivers:   drivers,
		enforcer:  enforcer,
		mcpServer: mcpServer,
		users:     users,
		parishes:  parishes,
		orgs:      orgs,
		locker:    newSessionLocker(),
		logger:    logger,
		nowFn:     func() time.Time { return time.Now().UTC() },
	}
}

// SetMCPServer binds the in-process MCP server used by tool calls.
func (s *Service) SetMCPServer(srv *mcp.Server) {
	s.mcpServer = srv
}

// Logger returns the service's structured logger.
func (s *Service) Logger() *slog.Logger {
	return s.logger
}

// SetChannelResolver binds the user channel resolver and messaging clients.
func (s *Service) SetChannelResolver(r ChannelUserResolver) {
	s.channelResolver = r
}

// GowaClient returns the active hot-reloaded GOWA WhatsApp client (or nil).
func (s *Service) GowaClient() *gowa.Client {
	if s.channelResolver == nil {
		return nil
	}
	return s.channelResolver.GowaClient()
}

// ChatTurnInput holds parameters for executing a single user turn.
type ChatTurnInput struct {
	Principal authx.Principal
	Channel   domain.Channel
	SessionID *uuid.UUID
	Message   string
	Continue  bool
	OnStep    StepCallback
}

// PreparedTurn holds a validated, locked session ready to execute a ReAct turn.
type PreparedTurn struct {
	Principal authx.Principal
	Channel   domain.Channel
	Session   *ent.AgentSession
	Message   string
	Continue  bool
	Gabriel   domain.AgentConfig
	locker    *sessionLocker
	released  bool
}

// Release unlocks the per-session turn mutex if not already released.
func (p *PreparedTurn) Release() {
	if p == nil || p.locker == nil || p.released {
		return
	}
	p.released = true
	p.locker.unlock(p.Session.ID)
}

// PrepareTurn validates the message, session ownership, and provider availability,
// and acquires the per-session turn mutex before any SSE headers are written.
func (s *Service) PrepareTurn(ctx context.Context, in ChatTurnInput) (*PreparedTurn, error) {
	msgText := strings.TrimSpace(in.Message)
	if in.Continue {
		if in.SessionID == nil || *in.SessionID == uuid.Nil {
			return nil, domain.ValidationError("sessionId", "REQUIRED", "Session ID is required to continue.")
		}
	} else if msgText == "" {
		return nil, domain.ErrEmptyMessage
	}
	ch := in.Channel
	if ch == "" {
		ch = domain.ChannelWeb
	}

	sess, err := s.resolveSession(ctx, in.Principal, ch, in.SessionID, msgText)
	if err != nil {
		return nil, err
	}

	gabrielCfg, ok := s.config.AgentConfig("gabriel")
	if !ok || strings.TrimSpace(gabrielCfg.Model.APIKey) == "" {
		return nil, domain.ErrProviderUnavailable
	}

	if !s.locker.tryLock(sess.ID) {
		return nil, domain.ErrSessionBusy
	}

	return &PreparedTurn{
		Principal: in.Principal,
		Channel:   ch,
		Session:   sess,
		Message:   msgText,
		Continue:  in.Continue,
		Gabriel:   gabrielCfg,
		locker:    s.locker,
	}, nil
}

// RunTurn executes a multi-step ReAct turn for Gabriel (and delegated specialists).
func (s *Service) RunTurn(ctx context.Context, in ChatTurnInput) (*ent.AgentSession, string, error) {
	prep, err := s.PrepareTurn(ctx, in)
	if err != nil {
		return nil, "", err
	}
	return s.RunPreparedTurn(ctx, prep, in.OnStep)
}

// RunPreparedTurn executes a turn that was already validated and locked via PrepareTurn.
func (s *Service) RunPreparedTurn(ctx context.Context, prep *PreparedTurn, onStep StepCallback) (*ent.AgentSession, string, error) {
	defer prep.Release()

	sess := prep.Session
	gabrielCfg := prep.Gabriel

	if !prep.Continue && prep.Message != "" {
		if _, err := s.repo.AppendMessage(ctx, domain.CreateMessageInput{
			ParishID:  prep.Principal.ParishID,
			UserID:    prep.Principal.UserID,
			SessionID: sess.ID,
			AgentName: gabrielCfg.Name,
			Role:      domain.RoleUser,
			Content:   prep.Message,
		}); err != nil {
			return nil, "", err
		}
	}

	if err := emitStep(onStep, domain.AgentSSEEvent{
		Event:     "session",
		SessionID: &sess.ID,
		AgentName: gabrielCfg.Name,
	}); err != nil {
		return nil, "", err
	}

	turnCtx := authx.WithPrincipal(ctx, prep.Principal)
	var mcpSession *mcp.ClientSession
	var err error
	if s.mcpServer != nil {
		mcpSession, err = mcpx.ConnectInMemory(turnCtx, s.mcpServer)
		if err != nil {
			return nil, "", err
		}
		defer func() { _ = mcpSession.Close() }()
	}

	sysPrompt, err := s.buildPromptForTurn(turnCtx, prep.Principal, prep.Channel, gabrielCfg.Instruction)
	if err != nil {
		return nil, "", err
	}

	recentRows, err := s.repo.ListRecentMessages(turnCtx, prep.Principal.ParishID, sess.ID, domain.HistoryWindow)
	if err != nil {
		return nil, "", err
	}
	history := make([]driver.ChatMessage, 0, len(recentRows))
	for _, row := range recentRows {
		if domain.Role(row.Role) != domain.RoleUser && row.AgentName != "gabriel" && row.AgentName != "" {
			continue
		}
		var tcs []domain.ToolCall
		if len(row.ToolCalls) > 0 {
			tcs = make([]domain.ToolCall, len(row.ToolCalls))
			for i, tc := range row.ToolCalls {
				tcs[i] = domain.ToolCall{
					ID:        tc.ID,
					Name:      tc.Name,
					Arguments: tc.Arguments,
					AgentName: tc.AgentName,
				}
			}
		}
		var tcID, tcName string
		if row.ToolCallID != nil {
			tcID = *row.ToolCallID
		}
		if row.ToolName != nil {
			tcName = *row.ToolName
		}
		history = append(history, driver.ChatMessage{
			Role:       domain.Role(row.Role),
			Content:    row.Content,
			ToolCalls:  tcs,
			ToolCallID: tcID,
			ToolName:   tcName,
		})
	}

	sanitized := sanitizeHistory(history)
	if len(sanitized) == 0 {
		return nil, "", domain.ErrEmptyMessage
	}

	answer, err := s.runReActLoop(
		turnCtx,
		prep.Principal,
		prep.Channel,
		sess.ID,
		gabrielCfg,
		sysPrompt,
		sanitized,
		mcpSession,
		onStep,
		true,
	)
	if err != nil {
		return nil, "", err
	}

	if err := emitStep(onStep, domain.AgentSSEEvent{
		Event:     "done",
		SessionID: &sess.ID,
		AgentName: gabrielCfg.Name,
	}); err != nil {
		return nil, "", err
	}
	return sess, answer, nil
}

func (s *Service) resolveSession(
	ctx context.Context,
	p authx.Principal,
	ch domain.Channel,
	sessionID *uuid.UUID,
	firstMsg string,
) (*ent.AgentSession, error) {
	if ch == domain.ChannelWeb {
		if sessionID != nil {
			return s.repo.FindSessionByID(ctx, p.ParishID, p.UserID, *sessionID)
		}
		return s.repo.CreateSession(ctx, domain.CreateSessionInput{
			ParishID: p.ParishID,
			UserID:   p.UserID,
			Channel:  domain.ChannelWeb,
			Title:    truncateRunes(firstMsg, 80),
		})
	}
	return s.repo.FindOrCreateChannelSession(ctx, p.ParishID, p.UserID, ch, truncateRunes(firstMsg, 80))
}

func (s *Service) buildPromptForTurn(
	ctx context.Context,
	p authx.Principal,
	ch domain.Channel,
	baseInstruction string,
) (string, error) {
	var userProf *UserProfileInfo
	if s.users != nil && p.UserID != uuid.Nil {
		u, err := s.users.GetUser(ctx, p.UserID)
		if err == nil {
			userProf = u
		}
	}
	var parishInfo *ParishInfo
	if s.parishes != nil && p.ParishID != uuid.Nil {
		par, err := s.parishes.GetParish(ctx, p.ParishID)
		if err == nil {
			parishInfo = par
		}
	}
	var orgs []OrgInfo
	if s.orgs != nil && p.ParishID != uuid.Nil && p.UserID != uuid.Nil {
		o, err := s.orgs.ListUserOrgs(ctx, p.ParishID, p.UserID)
		if err == nil {
			orgs = o
		}
	}
	pinnedRows, err := s.repo.ListPinnedMemories(ctx, p.ParishID, p.UserID, domain.MaxPinnedMemories)
	if err != nil {
		return "", err
	}
	pinned := make([]domain.Memory, len(pinnedRows))
	for i, r := range pinnedRows {
		pinned[i] = domain.Memory{
			ID:              r.ID,
			ParishID:        r.ParishID,
			UserID:          r.UserID,
			Scope:           domain.MemoryScope(r.Scope),
			Category:        domain.MemoryCategory(r.Category),
			Content:         r.Content,
			IsPinned:        r.IsPinned,
			SourceSessionID: r.SourceSessionID,
			CreatedAt:       r.CreatedAt,
			UpdatedAt:       r.UpdatedAt,
		}
	}

	return BuildSystemPrompt(PromptInput{
		Instruction:    baseInstruction,
		Channel:        ch,
		Now:            s.nowFn(),
		Parish:         parishInfo,
		Orgs:           orgs,
		User:           userProf,
		PinnedMemories: pinned,
	}), nil
}

func (s *Service) runReActLoop(
	ctx context.Context,
	p authx.Principal,
	ch domain.Channel,
	sessionID uuid.UUID,
	agentCfg domain.AgentConfig,
	systemPrompt string,
	messages []driver.ChatMessage,
	mcpSession *mcp.ClientSession,
	onStep StepCallback,
	emitDeltas bool,
) (string, error) {
	maxSteps := agentCfg.MaxSteps
	if maxSteps <= 0 {
		maxSteps = domain.MaxSteps
	}

	toolDefs := internalToolDefinitions(agentCfg.Tools)
	internalSet := make(map[string]struct{}, len(toolDefs))
	for _, td := range toolDefs {
		internalSet[td.Name] = struct{}{}
	}
	mcpDefs, mcpSet, err := discoverMCPTools(ctx, mcpSession, agentCfg.Tools)
	if err != nil {
		return "", err
	}
	toolDefs = append(toolDefs, mcpDefs...)

	history := make([]driver.ChatMessage, len(messages))
	copy(history, messages)

	for range maxSteps {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		resp, err := s.drivers.StreamChat(ctx, agentCfg.Model, driver.ChatRequest{
			SystemPrompt: systemPrompt,
			Messages:     history,
			Tools:        toolDefs,
		}, func(ev driver.StreamEvent) error {
			switch ev.Type {
			case driver.StreamEventThinking:
				return emitStep(onStep, domain.AgentSSEEvent{
					Event:     "thinking",
					SessionID: &sessionID,
					AgentName: agentCfg.Name,
					Content:   ev.Delta,
				})
			case driver.StreamEventDelta:
				if !emitDeltas {
					return nil
				}
				return emitStep(onStep, domain.AgentSSEEvent{
					Event:     "delta",
					SessionID: &sessionID,
					AgentName: agentCfg.Name,
					Content:   ev.Delta,
				})
			case driver.StreamEventToolCall:
				if ev.ToolCall == nil {
					return nil
				}
				return emitStep(onStep, domain.AgentSSEEvent{
					Event:      "tool_call",
					SessionID:  &sessionID,
					AgentName:  agentCfg.Name,
					ToolCallID: ev.ToolCall.ID,
					ToolName:   ev.ToolCall.Name,
					Arguments:  ev.ToolCall.Arguments,
				})
			}
			return nil
		})
		if err != nil {
			return "", err
		}

		var thinkingPtr *string
		if resp.Thinking != "" {
			t := resp.Thinking
			thinkingPtr = &t
		}
		if _, err := s.repo.AppendMessage(ctx, domain.CreateMessageInput{
			ParishID:  p.ParishID,
			UserID:    p.UserID,
			SessionID: sessionID,
			AgentName: agentCfg.Name,
			Role:      domain.RoleAssistant,
			Content:   resp.Content,
			Thinking:  thinkingPtr,
			ToolCalls: resp.ToolCalls,
		}); err != nil {
			return "", err
		}

		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		history = append(history, driver.ChatMessage{
			Role:      domain.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			var resultStr string
			var toolErr error
			if _, isInternal := internalSet[tc.Name]; isInternal {
				resultStr, toolErr = s.executeInternalTool(ctx, p, ch, sessionID, mcpSession, onStep, tc.Name, tc.Arguments)
			} else if _, isMCP := mcpSet[tc.Name]; isMCP {
				resultStr, toolErr = callMCPTool(ctx, mcpSession, tc.Name, tc.Arguments)
			} else {
				toolErr = errx.Invalid("TOOL_NOT_ALLOWED", "Tool is not allowed for this agent.")
			}

			if toolErr != nil {
				if errors.Is(toolErr, domain.ErrProviderUnavailable) || errors.Is(toolErr, domain.ErrMaxStepsExceeded) || ctx.Err() != nil {
					return "", toolErr
				}
				resultStr = formatToolErrorJSON(toolErr)
			}
			resultStr = truncateToolResult(resultStr)

			tcID := tc.ID
			tcName := tc.Name
			if _, err := s.repo.AppendMessage(ctx, domain.CreateMessageInput{
				ParishID:   p.ParishID,
				UserID:     p.UserID,
				SessionID:  sessionID,
				AgentName:  agentCfg.Name,
				Role:       domain.RoleTool,
				Content:    resultStr,
				ToolCallID: &tcID,
				ToolName:   &tcName,
			}); err != nil {
				return "", err
			}

			if err := emitStep(onStep, domain.AgentSSEEvent{
				Event:      "tool_result",
				SessionID:  &sessionID,
				AgentName:  agentCfg.Name,
				ToolCallID: tc.ID,
				ToolName:   tc.Name,
				Content:    resultStr,
			}); err != nil {
				return "", err
			}

			history = append(history, driver.ChatMessage{
				Role:       domain.RoleTool,
				Content:    resultStr,
				ToolCallID: tc.ID,
				ToolName:   tc.Name,
			})
		}
	}

	return "", domain.ErrMaxStepsExceeded
}

func formatToolErrorJSON(err error) string {
	code := "TOOL_ERROR"
	msg := err.Error()
	if de, ok := errors.AsType[*errx.Error](err); ok {
		if de.Code != "" {
			code = de.Code
		}
		if de.Message != "" {
			msg = de.Message
		}
	}
	b, _ := json.Marshal(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
	return string(b)
}

func emitStep(cb StepCallback, ev domain.AgentSSEEvent) error {
	if cb == nil {
		return nil
	}
	return cb(ev)
}

func truncateRunes(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max])
}

// --- Session, Memory, and Provider Model CRUD/Search methods for HTTP handlers ---

func (s *Service) SearchSessions(ctx context.Context, in domain.SearchSessionsInput) ([]*ent.AgentSession, error) {
	return s.repo.SearchSessions(ctx, in)
}

func (s *Service) CreateSession(ctx context.Context, in domain.CreateSessionInput) (*ent.AgentSession, error) {
	return s.repo.CreateSession(ctx, in)
}

func (s *Service) GetSession(ctx context.Context, parishID, userID, sessionID uuid.UUID) (*ent.AgentSession, error) {
	return s.repo.FindSessionByID(ctx, parishID, userID, sessionID)
}

func (s *Service) UpdateSessionTitle(ctx context.Context, parishID, userID, sessionID uuid.UUID, title string) (*ent.AgentSession, error) {
	return s.repo.UpdateSessionTitle(ctx, parishID, userID, sessionID, title)
}

func (s *Service) DeleteSession(ctx context.Context, parishID, userID, sessionID uuid.UUID) error {
	return s.repo.SoftDeleteSession(ctx, parishID, userID, sessionID)
}

func (s *Service) SearchSessionMessages(ctx context.Context, in domain.SearchSessionMessagesInput) ([]*ent.AgentMessage, error) {
	return s.repo.SearchSessionMessages(ctx, in)
}

func (s *Service) SearchMemories(ctx context.Context, in domain.SearchMemoriesInput) ([]*ent.AgentMemory, error) {
	return s.repo.SearchMemories(ctx, in)
}

func (s *Service) CreateMemory(ctx context.Context, p authx.Principal, req domain.CreateAgentMemoryRequest) (*ent.AgentMemory, error) {
	scope := domain.MemoryScope(req.Scope)
	if scope == domain.MemoryScopeParish {
		ok, err := s.isParishAdmin(p)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	}
	var userPtr *uuid.UUID
	if scope == domain.MemoryScopeUser {
		uid := p.UserID
		userPtr = &uid
	}
	pinned := req.IsPinned != nil && *req.IsPinned
	content := strings.TrimSpace(req.Content)
	return s.repo.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID:    p.ParishID,
		UserID:      userPtr,
		Scope:       scope,
		Category:    domain.MemoryCategory(req.Category),
		Content:     content,
		IsPinned:    pinned,
		SupersedeID: req.SupersedeID,
		Embedding:   s.bestEffortEmbed(ctx, content),
	})
}

func (s *Service) UpdateMemory(ctx context.Context, p authx.Principal, id uuid.UUID, req domain.UpdateAgentMemoryRequest) (*ent.AgentMemory, error) {
	row, err := s.repo.FindMemoryByID(ctx, p.ParishID, id)
	if err != nil {
		return nil, err
	}
	if row.Scope == "parish" {
		ok, err := s.isParishAdmin(p)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	}
	var catPtr *domain.MemoryCategory
	if req.Category != nil {
		c := domain.MemoryCategory(*req.Category)
		catPtr = &c
	}
	var emb []float32
	if req.Content != nil {
		emb = s.bestEffortEmbed(ctx, strings.TrimSpace(*req.Content))
	}
	return s.repo.UpdateMemory(ctx, domain.UpdateMemoryInput{
		ID:        id,
		ParishID:  p.ParishID,
		UserID:    p.UserID,
		Category:  catPtr,
		Content:   req.Content,
		IsPinned:  req.IsPinned,
		Embedding: emb,
	})
}

func (s *Service) DeleteMemory(ctx context.Context, p authx.Principal, id uuid.UUID) error {
	row, err := s.repo.FindMemoryByID(ctx, p.ParishID, id)
	if err != nil {
		return err
	}
	if row.Scope == "parish" {
		ok, err := s.isParishAdmin(p)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrForbidden
		}
	}
	return s.repo.SoftDeleteMemory(ctx, p.ParishID, p.UserID, id)
}
