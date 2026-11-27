package liturgi

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/liturgi/domain"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/user"
)

// MCPLiturgiMinistry is the MCP-safe representation of a liturgical ministry configuration.
type MCPLiturgiMinistry struct {
	ID           *string `json:"id,omitempty"`
	ParokiID     string  `json:"parokiId"`
	RoleSlug     string  `json:"roleSlug"`
	OrgID        *string `json:"orgId,omitempty"`
	OrgName      *string `json:"orgName,omitempty"`
	LeaderUserID *string `json:"leaderUserId,omitempty"`
	LeaderName   *string `json:"leaderName,omitempty"`
	LeaderEmail  *string `json:"leaderEmail,omitempty"`
	CanManage    bool    `json:"canManage"`
}

// MCPLiturgiPetugas is the MCP-safe representation of a liturgical officer assignment.
type MCPLiturgiPetugas struct {
	ID                   string  `json:"id"`
	ParokiID             string  `json:"parokiId"`
	EventID              string  `json:"eventId"`
	Role                 string  `json:"role"`
	SlotLabel            *string `json:"slotLabel,omitempty"`
	UserID               *string `json:"userId,omitempty"`
	OrgID                *string `json:"orgId,omitempty"`
	Name                 string  `json:"name"`
	Urutan               int     `json:"urutan"`
	Status               string  `json:"status"`
	Note                 *string `json:"note,omitempty"`
	RespondedAt          *string `json:"respondedAt,omitempty"`
	RespondedBy          *string `json:"respondedBy,omitempty"`
	SubstituteUserID     *string `json:"substituteUserId,omitempty"`
	SubstituteOrgID      *string `json:"substituteOrgId,omitempty"`
	SubstituteName       *string `json:"substituteName,omitempty"`
	CanConfirm           bool    `json:"canConfirm"`
	CanEdit              bool    `json:"canEdit"`
	CanRespond           bool    `json:"canRespond"`
	CanApproveSubstitute bool    `json:"canApproveSubstitute"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

// MCPLiturgiLagu is the MCP-safe representation of a liturgical song entry.
type MCPLiturgiLagu struct {
	ID        string  `json:"id"`
	ParokiID  string  `json:"parokiId"`
	EventID   string  `json:"eventId"`
	Urutan    int     `json:"urutan"`
	Bagian    string  `json:"bagian"`
	CantusID  *string `json:"cantusId,omitempty"`
	Sumber    *string `json:"sumber,omitempty"`
	Nomor     *string `json:"nomor,omitempty"`
	Judul     string  `json:"judul"`
	NadaDasar *string `json:"nadaDasar,omitempty"`
	Note      *string `json:"note,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// MCPLiturgiSchedule is the MCP-safe representation of a parish Mass schedule.
type MCPLiturgiSchedule struct {
	ID                string              `json:"id"`
	ParokiID          string              `json:"parokiId"`
	CreatedBy         string              `json:"createdBy"`
	Title             string              `json:"title"`
	Description       *string             `json:"description,omitempty"`
	Location          *string             `json:"location,omitempty"`
	StartAt           string              `json:"startAt"`
	EndAt             string              `json:"endAt"`
	Timezone          string              `json:"timezone"`
	Visibility        string              `json:"visibility"`
	RsvpEnabled       bool                `json:"rsvpEnabled"`
	RsvpDeadline      *string             `json:"rsvpDeadline,omitempty"`
	Status            string              `json:"status"`
	RsvpOpen          bool                `json:"rsvpOpen"`
	CanRsvp           bool                `json:"canRsvp"`
	CanManageSchedule bool                `json:"canManageSchedule"`
	ManageableRoles   []string            `json:"manageableRoles"`
	CanManageLagu     bool                `json:"canManageLagu"`
	Petugas           []MCPLiturgiPetugas `json:"petugas"`
	Lagu              []MCPLiturgiLagu    `json:"lagu"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
}

// LiturgiListSchedulesInput is the input schema for liturgi_list_schedules.
type LiturgiListSchedulesInput struct {
	Query         string `json:"query,omitempty" jsonschema:"Optional text search query matching title, description, or location"`
	From          string `json:"from,omitempty" jsonschema:"Optional inclusive start timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS, YYYY-MM-DD, or RFC3339"`
	To            string `json:"to,omitempty" jsonschema:"Optional inclusive end timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS, YYYY-MM-DD, or RFC3339"`
	Status        string `json:"status,omitempty" jsonschema:"Optional status filter: scheduled or cancelled"`
	Role          string `json:"role,omitempty" jsonschema:"Optional ministry role filter (lektor, mazmur, pembagi-komuni, misdinar, organis, kor)"`
	AssignedToMe  *bool  `json:"assignedToMe,omitempty"`
	AssignedOrgID string `json:"assignedOrgId,omitempty"`
	HasDeclined   *bool  `json:"hasDeclined,omitempty"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum number of schedules to return (default 20, max 100)"`
}

// LiturgiListSchedulesOutput is the output schema for liturgi_list_schedules.
type LiturgiListSchedulesOutput struct {
	Schedules []MCPLiturgiSchedule `json:"schedules"`
}

// LiturgiGetScheduleInput is the input schema for liturgi_get_schedule.
type LiturgiGetScheduleInput struct {
	EventID string `json:"eventId"`
}

// LiturgiCreateScheduleInput is the input schema for liturgi_create_schedule.
type LiturgiCreateScheduleInput struct {
	Title        string  `json:"title"`
	Description  *string `json:"description,omitempty"`
	Location     *string `json:"location,omitempty"`
	StartAt      string  `json:"startAt" jsonschema:"Start timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339"`
	EndAt        string  `json:"endAt" jsonschema:"End timestamp in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339"`
	RsvpEnabled  *bool   `json:"rsvpEnabled,omitempty"`
	RsvpDeadline *string `json:"rsvpDeadline,omitempty" jsonschema:"Optional RSVP deadline in parish local wall-clock YYYY-MM-DDTHH:MM:SS or RFC3339"`
}

// LiturgiUpdateScheduleInput is the input schema for liturgi_update_schedule.
type LiturgiUpdateScheduleInput struct {
	EventID           string  `json:"eventId"`
	Title             *string `json:"title,omitempty"`
	Description       *string `json:"description,omitempty"`
	Location          *string `json:"location,omitempty"`
	StartAt           *string `json:"startAt,omitempty"`
	EndAt             *string `json:"endAt,omitempty"`
	RsvpEnabled       *bool   `json:"rsvpEnabled,omitempty"`
	RsvpDeadline      *string `json:"rsvpDeadline,omitempty"`
	ClearRsvpDeadline *bool   `json:"clearRsvpDeadline,omitempty"`
	Status            *string `json:"status,omitempty"`
}

// LiturgiSetMinistryInput is the input schema for liturgi_set_ministry.
type LiturgiSetMinistryInput struct {
	RoleSlug          string `json:"roleSlug" jsonschema:"Ministry role slug: lektor, mazmur, pembagi-komuni, misdinar, organis, or kor"`
	OrgID             string `json:"orgId,omitempty"`
	ClearOrgID        *bool  `json:"clearOrgId,omitempty"`
	LeaderUserID      string `json:"leaderUserId,omitempty"`
	ClearLeaderUserID *bool  `json:"clearLeaderUserId,omitempty"`
}

// LiturgiSetPetugasInput is the input schema for liturgi_set_petugas.
type LiturgiSetPetugasInput struct {
	Action    string  `json:"action" jsonschema:"Action to perform: set or remove"`
	EventID   string  `json:"eventId"`
	PetugasID string  `json:"petugasId,omitempty" jsonschema:"Required for remove, or to update an existing assignment on set"`
	Role      *string `json:"role,omitempty"`
	SlotLabel *string `json:"slotLabel,omitempty"`
	UserID    string  `json:"userId,omitempty"`
	ClearUser *bool   `json:"clearUser,omitempty"`
	OrgID     string  `json:"orgId,omitempty"`
	ClearOrg  *bool   `json:"clearOrg,omitempty"`
	Name      *string `json:"name,omitempty"`
	Urutan    *int    `json:"urutan,omitempty"`
	Status    *string `json:"status,omitempty"`
	Note      *string `json:"note,omitempty"`
}

// LiturgiDeleteOutput is a simple deletion confirmation output.
type LiturgiDeleteOutput struct {
	Deleted bool `json:"deleted"`
}

// LiturgiRespondPetugasInput is the input schema for liturgi_respond_petugas.
type LiturgiRespondPetugasInput struct {
	EventID          string  `json:"eventId"`
	PetugasID        string  `json:"petugasId"`
	Status           string  `json:"status" jsonschema:"Response status: confirmed or declined"`
	Note             *string `json:"note,omitempty"`
	SubstituteUserID string  `json:"substituteUserId,omitempty"`
	SubstituteOrgID  string  `json:"substituteOrgId,omitempty"`
	SubstituteName   *string `json:"substituteName,omitempty"`
}

// LiturgiApproveSubstituteInput is the input schema for liturgi_approve_substitute.
type LiturgiApproveSubstituteInput struct {
	EventID       string `json:"eventId"`
	PetugasID     string `json:"petugasId"`
	MarkConfirmed bool   `json:"markConfirmed,omitempty"`
}

// LiturgiSetLaguInput is the input schema for liturgi_set_lagu.
type LiturgiSetLaguInput struct {
	Action        string  `json:"action" jsonschema:"Action to perform: set or remove"`
	EventID       string  `json:"eventId"`
	LaguID        string  `json:"laguId,omitempty" jsonschema:"Required for remove, or to update an existing song entry on set"`
	Urutan        *int    `json:"urutan,omitempty"`
	Bagian        *string `json:"bagian,omitempty"`
	CantusID      string  `json:"cantusId,omitempty"`
	ClearCantusID *bool   `json:"clearCantusId,omitempty"`
	Sumber        *string `json:"sumber,omitempty"`
	Nomor         *string `json:"nomor,omitempty"`
	Judul         *string `json:"judul,omitempty"`
	NadaDasar     *string `json:"nadaDasar,omitempty"`
	Note          *string `json:"note,omitempty"`
}

// RegisterMCP registers all 9 liturgi_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_list_schedules",
		Description: "List parish Mass schedules in the caller's active parish with officer assignments and song lists.",
	}, f.mcpListSchedules)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_get_schedule",
		Description: "Get a single parish Mass schedule by eventId with officers, songs, and permissions.",
	}, f.mcpGetSchedule)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_create_schedule",
		Description: "Create a new parish Mass schedule (requires liturgi:admin or ref:paroki-admin).",
	}, f.mcpCreateSchedule)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_update_schedule",
		Description: "Update or cancel a parish Mass schedule (requires liturgi:admin or ref:paroki-admin).",
	}, f.mcpUpdateSchedule)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_set_ministry",
		Description: "Configure the organization and/or leader user for a liturgical ministry role (requires liturgi:admin or ref:paroki-admin).",
	}, f.mcpSetMinistry)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_set_petugas",
		Description: "Assign, update, or remove a liturgical officer on a parish Mass schedule.",
	}, f.mcpSetPetugas)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_respond_petugas",
		Description: "Confirm or decline a liturgical officer assignment, optionally proposing a substitute.",
	}, f.mcpRespondPetugas)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_approve_substitute",
		Description: "Approve and promote a proposed substitute onto an officer assignment.",
	}, f.mcpApproveSubstitute)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "liturgi_set_lagu",
		Description: "Add, update, or remove a liturgical song entry on a parish Mass schedule (supports cantusId auto-fill).",
	}, f.mcpSetLagu)
}

