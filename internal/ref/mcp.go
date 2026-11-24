package ref

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/user"
)

// RefGetParishInput is the input schema for ref_get_parish.
type RefGetParishInput struct{}

// MCPRefParish is the MCP-safe representation of a Parish.
type MCPRefParish struct {
	ID          string    `json:"id"`
	DioceseID   string    `json:"dioceseId"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	PatronSaint *string   `json:"patronSaint,omitempty"`
	PastorName  *string   `json:"pastorName,omitempty"`
	Address     *string   `json:"address,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Email       *string   `json:"email,omitempty"`
	Website     *string   `json:"website,omitempty"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	Letterhead  *string   `json:"letterhead,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RefListZonesInput is the input schema for ref_list_zones.
type RefListZonesInput struct {
	Query  string `json:"query,omitempty"`
	Active *bool  `json:"active,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// MCPRefZone is the MCP-safe representation of a Zone.
type MCPRefZone struct {
	ID               string    `json:"id"`
	ParishID         string    `json:"parishId"`
	Type             string    `json:"type"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	CoordinatorName  *string   `json:"coordinatorName,omitempty"`
	CoordinatorPhone *string   `json:"coordinatorPhone,omitempty"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// RefListZonesOutput is the output schema for ref_list_zones.
type RefListZonesOutput struct {
	Zones []MCPRefZone `json:"zones"`
}

// RefListWardsInput is the input schema for ref_list_wards.
type RefListWardsInput struct {
	ZoneID string `json:"zoneId,omitempty"`
	Query  string `json:"query,omitempty"`
	Active *bool  `json:"active,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// MCPRefWard is the MCP-safe representation of a Ward.
type MCPRefWard struct {
	ID          string    `json:"id"`
	ParishID    string    `json:"parishId"`
	ZoneID      string    `json:"zoneId"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	PatronSaint *string   `json:"patronSaint,omitempty"`
	HeadName    *string   `json:"headName,omitempty"`
	HeadPhone   *string   `json:"headPhone,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RefListWardsOutput is the output schema for ref_list_wards.
type RefListWardsOutput struct {
	Wards []MCPRefWard `json:"wards"`
}

// RefListPeriodsInput is the input schema for ref_list_periods.
type RefListPeriodsInput struct {
	Query     string `json:"query,omitempty"`
	Status    string `json:"status,omitempty"`
	IsCurrent *bool  `json:"isCurrent,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// MCPRefPeriod is the MCP-safe representation of a Period.
type MCPRefPeriod struct {
	ID        string    `json:"id"`
	ParishID  string    `json:"parishId"`
	Type      string    `json:"type"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	StartDate string    `json:"startDate"`
	EndDate   string    `json:"endDate"`
	Status    string    `json:"status"`
	IsCurrent bool      `json:"isCurrent"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RefListPeriodsOutput is the output schema for ref_list_periods.
type RefListPeriodsOutput struct {
	Periods []MCPRefPeriod `json:"periods"`
}

// RefListLookupsInput is the input schema for ref_list_lookups.
type RefListLookupsInput struct {
	Category string `json:"category,omitempty"`
	Query    string `json:"query,omitempty"`
	Active   *bool  `json:"active,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// MCPRefLookup is the MCP-safe representation of a Lookup.
type MCPRefLookup struct {
	ID        string    `json:"id"`
	ParishID  *string   `json:"parishId,omitempty"`
	Category  string    `json:"category"`
	Code      string    `json:"code"`
	Label     string    `json:"label"`
	SortOrder int       `json:"sortOrder"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// RefListLookupsOutput is the output schema for ref_list_lookups.
type RefListLookupsOutput struct {
	Lookups []MCPRefLookup `json:"lookups"`
}

// RegisterMCP registers read-only ref_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_get_parish",
		Description: "Get the caller's active parish profile.",
	}, f.mcpRefGetParish)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_zones",
		Description: "List territorial zones/stations in the caller's active parish.",
	}, f.mcpRefListZones)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_wards",
		Description: "List wards (lingkungan) in the caller's active parish.",
	}, f.mcpRefListWards)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_periods",
		Description: "List fiscal and liturgical periods in the caller's active parish.",
	}, f.mcpRefListPeriods)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_lookups",
		Description: "List merged global and active-parish reference lookups.",
	}, f.mcpRefListLookups)
}

func (f *Feature) mcpRefGetParish(ctx context.Context, _ *mcp.CallToolRequest, _ RefGetParishInput) (*mcp.CallToolResult, MCPRefParish, error) {
	p, err := f.requireMCPAccess(ctx, PermParishes, authz.ActionRead)
	if err != nil {
		return nil, MCPRefParish{}, err
	}
	res, err := f.service.GetParish(ctx, p.ParishID)
	if err != nil {
		return nil, MCPRefParish{}, err
	}
	return nil, MCPRefParish{
		ID:          res.ID.String(),
		DioceseID:   res.DioceseID.String(),
		Code:        res.Code,
		Name:        res.Name,
		PatronSaint: res.PatronSaint,
		PastorName:  res.PastorName,
		Address:     res.Address,
		Phone:       res.Phone,
		Email:       res.Email,
		Website:     res.Website,
		LogoURL:     res.LogoURL,
		Letterhead:  res.Letterhead,
		Active:      res.Active,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
	}, nil
}

func (f *Feature) mcpRefListZones(ctx context.Context, _ *mcp.CallToolRequest, in RefListZonesInput) (*mcp.CallToolResult, RefListZonesOutput, error) {
	p, err := f.requireMCPAccess(ctx, PermMasterData, authz.ActionRead)
	if err != nil {
		return nil, RefListZonesOutput{}, err
	}
	var filter *domain.ZoneSearchFilter
	if in.Active != nil {
		filter = &domain.ZoneSearchFilter{Active: in.Active}
	}
	items, _, err := f.service.SearchZones(
		ctx,
		p.ParishID,
		domain.SearchZonesRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit, 50, 100),
	)
	if err != nil {
		return nil, RefListZonesOutput{}, err
	}
	out := make([]MCPRefZone, 0, len(items))
	for _, z := range items {
		out = append(out, MCPRefZone{
			ID:               z.ID.String(),
			ParishID:         z.ParishID.String(),
			Type:             z.Type,
			Code:             z.Code,
			Name:             z.Name,
			CoordinatorName:  z.CoordinatorName,
			CoordinatorPhone: z.CoordinatorPhone,
			Active:           z.Active,
			CreatedAt:        z.CreatedAt,
			UpdatedAt:        z.UpdatedAt,
		})
	}
	return nil, RefListZonesOutput{Zones: out}, nil
}

func (f *Feature) mcpRefListWards(ctx context.Context, _ *mcp.CallToolRequest, in RefListWardsInput) (*mcp.CallToolResult, RefListWardsOutput, error) {
	p, err := f.requireMCPAccess(ctx, PermMasterData, authz.ActionRead)
	if err != nil {
		return nil, RefListWardsOutput{}, err
	}
	var filter *domain.WardSearchFilter
	if in.ZoneID != "" || in.Active != nil {
		filter = &domain.WardSearchFilter{Active: in.Active}
		if in.ZoneID != "" {
			zid, err := uuid.Parse(in.ZoneID)
			if err != nil {
				return nil, RefListWardsOutput{}, httpx.BadRequest("Invalid zoneId.")
			}
			filter.ZoneID = &zid
		}
	}
	items, _, err := f.service.SearchWards(
		ctx,
		p.ParishID,
		domain.SearchWardsRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit, 50, 100),
	)
	if err != nil {
		return nil, RefListWardsOutput{}, err
	}
	out := make([]MCPRefWard, 0, len(items))
	for _, w := range items {
		out = append(out, MCPRefWard{
			ID:          w.ID.String(),
			ParishID:    w.ParishID.String(),
			ZoneID:      w.ZoneID.String(),
			Code:        w.Code,
			Name:        w.Name,
			PatronSaint: w.PatronSaint,
			HeadName:    w.HeadName,
			HeadPhone:   w.HeadPhone,
			Active:      w.Active,
			CreatedAt:   w.CreatedAt,
			UpdatedAt:   w.UpdatedAt,
		})
	}
	return nil, RefListWardsOutput{Wards: out}, nil
}

func (f *Feature) mcpRefListPeriods(ctx context.Context, _ *mcp.CallToolRequest, in RefListPeriodsInput) (*mcp.CallToolResult, RefListPeriodsOutput, error) {
	p, err := f.requireMCPAccess(ctx, PermMasterData, authz.ActionRead)
	if err != nil {
		return nil, RefListPeriodsOutput{}, err
	}
	var filter *domain.PeriodSearchFilter
	if in.Status != "" || in.IsCurrent != nil {
		filter = &domain.PeriodSearchFilter{Status: in.Status, IsCurrent: in.IsCurrent}
	}
	items, _, err := f.service.SearchPeriods(
		ctx,
		p.ParishID,
		domain.SearchPeriodsRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit, 50, 100),
	)
	if err != nil {
		return nil, RefListPeriodsOutput{}, err
	}
	out := make([]MCPRefPeriod, 0, len(items))
	for _, pr := range items {
		out = append(out, MCPRefPeriod{
			ID:        pr.ID.String(),
			ParishID:  pr.ParishID.String(),
			Type:      pr.Type,
			Code:      pr.Code,
			Name:      pr.Name,
			StartDate: pr.StartDate,
			EndDate:   pr.EndDate,
			Status:    pr.Status,
			IsCurrent: pr.IsCurrent,
			CreatedAt: pr.CreatedAt,
			UpdatedAt: pr.UpdatedAt,
		})
	}
	return nil, RefListPeriodsOutput{Periods: out}, nil
}

func (f *Feature) mcpRefListLookups(ctx context.Context, _ *mcp.CallToolRequest, in RefListLookupsInput) (*mcp.CallToolResult, RefListLookupsOutput, error) {
	p, err := f.requireMCPAccess(ctx, PermMasterData, authz.ActionRead)
	if err != nil {
		return nil, RefListLookupsOutput{}, err
	}
	var filter *domain.LookupSearchFilter
	if in.Category != "" || in.Active != nil {
		filter = &domain.LookupSearchFilter{Category: in.Category, Active: in.Active}
	}
	items, _, err := f.service.SearchLookups(
		ctx,
		p,
		domain.SearchLookupsRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit, 50, 100),
	)
	if err != nil {
		return nil, RefListLookupsOutput{}, err
	}
	out := make([]MCPRefLookup, 0, len(items))
	for _, l := range items {
		var parishID *string
		if l.ParishID != nil {
			s := l.ParishID.String()
			parishID = &s
		}
		out = append(out, MCPRefLookup{
			ID:        l.ID.String(),
			ParishID:  parishID,
			Category:  l.Category,
			Code:      l.Code,
			Label:     l.Label,
			SortOrder: l.SortOrder,
			Active:    l.Active,
			CreatedAt: l.CreatedAt,
			UpdatedAt: l.UpdatedAt,
		})
	}
	return nil, RefListLookupsOutput{Lookups: out}, nil
}

func (f *Feature) requireMCPAccess(ctx context.Context, resource string, action authz.Action) (user.Principal, error) {
	p, ok := user.UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return user.Principal{}, httpx.Unauthorized("Authentication is required.")
	}
	tp, err := user.RequireTenant(ctx)
	if err != nil {
		return user.Principal{}, err
	}
	if f.enforcer != nil {
		allowed, err := f.enforcer.Enforce(authz.UserSubject(tp.UserID), authz.DomainGlobal, resource, action)
		if err != nil {
			return user.Principal{}, err
		}
		if !allowed {
			return user.Principal{}, httpx.Forbidden("Forbidden.")
		}
	}
	return tp, nil
}

func clampMCPLimit(limit, def, max int) int {
	if limit <= 0 {
		return def
	}
	if limit > max {
		return max
	}
	return limit
}
