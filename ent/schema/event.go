package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Event represents a scheduled parish or organizational event.
type Event struct {
	ent.Schema
}

// Mixin returns the shared mixins for Event.
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the Event entity.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("paroki_id", uuid.UUID{}),
		field.UUID("org_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("created_by", uuid.UUID{}),
		field.String("title").NotEmpty(),
		field.String("jenis").Optional().Nillable(),
		field.String("description").Optional().Nillable(),
		field.String("location").Optional().Nillable(),
		field.Time("start_at"),
		field.Time("end_at"),
		field.Enum("visibility").Values("public", "private").Default("public"),
		field.Bool("rsvp_enabled").Default(true),
		field.Time("rsvp_deadline").Optional().Nillable(),
		field.Enum("status").Values("scheduled", "cancelled").Default("scheduled"),
	}
}

// Edges returns the intra-feature edges of the Event entity.
func (Event) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("rsvps", EventRSVP.Type),
	}
}

// Indexes returns the indexes for the Event entity.
func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("paroki_id", "start_at"),
		index.Fields("paroki_id", "jenis", "start_at"),
		index.Fields("paroki_id", "org_id", "start_at"),
		index.Fields("paroki_id", "visibility", "status", "start_at"),
		index.Fields("org_id"),
	}
}
