package authz_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleFeature struct{}

func (sampleFeature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	parishAdmin := authz.Role("ref", "parish-admin")
	orgAdmin := authz.Role("org", "admin")
	orgMember := authz.Role("org", "member")

	parishObj := authz.Permission("ref", "parishes")
	eventObj := authz.Permission("org", "events")

	for _, p := range []struct {
		role, dom, obj, act string
	}{
		{parishAdmin, authz.DomainGlobal, parishObj, "write"},
		{parishAdmin, authz.DomainGlobal, eventObj, "write"},
		{orgAdmin, authz.DomainGlobal, eventObj, "write"},
		{orgMember, authz.DomainGlobal, eventObj, "read"},
	} {
		if err := e.AddPermissionForRole(ctx, p.role, p.dom, p.obj, p.act); err != nil {
			return err
		}
	}
	return nil
}

func TestNamingHelpers(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	assert.Equal(t, "user:"+id.String(), authz.UserSubject(id))
	assert.Equal(t, "parish:"+id.String(), authz.ParishDomain(id))
	assert.Equal(t, "org:"+id.String(), authz.OrgDomain(id))
	assert.Equal(t, "org:parish-admin", authz.Role("org", "parish-admin"))
	assert.Equal(t, "user:profile_read", authz.Permission("user", "profile_read"))

	assert.Panics(t, func() { authz.Role("Org", "admin") })
	assert.Panics(t, func() { authz.Permission("org", "Bad Resource") })
}

func TestAuthzRBACAndPersistence(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	enforcer, err := authz.NewEnforcer(client)
	require.NoError(t, err)

	// RegisterFeatures ignores non-PermissionFeature items and registers sampleFeature.
	require.NoError(t, authz.RegisterFeatures(ctx, enforcer, struct{}{}, sampleFeature{}))

	superAdmin := authz.UserSubject(uuid.Must(uuid.NewV7()))
	alice := authz.UserSubject(uuid.Must(uuid.NewV7()))
	bob := authz.UserSubject(uuid.Must(uuid.NewV7()))
	carol := authz.UserSubject(uuid.Must(uuid.NewV7()))

	parishA := uuid.Must(uuid.NewV7())
	parishB := uuid.Must(uuid.NewV7())
	orgX := uuid.Must(uuid.NewV7())
	orgY := uuid.Must(uuid.NewV7())

	// Assign roles and g2 domain hierarchy.
	require.NoError(t, enforcer.AssignRole(ctx, superAdmin, authz.RoleSuperAdmin, authz.DomainGlobal))
	require.NoError(t, enforcer.AssignRole(ctx, alice, authz.Role("ref", "parish-admin"), authz.ParishDomain(parishA)))
	require.NoError(t, enforcer.AssignRole(ctx, bob, authz.Role("org", "admin"), authz.OrgDomain(orgX)))
	require.NoError(t, enforcer.AssignRole(ctx, carol, authz.Role("org", "member"), authz.OrgDomain(orgX)))

	require.NoError(t, enforcer.LinkOrgToParish(ctx, orgX, parishA))
	require.NoError(t, enforcer.LinkOrgToParish(ctx, orgY, parishB))

	// Super Admin bypasses all domain/resource checks.
	ok, err := enforcer.Enforce(superAdmin, authz.OrgDomain(orgY), "anything:resource", "delete")
	require.NoError(t, err)
	assert.True(t, ok)

	// Parish Admin of parishA is authorized on parishA and child orgX, denied on parishB and orgY.
	ok, err = enforcer.Enforce(alice, authz.ParishDomain(parishA), "ref:parishes", "write")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(alice, authz.OrgDomain(orgX), "org:events", "write")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(alice, authz.ParishDomain(parishB), "ref:parishes", "write")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = enforcer.Enforce(alice, authz.OrgDomain(orgY), "org:events", "write")
	require.NoError(t, err)
	assert.False(t, ok)

	// Org admin vs org member.
	ok, err = enforcer.Enforce(bob, authz.OrgDomain(orgX), "org:events", "write")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(carol, authz.OrgDomain(orgX), "org:events", "read")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(carol, authz.OrgDomain(orgX), "org:events", "write")
	require.NoError(t, err)
	assert.False(t, ok)

	// HasRoleInDomain and RevokeRole.
	has, err := enforcer.HasRoleInDomain(bob, authz.Role("org", "admin"), authz.OrgDomain(orgX))
	require.NoError(t, err)
	assert.True(t, has)

	require.NoError(t, enforcer.RevokeRole(ctx, bob, authz.Role("org", "admin"), authz.OrgDomain(orgX)))
	has, err = enforcer.HasRoleInDomain(bob, authz.Role("org", "admin"), authz.OrgDomain(orgX))
	require.NoError(t, err)
	assert.False(t, has)

	// Reload enforcer from the same Ent client and verify persistence.
	reloaded, err := authz.NewEnforcer(client)
	require.NoError(t, err)

	ok, err = reloaded.Enforce(alice, authz.OrgDomain(orgX), "org:events", "write")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = reloaded.Enforce(bob, authz.OrgDomain(orgX), "org:events", "write")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestNewMemoryEnforcer(t *testing.T) {
	ctx := context.Background()
	e, err := authz.NewMemoryEnforcer()
	require.NoError(t, err)

	uid := uuid.Must(uuid.NewV7())
	require.NoError(t, e.AddPermissionForRole(ctx, "org:member", authz.DomainGlobal, "org:events", "read"))
	require.NoError(t, e.AssignRole(ctx, authz.UserSubject(uid), "org:member", authz.DomainGlobal))

	ok, err := e.Enforce(authz.UserSubject(uid), authz.DomainGlobal, "org:events", "read")
	require.NoError(t, err)
	assert.True(t, ok)
}
