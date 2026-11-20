package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Period stores fiscal or liturgical periods for a Parish.
type Period struct {
	ent.Schema
}

// Mixin returns the shared mixins for Period.
func (Period) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Period entity.
func (Period) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("parish_id", uuid.UUID{}),
		field.Enum("type").Values("fiscal", "liturgical").Default("fiscal"),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("start_date").NotEmpty(),
		field.String("end_date").NotEmpty(),
		field.Enum("status").Values("open", "closed").Default("open"),
		field.Bool("is_current").Default(false),
	}
}

// Edges returns the edges of the Period entity.
func (Period) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("parish", Parish.Type).
			Ref("periods").
			Field("parish_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the Period entity.
func (Period) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parish_id", "code").Unique(),
		index.Fields("parish_id", "type", "is_current"),
	}
}
