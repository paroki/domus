package repo

import "github.com/paroki/domus/ent"

// Repository provides Ent data access for the umat feature.
type Repository struct {
	client *ent.Client
}

// New constructs a Repository backed by client.
func New(client *ent.Client) *Repository {
	return &Repository{client: client}
}
