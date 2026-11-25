package service_test

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/stretchr/testify/require"
)

func TestProjectHistory_CoalesceAndCap(t *testing.T) {
	toolName := "event_list"
	rows := []domain.MessageResponse{
		{Role: "user", Content: "Pesan batal 1"},
		{Role: "user", Content: "Pesan baru 2"},
		{Role: "assistant", Content: "Berikut daftarnya."},
		{Role: "tool", ToolName: &toolName, Content: `{"events":[]}`},
	}
	projected := service.ProjectHistory(rows)
	require.Len(t, projected, 2)
	require.Equal(t, schema.User, projected[0].Role)
	require.Equal(t, "Pesan batal 1\n\nPesan baru 2", projected[0].Content)
	require.Equal(t, schema.Assistant, projected[1].Role)
	require.Contains(t, projected[1].Content, "Berikut daftarnya.")
	require.Contains(t, projected[1].Content, "[Tool event_list]:")

	// Cap at 40 messages
	var many []domain.MessageResponse
	for i := 0; i < 50; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		many = append(many, domain.MessageResponse{Role: role, Content: "msg"})
	}
	capped := service.ProjectHistory(many)
	require.Len(t, capped, 40)
}

func TestBuildGenModelInput_PreservesCurlyBraces(t *testing.T) {
	fn := service.BuildGenModelInput("Always reply politely. Note: {User} {Time} {curly_braces}", service.PromptContext{
		UserName:    "Budi",
		ParokiName:  "Paroki Katedral",
		UserRoles:   []string{"user:user"},
		CurrentTime: "2026-11-24T12:00:00Z",
		RetrievedMemories: []domain.MemoryResponse{
			{ID: uuid.New(), Category: "fact", Content: "Suka jadwal misa pagi"},
		},
	})
	out, err := fn(context.Background(), "", &adk.AgentInput{
		Messages: []*schema.Message{schema.UserMessage("Halo")},
	})
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, schema.System, out[0].Role)
	require.Contains(t, out[0].Content, "{User} {Time} {curly_braces}")
	require.Contains(t, out[0].Content, "Paroki Katedral")
	require.Contains(t, out[0].Content, "Suka jadwal misa pagi")
}
