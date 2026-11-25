package authx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrincipalMethods(t *testing.T) {
	uid := uuid.Must(uuid.NewV7())
	pid := uuid.Must(uuid.NewV7())

	noTenant := authx.Principal{
		UserID:  uid,
		Email:   "admin@paroki.org",
		Name:    "Admin",
		Channel: "web",
	}
	assert.Equal(t, "user:"+uid.String(), noTenant.Subject())
	assert.False(t, noTenant.HasTenant())
	assert.Empty(t, noTenant.ParokiDomain())

	withTenant := authx.Principal{
		UserID:   uid,
		ParokiID: pid,
		Roles:    []string{"ref:paroki-admin"},
		Channel:  "mcp",
	}
	assert.True(t, withTenant.HasTenant())
	assert.Equal(t, "paroki:"+pid.String(), withTenant.ParokiDomain())
}

func TestPrincipalContext(t *testing.T) {
	_, ok := authx.UserContext(context.Background())
	assert.False(t, ok)

	//lint:ignore SA1012 verify nil context guard
	_, ok = authx.UserContext(nil) //nolint:staticcheck // testing nil context guard
	assert.False(t, ok)

	p := authx.Principal{
		UserID:    uuid.Must(uuid.NewV7()),
		SessionID: uuid.Must(uuid.NewV7()),
		Email:     "toni@paroki.org",
		Name:      "Toni",
		Channel:   "web",
	}
	ctx := authx.WithPrincipal(context.Background(), p)
	got, ok := authx.UserContext(ctx)
	require.True(t, ok)
	assert.Equal(t, p, got)

	got2, ok := authx.FromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, p, got2)
}

func TestRequireTenant(t *testing.T) {
	_, err := authx.RequireTenant(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, httpx.ErrUnauthorized))

	uid := uuid.Must(uuid.NewV7())
	noTenantCtx := authx.WithPrincipal(context.Background(), authx.Principal{UserID: uid})
	_, err = authx.RequireTenant(noTenantCtx)
	require.ErrorIs(t, err, authx.ErrTenantRequired)

	pid := uuid.Must(uuid.NewV7())
	tenantCtx := authx.WithPrincipal(context.Background(), authx.Principal{UserID: uid, ParokiID: pid})
	p, err := authx.RequireTenant(tenantCtx)
	require.NoError(t, err)
	assert.Equal(t, uid, p.UserID)
	assert.Equal(t, pid, p.ParokiID)
}
