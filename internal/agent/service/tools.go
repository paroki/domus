package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/ent/agentmemory"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
)

const roleParishAdmin = "ref:parish-admin"

type memorySaveArgs struct {
	Content     string  `json:"content" jsonschema:"description=Fakta singkat (maks 500 karakter)"`
	Category    string  `json:"category" jsonschema:"enum=profile,enum=preference,enum=ministry,enum=parish_faq,description=Kategori fakta memori"`
	IsPinned    bool    `json:"is_pinned,omitempty" jsonschema:"description=Pin agar selalu disuntikkan ke konteks"`
	Scope       string  `json:"scope,omitempty" jsonschema:"enum=user,enum=parish,description=Cakupan memori (user pribadi atau paroki)"`
	SupersedeID *string `json:"supersede_id,omitempty" jsonschema:"description=UUID memori lama yang digantikan"`
}

type memoryRecallArgs struct {
	Query    string `json:"query" jsonschema:"description=Kata kunci pencarian fakta memori"`
	Category string `json:"category,omitempty" jsonschema:"enum=profile,enum=preference,enum=ministry,enum=parish_faq,description=Filter kategori memori opsional"`
	Limit    int    `json:"limit,omitempty" jsonschema:"minimum=1,maximum=20,description=Jumlah maksimal hasil memori (1-20)"`
}

type memoryForgetArgs struct {
	MemoryID string `json:"memory_id" jsonschema:"description=UUID memori yang akan dihapus"`
}

type memorySearchHistoryArgs struct {
	Query string `json:"query" jsonschema:"description=Kata kunci pencarian riwayat percakapan lintas sesi"`
	Limit int    `json:"limit,omitempty" jsonschema:"minimum=1,maximum=20,description=Jumlah maksimal hasil riwayat (1-20)"`
}

type delegateSpecialistArgs struct {
	AgentName string `json:"agent_name" jsonschema:"enum=mikael,enum=rafael,description=Nama agen spesialis tujuan (mikael atau rafael)"`
	Task      string `json:"task" jsonschema:"description=Instruksi tugas jelas untuk spesialis"`
}

func (s *Service) internalTools(
	p authx.Principal,
	ch domain.Channel,
	sessionID uuid.UUID,
	mcpSession *mcp.ClientSession,
	onStep StepCallback,
	patterns []string,
) []ai.Tool {
	catalog := []ai.Tool{
		ai.NewTool(
			"memory_save",
			"Simpan fakta jangka panjang tentang pengguna atau FAQ paroki ke dalam memori.",
			func(tc *ai.ToolContext, args memorySaveArgs) (string, error) {
				return s.execMemorySave(tc.Context, p, sessionID, args)
			},
		),
		ai.NewTool(
			"memory_recall",
			"Cari fakta jangka panjang yang relevan dari memori pengguna dan FAQ paroki.",
			func(tc *ai.ToolContext, args memoryRecallArgs) (string, error) {
				return s.execMemoryRecall(tc.Context, p, args)
			},
		),
		ai.NewTool(
			"memory_forget",
			"Hapus (soft-delete) satu fakta memori berdasarkan UUID.",
			func(tc *ai.ToolContext, args memoryForgetArgs) (string, error) {
				return s.execMemoryForget(tc.Context, p, args)
			},
		),
		ai.NewTool(
			"memory_search_history",
			"Cari riwayat percakapan terdahulu lintas sesi dan saluran (web, telegram, whatsapp).",
			func(tc *ai.ToolContext, args memorySearchHistoryArgs) (string, error) {
				return s.execMemorySearchHistory(tc.Context, p, args)
			},
		),
		ai.NewTool(
			"delegate_specialist",
			"Delegasikan tugas spesialis ke Mikael (acara/RSVP) atau Rafael (organisasi/undangan).",
			func(tc *ai.ToolContext, args delegateSpecialistArgs) (string, error) {
				return s.execDelegateSpecialist(tc.Context, p, ch, sessionID, mcpSession, onStep, args)
			},
		),
	}

	var out []ai.Tool
	for _, t := range catalog {
		if matchesAnyPattern(patterns, t.Name()) {
			out = append(out, t)
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

func (s *Service) execMemorySave(ctx context.Context, p authx.Principal, sessionID uuid.UUID, args memorySaveArgs) (string, error) {
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

func (s *Service) execMemoryRecall(ctx context.Context, p authx.Principal, args memoryRecallArgs) (string, error) {
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

func (s *Service) execMemoryForget(ctx context.Context, p authx.Principal, args memoryForgetArgs) (string, error) {
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

func (s *Service) execMemorySearchHistory(ctx context.Context, p authx.Principal, args memorySearchHistoryArgs) (string, error) {
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
	args delegateSpecialistArgs,
) (string, error) {
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
