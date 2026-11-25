package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Lookup stores global or paroki-scoped reference lookup values.
type Lookup struct {
	ent.Schema
}

// Mixin returns the shared mixins for Lookup.
func (Lookup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Lookup entity.
func (Lookup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}).Optional().Nillable(),
		field.String("category").NotEmpty(),
		field.String("code").NotEmpty(),
		field.String("label").NotEmpty(),
		field.Int("sort_order").Default(0),
		field.Bool("active").Default(true),
	}
}

// Edges returns the edges of the Lookup entity.
func (Lookup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("paroki", Paroki.Type).
			Ref("lookups").
			Field("paroki_id").
			Unique(),
	}
}

// Indexes returns the indexes for the Lookup entity.
func (Lookup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "category", "code").Unique(),
		index.Fields("category", "code").
			Unique().
			Annotations(entsql.IndexWhere("paroki_id IS NULL")).
			StorageKey("lookup_global_category_code"),
		index.Fields("paroki_id", "category", "sort_order"),
	}
}