func (f *Feature) mcpListSchedules(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiListSchedulesInput) (*mcp.CallToolResult, LiturgiListSchedulesOutput, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, LiturgiListSchedulesOutput{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)

	var fromT, toT *time.Time
	if strings.TrimSpace(in.From) != "" {
		t, err := parseMCPTimestampInLocation(strings.TrimSpace(in.From), loc, false)
		if err != nil {
			return nil, LiturgiListSchedulesOutput{}, httpx.BadRequest("Invalid from timestamp.")
		}
		fromT = &t
	}
	if strings.TrimSpace(in.To) != "" {
		t, err := parseMCPTimestampInLocation(strings.TrimSpace(in.To), loc, true)
		if err != nil {
			return nil, LiturgiListSchedulesOutput{}, httpx.BadRequest("Invalid to timestamp.")
		}
		toT = &t
	}
	var assignedOrgID *uuid.UUID
	if strings.TrimSpace(in.AssignedOrgID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(in.AssignedOrgID))
		if err != nil {
			return nil, LiturgiListSchedulesOutput{}, httpx.BadRequest("Invalid assignedOrgId.")
		}
		assignedOrgID = &id
	}

	limit := in.Limit
	if limit <= 0 {
		limit = httpx.DefaultLimit
	}
	if limit > httpx.MaxLimit {
		limit = httpx.MaxLimit
	}

	items, _, err := f.service.SearchSchedules(ctx, p, domain.SearchLiturgiSchedulesRequest{
		Q: in.Query,
		Filter: &domain.LiturgiScheduleSearchFilter{
			From:          fromT,
			To:            toT,
			Status:        in.Status,
			Role:          in.Role,
			AssignedToMe:  in.AssignedToMe,
			AssignedOrgID: assignedOrgID,
			HasDeclined:   in.HasDeclined,
		},
		Limit: limit,
	}, nil, nil, limit)
	if err != nil {
		return nil, LiturgiListSchedulesOutput{}, err
	}

	out := make([]MCPLiturgiSchedule, 0, len(items))
	for _, item := range items {
		out = append(out, toMCPLiturgiSchedule(item, loc))
	}
	return nil, LiturgiListSchedulesOutput{Schedules: out}, nil
}

