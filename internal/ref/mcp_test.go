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

	keuskupan, err := db.Keuskupan.Create().SetKode("kaj").SetNama("Keuskupan Agung Jakarta").Save(ctx)
	require.NoError(t, err)
	kevikepan1, err := db.Kevikepan.Create().
		SetKeuskupanID(keuskupan.ID).
		SetKode("pusat").
		SetNama("Kevikepan Pusat").
		Save(ctx)
	require.NoError(t, err)
	paroki1, err := db.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p1").SetNama("Paroki Satu").Save(ctx)
	require.NoError(t, err)
	paroki2, err := db.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p2").SetNama("Paroki Dua").Save(ctx)
	require.NoError(t, err)

	wilayah1, err := db.Wilayah.Create().SetParokiID(paroki1.ID).SetKode("z1").SetNama("Wilayah Satu").Save(ctx)
	require.NoError(t, err)
	_, err = db.Wilayah.Create().SetParokiID(paroki2.ID).SetKode("z2").SetNama("Wilayah Dua").Save(ctx)
	require.NoError(t, err)

	_, err = db.Lingkungan.Create().SetParokiID(paroki1.ID).SetWilayahID(wilayah1.ID).SetKode("w1").SetNama("Lingkungan Satu").Save(ctx)
	require.NoError(t, err)

	_, err = db.Period.Create().
		SetParokiID(paroki1.ID).
		SetCode("2026").
		SetName("Tahun 2026").
		SetStartDate("2026-01-01").
		SetEndDate("2026-12-31").
		SetIsCurrent(true).
		Save(ctx)
	require.NoError(t, err)

	_, err = db.Lookup.Create().SetCategory("skill").SetCode("musik").SetLabel("Musik Liturgi").Save(ctx)
	require.NoError(t, err)
	_, err = db.Lookup.Create().SetParokiID(paroki1.ID).SetCategory("skill").SetCode("lektor").SetLabel("Lektor").Save(ctx)
	require.NoError(t, err)
	_, err = db.Lookup.Create().SetParokiID(paroki2.ID).SetCategory("skill").SetCode("other").SetLabel("Other Paroki").Save(ctx)
	require.NoError(t, err)

	srv := mcpx.NewServer("domus", userFeature, refFeature)

	// Without active ParokiID -> TENANT_REQUIRED
	noTenantCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID})
	noTenantSess, err := mcpx.ConnectInMemory(noTenantCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = noTenantSess.Close() })

	res, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{Name: "ref_get_paroki"})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "An active paroki tenant is required.")

	// With active ParokiID -> returns paroki1-scoped data only
	tenantCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID, ParokiID: paroki1.ID})
	sess, err := mcpx.ConnectInMemory(tenantCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sess.Close() })

	parokiRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{Name: "ref_get_paroki"})
	require.NoError(t, err)
	require.False(t, parokiRes.IsError)
	var gotParoki ref.MCPRefParoki
	require.NoError(t, json.Unmarshal([]byte(parokiRes.Content[0].(*mcp.TextContent).Text), &gotParoki))
	require.Equal(t, paroki1.ID.String(), gotParoki.ID)

	wilayahRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{Name: "ref_list_wilayah"})
	require.NoError(t, err)
	require.False(t, wilayahRes.IsError)
	var gotWilayah ref.RefListWilayahOutput
	require.NoError(t, json.Unmarshal([]byte(wilayahRes.Content[0].(*mcp.TextContent).Text), &gotWilayah))
	require.Len(t, gotWilayah.Wilayah, 1)
	require.Equal(t, wilayah1.ID.String(), gotWilayah.Wilayah[0].ID)

	lingkunganRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{
		Name:      "ref_list_lingkungan",
		Arguments: map[string]any{"wilayahId": wilayah1.ID.String()},
	})
	require.NoError(t, err)
	require.False(t, lingkunganRes.IsError)
	var gotLingkungan ref.RefListLingkunganOutput
	require.NoError(t, json.Unmarshal([]byte(lingkunganRes.Content[0].(*mcp.TextContent).Text), &gotLingkungan))
	require.Len(t, gotLingkungan.Lingkungan, 1)

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
		if l.ParokiID != nil {
			require.NotEqual(t, uuid.Nil.String(), *l.ParokiID)
			require.Equal(t, paroki1.ID.String(), *l.ParokiID)
		}
	}

	// ref_list_kevikepan: with explicit keuskupanId works without active tenant.
	kevikepanRes, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{
		Name:      "ref_list_kevikepan",
		Arguments: map[string]any{"keuskupanId": keuskupan.ID.String()},
	})
	require.NoError(t, err)
	require.False(t, kevikepanRes.IsError)
	var gotKevikepan ref.RefListKevikepanOutput
	require.NoError(t, json.Unmarshal([]byte(kevikepanRes.Content[0].(*mcp.TextContent).Text), &gotKevikepan))
	require.Len(t, gotKevikepan.Kevikepan, 1)
	require.Equal(t, kevikepan1.ID.String(), gotKevikepan.Kevikepan[0].ID)

	// ref_list_kevikepan: invalid keuskupanId returns error.
	badKevRes, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{
		Name:      "ref_list_kevikepan",
		Arguments: map[string]any{"keuskupanId": "not-a-uuid"},
	})
	require.NoError(t, err)
	require.True(t, badKevRes.IsError)

	// ref_list_kevikepan: without keuskupanId and without tenant requires active tenant.
	noTenantKevRes, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{Name: "ref_list_kevikepan"})
	require.NoError(t, err)
	require.True(t, noTenantKevRes.IsError)
	require.Contains(t, noTenantKevRes.Content[0].(*mcp.TextContent).Text, "An active paroki tenant is required.")

	// ref_list_kevikepan: without keuskupanId resolves active paroki's keuskupanId.
	tenantKevRes, err := sess.CallTool(tenantCtx, &mcp.CallToolParams{Name: "ref_list_kevikepan"})
	require.NoError(t, err)
	require.False(t, tenantKevRes.IsError)
	var gotTenantKev ref.RefListKevikepanOutput
	require.NoError(t, json.Unmarshal([]byte(tenantKevRes.Content[0].(*mcp.TextContent).Text), &gotTenantKev))
	require.Len(t, gotTenantKev.Kevikepan, 1)
	require.Equal(t, kevikepan1.ID.String(), gotTenantKev.Kevikepan[0].ID)
}
