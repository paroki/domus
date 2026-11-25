package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
)

const roleParokiAdmin = "ref:paroki-admin"

// UserGetMeInput is the input schema for user_get_me.
type UserGetMeInput struct{}

// MCPUserMe is the MCP-safe representation of the authenticated user's profile.
type MCPUserMe struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	Name           string   `json:"name"`
	Phone          *string  `json:"phone,omitempty"`
	Gender         *string  `json:"gender,omitempty"`
	BirthDate      *string  `json:"birthDate,omitempty"`
	ActiveParokiID *string  `json:"activeParokiId,omitempty"`
	Roles          []string `json:"roles,omitempty"`
}

// UserListParokiMembersInput is the input schema for user_list_paroki_members.
type UserListParokiMembersInput struct {
	Query string `json:"query,omitempty"`
	Role  string `json:"role,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// UserListParokiMembersOutput is the output schema for user_list_paroki_members.
type UserListParokiMembersOutput struct {
	Members []ParokiMemberItem `json:"members"`
}

// RegisterMCP registers user_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "user_get_me",
		Description: "Get the authenticated caller's own profile, active paroki ID, and roles.",
	}, f.mcpUserGetMe)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "user_list_paroki_members",
		Description: "List members of the caller's active paroki (requires user:super-admin or ref:paroki-admin).",
	}, f.mcpUserListParokiMembers)
}

func (f *Feature) mcpUserGetMe(ctx context.Context, _ *mcp.CallToolRequest, _ UserGetMeInput) (*mcp.CallToolResult, MCPUserMe, error) {
	p, ok := UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return nil, MCPUserMe{}, httpx.Unauthorized("Authentication is required.")
	}
	res, err := f.service.GetMe(ctx, p)
	if err != nil {
		return nil, MCPUserMe{}, err
	}
	var activeParokiID *string
	if res.ActiveParokiID != nil {
		s := res.ActiveParokiID.String()
		activeParokiID = &s
	}
	return nil, MCPUserMe{
		ID:             res.ID.String(),
		Email:          res.Email,
		Name:           res.Name,
		Phone:          res.Phone,
		Gender:         res.Gender,
		BirthDate:      res.BirthDate,
		ActiveParokiID: activeParokiID,
		Roles:          res.Roles,
	}, nil
}

func (f *Feature) mcpUserListParokiMembers(ctx context.Context, _ *mcp.CallToolRequest, in UserListParokiMembersInput) (*mcp.CallToolResult, UserListParokiMembersOutput, error) {
	p, ok := UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return nil, UserListParokiMembersOutput{}, httpx.Unauthorized("Authentication is required.")
	}
	tp, err := RequireTenant(ctx)
	if err != nil {
		return nil, UserListParokiMembersOutput{}, err
	}

	sub := authz.UserSubject(tp.UserID)
	isSuperAdmin, err := f.enforcer.HasRoleInDomain(sub, RoleSuperAdmin, authz.DomainGlobal)
	if err != nil {
		return nil, UserListParokiMembersOutput{}, err
	}
	isParokiAdmin, err := f.enforcer.HasRoleInDomain(sub, roleParokiAdmin, authz.ParokiDomain(tp.ParokiID))
	if err != nil {
		return nil, UserListParokiMembersOutput{}, err
	}
	if !isSuperAdmin && !isParokiAdmin {
		return nil, UserListParokiMembersOutput{}, httpx.Forbidden("Forbidden.")
	}
	if f.parokiMembers == nil {
		return nil, UserListParokiMembersOutput{}, errors.New("paroki member directory is not configured")
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	members, err := f.parokiMembers.ListParokiMembers(ctx, tp, tp.ParokiID, in.Query, in.Role, limit)
	if err != nil {
		return nil, UserListParokiMembersOutput{}, err
	}
	return nil, UserListParokiMembersOutput{Members: members}, nil
}
