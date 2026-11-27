package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserPushSubscription stores Web Push browser subscriptions bound to a User and optional UserSession.
type UserPushSubscription struct {
	ent.Schema
}

// Mixin returns the shared mixins for UserPushSubscription.
func (UserPushSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UserPushSubscription entity.
func (UserPushSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("session_id", uuid.UUID{}).Optional().Nillable(),
		field.String("endpoint").NotEmpty().Unique(),
		field.String("p256dh").NotEmpty(),
		field.String("auth").NotEmpty(),
		field.String("user_agent").Default(""),
		field.Time("last_used_at").Optional().Nillable(),
	}
}

// Edges returns the edges of the UserPushSubscription entity.
func (UserPushSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("push_subscriptions").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the UserPushSubscription entity.
func (UserPushSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("endpoint").Unique(),
		index.Fields("user_id"),
		index.Fields("session_id"),
	}
}
