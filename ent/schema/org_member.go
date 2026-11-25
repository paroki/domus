package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OrgMember links a User ID to an Organization with a role, position, and term.
type OrgMember struct {
	ent.Schema
}

// Mixin returns the shared mixins for OrgMember.
func (OrgMember) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the OrgMember entity.
func (OrgMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("org_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("role").Values("admin", "member").Default("member"),
		field.String("position").Optional().Nillable(),
		field.UUID("period_id", uuid.UUID{}).Optional().Nillable(),
		field.String("start_date").Optional().Nillable(),
		field.String("end_date").Optional().Nillable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
	}
}

// Edges returns the intra-feature edges of the OrgMember entity.
func (OrgMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("members").
			Field("org_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the OrgMember entity.
func (OrgMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "user_id").Unique(),
		index.Fields("paroki_id", "user_id"),
		index.Fields("org_id", "role", "status"),
	}
}