func (f *Feature) mcpGetSchedule(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiGetScheduleInput) (*mcp.CallToolResult, MCPLiturgiSchedule, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, err
	}
	eventID, err := uuid.Parse(strings.TrimSpace(in.EventID))
	if err != nil {
		return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid eventId.")
	}
	item, err := f.service.GetSchedule(ctx, p, eventID)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	return nil, toMCPLiturgiSchedule(item, loc), nil
}

func (f *Feature) mcpCreateSchedule(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiCreateScheduleInput) (*mcp.CallToolResult, MCPLiturgiSchedule, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, err
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	startAt, err := parseMCPTimestampInLocation(strings.TrimSpace(in.StartAt), loc, false)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid startAt timestamp.")
	}
	endAt, err := parseMCPTimestampInLocation(strings.TrimSpace(in.EndAt), loc, false)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid endAt timestamp.")
	}
	var deadline *time.Time
	if in.RsvpDeadline != nil && strings.TrimSpace(*in.RsvpDeadline) != "" {
		t, err := parseMCPTimestampInLocation(strings.TrimSpace(*in.RsvpDeadline), loc, false)
		if err != nil {
			return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid rsvpDeadline timestamp.")
		}
		deadline = &t
	}
	created, err := f.service.CreateSchedule(ctx, p, domain.CreateLiturgiScheduleRequest{
		Title:        in.Title,
		Description:  in.Description,
		Location:     in.Location,
		StartAt:      startAt,
		EndAt:        endAt,
		RsvpEnabled:  in.RsvpEnabled,
		RsvpDeadline: deadline,
	})
	if err != nil {
		return nil, MCPLiturgiSchedule{}, err
	}
	return nil, toMCPLiturgiSchedule(created, loc), nil
}

