package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LiturgiLagu records a liturgical song entry for a Mass schedule (scalar event_id and optional cantus_id).
type LiturgiLagu struct {
	ent.Schema
}

// Annotations returns the table annotations for LiturgiLagu.
func (LiturgiLagu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "liturgi_lagu"},
	}
}

// Mixin returns the shared mixins for LiturgiLagu.
func (LiturgiLagu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the LiturgiLagu entity.
func (LiturgiLagu) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}),
		field.Int("urutan").Default(0),
		field.String("bagian").NotEmpty(),
		field.UUID("cantus_id", uuid.UUID{}).Optional().Nillable(),
		field.String("sumber").Optional().Nillable(),
		field.String("nomor").Optional().Nillable(),
		field.String("judul").NotEmpty(),
		field.String("nada_dasar").Optional().Nillable(),
		field.String("note").Optional().Nillable(),
	}
}

// Indexes returns the indexes for the LiturgiLagu entity.
func (LiturgiLagu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "urutan"),
		index.Fields("paroki_id", "event_id"),
		index.Fields("cantus_id"),
	}
}
