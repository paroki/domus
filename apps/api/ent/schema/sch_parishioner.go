package schema

import "entgo.io/ent"

// Parishioner holds the schema definition for the Parishioner entity.
type Parishioner struct {
	ent.Schema
}

// Fields of the Parishioner.
func (Parishioner) Fields() []ent.Field {
	return nil
}

// Edges of the Parishioner.
func (Parishioner) Edges() []ent.Edge {
	return nil
}
