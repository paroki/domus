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

// UmatDokumen stores encrypted identity document images for a parishioner or household.
type UmatDokumen struct {
	ent.Schema
}

// Annotations returns the table annotations for UmatDokumen.
func (UmatDokumen) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "umat_dokumen"},
	}
}

// Mixin returns the shared mixins for UmatDokumen (hard-delete child row).
func (UmatDokumen) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UmatDokumen entity.
func (UmatDokumen) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("umat_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("keluarga_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("jenis").Values("ktp", "kia", "kartu_pelajar", "kk"),
		field.Enum("mime_type").NamedValues("WebP", "image/webp", "JPEG", "image/jpeg", "PNG", "image/png"),
		field.Int("size_bytes"),
		field.Bytes("data_enc"),
		field.UUID("uploaded_by", uuid.UUID{}),
	}
}

// Edges returns the intra-feature edges of the UmatDokumen entity.
func (UmatDokumen) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("umat", Umat.Type).
			Ref("dokumen").
			Field("umat_id").
			Unique(),
		edge.From("keluarga", Keluarga.Type).
			Ref("dokumen").
			Field("keluarga_id").
			Unique(),
	}
}

// Indexes returns the indexes for the UmatDokumen entity.
func (UmatDokumen) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "umat_id", "jenis").
			Unique().
			Annotations(entsql.IndexWhere("umat_id IS NOT NULL")).
			StorageKey("umat_dokumen_paroki_umat_jenis_uq"),
		index.Fields("paroki_id", "keluarga_id", "jenis").
			Unique().
			Annotations(entsql.IndexWhere("keluarga_id IS NOT NULL")).
			StorageKey("umat_dokumen_paroki_keluarga_jenis_uq"),
	}
}
