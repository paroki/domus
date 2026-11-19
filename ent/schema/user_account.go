package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserAccount links an external identity provider account to a User.
type UserAccount struct {
	ent.Schema
}

// Mixin returns the shared mixins for UserAccount.
func (UserAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UserAccount entity.
func (UserAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.String("provider").NotEmpty(),
		field.String("subject").NotEmpty(),
		field.String("email").Optional().Nillable(),
		field.String("display_name").Optional().Nillable(),
		field.String("avatar_url").Optional().Nillable(),
		field.UUID("active_parish_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("last_used_at").Optional().Nillable(),
	}
}

// Edges returns the edges of the UserAccount entity.
func (UserAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("accounts").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes returns the indexes for the UserAccount entity.
func (UserAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "subject").Unique(),
		index.Fields("user_id", "provider"),
	}
}
