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

// Paroki stores parish tenant metadata within a Keuskupan.
type Paroki struct {
	ent.Schema
}

// Annotations returns the table annotations for Paroki.
func (Paroki) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "paroki"},
	}
}

// Mixin returns the shared mixins for Paroki.
func (Paroki) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Paroki entity.
func (Paroki) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("keuskupan_id", uuid.UUID{}),
		field.String("kode").NotEmpty().Unique(),
		field.String("nama").NotEmpty(),
		field.String("pelindung").Optional().Nillable(),
		field.String("pastor").Optional().Nillable(),
		field.String("alamat").Optional().Nillable(),
		field.String("telepon").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
		field.String("website").Optional().Nillable(),
		field.String("logo_url").Optional().Nillable(),
		field.String("kop_surat").Optional().Nillable(),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Paroki entity.
func (Paroki) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("keuskupan", Keuskupan.Type).
			Ref("paroki").
			Field("keuskupan_id").
			Unique().
			Required(),
		edge.To("members", ParokiMember.Type),
		edge.To("wilayah", Wilayah.Type),
		edge.To("lingkungan", Lingkungan.Type),
		edge.To("lookups", Lookup.Type),
		edge.To("periods", Period.Type),
	}
}

// Indexes returns the indexes for the Paroki entity.
func (Paroki) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kode").Unique(),
		index.Fields("keuskupan_id", "active"),
	}
}
