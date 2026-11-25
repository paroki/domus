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

// Kelurahan stores civil sub-district/village (Kelurahan/Desa) reference data under a Kecamatan.
type Kelurahan struct {
	ent.Schema
}

// Annotations returns the table annotations for Kelurahan.
func (Kelurahan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "kelurahan"},
	}
}

// Mixin returns the shared mixins for Kelurahan.
func (Kelurahan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Kelurahan entity.
func (Kelurahan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("kecamatan_id", uuid.UUID{}),
		field.String("kode").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.String("kode_pos").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Kelurahan entity.
func (Kelurahan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("kecamatan", Kecamatan.Type).
			Ref("kelurahan").
			Field("kecamatan_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the Kelurahan entity.
func (Kelurahan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kecamatan_id", "kode").Unique(),
		index.Fields("kecamatan_id", "active"),
	}
}
