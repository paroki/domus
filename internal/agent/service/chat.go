package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/parish"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/mcpx"
)

const (
	maxChatMessageRunes = 16000
	maxSessionTitleRune = 60

	settingModelMikael  = "agent.model.mikael"
	settingModelRafael  = "agent.model.rafael"
	settingPromptGab    = "agent.prompt.gabriel"
	settingPromptMik    = "agent.prompt.mikael"
	settingPromptRaf    = "agent.prompt.rafael"
	chatTurnTimeout     = 3 * time.Minute
	chatKeepalivePeriod = 15 * time.Second
)

type resolvedModelConfig struct {
	provider *ent.AgentProvider
	plainKey string
	modelID  string
}

// ChatTurnPrep holds all validated state and resolved dependencies prepared before opening SSE.
type ChatTurnPrep struct {
	Principal       authx.Principal
	SessionID       *uuid.UUID
	LockedSessionID uuid.UUID
	SessionTitle    string
	Message         string
	GabrielModel    resolvedModelConfig
	MikaelModel     resolvedModelConfig
	RafaelModel     resolvedModelConfig
	GabrielPrompt   string
	MikaelPrompt    string
	RafaelPrompt    string
	PromptContext   PromptContext
	History         []*schema.Message
}

func (s *Service) tryLockSession(id uuid.UUID) bool {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	if s.busySessions[id] {
		return false
	}
	s.busySessions[id] = true
	return true
}

func (s *Service) unlockSession(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	s.sessionMu.Lock()
	delete(s.busySessions, id)
	s.sessionMu.Unlock()
}

// PrepareChat validates input, resolves models, loads context/history, and acquires the session busy lock.
// Everything here runs synchronously before opening the SSE stream.
func (s *Service) PrepareChat(ctx context.Context, p authx.Principal, req domain.ChatRequest) (*ChatTurnPrep, error) {
	msg := strings.TrimSpace(req.Message)
	n := utf8.RuneCountInString(msg)
	if n == 0 {
		return nil, domain.ValidationError("message", "REQUIRED", "Message is required.")
	}
	if n > maxChatMessageRunes {
		return nil, domain.ValidationError("message", "MAX_LENGTH", "Message must be at most 16000 characters.")
	}

	var existingSess *domain.SessionResponse
	var lockedID uuid.UUID
	if req.SessionID != nil {
		sess, err := s.repo.GetSession(ctx, p.ParishID, p.UserID, *req.SessionID)
		if err != nil {
			return nil, err
		}
		if !s.tryLockSession(sess.ID) {
			return nil, domain.ErrSessionBusy
		}
		lockedID = sess.ID
		existingSess = &sess
	}

	releaseOnError := func() {
		if lockedID != uuid.Nil {
			s.unlockSession(lockedID)
		}
	}

	// Resolve Gabriel model (session override if valid, otherwise agent.model.gabriel)
	var gabrielCfg resolvedModelConfig
	if existingSess != nil && existingSess.ProviderID != nil && existingSess.ModelID != nil && strings.TrimSpace(*existingSess.ModelID) != "" {
		if prov, key, err := s.loadProviderWithKey(ctx, *existingSess.ProviderID); err == nil {
			gabrielCfg = resolvedModelConfig{provider: prov, plainKey: key, modelID: strings.TrimSpace(*existingSess.ModelID)}
		}
	}
	if gabrielCfg.provider == nil {
		prov, key, mid, err := s.resolveSettingModel(ctx, settingModelGabriel, false)
		if err != nil {
			releaseOnError()
			return nil, err
		}
		gabrielCfg = resolvedModelConfig{provider: prov, plainKey: key, modelID: mid}
	}

	mikaelCfg := gabrielCfg
	if prov, key, mid, err := s.resolveSettingModel(ctx, settingModelMikael, true); err == nil {
		mikaelCfg = resolvedModelConfig{provider: prov, plainKey: key, modelID: mid}
	}
	rafaelCfg := gabrielCfg
	if prov, key, mid, err := s.resolveSettingModel(ctx, settingModelRafael, true); err == nil {
		rafaelCfg = resolvedModelConfig{provider: prov, plainKey: key, modelID: mid}
	}

	gabPrompt := s.resolvePrompt(settingPromptGab, prompts.Gabriel)
	mikPrompt := s.resolvePrompt(settingPromptMik, prompts.Mikael)
	rafPrompt := s.resolvePrompt(settingPromptRaf, prompts.Rafael)

	var userName, parishName string
	if s.db != nil {
		if u, err := s.db.User.Query().Where(entuser.IDEQ(p.UserID)).Only(ctx); err == nil {
			userName = u.Name
		}
		if pr, err := s.db.Parish.Query().Where(parish.IDEQ(p.ParishID)).Only(ctx); err == nil {
			parishName = pr.Name
		}
	}

	memories, _, _, _ := s.SearchMemories(ctx, p, domain.SearchMemoriesRequest{Q: msg, Limit: 5})

	var history []*schema.Message
	var title string
	if existingSess != nil {
		title = existingSess.Title
		rows, err := s.repo.ListMessages(ctx, p.ParishID, p.UserID, existingSess.ID, 0)
		if err != nil {
			releaseOnError()
			return nil, err
		}
		history = ProjectHistory(rows)
	} else {
		title = cleanTitle(msg)
	}

	return &ChatTurnPrep{
		Principal:       p,
		SessionID:       req.SessionID,
		LockedSessionID: lockedID,
		SessionTitle:    title,
		Message:         msg,
		GabrielModel:    gabrielCfg,
		MikaelModel:     mikaelCfg,
		RafaelModel:     rafaelCfg,
		GabrielPrompt:   gabPrompt,
		MikaelPrompt:    mikPrompt,
		RafaelPrompt:    rafPrompt,
		PromptContext: PromptContext{
			UserName:          userName,
			ParishName:        parishName,
			CurrentTime:       time.Now().UTC().Format(time.RFC3339),
			RetrievedMemories: memories,
		},
		History: history,
	}, nil
}

