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
	"github.com/paroki/domus/ent/cantus"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/keluarga"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/paroki"
	"github.com/paroki/domus/ent/umat"
	entuser "github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/mcpx"
)

const (
	maxChatMessageRunes = 16000
	maxSessionTitleRune = 60

	settingModelMikael    = "agent.model.mikael"
	settingModelRafael    = "agent.model.rafael"
	settingModelYosef     = "agent.model.yosef"
	settingModelGregorius = "agent.model.gregorius"
	settingModelSesilia   = "agent.model.sesilia"
	settingPromptGab      = "agent.prompt.gabriel"
	settingPromptMik      = "agent.prompt.mikael"
	settingPromptRaf      = "agent.prompt.rafael"
	settingPromptYos      = "agent.prompt.yosef"
	settingPromptGreg     = "agent.prompt.gregorius"
	settingPromptSes      = "agent.prompt.sesilia"
	chatTurnTimeout       = 3 * time.Minute
	chatKeepalivePeriod   = 15 * time.Second
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
	YosefModel      resolvedModelConfig
	GregoriusModel  resolvedModelConfig
	SesiliaModel    resolvedModelConfig
	GabrielPrompt   string
	MikaelPrompt    string
	RafaelPrompt    string
	YosefPrompt     string
	GregoriusPrompt string
	SesiliaPrompt   string
	PromptContext   PromptContext
	History         []*schema.Message
	SkipAppendUser  bool
}

type sessionLock struct {
	cancelling bool
	done       chan struct{}
}

func (s *Service) acquireSessionLock(ctx context.Context, id uuid.UUID) bool {
	for {
		s.sessionMu.Lock()
		lock, busy := s.busySessions[id]
		if !busy {
			s.busySessions[id] = &sessionLock{done: make(chan struct{})}
			s.sessionMu.Unlock()
			return true
		}
		if !lock.cancelling {
			s.sessionMu.Unlock()
			return false
		}
		done := lock.done
		s.sessionMu.Unlock()

		select {
		case <-done:
		case <-ctx.Done():
			return false
		}
	}
}

func (s *Service) tryLockSession(id uuid.UUID) bool {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	if _, busy := s.busySessions[id]; busy {
		return false
	}
	s.busySessions[id] = &sessionLock{done: make(chan struct{})}
	return true
}

func (s *Service) markSessionCancelling(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	s.sessionMu.Lock()
	if lock, ok := s.busySessions[id]; ok {
		lock.cancelling = true
	}
	s.sessionMu.Unlock()
}

