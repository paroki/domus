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

// Lingkungan stores a neighborhood community (lingkungan) within a Wilayah and Paroki.
type Lingkungan struct {
	ent.Schema
}

// Annotations returns the table annotations for Lingkungan.
func (Lingkungan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "lingkungan"},
	}
}

// Mixin returns the shared mixins for Lingkungan.
func (Lingkungan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Lingkungan entity.
func (Lingkungan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("wilayah_id", uuid.UUID{}),
		field.String("kode").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.String("pelindung").Optional().Nillable(),
		field.String("ketua").Optional().Nillable(),
		field.String("telepon_ketua").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Lingkungan entity.
func (Lingkungan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("paroki", Paroki.Type).
			Ref("lingkungan").
			Field("paroki_id").
			Unique().
			Required(),
		edge.From("wilayah", Wilayah.Type).
			Ref("lingkungan").
			Field("wilayah_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the Lingkungan entity.
func (Lingkungan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "kode").Unique(),
		index.Fields("paroki_id", "wilayah_id", "active"),
	}
}
