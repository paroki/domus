package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Meta stores baseline key-value metadata records.
type Meta struct {
	ent.Schema
}

// Mixin returns the shared mixins for Meta.
func (Meta) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the Meta entity.
func (Meta) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty().Unique(),
		field.String("value"),
	}
}
