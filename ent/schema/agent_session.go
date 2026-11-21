package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AgentSession represents a multi-turn AI agent conversation session.
type AgentSession struct {
	ent.Schema
}

// Annotations returns the table annotations for AgentSession.
func (AgentSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "agent_sessions"},
	}
}

// Mixin returns the shared mixins for AgentSession.
func (AgentSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the AgentSession entity.
func (AgentSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("channel").Values("web", "telegram", "whatsapp").Default("web"),
		field.String("title").NotEmpty().Default("Percakapan Baru"),
		field.Time("last_message_at").Default(time.Now),
	}
}

// Edges returns the intra-feature edges of the AgentSession entity.
func (AgentSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("messages", AgentMessage.Type),
	}
}

// Indexes returns the indexes for the AgentSession entity.
func (AgentSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "user_id", "last_message_at"),
		index.Fields("parish_id", "user_id", "channel", "last_message_at"),
	}
}
