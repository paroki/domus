package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/event/domain"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user"
)

// MCPEventRSVP is the MCP-safe representation of an EventRSVP (UUIDs and timestamps as strings).
type MCPEventRSVP struct {
	ID          string  `json:"id"`
	ParishID    string  `json:"parishId"`
	EventID     string  `json:"eventId"`
	UserID      string  `json:"userId"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Phone       *string `json:"phone,omitempty"`
	OrgRole     string  `json:"orgRole,omitempty"`
	OrgPosition *string `json:"orgPosition,omitempty"`
	Status      string  `json:"status"`
	Note        *string `json:"note,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// MCPEventRSVPSummary is the MCP-safe representation of EventRSVPSummaryResponse (UUIDs as strings).
type MCPEventRSVPSummary struct {
	EventID            string `json:"eventId"`
	Going              int    `json:"going"`
	Maybe              int    `json:"maybe"`
	NotGoing           int    `json:"notGoing"`
	TotalResponded     int    `json:"totalResponded"`
	TotalActiveMembers int    `json:"totalActiveMembers"`
	Unresponded        int    `json:"unresponded"`
}

// MCPEvent is the MCP-safe representation of an Event (UUIDs and timestamps as strings).
type MCPEvent struct {
	ID           string               `json:"id"`
	ParishID     string               `json:"parishId"`
	OrgID        string               `json:"orgId"`
	OrgCode      string               `json:"orgCode"`
	OrgName      string               `json:"orgName"`
	OrgKind      string               `json:"orgKind"`
	CreatedBy    string               `json:"createdBy"`
	Title        string               `json:"title"`
	Description  *string              `json:"description,omitempty"`
	Location     *string              `json:"location,omitempty"`
	StartAt      string               `json:"startAt"`
	EndAt        string               `json:"endAt"`
	Visibility   string               `json:"visibility"`
	RsvpEnabled  bool                 `json:"rsvpEnabled"`
	RsvpDeadline *string              `json:"rsvpDeadline,omitempty"`
	Status       string               `json:"status"`
	RsvpOpen     bool                 `json:"rsvpOpen"`
	MyOrgRole    *string              `json:"myOrgRole,omitempty"`
	MyRsvp       *MCPEventRSVP        `json:"myRsvp,omitempty"`
	RsvpSummary  *MCPEventRSVPSummary `json:"rsvpSummary,omitempty"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

// EventListInput is the input schema for event_list.
type EventListInput struct {
	Query      string `json:"query,omitempty" jsonschema:"Optional text search query matching event title, description, or location"`
	OrgID      string `json:"orgId,omitempty" jsonschema:"Optional organization UUID filter"`
	From       string `json:"from,omitempty" jsonschema:"Optional inclusive start timestamp in RFC3339 (e.g. 2026-11-23T00:00:00Z) or YYYY-MM-DD"`
	To         string `json:"to,omitempty" jsonschema:"Optional inclusive end timestamp in RFC3339 (e.g. 2026-11-29T23:59:59Z) or YYYY-MM-DD"`
	Visibility string `json:"visibility,omitempty" jsonschema:"Optional visibility filter: public or private"`
	Status     string `json:"status,omitempty" jsonschema:"Optional status filter: scheduled or cancelled"`
	MyOrgsOnly bool   `json:"myOrgsOnly,omitempty" jsonschema:"If true, restrict results to events owned by caller's organizations"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of events to return (default 20, max 100)"`
}

// EventListOutput is the output schema for event_list.
type EventListOutput struct {
	Events []MCPEvent `json:"events"`
}

// EventGetInput is the input schema for event_get.
type EventGetInput struct {
	EventID string `json:"eventId"`
}

// EventCreateInput is the input schema for event_create.
type EventCreateInput struct {
	OrgID        string  `json:"orgId"`
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Location     *string `json:"location,omitempty"`
	StartAt      string  `json:"startAt"`
	EndAt        string  `json:"endAt"`
	Visibility   string  `json:"visibility,omitempty"`
	RsvpEnabled  *bool   `json:"rsvpEnabled,omitempty"`
	RsvpDeadline *string `json:"rsvpDeadline,omitempty"`
}

// EventUpdateInput is the input schema for event_update.
type EventUpdateInput struct {
	EventID           string  `json:"eventId"`
	Title             *string `json:"title,omitempty"`
	Description       *string `json:"description,omitempty"`
	Location          *string `json:"location,omitempty"`
	StartAt           *string `json:"startAt,omitempty"`
	EndAt             *string `json:"endAt,omitempty"`
	Visibility        *string `json:"visibility,omitempty"`
	RsvpEnabled       *bool   `json:"rsvpEnabled,omitempty"`
	RsvpDeadline      *string `json:"rsvpDeadline,omitempty"`
	ClearRsvpDeadline *bool   `json:"clearRsvpDeadline,omitempty"`
	Status            *string `json:"status,omitempty"`
}

// EventRSVPInput is the input schema for event_rsvp.
type EventRSVPInput struct {
	EventID string  `json:"eventId"`
	Status  string  `json:"status"`
	Note    *string `json:"note,omitempty"`
}

// EventListRSVPsInput is the input schema for event_list_rsvps.
type EventListRSVPsInput struct {
	EventID string `json:"eventId"`
	Query   string `json:"query,omitempty"`
	Status  string `json:"status,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// EventListRSVPsOutput is the output schema for event_list_rsvps.
type EventListRSVPsOutput struct {
	Summary MCPEventRSVPSummary `json:"summary"`
	RSVPs   []MCPEventRSVP      `json:"rsvps"`
}

// RegisterMCP registers all event_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_list",
		Description: "List visible events in the caller's active parish (public parish events and private events for caller's organizations).",
	}, f.mcpEventList)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_get",
		Description: "Get details of a single event in the caller's active parish, including RSVP summary and caller's RSVP if an org member.",
	}, f.mcpEventGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_create",
		Description: "Create a new organization-owned event in the caller's active parish (requires org:admin of the owning organization).",
	}, f.mcpEventCreate)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_update",
		Description: "Update or cancel an existing organization-owned event (requires org:admin of the owning organization).",
	}, f.mcpEventUpdate)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_rsvp",
		Description: "Submit or update the caller's RSVP for an organization event (requires active membership in the owning organization).",
	}, f.mcpEventRSVP)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_list_rsvps",
		Description: "Get RSVP attendance summary and attendee roster for an event (requires active membership in the owning organization).",
	}, f.mcpEventListRSVPs)
}

func (f *Feature) mcpEventList(ctx context.Context, _ *mcp.CallToolRequest, in EventListInput) (*mcp.CallToolResult, EventListOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, EventListOutput{}, err
	}

	var filter *domain.EventSearchFilter
	if in.OrgID != "" || in.From != "" || in.To != "" || in.Visibility != "" || in.Status != "" || in.MyOrgsOnly {
		filter = &domain.EventSearchFilter{
			Visibility: in.Visibility,
			Status:     in.Status,
		}
		if in.OrgID != "" {
			parsed, err := uuid.Parse(in.OrgID)
			if err != nil {
				return nil, EventListOutput{}, httpx.BadRequest("Invalid organization ID.")
			}
			filter.OrgID = &parsed
		}
		if in.From != "" {
			t, err := parseMCPTimestamp(in.From, false)
			if err != nil {
				return nil, EventListOutput{}, httpx.BadRequest("Invalid 'from' timestamp; expected RFC3339 or YYYY-MM-DD.")
			}
			filter.From = &t
		}
		if in.To != "" {
			t, err := parseMCPTimestamp(in.To, true)
			if err != nil {
				return nil, EventListOutput{}, httpx.BadRequest("Invalid 'to' timestamp; expected RFC3339 or YYYY-MM-DD.")
			}
			filter.To = &t
		}
		if in.MyOrgsOnly {
			t := true
			filter.MyOrgsOnly = &t
		}
	}

	items, _, err := f.service.SearchEvents(
		ctx,
		p,
		domain.SearchEventsRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit),
	)
	if err != nil {
		return nil, EventListOutput{}, err
	}
	out := make([]MCPEvent, 0, len(items))
	for _, item := range items {
		out = append(out, toMCPEvent(item))
	}
	return nil, EventListOutput{Events: out}, nil
}

func (f *Feature) mcpEventGet(ctx context.Context, _ *mcp.CallToolRequest, in EventGetInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid event ID.")
	}
	res, err := f.service.GetEvent(ctx, p, eventID)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(res), nil
}

func (f *Feature) mcpEventCreate(ctx context.Context, _ *mcp.CallToolRequest, in EventCreateInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	orgID, err := uuid.Parse(in.OrgID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid organization ID.")
	}
	startAt, err := time.Parse(time.RFC3339, in.StartAt)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid 'startAt' timestamp; expected RFC3339.")
	}
	endAt, err := time.Parse(time.RFC3339, in.EndAt)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid 'endAt' timestamp; expected RFC3339.")
	}
	var rsvpDeadline *time.Time
	if in.RsvpDeadline != nil && *in.RsvpDeadline != "" {
		t, err := time.Parse(time.RFC3339, *in.RsvpDeadline)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'rsvpDeadline' timestamp; expected RFC3339.")
		}
		rsvpDeadline = &t
	}

	res, err := f.service.CreateEvent(ctx, p, domain.CreateEventRequest{
		OrgID:        orgID,
		Title:        in.Title,
		Description:  in.Description,
		Location:     in.Location,
		StartAt:      startAt,
		EndAt:        endAt,
		Visibility:   in.Visibility,
		RsvpEnabled:  in.RsvpEnabled,
		RsvpDeadline: rsvpDeadline,
	})
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(res), nil
}

func (f *Feature) mcpEventUpdate(ctx context.Context, _ *mcp.CallToolRequest, in EventUpdateInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid event ID.")
	}

	req := domain.UpdateEventRequest{
		Title:             in.Title,
		Description:       in.Description,
		Location:          in.Location,
		Visibility:        in.Visibility,
		RsvpEnabled:       in.RsvpEnabled,
		ClearRsvpDeadline: in.ClearRsvpDeadline,
		Status:            in.Status,
	}
	if in.StartAt != nil {
		t, err := time.Parse(time.RFC3339, *in.StartAt)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'startAt' timestamp; expected RFC3339.")
		}
		req.StartAt = &t
	}
	if in.EndAt != nil {
		t, err := time.Parse(time.RFC3339, *in.EndAt)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'endAt' timestamp; expected RFC3339.")
		}
		req.EndAt = &t
	}
	if in.RsvpDeadline != nil && *in.RsvpDeadline != "" {
		t, err := time.Parse(time.RFC3339, *in.RsvpDeadline)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'rsvpDeadline' timestamp; expected RFC3339.")
		}
		req.RsvpDeadline = &t
	}

	res, err := f.service.UpdateEvent(ctx, p, eventID, req)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(res), nil
}

func (f *Feature) mcpEventRSVP(ctx context.Context, _ *mcp.CallToolRequest, in EventRSVPInput) (*mcp.CallToolResult, MCPEventRSVP, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEventRSVP{}, err
	}
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEventRSVP{}, httpx.BadRequest("Invalid event ID.")
	}
	res, err := f.service.UpsertRSVP(ctx, p, eventID, domain.UpsertEventRSVPRequest{
		Status: in.Status,
		Note:   in.Note,
	})
	if err != nil {
		return nil, MCPEventRSVP{}, err
	}
	return nil, toMCPEventRSVP(res), nil
}

func (f *Feature) mcpEventListRSVPs(ctx context.Context, _ *mcp.CallToolRequest, in EventListRSVPsInput) (*mcp.CallToolResult, EventListRSVPsOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, EventListRSVPsOutput{}, err
	}
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, EventListRSVPsOutput{}, httpx.BadRequest("Invalid event ID.")
	}

	summary, err := f.service.GetRSVPSummary(ctx, p, eventID)
	if err != nil {
		return nil, EventListRSVPsOutput{}, err
	}

	var filter *domain.EventRSVPSearchFilter
	if in.Status != "" {
		filter = &domain.EventRSVPSearchFilter{Status: in.Status}
	}
	items, _, err := f.service.SearchRSVPs(
		ctx,
		p,
		eventID,
		domain.SearchEventRSVPsRequest{Q: in.Query, Filter: filter},
		nil,
		nil,
		clampMCPLimit(in.Limit),
	)
	if err != nil {
		return nil, EventListRSVPsOutput{}, err
	}
	out := make([]MCPEventRSVP, 0, len(items))
	for _, item := range items {
		out = append(out, toMCPEventRSVP(item))
	}
	return nil, EventListRSVPsOutput{
		Summary: toMCPEventRSVPSummary(summary),
		RSVPs:   out,
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

func toMCPEvent(e domain.EventResponse) MCPEvent {
	var deadline *string
	if e.RsvpDeadline != nil {
		s := e.RsvpDeadline.UTC().Format(time.RFC3339)
		deadline = &s
	}
	var myRsvp *MCPEventRSVP
	if e.MyRsvp != nil {
		r := toMCPEventRSVP(*e.MyRsvp)
		myRsvp = &r
	}
	var summary *MCPEventRSVPSummary
	if e.RsvpSummary != nil {
		sum := toMCPEventRSVPSummary(*e.RsvpSummary)
		summary = &sum
	}
	return MCPEvent{
		ID:           e.ID.String(),
		ParishID:     e.ParishID.String(),
		OrgID:        e.OrgID.String(),
		OrgCode:      e.OrgCode,
		OrgName:      e.OrgName,
		OrgKind:      e.OrgKind,
		CreatedBy:    e.CreatedBy.String(),
		Title:        e.Title,
		Description:  e.Description,
		Location:     e.Location,
		StartAt:      e.StartAt.UTC().Format(time.RFC3339),
		EndAt:        e.EndAt.UTC().Format(time.RFC3339),
		Visibility:   e.Visibility,
		RsvpEnabled:  e.RsvpEnabled,
		RsvpDeadline: deadline,
		Status:       e.Status,
		RsvpOpen:     e.RsvpOpen,
		MyOrgRole:    e.MyOrgRole,
		MyRsvp:       myRsvp,
		RsvpSummary:  summary,
		CreatedAt:    e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    e.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMCPEventRSVP(r domain.EventRSVPResponse) MCPEventRSVP {
	return MCPEventRSVP{
		ID:          r.ID.String(),
		ParishID:    r.ParishID.String(),
		EventID:     r.EventID.String(),
		UserID:      r.UserID.String(),
		Name:        r.Name,
		Email:       r.Email,
		Phone:       r.Phone,
		OrgRole:     r.OrgRole,
		OrgPosition: r.OrgPosition,
		Status:      r.Status,
		Note:        r.Note,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func parseMCPTimestamp(raw string, endOfDay bool) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	d, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		return d.Add(24*time.Hour - time.Nanosecond), nil
	}
	return d, nil
}

func toMCPEventRSVPSummary(s domain.EventRSVPSummaryResponse) MCPEventRSVPSummary {
	return MCPEventRSVPSummary{
		EventID:            s.EventID.String(),
		Going:              s.Going,
		Maybe:              s.Maybe,
		NotGoing:           s.NotGoing,
		TotalResponded:     s.TotalResponded,
		TotalActiveMembers: s.TotalActiveMembers,
		Unresponded:        s.Unresponded,
	}
}
