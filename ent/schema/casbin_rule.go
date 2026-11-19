package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CasbinRule stores Casbin policy and role-grouping tuples.
type CasbinRule struct {
	ent.Schema
}

// Mixin returns the shared mixins for CasbinRule.
func (CasbinRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
	}
}

// Fields returns the fields of CasbinRule.
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("ptype").NotEmpty(),
		field.String("v0").Default(""),
		field.String("v1").Default(""),
		field.String("v2").Default(""),
		field.String("v3").Default(""),
		field.String("v4").Default(""),
		field.String("v5").Default(""),
	}
}

// Indexes returns the composite lookup index for CasbinRule.
func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ptype", "v0", "v1", "v2", "v3"),
	}
}
