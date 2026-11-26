package domain

import "github.com/paroki/domus/internal/platform/authz"

// Feature-owned Casbin roles and resource permissions.
var (
	RoleAdmin    = authz.Role("umat", "admin")
	PermKeluarga = authz.Permission("umat", "keluarga")
	PermUmat     = authz.Permission("umat", "umat")
)