func (f *Feature) mcpUpdateSchedule(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiUpdateScheduleInput) (*mcp.CallToolResult, MCPLiturgiSchedule, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, err
	}
	eventID, err := uuid.Parse(strings.TrimSpace(in.EventID))
	if err != nil {
		return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid eventId.")
	}
	loc := f.service.ResolveParokiLocation(ctx, p.ParokiID)
	req := domain.UpdateLiturgiScheduleRequest{
		Title:             in.Title,
		Description:       in.Description,
		Location:          in.Location,
		RsvpEnabled:       in.RsvpEnabled,
		ClearRsvpDeadline: in.ClearRsvpDeadline,
		Status:            in.Status,
	}
	if in.StartAt != nil && strings.TrimSpace(*in.StartAt) != "" {
		t, err := parseMCPTimestampInLocation(strings.TrimSpace(*in.StartAt), loc, false)
		if err != nil {
			return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid startAt timestamp.")
		}
		req.StartAt = &t
	}
	if in.EndAt != nil && strings.TrimSpace(*in.EndAt) != "" {
		t, err := parseMCPTimestampInLocation(strings.TrimSpace(*in.EndAt), loc, false)
		if err != nil {
			return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid endAt timestamp.")
		}
		req.EndAt = &t
	}
	if in.RsvpDeadline != nil && strings.TrimSpace(*in.RsvpDeadline) != "" {
		t, err := parseMCPTimestampInLocation(strings.TrimSpace(*in.RsvpDeadline), loc, false)
		if err != nil {
			return nil, MCPLiturgiSchedule{}, httpx.BadRequest("Invalid rsvpDeadline timestamp.")
		}
		req.RsvpDeadline = &t
	}
	updated, err := f.service.UpdateSchedule(ctx, p, eventID, req)
	if err != nil {
		return nil, MCPLiturgiSchedule{}, err
	}
	return nil, toMCPLiturgiSchedule(updated, loc), nil
}

