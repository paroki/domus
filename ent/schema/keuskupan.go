package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Keuskupan stores diocesan organizational metadata.
type Keuskupan struct {
	ent.Schema
}

// Annotations returns the table annotations for Keuskupan.
func (Keuskupan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "keuskupan"},
	}
}

// Mixin returns the shared mixins for Keuskupan.
func (Keuskupan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Keuskupan entity.
func (Keuskupan) Fields() []ent.Field {
	return []ent.Field{
		field.String("kode").NotEmpty().Unique(),
		field.String("nama").NotEmpty(),
		field.String("uskup").Optional().Nillable(),
		field.String("alamat").Optional().Nillable(),
		field.String("telepon").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
		field.String("website").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Keuskupan entity.
func (Keuskupan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("paroki", Paroki.Type),
	}
}

// Indexes returns the indexes for the Keuskupan entity.
func (Keuskupan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kode").Unique(),
		index.Fields("active"),
	}
}
