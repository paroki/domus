package prompts

import (
	"context"
	"embed"
	"strings"
	"sync"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

//go:embed gabriel.md
var rawGabriel string

//go:embed system.prompt
var systemPromptFS embed.FS

// Embedded default system instructions for Domus AI agent personas.
var (
	GabrielInstruction = strings.TrimSpace(rawGabriel)
)

// SystemPromptData holds pre-escaped variables for system.prompt (Dotprompt).
type SystemPromptData struct {
	Instruction      string `json:"instruction"`
	Channel          string `json:"channel"`
	CurrentTime      string `json:"currentTime"`
	ParishXML        string `json:"parishXML"`
	RolesXML         string `json:"rolesXML"`
	OrgsXML          string `json:"orgsXML"`
	UserName         string `json:"userName"`
	Salutation       string `json:"salutation"`
	NeedsOnboarding  bool   `json:"needsOnboarding"`
	UserMemoryXML    string `json:"userMemoryXML"`
	A2UIInstructions string `json:"a2uiInstructions"`
}

var (
	promptOnce sync.Once
	promptTmpl ai.Prompt
)

func getSystemPrompt() ai.Prompt {
	promptOnce.Do(func() {
		g := genkit.Init(
			context.Background(),
			genkit.WithPromptFS(systemPromptFS),
			genkit.WithPromptDir("."),
		)
		promptTmpl = genkit.LookupPrompt(g, "system")
	})
	return promptTmpl
}

// RenderSystemPrompt renders system.prompt using Genkit's Dotprompt engine.
func RenderSystemPrompt(ctx context.Context, data SystemPromptData) (string, error) {
	opts, err := getSystemPrompt().Render(ctx, data)
	if err != nil {
		return "", err
	}
	if len(opts.Messages) == 0 {
		return "", nil
	}
	return strings.TrimSpace(opts.Messages[0].Text()), nil
}
