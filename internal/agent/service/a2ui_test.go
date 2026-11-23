package service

import (
	"testing"

	"github.com/paroki/domus/internal/agent/domain"
	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt_A2UIOnlyOnWebChannel(t *testing.T) {
	webPrompt := BuildSystemPrompt(PromptInput{
		Instruction:    "Halo",
		IsOrchestrator: true,
		Channel:        domain.ChannelWeb,
	})
	assert.Contains(t, webPrompt, "<a2ui_instructions>")
	assert.Contains(t, webPrompt, A2UIBasicCatalogID)

	specialistPrompt := BuildSystemPrompt(PromptInput{
		Instruction:    "Halo",
		IsOrchestrator: false,
		Channel:        domain.ChannelWeb,
	})
	assert.NotContains(t, specialistPrompt, "<a2ui_instructions>")

	tgPrompt := BuildSystemPrompt(PromptInput{
		Instruction:    "Halo",
		IsOrchestrator: true,
		Channel:        domain.ChannelTelegram,
	})
	assert.NotContains(t, tgPrompt, "<a2ui_instructions>")
}

func TestSanitizeA2UIBlocks(t *testing.T) {
	valid := "Berikut jadwalnya:\n\n```a2ui\n[{\"version\":\"v0.9\",\"createSurface\":{\"surfaceId\":\"s1\",\"catalogId\":\"" + A2UIBasicCatalogID + "\"}}]\n```"
	assert.Equal(t, valid, SanitizeA2UIBlocks(valid))

	broken := "Berikut jadwalnya:\n\n```a2ui\n{not valid json\n```"
	assert.Equal(t, "Berikut jadwalnya:", SanitizeA2UIBlocks(broken))

	wrongVersion := "Halo\n\n```a2ui\n[{\"version\":\"v0.8\",\"createSurface\":{\"surfaceId\":\"s1\"}}]\n```"
	assert.Equal(t, "Halo", SanitizeA2UIBlocks(wrongVersion))

	legacyProps := "Halo\n\n```a2ui\n[{\"version\":\"v0.9\",\"updateComponents\":{\"surfaceId\":\"s1\",\"components\":[{\"id\":\"root\",\"type\":\"Column\",\"props\":{\"children\":[]}}]}}]\n```"
	assert.Equal(t, "Halo", SanitizeA2UIBlocks(legacyProps))

	validComponents := "Halo\n\n```a2ui\n[{\"version\":\"v0.9\",\"updateComponents\":{\"surfaceId\":\"s1\",\"components\":[{\"id\":\"root\",\"component\":\"Text\",\"text\":\"OK\"}]}}]\n```"
	assert.Equal(t, validComponents, SanitizeA2UIBlocks(validComponents))
}
