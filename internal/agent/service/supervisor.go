package service

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/patchtoolcalls"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// StreamEvent represents a single SSE frame emitted during a chat turn.
type StreamEvent struct {
	Event string
	Data  any
}

type supervisorConfig struct {
	GabrielModel  model.ToolCallingChatModel
	MikaelModel   model.ToolCallingChatModel
	RafaelModel   model.ToolCallingChatModel
	GabrielPrompt string
	MikaelPrompt  string
	RafaelPrompt  string
	PromptContext PromptContext
	GabrielTools  []tool.BaseTool
	MikaelTools   []tool.BaseTool
	RafaelTools   []tool.BaseTool
	Emit          func(StreamEvent)
}

func newToolEventMiddleware(agentName string, emit func(StreamEvent), subAgents map[string]bool) compose.ToolMiddleware {
	emitStart := func(input *compose.ToolInput) {
		if emit == nil {
			return
		}
		if subAgents[input.Name] {
			emit(StreamEvent{
				Event: "agent_transfer",
				Data: map[string]string{
					"from": agentName,
					"to":   input.Name,
				},
			})
			return
		}
		emit(StreamEvent{
			Event: "tool_start",
			Data: map[string]any{
				"agent":      agentName,
				"toolCallId": input.CallID,
				"toolName":   input.Name,
				"args":       input.Arguments,
			},
		})
	}
	emitEnd := func(input *compose.ToolInput, out string, err error) {
		if emit == nil || subAgents[input.Name] {
			return
		}
		isErr := err != nil ||
			(strings.HasPrefix(out, "Tool ") && strings.Contains(out, " error: ")) ||
			strings.HasPrefix(out, "Error: ")
		emit(StreamEvent{
			Event: "tool_end",
			Data: map[string]any{
				"agent":      agentName,
				"toolCallId": input.CallID,
				"toolName":   input.Name,
				"isError":    isErr,
			},
		})
	}

	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.ToolOutput, error) {
				emitStart(input)
				out, err := next(ctx, input)
				var result string
				if out != nil {
					result = out.Result
				}
				emitEnd(input, result, err)
				return out, err
			}
		},
		Streamable: func(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
			return func(ctx context.Context, input *compose.ToolInput) (*compose.StreamToolOutput, error) {
				emitStart(input)
				out, err := next(ctx, input)
				emitEnd(input, "", err)
				return out, err
			}
		},
	}
}

func buildSupervisorRunner(ctx context.Context, cfg supervisorConfig) (*adk.Runner, error) {
	patchHandler, err := patchtoolcalls.New(ctx, nil)
	if err != nil {
		return nil, err
	}

	mikaelAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "mikael",
		Description:   "Parish events and RSVP specialist. Handles event_list, event_get, event_create, event_update, event_rsvp, and event_list_rsvps.",
		Model:         cfg.MikaelModel,
		MaxIterations: 6,
		Handlers:      []adk.ChatModelAgentMiddleware{patchHandler},
		GenModelInput: BuildGenModelInput(cfg.MikaelPrompt, cfg.PromptContext),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               cfg.MikaelTools,
				ToolCallMiddlewares: []compose.ToolMiddleware{newToolEventMiddleware("mikael", cfg.Emit, nil)},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	rafaelAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "rafael",
		Description:   "Parish organization, reference master data, and user directory specialist. Handles org_*, ref_*, and user_* tools.",
		Model:         cfg.RafaelModel,
		MaxIterations: 6,
		Handlers:      []adk.ChatModelAgentMiddleware{patchHandler},
		GenModelInput: BuildGenModelInput(cfg.RafaelPrompt, cfg.PromptContext),
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               cfg.RafaelTools,
				ToolCallMiddlewares: []compose.ToolMiddleware{newToolEventMiddleware("rafael", cfg.Emit, nil)},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	mikaelTool := adk.NewAgentTool(ctx, mikaelAgent)
	rafaelTool := adk.NewAgentTool(ctx, rafaelAgent)

	gabrielAllTools := make([]tool.BaseTool, 0, len(cfg.GabrielTools)+2)
	gabrielAllTools = append(gabrielAllTools, cfg.GabrielTools...)
	gabrielAllTools = append(gabrielAllTools, mikaelTool, rafaelTool)

	subAgents := map[string]bool{
		"mikael": true,
		"rafael": true,
	}

	gabrielAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "gabriel",
		Description:   "Primary Domus parish AI assistant supervisor.",
		Model:         cfg.GabrielModel,
		MaxIterations: 8,
		Handlers:      []adk.ChatModelAgentMiddleware{patchHandler},
		GenModelInput: BuildGenModelInput(cfg.GabrielPrompt, cfg.PromptContext),
		ToolsConfig: adk.ToolsConfig{
			EmitInternalEvents: true,
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools:               gabrielAllTools,
				ToolCallMiddlewares: []compose.ToolMiddleware{newToolEventMiddleware("gabriel", cfg.Emit, subAgents)},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           gabrielAgent,
		EnableStreaming: true,
	}), nil
}

// consumeStreamMessage reads a single AgentEvent's MessageOutput stream, closing it on return.
func consumeStreamMessage(mv *adk.MessageVariant, agentName string, emit func(StreamEvent)) (string, []*schema.ToolCall, error) {
	if mv == nil {
		return "", nil, nil
	}
	if !mv.IsStreaming || mv.MessageStream == nil {
		if mv.Message == nil {
			return "", nil, nil
		}
		var tcs []*schema.ToolCall
		for i := range mv.Message.ToolCalls {
			tc := mv.Message.ToolCalls[i]
			tcs = append(tcs, &tc)
		}
		if mv.Role == schema.Assistant && mv.Message.Content != "" && emit != nil {
			emit(StreamEvent{
				Event: "token",
				Data: map[string]string{
					"agent": agentName,
					"delta": mv.Message.Content,
				},
			})
		}
		return mv.Message.Content, tcs, nil
	}

	defer mv.MessageStream.Close()
	var content strings.Builder
	var toolCalls []*schema.ToolCall
	for {
		chunk, err := mv.MessageStream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return content.String(), toolCalls, err
		}
		if chunk == nil {
			continue
		}
		if mv.Role == schema.Assistant && chunk.Content != "" {
			content.WriteString(chunk.Content)
			if emit != nil {
				emit(StreamEvent{
					Event: "token",
					Data: map[string]string{
						"agent": agentName,
						"delta": chunk.Content,
					},
				})
			}
		}
		for i := range chunk.ToolCalls {
			tc := chunk.ToolCalls[i]
			toolCalls = append(toolCalls, &tc)
		}
	}
	return content.String(), toolCalls, nil
}
