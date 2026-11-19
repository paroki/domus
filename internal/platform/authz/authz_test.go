package authz_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
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
		role, dom, obj string
		act            authz.Action
	}{
		{parishAdmin, authz.DomainGlobal, parishObj, authz.ActionWrite},
		{parishAdmin, authz.DomainGlobal, eventObj, authz.ActionWrite},
		{orgAdmin, authz.DomainGlobal, eventObj, authz.ActionWrite},
		{orgMember, authz.DomainGlobal, eventObj, authz.ActionRead},
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

func TestAuthzLeastPrivilegeAndPersistence(t *testing.T) {
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

	// Assign roles directly in their explicit domains.
	require.NoError(t, enforcer.AssignRole(ctx, superAdmin, "user:super-admin", authz.DomainGlobal))
	require.NoError(t, enforcer.AssignRole(ctx, alice, authz.Role("ref", "parish-admin"), authz.ParishDomain(parishA)))
	require.NoError(t, enforcer.AssignRole(ctx, bob, authz.Role("org", "admin"), authz.OrgDomain(orgX)))
	require.NoError(t, enforcer.AssignRole(ctx, carol, authz.Role("org", "member"), authz.OrgDomain(orgX)))

	// No user:super-admin cross-domain or arbitrary action bypass.
	ok, err := enforcer.Enforce(superAdmin, authz.OrgDomain(orgX), "org:events", authz.ActionWrite)
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = enforcer.Enforce(superAdmin, authz.DomainGlobal, "anything:resource", authz.Action("delete"))
	require.NoError(t, err)
	assert.False(t, ok)

	// Parish Admin of parishA is authorized on parishA, denied on parishB and unlinked orgX.
	ok, err = enforcer.Enforce(alice, authz.ParishDomain(parishA), "ref:parishes", authz.ActionWrite)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(alice, authz.ParishDomain(parishB), "ref:parishes", authz.ActionWrite)
	require.NoError(t, err)
	assert.False(t, ok)

	// No implicit parish:<id> -> org:<id> domain inheritance.
	ok, err = enforcer.Enforce(alice, authz.OrgDomain(orgX), "org:events", authz.ActionWrite)
	require.NoError(t, err)
	assert.False(t, ok)

	// Org admin vs org member.
	ok, err = enforcer.Enforce(bob, authz.OrgDomain(orgX), "org:events", authz.ActionWrite)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(carol, authz.OrgDomain(orgX), "org:events", authz.ActionRead)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce(carol, authz.OrgDomain(orgX), "org:events", authz.ActionWrite)
	require.NoError(t, err)
	assert.False(t, ok)

	// Query helpers: RolesForUserInDomain and UsersForRoleInDomain.
	roles, err := enforcer.RolesForUserInDomain(bob, authz.OrgDomain(orgX))
	require.NoError(t, err)
	assert.Equal(t, []string{authz.Role("org", "admin")}, roles)

	emptyRoles, err := enforcer.RolesForUserInDomain(bob, authz.ParishDomain(parishB))
	require.NoError(t, err)
	assert.NotNil(t, emptyRoles)
	assert.Empty(t, emptyRoles)

	users, err := enforcer.UsersForRoleInDomain(authz.Role("org", "member"), authz.OrgDomain(orgX))
	require.NoError(t, err)
	assert.Equal(t, []string{carol}, users)

	emptyUsers, err := enforcer.UsersForRoleInDomain(authz.Role("org", "member"), authz.ParishDomain(parishB))
	require.NoError(t, err)
	assert.NotNil(t, emptyUsers)
	assert.Empty(t, emptyUsers)

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

	ok, err = reloaded.Enforce(alice, authz.ParishDomain(parishA), "ref:parishes", authz.ActionWrite)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = reloaded.Enforce(bob, authz.OrgDomain(orgX), "org:events", authz.ActionWrite)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestNewMemoryEnforcer(t *testing.T) {
	ctx := context.Background()
	e, err := authz.NewMemoryEnforcer()
	require.NoError(t, err)

	uid := uuid.Must(uuid.NewV7())
	require.NoError(t, e.AddPermissionForRole(ctx, "org:member", authz.DomainGlobal, "org:events", authz.ActionRead))
	require.NoError(t, e.AssignRole(ctx, authz.UserSubject(uid), "org:member", authz.DomainGlobal))

	ok, err := e.Enforce(authz.UserSubject(uid), authz.DomainGlobal, "org:events", authz.ActionRead)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRequireMiddleware(t *testing.T) {
	ctx := context.Background()
	e, err := authz.NewMemoryEnforcer()
	require.NoError(t, err)

	userID := uuid.Must(uuid.NewV7())
	parishID := uuid.Must(uuid.NewV7())
	otherParishID := uuid.Must(uuid.NewV7())

	require.NoError(t, e.AddPermissionForRole(ctx, "user:global-reader", authz.DomainGlobal, "user:users", authz.ActionRead))
	require.NoError(t, e.AddPermissionForRole(ctx, "ref:parish-admin", authz.DomainGlobal, "ref:parishes", authz.ActionWrite))

	require.NoError(t, e.AssignRole(ctx, authz.UserSubject(userID), "user:global-reader", authz.DomainGlobal))
	require.NoError(t, e.AssignRole(ctx, authz.UserSubject(userID), "ref:parish-admin", authz.ParishDomain(parishID)))

	app := fiber.New(fiber.Config{
		ErrorHandler: httpx.ErrorHandler(nil),
	})

	// Attach optional principal from X-User-ID header for testing.
	app.Use(func(c fiber.Ctx) error {
		if raw := c.Get("X-User-ID"); raw != "" {
			id := uuid.MustParse(raw)
			c.SetContext(authx.WithPrincipal(c.Context(), authx.Principal{UserID: id}))
		}
		return c.Next()
	})

	app.Get("/users", e.Require("user:users", authz.ActionRead), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Post("/users", e.Require("user:users", authz.ActionWrite), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Post("/parishes/:id", e.Require("ref:parishes", authz.ActionWrite, func(c fiber.Ctx) (string, error) {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return "", httpx.BadRequest("invalid parish id")
		}
		return authz.ParishDomain(id), nil
	}), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// 1. Missing principal -> 401
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// 2. Default global domain allowed -> 200
	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("X-User-ID", userID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 3. Default global domain missing write permission -> 403
	req = httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("X-User-ID", userID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// 4. Custom domFn allowed parish -> 200
	req = httptest.NewRequest(http.MethodPost, "/parishes/"+parishID.String(), nil)
	req.Header.Set("X-User-ID", userID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 5. Custom domFn denied parish -> 403
	req = httptest.NewRequest(http.MethodPost, "/parishes/"+otherParishID.String(), nil)
	req.Header.Set("X-User-ID", userID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	// 6. Custom domFn error -> 400
	req = httptest.NewRequest(http.MethodPost, "/parishes/not-a-uuid", nil)
	req.Header.Set("X-User-ID", userID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
