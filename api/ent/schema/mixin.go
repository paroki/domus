package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// UUIDv7Mixin defines a mixin for adding a UUID v7 ID field.
type UUIDv7Mixin struct {
	mixin.Schema
}

// Fields of the UUIDv7Mixin.
func (UUIDv7Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID {
				id, err := uuid.NewV7()
				if err != nil {
					panic(err)
				}
				return id
			}),
	}
}
