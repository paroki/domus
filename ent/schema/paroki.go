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
		field.UUID("kevikepan_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("induk_id", uuid.UUID{}).Optional().Nillable(),
		field.String("kode").NotEmpty().Unique(),
		field.String("kode_resmi").Optional().Nillable(),
		field.String("nama").NotEmpty(),
		field.String("pelindung").Optional().Nillable(),
		field.String("pastor").Optional().Nillable(),
		field.String("alamat").Optional().Nillable(),
		field.UUID("kota_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("kecamatan_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("kelurahan_id", uuid.UUID{}).Optional().Nillable(),
		field.String("timezone").NotEmpty().Default("Asia/Makassar"),
		field.String("telepon").Optional().Nillable(),
		field.String("fax").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
		field.String("website").Optional().Nillable(),
		field.String("logo_url").Optional().Nillable(),
		field.String("kop_surat").Optional().Nillable(),
		field.Float("latitude").Optional().Nillable(),
		field.Float("longitude").Optional().Nillable(),
		field.String("status_paroki").Optional().Nillable(),
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
		edge.From("kevikepan", Kevikepan.Type).
			Ref("paroki").
			Field("kevikepan_id").
			Unique(),
		edge.To("anak", Paroki.Type).
			From("induk").
			Field("induk_id").
			Unique(),
		edge.From("kota", Kota.Type).
			Ref("paroki").
			Field("kota_id").
			Unique(),
		edge.From("kecamatan", Kecamatan.Type).
			Ref("paroki").
			Field("kecamatan_id").
			Unique(),
		edge.From("kelurahan", Kelurahan.Type).
			Ref("paroki").
			Field("kelurahan_id").
			Unique(),
		edge.To("members", ParokiMember.Type),
		edge.To("wilayah", Wilayah.Type),
		edge.To("lingkungan", Lingkungan.Type),
		edge.To("lookups", Lookup.Type),
	}
}

// Indexes returns the indexes for the Paroki entity.
func (Paroki) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("kode").Unique(),
		index.Fields("keuskupan_id", "kevikepan_id", "active"),
		index.Fields("induk_id"),
	}
}