func (f *Feature) mcpSetMinistry(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiSetMinistryInput) (*mcp.CallToolResult, MCPLiturgiMinistry, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, MCPLiturgiMinistry{}, err
	}
	req := domain.UpdateLiturgiMinistryRequest{
		ClearOrgID:        in.ClearOrgID,
		ClearLeaderUserID: in.ClearLeaderUserID,
	}
	if strings.TrimSpace(in.OrgID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(in.OrgID))
		if err != nil {
			return nil, MCPLiturgiMinistry{}, httpx.BadRequest("Invalid orgId.")
		}
		req.OrgID = &id
	}
	if strings.TrimSpace(in.LeaderUserID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(in.LeaderUserID))
		if err != nil {
			return nil, MCPLiturgiMinistry{}, httpx.BadRequest("Invalid leaderUserId.")
		}
		req.LeaderUserID = &id
	}
	res, err := f.service.UpdateMinistry(ctx, p, strings.TrimSpace(in.RoleSlug), req)
	if err != nil {
		return nil, MCPLiturgiMinistry{}, err
	}
	return nil, toMCPLiturgiMinistry(res), nil
}

func (f *Feature) mcpSetPetugas(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiSetPetugasInput) (*mcp.CallToolResult, any, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, nil, err
	}
	eventID, err := uuid.Parse(strings.TrimSpace(in.EventID))
	if err != nil {
		return nil, nil, httpx.BadRequest("Invalid eventId.")
	}
	action := strings.ToLower(strings.TrimSpace(in.Action))
	switch action {
	case "remove":
		petugasID, err := uuid.Parse(strings.TrimSpace(in.PetugasID))
		if err != nil {
			return nil, nil, httpx.BadRequest("Invalid petugasId.")
		}
		if err := f.service.DeletePetugas(ctx, p, eventID, petugasID); err != nil {
			return nil, nil, err
		}
		return nil, LiturgiDeleteOutput{Deleted: true}, nil
	case "set":
		var userID, orgID *uuid.UUID
		if strings.TrimSpace(in.UserID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(in.UserID))
			if err != nil {
				return nil, nil, httpx.BadRequest("Invalid userId.")
			}
			userID = &id
		}
		if strings.TrimSpace(in.OrgID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(in.OrgID))
			if err != nil {
				return nil, nil, httpx.BadRequest("Invalid orgId.")
			}
			orgID = &id
		}
		if strings.TrimSpace(in.PetugasID) != "" {
			petugasID, err := uuid.Parse(strings.TrimSpace(in.PetugasID))
			if err != nil {
				return nil, nil, httpx.BadRequest("Invalid petugasId.")
			}
			res, err := f.service.UpdatePetugas(ctx, p, eventID, petugasID, domain.UpdateLiturgiPetugasRequest{
				Role:      in.Role,
				SlotLabel: in.SlotLabel,
				UserID:    userID,
				ClearUser: in.ClearUser,
				OrgID:     orgID,
				ClearOrg:  in.ClearOrg,
				Name:      in.Name,
				Urutan:    in.Urutan,
				Status:    in.Status,
				Note:      in.Note,
			})
			if err != nil {
				return nil, nil, err
			}
			return nil, toMCPLiturgiPetugas(res), nil
		}
		role := ""
		if in.Role != nil {
			role = *in.Role
		}
		name := ""
		if in.Name != nil {
			name = *in.Name
		}
		res, err := f.service.CreatePetugas(ctx, p, eventID, domain.CreateLiturgiPetugasRequest{
			Role:      role,
			SlotLabel: in.SlotLabel,
			UserID:    userID,
			OrgID:     orgID,
			Name:      name,
			Urutan:    in.Urutan,
			Status:    in.Status,
			Note:      in.Note,
		})
		if err != nil {
			return nil, nil, err
		}
		return nil, toMCPLiturgiPetugas(res), nil
	default:
		return nil, nil, httpx.BadRequest("action must be 'set' or 'remove'.")
	}
}

