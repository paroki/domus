package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/prompts"
	"github.com/paroki/domus/internal/platform/authx"
)

const (
	settingModelMemoryExtractor  = "agent.model.memory_extractor"
	settingPromptMemoryExtractor = "agent.prompt.memory_extractor"
)

type extractedMemoryItem struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}

func (s *Service) dispatchMemoryExtraction(
	p authx.Principal,
	sessionID uuid.UUID,
	userMsg, assistantMsg string,
) {
	if s.repo == nil || s.db == nil || s.drivers == nil || s.extractSem == nil {
		return
	}
	if strings.TrimSpace(userMsg) == "" || strings.TrimSpace(assistantMsg) == "" {
		return
	}

	select {
	case s.extractSem <- struct{}{}:
	default:
		return
	}

	go func() {
		defer func() { <-s.extractSem }()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := s.runMemoryExtraction(ctx, p, sessionID, userMsg, assistantMsg); err != nil {
			s.logger.WarnContext(ctx, "async memory extraction failed", "sessionId", sessionID, "err", err)
		}
	}()
}

func (s *Service) runMemoryExtraction(
	ctx context.Context,
	p authx.Principal,
	sessionID uuid.UUID,
	userMsg, assistantMsg string,
) error {
	prov, plainKey, modelID, err := s.resolveSettingModel(ctx, settingModelMemoryExtractor, true)
	if err != nil {
		return err
	}
	cm, err := s.drivers.NewChatModel(ctx, string(prov.Provider), prov.BaseURL, plainKey, modelID)
	if err != nil {
		return err
	}

	sysPrompt := strings.TrimSpace(prompts.MemoryExtractor)
	if s.settings != nil {
		if custom := strings.TrimSpace(s.settings.Get(settingPromptMemoryExtractor)); custom != "" {
			sysPrompt = custom
		}
	}
	sysPrompt += "\n\nReturn ONLY a JSON array of objects with fields {\"category\":\"preference|fact|context\",\"content\":\"...\"}. Return [] if there are no durable net-new facts."

	turnText := fmt.Sprintf("User: %s\nAssistant: %s", userMsg, assistantMsg)
	resp, err := cm.Generate(ctx, []*schema.Message{
		schema.SystemMessage(sysPrompt),
		schema.UserMessage(turnText),
	})
	if err != nil {
		return err
	}

	raw := strings.TrimSpace(resp.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}

	var items []extractedMemoryItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	existing, _, err := s.repo.SearchMemories(ctx, p.ParishID, p.UserID, "", "", nil, nil, nil, 100)
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(existing))
	for _, m := range existing {
		seen[strings.ToLower(strings.TrimSpace(m.Content))] = true
	}

	for _, item := range items {
		content := strings.TrimSpace(item.Content)
		if err := validateMemoryContent(content); err != nil {
			continue
		}
		norm := strings.ToLower(content)
		if seen[norm] {
			continue
		}
		cat := strings.TrimSpace(item.Category)
		if !isValidMemoryCategory(cat) {
			cat = domain.MemoryCategoryFact
		}
		vec := s.embedTextBestEffort(ctx, content)
		if _, err := s.repo.CreateMemory(ctx, domain.CreateMemoryInput{
			ParishID:        p.ParishID,
			UserID:          p.UserID,
			Category:        cat,
			Content:         content,
			SourceSessionID: &sessionID,
		}, vec); err != nil {
			s.logger.WarnContext(ctx, "save extracted memory failed", "err", err)
			continue
		}
		seen[norm] = true
	}
	return nil
}
