package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LiturgiPetugas records an officer/minister assignment for a Mass schedule (scalar event_id).
type LiturgiPetugas struct {
	ent.Schema
}

// Annotations returns the table annotations for LiturgiPetugas.
func (LiturgiPetugas) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "liturgi_petugas"},
	}
}

// Mixin returns the shared mixins for LiturgiPetugas.
func (LiturgiPetugas) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the LiturgiPetugas entity.
func (LiturgiPetugas) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}),
		field.Enum("role").Values("lektor", "mazmur", "pembagi-komuni", "misdinar", "organis", "kor"),
		field.String("slot_label").Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("org_id", uuid.UUID{}).Optional().Nillable(),
		field.String("name").NotEmpty(),
		field.Int("urutan").Default(0),
		field.Enum("status").Values("assigned", "confirmed", "declined").Default("assigned"),
		field.String("note").Optional().Nillable(),
		field.Time("responded_at").Optional().Nillable(),
		field.UUID("responded_by", uuid.UUID{}).Optional().Nillable(),
		field.UUID("substitute_user_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("substitute_org_id", uuid.UUID{}).Optional().Nillable(),
		field.String("substitute_name").Optional().Nillable(),
	}
}

// Indexes returns the indexes for the LiturgiPetugas entity.
func (LiturgiPetugas) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "role", "urutan"),
		index.Fields("paroki_id", "event_id"),
		index.Fields("paroki_id", "user_id"),
		index.Fields("paroki_id", "org_id"),
		index.Fields("paroki_id", "substitute_org_id"),
		index.Fields("paroki_id", "status"),
	}
}
