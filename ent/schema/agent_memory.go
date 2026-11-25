package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AgentMemory stores a long-term parish/user memory for hybrid RAG search.
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
	}
}

// Fields returns the fields of the AgentMemory entity.
func (AgentMemory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("category").Values("preference", "fact", "context").Default("fact"),
		field.Text("content").NotEmpty(),
		field.UUID("source_session_id", uuid.UUID{}).Optional().Nillable(),
	}
}

// Indexes returns the indexes of the AgentMemory entity.
func (AgentMemory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "user_id", "created_at"),
	}
}
