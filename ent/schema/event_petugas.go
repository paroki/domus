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

// EventPetugas records an officer/minister assignment for an Event.
type EventPetugas struct {
	ent.Schema
}

// Annotations returns the table annotations for EventPetugas.
func (EventPetugas) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "event_petugas"},
	}
}

// Mixin returns the shared mixins for EventPetugas.
func (EventPetugas) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the EventPetugas entity.
func (EventPetugas) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}),
		field.String("role").NotEmpty(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("org_id", uuid.UUID{}).Optional().Nillable(),
		field.String("name").NotEmpty(),
		field.Int("urutan").Default(0),
		field.Enum("status").Values("assigned", "confirmed", "declined").Default("assigned"),
		field.String("note").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the EventPetugas entity.
func (EventPetugas) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("petugas").
			Field("event_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the EventPetugas entity.
func (EventPetugas) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "urutan"),
		index.Fields("paroki_id", "user_id"),
		index.Fields("paroki_id", "org_id"),
	}
}
