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

// RefGetParokiInput is the input schema for ref_get_paroki.
type RefGetParokiInput struct{}

// MCPRefParoki is the MCP-safe representation of a Paroki.
type MCPRefParoki struct {
	ID          string    `json:"id"`
	KeuskupanID string    `json:"keuskupanId"`
	Kode        string    `json:"kode"`
	Nama        string    `json:"nama"`
	Pelindung   *string   `json:"pelindung,omitempty"`
	Pastor      *string   `json:"pastor,omitempty"`
	Alamat      *string   `json:"alamat,omitempty"`
	Telepon     *string   `json:"telepon,omitempty"`
	Email       *string   `json:"email,omitempty"`
	Website     *string   `json:"website,omitempty"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	KopSurat    *string   `json:"kopSurat,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// RefListWilayahInput is the input schema for ref_list_wilayah.
type RefListWilayahInput struct {
	Query  string `json:"query,omitempty"`
	Active *bool  `json:"active,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// MCPRefWilayah is the MCP-safe representation of a Wilayah.
type MCPRefWilayah struct {
	ID                 string    `json:"id"`
	ParokiID           string    `json:"parokiId"`
	Jenis              string    `json:"jenis"`
	Kode               string    `json:"kode"`
	Nama               string    `json:"nama"`
	Koordinator        *string   `json:"koordinator,omitempty"`
	TeleponKoordinator *string   `json:"teleponKoordinator,omitempty"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// RefListWilayahOutput is the output schema for ref_list_wilayah.
type RefListWilayahOutput struct {
	Wilayah []MCPRefWilayah `json:"wilayah"`
}

// RefListLingkunganInput is the input schema for ref_list_lingkungan.
type RefListLingkunganInput struct {
	WilayahID string `json:"wilayahId,omitempty"`
	Query     string `json:"query,omitempty"`
	Active    *bool  `json:"active,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

// MCPRefLingkungan is the MCP-safe representation of a Lingkungan.
type MCPRefLingkungan struct {
	ID           string    `json:"id"`
	ParokiID     string    `json:"parokiId"`
	WilayahID    string    `json:"wilayahId"`
	Kode         string    `json:"kode"`
	Nama         string    `json:"nama"`
	Pelindung    *string   `json:"pelindung,omitempty"`
	Ketua        *string   `json:"ketua,omitempty"`
	TeleponKetua *string   `json:"teleponKetua,omitempty"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// RefListLingkunganOutput is the output schema for ref_list_lingkungan.
type RefListLingkunganOutput struct {
	Lingkungan []MCPRefLingkungan `json:"lingkungan"`
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
	ParokiID  string    `json:"parokiId"`
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
	ParokiID  *string   `json:"parokiId,omitempty"`
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
		Name:        "ref_get_paroki",
		Description: "Get the caller's active paroki profile.",
	}, f.mcpRefGetParoki)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_wilayah",
		Description: "List territorial wilayah/stasi in the caller's active paroki.",
	}, f.mcpRefListWilayah)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_lingkungan",
		Description: "List lingkungan in the caller's active paroki.",
	}, f.mcpRefListLingkungan)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_periods",
		Description: "List fiscal and liturgical periods in the caller's active paroki.",
	}, f.mcpRefListPeriods)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "ref_list_lookups",
		Description: "List merged global and active-paroki reference lookups.",
	}, f.mcpRefListLookups)
}

func (f *Feature) mcpRefGetParoki(ctx context.Context, _ *mcp.CallToolRequest, _ RefGetParokiInput) (*mcp.CallToolResult, MCPRefParoki, error) {
	p, err := f.requireMCPAccess(ctx, PermParoki, authz.ActionRead)
	if err != nil {
		return nil, MCPRefParoki{}, err
	}
	res, err := f.service.GetParoki(ctx, p.ParokiID)
	if err != nil {
		return nil, MCPRefParoki{}, err
	}
	return nil, MCPRefParoki{
		ID:          res.ID.String(),
		KeuskupanID: res.KeuskupanID.String(),
		Kode:        res.Kode,
		Nama:        res.Nama,
		Pelindung:   res.Pelindung,
		Pastor:      res.Pastor,
		Alamat:      res.Alamat,
		Telepon:     res.Telepon,
		Email:       res.Email,
		Website:     res.Website,
		LogoURL:     res.LogoURL,
		KopSurat:    res.KopSurat,
		Active:      res.Active,
		CreatedAt:   res.CreatedAt,
		UpdatedAt:   res.UpdatedAt,
	}, nil
}