func (f *Feature) mcpRespondPetugas(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiRespondPetugasInput) (*mcp.CallToolResult, MCPLiturgiPetugas, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, MCPLiturgiPetugas{}, err
	}
	eventID, err := uuid.Parse(strings.TrimSpace(in.EventID))
	if err != nil {
		return nil, MCPLiturgiPetugas{}, httpx.BadRequest("Invalid eventId.")
	}
	petugasID, err := uuid.Parse(strings.TrimSpace(in.PetugasID))
	if err != nil {
		return nil, MCPLiturgiPetugas{}, httpx.BadRequest("Invalid petugasId.")
	}
	req := domain.RespondLiturgiPetugasRequest{
		Status:         strings.TrimSpace(in.Status),
		Note:           in.Note,
		SubstituteName: in.SubstituteName,
	}
	if strings.TrimSpace(in.SubstituteUserID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(in.SubstituteUserID))
		if err != nil {
			return nil, MCPLiturgiPetugas{}, httpx.BadRequest("Invalid substituteUserId.")
		}
		req.SubstituteUserID = &id
	}
	if strings.TrimSpace(in.SubstituteOrgID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(in.SubstituteOrgID))
		if err != nil {
			return nil, MCPLiturgiPetugas{}, httpx.BadRequest("Invalid substituteOrgId.")
		}
		req.SubstituteOrgID = &id
	}
	res, err := f.service.RespondPetugas(ctx, p, eventID, petugasID, req)
	if err != nil {
		return nil, MCPLiturgiPetugas{}, err
	}
	return nil, toMCPLiturgiPetugas(res), nil
}

