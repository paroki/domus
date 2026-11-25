package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserSession stores active authenticated sessions for a User.
type UserSession struct {
	ent.Schema
}

// Mixin returns the shared mixins for UserSession.
func (UserSession) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UserSession entity.
func (UserSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.String("token_hash").NotEmpty().Unique(),
		field.String("user_agent").Default(""),
		field.String("ip_address").Default(""),
		field.UUID("active_paroki_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("last_active_at").Default(time.Now),
		field.Time("expires_at"),
	}
}

// Edges returns the edges of the UserSession entity.
func (UserSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the UserSession entity.
func (UserSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("user_id", "expires_at"),
	}
}
