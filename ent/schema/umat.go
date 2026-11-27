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

// Umat stores individual parishioner profile records.
type Umat struct {
	ent.Schema
}

// Annotations returns the table annotations for Umat.
func (Umat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "umat"},
	}
}

// Mixin returns the shared mixins for Umat.
func (Umat) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Umat entity.
func (Umat) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("keluarga_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("wilayah_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("lingkungan_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.String("nomor_induk").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.String("nama_baptis").Optional().Nillable(),
		field.String("nik_enc").Optional().Nillable(),
		field.String("nik_hash").Optional().Nillable(),
		field.Enum("jenis_kelamin").Values("pria", "wanita"),
		field.String("tempat_lahir").Optional().Nillable(),
		field.String("tanggal_lahir").Optional().Nillable(),
		field.String("golongan_darah").Optional().Nillable(),
		field.String("suku").Optional().Nillable(),
		field.String("agama").Optional().Nillable(),
		field.String("kewarganegaraan").Optional().Nillable(),
		field.String("negara").Optional().Nillable(),
		field.String("hubungan_keluarga").Optional().Nillable(),
		field.String("status_pernikahan").Optional().Nillable(),
		field.Bool("almarhum").Default(false),
		field.String("tanggal_meninggal").Optional().Nillable(),
		field.String("alamat").Optional().Nillable(),
		field.UUID("kota_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("kecamatan_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("kelurahan_id", uuid.UUID{}).Optional().Nillable(),
		field.String("telepon").Optional().Nillable(),
		field.String("email").Optional().Nillable(),
		field.String("foto_url").Optional().Nillable(),
		field.String("pendidikan").Optional().Nillable(),
		field.String("jurusan").Optional().Nillable(),
		field.String("pekerjaan").Optional().Nillable(),
		field.String("profesi").Optional().Nillable(),
		field.String("keterampilan").Optional().Nillable(),
		field.String("disabilitas").Optional().Nillable(),
		field.String("aktivitas_sosial").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the Umat entity.
func (Umat) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("keluarga", Keluarga.Type).
			Ref("anggota").
			Field("keluarga_id").
			Unique(),
		edge.To("sakramen", UmatSakramen.Type),
		edge.To("dokumen", UmatDokumen.Type),
	}
}

// Indexes returns the indexes for the Umat entity.
func (Umat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "nomor_induk").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")).
			StorageKey("umat_paroki_nomor_induk_active"),
		index.Fields("paroki_id", "nik_hash").
			Unique().
			Annotations(entsql.IndexWhere("nik_hash IS NOT NULL AND deleted_at IS NULL")).
			StorageKey("umat_paroki_nik_hash_active"),
		index.Fields("paroki_id", "user_id").
			Unique().
			Annotations(entsql.IndexWhere("user_id IS NOT NULL AND deleted_at IS NULL")).
			StorageKey("umat_paroki_user_id_active"),
		index.Fields("paroki_id", "keluarga_id"),
		index.Fields("paroki_id", "wilayah_id", "lingkungan_id", "almarhum"),
		index.Fields("paroki_id", "nama"),
	}
}
