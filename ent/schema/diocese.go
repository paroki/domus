package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Diocese stores diocesan organizational metadata.
type Diocese struct {
	ent.Schema
}

// Mixin returns the shared mixins for Diocese.
func (Diocese) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Diocese entity.
func (Diocese) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().Unique(),
		field.String("name").NotEmpty(),
		field.String("bishop_name").Optional().Nillable(),
		field.String("address").Optional().Nillable(),
		field.String("phone").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
		field.String("website").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Diocese entity.
func (Diocese) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("parishes", Parish.Type),
	}
}

// Indexes returns the indexes for the Diocese entity.
func (Diocese) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("active"),
	}
}
