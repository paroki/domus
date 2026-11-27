package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/liturgipetugas"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/ent/user"
)

// CreatePetugasInput holds fields for creating a liturgi_petugas row.
type CreatePetugasInput struct {
	ParokiID    uuid.UUID
	EventID     uuid.UUID
	Role        liturgipetugas.Role
	SlotLabel   *string
	UserID      *uuid.UUID
	OrgID       *uuid.UUID
	Name        string
	Urutan      int
	Status      liturgipetugas.Status
	Note        *string
	RespondedAt *time.Time
	RespondedBy *uuid.UUID
}

// UpdatePetugasInput holds fields for updating a liturgi_petugas row.
type UpdatePetugasInput struct {
	Role                  *liturgipetugas.Role
	SlotLabel             *string
	ClearSlotLabel        bool
	UserID                *uuid.UUID
	ClearUserID           bool
	OrgID                 *uuid.UUID
	ClearOrgID            bool
	Name                  *string
	Urutan                *int
	Status                *liturgipetugas.Status
	Note                  *string
	ClearNote             bool
	RespondedAt           *time.Time
	ClearResponded        bool
	RespondedBy           *uuid.UUID
	SubstituteUserID      *uuid.UUID
	ClearSubstituteUserID bool
	SubstituteOrgID       *uuid.UUID
	ClearSubstituteOrgID  bool
	SubstituteName        *string
	ClearSubstituteName   bool
}

// FindPetugasByID returns the LiturgiPetugas row matching (parokiID, eventID, petugasID).
func (r *Repository) FindPetugasByID(ctx context.Context, parokiID, eventID, petugasID uuid.UUID) (*ent.LiturgiPetugas, error) {
	return r.client.LiturgiPetugas.Query().
		Where(
			liturgipetugas.IDEQ(petugasID),
			liturgipetugas.ParokiIDEQ(parokiID),
			liturgipetugas.EventIDEQ(eventID),
		).
		Only(ctx)
}

// CreatePetugas inserts a new LiturgiPetugas row.
func (r *Repository) CreatePetugas(ctx context.Context, in CreatePetugasInput) (*ent.LiturgiPetugas, error) {
	q := r.client.LiturgiPetugas.Create().
		SetParokiID(in.ParokiID).
		SetEventID(in.EventID).
		SetRole(in.Role).
		SetName(in.Name).
		SetUrutan(in.Urutan).
		SetStatus(in.Status)
	if in.SlotLabel != nil {
		q.SetSlotLabel(*in.SlotLabel)
	}
	if in.UserID != nil {
		q.SetUserID(*in.UserID)
	}
	if in.OrgID != nil {
		q.SetOrgID(*in.OrgID)
	}
	if in.Note != nil {
		q.SetNote(*in.Note)
	}
	if in.RespondedAt != nil {
		q.SetRespondedAt(*in.RespondedAt)
	}
	if in.RespondedBy != nil {
		q.SetRespondedBy(*in.RespondedBy)
	}
	return q.Save(ctx)
}

// UpdatePetugas updates an existing LiturgiPetugas row.
func (r *Repository) UpdatePetugas(ctx context.Context, id uuid.UUID, in UpdatePetugasInput) (*ent.LiturgiPetugas, error) {
	upd := r.client.LiturgiPetugas.UpdateOneID(id)
	if in.Role != nil {
		upd.SetRole(*in.Role)
	}
	if in.ClearSlotLabel {
		upd.ClearSlotLabel()
	} else if in.SlotLabel != nil {
		upd.SetSlotLabel(*in.SlotLabel)
	}
	if in.ClearUserID {
		upd.ClearUserID()
	} else if in.UserID != nil {
		upd.SetUserID(*in.UserID)
	}
	if in.ClearOrgID {
		upd.ClearOrgID()
	} else if in.OrgID != nil {
		upd.SetOrgID(*in.OrgID)
	}
	if in.Name != nil {
		upd.SetName(*in.Name)
	}
	if in.Urutan != nil {
		upd.SetUrutan(*in.Urutan)
	}
	if in.Status != nil {
		upd.SetStatus(*in.Status)
	}
	if in.ClearNote {
		upd.ClearNote()
	} else if in.Note != nil {
		upd.SetNote(*in.Note)
	}
	if in.ClearResponded {
		upd.ClearRespondedAt().ClearRespondedBy()
	} else {
		if in.RespondedAt != nil {
			upd.SetRespondedAt(*in.RespondedAt)
		}
		if in.RespondedBy != nil {
			upd.SetRespondedBy(*in.RespondedBy)
		}
	}
	if in.ClearSubstituteUserID {
		upd.ClearSubstituteUserID()
	} else if in.SubstituteUserID != nil {
		upd.SetSubstituteUserID(*in.SubstituteUserID)
	}
	if in.ClearSubstituteOrgID {
		upd.ClearSubstituteOrgID()
	} else if in.SubstituteOrgID != nil {
		upd.SetSubstituteOrgID(*in.SubstituteOrgID)
	}
	if in.ClearSubstituteName {
		upd.ClearSubstituteName()
	} else if in.SubstituteName != nil {
		upd.SetSubstituteName(*in.SubstituteName)
	}
	return upd.Save(ctx)
}

// DeletePetugas hard-deletes a LiturgiPetugas row by ID.
func (r *Repository) DeletePetugas(ctx context.Context, id uuid.UUID) error {
	return r.client.LiturgiPetugas.DeleteOneID(id).Exec(ctx)
}

// GetActiveUserNameInParoki returns the display name of userID if non-deleted and a member of parokiID.
func (r *Repository) GetActiveUserNameInParoki(ctx context.Context, parokiID, userID uuid.UUID) (string, bool, error) {
	isMember, err := r.client.ParokiMember.Query().
		Where(parokimember.ParokiIDEQ(parokiID), parokimember.UserIDEQ(userID)).
		Exist(ctx)
	if err != nil || !isMember {
		return "", false, err
	}
	u, err := r.client.User.Query().
		Where(user.IDEQ(userID), user.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return u.Name, true, nil
}

// ListActiveOrgMemberUserIDs returns active member user IDs for orgID.
func (r *Repository) ListActiveOrgMemberUserIDs(ctx context.Context, orgID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.client.OrgMember.Query().
		Where(orgmember.OrgIDEQ(orgID), orgmember.StatusEQ(orgmember.StatusActive)).
		Select(orgmember.FieldUserID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	return ids, nil
}

// ListActivePetugasForEvent returns non-declined LiturgiPetugas rows for eventID.
func (r *Repository) ListActivePetugasForEvent(ctx context.Context, eventID uuid.UUID) ([]*ent.LiturgiPetugas, error) {
	return r.client.LiturgiPetugas.Query().
		Where(
			liturgipetugas.EventIDEQ(eventID),
			liturgipetugas.StatusNEQ(liturgipetugas.StatusDeclined),
		).
		All(ctx)
}

// ListScheduledMassEventsBetween returns non-deleted scheduled parish Mass events with start_at in [from, to].
func (r *Repository) ListScheduledMassEventsBetween(ctx context.Context, from, to time.Time) ([]*ent.Event, error) {
	return r.client.Event.Query().
		Where(
			event.DeletedAtIsNil(),
			event.StatusEQ(event.StatusScheduled),
			event.OrgIDIsNil(),
			event.JenisEQ("misa"),
			event.StartAtGTE(from),
			event.StartAtLTE(to),
		).
		All(ctx)
}
