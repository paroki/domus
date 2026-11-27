package cantus_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	agentservice "github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/cantus"
	"github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestCantusMCPTools_RBACAndGabrielRouting(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	enforcer, err := authz.NewEnforcer(db)
	require.NoError(t, err)

	userCfg := user.Config{
		SuperAdminEmail: "super-mcp@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	userFeature := user.New(db, enforcer, userCfg, slog.New(slog.DiscardHandler))
	cantusFeature := cantus.New(db, enforcer, userFeature)

	require.NoError(t, authz.RegisterFeatures(ctx, enforcer, userFeature, cantusFeature))
	require.NoError(t, userFeature.Bootstrap(ctx))

	superUser, err := db.User.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, superUser, 1)
	superAdminID := superUser[0].ID

	cantusAdminUser, err := userFeature.EnsureUser(ctx, "cantus-mcp@paroki.org", "Cantus Admin", "", "email", "cantus-mcp@paroki.org")
	require.NoError(t, err)
	require.NoError(t, enforcer.AssignRole(ctx, authz.UserSubject(cantusAdminUser.ID), cantus.RoleAdmin, authz.DomainGlobal))

	regularUser, err := userFeature.EnsureUser(ctx, "member-mcp@paroki.org", "Regular Member", "", "email", "member-mcp@paroki.org")
	require.NoError(t, err)

	srv := mcpx.NewServer("domus", userFeature, cantusFeature)

	// 1. G8: Super Admin without cantus:admin cannot call cantus_create (403 CANTUS_FORBIDDEN)
	superCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: superAdminID})
	superSess, err := mcpx.ConnectInMemory(superCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = superSess.Close() })

	res, err := superSess.CallTool(superCtx, &mcp.CallToolParams{
		Name: "cantus_create",
		Arguments: map[string]any{
			"slug":  "mcp-forbidden",
			"jenis": "lagu",
			"judul": "Forbidden Song",
		},
	})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "Cantus admin role is required.")

	// 2. cantus:admin can call cantus_create and cantus_update without ParokiID
	adminCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: cantusAdminUser.ID})
	adminSess, err := mcpx.ConnectInMemory(adminCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = adminSess.Close() })

	createRes, err := adminSess.CallTool(adminCtx, &mcp.CallToolParams{
		Name: "cantus_create",
		Arguments: map[string]any{
			"slug":     "ps-900-kristus-raja",
			"jenis":    "lagu",
			"sumber":   "PS",
			"nomor":    "PS-900",
			"judul":    "Kristus Raja Semesta",
			"kategori": "Pembuka",
			"tag":      []string{"kristus-raja", "pembuka"},
		},
	})
	require.NoError(t, err)
	require.False(t, createRes.IsError)

	var created domain.CantusResponse
	require.NoError(t, json.Unmarshal([]byte(createRes.Content[0].(*mcp.TextContent).Text), &created))
	require.Equal(t, "ps-900-kristus-raja", created.Slug)

	// Create an inactive song to verify cantus_search defaults active=true
	inactiveRes, err := adminSess.CallTool(adminCtx, &mcp.CallToolParams{
		Name: "cantus_create",
		Arguments: map[string]any{
			"slug":   "ps-901-inactive",
			"jenis":  "lagu",
			"judul":  "Inactive Song",
			"active": false,
		},
	})
	require.NoError(t, err)
	require.False(t, inactiveRes.IsError)

	// Update via cantus_update
	updateRes, err := adminSess.CallTool(adminCtx, &mcp.CallToolParams{
		Name: "cantus_update",
		Arguments: map[string]any{
			"cantusId": created.ID.String(),
			"judul":    "Kristus Raja Semesta Alam",
		},
	})
	require.NoError(t, err)
	require.False(t, updateRes.IsError)

	var updated domain.CantusResponse
	require.NoError(t, json.Unmarshal([]byte(updateRes.Content[0].(*mcp.TextContent).Text), &updated))
	require.Equal(t, "Kristus Raja Semesta Alam", updated.Judul)

	// 3. Regular user can call cantus_search and cantus_get without ParokiID
	memberCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID})
	memberSess, err := mcpx.ConnectInMemory(memberCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = memberSess.Close() })

	searchRes, err := memberSess.CallTool(memberCtx, &mcp.CallToolParams{
		Name: "cantus_search",
		Arguments: map[string]any{
			"tag": "kristus-raja",
		},
	})
	require.NoError(t, err)
	require.False(t, searchRes.IsError)

	var searchOut cantus.MCPSearchOutput
	require.NoError(t, json.Unmarshal([]byte(searchRes.Content[0].(*mcp.TextContent).Text), &searchOut))
	require.Len(t, searchOut.Cantus, 1)
	require.Equal(t, created.ID.String(), searchOut.Cantus[0].ID)

	// cantus_get by slug
	getBySlugRes, err := memberSess.CallTool(memberCtx, &mcp.CallToolParams{
		Name: "cantus_get",
		Arguments: map[string]any{
			"slug": "ps-900-kristus-raja",
		},
	})
	require.NoError(t, err)
	require.False(t, getBySlugRes.IsError)

	var gotBySlug domain.CantusResponse
	require.NoError(t, json.Unmarshal([]byte(getBySlugRes.Content[0].(*mcp.TextContent).Text), &gotBySlug))
	require.Equal(t, created.ID, gotBySlug.ID)
	require.False(t, gotBySlug.CanManage)

	// 4. Sesilia LoadMCPTools routing: cantus_* tools go to sesilia
	set, err := agentservice.LoadMCPTools(memberCtx, memberSess)
	require.NoError(t, err)
	require.Empty(t, set.Mikael)

	var sesiliaNames []string
	for _, tl := range set.Sesilia {
		info, infoErr := tl.Info(memberCtx)
		require.NoError(t, infoErr)
		sesiliaNames = append(sesiliaNames, info.Name)
	}
	require.ElementsMatch(t, []string{"cantus_search", "cantus_get", "cantus_create", "cantus_update"}, sesiliaNames)

	for _, tl := range set.Rafael {
		info, infoErr := tl.Info(memberCtx)
		require.NoError(t, infoErr)
		require.NotContains(t, []string{"cantus_create", "cantus_update"}, info.Name)
	}
}
