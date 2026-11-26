package repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/eventlagu"
)

// CreateEventLaguInput holds fields for creating an EventLagu row.
type CreateEventLaguInput struct {
	ParokiID  uuid.UUID
	EventID   uuid.UUID
	Urutan    int
	Bagian    string
	Sumber    *string
	Nomor     *string
	Judul     string
	NadaDasar *string
	Note      *string
}

// UpdateEventLaguInput holds optional fields for updating an EventLagu row.
type UpdateEventLaguInput struct {
	Urutan         *int
	Bagian         *string
	Sumber         *string
	ClearSumber    bool
	Nomor          *string
	ClearNomor     bool
	Judul          *string
	NadaDasar      *string
	ClearNadaDasar bool
	Note           *string
	ClearNote      bool
}

// ListLaguForEvent returns all EventLagu rows for eventID ordered by urutan ASC, created_at ASC.
func (r *Repository) ListLaguForEvent(ctx context.Context, eventID uuid.UUID) ([]*ent.EventLagu, error) {
	return r.client.EventLagu.Query().
		Where(eventlagu.EventIDEQ(eventID)).
		Order(eventlagu.ByUrutan(sql.OrderAsc()), eventlagu.ByCreatedAt(sql.OrderAsc())).
		All(ctx)
}

// FindLaguInEvent returns an EventLagu row belonging to parokiID and eventID.
func (r *Repository) FindLaguInEvent(ctx context.Context, parokiID, eventID, laguID uuid.UUID) (*ent.EventLagu, error) {
	return r.client.EventLagu.Query().
		Where(
			eventlagu.IDEQ(laguID),
			eventlagu.EventIDEQ(eventID),
			eventlagu.ParokiIDEQ(parokiID),
		).
		Only(ctx)
}

// CreateLagu inserts a new EventLagu row.
func (r *Repository) CreateLagu(ctx context.Context, in CreateEventLaguInput) (*ent.EventLagu, error) {
	q := r.client.EventLagu.Create().
		SetParokiID(in.ParokiID).
		SetEventID(in.EventID).
		SetUrutan(in.Urutan).
		SetBagian(in.Bagian).
		SetJudul(in.Judul)

	if in.Sumber != nil && *in.Sumber != "" {
		q.SetSumber(*in.Sumber)
	}
	if in.Nomor != nil && *in.Nomor != "" {
		q.SetNomor(*in.Nomor)
	}
	if in.NadaDasar != nil && *in.NadaDasar != "" {
		q.SetNadaDasar(*in.NadaDasar)
	}
	if in.Note != nil && *in.Note != "" {
		q.SetNote(*in.Note)
	}
	return q.Save(ctx)
}

// UpdateLagu updates an existing EventLagu row.
func (r *Repository) UpdateLagu(ctx context.Context, laguID uuid.UUID, in UpdateEventLaguInput) (*ent.EventLagu, error) {
	q := r.client.EventLagu.UpdateOneID(laguID)
	if in.Urutan != nil {
		q.SetUrutan(*in.Urutan)
	}
	if in.Bagian != nil {
		q.SetBagian(*in.Bagian)
	}
	if in.ClearSumber {
		q.ClearSumber()
	} else if in.Sumber != nil {
		q.SetSumber(*in.Sumber)
	}
	if in.ClearNomor {
		q.ClearNomor()
	} else if in.Nomor != nil {
		q.SetNomor(*in.Nomor)
	}
	if in.Judul != nil {
		q.SetJudul(*in.Judul)
	}
	if in.ClearNadaDasar {
		q.ClearNadaDasar()
	} else if in.NadaDasar != nil {
		q.SetNadaDasar(*in.NadaDasar)
	}
	if in.ClearNote {
		q.ClearNote()
	} else if in.Note != nil {
		q.SetNote(*in.Note)
	}
	return q.Save(ctx)
}

// DeleteLagu hard-deletes an EventLagu row by ID.
func (r *Repository) DeleteLagu(ctx context.Context, laguID uuid.UUID) error {
	return r.client.EventLagu.DeleteOneID(laguID).Exec(ctx)
}
