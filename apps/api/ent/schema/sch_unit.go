package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Unit holds the schema definition for the Unit entity.
type Unit struct {
	ent.Schema
}

func (Unit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDV7Mixin{},
		AuditMixins{},
	}
}

// Fields of the Unit.
func (Unit) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
	}
}

// Edges of the Unit.
func (Unit) Edges() []ent.Edge {
	return nil
}
