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

// AgentSession stores an AI assistant conversation session scoped to a parish and user.
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
		field.String("title").Default("Percakapan Baru"),
		field.UUID("provider_id", uuid.UUID{}).Optional().Nillable(),
		field.String("model_id").Optional().Nillable(),
		field.Time("last_message_at").Default(func() time.Time { return time.Now().UTC() }),
	}
}

// Edges returns the edges of the AgentSession entity.
func (AgentSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("messages", AgentMessage.Type),
	}
}

// Indexes returns the indexes of the AgentSession entity.
func (AgentSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "user_id", "last_message_at"),
	}
}
