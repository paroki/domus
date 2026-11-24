package ref_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestRefMCPTools(t *testing.T) {
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

	require.NoError(t, authz.RegisterFeatures(ctx, enforcer, userFeature, refFeature))
	require.NoError(t, userFeature.Bootstrap(ctx))

	regularUser, err := userFeature.EnsureUser(ctx, "umat@paroki.org", "Umat Biasa", "", "email", "umat@paroki.org")
	require.NoError(t, err)

	diocese, err := db.Diocese.Create().SetCode("kaj").SetName("Keuskupan Agung Jakarta").Save(ctx)
	require.NoError(t, err)
	parish1, err := db.Parish.Create().SetDioceseID(diocese.ID).SetCode("p1").SetName("Paroki Satu").Save(ctx)
	require.NoError(t, err)
	parish2, err := db.Parish.Create().SetDioceseID(diocese.ID).SetCode("p2").SetName("Paroki Dua").Save(ctx)
	require.NoError(t, err)

	zone1, err := db.Zone.Create().SetParishID(parish1.ID).SetCode("z1").SetName("Wilayah Satu").Save(ctx)
	require.NoError(t, err)
	_, err = db.Zone.Create().SetParishID(parish2.ID).SetCode("z2").SetName("Wilayah Dua").Save(ctx)
	require.NoError(t, err)

	_, err = db.Ward.Create().SetParishID(parish1.ID).SetZoneID(zone1.ID).SetCode("w1").SetName("Lingkungan Satu").Save(ctx)
	require.NoError(t, err)

	_, err = db.Period.Create().
		SetParishID(parish1.ID).
		SetCode("2026").
		SetName("Tahun 2026").
		SetStartDate("2026-01-01").
		SetEndDate("2026-12-31").
		SetIsCurrent(true).
		Save(ctx)
	require.NoError(t, err)

	_, err = db.Lookup.Create().SetCategory("skill").SetCode("musik").SetLabel("Musik Liturgi").Save(ctx)
	require.NoError(t, err)
	_, err = db.Lookup.Create().SetParishID(parish1.ID).SetCategory("skill").SetCode("lektor").SetLabel("Lektor").Save(ctx)
	require.NoError(t, err)
	_, err = db.Lookup.Create().SetParishID(parish2.ID).SetCategory("skill").SetCode("other").SetLabel("Other Parish").Save(ctx)
	require.NoError(t, err)

	srv := mcpx.NewServer("domus", userFeature, refFeature)

	// Without active ParishID -> TENANT_REQUIRED
	noTenantCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID})
	noTenantSess, err := mcpx.ConnectInMemory(noTenantCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = noTenantSess.Close() })

	res, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{Name: "ref_get_parish"})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "An active parish tenant is required.")

	// With active ParishID -> returns parish1-scoped data only
	tenantCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID, ParishID: parish1.ID})
	sess, err := mcpx.ConnectInMemory(tenantCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.Close() })

	parishRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{Name: "ref_get_parish"})
	require.NoError(t, err)
	require.False(t, parishRes.IsError)
	var gotParish ref.MCPRefParish
	require.NoError(t, json.Unmarshal([]byte(parishRes.Content[0].(*mcp.TextContent).Text), &gotParish))
	require.Equal(t, parish1.ID.String(), gotParish.ID)

	zonesRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{Name: "ref_list_zones"})
	require.NoError(t, err)
	require.False(t, zonesRes.IsError)
	var gotZones ref.RefListZonesOutput
	require.NoError(t, json.Unmarshal([]byte(zonesRes.Content[0].(*mcp.TextContent).Text), &gotZones))
	require.Len(t, gotZones.Zones, 1)
	require.Equal(t, zone1.ID.String(), gotZones.Zones[0].ID)

	wardsRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{
		Name:      "ref_list_wards",
		Arguments: map[string]any{"zoneId": zone1.ID.String()},
	})
	require.NoError(t, err)
	require.False(t, wardsRes.IsError)
	var gotWards ref.RefListWardsOutput
	require.NoError(t, json.Unmarshal([]byte(wardsRes.Content[0].(*mcp.TextContent).Text), &gotWards))
	require.Len(t, gotWards.Wards, 1)

	periodsRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{Name: "ref_list_periods"})
	require.NoError(t, err)
	require.False(t, periodsRes.IsError)
	var gotPeriods ref.RefListPeriodsOutput
	require.NoError(t, json.Unmarshal([]byte(periodsRes.Content[0].(*mcp.TextContent).Text), &gotPeriods))
	require.Len(t, gotPeriods.Periods, 1)

	lookupsRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{
		Name:      "ref_list_lookups",
		Arguments: map[string]any{"category": "skill"},
	})
	require.NoError(t, err)
	require.False(t, lookupsRes.IsError)
	var gotLookups ref.RefListLookupsOutput
	require.NoError(t, json.Unmarshal([]byte(lookupsRes.Content[0].(*mcp.TextContent).Text), &gotLookups))
	require.Len(t, gotLookups.Lookups, 2)
	for _, l := range gotLookups.Lookups {
		if l.ParishID != nil {
			require.NotEqual(t, uuid.Nil.String(), *l.ParishID)
			require.Equal(t, parish1.ID.String(), *l.ParishID)
		}
	}
}
