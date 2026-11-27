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
	noopHandler := func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, map[string]string, error) {
		return nil, map[string]string{"ok": "1"}, nil
	}
	for _, name := range []string{
		"event_list",
		"umat_search",
		"liturgi_list_schedules",
		"liturgi_get_schedule",
		"liturgi_create_schedule",
		"liturgi_set_lagu",
		"cantus_search",
		"cantus_create",
	} {
		mcp.AddTool(srv, &mcp.Tool{Name: name, Description: name}, noopHandler)
	}

	turnCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: u.ID, ParokiID: paroki.ID})
	sess, err := mcpx.ConnectInMemory(turnCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.Close() })

	set, err := service.LoadMCPTools(turnCtx, sess)
	require.NoError(t, err)

	collectNames := func(tools []tool.BaseTool) []string {
		names := make([]string, 0, len(tools))
		for _, tl := range tools {
			info, infoErr := tl.Info(turnCtx)
			require.NoError(t, infoErr)
			names = append(names, info.Name)
		}
		return names
	}

	mikaelNames := collectNames(set.Mikael)
	rafaelNames := collectNames(set.Rafael)
	yosefNames := collectNames(set.Yosef)
	gregoriusNames := collectNames(set.Gregorius)
	sesiliaNames := collectNames(set.Sesilia)

	require.Equal(t, []string{"event_list"}, mikaelNames)

	require.Contains(t, rafaelNames, "ref_get_paroki")
	require.Contains(t, rafaelNames, "user_get_me")
	require.NotContains(t, rafaelNames, "agent_remember")
	require.NotContains(t, rafaelNames, "cantus_search")

	require.Contains(t, yosefNames, "umat_search")
	require.Contains(t, yosefNames, "ref_get_paroki")
	require.Contains(t, yosefNames, "ref_list_wilayah")
	require.Contains(t, yosefNames, "ref_list_lingkungan")
	require.NotContains(t, yosefNames, "user_get_me")

	require.Contains(t, gregoriusNames, "liturgi_list_schedules")
	require.Contains(t, gregoriusNames, "liturgi_get_schedule")
	require.Contains(t, gregoriusNames, "liturgi_create_schedule")
	require.NotContains(t, gregoriusNames, "liturgi_set_lagu")

	require.Contains(t, sesiliaNames, "cantus_search")
	require.Contains(t, sesiliaNames, "cantus_create")
	require.Contains(t, sesiliaNames, "liturgi_list_schedules")
	require.Contains(t, sesiliaNames, "liturgi_get_schedule")
	require.Contains(t, sesiliaNames, "liturgi_set_lagu")
	require.NotContains(t, sesiliaNames, "liturgi_create_schedule")

	var refGetParoki tool.InvokableTool
	for _, tl := range set.Rafael {
		info, infoErr := tl.Info(turnCtx)
		require.NoError(t, infoErr)
		if info.Name == "ref_get_paroki" {
			refGetParoki = tl.(tool.InvokableTool)
			break
		}
	}
	require.NotNil(t, refGetParoki)
	out, err := refGetParoki.InvokableRun(turnCtx, `{}`)
	require.NoError(t, err)
	require.Contains(t, out, paroki.ID.String())
}
