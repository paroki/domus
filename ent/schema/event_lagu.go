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

// EventLagu records a liturgical song entry for an Event.
type EventLagu struct {
	ent.Schema
}

// Annotations returns the table annotations for EventLagu.
func (EventLagu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "event_lagu"},
	}
}

// Mixin returns the shared mixins for EventLagu.
func (EventLagu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the EventLagu entity.
func (EventLagu) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}),
		field.Int("urutan").Default(0),
		field.String("bagian").NotEmpty(),
		field.String("sumber").Optional().Nillable(),
		field.String("nomor").Optional().Nillable(),
		field.String("judul").NotEmpty(),
		field.String("nada_dasar").Optional().Nillable(),
		field.String("note").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the EventLagu entity.
func (EventLagu) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("lagu").
			Field("event_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the EventLagu entity.
func (EventLagu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "urutan"),
	}
}
