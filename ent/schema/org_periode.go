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

// OrgPeriode stores service/term periods scoped to an Organization.
type OrgPeriode struct {
	ent.Schema
}

// Annotations returns the table annotations for OrgPeriode.
func (OrgPeriode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "org_periode"},
	}
}

// Mixin returns the shared mixins for OrgPeriode.
func (OrgPeriode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the OrgPeriode entity.
func (OrgPeriode) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("org_id", uuid.UUID{}),
		field.String("kode").NotEmpty(),
		field.String("nama").NotEmpty(),
		field.String("tanggal_mulai").NotEmpty(),
		field.String("tanggal_selesai").NotEmpty(),
		field.Enum("status").Values("open", "closed").Default("open"),
		field.Bool("is_current").Default(false),
	}
}

// Edges returns the intra-feature edges of the OrgPeriode entity.
func (OrgPeriode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("periode").
			Field("org_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the OrgPeriode entity.
func (OrgPeriode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "kode").Unique(),
		index.Fields("org_id", "is_current"),
		index.Fields("paroki_id", "org_id"),
	}
}
