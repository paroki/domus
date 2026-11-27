package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserNotificationLog stores ephemeral deduplication records for dispatched notifications.
type UserNotificationLog struct {
	ent.Schema
}

// Mixin returns the shared mixins for UserNotificationLog.
func (UserNotificationLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UserNotificationLog entity.
func (UserNotificationLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("paroki_id", uuid.UUID{}).Optional().Nillable(),
		field.String("kind").NotEmpty(),
		field.UUID("ref_id", uuid.UUID{}),
		field.String("window").NotEmpty(),
		field.Time("sent_at"),
	}
}

// Indexes returns the indexes for the UserNotificationLog entity.
func (UserNotificationLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "kind", "ref_id", "window").Unique(),
		index.Fields("sent_at"),
	}
}
