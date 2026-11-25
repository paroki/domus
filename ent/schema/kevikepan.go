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

// Kevikepan stores deanery/vicariate organizational metadata under a Keuskupan.
type Kevikepan struct {
	ent.Schema
}

// Annotations returns the table annotations for Kevikepan.
func (Kevikepan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "kevikepan"},
	}
}

// Mixin returns the shared mixins for Kevikepan.
func (Kevikepan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Kevikepan entity.
func (Kevikepan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("keuskupan_id", uuid.UUID{}),
		field.String("kode").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.String("vikep").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Kevikepan entity.
func (Kevikepan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("keuskupan", Keuskupan.Type).
			Ref("kevikepan").
			Field("keuskupan_id").
			Unique().
			Required(),
		edge.To("paroki", Paroki.Type),
	}
}

// Indexes returns the indexes for the Kevikepan entity.
func (Kevikepan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("keuskupan_id", "kode").Unique(),
		index.Fields("keuskupan_id", "active"),
	}
}