func (f *Feature) mcpApproveSubstitute(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiApproveSubstituteInput) (*mcp.CallToolResult, MCPLiturgiPetugas, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, MCPLiturgiPetugas{}, err
	}
	eventID, err := uuid.Parse(strings.TrimSpace(in.EventID))
	if err != nil {
		return nil, MCPLiturgiPetugas{}, httpx.BadRequest("Invalid eventId.")
	}
	petugasID, err := uuid.Parse(strings.TrimSpace(in.PetugasID))
	if err != nil {
		return nil, MCPLiturgiPetugas{}, httpx.BadRequest("Invalid petugasId.")
	}
	res, err := f.service.ApproveSubstitute(ctx, p, eventID, petugasID, domain.ApproveLiturgiSubstituteRequest{
		MarkConfirmed: in.MarkConfirmed,
	})
	if err != nil {
		return nil, MCPLiturgiPetugas{}, err
	}
	return nil, toMCPLiturgiPetugas(res), nil
}

func (f *Feature) mcpSetLagu(ctx context.Context, _ *mcp.CallToolRequest, in LiturgiSetLaguInput) (*mcp.CallToolResult, any, error) {
	p, err := user.RequireTenant(ctx)
	if err != nil {
		return nil, nil, err
	}
	eventID, err := uuid.Parse(strings.TrimSpace(in.EventID))
	if err != nil {
		return nil, nil, httpx.BadRequest("Invalid eventId.")
	}
	action := strings.ToLower(strings.TrimSpace(in.Action))
	switch action {
	case "remove":
		laguID, err := uuid.Parse(strings.TrimSpace(in.LaguID))
		if err != nil {
			return nil, nil, httpx.BadRequest("Invalid laguId.")
		}
		if err := f.service.DeleteLagu(ctx, p, eventID, laguID); err != nil {
			return nil, nil, err
		}
		return nil, LiturgiDeleteOutput{Deleted: true}, nil
	case "set":
		var cantusID *uuid.UUID
		if strings.TrimSpace(in.CantusID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(in.CantusID))
			if err != nil {
				return nil, nil, httpx.BadRequest("Invalid cantusId.")
			}
			cantusID = &id
		}
		if strings.TrimSpace(in.LaguID) != "" {
			laguID, err := uuid.Parse(strings.TrimSpace(in.LaguID))
			if err != nil {
				return nil, nil, httpx.BadRequest("Invalid laguId.")
			}
			res, err := f.service.UpdateLagu(ctx, p, eventID, laguID, domain.UpdateLiturgiLaguRequest{
				Urutan:        in.Urutan,
				Bagian:        in.Bagian,
				CantusID:      cantusID,
				ClearCantusID: in.ClearCantusID,
				Sumber:        in.Sumber,
				Nomor:         in.Nomor,
				Judul:         in.Judul,
				NadaDasar:     in.NadaDasar,
				Note:          in.Note,
			})
			if err != nil {
				return nil, nil, err
			}
			return nil, toMCPLiturgiLagu(res), nil
		}
		bagian := ""
		if in.Bagian != nil {
			bagian = *in.Bagian
		}
		judul := ""
		if in.Judul != nil {
			judul = *in.Judul
		}
		res, err := f.service.CreateLagu(ctx, p, eventID, domain.CreateLiturgiLaguRequest{
			Urutan:    in.Urutan,
			Bagian:    bagian,
			CantusID:  cantusID,
			Sumber:    in.Sumber,
			Nomor:     in.Nomor,
			Judul:     judul,
			NadaDasar: in.NadaDasar,
			Note:      in.Note,
		})
		if err != nil {
			return nil, nil, err
		}
		return nil, toMCPLiturgiLagu(res), nil
	default:
		return nil, nil, httpx.BadRequest("action must be 'set' or 'remove'.")
	}
}

func toMCPLiturgiMinistry(m domain.LiturgiMinistryResponse) MCPLiturgiMinistry {
	return MCPLiturgiMinistry{
		ID:           uuidPtrStr(m.ID),
		ParokiID:     m.ParokiID.String(),
		RoleSlug:     m.RoleSlug,
		OrgID:        uuidPtrStr(m.OrgID),
		OrgName:      m.OrgName,
		LeaderUserID: uuidPtrStr(m.LeaderUserID),
		LeaderName:   m.LeaderName,
		LeaderEmail:  m.LeaderEmail,
		CanManage:    m.CanManage,
	}
}

