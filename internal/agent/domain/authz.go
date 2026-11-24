package domain

import "github.com/paroki/domus/internal/platform/authz"

// Casbin permissions for the agent feature.
var (
	PermProviders = authz.Permission("agent", "providers")
	PermChat      = authz.Permission("agent", "chat")
)
