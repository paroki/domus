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

// Kecamatan stores civil district (Kecamatan) reference data under a Kota.
type Kecamatan struct {
	ent.Schema
}

// Annotations returns the table annotations for Kecamatan.
func (Kecamatan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "kecamatan"},
	}
}

// Mixin returns the shared mixins for Kecamatan.
func (Kecamatan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Kecamatan entity.
func (Kecamatan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("kota_id", uuid.UUID{}),
		field.String("kode").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Kecamatan entity.
func (Kecamatan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("kota", Kota.Type).
			Ref("kecamatan").
			Field("kota_id").
			Unique().
			Required(),
		edge.To("kelurahan", Kelurahan.Type),
	}
}

// Indexes returns the indexes for the Kecamatan entity.
func (Kecamatan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kota_id", "kode").Unique(),
		index.Fields("kota_id", "active"),
	}
}
