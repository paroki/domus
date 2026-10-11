package config

import (
	"context"
	"log"

	"github.com/paroki/domus/api/ent"
	"github.com/paroki/domus/api/internal/platform/authz"
)

// GetAuthz builds the enforcer and loads memberships from the database.
func GetAuthz(entcli *ent.Client) *authz.Enforcer {
	e, err := authz.New()
	if err != nil {
		log.Fatalf("failed creating authz enforcer: %v", err)
	}
	if err := authz.LoadMemberships(context.Background(), entcli, e); err != nil {
		log.Fatalf("failed loading memberships: %v", err)
	}
	return e
}
