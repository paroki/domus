package authz

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
)

const (
	// DomainGlobal represents the wildcard global domain for shared policies.
	DomainGlobal = "*"
)

// Action represents a Casbin permission action.
type Action string

const (
	// ActionRead grants read access to a resource.
	ActionRead Action = "read"
	// ActionWrite grants write access to a resource.
	ActionWrite Action = "write"
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
	return &Enforcer{inner: e}, nil
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
func (e *Enforcer) Enforce(sub, dom, obj string, act Action) (bool, error) {
	return e.inner.Enforce(sub, dom, obj, string(act))
}

// HasRoleInDomain reports whether sub holds role directly in dom.
func (e *Enforcer) HasRoleInDomain(sub, role, dom string) (bool, error) {
	return e.inner.HasGroupingPolicy(sub, role, dom)
}

// RolesForUserInDomain returns all roles held by sub within dom.
func (e *Enforcer) RolesForUserInDomain(sub, dom string) ([]string, error) {
	roles := e.inner.GetRolesForUserInDomain(sub, dom)
	if roles == nil {
		return []string{}, nil
	}
	return roles, nil
}

// UsersForRoleInDomain returns all subjects holding role within dom.
func (e *Enforcer) UsersForRoleInDomain(role, dom string) ([]string, error) {
	users := e.inner.GetUsersForRoleInDomain(role, dom)
	if users == nil {
		return []string{}, nil
	}
	return users, nil
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

// AddPermissionForRole grants role permission to perform act on obj within dom.
func (e *Enforcer) AddPermissionForRole(_ context.Context, role, dom, obj string, act Action) error {
	_, err := e.inner.AddPolicy(role, dom, obj, string(act))
	return err
}

// Require returns a Fiber middleware enforcing that the authenticated principal
// has permission to perform act on obj within the resolved domain (defaulting to DomainGlobal).
func (e *Enforcer) Require(obj string, act Action, domFn ...func(fiber.Ctx) (string, error)) fiber.Handler {
	return func(c fiber.Ctx) error {
		p, ok := authx.UserContext(c.Context())
		if !ok || p.UserID == uuid.Nil {
			return httpx.Unauthorized("authentication required")
		}

		dom := DomainGlobal
		if len(domFn) > 0 && domFn[0] != nil {
			var err error
			dom, err = domFn[0](c)
			if err != nil {
				return err
			}
		}

		allowed, err := e.Enforce(p.Subject(), dom, obj, act)
		if err != nil {
			return httpx.InternalError(err)
		}
		if !allowed {
			return httpx.Forbidden("insufficient permissions")
		}
		return c.Next()
	}
}
