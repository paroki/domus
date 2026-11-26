package org

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user"
)

// OrgListInput is the input schema for org_list.
type OrgListInput struct {
	Query      string `json:"query,omitempty"`
	Kind       string `json:"kind,omitempty"`
	MyOrgsOnly bool   `json:"myOrgsOnly,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// MCPOrganization is the MCP-safe representation of an Organization (UUIDs as strings).
type MCPOrganization struct {
	ID           string    `json:"id"`
	ParokiID     string    `json:"parokiId"`
	ParentID     *string   `json:"parentId,omitempty"`
	WilayahID    *string   `json:"wilayahId,omitempty"`
	LingkunganID *string   `json:"lingkunganId,omitempty"`
	Kind         string    `json:"kind"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	Active       bool      `json:"active"`
	MyRole       *string   `json:"myRole,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// OrgListOutput is the output schema for org_list.
type OrgListOutput struct {
	Organizations []MCPOrganization `json:"organizations"`
}

// OrgGetInput is the input schema for org_get.
type OrgGetInput struct {
	OrgID string `json:"orgId"`
}

// OrgCreateInput is the input schema for org_create.
type OrgCreateInput struct {
	ParentID           *string `json:"parentId,omitempty"`
	WilayahID          *string `json:"wilayahId,omitempty"`
	LingkunganID       *string `json:"lingkunganId,omitempty"`
	Kind               string  `json:"kind,omitempty"`
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	Active             *bool   `json:"active,omitempty"`
	InitialAdminUserID *string `json:"initialAdminUserId,omitempty"`
	InitialAdminEmail  string  `json:"initialAdminEmail,omitempty"`
	InitialAdminName   string  `json:"initialAdminName,omitempty"`
}

// OrgListMembersInput is the input schema for org_list_members.
type OrgListMembersInput struct {
	OrgID  string `json:"orgId"`
	Query  string `json:"query,omitempty"`
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// MCPOrgMember is the MCP-safe representation of an OrgMember (UUIDs as strings).
type MCPOrgMember struct {
	ID        string    `json:"id"`
	ParokiID  string    `json:"parokiId"`
	OrgID     string    `json:"orgId"`
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// OrgListMembersOutput is the output schema for org_list_members.
type OrgListMembersOutput struct {
	Members []MCPOrgMember `json:"members"`
}

// OrgUpdateMemberInput is the input schema for org_update_member.
type OrgUpdateMemberInput struct {
	OrgID  string  `json:"orgId"`
	UserID string  `json:"userId,omitempty"`
	Email  string  `json:"email,omitempty"`
	Name   string  `json:"name,omitempty"`
	Phone  *string `json:"phone,omitempty"`
	Role   string  `json:"role,omitempty"`
	Status *string `json:"status,omitempty"`
}

// OrgCreateInviteInput is the input schema for org_create_invite.
type OrgCreateInviteInput struct {
	OrgID      string `json:"orgId"`
	Role       string `json:"role,omitempty"`
	MaxUses    *int   `json:"maxUses,omitempty"`
	TTLSeconds *int   `json:"ttlSeconds,omitempty"`
}

// MCPOrgInvite is the MCP-safe representation of an OrgInvite (UUIDs as strings).
type MCPOrgInvite struct {
	ID        string     `json:"id"`
	ParokiID  string     `json:"parokiId"`
	OrgID     string     `json:"orgId"`
	Code      string     `json:"code"`
	Role      string     `json:"role"`
	MaxUses   *int       `json:"maxUses,omitempty"`
	UsedCount int        `json:"usedCount"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedBy string     `json:"createdBy"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// OrgListPeriodeInput is the input schema for org_list_periode.
type OrgListPeriodeInput struct {
	OrgID     string `json:"orgId"`
	Query     string `json:"query,omitempty"`
	Status    string `json:"status,omitempty"`
	IsCurrent *bool  `json:"isCurrent,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// MCPOrgPeriode is the MCP-safe representation of an OrgPeriode (UUIDs as strings).
type MCPOrgPeriode struct {
	ID             string    `json:"id"`
	ParokiID       string    `json:"parokiId"`
	OrgID          string    `json:"orgId"`
	Kode           string    `json:"kode"`
	Nama           string    `json:"nama"`
	TanggalMulai   string    `json:"tanggalMulai"`
	TanggalSelesai string    `json:"tanggalSelesai"`
	Status         string    `json:"status"`
	IsCurrent      bool      `json:"isCurrent"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// OrgListPeriodeOutput is the output schema for org_list_periode.
type OrgListPeriodeOutput struct {
	Periode []MCPOrgPeriode `json:"periode"`
}

// OrgRevokeInviteInput is the input schema for org_revoke_invite.
type OrgRevokeInviteInput struct {
	OrgID    string `json:"orgId"`
	InviteID string `json:"inviteId"`
}

// OrgRevokeInviteOutput is the output schema for org_revoke_invite.
type OrgRevokeInviteOutput struct {
	Revoked bool `json:"revoked"`
}

// OrgRedeemInviteInput is the input schema for org_redeem_invite.
type OrgRedeemInviteInput struct {
	Code string `json:"code"`
}

// MCPRedeemOrgInviteOutput is the MCP-safe representation of RedeemOrgInviteResponse.
type MCPRedeemOrgInviteOutput struct {
	Organization MCPOrganization `json:"organization"`
	Membership   MCPOrgMember    `json:"membership"`
	ParokiID     string          `json:"parokiId"`
}

// RegisterMCP registers all org_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_list",
		Description: "List organizations in the caller's active paroki.",
	}, f.mcpOrgList)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_get",
		Description: "Get details of a single organization in the caller's active paroki.",
	}, f.mcpOrgGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_create",
		Description: "Create an organization (or reactivate a soft-deleted one) in the caller's active paroki (requires ref:paroki-admin).",
	}, f.mcpOrgCreate)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_list_members",
		Description: "List members of an organization (requires org:admin or org:member).",
	}, f.mcpOrgListMembers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_update_member",
		Description: "Add, onboard, or update an organization member (requires org:admin, or ref:paroki-admin for admin appointment).",
	}, f.mcpOrgUpdateMember)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_create_invite",
		Description: "Create a time-limited multi-use invite link for an organization (requires org:admin).",
	}, f.mcpOrgCreateInvite)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_revoke_invite",
		Description: "Revoke an organization invite link (requires org:admin).",
	}, f.mcpOrgRevokeInvite)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_redeem_invite",
		Description: "Redeem an organization invite code to join the organization and its paroki.",
	}, f.mcpOrgRedeemInvite)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "org_list_periode",
		Description: "List periods of an organization (requires org:admin, org:member, or ref:paroki-admin).",
	}, f.mcpOrgListPeriode)
}

func (f *Feature) mcpOrgList(ctx context.Context, _ *mcp.CallToolRequest, in OrgListInput) (*mcp.CallToolResult, OrgListOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, OrgListOutput{}, err
	}
	var filter *domain.OrganizationSearchFilter
	if in.Kind != "" || in.MyOrgsOnly {
		filter = &domain.OrganizationSearchFilter{Kind: in.Kind}
		if in.MyOrgsOnly {
			t := true
			filter.MyOrgsOnly = &t
		}
	}
	items, _, err := f.service.SearchOrganizations(
		ctx,
		p,
		domain.SearchOrganizationsRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit),
	)
	if err != nil {
		return nil, OrgListOutput{}, err
	}
	out := make([]MCPOrganization, 0, len(items))
	for _, item := range items {
		out = append(out, toMCPOrganization(item))
	}
	return nil, OrgListOutput{Organizations: out}, nil
}

func (f *Feature) mcpOrgGet(ctx context.Context, _ *mcp.CallToolRequest, in OrgGetInput) (*mcp.CallToolResult, MCPOrganization, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, MCPOrganization{}, httpx.BadRequest("Invalid organization ID.")
	}
	res, err := f.service.GetOrganizationForCaller(ctx, p, orgID)
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	return nil, toMCPOrganization(res), nil
}

func parseOptionalUUID(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(*raw)
	if err != nil {
		return nil, httpx.BadRequest("Invalid " + field + ".")
	}
	return &parsed, nil
}

func (f *Feature) mcpOrgCreate(ctx context.Context, _ *mcp.CallToolRequest, in OrgCreateInput) (*mcp.CallToolResult, MCPOrganization, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	parentID, err := parseOptionalUUID(in.ParentID, "parentId")
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	wilayahID, err := parseOptionalUUID(in.WilayahID, "wilayahId")
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	lingkunganID, err := parseOptionalUUID(in.LingkunganID, "lingkunganId")
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	initialAdminUserID, err := parseOptionalUUID(in.InitialAdminUserID, "initialAdminUserId")
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	res, err := f.service.CreateOrganization(ctx, p, domain.CreateOrganizationRequest{
		ParentID:           parentID,
		WilayahID:          wilayahID,
		LingkunganID:       lingkunganID,
		Kind:               in.Kind,
		Code:               in.Code,
		Name:               in.Name,
		Description:        in.Description,
		Active:             in.Active,
		InitialAdminUserID: initialAdminUserID,
		InitialAdminEmail:  in.InitialAdminEmail,
		InitialAdminName:   in.InitialAdminName,
	})
	if err != nil {
		return nil, MCPOrganization{}, err
	}
	return nil, toMCPOrganization(res), nil
}

func (f *Feature) mcpOrgListMembers(ctx context.Context, _ *mcp.CallToolRequest, in OrgListMembersInput) (*mcp.CallToolResult, OrgListMembersOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, OrgListMembersOutput{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, OrgListMembersOutput{}, httpx.BadRequest("Invalid organization ID.")
	}
	var filter *domain.OrgMemberSearchFilter
	if in.Role != "" || in.Status != "" {
		filter = &domain.OrgMemberSearchFilter{Role: in.Role, Status: in.Status}
	}
	items, _, err := f.service.SearchOrgMembers(
		ctx,
		p,
		orgID,
		domain.SearchOrgMembersRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit),
	)
	if err != nil {
		return nil, OrgListMembersOutput{}, err
	}
	out := make([]MCPOrgMember, 0, len(items))
	for _, item := range items {
		out = append(out, toMCPOrgMember(item))
	}
	return nil, OrgListMembersOutput{Members: out}, nil
}

func (f *Feature) mcpOrgUpdateMember(ctx context.Context, _ *mcp.CallToolRequest, in OrgUpdateMemberInput) (*mcp.CallToolResult, MCPOrgMember, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPOrgMember{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, MCPOrgMember{}, httpx.BadRequest("Invalid organization ID.")
	}

	var userID *uuid.UUID
	if in.UserID != "" {
		parsed, err := uuid.Parse(in.UserID)
		if err != nil {
			return nil, MCPOrgMember{}, httpx.BadRequest("Invalid user ID.")
		}
		userID = &parsed

		var rolePtr *string
		if in.Role != "" {
			r := in.Role
			rolePtr = &r
		}
		res, err := f.service.UpdateOrgMember(ctx, p, orgID, parsed, domain.UpdateOrgMemberRequest{
			Role:   rolePtr,
			Status: in.Status,
		})
		if err == nil {
			return nil, toMCPOrgMember(res), nil
		}
		if !errors.Is(err, domain.ErrMemberNotFound) {
			return nil, MCPOrgMember{}, err
		}
	}

	var status string
	if in.Status != nil {
		status = *in.Status
	}
	res, err := f.service.AddOrgMember(ctx, p, orgID, domain.AddOrgMemberRequest{
		UserID: userID,
		Email:  in.Email,
		Name:   in.Name,
		Phone:  in.Phone,
		Role:   in.Role,
		Status: status,
	})
	if err != nil {
		return nil, MCPOrgMember{}, err
	}
	return nil, toMCPOrgMember(res), nil
}

func (f *Feature) mcpOrgCreateInvite(ctx context.Context, _ *mcp.CallToolRequest, in OrgCreateInviteInput) (*mcp.CallToolResult, MCPOrgInvite, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPOrgInvite{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, MCPOrgInvite{}, httpx.BadRequest("Invalid organization ID.")
	}
	res, err := f.service.CreateOrgInvite(ctx, p, orgID, domain.CreateOrgInviteRequest{
		Role:       in.Role,
		MaxUses:    in.MaxUses,
		TTLSeconds: in.TTLSeconds,
	})
	if err != nil {
		return nil, MCPOrgInvite{}, err
	}
	return nil, toMCPOrgInvite(res), nil
}

func (f *Feature) mcpOrgRevokeInvite(ctx context.Context, _ *mcp.CallToolRequest, in OrgRevokeInviteInput) (*mcp.CallToolResult, OrgRevokeInviteOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, OrgRevokeInviteOutput{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, OrgRevokeInviteOutput{}, httpx.BadRequest("Invalid organization ID.")
	}
	inviteID, err := uuid.Parse(in.InviteID)
	if err != nil {
		return nil, OrgRevokeInviteOutput{}, httpx.BadRequest("Invalid invite ID.")
	}
	if err := f.service.RevokeOrgInvite(ctx, p, orgID, inviteID); err != nil {
		return nil, OrgRevokeInviteOutput{}, err
	}
	return nil, OrgRevokeInviteOutput{Revoked: true}, nil
}

func (f *Feature) mcpOrgRedeemInvite(ctx context.Context, _ *mcp.CallToolRequest, in OrgRedeemInviteInput) (*mcp.CallToolResult, MCPRedeemOrgInviteOutput, error) {
	p, ok := user.UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return nil, MCPRedeemOrgInviteOutput{}, httpx.Unauthorized("Authentication is required.")
	}
	res, err := f.service.RedeemInvite(ctx, p, in.Code)
	if err != nil {
		return nil, MCPRedeemOrgInviteOutput{}, err
	}
	return nil, MCPRedeemOrgInviteOutput{
		Organization: toMCPOrganization(res.Organization),
		Membership:   toMCPOrgMember(res.Membership),
		ParokiID:     res.ParokiID.String(),
	}, nil
}

func requireMCPTenant(ctx context.Context) (user.Principal, error) {
	p, ok := user.UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return user.Principal{}, httpx.Unauthorized("Authentication is required.")
	}
	return user.RequireTenant(ctx)
}

func clampMCPLimit(limit int) int {
	if limit <= 0 {
		return httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		return httpx.MaxLimit
	}
	return limit
}

func uuidPtrStr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func toMCPOrganization(o domain.OrganizationResponse) MCPOrganization {
	return MCPOrganization{
		ID:           o.ID.String(),
		ParokiID:     o.ParokiID.String(),
		ParentID:     uuidPtrStr(o.ParentID),
		WilayahID:    uuidPtrStr(o.WilayahID),
		LingkunganID: uuidPtrStr(o.LingkunganID),
		Kind:         o.Kind,
		Code:         o.Code,
		Name:         o.Name,
		Description:  o.Description,
		Active:       o.Active,
		MyRole:       o.MyRole,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

func toMCPOrgMember(m domain.OrgMemberResponse) MCPOrgMember {
	return MCPOrgMember{
		ID:        m.ID.String(),
		ParokiID:  m.ParokiID.String(),
		OrgID:     m.OrgID.String(),
		UserID:    m.UserID.String(),
		Email:     m.Email,
		Name:      m.Name,
		Phone:     m.Phone,
		Role:      m.Role,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func toMCPOrgInvite(inv domain.OrgInviteResponse) MCPOrgInvite {
	return MCPOrgInvite{
		ID:        inv.ID.String(),
		ParokiID:  inv.ParokiID.String(),
		OrgID:     inv.OrgID.String(),
		Code:      inv.Code,
		Role:      inv.Role,
		MaxUses:   inv.MaxUses,
		UsedCount: inv.UsedCount,
		ExpiresAt: inv.ExpiresAt,
		RevokedAt: inv.RevokedAt,
		CreatedBy: inv.CreatedBy.String(),
		CreatedAt: inv.CreatedAt,
		UpdatedAt: inv.UpdatedAt,
	}
}

func (f *Feature) mcpOrgListPeriode(ctx context.Context, _ *mcp.CallToolRequest, in OrgListPeriodeInput) (*mcp.CallToolResult, OrgListPeriodeOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, OrgListPeriodeOutput{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, OrgListPeriodeOutput{}, httpx.BadRequest("Invalid organization ID.")
	}
	var filter *domain.OrgPeriodeSearchFilter
	if in.Status != "" || in.IsCurrent != nil {
		filter = &domain.OrgPeriodeSearchFilter{
			Status:    in.Status,
			IsCurrent: in.IsCurrent,
		}
	}
	items, _, err := f.service.SearchOrgPeriode(
		ctx,
		p,
		orgID,
		domain.SearchOrgPeriodeRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit),
	)
	if err != nil {
		return nil, OrgListPeriodeOutput{}, err
	}
	out := make([]MCPOrgPeriode, 0, len(items))
	for _, item := range items {
		out = append(out, MCPOrgPeriode{
			ID:             item.ID.String(),
			ParokiID:       item.ParokiID.String(),
			OrgID:          item.OrgID.String(),
			Kode:           item.Kode,
			Nama:           item.Nama,
			TanggalMulai:   item.TanggalMulai,
			TanggalSelesai: item.TanggalSelesai,
			Status:         item.Status,
			IsCurrent:      item.IsCurrent,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return nil, OrgListPeriodeOutput{Periode: out}, nil
}
