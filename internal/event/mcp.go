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
	ParokiID    string  `json:"parokiId"`
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

// MCPEventPetugas is the MCP-safe representation of an EventPetugas assignment.
type MCPEventPetugas struct {
	ID         string  `json:"id"`
	ParokiID   string  `json:"parokiId"`
	EventID    string  `json:"eventId"`
	Role       string  `json:"role"`
	UserID     *string `json:"userId,omitempty"`
	OrgID      *string `json:"orgId,omitempty"`
	Name       string  `json:"name"`
	Urutan     int     `json:"urutan"`
	Status     string  `json:"status"`
	Note       *string `json:"note,omitempty"`
	CanConfirm bool    `json:"canConfirm"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

// MCPEventLagu is the MCP-safe representation of an EventLagu entry.
type MCPEventLagu struct {
	ID        string  `json:"id"`
	ParokiID  string  `json:"parokiId"`
	EventID   string  `json:"eventId"`
	Urutan    int     `json:"urutan"`
	Bagian    string  `json:"bagian"`
	Sumber    *string `json:"sumber,omitempty"`
	Nomor     *string `json:"nomor,omitempty"`
	Judul     string  `json:"judul"`
	NadaDasar *string `json:"nadaDasar,omitempty"`
	Note      *string `json:"note,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// MCPEvent is the MCP-safe representation of an Event (UUIDs and timestamps as strings).
type MCPEvent struct {
	ID           string               `json:"id"`
	ParokiID     string               `json:"parokiId"`
	OrgID        *string              `json:"orgId,omitempty"`
	OrgCode      string               `json:"orgCode,omitempty"`
	OrgName      string               `json:"orgName,omitempty"`
	OrgKind      string               `json:"orgKind,omitempty"`
	CreatedBy    string               `json:"createdBy"`
	Title        string               `json:"title"`
	Jenis        *string              `json:"jenis,omitempty"`
	Description  *string              `json:"description,omitempty"`
	Location     *string              `json:"location,omitempty"`
	StartAt      string               `json:"startAt"`
	EndAt        string               `json:"endAt"`
	Timezone     string               `json:"timezone"`
	Visibility   string               `json:"visibility"`
	RsvpEnabled  bool                 `json:"rsvpEnabled"`
	RsvpDeadline *string              `json:"rsvpDeadline,omitempty"`
	Status       string               `json:"status"`
	RsvpOpen     bool                 `json:"rsvpOpen"`
	CanManage    bool                 `json:"canManage"`
	CanRsvp      bool                 `json:"canRsvp"`
	MyOrgRole    *string              `json:"myOrgRole,omitempty"`
	MyRsvp       *MCPEventRSVP        `json:"myRsvp,omitempty"`
	RsvpSummary  *MCPEventRSVPSummary `json:"rsvpSummary,omitempty"`
	Petugas      []MCPEventPetugas    `json:"petugas,omitempty"`
	Lagu         []MCPEventLagu       `json:"lagu,omitempty"`
	CreatedAt    string               `json:"createdAt"`
	UpdatedAt    string               `json:"updatedAt"`
}

// EventListInput is the input schema for event_list.
type EventListInput struct {
	Query          string `json:"query,omitempty" jsonschema:"Optional text search query matching event title, description, or location"`
	OrgID          string `json:"orgId,omitempty" jsonschema:"Optional organization UUID filter"`
	ParokiWideOnly bool   `json:"parokiWideOnly,omitempty" jsonschema:"If true, restrict results to parish-level events (not owned by an organization)"`
	Jenis          string `json:"jenis,omitempty" jsonschema:"Optional event type code filter (e.g. misa, ibadat, rapat, latihan, rekoleksi, katekese, bakti-sosial, lainnya, or active event_type lookup code)"`
	AssignedToMe   bool   `json:"assignedToMe,omitempty" jsonschema:"If true, restrict results to events where caller or caller's organizations are assigned as petugas"`
	AssignedOrgID  string `json:"assignedOrgId,omitempty" jsonschema:"Optional organization UUID filter for events where the organization is assigned as petugas"`
	From           string `json:"from,omitempty" jsonschema:"Optional inclusive start timestamp in local wall-clock YYYY-MM-DDTHH:MM:SS, YYYY-MM-DD, or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
	To             string `json:"to,omitempty" jsonschema:"Optional inclusive end timestamp in local wall-clock YYYY-MM-DDTHH:MM:SS, YYYY-MM-DD, or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
	Visibility     string `json:"visibility,omitempty" jsonschema:"Optional visibility filter: public or private"`
	Status         string `json:"status,omitempty" jsonschema:"Optional status filter: scheduled or cancelled"`
	MyOrgsOnly     bool   `json:"myOrgsOnly,omitempty" jsonschema:"If true, restrict results to events owned by caller's organizations"`
	Limit          int    `json:"limit,omitempty" jsonschema:"Maximum number of events to return (default 20, max 100)"`
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
	OrgID        string  `json:"orgId,omitempty" jsonschema:"Optional organization UUID; omit to create a parish-level event (requires ref:paroki-admin)"`
	Title        string  `json:"title"`
	Jenis        *string `json:"jenis,omitempty" jsonschema:"Optional event type code (e.g. misa, ibadat, rapat, latihan, rekoleksi, katekese, bakti-sosial, lainnya, or active event_type lookup code)"`
	Description  *string `json:"description,omitempty"`
	Location     *string `json:"location,omitempty"`
	StartAt      string  `json:"startAt" jsonschema:"Start timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
	EndAt        string  `json:"endAt" jsonschema:"End timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
	Visibility   string  `json:"visibility,omitempty"`
	RsvpEnabled  *bool   `json:"rsvpEnabled,omitempty"`
	RsvpDeadline *string `json:"rsvpDeadline,omitempty" jsonschema:"Optional RSVP deadline in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
}

// EventUpdateInput is the input schema for event_update.
type EventUpdateInput struct {
	EventID           string  `json:"eventId"`
	Title             *string `json:"title,omitempty"`
	Jenis             *string `json:"jenis,omitempty" jsonschema:"Optional event type code; pass empty string to clear"`
	Description       *string `json:"description,omitempty"`
	Location          *string `json:"location,omitempty"`
	StartAt           *string `json:"startAt,omitempty" jsonschema:"Start timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
	EndAt             *string `json:"endAt,omitempty" jsonschema:"End timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
	Visibility        *string `json:"visibility,omitempty"`
	RsvpEnabled       *bool   `json:"rsvpEnabled,omitempty"`
	RsvpDeadline      *string `json:"rsvpDeadline,omitempty" jsonschema:"RSVP deadline in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339 with explicit offset (never append Z to a local wall-clock time)"`
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

// EventSetPetugasInput is the input schema for event_set_petugas.
type EventSetPetugasInput struct {
	EventID   string  `json:"eventId"`
	PetugasID string  `json:"petugasId,omitempty" jsonschema:"Existing petugas UUID when updating or removing; omit to add a new officer"`
	Remove    bool    `json:"remove,omitempty" jsonschema:"If true (with petugasId), remove the officer assignment"`
	Role      *string `json:"role,omitempty" jsonschema:"Officer role (e.g. Selebran, Organis, Lektor 1, Tata Tertib)"`
	UserID    *string `json:"userId,omitempty" jsonschema:"Optional user UUID in the active parish (mutually exclusive with orgId)"`
	OrgID     *string `json:"orgId,omitempty" jsonschema:"Optional organization UUID in the active parish (mutually exclusive with userId)"`
	Name      *string `json:"name,omitempty" jsonschema:"Display name (required when neither userId nor orgId is set)"`
	Urutan    *int    `json:"urutan,omitempty" jsonschema:"Display sort order"`
	Status    *string `json:"status,omitempty" jsonschema:"Assignment status: assigned, confirmed, or declined"`
	Note      *string `json:"note,omitempty" jsonschema:"Optional note"`
}

// EventConfirmPetugasInput is the input schema for event_confirm_petugas.
type EventConfirmPetugasInput struct {
	EventID   string  `json:"eventId"`
	PetugasID string  `json:"petugasId"`
	Status    string  `json:"status" jsonschema:"Confirmation status: confirmed or declined"`
	Note      *string `json:"note,omitempty" jsonschema:"Optional confirmation note"`
}

// EventSetLaguInput is the input schema for event_set_lagu.
type EventSetLaguInput struct {
	EventID   string  `json:"eventId"`
	LaguID    string  `json:"laguId,omitempty" jsonschema:"Existing lagu UUID when updating or removing; omit to add a new song entry"`
	Remove    bool    `json:"remove,omitempty" jsonschema:"If true (with laguId), remove the song entry"`
	Urutan    *int    `json:"urutan,omitempty" jsonschema:"Display sort order"`
	Bagian    *string `json:"bagian,omitempty" jsonschema:"Liturgical song part (e.g. Pembuka, Kemuliaan, Komuni, Penutup)"`
	Sumber    *string `json:"sumber,omitempty" jsonschema:"Songbook source (e.g. Puji Syukur, Madah Bakti); pass empty string to clear"`
	Nomor     *string `json:"nomor,omitempty" jsonschema:"Song number; pass empty string to clear"`
	Judul     *string `json:"judul,omitempty" jsonschema:"Song title"`
	NadaDasar *string `json:"nadaDasar,omitempty" jsonschema:"Musical key (e.g. G, Do=F); pass empty string to clear"`
	Note      *string `json:"note,omitempty" jsonschema:"Optional note; pass empty string to clear"`
}

// RegisterMCP registers all event_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_list",
		Description: "List visible events in the caller's active paroki (public paroki events and private events for caller's organizations).",
	}, f.mcpEventList)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_get",
		Description: "Get details of a single event in the caller's active paroki, including RSVP summary and caller's RSVP if an org member.",
	}, f.mcpEventGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_create",
		Description: "Create a new organization-owned event (requires org:admin) or parish-level event (omit orgId, requires ref:paroki-admin) in the caller's active paroki.",
	}, f.mcpEventCreate)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_update",
		Description: "Update or cancel an existing event (requires org:admin for org events or ref:paroki-admin for parish-level events).",
	}, f.mcpEventUpdate)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_rsvp",
		Description: "Submit or update the caller's RSVP for an organization or parish-level event.",
	}, f.mcpEventRSVP)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_list_rsvps",
		Description: "Get RSVP attendance summary and attendee roster for an event.",
	}, f.mcpEventListRSVPs)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_set_petugas",
		Description: "Add, update, or remove an officer/minister assignment on an event (requires event manager).",
	}, f.mcpEventSetPetugas)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_confirm_petugas",
		Description: "Confirm or decline an officer/minister assignment (requires assigned user, org:admin of assigned org, or event manager).",
	}, f.mcpEventConfirmPetugas)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "event_set_lagu",
		Description: "Add, update, or remove a liturgical song entry on an event (requires event manager).",
	}, f.mcpEventSetLagu)
}

func (f *Feature) mcpEventList(ctx context.Context, _ *mcp.CallToolRequest, in EventListInput) (*mcp.CallToolResult, EventListOutput, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, EventListOutput{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)

	var filter *domain.EventSearchFilter
	if in.OrgID != "" || in.ParokiWideOnly || in.Jenis != "" || in.AssignedToMe || in.AssignedOrgID != "" || in.From != "" || in.To != "" || in.Visibility != "" || in.Status != "" || in.MyOrgsOnly {
		filter = &domain.EventSearchFilter{
			Jenis:      in.Jenis,
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
		if in.ParokiWideOnly {
			t := true
			filter.ParokiWideOnly = &t
		}
		if in.AssignedToMe {
			t := true
			filter.AssignedToMe = &t
		}
		if in.AssignedOrgID != "" {
			parsed, err := uuid.Parse(in.AssignedOrgID)
			if err != nil {
				return nil, EventListOutput{}, httpx.BadRequest("Invalid assignedOrgId.")
			}
			filter.AssignedOrgID = &parsed
		}
		if in.From != "" {
			t, err := parseMCPTimestampInLocation(in.From, loc, false)
			if err != nil {
				return nil, EventListOutput{}, httpx.BadRequest("Invalid 'from' timestamp; expected RFC3339, YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DD.")
			}
			filter.From = &t
		}
		if in.To != "" {
			t, err := parseMCPTimestampInLocation(in.To, loc, true)
			if err != nil {
				return nil, EventListOutput{}, httpx.BadRequest("Invalid 'to' timestamp; expected RFC3339, YYYY-MM-DDTHH:MM:SS, or YYYY-MM-DD.")
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
		out = append(out, toMCPEvent(item, loc))
	}
	return nil, EventListOutput{Events: out}, nil
}

func (f *Feature) mcpEventGet(ctx context.Context, _ *mcp.CallToolRequest, in EventGetInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid event ID.")
	}
	res, err := f.service.GetEvent(ctx, p, eventID)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(res, loc), nil
}

func (f *Feature) mcpEventCreate(ctx context.Context, _ *mcp.CallToolRequest, in EventCreateInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)

	var orgID *uuid.UUID
	if in.OrgID != "" {
		parsed, err := uuid.Parse(in.OrgID)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid organization ID.")
		}
		orgID = &parsed
	}
	startAt, err := parseMCPTimestampInLocation(in.StartAt, loc, false)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid 'startAt' timestamp; expected RFC3339 or YYYY-MM-DDTHH:MM:SS.")
	}
	endAt, err := parseMCPTimestampInLocation(in.EndAt, loc, false)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid 'endAt' timestamp; expected RFC3339 or YYYY-MM-DDTHH:MM:SS.")
	}
	var rsvpDeadline *time.Time
	if in.RsvpDeadline != nil && *in.RsvpDeadline != "" {
		t, err := parseMCPTimestampInLocation(*in.RsvpDeadline, loc, false)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'rsvpDeadline' timestamp; expected RFC3339 or YYYY-MM-DDTHH:MM:SS.")
		}
		rsvpDeadline = &t
	}

	res, err := f.service.CreateEvent(ctx, p, domain.CreateEventRequest{
		OrgID:        orgID,
		Title:        in.Title,
		Jenis:        in.Jenis,
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
	return nil, toMCPEvent(res, loc), nil
}

func (f *Feature) mcpEventUpdate(ctx context.Context, _ *mcp.CallToolRequest, in EventUpdateInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid event ID.")
	}

	req := domain.UpdateEventRequest{
		Title:             in.Title,
		Jenis:             in.Jenis,
		Description:       in.Description,
		Location:          in.Location,
		Visibility:        in.Visibility,
		RsvpEnabled:       in.RsvpEnabled,
		ClearRsvpDeadline: in.ClearRsvpDeadline,
		Status:            in.Status,
	}
	if in.StartAt != nil {
		t, err := parseMCPTimestampInLocation(*in.StartAt, loc, false)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'startAt' timestamp; expected RFC3339 or YYYY-MM-DDTHH:MM:SS.")
		}
		req.StartAt = &t
	}
	if in.EndAt != nil {
		t, err := parseMCPTimestampInLocation(*in.EndAt, loc, false)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'endAt' timestamp; expected RFC3339 or YYYY-MM-DDTHH:MM:SS.")
		}
		req.EndAt = &t
	}
	if in.RsvpDeadline != nil && *in.RsvpDeadline != "" {
		t, err := parseMCPTimestampInLocation(*in.RsvpDeadline, loc, false)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid 'rsvpDeadline' timestamp; expected RFC3339 or YYYY-MM-DDTHH:MM:SS.")
		}
		req.RsvpDeadline = &t
	}

	res, err := f.service.UpdateEvent(ctx, p, eventID, req)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(res, loc), nil
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

func (f *Feature) mcpEventSetPetugas(ctx context.Context, _ *mcp.CallToolRequest, in EventSetPetugasInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid event ID.")
	}

	if in.Remove {
		petugasID, err := uuid.Parse(in.PetugasID)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid petugas ID.")
		}
		if err := f.service.DeleteEventPetugas(ctx, p, eventID, petugasID); err != nil {
			return nil, MCPEvent{}, err
		}
	} else {
		var userID *uuid.UUID
		if in.UserID != nil {
			if *in.UserID == "" {
				nilID := uuid.Nil
				userID = &nilID
			} else {
				parsed, err := uuid.Parse(*in.UserID)
				if err != nil {
					return nil, MCPEvent{}, httpx.BadRequest("Invalid userId.")
				}
				userID = &parsed
			}
		}
		var orgID *uuid.UUID
		if in.OrgID != nil {
			if *in.OrgID == "" {
				nilID := uuid.Nil
				orgID = &nilID
			} else {
				parsed, err := uuid.Parse(*in.OrgID)
				if err != nil {
					return nil, MCPEvent{}, httpx.BadRequest("Invalid orgId.")
				}
				orgID = &parsed
			}
		}
		if in.PetugasID != "" {
			petugasID, err := uuid.Parse(in.PetugasID)
			if err != nil {
				return nil, MCPEvent{}, httpx.BadRequest("Invalid petugas ID.")
			}
			if _, err := f.service.UpdateEventPetugas(ctx, p, eventID, petugasID, domain.UpdateEventPetugasRequest{
				Role:   in.Role,
				UserID: userID,
				OrgID:  orgID,
				Name:   in.Name,
				Urutan: in.Urutan,
				Status: in.Status,
				Note:   in.Note,
			}); err != nil {
				return nil, MCPEvent{}, err
			}
		} else {
			var role, name, status string
			if in.Role != nil {
				role = *in.Role
			}
			if in.Name != nil {
				name = *in.Name
			}
			if in.Status != nil {
				status = *in.Status
			}
			if _, err := f.service.CreateEventPetugas(ctx, p, eventID, domain.CreateEventPetugasRequest{
				Role:   role,
				UserID: userID,
				OrgID:  orgID,
				Name:   name,
				Urutan: in.Urutan,
				Status: status,
				Note:   in.Note,
			}); err != nil {
				return nil, MCPEvent{}, err
			}
		}
	}

	ev, err := f.service.GetEvent(ctx, p, eventID)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(ev, loc), nil
}

func (f *Feature) mcpEventConfirmPetugas(ctx context.Context, _ *mcp.CallToolRequest, in EventConfirmPetugasInput) (*mcp.CallToolResult, MCPEventPetugas, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEventPetugas{}, err
	}
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEventPetugas{}, httpx.BadRequest("Invalid event ID.")
	}
	petugasID, err := uuid.Parse(in.PetugasID)
	if err != nil {
		return nil, MCPEventPetugas{}, httpx.BadRequest("Invalid petugas ID.")
	}
	status := in.Status
	res, err := f.service.UpdateEventPetugas(ctx, p, eventID, petugasID, domain.UpdateEventPetugasRequest{
		Status: &status,
		Note:   in.Note,
	})
	if err != nil {
		return nil, MCPEventPetugas{}, err
	}
	return nil, toMCPEventPetugas(res), nil
}

func (f *Feature) mcpEventSetLagu(ctx context.Context, _ *mcp.CallToolRequest, in EventSetLaguInput) (*mcp.CallToolResult, MCPEvent, error) {
	p, err := requireMCPTenant(ctx)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	eventID, err := uuid.Parse(in.EventID)
	if err != nil {
		return nil, MCPEvent{}, httpx.BadRequest("Invalid event ID.")
	}

	if in.Remove {
		laguID, err := uuid.Parse(in.LaguID)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid lagu ID.")
		}
		if err := f.service.DeleteEventLagu(ctx, p, eventID, laguID); err != nil {
			return nil, MCPEvent{}, err
		}
	} else if in.LaguID != "" {
		laguID, err := uuid.Parse(in.LaguID)
		if err != nil {
			return nil, MCPEvent{}, httpx.BadRequest("Invalid lagu ID.")
		}
		if _, err := f.service.UpdateEventLagu(ctx, p, eventID, laguID, domain.UpdateEventLaguRequest{
			Urutan:    in.Urutan,
			Bagian:    in.Bagian,
			Sumber:    in.Sumber,
			Nomor:     in.Nomor,
			Judul:     in.Judul,
			NadaDasar: in.NadaDasar,
			Note:      in.Note,
		}); err != nil {
			return nil, MCPEvent{}, err
		}
	} else {
		var bagian, judul string
		if in.Bagian != nil {
			bagian = *in.Bagian
		}
		if in.Judul != nil {
			judul = *in.Judul
		}
		if _, err := f.service.CreateEventLagu(ctx, p, eventID, domain.CreateEventLaguRequest{
			Urutan:    in.Urutan,
			Bagian:    bagian,
			Sumber:    in.Sumber,
			Nomor:     in.Nomor,
			Judul:     judul,
			NadaDasar: in.NadaDasar,
			Note:      in.Note,
		}); err != nil {
			return nil, MCPEvent{}, err
		}
	}

	ev, err := f.service.GetEvent(ctx, p, eventID)
	if err != nil {
		return nil, MCPEvent{}, err
	}
	return nil, toMCPEvent(ev, loc), nil
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

func toMCPEvent(e domain.EventResponse, loc *time.Location) MCPEvent {
	if loc == nil {
		loc = time.UTC
	}
	var orgID *string
	if e.OrgID != nil {
		s := e.OrgID.String()
		orgID = &s
	}
	var deadline *string
	if e.RsvpDeadline != nil {
		s := e.RsvpDeadline.In(loc).Format(time.RFC3339)
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
	var petugas []MCPEventPetugas
	if len(e.Petugas) > 0 {
		petugas = make([]MCPEventPetugas, 0, len(e.Petugas))
		for _, item := range e.Petugas {
			petugas = append(petugas, toMCPEventPetugas(item))
		}
	}
	var lagu []MCPEventLagu
	if len(e.Lagu) > 0 {
		lagu = make([]MCPEventLagu, 0, len(e.Lagu))
		for _, item := range e.Lagu {
			lagu = append(lagu, toMCPEventLagu(item))
		}
	}
	return MCPEvent{
		ID:           e.ID.String(),
		ParokiID:     e.ParokiID.String(),
		OrgID:        orgID,
		OrgCode:      e.OrgCode,
		OrgName:      e.OrgName,
		OrgKind:      e.OrgKind,
		CreatedBy:    e.CreatedBy.String(),
		Title:        e.Title,
		Jenis:        e.Jenis,
		Description:  e.Description,
		Location:     e.Location,
		StartAt:      e.StartAt.In(loc).Format(time.RFC3339),
		EndAt:        e.EndAt.In(loc).Format(time.RFC3339),
		Timezone:     loc.String(),
		Visibility:   e.Visibility,
		RsvpEnabled:  e.RsvpEnabled,
		RsvpDeadline: deadline,
		Status:       e.Status,
		RsvpOpen:     e.RsvpOpen,
		CanManage:    e.CanManage,
		CanRsvp:      e.CanRsvp,
		MyOrgRole:    e.MyOrgRole,
		MyRsvp:       myRsvp,
		RsvpSummary:  summary,
		Petugas:      petugas,
		Lagu:         lagu,
		CreatedAt:    e.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    e.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMCPEventPetugas(p domain.EventPetugasResponse) MCPEventPetugas {
	var userID *string
	if p.UserID != nil {
		s := p.UserID.String()
		userID = &s
	}
	var orgID *string
	if p.OrgID != nil {
		s := p.OrgID.String()
		orgID = &s
	}
	return MCPEventPetugas{
		ID:         p.ID.String(),
		ParokiID:   p.ParokiID.String(),
		EventID:    p.EventID.String(),
		Role:       p.Role,
		UserID:     userID,
		OrgID:      orgID,
		Name:       p.Name,
		Urutan:     p.Urutan,
		Status:     p.Status,
		Note:       p.Note,
		CanConfirm: p.CanConfirm,
		CreatedAt:  p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMCPEventLagu(l domain.EventLaguResponse) MCPEventLagu {
	return MCPEventLagu{
		ID:        l.ID.String(),
		ParokiID:  l.ParokiID.String(),
		EventID:   l.EventID.String(),
		Urutan:    l.Urutan,
		Bagian:    l.Bagian,
		Sumber:    l.Sumber,
		Nomor:     l.Nomor,
		Judul:     l.Judul,
		NadaDasar: l.NadaDasar,
		Note:      l.Note,
		CreatedAt: l.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: l.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMCPEventRSVP(r domain.EventRSVPResponse) MCPEventRSVP {
	return MCPEventRSVP{
		ID:          r.ID.String(),
		ParokiID:    r.ParokiID.String(),
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

func parseMCPTimestampInLocation(raw string, loc *time.Location, endOfDay bool) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02T15:04"} {
		if t, err := time.ParseInLocation(layout, raw, loc); err == nil {
			return t.UTC(), nil
		}
	}
	d, err := time.ParseInLocation(time.DateOnly, raw, loc)
	if err != nil {
		return time.Time{}, err
	}
	if endOfDay {
		return d.Add(24*time.Hour - time.Nanosecond).UTC(), nil
	}
	return d.UTC(), nil
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
