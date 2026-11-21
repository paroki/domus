package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddedPersonas(t *testing.T) {
	assert.NotEmpty(t, GabrielInstruction)
	assert.Contains(t, GabrielInstruction, "Bahasa Indonesia")
	assert.Contains(t, GabrielInstruction, "delegate_specialist")
	assert.Contains(t, GabrielInstruction, "memory_save")
	assert.Contains(t, GabrielInstruction, "<user_profile>")
	assert.Contains(t, GabrielInstruction, "ref:parish-admin")
	assert.Contains(t, GabrielInstruction, "Google")

	assert.NotEmpty(t, MikaelInstruction)
	assert.Contains(t, MikaelInstruction, "event_list")
	assert.Contains(t, MikaelInstruction, "RFC3339")
	assert.Contains(t, MikaelInstruction, "EVENT_RSVP_CLOSED")

	assert.NotEmpty(t, RafaelInstruction)
	assert.Contains(t, RafaelInstruction, "org_list")
	assert.Contains(t, RafaelInstruction, "org_create_invite")
	assert.Contains(t, RafaelInstruction, "org:admin")
}
