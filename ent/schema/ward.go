package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Ward stores a neighborhood community (lingkungan) within a Zone and Parish.
type Ward struct {
	ent.Schema
}

// Mixin returns the shared mixins for Ward.
func (Ward) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Ward entity.
func (Ward) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.UUID("zone_id", uuid.UUID{}),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("patron_saint").Optional().Nillable(),
		field.String("head_name").Optional().Nillable(),
		field.String("head_phone").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Ward entity.
func (Ward) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parish", Parish.Type).
			Ref("wards").
			Field("parish_id").
			Unique().
			Required(),
		edge.From("zone", Zone.Type).
			Ref("wards").
			Field("zone_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the Ward entity.
func (Ward) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "code").Unique(),
		index.Fields("parish_id", "zone_id", "active"),
	}
}
