package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// AgentProviderModel describes a cached model entry stored on an AgentProvider record.
type AgentProviderModel struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// AgentProvider stores a configured LLM provider credential and its discovered models.
type AgentProvider struct {
	ent.Schema
}

// Annotations returns the table annotations for AgentProvider.
func (AgentProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "agent_providers"},
	}
}

// Mixin returns the shared mixins for AgentProvider.
func (AgentProvider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		IDMixin{},
		TimeMixin{},
	}
}

// Fields returns the fields of the AgentProvider entity.
func (AgentProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Default(""),
		field.Enum("provider").Values("gemini", "anthropic", "openai", "openrouter"),
		field.String("api_key").NotEmpty(),
		field.String("base_url").Default(""),
		field.JSON("models", []AgentProviderModel{}).Default([]AgentProviderModel{}),
	}
}
