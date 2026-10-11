package authz

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const modelConf = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.obj == p.obj && r.act == p.act
`

// Enforcer wraps a casbin enforcer (RBAC with domain). It is safe for
// concurrent use, so memberships can be added/removed at runtime.
type Enforcer struct {
	e *casbin.SyncedEnforcer
}

// New builds an enforcer with the static policies loaded.
// Memberships are loaded separately (see LoadMemberships).
func New() (*Enforcer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, fmt.Errorf("authz model: %w", err)
	}

	e, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("authz enforcer: %w", err)
	}

	if _, err := e.AddPolicies(policies()); err != nil {
		return nil, fmt.Errorf("authz policies: %w", err)
	}

	return &Enforcer{e: e}, nil
}

// Can reports whether userID may perform act on obj within scope.
func (e *Enforcer) Can(userID string, scope Scope, obj Resource, act Action) (bool, error) {
	return e.e.Enforce(userID, scope.Domain(), string(obj), string(act))
}

// AddMembership registers user -> role in scope. Call after the
// Membership row is committed.
func (e *Enforcer) AddMembership(userID, role string, scope Scope) error {
	_, err := e.e.AddGroupingPolicy(userID, role, scope.Domain())
	return err
}

// RemoveMembership removes user -> role in scope. Call after the
// Membership row deletion is committed.
func (e *Enforcer) RemoveMembership(userID, role string, scope Scope) error {
	_, err := e.e.RemoveGroupingPolicy(userID, role, scope.Domain())
	return err
}
