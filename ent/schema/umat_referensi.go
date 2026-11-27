package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UmatReferensi stores global or paroki-scoped Data Umat reference lookup values.
type UmatReferensi struct {
	ent.Schema
}

// Annotations returns the table annotations for UmatReferensi.
func (UmatReferensi) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "umat_referensi"},
	}
}

// Mixin returns the shared mixins for UmatReferensi.
func (UmatReferensi) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the UmatReferensi entity.
func (UmatReferensi) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}).Optional().Nillable(),
		field.String("category").NotEmpty(),
		field.String("code").NotEmpty(),
		field.String("label").NotEmpty(),
		field.Int("sort_order").Default(0),
		field.Bool("active").Default(true),
	}
}

// Indexes returns the indexes for the UmatReferensi entity.
func (UmatReferensi) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "category", "code").Unique(),
		index.Fields("category", "code").
			Unique().
			Annotations(entsql.IndexWhere("paroki_id IS NULL")).
			StorageKey("umat_referensi_global_category_code"),
		index.Fields("paroki_id", "category", "sort_order"),
	}
}
