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

// OrgJabatan stores hierarchical per-period position assignments within an Organization.
type OrgJabatan struct {
	ent.Schema
}

// Annotations returns the table annotations for OrgJabatan.
func (OrgJabatan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "org_jabatan"},
	}
}

// Mixin returns the shared mixins for OrgJabatan.
func (OrgJabatan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the OrgJabatan entity.
func (OrgJabatan) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("org_id", uuid.UUID{}),
		field.UUID("periode_id", uuid.UUID{}),
		field.UUID("parent_id", uuid.UUID{}).Optional().Nillable(),
		field.String("jabatan").NotEmpty(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("umat_id", uuid.UUID{}).Optional().Nillable(),
		field.String("nama").Optional().Nillable(),
		field.String("start_date").Optional().Nillable(),
		field.String("end_date").Optional().Nillable(),
		field.Int("urutan").Default(0),
		field.Bool("active").Default(true),
	}
}

// Edges returns the intra-feature edges of the OrgJabatan entity.
func (OrgJabatan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("jabatan").
			Field("org_id").
			Unique().
			Required(),
		edge.From("periode", OrgPeriode.Type).
			Ref("jabatan").
			Field("periode_id").
			Unique().
			Required(),
		edge.To("children", OrgJabatan.Type).
			From("parent").
			Field("parent_id").
			Unique(),
	}
}

// Indexes returns the indexes for the OrgJabatan entity.
func (OrgJabatan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("org_id", "periode_id", "parent_id", "urutan"),
		index.Fields("org_id", "periode_id", "active"),
		index.Fields("paroki_id", "user_id"),
		index.Fields("paroki_id", "umat_id"),
		index.Fields("paroki_id", "org_id"),
	}
}
