package service_test

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/stretchr/testify/require"
)

func TestIsSafeInternalPath(t *testing.T) {
	valid := []string{
		"/",
		"/org",
		"/org/123",
		"/event/abc?tab=rsvp",
		"/umat/keluarga/123#members",
		"/liturgi/123",
		"/cantus/456",
		"/me",
		"/users",
		"/settings",
	}
	for _, p := range valid {
		require.True(t, service.IsSafeInternalPath(p), "expected valid path %q", p)
	}

	invalid := []string{
		"",
		"org/123",
		"//evil.com",
		"https://evil.com",
		"/org\n/evil",
		"/organization-evil",
		"/unknown/route",
		"/admin",
	}
	for _, p := range invalid {
		require.False(t, service.IsSafeInternalPath(p), "expected invalid path %q", p)
	}
}

func TestBuildGenModelInput_RuntimeContextAndRules(t *testing.T) {
	fn := service.BuildGenModelInput("Custom admin override", service.PromptContext{
		UserName:       "Budi",
		ParokiName:     "Paroki Katedral",
		UserRoles:      []string{"user"},
		UserTimezone:   "Asia/Makassar",
		ParokiTimezone: "Asia/Jakarta",
		CurrentTime:    "Friday, 2026-11-27T17:00:00+07:00",
		PagePath:       "/org/123",
		PageTitle:      "OMK",
		ActiveEntity:   "Active Organization: OMK (orgId: 123)",
	})
	msgs, err := fn(context.Background(), "gabriel", &adk.AgentInput{
		Messages: []*schema.Message{schema.UserMessage("Halo")},
	})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	sys := msgs[0].Content
	require.Contains(t, sys, "Custom admin override")
	require.Contains(t, sys, "- User Timezone: Asia/Makassar")
	require.Contains(t, sys, "- Paroki Timezone: Asia/Jakarta")
	require.Contains(t, sys, "- Current Time: Friday, 2026-11-27T17:00:00+07:00")
	require.Contains(t, sys, "- Active Page: /org/123 (OMK)")
	require.Contains(t, sys, "- Active Organization: OMK (orgId: 123)")
	require.Contains(t, sys, "- Language Rule: Always reply to the user in Bahasa Indonesia.")
	require.Contains(t, sys, "- Timezone Rule: Interpret and display all parish and organization event schedules in Paroki Timezone (Asia/Jakarta)")
}
