package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/platform/authx"
)

type rememberInput struct {
	Content  string `json:"content" jsonschema:"required" jsonschema_description:"Durable preference, fact, or parish context to remember (1-2000 chars)"`
	Category string `json:"category,omitempty" jsonschema:"enum=preference,enum=fact,enum=context" jsonschema_description:"Memory category (default: fact)"`
}

type forgetInput struct {
	MemoryID string `json:"memoryId" jsonschema:"required" jsonschema_description:"UUID of the memory item to delete"`
}

// NewMemoryTools constructs internal agent_remember and agent_forget tools for gabriel.
// These tools are never exposed on the public /mcp server.
func NewMemoryTools(s *Service, p authx.Principal) ([]tool.BaseTool, error) {
	rememberTool, err := utils.InferTool(
		"agent_remember",
		"Save a durable user preference, fact, or parish context to long-term memory.",
		func(ctx context.Context, in rememberInput) (string, error) {
			res, err := s.CreateMemory(ctx, p, domain.CreateMemoryRequest{
				Content:  in.Content,
				Category: in.Category,
			})
			if err != nil {
				return fmt.Sprintf("Error: %s", err.Error()), nil
			}
			b, _ := json.Marshal(map[string]any{
				"id":       res.ID.String(),
				"category": res.Category,
				"content":  res.Content,
			})
			return string(b), nil
		},
	)
	if err != nil {
		return nil, err
	}

	forgetTool, err := utils.InferTool(
		"agent_forget",
		"Delete an existing long-term memory item by its UUID.",
		func(ctx context.Context, in forgetInput) (string, error) {
			id, err := uuid.Parse(in.MemoryID)
			if err != nil {
				return "Error: invalid memoryId UUID", nil
			}
			if err := s.DeleteMemory(ctx, p, id); err != nil {
				return fmt.Sprintf("Error: %s", err.Error()), nil
			}
			return `{"deleted":true}`, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{rememberTool, forgetTool}, nil
}
