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

// EventRSVP records a user's RSVP response for an Event.
type EventRSVP struct {
	ent.Schema
}

// Annotations returns the table annotations for EventRSVP.
func (EventRSVP) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "event_rsvps"},
	}
}

// Mixin returns the shared mixins for EventRSVP.
func (EventRSVP) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the EventRSVP entity.
func (EventRSVP) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("event_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("status").Values("going", "maybe", "not_going"),
		field.String("note").Optional().Nillable(),
	}
}

// Edges returns the intra-feature edges of the EventRSVP entity.
func (EventRSVP) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("event", Event.Type).
			Ref("rsvps").
			Field("event_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the EventRSVP entity.
func (EventRSVP) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_id", "user_id").Unique(),
		index.Fields("event_id", "status", "updated_at"),
		index.Fields("paroki_id", "user_id"),
	}
}
