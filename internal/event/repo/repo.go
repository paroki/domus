package repo

import (
	"github.com/paroki/domus/ent"
)

// Repository provides Ent-backed persistence for event entities.
type Repository struct {
	client *ent.Client
}

// New creates a Repository backed by the given Ent client.
func New(client *ent.Client) *Repository {
	return &Repository{client: client}
}
