package user_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestUserMCPTools(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	enforcer, err := authz.NewEnforcer(db)
	require.NoError(t, err)

	userCfg := user.Config{
		SuperAdminEmail: "super@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	userFeature := user.New(db, enforcer, userCfg, slog.New(slog.DiscardHandler))
	refFeature := ref.New(db, enforcer, userFeature)
	userFeature.SetParokiMemberDirectory(refFeature)

	require.NoError(t, authz.RegisterFeatures(ctx, enforcer, userFeature, refFeature))
	require.NoError(t, userFeature.Bootstrap(ctx))

	superUser, err := userFeature.EnsureUser(ctx, "super@paroki.org", "Super Admin", "", "email", "super@paroki.org")
	require.NoError(t, err)
	parokiAdmin, err := userFeature.EnsureUser(ctx, "admin@paroki.org", "Admin Paroki", "", "email", "admin@paroki.org")
	require.NoError(t, err)
	regularUser, err := userFeature.EnsureUser(ctx, "umat@paroki.org", "Umat Biasa", "", "email", "umat@paroki.org")
	require.NoError(t, err)

	keuskupan, err := db.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	require.NoError(t, err)
	paroki, err := db.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p1").SetNama("Paroki Satu").Save(ctx)
	require.NoError(t, err)

	_, err = refFeature.EnsureParokiMember(ctx, paroki.ID, regularUser.ID)
	require.NoError(t, err)
	_, err = refFeature.EnsureParokiMember(ctx, paroki.ID, parokiAdmin.ID)
	require.NoError(t, err)
	require.NoError(t, enforcer.AssignRole(ctx, authz.UserSubject(parokiAdmin.ID), ref.RoleParokiAdmin, authz.ParokiDomain(paroki.ID)))

	srv := mcpx.NewServer("domus", userFeature, refFeature)

	// 1. user_get_me returns caller's own profile
	meCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID, ParokiID: paroki.ID})
	meSess, err := mcpx.ConnectInMemory(meCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = meSess.Close() })

	meRes, err := meSess.CallTool(meCtx, &mcp.CallToolParams{Name: "user_get_me"})
	require.NoError(t, err)
	require.False(t, meRes.IsError)
	var gotMe user.MCPUserMe
	require.NoError(t, json.Unmarshal([]byte(meRes.Content[0].(*mcp.TextContent).Text), &gotMe))
	require.Equal(t, regularUser.ID.String(), gotMe.ID)
	require.Equal(t, "umat@paroki.org", gotMe.Email)
	require.Equal(t, "Asia/Makassar", gotMe.Timezone)
	require.NotNil(t, gotMe.ActiveParokiID)
	require.Equal(t, paroki.ID.String(), *gotMe.ActiveParokiID)

	_, err = db.User.UpdateOneID(regularUser.ID).SetTimezone("Asia/Jakarta").Save(ctx)
	require.NoError(t, err)
	meRes, err = meSess.CallTool(meCtx, &mcp.CallToolParams{Name: "user_get_me"})
	require.NoError(t, err)
	require.False(t, meRes.IsError)
	require.NoError(t, json.Unmarshal([]byte(meRes.Content[0].(*mcp.TextContent).Text), &gotMe))
	require.Equal(t, "Asia/Jakarta", gotMe.Timezone)

	// 2. user_list_paroki_members without ParokiID -> TENANT_REQUIRED
	noTenantCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: superUser.ID})
	noTenantSess, err := mcpx.ConnectInMemory(noTenantCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = noTenantSess.Close() })

	noTenantRes, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{Name: "user_list_paroki_members"})
	require.NoError(t, err)
	require.True(t, noTenantRes.IsError)
	require.Contains(t, noTenantRes.Content[0].(*mcp.TextContent).Text, "An active paroki tenant is required.")

	// 3. user_list_paroki_members as regular user -> FORBIDDEN
	regRes, err := meSess.CallTool(meCtx, &mcp.CallToolParams{Name: "user_list_paroki_members"})
	require.NoError(t, err)
	require.True(t, regRes.IsError)
	require.Contains(t, regRes.Content[0].(*mcp.TextContent).Text, "Forbidden.")

	// 4. user_list_paroki_members as user:super-admin -> succeeds
	superCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: superUser.ID, ParokiID: paroki.ID})
	superSess, err := mcpx.ConnectInMemory(superCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = superSess.Close() })

	superRes, err := superSess.CallTool(superCtx, &mcp.CallToolParams{Name: "user_list_paroki_members"})
	require.NoError(t, err)
	require.False(t, superRes.IsError)
	var gotMembers user.UserListParokiMembersOutput
	require.NoError(t, json.Unmarshal([]byte(superRes.Content[0].(*mcp.TextContent).Text), &gotMembers))
	require.Len(t, gotMembers.Members, 2)

	// 5. user_list_paroki_members as ref:paroki-admin -> succeeds
	adminCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: parokiAdmin.ID, ParokiID: paroki.ID})
	adminSess, err := mcpx.ConnectInMemory(adminCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = adminSess.Close() })

	adminRes, err := adminSess.CallTool(adminCtx, &mcp.CallToolParams{Name: "user_list_paroki_members"})
	require.NoError(t, err)
	require.False(t, adminRes.IsError)
}