// ExecuteChatStream runs the multi-agent turn inside Fiber's SendStreamWriter callback.
// It never touches fiber.Ctx and releases the session busy lock only after persistence completes.
func (s *Service) ExecuteChatStream(prep *ChatTurnPrep, writeEvent func(StreamEvent) error) {
	defer func() {
		if prep.LockedSessionID != uuid.Nil {
			s.unlockSession(prep.LockedSessionID)
		}
	}()

	bgCtx := context.Background()
	isNew := false
	var sessionID uuid.UUID
	if prep.SessionID == nil {
		created, err := s.repo.CreateSession(bgCtx, prep.Principal.ParishID, prep.Principal.UserID, prep.SessionTitle, nil, nil)
		if err != nil {
			_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "INTERNAL_ERROR", "message": err.Error()}})
			return
		}
		sessionID = created.ID
		isNew = true
		s.tryLockSession(sessionID)
		prep.LockedSessionID = sessionID
	} else {
		sessionID = *prep.SessionID
	}

	if _, err := s.repo.AppendMessages(bgCtx, prep.Principal.ParishID, prep.Principal.UserID, sessionID, []domain.AppendMessageInput{
		{Role: "user", Content: prep.Message},
	}); err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	if err := writeEvent(StreamEvent{
		Event: "session",
		Data: map[string]any{
			"sessionId": sessionID.String(),
			"title":     prep.SessionTitle,
			"isNew":     isNew,
		},
	}); err != nil {
		return
	}

	turnCtx, turnCancel := context.WithTimeout(authx.WithPrincipal(bgCtx, prep.Principal), chatTurnTimeout)
	defer turnCancel()

	var mcpSess *mcp.ClientSession
	var gabrielMCPTools, mikaelTools, rafaelTools []tool.BaseTool
	if s.mcpServer != nil {
		var err error
		mcpSess, err = mcpx.ConnectInMemory(turnCtx, s.mcpServer)
		if err != nil {
			_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "MCP_CONNECT_FAILED", "message": err.Error()}})
			return
		}
		defer func() { _ = mcpSess.Close() }()

		gabrielMCPTools, mikaelTools, rafaelTools, err = LoadMCPTools(turnCtx, mcpSess)
		if err != nil {
			_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "MCP_TOOLS_FAILED", "message": err.Error()}})
			return
		}
	}

	memTools, err := NewMemoryTools(s, prep.Principal)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "MEMORY_TOOLS_FAILED", "message": err.Error()}})
		return
	}
	gabrielTools := append(append([]tool.BaseTool(nil), gabrielMCPTools...), memTools...)

	gabModel, err := s.drivers.NewChatModel(turnCtx, string(prep.GabrielModel.provider.Provider), prep.GabrielModel.provider.BaseURL, prep.GabrielModel.plainKey, prep.GabrielModel.modelID)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "AGENT_PROVIDER_UNAVAILABLE", "message": err.Error()}})
		return
	}
	mikModel, err := s.drivers.NewChatModel(turnCtx, string(prep.MikaelModel.provider.Provider), prep.MikaelModel.provider.BaseURL, prep.MikaelModel.plainKey, prep.MikaelModel.modelID)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "AGENT_PROVIDER_UNAVAILABLE", "message": err.Error()}})
		return
	}
	rafModel, err := s.drivers.NewChatModel(turnCtx, string(prep.RafaelModel.provider.Provider), prep.RafaelModel.provider.BaseURL, prep.RafaelModel.plainKey, prep.RafaelModel.modelID)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "AGENT_PROVIDER_UNAVAILABLE", "message": err.Error()}})
		return
	}

	eventCh := make(chan StreamEvent, 64)
	emitAsync := func(ev StreamEvent) {
		select {
		case eventCh <- ev:
		case <-turnCtx.Done():
		}
	}

	runner, err := buildSupervisorRunner(turnCtx, supervisorConfig{
		GabrielModel:  gabModel,
		MikaelModel:   mikModel,
		RafaelModel:   rafModel,
		GabrielPrompt: prep.GabrielPrompt,
		MikaelPrompt:  prep.MikaelPrompt,
		RafaelPrompt:  prep.RafaelPrompt,
		PromptContext: prep.PromptContext,
		GabrielTools:  gabrielTools,
		MikaelTools:   mikaelTools,
		RafaelTools:   rafaelTools,
		Emit:          emitAsync,
	})
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "SUPERVISOR_INIT_FAILED", "message": err.Error()}})
		return
	}

	inputMsgs := append(append([]*schema.Message(nil), prep.History...), schema.UserMessage(prep.Message))
	cancelOpt, cancelFn := adk.WithCancel()
	iter := runner.Run(turnCtx, inputMsgs, cancelOpt)

	type runnerResult struct {
		completedTools []domain.AppendMessageInput
		finalText      string
		finalAgent     string
		err            error
	}
	doneCh := make(chan runnerResult, 1)

	go func() {
		var res runnerResult
		res.finalAgent = "gabriel"
		var finalBuilder strings.Builder
		for {
			ev, ok := iter.Next()
			if !ok {
				break
			}
			if ev.Err != nil {
				res.err = ev.Err
				break
			}
			if ev.Output != nil && ev.Output.MessageOutput != nil {
				mv := ev.Output.MessageOutput
				text, tcs, err := consumeStreamMessage(mv, ev.AgentName, emitAsync)
				if err != nil {
					res.err = err
					break
				}
				switch mv.Role {
				case schema.Assistant:
					if len(tcs) > 0 {
						records := make([]domain.ToolCallRecord, 0, len(tcs))
						for _, tc := range tcs {
							records = append(records, domain.ToolCallRecord{
								ID:        tc.ID,
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							})
						}
						res.completedTools = append(res.completedTools, domain.AppendMessageInput{
							AgentName: ev.AgentName,
							Role:      "assistant",
							Content:   text,
							ToolCalls: records,
						})
					} else if text != "" {
						finalBuilder.WriteString(text)
						if ev.AgentName != "" {
							res.finalAgent = ev.AgentName
						}
					}
				case schema.Tool:
					var tcID, tName *string
					if mv.Message != nil && mv.Message.ToolCallID != "" {
						v := mv.Message.ToolCallID
						tcID = &v
					}
					if mv.ToolName != "" {
						v := mv.ToolName
						tName = &v
					}
					res.completedTools = append(res.completedTools, domain.AppendMessageInput{
						AgentName:  ev.AgentName,
						Role:       "tool",
						Content:    text,
						ToolCallID: tcID,
						ToolName:   tName,
					})
				}
			}
		}
		res.finalText = finalBuilder.String()
		doneCh <- res
	}()

	ticker := time.NewTicker(chatKeepalivePeriod)
	defer ticker.Stop()

	clientDisconnected := false
	var res runnerResult

