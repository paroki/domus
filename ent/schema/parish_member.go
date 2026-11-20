package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ParishMember links a User ID to a Parish with a tenant membership role.
type ParishMember struct {
	ent.Schema
}

// Mixin returns the shared mixins for ParishMember.
func (ParishMember) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the ParishMember entity.
func (ParishMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("role").Values("admin", "member").Default("member"),
	}
}

// Edges returns the edges of the ParishMember entity.
func (ParishMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parish", Parish.Type).
			Ref("members").
			Field("parish_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the ParishMember entity.
func (ParishMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}
