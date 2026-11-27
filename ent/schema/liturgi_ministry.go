package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LiturgiMinistry configures the organization and leader for one of the 6 parish liturgical ministry roles.
type LiturgiMinistry struct {
	ent.Schema
}

// Annotations returns the table annotations for LiturgiMinistry.
func (LiturgiMinistry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "liturgi_ministries"},
	}
}

// Mixin returns the shared mixins for LiturgiMinistry.
func (LiturgiMinistry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the LiturgiMinistry entity.
func (LiturgiMinistry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.Enum("role_slug").Values("lektor", "mazmur", "pembagi-komuni", "misdinar", "organis", "kor"),
		field.UUID("org_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("leader_user_id", uuid.UUID{}).Optional().Nillable(),
	}
}

// Indexes returns the indexes for the LiturgiMinistry entity.
func (LiturgiMinistry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "role_slug").Unique(),
		index.Fields("org_id"),
		index.Fields("paroki_id", "leader_user_id"),
	}
}
