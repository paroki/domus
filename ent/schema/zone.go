package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Zone stores a territorial zone or mission station within a Parish.
type Zone struct {
	ent.Schema
}

// Mixin returns the shared mixins for Zone.
func (Zone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Zone entity.
func (Zone) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.Enum("type").Values("zone", "station").Default("zone"),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("coordinator_name").Optional().Nillable(),
		field.String("coordinator_phone").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Zone entity.
func (Zone) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parish", Parish.Type).
			Ref("zones").
			Field("parish_id").
			Unique().
			Required(),
		edge.To("wards", Ward.Type),
	}
}

// Indexes returns the indexes for the Zone entity.
func (Zone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "code").Unique(),
		index.Fields("parish_id", "type", "active"),
	}
}
