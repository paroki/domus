package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	"github.com/paroki/domus/api/internal/core"
)

type IDV7Mixin struct {
	mixin.Schema
}

func (IDV7Mixin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(core.GenerateID).Immutable(),
	}
}
