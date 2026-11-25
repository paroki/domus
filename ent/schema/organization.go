package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Organization represents a parish organizational unit (DPP, BGKP, section, zone, ward, etc.).
type Organization struct {
	ent.Schema
}

// Mixin returns the shared mixins for Organization.
func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Organization entity.
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("wilayah_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("lingkungan_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("kind").Values("dpp", "bgkp", "section", "categorical", "wilayah", "lingkungan", "community").Default("section"),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("description").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the intra-feature edges of the Organization entity.
func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Organization.Type).
			From("parent").
			Field("parent_id").
			Unique(),
		edge.To("members", OrgMember.Type),
		edge.To("invites", OrgInvite.Type),
	}
}

// Indexes returns the indexes for the Organization entity.
func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "code").Unique(),
		index.Fields("paroki_id", "kind", "active"),
		index.Fields("paroki_id", "parent_id"),
		index.Fields("paroki_id", "wilayah_id"),
		index.Fields("paroki_id", "lingkungan_id"),
	}
}
