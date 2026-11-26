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

// UmatSakramen stores sacramental history rows for a parishioner.
type UmatSakramen struct {
	ent.Schema
}

// Annotations returns the table annotations for UmatSakramen.
func (UmatSakramen) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "umat_sakramen"},
	}
}

// Mixin returns the shared mixins for UmatSakramen (hard-delete child row).
func (UmatSakramen) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UmatSakramen entity.
func (UmatSakramen) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("umat_id", uuid.UUID{}),
		field.Enum("jenis").Values("baptis", "komuni_pertama", "krisma", "perkawinan", "imamat", "perminyakan", "kematian"),
		field.String("tanggal").Optional().Nillable(),
		field.UUID("keuskupan_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("paroki_sakramen_id", uuid.UUID{}).Optional().Nillable(),
		field.String("keuskupan_nama").Optional().Nillable(),
		field.String("paroki_nama").Optional().Nillable(),
		field.String("buku").Optional().Nillable(),
		field.String("nomor").Optional().Nillable(),
		field.String("catatan").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the UmatSakramen entity.
func (UmatSakramen) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("umat", Umat.Type).
			Ref("sakramen").
			Field("umat_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the UmatSakramen entity.
func (UmatSakramen) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "umat_id", "jenis").Unique(),
		index.Fields("paroki_id", "jenis", "tanggal"),
	}
}
