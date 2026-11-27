package domain

import (
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
)

// Feature-owned Casbin role and resource permissions.
var (
	RoleAdmin      = authz.Role("liturgi", "admin")
	PermSchedules  = authz.Permission("liturgi", "schedules")
	PermMinistries = authz.Permission("liturgi", "ministries")
	PermPetugas    = authz.Permission("liturgi", "petugas")
	PermLagu       = authz.Permission("liturgi", "lagu")
)

// MinistryRoles is the fixed canonical order of the 6 liturgical ministry roles.
var MinistryRoles = []string{
	"lektor",
	"mazmur",
	"pembagi-komuni",
	"misdinar",
	"organis",
	"kor",
}

// IsValidMinistryRole reports whether slug is one of the 6 canonical ministry roles.
func IsValidMinistryRole(slug string) bool {
	return slices.Contains(MinistryRoles, slug)
}

// LiturgiMinistryResponse is the public API representation of a parish liturgical ministry configuration.
type LiturgiMinistryResponse struct {
	ID           *uuid.UUID `json:"id,omitempty"`
	ParokiID     uuid.UUID  `json:"parokiId"`
	RoleSlug     string     `json:"roleSlug"`
	OrgID        *uuid.UUID `json:"orgId,omitempty"`
	OrgName      *string    `json:"orgName,omitempty"`
	LeaderUserID *uuid.UUID `json:"leaderUserId,omitempty"`
	LeaderName   *string    `json:"leaderName,omitempty"`
	LeaderEmail  *string    `json:"leaderEmail,omitempty"`
	CanManage    bool       `json:"canManage"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

// UpdateLiturgiMinistryRequest updates the assigned organization and/or leader user for a ministry role.
type UpdateLiturgiMinistryRequest struct {
	OrgID             *uuid.UUID `json:"orgId,omitempty"`
	ClearOrgID        *bool      `json:"clearOrgId,omitempty"`
	LeaderUserID      *uuid.UUID `json:"leaderUserId,omitempty"`
	ClearLeaderUserID *bool      `json:"clearLeaderUserId,omitempty"`
}

// LiturgiPetugasResponse is the public API representation of a liturgical officer assignment.
type LiturgiPetugasResponse struct {
	ID               uuid.UUID  `json:"id"`
	ParokiID         uuid.UUID  `json:"parokiId"`
	EventID          uuid.UUID  `json:"eventId"`
	Role             string     `json:"role"`
	SlotLabel        *string    `json:"slotLabel,omitempty"`
	UserID           *uuid.UUID `json:"userId,omitempty"`
	OrgID            *uuid.UUID `json:"orgId,omitempty"`
	Name             string     `json:"name"`
	Urutan           int        `json:"urutan"`
	Status           string     `json:"status"`
	Note             *string    `json:"note,omitempty"`
	RespondedAt      *time.Time `json:"respondedAt,omitempty"`
	RespondedBy      *uuid.UUID `json:"respondedBy,omitempty"`
	SubstituteUserID *uuid.UUID `json:"substituteUserId,omitempty"`
	SubstituteOrgID  *uuid.UUID `json:"substituteOrgId,omitempty"`
	SubstituteName   *string    `json:"substituteName,omitempty"`
	CanConfirm       bool       `json:"canConfirm"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// LiturgiLaguResponse is the public API representation of a liturgical song entry.
type LiturgiLaguResponse struct {
	ID        uuid.UUID  `json:"id"`
	ParokiID  uuid.UUID  `json:"parokiId"`
	EventID   uuid.UUID  `json:"eventId"`
	Urutan    int        `json:"urutan"`
	Bagian    string     `json:"bagian"`
	CantusID  *uuid.UUID `json:"cantusId,omitempty"`
	Sumber    *string    `json:"sumber,omitempty"`
	Nomor     *string    `json:"nomor,omitempty"`
	Judul     string     `json:"judul"`
	NadaDasar *string    `json:"nadaDasar,omitempty"`
	Note      *string    `json:"note,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// LiturgiRSVPResponse represents the caller's RSVP for a Mass schedule.
type LiturgiRSVPResponse struct {
	ID        uuid.UUID `json:"id"`
	ParokiID  uuid.UUID `json:"parokiId"`
	EventID   uuid.UUID `json:"eventId"`
	UserID    uuid.UUID `json:"userId"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	OrgRole   string    `json:"orgRole,omitempty"`
	Status    string    `json:"status"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// LiturgiRSVPSummaryResponse summarizes RSVP counts for a Mass schedule.
type LiturgiRSVPSummaryResponse struct {
	EventID            uuid.UUID `json:"eventId"`
	Going              int       `json:"going"`
	Maybe              int       `json:"maybe"`
	NotGoing           int       `json:"notGoing"`
	TotalResponded     int       `json:"totalResponded"`
	TotalActiveMembers int       `json:"totalActiveMembers"`
	Unresponded        int       `json:"unresponded"`
}

// LiturgiScheduleResponse is the public API representation of a parish Mass schedule.
type LiturgiScheduleResponse struct {
	ID                uuid.UUID                   `json:"id"`
	ParokiID          uuid.UUID                   `json:"parokiId"`
	CreatedBy         uuid.UUID                   `json:"createdBy"`
	Title             string                      `json:"title"`
	Description       *string                     `json:"description,omitempty"`
	Location          *string                     `json:"location,omitempty"`
	StartAt           time.Time                   `json:"startAt"`
	EndAt             time.Time                   `json:"endAt"`
	Visibility        string                      `json:"visibility"`
	RsvpEnabled       bool                        `json:"rsvpEnabled"`
	RsvpDeadline      *time.Time                  `json:"rsvpDeadline,omitempty"`
	Status            string                      `json:"status"`
	RsvpOpen          bool                        `json:"rsvpOpen"`
	CanRsvp           bool                        `json:"canRsvp"`
	MyRsvp            *LiturgiRSVPResponse        `json:"myRsvp,omitempty"`
	RsvpSummary       *LiturgiRSVPSummaryResponse `json:"rsvpSummary,omitempty"`
	CanManageSchedule bool                        `json:"canManageSchedule"`
	ManageableRoles   []string                    `json:"manageableRoles"`
	CanManageLagu     bool                        `json:"canManageLagu"`
	Petugas           []LiturgiPetugasResponse    `json:"petugas"`
	Lagu              []LiturgiLaguResponse       `json:"lagu"`
	CreatedAt         time.Time                   `json:"createdAt"`
	UpdatedAt         time.Time                   `json:"updatedAt"`
}

// LiturgiScheduleSearchFilter filters parish Mass schedule search results.
type LiturgiScheduleSearchFilter struct {
	From          *time.Time `json:"from,omitempty"`
	To            *time.Time `json:"to,omitempty"`
	Status        string     `json:"status,omitempty" validate:"omitempty,oneof=scheduled cancelled"`
	Role          string     `json:"role,omitempty"`
	AssignedToMe  *bool      `json:"assignedToMe,omitempty"`
	AssignedOrgID *uuid.UUID `json:"assignedOrgId,omitempty"`
	HasDeclined   *bool      `json:"hasDeclined,omitempty"`
}

// SearchLiturgiSchedulesRequest searches parish Mass schedules in the active paroki.
type SearchLiturgiSchedulesRequest struct {
	Q      string                       `json:"q,omitempty"`
	Filter *LiturgiScheduleSearchFilter `json:"filter,omitempty"`
	Cursor string                       `json:"cursor,omitempty"`
	Limit  int                          `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateLiturgiScheduleRequest creates a new parish-level Mass schedule.
type CreateLiturgiScheduleRequest struct {
	Title        string     `json:"title" validate:"required,min=1"`
	Description  *string    `json:"description,omitempty"`
	Location     *string    `json:"location,omitempty"`
	StartAt      time.Time  `json:"startAt" validate:"required"`
	EndAt        time.Time  `json:"endAt" validate:"required"`
	RsvpEnabled  *bool      `json:"rsvpEnabled,omitempty"`
	RsvpDeadline *time.Time `json:"rsvpDeadline,omitempty"`
}

// UpdateLiturgiScheduleRequest updates an existing parish-level Mass schedule.
type UpdateLiturgiScheduleRequest struct {
	Title             *string    `json:"title,omitempty" validate:"omitempty,min=1"`
	Description       *string    `json:"description,omitempty"`
	Location          *string    `json:"location,omitempty"`
	StartAt           *time.Time `json:"startAt,omitempty"`
	EndAt             *time.Time `json:"endAt,omitempty"`
	RsvpEnabled       *bool      `json:"rsvpEnabled,omitempty"`
	RsvpDeadline      *time.Time `json:"rsvpDeadline,omitempty"`
	ClearRsvpDeadline *bool      `json:"clearRsvpDeadline,omitempty"`
	Status            *string    `json:"status,omitempty" validate:"omitempty,oneof=scheduled cancelled"`
}
