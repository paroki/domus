package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/eventpetugas"
	"github.com/paroki/domus/ent/predicate"
)

// CreateEventInput holds fields for creating an Event.
type CreateEventInput struct {
	ParokiID     uuid.UUID
	OrgID        *uuid.UUID
	CreatedBy    uuid.UUID
	Title        string
	Jenis        *string
	Description  *string
	Location     *string
	StartAt      time.Time
	EndAt        time.Time
	Visibility   event.Visibility
	RsvpEnabled  bool
	RsvpDeadline *time.Time
}

// SearchEventsInput holds query filters and pagination for SearchEvents.
type SearchEventsInput struct {
	ParokiID       uuid.UUID
	MemberOrgIDs   []uuid.UUID
	OrgID          *uuid.UUID
	ParokiWideOnly bool
	Jenis          string
	Visibility     string
	Status         string
	MyOrgsOnly     bool
	AssignedToMe   bool
	AssignedUserID uuid.UUID
	AssignedOrgID  *uuid.UUID
	From           *time.Time
	To             *time.Time
	Q              string
	CursorAt       *time.Time
	CursorID       *uuid.UUID
	Limit          int
}

// SearchEvents returns paginated non-deleted events visible to the caller in parokiID.
func (r *Repository) SearchEvents(ctx context.Context, in SearchEventsInput) ([]*ent.Event, error) {
	query := r.client.Event.Query().
		Where(event.ParokiIDEQ(in.ParokiID), event.DeletedAtIsNil())

	if in.MyOrgsOnly {
		if len(in.MemberOrgIDs) == 0 {
			return []*ent.Event{}, nil
		}
		query = query.Where(event.OrgIDIn(in.MemberOrgIDs...))
	} else {
		var visPreds []predicate.Event
		visPreds = append(visPreds, event.VisibilityEQ(event.VisibilityPublic))
		if len(in.MemberOrgIDs) > 0 {
			visPreds = append(visPreds, event.OrgIDIn(in.MemberOrgIDs...))
		}
		query = query.Where(event.Or(visPreds...))
	}

	if in.OrgID != nil {
		query = query.Where(event.OrgIDEQ(*in.OrgID))
	}
	if in.ParokiWideOnly {
		query = query.Where(event.OrgIDIsNil())
	}
	if in.Jenis != "" {
		query = query.Where(event.JenisEQ(in.Jenis))
	}
	if in.AssignedToMe {
		petugasPreds := []predicate.EventPetugas{
			eventpetugas.UserIDEQ(in.AssignedUserID),
		}
		if len(in.MemberOrgIDs) > 0 {
			petugasPreds = append(petugasPreds, eventpetugas.OrgIDIn(in.MemberOrgIDs...))
		}
		query = query.Where(event.HasPetugasWith(eventpetugas.Or(petugasPreds...)))
	}
	if in.AssignedOrgID != nil {
		query = query.Where(event.HasPetugasWith(eventpetugas.OrgIDEQ(*in.AssignedOrgID)))
	}
	if in.Visibility != "" {
		query = query.Where(event.VisibilityEQ(event.Visibility(in.Visibility)))
	}
	if in.Status != "" {
		query = query.Where(event.StatusEQ(event.Status(in.Status)))
	}
	if in.From != nil {
		query = query.Where(event.StartAtGTE(in.From.UTC()))
	}
	if in.To != nil {
		query = query.Where(event.StartAtLTE(in.To.UTC()))
	}
	if in.Q != "" {
		query = query.Where(event.Or(
			event.TitleContainsFold(in.Q),
			event.DescriptionContainsFold(in.Q),
			event.LocationContainsFold(in.Q),
		))
	}
	if in.CursorAt != nil && in.CursorID != nil {
		query = query.Where(event.Or(
			event.CreatedAtLT(*in.CursorAt),
			event.And(event.CreatedAtEQ(*in.CursorAt), event.IDLT(*in.CursorID)),
		))
	}

	return query.
		Order(event.ByCreatedAt(sql.OrderDesc()), event.ByID(sql.OrderDesc())).
		Limit(in.Limit).
		All(ctx)
}

