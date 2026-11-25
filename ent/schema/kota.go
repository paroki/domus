package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Kota stores civil regency/city (Kabupaten/Kota) reference data.
type Kota struct {
	ent.Schema
}

// Annotations returns the table annotations for Kota.
func (Kota) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "kota"},
	}
}

// Mixin returns the shared mixins for Kota.
func (Kota) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Kota entity.
func (Kota) Fields() []ent.Field {
	return []ent.Field{
		field.String("kode").NotEmpty().Unique(),
		field.String("nama").NotEmpty(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Kota entity.
func (Kota) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("kecamatan", Kecamatan.Type),
	}
}

// Indexes returns the indexes for the Kota entity.
func (Kota) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kode").Unique(),
		index.Fields("active"),
	}
}