loop:
	for {
		select {
		case ev := <-eventCh:
			if !clientDisconnected {
				if err := writeEvent(ev); err != nil {
					clientDisconnected = true
					cancelFn(
						adk.WithAgentCancelMode(adk.CancelAfterToolCalls),
						adk.WithAgentCancelTimeout(3*time.Second),
						adk.WithRecursive(),
					)
					turnCancel()
				}
			}
		case <-ticker.C:
			if !clientDisconnected {
				if err := writeEvent(StreamEvent{Event: "ping"}); err != nil {
					clientDisconnected = true
					cancelFn(
						adk.WithAgentCancelMode(adk.CancelAfterToolCalls),
						adk.WithAgentCancelTimeout(3*time.Second),
						adk.WithRecursive(),
					)
					turnCancel()
				}
			}
		case res = <-doneCh:
			// Drain any buffered events before exiting
			for {
				select {
				case ev := <-eventCh:
					if !clientDisconnected {
						if err := writeEvent(ev); err != nil {
							clientDisconnected = true
						}
					}
				default:
					break loop
				}
			}
		}
	}

	persistCtx, persistCancel := context.WithTimeout(bgCtx, 10*time.Second)
	defer persistCancel()

	var cancelErr *adk.CancelError
	if clientDisconnected || errors.As(res.err, &cancelErr) {
		// Persist any completed assistant(tool_calls) + tool pairs; discard partial incomplete assistant text.
		if len(res.completedTools) > 0 {
			_, _ = s.repo.AppendMessages(persistCtx, prep.Principal.ParishID, prep.Principal.UserID, sessionID, res.completedTools)
		}
		return
	}

	if res.err != nil {
		if len(res.completedTools) > 0 {
			_, _ = s.repo.AppendMessages(persistCtx, prep.Principal.ParishID, prep.Principal.UserID, sessionID, res.completedTools)
		}
		_ = writeEvent(StreamEvent{
			Event: "error",
			Data: map[string]string{
				"code":    "AGENT_STREAM_ERROR",
				"message": res.err.Error(),
			},
		})
		return
	}

	toPersist := append([]domain.AppendMessageInput(nil), res.completedTools...)
	toPersist = append(toPersist, domain.AppendMessageInput{
		AgentName: res.finalAgent,
		Role:      "assistant",
		Content:   res.finalText,
	})
	persisted, err := s.repo.AppendMessages(persistCtx, prep.Principal.ParishID, prep.Principal.UserID, sessionID, toPersist)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "INTERNAL_ERROR", "message": err.Error()}})
		return
	}

	var lastMsgID string
	if len(persisted) > 0 {
		lastMsgID = persisted[len(persisted)-1].ID.String()
	}
	_ = writeEvent(StreamEvent{
		Event: "done",
		Data: map[string]string{
			"sessionId": sessionID.String(),
			"messageId": lastMsgID,
			"agent":     res.finalAgent,
		},
	})

	s.dispatchMemoryExtraction(prep.Principal, sessionID, prep.Message, res.finalText)
}

func (s *Service) resolvePrompt(key, fallback string) string {
	if s.settings != nil {
		if custom := strings.TrimSpace(s.settings.Get(key)); custom != "" {
			return custom
		}
	}
	return strings.TrimSpace(fallback)
}

func cleanTitle(msg string) string {
	cleaned := strings.Join(strings.Fields(msg), " ")
	r := []rune(cleaned)
	if len(r) <= maxSessionTitleRune {
		return cleaned
	}
	return string(r[:maxSessionTitleRune]) + "…"
}
