package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/liturgilagu"
)

// CreateLaguInput holds fields for creating a liturgi_lagu row.
type CreateLaguInput struct {
	ParokiID  uuid.UUID
	EventID   uuid.UUID
	Urutan    int
	Bagian    string
	CantusID  *uuid.UUID
	Sumber    *string
	Nomor     *string
	Judul     string
	NadaDasar *string
	Note      *string
}

// UpdateLaguInput holds fields for updating a liturgi_lagu row.
type UpdateLaguInput struct {
	Urutan        *int
	Bagian        *string
	CantusID      *uuid.UUID
	ClearCantusID bool
	Sumber        *string
	ClearSumber   bool
	Nomor         *string
	ClearNomor    bool
	Judul         *string
	NadaDasar     *string
	ClearNada     bool
	Note          *string
	ClearNote     bool
}

// FindLaguByID returns the LiturgiLagu row matching (parokiID, eventID, laguID).
func (r *Repository) FindLaguByID(ctx context.Context, parokiID, eventID, laguID uuid.UUID) (*ent.LiturgiLagu, error) {
	return r.client.LiturgiLagu.Query().
		Where(
			liturgilagu.IDEQ(laguID),
			liturgilagu.ParokiIDEQ(parokiID),
			liturgilagu.EventIDEQ(eventID),
		).
		Only(ctx)
}

// CreateLagu inserts a new LiturgiLagu row.
func (r *Repository) CreateLagu(ctx context.Context, in CreateLaguInput) (*ent.LiturgiLagu, error) {
	q := r.client.LiturgiLagu.Create().
		SetParokiID(in.ParokiID).
		SetEventID(in.EventID).
		SetUrutan(in.Urutan).
		SetBagian(in.Bagian).
		SetJudul(in.Judul)
	if in.CantusID != nil {
		q.SetCantusID(*in.CantusID)
	}
	if in.Sumber != nil {
		q.SetSumber(*in.Sumber)
	}
	if in.Nomor != nil {
		q.SetNomor(*in.Nomor)
	}
	if in.NadaDasar != nil {
		q.SetNadaDasar(*in.NadaDasar)
	}
	if in.Note != nil {
		q.SetNote(*in.Note)
	}
	return q.Save(ctx)
}

// UpdateLagu updates an existing LiturgiLagu row.
func (r *Repository) UpdateLagu(ctx context.Context, id uuid.UUID, in UpdateLaguInput) (*ent.LiturgiLagu, error) {
	upd := r.client.LiturgiLagu.UpdateOneID(id)
	if in.Urutan != nil {
		upd.SetUrutan(*in.Urutan)
	}
	if in.Bagian != nil {
		upd.SetBagian(*in.Bagian)
	}
	if in.ClearCantusID {
		upd.ClearCantusID()
	} else if in.CantusID != nil {
		upd.SetCantusID(*in.CantusID)
	}
	if in.ClearSumber {
		upd.ClearSumber()
	} else if in.Sumber != nil {
		upd.SetSumber(*in.Sumber)
	}
	if in.ClearNomor {
		upd.ClearNomor()
	} else if in.Nomor != nil {
		upd.SetNomor(*in.Nomor)
	}
	if in.Judul != nil {
		upd.SetJudul(*in.Judul)
	}
	if in.ClearNada {
		upd.ClearNadaDasar()
	} else if in.NadaDasar != nil {
		upd.SetNadaDasar(*in.NadaDasar)
	}
	if in.ClearNote {
		upd.ClearNote()
	} else if in.Note != nil {
		upd.SetNote(*in.Note)
	}
	return upd.Save(ctx)
}

// DeleteLagu hard-deletes a LiturgiLagu row by ID.
func (r *Repository) DeleteLagu(ctx context.Context, id uuid.UUID) error {
	return r.client.LiturgiLagu.DeleteOneID(id).Exec(ctx)
}

// IsCantusInUse reports whether cantusID is referenced by any liturgi_lagu row belonging to a non-deleted Event.
func (r *Repository) IsCantusInUse(ctx context.Context, cantusID uuid.UUID) (bool, error) {
	rows, err := r.client.LiturgiLagu.Query().
		Where(liturgilagu.CantusIDEQ(cantusID)).
		Select(liturgilagu.FieldEventID).
		All(ctx)
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, nil
	}
	eventIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		eventIDs = append(eventIDs, row.EventID)
	}
	return r.client.Event.Query().
		Where(event.IDIn(eventIDs...), event.DeletedAtIsNil()).
		Exist(ctx)
}