func toMCPLiturgiPetugas(p domain.LiturgiPetugasResponse) MCPLiturgiPetugas {
	var respondedAt *string
	if p.RespondedAt != nil {
		s := p.RespondedAt.UTC().Format(time.RFC3339)
		respondedAt = &s
	}
	return MCPLiturgiPetugas{
		ID:                   p.ID.String(),
		ParokiID:             p.ParokiID.String(),
		EventID:              p.EventID.String(),
		Role:                 p.Role,
		SlotLabel:            p.SlotLabel,
		UserID:               uuidPtrStr(p.UserID),
		OrgID:                uuidPtrStr(p.OrgID),
		Name:                 p.Name,
		Urutan:               p.Urutan,
		Status:               p.Status,
		Note:                 p.Note,
		RespondedAt:          respondedAt,
		RespondedBy:          uuidPtrStr(p.RespondedBy),
		SubstituteUserID:     uuidPtrStr(p.SubstituteUserID),
		SubstituteOrgID:      uuidPtrStr(p.SubstituteOrgID),
		SubstituteName:       p.SubstituteName,
		CanConfirm:           p.CanConfirm,
		CanEdit:              p.CanEdit,
		CanRespond:           p.CanRespond,
		CanApproveSubstitute: p.CanApproveSubstitute,
		CreatedAt:            p.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:            p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMCPLiturgiLagu(l domain.LiturgiLaguResponse) MCPLiturgiLagu {
	return MCPLiturgiLagu{
		ID:        l.ID.String(),
		ParokiID:  l.ParokiID.String(),
		EventID:   l.EventID.String(),
		Urutan:    l.Urutan,
		Bagian:    l.Bagian,
		CantusID:  uuidPtrStr(l.CantusID),
		Sumber:    l.Sumber,
		Nomor:     l.Nomor,
		Judul:     l.Judul,
		NadaDasar: l.NadaDasar,
		Note:      l.Note,
		CreatedAt: l.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: l.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toMCPLiturgiSchedule(s domain.LiturgiScheduleResponse, loc *time.Location) MCPLiturgiSchedule {
	if loc == nil {
		loc = time.UTC
	}
	var deadline *string
	if s.RsvpDeadline != nil {
		d := s.RsvpDeadline.In(loc).Format(time.RFC3339)
		deadline = &d
	}
	petugas := make([]MCPLiturgiPetugas, 0, len(s.Petugas))
	for _, p := range s.Petugas {
		petugas = append(petugas, toMCPLiturgiPetugas(p))
	}
	lagu := make([]MCPLiturgiLagu, 0, len(s.Lagu))
	for _, l := range s.Lagu {
		lagu = append(lagu, toMCPLiturgiLagu(l))
	}
	return MCPLiturgiSchedule{
		ID:                s.ID.String(),
		ParokiID:          s.ParokiID.String(),
		CreatedBy:         s.CreatedBy.String(),
		Title:             s.Title,
		Description:       s.Description,
		Location:          s.Location,
		StartAt:           s.StartAt.In(loc).Format(time.RFC3339),
		EndAt:             s.EndAt.In(loc).Format(time.RFC3339),
		Timezone:          loc.String(),
		Visibility:        s.Visibility,
		RsvpEnabled:       s.RsvpEnabled,
		RsvpDeadline:      deadline,
		Status:            s.Status,
		RsvpOpen:          s.RsvpOpen,
		CanRsvp:           s.CanRsvp,
		CanManageSchedule: s.CanManageSchedule,
		ManageableRoles:   s.ManageableRoles,
		CanManageLagu:     s.CanManageLagu,
		Petugas:           petugas,
		Lagu:              lagu,
		CreatedAt:         s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:         s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func uuidPtrStr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
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
