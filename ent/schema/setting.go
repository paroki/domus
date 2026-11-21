package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Setting stores key-value runtime configuration records.
type Setting struct {
	ent.Schema
}

// Mixin returns the shared mixins for Setting.
func (Setting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the Setting entity.
func (Setting) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty().Unique(),
		field.String("value").Default(""),
	}
}