// FindEventInParoki returns a non-deleted Event belonging to parokiID.
func (r *Repository) FindEventInParoki(ctx context.Context, parokiID, eventID uuid.UUID) (*ent.Event, error) {
	return r.client.Event.Query().
		Where(event.IDEQ(eventID), event.ParokiIDEQ(parokiID), event.DeletedAtIsNil()).
		Only(ctx)
}

// CreateEvent inserts a new Event row.
func (r *Repository) CreateEvent(ctx context.Context, in CreateEventInput) (*ent.Event, error) {
	q := r.client.Event.Create().
		SetParokiID(in.ParokiID).
		SetCreatedBy(in.CreatedBy).
		SetTitle(in.Title).
		SetStartAt(in.StartAt.UTC()).
		SetEndAt(in.EndAt.UTC()).
		SetVisibility(in.Visibility).
		SetRsvpEnabled(in.RsvpEnabled).
		SetStatus(event.StatusScheduled)

	if in.OrgID != nil {
		q.SetOrgID(*in.OrgID)
	}
	if in.Jenis != nil && *in.Jenis != "" {
		q.SetJenis(*in.Jenis)
	}

	if in.Description != nil && *in.Description != "" {
		q.SetDescription(*in.Description)
	}
	if in.Location != nil && *in.Location != "" {
		q.SetLocation(*in.Location)
	}
	if in.RsvpDeadline != nil {
		q.SetRsvpDeadline(in.RsvpDeadline.UTC())
	}
	return q.Save(ctx)
}

// UpdateEventInput holds optional fields for updating an Event.
type UpdateEventInput struct {
	Title             *string
	Jenis             *string
	ClearJenis        bool
	Description       *string
	Location          *string
	StartAt           *time.Time
	EndAt             *time.Time
	Visibility        *event.Visibility
	RsvpEnabled       *bool
	RsvpDeadline      *time.Time
	ClearRsvpDeadline bool
	Status            *event.Status
}

// UpdateEvent updates an existing Event row.
func (r *Repository) UpdateEvent(ctx context.Context, eventID uuid.UUID, in UpdateEventInput) (*ent.Event, error) {
	q := r.client.Event.UpdateOneID(eventID)
	if in.Title != nil {
		q.SetTitle(*in.Title)
	}
	if in.ClearJenis {
		q.ClearJenis()
	} else if in.Jenis != nil {
		q.SetJenis(*in.Jenis)
	}
	if in.Description != nil {
		if *in.Description == "" {
			q.ClearDescription()
		} else {
			q.SetDescription(*in.Description)
		}
	}
	if in.Location != nil {
		if *in.Location == "" {
			q.ClearLocation()
		} else {
			q.SetLocation(*in.Location)
		}
	}
	if in.StartAt != nil {
		q.SetStartAt(in.StartAt.UTC())
	}
	if in.EndAt != nil {
		q.SetEndAt(in.EndAt.UTC())
	}
	if in.Visibility != nil {
		q.SetVisibility(*in.Visibility)
	}
	if in.RsvpEnabled != nil {
		q.SetRsvpEnabled(*in.RsvpEnabled)
	}
	if in.ClearRsvpDeadline {
		q.ClearRsvpDeadline()
	} else if in.RsvpDeadline != nil {
		q.SetRsvpDeadline(in.RsvpDeadline.UTC())
	}
	if in.Status != nil {
		q.SetStatus(*in.Status)
	}
	return q.Save(ctx)
}

// SoftDeleteEvent sets deleted_at on an Event row.
func (r *Repository) SoftDeleteEvent(ctx context.Context, eventID uuid.UUID, at time.Time) error {
	return r.client.Event.UpdateOneID(eventID).
		SetDeletedAt(at.UTC()).
		Exec(ctx)
}
