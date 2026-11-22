package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent/agentmemory"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/errx"
)

const roleParishAdmin = "ref:parish-admin"

func internalToolDefinitions(patterns []string) []driver.ToolDefinition {
	catalog := []driver.ToolDefinition{
		{
			Name:        "memory_save",
			Description: "Simpan fakta jangka panjang tentang pengguna atau FAQ paroki ke dalam memori.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"content": {"type": "string", "description": "Fakta singkat (maks 500 karakter)"},
					"category": {"type": "string", "enum": ["profile", "preference", "ministry", "parish_faq"]},
					"is_pinned": {"type": "boolean", "description": "Pin agar selalu disuntikkan ke konteks"},
					"scope": {"type": "string", "enum": ["user", "parish"]},
					"supersede_id": {"type": "string", "description": "UUID memori lama yang digantikan"}
				},
				"required": ["content", "category"]
			}`),
		},
		{
			Name:        "memory_recall",
			Description: "Cari fakta jangka panjang yang relevan dari memori pengguna dan FAQ paroki.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {"type": "string"},
					"category": {"type": "string", "enum": ["profile", "preference", "ministry", "parish_faq"]},
					"limit": {"type": "integer", "minimum": 1, "maximum": 20}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        "memory_forget",
			Description: "Hapus (soft-delete) satu fakta memori berdasarkan UUID.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"memory_id": {"type": "string"}
				},
				"required": ["memory_id"]
			}`),
		},
		{
			Name:        "memory_search_history",
			Description: "Cari riwayat percakapan terdahulu lintas sesi dan saluran (web, telegram, whatsapp).",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {"type": "string"},
					"limit": {"type": "integer", "minimum": 1, "maximum": 20}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        "delegate_specialist",
			Description: "Delegasikan tugas spesialis ke Mikael (acara/RSVP) atau Rafael (organisasi/undangan).",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"agent_name": {"type": "string", "enum": ["mikael", "rafael"]},
					"task": {"type": "string", "description": "Instruksi tugas jelas untuk spesialis"}
				},
				"required": ["agent_name", "task"]
			}`),
		},
	}

	var out []driver.ToolDefinition
	for _, td := range catalog {
		if matchesAnyPattern(patterns, td.Name) {
			out = append(out, td)
		}
	}
	return out
}

func (s *Service) isParishAdmin(p authx.Principal) (bool, error) {
	if s.enforcer == nil || p.ParishID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(p.Subject(), roleParishAdmin, authz.ParishDomain(p.ParishID))
}

func (s *Service) bestEffortEmbed(ctx context.Context, text string) []float32 {
	embCfg := s.config.EmbeddingConfig()
	if embCfg.Provider == "" || embCfg.Provider == string(domain.ProviderNone) || strings.TrimSpace(embCfg.APIKey) == "" {
		return nil
	}
	vec, err := s.drivers.Embed(ctx, embCfg.Provider, embCfg.BaseURL, embCfg.APIKey, embCfg.Model, text)
	if err != nil {
		s.logger.WarnContext(ctx, "best-effort embedding failed", "provider", embCfg.Provider, "err", err)
		return nil
	}
	return vec
}

func (s *Service) executeInternalTool(
	ctx context.Context,
	p authx.Principal,
	ch domain.Channel,
	sessionID uuid.UUID,
	mcpSession *mcp.ClientSession,
	onStep StepCallback,
	name string,
	argsJSON string,
) (string, error) {
	switch name {
	case "memory_save":
		return s.execMemorySave(ctx, p, sessionID, argsJSON)
	case "memory_recall":
		return s.execMemoryRecall(ctx, p, argsJSON)
	case "memory_forget":
		return s.execMemoryForget(ctx, p, argsJSON)
	case "memory_search_history":
		return s.execMemorySearchHistory(ctx, p, argsJSON)
	case "delegate_specialist":
		return s.execDelegateSpecialist(ctx, p, ch, sessionID, mcpSession, onStep, argsJSON)
	default:
		return "", errx.Invalid("UNKNOWN_TOOL", fmt.Sprintf("Unknown internal tool %q.", name))
	}
}

func (s *Service) execMemorySave(ctx context.Context, p authx.Principal, sessionID uuid.UUID, argsJSON string) (string, error) {
	var args struct {
		Content     string  `json:"content"`
		Category    string  `json:"category"`
		IsPinned    bool    `json:"is_pinned"`
		Scope       string  `json:"scope"`
		SupersedeID *string `json:"supersede_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", errx.Invalid("INVALID_TOOL_ARGS", "Invalid JSON arguments for memory_save.")
	}
	content := strings.TrimSpace(args.Content)
	if content == "" || len([]rune(content)) > domain.MaxMemoryContentLen {
		return "", domain.ValidationError("content", "INVALID_CONTENT", "Memory content must be 1..500 characters.")
	}
	cat := domain.MemoryCategory(strings.TrimSpace(args.Category))
	switch cat {
	case domain.MemoryCategoryProfile, domain.MemoryCategoryPreference, domain.MemoryCategoryMinistry, domain.MemoryCategoryParishFAQ:
	default:
		return "", domain.ValidationError("category", "INVALID_CATEGORY", "Invalid memory category.")
	}
	scope := domain.MemoryScope(strings.TrimSpace(args.Scope))
	if scope == "" {
		scope = domain.MemoryScopeUser
	}
	if scope != domain.MemoryScopeUser && scope != domain.MemoryScopeParish {
		return "", domain.ValidationError("scope", "INVALID_SCOPE", "Scope must be user or parish.")
	}
	if scope == domain.MemoryScopeParish {
		ok, err := s.isParishAdmin(p)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", domain.ErrForbidden
		}
	}

	var supersedeUUID *uuid.UUID
	if args.SupersedeID != nil && strings.TrimSpace(*args.SupersedeID) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*args.SupersedeID))
		if err != nil {
			return "", domain.ValidationError("supersede_id", "INVALID_UUID", "Invalid supersede_id UUID.")
		}
		supersedeUUID = &parsed
	}

	var userPtr *uuid.UUID
	if scope == domain.MemoryScopeUser {
		uid := p.UserID
		userPtr = &uid
	}
	emb := s.bestEffortEmbed(ctx, content)
	saved, err := s.repo.SaveMemory(ctx, domain.SaveMemoryInput{
		ParishID:        p.ParishID,
		UserID:          userPtr,
		Scope:           scope,
		Category:        cat,
		Content:         content,
		IsPinned:        args.IsPinned,
		SourceSessionID: &sessionID,
		SupersedeID:     supersedeUUID,
		Embedding:       emb,
	})
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(map[string]any{
		"id":        saved.ID,
		"scope":     saved.Scope,
		"category":  saved.Category,
		"content":   saved.Content,
		"is_pinned": saved.IsPinned,
	})
	return string(b), nil
}

