package mcpx

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authx"
)

// MCPFeature is implemented by feature facades that expose MCP tools.
type MCPFeature interface {
	RegisterMCP(s *mcp.Server)
}

// PrincipalResolver extracts an authenticated Principal from a Fiber request context.
type PrincipalResolver func(c fiber.Ctx) (authx.Principal, bool)

// PingInput is the empty input schema for the built-in system_ping tool.
type PingInput struct{}

// PingOutput is returned by the built-in system_ping tool.
type PingOutput struct {
	Message string `json:"message"`
	UserID  string `json:"userId,omitempty"`
}

// NewServer initializes an MCP server, registers system_ping, and wires any MCPFeature items.
func NewServer(appName string, features ...any) *mcp.Server {
	if appName == "" {
		appName = "domus"
	}
	s := mcp.NewServer(&mcp.Implementation{
		Name:    appName,
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "system_ping",
		Description: "Diagnostic ping tool returning pong and the authenticated user ID if present.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ PingInput) (*mcp.CallToolResult, PingOutput, error) {
		out := PingOutput{Message: "pong"}
		if p, ok := authx.UserContext(ctx); ok && p.UserID != uuid.Nil {
			out.UserID = p.UserID.String()
		}
		return nil, out, nil
	})

	RegisterFeatures(s, features...)
	return s
}

// RegisterFeatures calls RegisterMCP on every feature implementing MCPFeature.
func RegisterFeatures(s *mcp.Server, features ...any) {
	for _, f := range features {
		if mf, ok := f.(MCPFeature); ok {
			mf.RegisterMCP(s)
		}
	}
}

// NewHandler wraps an MCP Streamable HTTP handler for Fiber v3 and propagates authx.Principal.
func NewHandler(server *mcp.Server, resolver PrincipalResolver) fiber.Handler {
	httpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		JSONResponse: true,
	})
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if localCtx, ok := adaptor.LocalContextFromHTTPRequest(r); ok {
			if p, ok := authx.UserContext(localCtx); ok {
				r = r.WithContext(authx.WithPrincipal(r.Context(), p))
			}
		}
		httpHandler.ServeHTTP(w, r)
	})
	adapted := adaptor.HTTPHandlerWithContext(wrapped)

	return func(c fiber.Ctx) error {
		if resolver != nil {
			if p, ok := resolver(c); ok {
				c.SetContext(authx.WithPrincipal(c.Context(), p))
			}
		}
		return adapted(c)
	}
}

// ConnectInMemory connects an in-process MCP client to server, binding ctx (and any authx.Principal)
// as the parent context for every tool invocation.
func ConnectInMemory(ctx context.Context, server *mcp.Server) (*mcp.ClientSession, error) {
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		return nil, err
	}
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "domus-agent",
		Version: "0.1.0",
	}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		_ = serverSession.Close()
		return nil, err
	}
	return clientSession, nil
}
