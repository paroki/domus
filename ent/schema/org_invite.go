package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OrgInvite stores an invitation code for joining an Organization.
type OrgInvite struct {
	ent.Schema
}

// Mixin returns the shared mixins for OrgInvite.
func (OrgInvite) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the OrgInvite entity.
func (OrgInvite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("org_id", uuid.UUID{}),
		field.String("code").NotEmpty(),
		field.Enum("role").Values("admin", "member").Default("member"),
		field.String("default_position").Optional().Nillable(),
		field.Int("max_uses").Optional().Nillable(),
		field.Int("used_count").Default(0),
		field.Time("expires_at"),
		field.Time("revoked_at").Optional().Nillable(),
		field.UUID("created_by", uuid.UUID{}),
	}
}

// Edges returns the intra-feature edges of the OrgInvite entity.
func (OrgInvite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("invites").
			Field("org_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the OrgInvite entity.
func (OrgInvite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("org_id", "created_at"),
		index.Fields("paroki_id", "org_id"),
	}
}
