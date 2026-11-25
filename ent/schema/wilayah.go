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

// Wilayah stores a territorial zone or mission station within a Paroki.
type Wilayah struct {
	ent.Schema
}

// Annotations returns the table annotations for Wilayah.
func (Wilayah) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "wilayah"},
	}
}

// Mixin returns the shared mixins for Wilayah.
func (Wilayah) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Wilayah entity.
func (Wilayah) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.Enum("jenis").Values("wilayah", "stasi").Default("wilayah"),
		field.String("kode").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.String("koordinator").Optional().Nillable(),
		field.String("telepon_koordinator").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Wilayah entity.
func (Wilayah) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("paroki", Paroki.Type).
			Ref("wilayah").
			Field("paroki_id").
			Unique().
			Required(),
		edge.To("lingkungan", Lingkungan.Type),
	}
}

// Indexes returns the indexes for the Wilayah entity.
func (Wilayah) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "kode").Unique(),
		index.Fields("paroki_id", "jenis", "active"),
	}
}