func (s *Service) unlockSession(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	s.sessionMu.Lock()
	if lock, ok := s.busySessions[id]; ok {
		delete(s.busySessions, id)
		close(lock.done)
	}
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
		sess, err := s.repo.GetSession(ctx, p.ParokiID, p.UserID, *req.SessionID)
		if err != nil {
			return nil, err
		}
		if !s.acquireSessionLock(ctx, sess.ID) {
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

	resolveOptionalModel := func(key string) resolvedModelConfig {
		if prov, keyStr, mid, err := s.resolveSettingModel(ctx, key, true); err == nil {
			return resolvedModelConfig{provider: prov, plainKey: keyStr, modelID: mid}
		}
		return gabrielCfg
	}
	mikaelCfg := resolveOptionalModel(settingModelMikael)
	rafaelCfg := resolveOptionalModel(settingModelRafael)
	yosefCfg := resolveOptionalModel(settingModelYosef)
	gregoriusCfg := resolveOptionalModel(settingModelGregorius)
	sesiliaCfg := resolveOptionalModel(settingModelSesilia)

	gabPrompt := s.resolvePrompt(settingPromptGab, prompts.Gabriel)
	mikPrompt := s.resolvePrompt(settingPromptMik, prompts.Mikael)
	rafPrompt := s.resolvePrompt(settingPromptRaf, prompts.Rafael)
	yosPrompt := s.resolvePrompt(settingPromptYos, prompts.Yosef)
	gregPrompt := s.resolvePrompt(settingPromptGreg, prompts.Gregorius)
	sesPrompt := s.resolvePrompt(settingPromptSes, prompts.Sesilia)

	var userName, parokiName, userTZ, parokiTZ string
	if s.db != nil {
		if u, err := s.db.User.Query().Where(entuser.IDEQ(p.UserID)).Only(ctx); err == nil {
			userName = u.Name
			userTZ = strings.TrimSpace(u.Timezone)
		}
		if pr, err := s.db.Paroki.Query().Where(paroki.IDEQ(p.ParokiID)).Only(ctx); err == nil {
			parokiName = pr.Nama
			parokiTZ = strings.TrimSpace(pr.Timezone)
		}
	}
	parokiLoc, resolvedParokiTZ := loadTimezoneOrDefault(parokiTZ, defaultParokiTimezone)
	if userTZ == "" {
		userTZ = resolvedParokiTZ
	}

	var pagePath, pageTitle, activeEntity string
	if req.PageContext != nil {
		pagePath = strings.TrimSpace(req.PageContext.Path)
		pageTitle = strings.TrimSpace(req.PageContext.Title)
		if pagePath != "" {
			activeEntity = s.resolvePageContextEntity(ctx, p.ParokiID, pagePath)
		}
	}

	memories, _, _, _ := s.SearchMemories(ctx, p, domain.SearchMemoriesRequest{Q: msg, Limit: 5})

	var history []*schema.Message
	var title string
	skipAppendUser := false
	if existingSess != nil {
		title = existingSess.Title
		rows, err := s.repo.ListMessages(ctx, p.ParokiID, p.UserID, existingSess.ID, 0)
		if err != nil {
			releaseOnError()
			return nil, err
		}
		if req.Retry && len(rows) > 0 && rows[len(rows)-1].Role == "user" && strings.TrimSpace(rows[len(rows)-1].Content) == msg {
			skipAppendUser = true
			rows = rows[:len(rows)-1]
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
		YosefModel:      yosefCfg,
		GregoriusModel:  gregoriusCfg,
		SesiliaModel:    sesiliaCfg,
		GabrielPrompt:   gabPrompt,
		MikaelPrompt:    mikPrompt,
		RafaelPrompt:    rafPrompt,
		YosefPrompt:     yosPrompt,
		GregoriusPrompt: gregPrompt,
		SesiliaPrompt:   sesPrompt,
		PromptContext: PromptContext{
			UserName:          userName,
			ParokiName:        parokiName,
			UserRoles:         p.Roles,
			UserTimezone:      userTZ,
			ParokiTimezone:    resolvedParokiTZ,
			CurrentTime:       time.Now().In(parokiLoc).Format("Monday, 2006-01-02T15:04:05-07:00"),
			PagePath:          pagePath,
			PageTitle:         pageTitle,
			ActiveEntity:      activeEntity,
			RetrievedMemories: memories,
		},
		History:        history,
		SkipAppendUser: skipAppendUser,
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
		created, err := s.repo.CreateSession(bgCtx, prep.Principal.ParokiID, prep.Principal.UserID, prep.SessionTitle, nil, nil)
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

	if !prep.SkipAppendUser {
		if _, err := s.repo.AppendMessages(bgCtx, prep.Principal.ParokiID, prep.Principal.UserID, sessionID, []domain.AppendMessageInput{
			{Role: "user", Content: prep.Message},
		}); err != nil {
			_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "INTERNAL_ERROR", "message": err.Error()}})
			return
		}
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
	var mcpTools MCPToolSet
	if s.mcpServer != nil {
		var err error
		mcpSess, err = mcpx.ConnectInMemory(turnCtx, s.mcpServer)
		if err != nil {
			_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "MCP_CONNECT_FAILED", "message": err.Error()}})
			return
		}
		defer func() { _ = mcpSess.Close() }()

		mcpTools, err = LoadMCPTools(turnCtx, mcpSess)
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

	eventCh := make(chan StreamEvent, 64)
	emitAsync := func(ev StreamEvent) {
		select {
		case eventCh <- ev:
		case <-turnCtx.Done():
		}
	}

	recorder := &TurnMetaRecorder{}
	uiTools, err := NewUITools(emitAsync, recorder)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "UI_TOOLS_FAILED", "message": err.Error()}})
		return
	}

	appendTools := func(groups ...[]tool.BaseTool) []tool.BaseTool {
		var out []tool.BaseTool
		for _, g := range groups {
			out = append(out, g...)
		}
		return out
	}

	gabrielTools := appendTools(memTools, uiTools)
	mikaelTools := appendTools(mcpTools.Mikael, uiTools)
	rafaelTools := appendTools(mcpTools.Rafael, uiTools)
	yosefTools := appendTools(mcpTools.Yosef, uiTools)
	gregoriusTools := appendTools(mcpTools.Gregorius, uiTools)
	sesiliaTools := appendTools(mcpTools.Sesilia, uiTools)

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
	yosModel, err := s.drivers.NewChatModel(turnCtx, string(prep.YosefModel.provider.Provider), prep.YosefModel.provider.BaseURL, prep.YosefModel.plainKey, prep.YosefModel.modelID)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "AGENT_PROVIDER_UNAVAILABLE", "message": err.Error()}})
		return
	}
	gregModel, err := s.drivers.NewChatModel(turnCtx, string(prep.GregoriusModel.provider.Provider), prep.GregoriusModel.provider.BaseURL, prep.GregoriusModel.plainKey, prep.GregoriusModel.modelID)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "AGENT_PROVIDER_UNAVAILABLE", "message": err.Error()}})
		return
	}
	sesModel, err := s.drivers.NewChatModel(turnCtx, string(prep.SesiliaModel.provider.Provider), prep.SesiliaModel.provider.BaseURL, prep.SesiliaModel.plainKey, prep.SesiliaModel.modelID)
	if err != nil {
		_ = writeEvent(StreamEvent{Event: "error", Data: map[string]string{"code": "AGENT_PROVIDER_UNAVAILABLE", "message": err.Error()}})
		return
	}

	runner, err := buildSupervisorRunner(turnCtx, supervisorConfig{
		GabrielModel:    gabModel,
		MikaelModel:     mikModel,
		RafaelModel:     rafModel,
		YosefModel:      yosModel,
		GregoriusModel:  gregModel,
		SesiliaModel:    sesModel,
		GabrielPrompt:   prep.GabrielPrompt,
		MikaelPrompt:    prep.MikaelPrompt,
		RafaelPrompt:    prep.RafaelPrompt,
		YosefPrompt:     prep.YosefPrompt,
		GregoriusPrompt: prep.GregoriusPrompt,
		SesiliaPrompt:   prep.SesiliaPrompt,
		PromptContext:   prep.PromptContext,
		GabrielTools:    gabrielTools,
		MikaelTools:     mikaelTools,
		RafaelTools:     rafaelTools,
		YosefTools:      yosefTools,
		GregoriusTools:  gregoriusTools,
		SesiliaTools:    sesiliaTools,
		Emit:            emitAsync,
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
				msg, err := consumeStreamMessage(mv, ev.AgentName, emitAsync)
				if err != nil {
					res.err = err
					break
				}
				if msg == nil {
					continue
				}
				switch mv.Role {
				case schema.Assistant:
					if len(msg.ToolCalls) > 0 {
						records := make([]domain.ToolCallRecord, 0, len(msg.ToolCalls))
						for _, tc := range msg.ToolCalls {
							records = append(records, domain.ToolCallRecord{
								ID:        tc.ID,
								Name:      tc.Function.Name,
								Arguments: tc.Function.Arguments,
							})
						}
						res.completedTools = append(res.completedTools, domain.AppendMessageInput{
							AgentName: ev.AgentName,
							Role:      "assistant",
							Content:   msg.Content,
							ToolCalls: records,
						})
					} else if msg.Content != "" && (ev.AgentName == "" || ev.AgentName == "gabriel") {
						finalBuilder.WriteString(msg.Content)
						res.finalAgent = "gabriel"
					}
				case schema.Tool:
					var tcID, tName *string
					if msg.ToolCallID != "" {
						v := msg.ToolCallID
						tcID = &v
					}
					if mv.ToolName != "" {
						v := mv.ToolName
						tName = &v
					}
					res.completedTools = append(res.completedTools, domain.AppendMessageInput{
						AgentName:  ev.AgentName,
						Role:       "tool",
						Content:    msg.Content,
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
	var cancelFallback *time.Timer
	defer func() {
		if cancelFallback != nil {
			cancelFallback.Stop()
		}
	}()
	onDisconnect := func() {
		clientDisconnected = true
		s.markSessionCancelling(sessionID)
		cancelFn(
			adk.WithAgentCancelMode(adk.CancelAfterToolCalls),
			adk.WithAgentCancelTimeout(3*time.Second),
			adk.WithRecursive(),
		)
		cancelFallback = time.AfterFunc(4*time.Second, turnCancel)
	}
	var res runnerResult

loop:
	for {
		select {
		case ev := <-eventCh:
			if !clientDisconnected {
				if err := writeEvent(ev); err != nil {
					onDisconnect()
				}
			}
		case <-ticker.C:
			if !clientDisconnected {
				if err := writeEvent(StreamEvent{Event: "ping"}); err != nil {
					onDisconnect()
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
			_, _ = s.repo.AppendMessages(persistCtx, prep.Principal.ParokiID, prep.Principal.UserID, sessionID, res.completedTools)
		}
		return
	}

	if res.err != nil {
		if len(res.completedTools) > 0 {
			_, _ = s.repo.AppendMessages(persistCtx, prep.Principal.ParokiID, prep.Principal.UserID, sessionID, res.completedTools)
		}
		wrappedErr := driver.WrapProviderError(res.err, prep.GabrielModel.plainKey)
		wrappedErr = driver.WrapProviderError(wrappedErr, prep.MikaelModel.plainKey)
		wrappedErr = driver.WrapProviderError(wrappedErr, prep.RafaelModel.plainKey)
		wrappedErr = driver.WrapProviderError(wrappedErr, prep.YosefModel.plainKey)
		wrappedErr = driver.WrapProviderError(wrappedErr, prep.GregoriusModel.plainKey)
		wrappedErr = driver.WrapProviderError(wrappedErr, prep.SesiliaModel.plainKey)
		_ = writeEvent(StreamEvent{
			Event: "error",
			Data: map[string]string{
				"code":    "AGENT_STREAM_ERROR",
				"message": wrappedErr.Error(),
			},
		})
		return
	}

	var meta map[string]any
	q, nav := recorder.Snapshot()
	if q != nil || nav != nil {
		meta = make(map[string]any)
		if q != nil {
			meta["question"] = q
		}
		if nav != nil {
			meta["navigation"] = nav
		}
	}
	if strings.TrimSpace(res.finalText) == "" && q != nil {
		res.finalText = q.Question
	}

	toPersist := append([]domain.AppendMessageInput(nil), res.completedTools...)
	toPersist = append(toPersist, domain.AppendMessageInput{
		AgentName: res.finalAgent,
		Role:      "assistant",
		Content:   res.finalText,
		Meta:      meta,
	})
	persisted, err := s.repo.AppendMessages(persistCtx, prep.Principal.ParokiID, prep.Principal.UserID, sessionID, toPersist)
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

func loadTimezoneOrDefault(tz, fallback string) (*time.Location, string) {
	if tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc, tz
		}
	}
	if loc, err := time.LoadLocation(fallback); err == nil {
		return loc, fallback
	}
	return time.UTC, "UTC"
}

func (s *Service) resolvePageContextEntity(ctx context.Context, parokiID uuid.UUID, rawPath string) string {
	if s.db == nil {
		return ""
	}
	cleanPath := rawPath
	if idx := strings.IndexAny(cleanPath, "?#"); idx >= 0 {
		cleanPath = cleanPath[:idx]
	}
	segs := strings.Split(strings.Trim(cleanPath, "/"), "/")
	if len(segs) < 2 {
		return ""
	}

	switch segs[0] {
	case "org":
		id, err := uuid.Parse(segs[1])
		if err != nil {
			return ""
		}
		row, err := s.db.Organization.Query().
			Where(organization.IDEQ(id), organization.ParokiIDEQ(parokiID), organization.DeletedAtIsNil()).
			Only(ctx)
		if err != nil {
			return ""
		}
		return "Active Organization: " + row.Name + " (orgId: " + row.ID.String() + ")"

	case "event", "liturgi":
		id, err := uuid.Parse(segs[1])
		if err != nil {
			return ""
		}
		row, err := s.db.Event.Query().
			Where(event.IDEQ(id), event.ParokiIDEQ(parokiID), event.DeletedAtIsNil()).
			Only(ctx)
		if err != nil {
			return ""
		}
		return "Active Event: " + row.Title + " (eventId: " + row.ID.String() + ")"

	case "umat":
		if segs[1] == "keluarga" {
			if len(segs) < 3 {
				return ""
			}
			id, err := uuid.Parse(segs[2])
			if err != nil {
				return ""
			}
			row, err := s.db.Keluarga.Query().
				Where(keluarga.IDEQ(id), keluarga.ParokiIDEQ(parokiID), keluarga.DeletedAtIsNil()).
				Only(ctx)
			if err != nil {
				return ""
			}
			return "Active Keluarga: " + row.NamaKeluarga + " (keluargaId: " + row.ID.String() + ")"
		}
		if segs[1] == "new" || segs[1] == "referensi" {
			return ""
		}
		id, err := uuid.Parse(segs[1])
		if err != nil {
			return ""
		}
		row, err := s.db.Umat.Query().
			Where(umat.IDEQ(id), umat.ParokiIDEQ(parokiID), umat.DeletedAtIsNil()).
			Only(ctx)
		if err != nil {
			return ""
		}
		return "Active Umat: " + row.Nama + " (umatId: " + row.ID.String() + ")"

	case "cantus":
		if segs[1] == "new" {
			return ""
		}
		id, err := uuid.Parse(segs[1])
		if err != nil {
			return ""
		}
		row, err := s.db.Cantus.Query().
			Where(cantus.IDEQ(id), cantus.DeletedAtIsNil()).
			Only(ctx)
		if err != nil {
			return ""
		}
		return "Active Cantus: " + row.Judul + " (cantusId: " + row.ID.String() + ")"
	}
	return ""
}

func cleanTitle(msg string) string {
	cleaned := strings.Join(strings.Fields(msg), " ")
	r := []rune(cleaned)
	if len(r) <= maxSessionTitleRune {
		return cleaned
	}
	return string(r[:maxSessionTitleRune]) + "…"
}
