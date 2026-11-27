package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventResponse is the public API representation of an Event.
type EventResponse struct {
	ID           uuid.UUID                 `json:"id"`
	ParokiID     uuid.UUID                 `json:"parokiId"`
	OrgID        *uuid.UUID                `json:"orgId,omitempty"`
	OrgCode      string                    `json:"orgCode,omitempty"`
	OrgName      string                    `json:"orgName,omitempty"`
	OrgKind      string                    `json:"orgKind,omitempty"`
	CreatedBy    uuid.UUID                 `json:"createdBy"`
	Title        string                    `json:"title"`
	Jenis        *string                   `json:"jenis,omitempty"`
	Description  *string                   `json:"description,omitempty"`
	Location     *string                   `json:"location,omitempty"`
	StartAt      time.Time                 `json:"startAt"`
	EndAt        time.Time                 `json:"endAt"`
	Visibility   string                    `json:"visibility"`
	RsvpEnabled  bool                      `json:"rsvpEnabled"`
	RsvpDeadline *time.Time                `json:"rsvpDeadline,omitempty"`
	Status       string                    `json:"status"`
	RsvpOpen     bool                      `json:"rsvpOpen"`
	CanManage    bool                      `json:"canManage"`
	CanRsvp      bool                      `json:"canRsvp"`
	MyOrgRole    *string                   `json:"myOrgRole,omitempty"`
	MyRsvp       *EventRSVPResponse        `json:"myRsvp,omitempty"`
	RsvpSummary  *EventRSVPSummaryResponse `json:"rsvpSummary,omitempty"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
}

// EventSearchFilter filters event search results.
type EventSearchFilter struct {
	OrgID          *uuid.UUID  `json:"orgId,omitempty"`
	ParokiWideOnly *bool       `json:"parokiWideOnly,omitempty"`
	Jenis          string      `json:"jenis,omitempty"`
	Visibility     string      `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
	Status         string      `json:"status,omitempty" validate:"omitempty,oneof=scheduled cancelled"`
	MyOrgsOnly     *bool       `json:"myOrgsOnly,omitempty"`
	From           *time.Time  `json:"from,omitempty"`
	To             *time.Time  `json:"to,omitempty"`
	EventIDs       []uuid.UUID `json:"-"`
}

// SearchEventsRequest searches visible events within the active paroki.
type SearchEventsRequest struct {
	Q      string             `json:"q,omitempty"`
	Filter *EventSearchFilter `json:"filter,omitempty"`
	Cursor string             `json:"cursor,omitempty"`
	Limit  int                `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateEventRequest creates a new Event owned by an organization or the active paroki.
type CreateEventRequest struct {
	OrgID        *uuid.UUID `json:"orgId,omitempty"`
	Title        string     `json:"title" validate:"required,min=1"`
	Jenis        *string    `json:"jenis,omitempty"`
	Description  *string    `json:"description,omitempty"`
	Location     *string    `json:"location,omitempty"`
	StartAt      time.Time  `json:"startAt" validate:"required"`
	EndAt        time.Time  `json:"endAt" validate:"required"`
	Visibility   string     `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
	RsvpEnabled  *bool      `json:"rsvpEnabled,omitempty"`
	RsvpDeadline *time.Time `json:"rsvpDeadline,omitempty"`
}

// UpdateEventRequest updates an existing Event.
type UpdateEventRequest struct {
	Title             *string    `json:"title,omitempty" validate:"omitempty,min=1"`
	Jenis             *string    `json:"jenis,omitempty"`
	Description       *string    `json:"description,omitempty"`
	Location          *string    `json:"location,omitempty"`
	StartAt           *time.Time `json:"startAt,omitempty"`
	EndAt             *time.Time `json:"endAt,omitempty"`
	Visibility        *string    `json:"visibility,omitempty" validate:"omitempty,oneof=public private"`
	RsvpEnabled       *bool      `json:"rsvpEnabled,omitempty"`
	RsvpDeadline      *time.Time `json:"rsvpDeadline,omitempty"`
	ClearRsvpDeadline *bool      `json:"clearRsvpDeadline,omitempty"`
	Status            *string    `json:"status,omitempty" validate:"omitempty,oneof=scheduled cancelled"`
}

// EventRSVPResponse is the public API representation of an EventRSVP.
type EventRSVPResponse struct {
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

// UpsertEventRSVPRequest creates or updates the caller's RSVP for an Event.
type UpsertEventRSVPRequest struct {
	Status string  `json:"status" validate:"required,oneof=going maybe not_going"`
	Note   *string `json:"note,omitempty"`
}

// EventRSVPSummaryResponse summarizes RSVP counts for an Event.
type EventRSVPSummaryResponse struct {
	EventID            uuid.UUID `json:"eventId"`
	Going              int       `json:"going"`
	Maybe              int       `json:"maybe"`
	NotGoing           int       `json:"notGoing"`
	TotalResponded     int       `json:"totalResponded"`
	TotalActiveMembers int       `json:"totalActiveMembers"`
	Unresponded        int       `json:"unresponded"`
}

// EventRSVPSearchFilter filters event RSVP roster search results.
type EventRSVPSearchFilter struct {
	Status string `json:"status,omitempty" validate:"omitempty,oneof=going maybe not_going"`
}

// SearchEventRSVPsRequest searches RSVPs for an Event.
type SearchEventRSVPsRequest struct {
	Q      string                 `json:"q,omitempty"`
	Filter *EventRSVPSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}
