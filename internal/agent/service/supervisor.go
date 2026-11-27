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
	GabrielModel    model.ToolCallingChatModel
	MikaelModel     model.ToolCallingChatModel
	RafaelModel     model.ToolCallingChatModel
	YosefModel      model.ToolCallingChatModel
	GregoriusModel  model.ToolCallingChatModel
	SesiliaModel    model.ToolCallingChatModel
	GabrielPrompt   string
	MikaelPrompt    string
	RafaelPrompt    string
	YosefPrompt     string
	GregoriusPrompt string
	SesiliaPrompt   string
	PromptContext   PromptContext
	GabrielTools    []tool.BaseTool
	MikaelTools     []tool.BaseTool
	RafaelTools     []tool.BaseTool
	YosefTools      []tool.BaseTool
	GregoriusTools  []tool.BaseTool
	SesiliaTools    []tool.BaseTool
	Emit            func(StreamEvent)
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
		output := out
		if err != nil && output == "" {
			output = err.Error()
		}
		if runes := []rune(output); len(runes) > 2000 {
			output = string(runes[:2000]) + "…"
		}
		emit(StreamEvent{
			Event: "tool_end",
			Data: map[string]any{
				"agent":      agentName,
				"toolCallId": input.CallID,
				"toolName":   input.Name,
				"isError":    isErr,
				"output":     output,
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

type subAgentSpec struct {
	name   string
	desc   string
	model  model.ToolCallingChatModel
	prompt string
	tools  []tool.BaseTool
}

func buildSupervisorRunner(ctx context.Context, cfg supervisorConfig) (*adk.Runner, error) {
	patchHandler, err := patchtoolcalls.New(ctx, nil)
	if err != nil {
		return nil, err
	}

	specs := []subAgentSpec{
		{
			name:   "mikael",
			desc:   "Parish events and RSVP specialist. Handles event_list, event_get, event_create, event_update, event_rsvp, and event_list_rsvps.",
			model:  cfg.MikaelModel,
			prompt: cfg.MikaelPrompt,
			tools:  cfg.MikaelTools,
		},
		{
			name:   "rafael",
			desc:   "Parish organization, reference master data, and user directory specialist. Handles org_*, ref_*, and user_* tools.",
			model:  cfg.RafaelModel,
			prompt: cfg.RafaelPrompt,
			tools:  cfg.RafaelTools,
		},
		{
			name:   "yosef",
			desc:   "Buku Induk Keluarga & Data Umat specialist. Handles umat_* and read-only territorial ref_get_paroki, ref_list_wilayah, ref_list_lingkungan.",
			model:  cfg.YosefModel,
			prompt: cfg.YosefPrompt,
			tools:  cfg.YosefTools,
		},
		{
			name:   "gregorius",
			desc:   "Koordinator Liturgi & Petugas Liturgi specialist. Handles liturgi_list_schedules, liturgi_get_schedule, liturgi_create_schedule, liturgi_update_schedule, liturgi_set_ministry, liturgi_set_petugas, liturgi_respond_petugas, and liturgi_approve_substitute.",
			model:  cfg.GregoriusModel,
			prompt: cfg.GregoriusPrompt,
			tools:  cfg.GregoriusTools,
		},
		{
			name:   "sesilia",
			desc:   "Musik & Lagu Liturgi specialist. Handles cantus_* and liturgi_list_schedules, liturgi_get_schedule, liturgi_set_lagu.",
			model:  cfg.SesiliaModel,
			prompt: cfg.SesiliaPrompt,
			tools:  cfg.SesiliaTools,
		},
	}

	gabrielAllTools := make([]tool.BaseTool, 0, len(cfg.GabrielTools)+len(specs))
	gabrielAllTools = append(gabrielAllTools, cfg.GabrielTools...)
	subAgents := make(map[string]bool, len(specs))

	for _, sp := range specs {
		agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
			Name:          sp.name,
			Description:   sp.desc,
			Model:         sp.model,
			MaxIterations: 6,
			Handlers:      []adk.ChatModelAgentMiddleware{patchHandler},
			GenModelInput: BuildGenModelInput(sp.prompt, cfg.PromptContext),
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools:               sp.tools,
					ToolCallMiddlewares: []compose.ToolMiddleware{newToolEventMiddleware(sp.name, cfg.Emit, nil)},
				},
			},
		})
		if err != nil {
			return nil, err
		}
		gabrielAllTools = append(gabrielAllTools, adk.NewAgentTool(ctx, agent))
		subAgents[sp.name] = true
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
// Only the supervisor ("gabriel" or unnamed root) emits user-facing token deltas.
func consumeStreamMessage(mv *adk.MessageVariant, agentName string, emit func(StreamEvent)) (*schema.Message, error) {
	if mv == nil {
		return nil, nil
	}
	isSupervisor := agentName == "" || agentName == "gabriel"
	if !mv.IsStreaming || mv.MessageStream == nil {
		if mv.Message == nil {
			return nil, nil
		}
		if mv.Role == schema.Assistant && isSupervisor && mv.Message.Content != "" && emit != nil {
			emit(StreamEvent{
				Event: "token",
				Data: map[string]string{
					"agent": agentName,
					"delta": mv.Message.Content,
				},
			})
		}
		return mv.Message, nil
	}

	defer mv.MessageStream.Close()
	var chunks []*schema.Message
	for {
		chunk, err := mv.MessageStream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if chunk == nil {
			continue
		}
		chunks = append(chunks, chunk)
		if mv.Role == schema.Assistant && isSupervisor && chunk.Content != "" && emit != nil {
			emit(StreamEvent{
				Event: "token",
				Data: map[string]string{
					"agent": agentName,
					"delta": chunk.Content,
				},
			})
		}
	}
	if len(chunks) == 0 {
		return nil, nil
	}
	return schema.ConcatMessages(chunks)
}