func (s *Service) execMemoryRecall(ctx context.Context, p authx.Principal, argsJSON string) (string, error) {
	var args struct {
		Query    string `json:"query"`
		Category string `json:"category"`
		Limit    int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", errx.Invalid("INVALID_TOOL_ARGS", "Invalid JSON arguments for memory_recall.")
	}
	q := strings.TrimSpace(args.Query)
	if q == "" {
		return "", domain.ValidationError("query", "REQUIRED", "Query is required.")
	}
	limit := clampLimit(args.Limit, 10, 20)
	var catPtr *domain.MemoryCategory
	if strings.TrimSpace(args.Category) != "" {
		c := domain.MemoryCategory(strings.TrimSpace(args.Category))
		catPtr = &c
	}
	emb := s.bestEffortEmbed(ctx, q)
	rows, err := s.repo.RecallMemories(ctx, domain.RecallMemoriesInput{
		ParishID:       p.ParishID,
		UserID:         p.UserID,
		Query:          q,
		QueryEmbedding: emb,
		Category:       catPtr,
		Limit:          limit,
	})
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(map[string]any{"memories": rows})
	return string(b), nil
}

func (s *Service) execMemoryForget(ctx context.Context, p authx.Principal, argsJSON string) (string, error) {
	var args struct {
		MemoryID string `json:"memory_id"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", errx.Invalid("INVALID_TOOL_ARGS", "Invalid JSON arguments for memory_forget.")
	}
	memID, err := uuid.Parse(strings.TrimSpace(args.MemoryID))
	if err != nil {
		return "", domain.ValidationError("memory_id", "INVALID_UUID", "Invalid memory_id UUID.")
	}
	row, err := s.repo.FindMemoryByID(ctx, p.ParishID, memID)
	if err != nil {
		return "", err
	}
	if row.Scope == agentmemory.ScopeUser {
		if row.UserID == nil || *row.UserID != p.UserID {
			return "", domain.ErrForbidden
		}
	} else {
		ok, err := s.isParishAdmin(p)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", domain.ErrForbidden
		}
	}
	if err := s.repo.SoftDeleteMemory(ctx, p.ParishID, p.UserID, memID); err != nil {
		return "", err
	}
	return `{"deleted":true}`, nil
}

func (s *Service) execMemorySearchHistory(ctx context.Context, p authx.Principal, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", errx.Invalid("INVALID_TOOL_ARGS", "Invalid JSON arguments for memory_search_history.")
	}
	q := strings.TrimSpace(args.Query)
	if q == "" {
		return "", domain.ValidationError("query", "REQUIRED", "Query is required.")
	}
	limit := clampLimit(args.Limit, 10, 20)
	rows, err := s.repo.SearchUserHistory(ctx, domain.SearchUserHistoryInput{
		ParishID: p.ParishID,
		UserID:   p.UserID,
		Q:        q,
		Limit:    limit,
	})
	if err != nil {
		return "", err
	}
	items := make([]map[string]any, 0, len(rows))
	for _, m := range rows {
		items = append(items, map[string]any{
			"session_id": m.SessionID,
			"role":       m.Role,
			"agent_name": m.AgentName,
			"content":    m.Content,
			"created_at": m.CreatedAt,
		})
	}
	b, _ := json.Marshal(map[string]any{"messages": items})
	return string(b), nil
}

func (s *Service) execDelegateSpecialist(
	ctx context.Context,
	p authx.Principal,
	ch domain.Channel,
	sessionID uuid.UUID,
	mcpSession *mcp.ClientSession,
	onStep StepCallback,
	argsJSON string,
) (string, error) {
	var args struct {
		AgentName string `json:"agent_name"`
		Task      string `json:"task"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", errx.Invalid("INVALID_TOOL_ARGS", "Invalid JSON arguments for delegate_specialist.")
	}
	if inDelegation(ctx) {
		return "", domain.ValidationError("agent_name", "NESTED_DELEGATION", "Specialists cannot delegate to other specialists.")
	}
	target := strings.TrimSpace(args.AgentName)
	if target != "mikael" && target != "rafael" {
		return "", domain.ValidationError("agent_name", "INVALID_AGENT", "agent_name must be mikael or rafael.")
	}
	task := strings.TrimSpace(args.Task)
	if task == "" {
		return "", domain.ValidationError("task", "REQUIRED", "Task is required.")
	}

	specCfg, ok := s.config.AgentConfig(target)
	if !ok {
		return "", domain.ErrProviderUnavailable
	}
	sysPrompt, err := s.buildPromptForTurn(ctx, p, ch, specCfg.Instruction)
	if err != nil {
		return "", err
	}

	answer, err := s.runReActLoop(
		withDelegation(ctx),
		p,
		ch,
		sessionID,
		specCfg,
		sysPrompt,
		[]driver.ChatMessage{{Role: domain.RoleUser, Content: task}},
		mcpSession,
		onStep,
		true, // stream specialist reply directly to user
	)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(map[string]any{
		"agent_name": target,
		"answer":     answer,
	})
	return string(b), nil
}

func clampLimit(v, def, max int) int {
	if v <= 0 {
		return def
	}
	if v > max {
		return max
	}
	return v
}
