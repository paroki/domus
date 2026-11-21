package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AgentToolCall represents a serialized LLM tool invocation on an AgentMessage.
type AgentToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
	AgentName string `json:"agent_name,omitempty"`
}

// AgentMessage stores a single message turn within an AgentSession.
type AgentMessage struct {
	ent.Schema
}

// Annotations returns the table annotations for AgentMessage.
func (AgentMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "agent_messages"},
	}
}

// Mixin returns the shared mixins for AgentMessage.
func (AgentMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the AgentMessage entity.
func (AgentMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("session_id", uuid.UUID{}),
		field.String("agent_name").NotEmpty().Default("gabriel"),
		field.Enum("role").Values("user", "assistant", "tool"),
		field.Text("content").Default(""),
		field.Text("thinking").Optional().Nillable(),
		field.JSON("tool_calls", []AgentToolCall{}).Optional(),
		field.String("tool_call_id").Optional().Nillable(),
		field.String("tool_name").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the AgentMessage entity.
func (AgentMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", AgentSession.Type).
			Ref("messages").
			Field("session_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the AgentMessage entity.
func (AgentMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id", "created_at"),
		index.Fields("parish_id", "user_id", "created_at"),
	}
}
