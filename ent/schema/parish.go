package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Parish stores parish tenant metadata within a Diocese.
type Parish struct {
	ent.Schema
}

// Mixin returns the shared mixins for Parish.
func (Parish) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Parish entity.
func (Parish) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("diocese_id", uuid.UUID{}),
		field.String("code").NotEmpty().Unique(),
		field.String("name").NotEmpty(),
		field.String("patron_saint").Optional().Nillable(),
		field.String("pastor_name").Optional().Nillable(),
		field.String("address").Optional().Nillable(),
		field.String("phone").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
		field.String("website").Optional().Nillable(),
		field.String("logo_url").Optional().Nillable(),
		field.String("letterhead").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Parish entity.
func (Parish) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("diocese", Diocese.Type).
			Ref("parishes").
			Field("diocese_id").
			Unique().
			Required(),
		edge.To("members", ParishMember.Type),
		edge.To("zones", Zone.Type),
		edge.To("wards", Ward.Type),
		edge.To("lookups", Lookup.Type),
		edge.To("periods", Period.Type),
	}
}

// Indexes returns the indexes for the Parish entity.
func (Parish) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("diocese_id", "active"),
	}
}
