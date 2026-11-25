package repo

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentmemory"
	"github.com/paroki/domus/ent/agentsession"
	"github.com/paroki/domus/ent/event"
	"github.com/paroki/domus/ent/lingkungan"
	"github.com/paroki/domus/ent/lookup"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/paroki"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/wilayah"
)

// ParokiInput holds fields for creating or reactivating a Paroki.
type ParokiInput struct {
	KeuskupanID uuid.UUID
	Kode        string
	Nama        string
	Pelindung   *string
	Pastor      *string
	Alamat      *string
	Telepon     *string
	Email       *string
	Website     *string
	LogoURL     *string
	KopSurat    *string
	Active      bool
}

// SearchParoki returns non-deleted paroki matching optional filters and cursor.
func (r *Repository) SearchParoki(
	ctx context.Context,
	q string,
	keuskupanID *uuid.UUID,
	active *bool,
	memberUserID *uuid.UUID,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Paroki, error) {
	query := r.client.Paroki.Query().Where(paroki.DeletedAtIsNil())
	if q != "" {
		query = query.Where(paroki.Or(
			paroki.KodeContainsFold(q),
			paroki.NamaContainsFold(q),
		))
	}
	if keuskupanID != nil && *keuskupanID != uuid.Nil {
		query = query.Where(paroki.KeuskupanIDEQ(*keuskupanID))
	}
	if active != nil {
		query = query.Where(paroki.ActiveEQ(*active))
	}
	if memberUserID != nil && *memberUserID != uuid.Nil {
		query = query.Where(paroki.HasMembersWith(parokimember.UserIDEQ(*memberUserID)))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(paroki.Or(
			paroki.CreatedAtLT(*cursorCreatedAt),
			paroki.And(
				paroki.CreatedAtEQ(*cursorCreatedAt),
				paroki.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(paroki.FieldCreatedAt), ent.Desc(paroki.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindParokiByID returns a non-deleted Paroki by ID.
func (r *Repository) FindParokiByID(ctx context.Context, id uuid.UUID) (*ent.Paroki, error) {
	return r.client.Paroki.Query().
		Where(paroki.IDEQ(id), paroki.DeletedAtIsNil()).
		Only(ctx)
}

// FindParokiByKode returns an active, non-deleted Paroki by normalized kode.
func (r *Repository) FindParokiByKode(ctx context.Context, kode string) (*ent.Paroki, error) {
	return r.client.Paroki.Query().
		Where(
			paroki.KodeEQ(strings.ToLower(strings.TrimSpace(kode))),
			paroki.DeletedAtIsNil(),
			paroki.ActiveEQ(true),
		).
		Only(ctx)
}

// FindParokiByKodeIncludeDeleted returns a Paroki by kode, including soft-deleted rows.
func (r *Repository) FindParokiByKodeIncludeDeleted(ctx context.Context, kode string) (*ent.Paroki, error) {
	return r.client.Paroki.Query().
		Where(paroki.KodeEQ(kode)).
		Only(ctx)
}

// CreateParoki inserts a new Paroki.
func (r *Repository) CreateParoki(ctx context.Context, in ParokiInput) (*ent.Paroki, error) {
	return r.client.Paroki.Create().
		SetKeuskupanID(in.KeuskupanID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetNillablePelindung(in.Pelindung).
		SetNillablePastor(in.Pastor).
		SetNillableAlamat(in.Alamat).
		SetNillableTelepon(in.Telepon).
		SetNillableEmail(in.Email).
		SetNillableWebsite(in.Website).
		SetNillableLogoURL(in.LogoURL).
		SetNillableKopSurat(in.KopSurat).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateParoki restores a soft-deleted Paroki with new field values.
func (r *Repository) ReactivateParoki(ctx context.Context, id uuid.UUID, in ParokiInput) (*ent.Paroki, error) {
	upd := r.client.Paroki.UpdateOneID(id).
		ClearDeletedAt().
		SetKeuskupanID(in.KeuskupanID).
		SetKode(in.Kode).
		SetNama(in.Nama).
		SetActive(true)
	setOrClearParokiString(in.Pelindung, upd.SetPelindung, upd.ClearPelindung)
	setOrClearParokiString(in.Pastor, upd.SetPastor, upd.ClearPastor)
	setOrClearParokiString(in.Alamat, upd.SetAlamat, upd.ClearAlamat)
	setOrClearParokiString(in.Telepon, upd.SetTelepon, upd.ClearTelepon)
	setOrClearParokiString(in.Email, upd.SetEmail, upd.ClearEmail)
	setOrClearParokiString(in.Website, upd.SetWebsite, upd.ClearWebsite)
	setOrClearParokiString(in.LogoURL, upd.SetLogoURL, upd.ClearLogoURL)
	setOrClearParokiString(in.KopSurat, upd.SetKopSurat, upd.ClearKopSurat)
	return upd.Save(ctx)
}

// UpdateParoki updates mutable fields of a Paroki.
func (r *Repository) UpdateParoki(
	ctx context.Context,
	id uuid.UUID,
	keuskupanID *uuid.UUID,
	kode, nama, pelindung, pastor, alamat, telepon, email, website, logoURL, kopSurat *string,
	active *bool,
) (*ent.Paroki, error) {
	upd := r.client.Paroki.UpdateOneID(id)
	if keuskupanID != nil {
		upd.SetKeuskupanID(*keuskupanID)
	}
	if kode != nil {
		upd.SetKode(*kode)
	}
	if nama != nil {
		upd.SetNama(*nama)
	}
	if pelindung != nil {
		setOrClearParokiString(pelindung, upd.SetPelindung, upd.ClearPelindung)
	}
	if pastor != nil {
		setOrClearParokiString(pastor, upd.SetPastor, upd.ClearPastor)
	}
	if alamat != nil {
		setOrClearParokiString(alamat, upd.SetAlamat, upd.ClearAlamat)
	}
	if telepon != nil {
		setOrClearParokiString(telepon, upd.SetTelepon, upd.ClearTelepon)
	}
	if email != nil {
		setOrClearParokiString(email, upd.SetEmail, upd.ClearEmail)
	}
	if website != nil {
		setOrClearParokiString(website, upd.SetWebsite, upd.ClearWebsite)
	}
	if logoURL != nil {
		setOrClearParokiString(logoURL, upd.SetLogoURL, upd.ClearLogoURL)
	}
	if kopSurat != nil {
		setOrClearParokiString(kopSurat, upd.SetKopSurat, upd.ClearKopSurat)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteParoki sets deleted_at and marks the Paroki inactive.
func (r *Repository) SoftDeleteParoki(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Paroki.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasParokiChildren reports whether any non-deleted Wilayah, Lingkungan, Lookup, Organization, Event, Period, AgentSession, or AgentMemory belongs to parokiID.
func (r *Repository) HasParokiChildren(ctx context.Context, parokiID uuid.UUID) (bool, error) {
	if has, err := r.client.Wilayah.Query().Where(wilayah.ParokiIDEQ(parokiID), wilayah.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Lingkungan.Query().Where(lingkungan.ParokiIDEQ(parokiID), lingkungan.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Lookup.Query().Where(lookup.ParokiIDEQ(parokiID), lookup.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Organization.Query().Where(organization.ParokiIDEQ(parokiID), organization.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Event.Query().Where(event.ParokiIDEQ(parokiID), event.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Period.Query().Where(period.ParokiIDEQ(parokiID), period.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.AgentSession.Query().Where(agentsession.ParokiIDEQ(parokiID), agentsession.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	return r.client.AgentMemory.Query().Where(agentmemory.ParokiIDEQ(parokiID)).Exist(ctx)
}

// SearchParokiMembers returns paginated ParokiMember rows along with a User lookup map.
func (r *Repository) SearchParokiMembers(
	ctx context.Context,
	parokiID uuid.UUID,
	q, role string,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.ParokiMember, map[uuid.UUID]*ent.User, error) {
	query := r.client.ParokiMember.Query().Where(parokimember.ParokiIDEQ(parokiID))
	if q != "" {
		userIDs, err := r.client.User.Query().
			Where(user.DeletedAtIsNil(), user.Or(
				user.NameContainsFold(q),
				user.EmailContainsFold(q),
			)).
			IDs(ctx)
		if err != nil {
			return nil, nil, err
		}
		if len(userIDs) == 0 {
			return []*ent.ParokiMember{}, map[uuid.UUID]*ent.User{}, nil
		}
		query = query.Where(parokimember.UserIDIn(userIDs...))
	}
	if role != "" {
		query = query.Where(parokimember.RoleEQ(parokimember.Role(role)))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(parokimember.Or(
			parokimember.CreatedAtLT(*cursorCreatedAt),
			parokimember.And(
				parokimember.CreatedAtEQ(*cursorCreatedAt),
				parokimember.IDLT(*cursorID),
			),
		))
	}
	rows, err := query.
		Order(ent.Desc(parokimember.FieldCreatedAt), ent.Desc(parokimember.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	usersMap := make(map[uuid.UUID]*ent.User, len(rows))
	if len(rows) > 0 {
		ids := make([]uuid.UUID, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.UserID)
		}
		users, err := r.client.User.Query().Where(user.IDIn(ids...)).All(ctx)
		if err != nil {
			return nil, nil, err
		}
		for _, u := range users {
			usersMap[u.ID] = u
		}
	}
	return rows, usersMap, nil
}

// FindParokiMember returns the ParokiMember record for (parokiID, userID).
func (r *Repository) FindParokiMember(ctx context.Context, parokiID, userID uuid.UUID) (*ent.ParokiMember, error) {
	return r.client.ParokiMember.Query().
		Where(parokimember.ParokiIDEQ(parokiID), parokimember.UserIDEQ(userID)).
		Only(ctx)
}

// CreateParokiMember inserts a new ParokiMember record.
func (r *Repository) CreateParokiMember(ctx context.Context, parokiID, userID uuid.UUID, role parokimember.Role) (*ent.ParokiMember, error) {
	return r.client.ParokiMember.Create().
		SetParokiID(parokiID).
		SetUserID(userID).
		SetRole(role).
		Save(ctx)
}

// UpdateParokiMemberRole updates a ParokiMember's role.
func (r *Repository) UpdateParokiMemberRole(ctx context.Context, id uuid.UUID, role parokimember.Role) (*ent.ParokiMember, error) {
	return r.client.ParokiMember.UpdateOneID(id).
		SetRole(role).
		Save(ctx)
}

// DeleteParokiMember hard-deletes a ParokiMember record.
func (r *Repository) DeleteParokiMember(ctx context.Context, id uuid.UUID) error {
	return r.client.ParokiMember.DeleteOneID(id).Exec(ctx)
}

// CountParokiAdmins returns the number of admins in parokiID.
func (r *Repository) CountParokiAdmins(ctx context.Context, parokiID uuid.UUID) (int, error) {
	return r.client.ParokiMember.Query().
		Where(parokimember.ParokiIDEQ(parokiID), parokimember.RoleEQ(parokimember.RoleAdmin)).
		Count(ctx)
}

func setOrClearParokiString(val *string, setFn func(string) *ent.ParokiUpdateOne, clearFn func() *ent.ParokiUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
