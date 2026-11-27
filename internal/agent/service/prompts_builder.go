package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/internal/agent/domain"
)

const defaultParokiTimezone = "Asia/Makassar"

// PromptContext holds dynamic per-turn context injected into GenModelInput.
type PromptContext struct {
	UserName          string
	ParokiName        string
	UserRoles         []string
	UserTimezone      string
	ParokiTimezone    string
	CurrentTime       string
	PagePath          string
	PageTitle         string
	ActiveEntity      string
	RetrievedMemories []domain.MemoryResponse
}

// BuildGenModelInput constructs a TypedGenModelInput that prepends personaText + context block
// as a SystemMessage without passing personaText through Eino's {Var} Instruction template parser.
func BuildGenModelInput(personaText string, pctx PromptContext) adk.TypedGenModelInput[*schema.Message] {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(personaText))
	b.WriteString("\n\n## Runtime Context\n")
	if pctx.UserName != "" {
		fmt.Fprintf(&b, "- User: %s\n", pctx.UserName)
	}
	if pctx.ParokiName != "" {
		fmt.Fprintf(&b, "- Active Paroki: %s\n", pctx.ParokiName)
	}
	if len(pctx.UserRoles) > 0 {
		fmt.Fprintf(&b, "- User Roles: %s\n", strings.Join(pctx.UserRoles, ", "))
	}
	if pctx.UserTimezone != "" {
		fmt.Fprintf(&b, "- User Timezone: %s\n", pctx.UserTimezone)
	}
	if pctx.ParokiTimezone != "" {
		fmt.Fprintf(&b, "- Paroki Timezone: %s\n", pctx.ParokiTimezone)
	}
	if pctx.CurrentTime != "" {
		fmt.Fprintf(&b, "- Current Time: %s\n", pctx.CurrentTime)
	}
	if pctx.PagePath != "" {
		if pctx.PageTitle != "" {
			fmt.Fprintf(&b, "- Active Page: %s (%s)\n", pctx.PagePath, pctx.PageTitle)
		} else {
			fmt.Fprintf(&b, "- Active Page: %s\n", pctx.PagePath)
		}
	}
	if pctx.ActiveEntity != "" {
		fmt.Fprintf(&b, "- %s\n", pctx.ActiveEntity)
	}
	parokiTZ := pctx.ParokiTimezone
	if parokiTZ == "" {
		parokiTZ = defaultParokiTimezone
	}
	b.WriteString("- Language Rule: Always reply to the user in Bahasa Indonesia.\n")
	fmt.Fprintf(&b, "- Timezone Rule: Interpret and display all parish and organization event schedules in Paroki Timezone (%s) unless the user explicitly specifies another timezone. Never append a 'Z' (UTC) suffix to local wall-clock times when calling tools.\n", parokiTZ)

	if len(pctx.RetrievedMemories) > 0 {
		b.WriteString("\n## Relevant Long-Term Memories\n")
		for _, m := range pctx.RetrievedMemories {
			fmt.Fprintf(&b, "- [%s] (%s) %s\n", m.ID.String(), m.Category, m.Content)
		}
	}

	sysMsg := schema.SystemMessage(b.String())
	return func(_ context.Context, _ string, input *adk.AgentInput) ([]*schema.Message, error) {
		msgs := make([]*schema.Message, 0, 1+len(input.Messages))
		msgs = append(msgs, sysMsg)
		msgs = append(msgs, input.Messages...)
		return msgs, nil
	}
}
