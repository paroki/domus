package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// UserOTP stores one-time passcodes for passwordless authentication.
type UserOTP struct {
	ent.Schema
}

// Annotations returns the schema annotations for UserOTP.
func (UserOTP) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_otps"},
	}
}

// Mixin returns the shared mixins for UserOTP.
func (UserOTP) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the UserOTP entity.
func (UserOTP) Fields() []ent.Field {
	return []ent.Field{
		field.String("channel").Default("email").NotEmpty(),
		field.String("target").NotEmpty(),
		field.String("code_hash").NotEmpty(),
		field.Int("attempts").Default(0),
		field.Time("expires_at"),
	}
}

// Indexes returns the indexes for the UserOTP entity.
func (UserOTP) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel", "target"),
		index.Fields("expires_at"),
	}
}
