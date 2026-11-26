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

// Keluarga stores parish household (Kartu Keluarga) records.
type Keluarga struct {
	ent.Schema
}

// Annotations returns the table annotations for Keluarga.
func (Keluarga) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "keluarga"},
	}
}

// Mixin returns the shared mixins for Keluarga.
func (Keluarga) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Keluarga entity.
func (Keluarga) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("wilayah_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("lingkungan_id", uuid.UUID{}).Optional().Nillable(),
		field.String("nomor_kk_enc").Optional().Nillable(),
		field.String("nomor_kk_hash").Optional().Nillable(),
		field.String("nama_keluarga").NotEmpty(),
		field.String("alamat").Optional().Nillable(),
		field.UUID("kota_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("kecamatan_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("kelurahan_id", uuid.UUID{}).Optional().Nillable(),
		field.String("kode_pos").Optional().Nillable(),
		field.String("status_rumah").Optional().Nillable(),
		field.String("status_ekonomi").Optional().Nillable(),
		field.Bool("aktif").Default(true),
		field.String("catatan").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the Keluarga entity.
func (Keluarga) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("anggota", Umat.Type),
	}
}

// Indexes returns the indexes for the Keluarga entity.
func (Keluarga) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "nomor_kk_hash").
			Unique().
			Annotations(entsql.IndexWhere("nomor_kk_hash IS NOT NULL AND deleted_at IS NULL")).
			StorageKey("keluarga_paroki_nomor_kk_hash_active"),
		index.Fields("paroki_id", "wilayah_id", "lingkungan_id", "aktif"),
		index.Fields("paroki_id", "nama_keluarga"),
	}
}
