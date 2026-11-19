package authz

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
)

const (
	// DomainGlobal represents the wildcard global domain for Super Admin and shared policies.
	DomainGlobal = "*"
	// RoleSuperAdmin is the global Super Admin role identifier.
	RoleSuperAdmin = "user:super-admin"
)

var (
	//go:embed model.conf
	modelConf string

	kebabRe    = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	resourceRe = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
)

// PermissionFeature is implemented by feature facades that register their own Casbin permissions.
type PermissionFeature interface {
	RegisterPermissions(ctx context.Context, e *Enforcer) error
}

// Enforcer wraps Casbin with Domus domain-hierarchy helpers.
type Enforcer struct {
	inner *casbin.Enforcer
}

// UserSubject formats a user UUID as a Casbin subject ("user:<uuid>").
func UserSubject(userID uuid.UUID) string {
	return "user:" + userID.String()
}

// ParishDomain formats a parish UUID as a Casbin domain ("parish:<uuid>").
func ParishDomain(parishID uuid.UUID) string {
	return "parish:" + parishID.String()
}

// OrgDomain formats an organization UUID as a Casbin domain ("org:<uuid>").
func OrgDomain(orgID uuid.UUID) string {
	return "org:" + orgID.String()
}

// Role formats and validates a feature-scoped role ("<feature>:<kebab-role>").
func Role(feature, role string) string {
	if !kebabRe.MatchString(feature) || !kebabRe.MatchString(role) {
		panic(fmt.Sprintf("authz.Role: invalid <feature>:<kebab-role> (%q, %q)", feature, role))
	}
	return feature + ":" + role
}

// Permission formats and validates a feature-scoped resource ("<feature>:<resource>").
func Permission(feature, resource string) string {
	if !kebabRe.MatchString(feature) || !resourceRe.MatchString(resource) {
		panic(fmt.Sprintf("authz.Permission: invalid <feature>:<resource> (%q, %q)", feature, resource))
	}
	return feature + ":" + resource
}

// NewEnforcer creates a Casbin enforcer persisted via Ent's CasbinRule table.
func NewEnforcer(client *ent.Client) (*Enforcer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, fmt.Errorf("parse casbin model: %w", err)
	}
	e, err := casbin.NewEnforcer(m, NewEntAdapter(client))
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}
	registerMatcherFuncs(e)
	return &Enforcer{inner: e}, nil
}

// NewMemoryEnforcer creates an in-memory Casbin enforcer for unit tests.
func NewMemoryEnforcer() (*Enforcer, error) {
	m, err := model.NewModelFromString(modelConf)
	if err != nil {
		return nil, fmt.Errorf("parse casbin model: %w", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("create memory casbin enforcer: %w", err)
	}
	registerMatcherFuncs(e)
	return &Enforcer{inner: e}, nil
}

func registerMatcherFuncs(e *casbin.Enforcer) {
	e.AddFunction("domainMatch", func(args ...any) (any, error) {
		if len(args) != 3 {
			return false, nil
		}
		sub, ok1 := args[0].(string)
		role, ok2 := args[1].(string)
		dom, ok3 := args[2].(string)
		if !ok1 || !ok2 || !ok3 {
			return false, nil
		}

		// Check if dom has any parent domain via g2(dom, parent) where sub holds role in parent.
		g2Rules, err := e.GetNamedGroupingPolicy("g2")
		if err != nil {
			return false, err
		}
		for _, rule := range g2Rules {
			if len(rule) >= 2 && rule[0] == dom {
				has, err := e.HasGroupingPolicy(sub, role, rule[1])
				if err != nil {
					return false, err
				}
				if has {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

// RegisterFeatures calls RegisterPermissions on every feature implementing PermissionFeature.
func RegisterFeatures(ctx context.Context, e *Enforcer, features ...any) error {
	for _, f := range features {
		pf, ok := f.(PermissionFeature)
		if !ok {
			continue
		}
		if err := pf.RegisterPermissions(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// Enforce checks whether sub is allowed to perform act on obj within dom.
func (e *Enforcer) Enforce(sub, dom, obj, act string) (bool, error) {
	return e.inner.Enforce(sub, dom, obj, act)
}

// HasRoleInDomain reports whether sub holds role directly in dom.
func (e *Enforcer) HasRoleInDomain(sub, role, dom string) (bool, error) {
	return e.inner.HasGroupingPolicy(sub, role, dom)
}

// AssignRole assigns role to sub in dom.
func (e *Enforcer) AssignRole(_ context.Context, sub, role, dom string) error {
	_, err := e.inner.AddGroupingPolicy(sub, role, dom)
	return err
}

// RevokeRole revokes role from sub in dom.
func (e *Enforcer) RevokeRole(_ context.Context, sub, role, dom string) error {
	_, err := e.inner.RemoveGroupingPolicy(sub, role, dom)
	return err
}

// LinkOrgToParish records a g2 domain hierarchy link from org:<orgID> to parish:<parishID>.
func (e *Enforcer) LinkOrgToParish(_ context.Context, orgID, parishID uuid.UUID) error {
	_, err := e.inner.AddNamedGroupingPolicy("g2", OrgDomain(orgID), ParishDomain(parishID))
	return err
}

// AddPermissionForRole grants role permission to perform act on obj within dom.
func (e *Enforcer) AddPermissionForRole(_ context.Context, role, dom, obj, act string) error {
	_, err := e.inner.AddPolicy(role, dom, obj, act)
	return err
}
