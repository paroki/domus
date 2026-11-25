package service_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/agent/service"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestLoadMCPTools_RoutingAndTenantIsolation(t *testing.T) {
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

	u, err := userFeature.EnsureUser(ctx, "umat@paroki.org", "Umat", "", "email", "umat@paroki.org")
	require.NoError(t, err)
	keuskupan, err := db.Keuskupan.Create().SetKode("kaj").SetNama("KAJ").Save(ctx)
	require.NoError(t, err)
	paroki, err := db.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p1").SetNama("Paroki Satu").Save(ctx)
	require.NoError(t, err)

	srv := mcpx.NewServer("domus", userFeature, refFeature)
	// Add a cantus_search tool to verify Gabriel allowlist routing
	mcp.AddTool(srv, &mcp.Tool{Name: "cantus_search", Description: "Search hymns"}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, map[string]string, error) {
		return nil, map[string]string{"ok": "1"}, nil
	})

	turnCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: u.ID, ParokiID: paroki.ID})
	sess, err := mcpx.ConnectInMemory(turnCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.Close() })

	gabriel, mikael, rafael, err := service.LoadMCPTools(turnCtx, sess)
	require.NoError(t, err)
	require.Empty(t, mikael) // no event_* registered on srv

	var gabrielNames, rafaelNames []string
	for _, tl := range gabriel {
		info, err := tl.Info(turnCtx)
		require.NoError(t, err)
		gabrielNames = append(gabrielNames, info.Name)
	}
	var refGetParoki tool.InvokableTool
	for _, tl := range rafael {
		info, err := tl.Info(turnCtx)
		require.NoError(t, err)
		rafaelNames = append(rafaelNames, info.Name)
		if info.Name == "ref_get_paroki" {
			refGetParoki = tl.(tool.InvokableTool)
		}
	}

	require.Equal(t, []string{"cantus_search"}, gabrielNames)
	require.Contains(t, rafaelNames, "ref_get_paroki")
	require.Contains(t, rafaelNames, "user_get_me")
	require.NotContains(t, rafaelNames, "agent_remember")

	require.NotNil(t, refGetParoki)
	out, err := refGetParoki.InvokableRun(turnCtx, `{}`)
	require.NoError(t, err)
	require.Contains(t, out, paroki.ID.String())
}