func (f *Feature) mcpRefListWilayah(ctx context.Context, _ *mcp.CallToolRequest, in RefListWilayahInput) (*mcp.CallToolResult, RefListWilayahOutput, error) {
	p, err := f.requireMCPAccess(ctx, PermMasterData, authz.ActionRead)
	if err != nil {
		return nil, RefListWilayahOutput{}, err
	}
	var filter *domain.WilayahSearchFilter
	if in.Active != nil {
		filter = &domain.WilayahSearchFilter{Active: in.Active}
	}
	items, _, err := f.service.SearchWilayah(
		ctx,
		p.ParokiID,
		domain.SearchWilayahRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit, 50, 100),
	)
	if err != nil {
		return nil, RefListWilayahOutput{}, err
	}
	out := make([]MCPRefWilayah, 0, len(items))
	for _, w := range items {
		out = append(out, MCPRefWilayah{
			ID:                 w.ID.String(),
			ParokiID:           w.ParokiID.String(),
			Jenis:              w.Jenis,
			Kode:               w.Kode,
			Nama:               w.Nama,
			Koordinator:        w.Koordinator,
			TeleponKoordinator: w.TeleponKoordinator,
			Active:             w.Active,
			CreatedAt:          w.CreatedAt,
			UpdatedAt:          w.UpdatedAt,
		})
	}
	return nil, RefListWilayahOutput{Wilayah: out}, nil
}

func (f *Feature) mcpRefListLingkungan(ctx context.Context, _ *mcp.CallToolRequest, in RefListLingkunganInput) (*mcp.CallToolResult, RefListLingkunganOutput, error) {
	p, err := f.requireMCPAccess(ctx, PermMasterData, authz.ActionRead)
	if err != nil {
		return nil, RefListLingkunganOutput{}, err
	}
	var filter *domain.LingkunganSearchFilter
	if in.WilayahID != "" || in.Active != nil {
		filter = &domain.LingkunganSearchFilter{Active: in.Active}
		if in.WilayahID != "" {
			wid, err := uuid.Parse(in.WilayahID)
			if err != nil {
				return nil, RefListLingkunganOutput{}, httpx.BadRequest("Invalid wilayahId.")
			}
			filter.WilayahID = &wid
		}
	}
	items, _, err := f.service.SearchLingkungan(
		ctx,
		p.ParokiID,
		domain.SearchLingkunganRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit, 50, 100),
	)
	if err != nil {
		return nil, RefListLingkunganOutput{}, err
	}
	out := make([]MCPRefLingkungan, 0, len(items))
	for _, l := range items {
		out = append(out, MCPRefLingkungan{
			ID:           l.ID.String(),
			ParokiID:     l.ParokiID.String(),
			WilayahID:    l.WilayahID.String(),
			Kode:         l.Kode,
			Nama:         l.Nama,
			Pelindung:    l.Pelindung,
			Ketua:        l.Ketua,
			TeleponKetua: l.TeleponKetua,
			Active:       l.Active,
			CreatedAt:    l.CreatedAt,
			UpdatedAt:    l.UpdatedAt,
		})
	}
	return nil, RefListLingkunganOutput{Lingkungan: out}, nil
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
		p.ParokiID,
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
			ParokiID:  pr.ParokiID.String(),
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
		var parokiID *string
		if l.ParokiID != nil {
			s := l.ParokiID.String()
			parokiID = &s
		}
		out = append(out, MCPRefLookup{
			ID:        l.ID.String(),
			ParokiID:  parokiID,
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
