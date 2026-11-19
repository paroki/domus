package authx_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authx"
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
	assert.Empty(t, noTenant.ParishDomain())

	withTenant := authx.Principal{
		UserID:   uid,
		ParishID: pid,
		Roles:    []string{"ref:parish-admin"},
		Channel:  "mcp",
	}
	assert.True(t, withTenant.HasTenant())
	assert.Equal(t, "parish:"+pid.String(), withTenant.ParishDomain())
}

func TestPrincipalContext(t *testing.T) {
	_, ok := authx.UserContext(context.Background())
	assert.False(t, ok)

	//lint:ignore SA1012 verify nil context guard
	_, ok = authx.UserContext(nil) //nolint:staticcheck // testing nil context guard
	assert.False(t, ok)

	p := authx.Principal{
		UserID:  uuid.Must(uuid.NewV7()),
		Email:   "toni@paroki.org",
		Name:    "Toni",
		Channel: "web",
	}
	ctx := authx.WithPrincipal(context.Background(), p)
	got, ok := authx.UserContext(ctx)
	require.True(t, ok)
	assert.Equal(t, p, got)
}
