package prompts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedPersonas(t *testing.T) {
	assert.NotEmpty(t, GabrielInstruction)
	assert.Contains(t, GabrielInstruction, "Bahasa Indonesia")
	assert.Contains(t, GabrielInstruction, "event_list")
	assert.Contains(t, GabrielInstruction, "RFC3339")
	assert.Contains(t, GabrielInstruction, "EVENT_RSVP_CLOSED")
	assert.Contains(t, GabrielInstruction, "org_list")
	assert.Contains(t, GabrielInstruction, "org_create_invite")
	assert.Contains(t, GabrielInstruction, "org:admin")
	assert.Contains(t, GabrielInstruction, "memory_save")
	assert.Contains(t, GabrielInstruction, "<user salutation=")
	assert.Contains(t, GabrielInstruction, "ref:parish-admin")
	assert.Contains(t, GabrielInstruction, "Google")
}

func TestRenderSystemPrompt(t *testing.T) {
	withOnboarding, err := RenderSystemPrompt(context.Background(), SystemPromptData{
		Instruction:     "Base",
		Channel:         "web",
		CurrentTime:     "2026-11-23T00:00:00Z",
		UserName:        "Anthonius",
		Salutation:      "Bapak Anthonius",
		NeedsOnboarding: true,
	})
	require.NoError(t, err)
	assert.Contains(t, withOnboarding, `<user name="Anthonius" salutation="Bapak Anthonius" />`)
	assert.Contains(t, withOnboarding, "First-Time Preference Interview")

	withoutOnboarding, err := RenderSystemPrompt(context.Background(), SystemPromptData{
		Instruction:     "Base",
		Channel:         "web",
		CurrentTime:     "2026-11-23T00:00:00Z",
		UserName:        "Anthonius",
		Salutation:      "Bapak Anthonius",
		NeedsOnboarding: false,
	})
	require.NoError(t, err)
	assert.NotContains(t, withoutOnboarding, "First-Time Preference Interview")
}
