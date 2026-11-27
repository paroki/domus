package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/paroki/domus/internal/agent/domain"
)

var allowedRoutePrefixes = []string{
	"/org",
	"/event",
	"/umat",
	"/liturgi",
	"/cantus",
	"/me",
	"/users",
	"/settings",
}

// TurnMetaRecorder captures structured UI tool payloads emitted during a single chat turn.
type TurnMetaRecorder struct {
	mu         sync.Mutex
	question   *domain.QuestionPayload
	navigation *domain.NavigationPayload
}

// SetQuestion records the latest question payload emitted in the turn.
func (r *TurnMetaRecorder) SetQuestion(q domain.QuestionPayload) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := q
	cp.Options = append([]domain.QuestionOption(nil), q.Options...)
	r.question = &cp
}

// SetNavigation records the latest navigation payload emitted in the turn.
func (r *TurnMetaRecorder) SetNavigation(nav domain.NavigationPayload) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := nav
	r.navigation = &cp
}

// Snapshot returns copies of any recorded question and navigation payloads.
func (r *TurnMetaRecorder) Snapshot() (*domain.QuestionPayload, *domain.NavigationPayload) {
	if r == nil {
		return nil, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var q *domain.QuestionPayload
	if r.question != nil {
		cp := *r.question
		cp.Options = append([]domain.QuestionOption(nil), r.question.Options...)
		q = &cp
	}
	var nav *domain.NavigationPayload
	if r.navigation != nil {
		cp := *r.navigation
		nav = &cp
	}
	return q, nav
}

// IsSafeInternalPath validates that path is a safe internal Domus SPA route.
func IsSafeInternalPath(path string) bool {
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return false
	}
	if strings.Contains(path, "://") {
		return false
	}
	for _, r := range path {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	seg := path
	if idx := strings.IndexAny(seg, "?#"); idx >= 0 {
		seg = seg[:idx]
	}
	if seg == "/" {
		return true
	}
	for _, prefix := range allowedRoutePrefixes {
		if seg == prefix || strings.HasPrefix(seg, prefix+"/") {
			return true
		}
	}
	return false
}

type askUserOptionInput struct {
	Label       string `json:"label" jsonschema:"required" jsonschema_description:"Short option label shown to the user (max 80 chars)"`
	Value       string `json:"value,omitempty" jsonschema_description:"Option value submitted when selected (defaults to label)"`
	Description string `json:"description,omitempty" jsonschema_description:"Optional helper description for this option"`
}

type askUserInput struct {
	Question    string               `json:"question" jsonschema:"required" jsonschema_description:"Clarifying or confirmation question to ask the user"`
	Options     []askUserOptionInput `json:"options" jsonschema:"required" jsonschema_description:"2 to 6 selectable options"`
	MultiSelect bool                 `json:"multiSelect,omitempty" jsonschema_description:"Allow selecting multiple options (default false)"`
	AllowCustom *bool                `json:"allowCustom,omitempty" jsonschema_description:"Allow user to type a custom answer (default true)"`
}

type navigateInput struct {
	Path         string `json:"path" jsonschema:"required" jsonschema_description:"Internal Domus route path starting with /"`
	Label        string `json:"label" jsonschema:"required" jsonschema_description:"Button or link label shown to the user"`
	Description  string `json:"description,omitempty" jsonschema_description:"Optional short description of the destination"`
	AutoRedirect bool   `json:"autoRedirect,omitempty" jsonschema_description:"Automatically redirect the user immediately (true) or show a navigation link (false)"`
}

// NewUITools constructs internal agent_ask_user and agent_navigate tools.
// These tools are never exposed on the public /mcp server.
func NewUITools(emit func(StreamEvent), recorder *TurnMetaRecorder) ([]tool.BaseTool, error) {
	askTool, err := utils.InferTool(
		"agent_ask_user",
		"Ask the user a structured clarifying or confirmation question with 2-6 selectable options.",
		func(_ context.Context, in askUserInput) (string, error) {
			q := strings.TrimSpace(in.Question)
			if q == "" {
				return "Error: question is required", nil
			}
			if len(in.Options) < 2 || len(in.Options) > 6 {
				return "Error: options must contain between 2 and 6 items", nil
			}
			opts := make([]domain.QuestionOption, 0, len(in.Options))
			for _, o := range in.Options {
				label := strings.TrimSpace(o.Label)
				if label == "" {
					return "Error: option label is required", nil
				}
				if utf8.RuneCountInString(label) > 80 {
					return "Error: option label must be at most 80 characters", nil
				}
				val := strings.TrimSpace(o.Value)
				if val == "" {
					val = label
				}
				opts = append(opts, domain.QuestionOption{
					Label:       label,
					Value:       val,
					Description: strings.TrimSpace(o.Description),
				})
			}
			allowCustom := true
			if in.AllowCustom != nil {
				allowCustom = *in.AllowCustom
			}
			payload := domain.QuestionPayload{
				Question:    q,
				Options:     opts,
				MultiSelect: in.MultiSelect,
				AllowCustom: allowCustom,
			}
			recorder.SetQuestion(payload)
			if emit != nil {
				emit(StreamEvent{Event: "question", Data: payload})
			}
			b, _ := json.Marshal(map[string]any{
				"asked":       true,
				"instruction": "Question sent to the user. Stop calling tools and wait for the user's reply in the next turn.",
			})
			return string(b), nil
		},
	)
	if err != nil {
		return nil, err
	}

	navTool, err := utils.InferTool(
		"agent_navigate",
		"Navigate or offer a navigation link to an internal Domus page.",
		func(_ context.Context, in navigateInput) (string, error) {
			p := strings.TrimSpace(in.Path)
			if !IsSafeInternalPath(p) {
				return "Error: invalid or unsafe internal path", nil
			}
			label := strings.TrimSpace(in.Label)
			if label == "" {
				return "Error: label is required", nil
			}
			payload := domain.NavigationPayload{
				Path:         p,
				Label:        label,
				Description:  strings.TrimSpace(in.Description),
				AutoRedirect: in.AutoRedirect,
			}
			recorder.SetNavigation(payload)
			if emit != nil {
				emit(StreamEvent{Event: "navigate", Data: payload})
			}
			b, _ := json.Marshal(map[string]any{
				"navigated": true,
				"path":      p,
			})
			return string(b), nil
		},
	)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{askTool, navTool}, nil
}
