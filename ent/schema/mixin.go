package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// NewID generates a time-ordered UUIDv7 identifier.
func NewID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// IDMixin provides a UUIDv7 primary key field.
type IDMixin struct {
	mixin.Schema
}

// Fields returns the IDMixin field definitions.
func (IDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(NewID).Immutable(),
	}
}

// TimeMixin provides immutable created_at and auto-updated updated_at timestamps.
type TimeMixin struct {
	mixin.Schema
}

// Fields returns the TimeMixin field definitions.
func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// SoftDeleteMixin provides an optional nullable deleted_at timestamp.
type SoftDeleteMixin struct {
	mixin.Schema
}

// Fields returns the SoftDeleteMixin field definitions.
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("deleted_at").Optional().Nillable(),
	}
}
