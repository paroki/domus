package mcpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/app"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/mcpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type echoInput struct {
	Text string `json:"text"`
}

type echoOutput struct {
	Echo   string `json:"echo"`
	UserID string `json:"userId,omitempty"`
}

type customMCPFeature struct{}

func (customMCPFeature) Register(api fiber.Router) {
	api.Get("/custom-ping", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
}

func (customMCPFeature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "custom_echo",
		Description: "Echoes input text and authenticated user ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in echoInput) (*mcp.CallToolResult, echoOutput, error) {
		out := echoOutput{Echo: in.Text}
		if p, ok := authx.UserContext(ctx); ok {
			out.UserID = p.UserID.String()
		}
		return nil, out, nil
	})
}

func mcpPost(t *testing.T, f *fiber.App, sessionID string, payload any) *http.Response {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	resp, err := f.Test(req)
	require.NoError(t, err)
	return resp
}

func initMCPSession(t *testing.T, f *fiber.App) string {
	t.Helper()
	initResp := mcpPost(t, f, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-06-18",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "test-client", "version": "1.0.0"},
		},
	})
	defer func() { _ = initResp.Body.Close() }()
	require.Equal(t, http.StatusOK, initResp.StatusCode)

	sessionID := initResp.Header.Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID)

	notifResp := mcpPost(t, f, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	defer func() { _ = notifResp.Body.Close() }()
	require.Equal(t, http.StatusAccepted, notifResp.StatusCode)

	return sessionID
}

func TestAppGetFiber_MCPDiscoveryAndSPAExclusion(t *testing.T) {
	f := app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler), customMCPFeature{})

	// Verify GET /mcp does NOT fall back to SPA index.html (returns 400 from MCP handler, not 200 HTML).
	getResp, err := f.Test(httptest.NewRequest(http.MethodGet, "/mcp", nil))
	require.NoError(t, err)
	defer func() { _ = getResp.Body.Close() }()
	getBody, _ := io.ReadAll(getResp.Body)
	assert.NotEqual(t, http.StatusOK, getResp.StatusCode)
	assert.False(t, strings.Contains(string(getBody), "<!doctype html"), "SPA fallback must not serve /mcp")

	sessionID := initMCPSession(t, f)

	// tools/list includes both system_ping and custom_echo.
	listResp := mcpPost(t, f, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	defer func() { _ = listResp.Body.Close() }()
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	listRaw, _ := io.ReadAll(listResp.Body)
	assert.Contains(t, string(listRaw), `"system_ping"`)
	assert.Contains(t, string(listRaw), `"custom_echo"`)

	// tools/call system_ping returns pong.
	callResp := mcpPost(t, f, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "system_ping",
			"arguments": map[string]any{},
		},
	})
	defer func() { _ = callResp.Body.Close() }()
	require.Equal(t, http.StatusOK, callResp.StatusCode)
	callRaw, _ := io.ReadAll(callResp.Body)
	assert.Contains(t, string(callRaw), "pong")
}

func TestNewHandler_PrincipalContextPropagation(t *testing.T) {
	uid := uuid.Must(uuid.NewV7())
	srv := mcpx.NewServer("domus", customMCPFeature{})

	f := fiber.New()
	f.All("/mcp", mcpx.NewHandler(srv, func(_ fiber.Ctx) (authx.Principal, bool) {
		return authx.Principal{
			UserID:  uid,
			Email:   "toni@paroki.org",
			Name:    "Toni",
			Channel: "mcp",
		}, true
	}))

	sessionID := initMCPSession(t, f)

	callResp := mcpPost(t, f, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "custom_echo",
			"arguments": map[string]any{"text": "hello"},
		},
	})
	defer func() { _ = callResp.Body.Close() }()
	require.Equal(t, http.StatusOK, callResp.StatusCode)
	raw, _ := io.ReadAll(callResp.Body)
	assert.Contains(t, string(raw), "hello")
	assert.Contains(t, string(raw), uid.String())
}

func TestConnectInMemory_PrincipalPropagation(t *testing.T) {
	uid := uuid.New()
	pid := uuid.New()
	ctx := authx.WithPrincipal(context.Background(), authx.Principal{
		UserID:   uid,
		ParokiID: pid,
	})

	srv := mcpx.NewServer("test-inmem")
	sess, err := mcpx.ConnectInMemory(ctx, srv)
	require.NoError(t, err)
	defer func() { _ = sess.Close() }()

	res, err := sess.CallTool(ctx, &mcp.CallToolParams{
		Name:      "system_ping",
		Arguments: map[string]any{},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.NotEmpty(t, res.Content)
	tc, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, tc.Text, uid.String())
}
