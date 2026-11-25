package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserToken stores Personal Access Tokens (PATs) for a User.
type UserToken struct {
	ent.Schema
}

// Mixin returns the shared mixins for UserToken.
func (UserToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UserToken entity.
func (UserToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.String("name").NotEmpty().MaxLen(100),
		field.String("token_prefix").NotEmpty(),
		field.String("token_hash").NotEmpty().Unique(),
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("expires_at").Optional().Nillable(),
	}
}

// Edges returns the edges of the UserToken entity.
func (UserToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("tokens").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the UserToken entity.
func (UserToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("user_id", "created_at"),
	}
}
