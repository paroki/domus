package repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/eventpetugas"
	"github.com/paroki/domus/ent/user"
)

// CreateEventPetugasInput holds fields for creating an EventPetugas row.
type CreateEventPetugasInput struct {
	ParokiID uuid.UUID
	EventID  uuid.UUID
	Role     string
	UserID   *uuid.UUID
	OrgID    *uuid.UUID
	Name     string
	Urutan   int
	Status   eventpetugas.Status
	Note     *string
}

// UpdateEventPetugasInput holds optional fields for updating an EventPetugas row.
type UpdateEventPetugasInput struct {
	Role        *string
	UserID      *uuid.UUID
	ClearUserID bool
	OrgID       *uuid.UUID
	ClearOrgID  bool
	Name        *string
	Urutan      *int
	Status      *eventpetugas.Status
	Note        *string
	ClearNote   bool
}

// ListPetugasForEvent returns all EventPetugas rows for eventID ordered by urutan ASC, created_at ASC.
func (r *Repository) ListPetugasForEvent(ctx context.Context, eventID uuid.UUID) ([]*ent.EventPetugas, error) {
	return r.client.EventPetugas.Query().
		Where(eventpetugas.EventIDEQ(eventID)).
		Order(eventpetugas.ByUrutan(sql.OrderAsc()), eventpetugas.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
}

// FindPetugasInEvent returns an EventPetugas row belonging to parokiID and eventID.
func (r *Repository) FindPetugasInEvent(ctx context.Context, parokiID, eventID, petugasID uuid.UUID) (*ent.EventPetugas, error) {
	return r.client.EventPetugas.Query().
		Where(
			eventpetugas.IDEQ(petugasID),
			eventpetugas.EventIDEQ(eventID),
			eventpetugas.ParokiIDEQ(parokiID),
		).
		Only(ctx)
}

// CreatePetugas inserts a new EventPetugas row.
func (r *Repository) CreatePetugas(ctx context.Context, in CreateEventPetugasInput) (*ent.EventPetugas, error) {
	q := r.client.EventPetugas.Create().
		SetParokiID(in.ParokiID).
		SetEventID(in.EventID).
		SetRole(in.Role).
		SetName(in.Name).
		SetUrutan(in.Urutan).
		SetStatus(in.Status)

	if in.UserID != nil {
		q.SetUserID(*in.UserID)
	}
	if in.OrgID != nil {
		q.SetOrgID(*in.OrgID)
	}
	if in.Note != nil && *in.Note != "" {
		q.SetNote(*in.Note)
	}
	return q.Save(ctx)
}

// UpdatePetugas updates an existing EventPetugas row.
func (r *Repository) UpdatePetugas(ctx context.Context, petugasID uuid.UUID, in UpdateEventPetugasInput) (*ent.EventPetugas, error) {
	q := r.client.EventPetugas.UpdateOneID(petugasID)
	if in.Role != nil {
		q.SetRole(*in.Role)
	}
	if in.ClearUserID {
		q.ClearUserID()
	} else if in.UserID != nil {
		q.SetUserID(*in.UserID)
	}
	if in.ClearOrgID {
		q.ClearOrgID()
	} else if in.OrgID != nil {
		q.SetOrgID(*in.OrgID)
	}
	if in.Name != nil {
		q.SetName(*in.Name)
	}
	if in.Urutan != nil {
		q.SetUrutan(*in.Urutan)
	}
	if in.Status != nil {
		q.SetStatus(*in.Status)
	}
	if in.ClearNote {
		q.ClearNote()
	} else if in.Note != nil {
		q.SetNote(*in.Note)
	}
	return q.Save(ctx)
}

// DeletePetugas hard-deletes an EventPetugas row by ID.
func (r *Repository) DeletePetugas(ctx context.Context, petugasID uuid.UUID) error {
	return r.client.EventPetugas.DeleteOneID(petugasID).Exec(ctx)
}

// FindActiveUser returns a non-deleted User by ID.
func (r *Repository) FindActiveUser(ctx context.Context, userID uuid.UUID) (*ent.User, error) {
	return r.client.User.Query().
		Where(user.IDEQ(userID), user.DeletedAtIsNil()).
		Only(ctx)
}
