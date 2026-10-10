package authz

import "github.com/paroki/domus/api/internal/core"

// Can checks if the authenticated user has permission to perform act on obj within their active workspace.
func Can(user core.AuthenticatedUser, obj Resource, act Action) bool {
	enforcer := GetEnforcer()

	for _, role := range user.Roles {
		ok, _ := enforcer.Enforce(string(role), "*", string(obj), string(act))
		if ok {
			return true
		}
	}
	return false
}
