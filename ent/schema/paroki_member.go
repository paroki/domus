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

// ParokiMember links a User ID to a Paroki with a tenant membership role.
type ParokiMember struct {
	ent.Schema
}

// Annotations returns the table annotations for ParokiMember.
func (ParokiMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "paroki_members"},
	}
}

// Mixin returns the shared mixins for ParokiMember.
func (ParokiMember) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the ParokiMember entity.
func (ParokiMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("role").Values("admin", "member").Default("member"),
		field.Bool("umat_admin").Default(false),
		field.Bool("liturgi_admin").Default(false),
	}
}

// Edges returns the edges of the ParokiMember entity.
func (ParokiMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("paroki", Paroki.Type).
			Ref("members").
			Field("paroki_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the ParokiMember entity.
func (ParokiMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}
