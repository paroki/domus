package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User stores core user profile and account status.
type User struct {
	ent.Schema
}

// Mixin returns the shared mixins for User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
		SoftDeleteMixin{},
	}
}

// Fields returns the fields of the User entity.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().Unique(),
		field.String("name").NotEmpty(),
		field.String("phone").Optional().Nillable(),
		field.String("avatar_url").Optional().Nillable(),
		field.Enum("status").Values("active", "disabled").Default("active"),
		field.Bool("onboarded").Default(false),
		field.UUID("default_parish_id", uuid.UUID{}).Optional().Nillable(),
		field.Time("last_login_at").Optional().Nillable(),
	}
}

// Edges returns the edges of the User entity.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("accounts", UserAccount.Type),
		edge.To("sessions", UserSession.Type),
	}
}

// Indexes returns the indexes for the User entity.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("status"),
	}
}
