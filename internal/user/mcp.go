package user

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
)

const roleParishAdmin = "ref:parish-admin"

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
	ActiveParishID *string  `json:"activeParishId,omitempty"`
	Roles          []string `json:"roles,omitempty"`
}

// UserListParishMembersInput is the input schema for user_list_parish_members.
type UserListParishMembersInput struct {
	Query string `json:"query,omitempty"`
	Role  string `json:"role,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// UserListParishMembersOutput is the output schema for user_list_parish_members.
type UserListParishMembersOutput struct {
	Members []ParishMemberItem `json:"members"`
}

// RegisterMCP registers user_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "user_get_me",
		Description: "Get the authenticated caller's own profile, active parish ID, and roles.",
	}, f.mcpUserGetMe)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "user_list_parish_members",
		Description: "List members of the caller's active parish (requires user:super-admin or ref:parish-admin).",
	}, f.mcpUserListParishMembers)
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
	var activeParishID *string
	if res.ActiveParishID != nil {
		s := res.ActiveParishID.String()
		activeParishID = &s
	}
	return nil, MCPUserMe{
		ID:             res.ID.String(),
		Email:          res.Email,
		Name:           res.Name,
		Phone:          res.Phone,
		Gender:         res.Gender,
		BirthDate:      res.BirthDate,
		ActiveParishID: activeParishID,
		Roles:          res.Roles,
	}, nil
}

func (f *Feature) mcpUserListParishMembers(ctx context.Context, _ *mcp.CallToolRequest, in UserListParishMembersInput) (*mcp.CallToolResult, UserListParishMembersOutput, error) {
	p, ok := UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return nil, UserListParishMembersOutput{}, httpx.Unauthorized("Authentication is required.")
	}
	tp, err := RequireTenant(ctx)
	if err != nil {
		return nil, UserListParishMembersOutput{}, err
	}

	sub := authz.UserSubject(tp.UserID)
	isSuperAdmin, err := f.enforcer.HasRoleInDomain(sub, RoleSuperAdmin, authz.DomainGlobal)
	if err != nil {
		return nil, UserListParishMembersOutput{}, err
	}
	isParishAdmin, err := f.enforcer.HasRoleInDomain(sub, roleParishAdmin, authz.ParishDomain(tp.ParishID))
	if err != nil {
		return nil, UserListParishMembersOutput{}, err
	}
	if !isSuperAdmin && !isParishAdmin {
		return nil, UserListParishMembersOutput{}, httpx.Forbidden("Forbidden.")
	}
	if f.parishMembers == nil {
		return nil, UserListParishMembersOutput{}, errors.New("parish member directory is not configured")
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	members, err := f.parishMembers.ListParishMembers(ctx, tp, tp.ParishID, in.Query, in.Role, limit)
	if err != nil {
		return nil, UserListParishMembersOutput{}, err
	}
	return nil, UserListParishMembersOutput{Members: members}, nil
}
