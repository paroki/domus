package cantus

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user"
)

// MCPSearchInput holds parameters for the cantus_search MCP tool.
type MCPSearchInput struct {
	Query     string `json:"query,omitempty" jsonschema:"Search text across judul, nomor, lirik, and kredit"`
	Slug      string `json:"slug,omitempty" jsonschema:"Exact cantus slug"`
	Jenis     string `json:"jenis,omitempty" jsonschema:"Cantus type: lagu, mazmur, or profan"`
	Sumber    string `json:"sumber,omitempty" jsonschema:"Source book (e.g. PS, MB, KA, Mazmur)"`
	Kategori  string `json:"kategori,omitempty" jsonschema:"Liturgical category (e.g. Pembuka, Persembahan, Komuni)"`
	Tag       string `json:"tag,omitempty" jsonschema:"Tag filter"`
	NadaDasar string `json:"nadaDasar,omitempty" jsonschema:"Base musical key (e.g. D, C, 1=D)"`
	Active    *bool  `json:"active,omitempty" jsonschema:"Filter by active status (defaults to true)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of results (1-100, default 20)"`
}

// MCPCantusItem is the MCP-safe representation of a Cantus entry (string UUIDs for JSON Schema).
type MCPCantusItem struct {
	ID        string               `json:"id"`
	Slug      string               `json:"slug"`
	Jenis     string               `json:"jenis"`
	Sumber    *string              `json:"sumber,omitempty"`
	Nomor     *string              `json:"nomor,omitempty"`
	Judul     string               `json:"judul"`
	Kategori  *string              `json:"kategori,omitempty"`
	Tag       []string             `json:"tag"`
	NadaDasar *string              `json:"nadaDasar,omitempty"`
	Lirik     *string              `json:"lirik,omitempty"`
	Kredit    map[string]string    `json:"kredit,omitempty"`
	Media     []domain.CantusMedia `json:"media,omitempty"`
	Meta      map[string]any       `json:"meta,omitempty"`
	Active    bool                 `json:"active"`
	CanManage bool                 `json:"canManage"`
	CreatedBy *string              `json:"createdBy,omitempty"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

func toMCPCantusItem(c domain.CantusResponse) MCPCantusItem {
	var createdBy *string
	if c.CreatedBy != nil {
		s := c.CreatedBy.String()
		createdBy = &s
	}
	return MCPCantusItem{
		ID:        c.ID.String(),
		Slug:      c.Slug,
		Jenis:     c.Jenis,
		Sumber:    c.Sumber,
		Nomor:     c.Nomor,
		Judul:     c.Judul,
		Kategori:  c.Kategori,
		Tag:       c.Tag,
		NadaDasar: c.NadaDasar,
		Lirik:     c.Lirik,
		Kredit:    c.Kredit,
		Media:     c.Media,
		Meta:      c.Meta,
		Active:    c.Active,
		CanManage: c.CanManage,
		CreatedBy: createdBy,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// MCPSearchOutput holds the result list for cantus_search.
type MCPSearchOutput struct {
	Cantus []MCPCantusItem `json:"cantus"`
}

// MCPGetInput holds parameters for the cantus_get MCP tool.
type MCPGetInput struct {
	CantusID string `json:"cantusId,omitempty" jsonschema:"Cantus UUID"`
	Slug     string `json:"slug,omitempty" jsonschema:"Cantus slug"`
}

// MCPCreateInput holds parameters for the cantus_create MCP tool.
type MCPCreateInput struct {
	Slug      string               `json:"slug" jsonschema:"Unique lowercase slug"`
	Jenis     string               `json:"jenis" jsonschema:"Cantus type: lagu, mazmur, or profan"`
	Sumber    *string              `json:"sumber,omitempty" jsonschema:"Source book"`
	Nomor     *string              `json:"nomor,omitempty" jsonschema:"Song or psalm number"`
	Judul     string               `json:"judul" jsonschema:"Song or psalm title"`
	Kategori  *string              `json:"kategori,omitempty" jsonschema:"Liturgical category"`
	Tag       []string             `json:"tag,omitempty" jsonschema:"Tags"`
	NadaDasar *string              `json:"nadaDasar,omitempty" jsonschema:"Base musical key"`
	Lirik     *string              `json:"lirik,omitempty" jsonschema:"Full lyrics"`
	Kredit    map[string]string    `json:"kredit,omitempty" jsonschema:"Attribution map (lagu, syair, aransemen)"`
	Media     []domain.CantusMedia `json:"media,omitempty" jsonschema:"Media attachments"`
	Meta      map[string]any       `json:"meta,omitempty" jsonschema:"Additional metadata"`
	Active    *bool                `json:"active,omitempty" jsonschema:"Active status (defaults to true)"`
}

// MCPUpdateInput holds parameters for the cantus_update MCP tool.
type MCPUpdateInput struct {
	CantusID  string               `json:"cantusId" jsonschema:"Cantus UUID to update"`
	Slug      *string              `json:"slug,omitempty" jsonschema:"Unique lowercase slug"`
	Jenis     *string              `json:"jenis,omitempty" jsonschema:"Cantus type: lagu, mazmur, or profan"`
	Sumber    *string              `json:"sumber,omitempty" jsonschema:"Source book"`
	Nomor     *string              `json:"nomor,omitempty" jsonschema:"Song or psalm number"`
	Judul     *string              `json:"judul,omitempty" jsonschema:"Song or psalm title"`
	Kategori  *string              `json:"kategori,omitempty" jsonschema:"Liturgical category"`
	Tag       []string             `json:"tag,omitempty" jsonschema:"Tags"`
	NadaDasar *string              `json:"nadaDasar,omitempty" jsonschema:"Base musical key"`
	Lirik     *string              `json:"lirik,omitempty" jsonschema:"Full lyrics"`
	Kredit    map[string]string    `json:"kredit,omitempty" jsonschema:"Attribution map"`
	Media     []domain.CantusMedia `json:"media,omitempty" jsonschema:"Media attachments"`
	Meta      map[string]any       `json:"meta,omitempty" jsonschema:"Additional metadata"`
	Active    *bool                `json:"active,omitempty" jsonschema:"Active status"`
}

// RegisterMCP registers global cantus_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "cantus_search",
		Description: "Search the global cross-parish song, psalm, and BPI catalog.",
	}, f.mcpCantusSearch)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "cantus_get",
		Description: "Get a single cantus entry by cantusId (UUID) or slug.",
	}, f.mcpCantusGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "cantus_create",
		Description: "Create a new global cantus entry (requires cantus:admin).",
	}, f.mcpCantusCreate)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "cantus_update",
		Description: "Update an existing global cantus entry by cantusId (requires cantus:admin).",
	}, f.mcpCantusUpdate)
}

func (f *Feature) authorizeMCP(ctx context.Context, act authz.Action) (user.Principal, error) {
	p, ok := user.UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return user.Principal{}, httpx.Unauthorized("Authentication is required.")
	}
	if f.enforcer == nil {
		if act == authz.ActionWrite {
			return user.Principal{}, domain.ErrCantusForbidden
		}
		return user.Principal{}, httpx.Forbidden("Forbidden.")
	}
	allowed, err := f.enforcer.Enforce(authz.UserSubject(p.UserID), authz.DomainGlobal, PermCantus, act)
	if err != nil {
		return user.Principal{}, err
	}
	if !allowed {
		if act == authz.ActionWrite {
			return user.Principal{}, domain.ErrCantusForbidden
		}
		return user.Principal{}, httpx.Forbidden("Forbidden.")
	}
	return p, nil
}

func (f *Feature) mcpCantusSearch(ctx context.Context, _ *mcp.CallToolRequest, in MCPSearchInput) (*mcp.CallToolResult, MCPSearchOutput, error) {
	p, err := f.authorizeMCP(ctx, authz.ActionRead)
	if err != nil {
		return nil, MCPSearchOutput{}, err
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	limit := in.Limit
	if limit <= 0 {
		limit = httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		limit = httpx.MaxLimit
	}
	items, _, err := f.service.Search(ctx, p.UserID, domain.SearchCantusRequest{
		Q: in.Query,
		Filter: &domain.CantusSearchFilter{
			Slug:      in.Slug,
			Jenis:     in.Jenis,
			Sumber:    in.Sumber,
			Kategori:  in.Kategori,
			Tag:       in.Tag,
			NadaDasar: in.NadaDasar,
			Active:    &active,
		},
		Limit: limit,
	}, nil, nil, limit)
	if err != nil {
		return nil, MCPSearchOutput{}, err
	}
	out := make([]MCPCantusItem, 0, len(items))
	for _, item := range items {
		out = append(out, toMCPCantusItem(item))
	}
	return nil, MCPSearchOutput{Cantus: out}, nil
}

func (f *Feature) mcpCantusGet(ctx context.Context, _ *mcp.CallToolRequest, in MCPGetInput) (*mcp.CallToolResult, MCPCantusItem, error) {
	p, err := f.authorizeMCP(ctx, authz.ActionRead)
	if err != nil {
		return nil, MCPCantusItem{}, err
	}
	if rawID := strings.TrimSpace(in.CantusID); rawID != "" {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, MCPCantusItem{}, httpx.BadRequest("Invalid cantusId.")
		}
		res, err := f.service.Get(ctx, p.UserID, id)
		if err != nil {
			return nil, MCPCantusItem{}, err
		}
		return nil, toMCPCantusItem(res), nil
	}
	if slug := strings.TrimSpace(in.Slug); slug != "" {
		res, err := f.service.GetBySlug(ctx, p.UserID, slug)
		if err != nil {
			return nil, MCPCantusItem{}, err
		}
		return nil, toMCPCantusItem(res), nil
	}
	return nil, MCPCantusItem{}, httpx.BadRequest("Either cantusId or slug is required.")
}

func (f *Feature) mcpCantusCreate(ctx context.Context, _ *mcp.CallToolRequest, in MCPCreateInput) (*mcp.CallToolResult, MCPCantusItem, error) {
	p, err := f.authorizeMCP(ctx, authz.ActionWrite)
	if err != nil {
		return nil, MCPCantusItem{}, err
	}
	res, err := f.service.Create(ctx, p.UserID, domain.CreateCantusRequest{
		Slug:      in.Slug,
		Jenis:     in.Jenis,
		Sumber:    in.Sumber,
		Nomor:     in.Nomor,
		Judul:     in.Judul,
		Kategori:  in.Kategori,
		Tag:       in.Tag,
		NadaDasar: in.NadaDasar,
		Lirik:     in.Lirik,
		Kredit:    in.Kredit,
		Media:     in.Media,
		Meta:      in.Meta,
		Active:    in.Active,
	})
	if err != nil {
		return nil, MCPCantusItem{}, err
	}
	return nil, toMCPCantusItem(res), nil
}

func (f *Feature) mcpCantusUpdate(ctx context.Context, _ *mcp.CallToolRequest, in MCPUpdateInput) (*mcp.CallToolResult, MCPCantusItem, error) {
	p, err := f.authorizeMCP(ctx, authz.ActionWrite)
	if err != nil {
		return nil, MCPCantusItem{}, err
	}
	id, err := uuid.Parse(strings.TrimSpace(in.CantusID))
	if err != nil {
		return nil, MCPCantusItem{}, httpx.BadRequest("Invalid cantusId.")
	}
	res, err := f.service.Update(ctx, p.UserID, id, domain.UpdateCantusRequest{
		Slug:      in.Slug,
		Jenis:     in.Jenis,
		Sumber:    in.Sumber,
		Nomor:     in.Nomor,
		Judul:     in.Judul,
		Kategori:  in.Kategori,
		Tag:       in.Tag,
		NadaDasar: in.NadaDasar,
		Lirik:     in.Lirik,
		Kredit:    in.Kredit,
		Media:     in.Media,
		Meta:      in.Meta,
		Active:    in.Active,
	})
	if err != nil {
		return nil, MCPCantusItem{}, err
	}
	return nil, toMCPCantusItem(res), nil
}
