package umat_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/umat"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestUmatMCPTools_RBACAndPIIMasking(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	enforcer, err := authz.NewEnforcer(db)
	require.NoError(t, err)
	cipher, err := crypt.New("test-secret-key-at-least-32-bytes!!")
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
	umatFeature := umat.New(db, enforcer, userFeature, refFeature, cipher)

	require.NoError(t, authz.RegisterFeatures(ctx, enforcer, userFeature, refFeature, umatFeature))
	require.NoError(t, userFeature.Bootstrap(ctx))

	umatAdminUser, err := userFeature.EnsureUser(ctx, "uadmin-mcp@paroki.org", "Umat Admin", "", "email", "uadmin-mcp@paroki.org")
	require.NoError(t, err)
	regularUser, err := userFeature.EnsureUser(ctx, "member-mcp@paroki.org", "Regular Member", "", "email", "member-mcp@paroki.org")
	require.NoError(t, err)

	keuskupan, err := db.Keuskupan.Create().SetKode("kaj-mcp").SetNama("KAJ MCP").Save(ctx)
	require.NoError(t, err)
	paroki1, err := db.Paroki.Create().SetKeuskupanID(keuskupan.ID).SetKode("p1-mcp").SetNama("Paroki 1").Save(ctx)
	require.NoError(t, err)

	require.NoError(t, enforcer.AssignRole(ctx, authz.UserSubject(umatAdminUser.ID), umat.RoleAdmin, authz.ParokiDomain(paroki1.ID)))

	plainKK := "3174010101901111"
	plainNIK := "3174010101902222"
	kkEnc, err := cipher.EncryptString(plainKK)
	require.NoError(t, err)
	nikEnc, err := cipher.EncryptString(plainNIK)
	require.NoError(t, err)

	kel, err := db.Keluarga.Create().
		SetParokiID(paroki1.ID).
		SetNomorRegistrasi("KK-2026-000001").
		SetNamaKeluarga("Keluarga MCP").
		SetNomorKkEnc(kkEnc).
		SetNomorKkHash(cipher.BlindIndex(plainKK)).
		Save(ctx)
	require.NoError(t, err)

	uRow, err := db.Umat.Create().
		SetParokiID(paroki1.ID).
		SetNomorInduk("NIU-2026-000001").
		SetKeluargaID(kel.ID).
		SetNama("Budi MCP").
		SetJenisKelamin("pria").
		SetNikEnc(nikEnc).
		SetNikHash(cipher.BlindIndex(plainNIK)).
		Save(ctx)
	require.NoError(t, err)

	srv := mcpx.NewServer("domus", userFeature, refFeature, umatFeature)

	// 1. Missing active paroki -> 400 TENANT_REQUIRED
	noTenantCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: umatAdminUser.ID})
	noTenantSess, err := mcpx.ConnectInMemory(noTenantCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = noTenantSess.Close() })

	res, err := noTenantSess.CallTool(noTenantCtx, &mcp.CallToolParams{Name: "umat_search"})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "An active paroki tenant is required.")

	// 2. Non-umat:admin caller -> 403 Forbidden
	nonAdminCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: regularUser.ID, ParokiID: paroki1.ID})
	nonAdminSess, err := mcpx.ConnectInMemory(nonAdminCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = nonAdminSess.Close() })

	res, err = nonAdminSess.CallTool(nonAdminCtx, &mcp.CallToolParams{Name: "umat_list_keluarga"})
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Contains(t, res.Content[0].(*mcp.TextContent).Text, "Forbidden.")

	// 3. Authorized umat:admin caller -> returns masked PII only, never raw NIK or NomorKK
	adminCtx := authx.WithPrincipal(ctx, authx.Principal{UserID: umatAdminUser.ID, ParokiID: paroki1.ID})
	adminSess, err := mcpx.ConnectInMemory(adminCtx, srv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = adminSess.Close() })

	for _, call := range []*mcp.CallToolParams{
		{Name: "umat_search", Arguments: map[string]any{"query": "Budi"}},
		{Name: "umat_get", Arguments: map[string]any{"umatId": uRow.ID.String()}},
		{Name: "umat_list_keluarga", Arguments: map[string]any{"query": "MCP"}},
		{Name: "umat_get_keluarga", Arguments: map[string]any{"keluargaId": kel.ID.String()}},
	} {
		out, err := adminSess.CallTool(adminCtx, call)
		require.NoError(t, err)
		require.False(t, out.IsError)
		raw := out.Content[0].(*mcp.TextContent).Text
		require.NotContains(t, raw, plainKK, "MCP tool %s must not leak raw nomorKk", call.Name)
		require.NotContains(t, raw, plainNIK, "MCP tool %s must not leak raw nik", call.Name)
		var parsed map[string]any
		require.NoError(t, json.Unmarshal([]byte(raw), &parsed))
		_, hasNik := parsed["nik"]
		_, hasKK := parsed["nomorKk"]
		require.False(t, hasNik)
		require.False(t, hasKK)
	}

	// 4. umat_list_referensi: global + parish merge, regular user allowed, invalid category rejected
	_, err = db.UmatReferensi.Create().
		SetCategory("agama").
		SetCode("katolik").
		SetLabel("Katolik Global").
		SetSortOrder(1).
		SetActive(true).
		Save(ctx)
	require.NoError(t, err)
	_, err = db.UmatReferensi.Create().
		SetParokiID(paroki1.ID).
		SetCategory("agama").
		SetCode("katolik").
		SetLabel("Katolik Paroki").
		SetSortOrder(1).
		SetActive(true).
		Save(ctx)
	require.NoError(t, err)
	_, err = db.UmatReferensi.Create().
		SetParokiID(paroki1.ID).
		SetCategory("agama").
		SetCode("nonaktif").
		SetLabel("Nonaktif").
		SetSortOrder(2).
		SetActive(false).
		Save(ctx)
	require.NoError(t, err)

	refRes, err := nonAdminSess.CallTool(nonAdminCtx, &mcp.CallToolParams{
		Name:      "umat_list_referensi",
		Arguments: map[string]any{"category": "agama"},
	})
	require.NoError(t, err)
	require.False(t, refRes.IsError)
	refText := refRes.Content[0].(*mcp.TextContent).Text
	require.Contains(t, refText, "Katolik Paroki")
	require.NotContains(t, refText, "Katolik Global")
	require.NotContains(t, refText, "Nonaktif")

	badRefRes, err := nonAdminSess.CallTool(nonAdminCtx, &mcp.CallToolParams{
		Name:      "umat_list_referensi",
		Arguments: map[string]any{"category": "invalid_cat"},
	})
	require.NoError(t, err)
	require.True(t, badRefRes.IsError)
}
