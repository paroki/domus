package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AgentMemory stores long-term user or parish knowledge for AI agent retrieval.
type AgentMemory struct {
	ent.Schema
}

// Annotations returns the table annotations for AgentMemory.
func (AgentMemory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "agent_memories"},
	}
}

// Mixin returns the shared mixins for AgentMemory.
func (AgentMemory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the AgentMemory entity.
func (AgentMemory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("scope").Values("user", "parish").Default("user"),
		field.Enum("category").Values("profile", "preference", "ministry", "parish_faq").Default("preference"),
		field.Text("content").NotEmpty(),
		field.Bool("is_pinned").Default(false),
		field.UUID("source_session_id", uuid.UUID{}).Optional().Nillable(),
	}
}

// Indexes returns the indexes for the AgentMemory entity.
func (AgentMemory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "user_id", "is_pinned"),
		index.Fields("parish_id", "scope", "category"),
	}
}
