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
	"github.com/paroki/domus/ent/keluarga"
	"github.com/paroki/domus/ent/lingkungan"
	"github.com/paroki/domus/ent/liturgilagu"
	"github.com/paroki/domus/ent/liturgiministry"
	"github.com/paroki/domus/ent/liturgipetugas"
	"github.com/paroki/domus/ent/lookup"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/paroki"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/ent/umat"
	"github.com/paroki/domus/ent/umatreferensi"
	"github.com/paroki/domus/ent/umatsakramen"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/wilayah"
)

// ParokiInput holds fields for creating or reactivating a Paroki.
type ParokiInput struct {
	KeuskupanID  uuid.UUID
	KevikepanID  *uuid.UUID
	IndukID      *uuid.UUID
	Kode         string
	KodeResmi    *string
	Nama         string
	Pelindung    *string
	Pastor       *string
	Alamat       *string
	KotaID       *uuid.UUID
	KecamatanID  *uuid.UUID
	KelurahanID  *uuid.UUID
	Timezone     string
	Telepon      *string
	Fax          *string
	Email        *string
	Website      *string
	LogoURL      *string
	KopSurat     *string
	Latitude     *float64
	Longitude    *float64
	StatusParoki *string
	Active       bool
}

// ParokiUpdate holds optional mutable fields for updating a Paroki.
type ParokiUpdate struct {
	KeuskupanID      *uuid.UUID
	KevikepanID      *uuid.UUID
	IndukID          *uuid.UUID
	Kode             *string
	KodeResmi        *string
	Nama             *string
	Pelindung        *string
	Pastor           *string
	Alamat           *string
	KotaID           *uuid.UUID
	KecamatanID      *uuid.UUID
	KelurahanID      *uuid.UUID
	Timezone         *string
	Telepon          *string
	Fax              *string
	Email            *string
	Website          *string
	LogoURL          *string
	KopSurat         *string
	Latitude         *float64
	Longitude        *float64
	ClearCoordinates bool
	StatusParoki     *string
	Active           *bool
}

// SearchParoki returns non-deleted paroki matching optional filters and cursor.
func (r *Repository) SearchParoki(
	ctx context.Context,
	q string,
	keuskupanID *uuid.UUID,
	kevikepanID *uuid.UUID,
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
	if kevikepanID != nil && *kevikepanID != uuid.Nil {
		query = query.Where(paroki.KevikepanIDEQ(*kevikepanID))
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
		SetNillableKevikepanID(in.KevikepanID).
		SetNillableIndukID(in.IndukID).
		SetKode(in.Kode).
		SetNillableKodeResmi(in.KodeResmi).
		SetNama(in.Nama).
		SetNillablePelindung(in.Pelindung).
		SetNillablePastor(in.Pastor).
		SetNillableAlamat(in.Alamat).
		SetNillableKotaID(in.KotaID).
		SetNillableKecamatanID(in.KecamatanID).
		SetNillableKelurahanID(in.KelurahanID).
		SetTimezone(in.Timezone).
		SetNillableTelepon(in.Telepon).
		SetNillableFax(in.Fax).
		SetNillableEmail(in.Email).
		SetNillableWebsite(in.Website).
		SetNillableLogoURL(in.LogoURL).
		SetNillableKopSurat(in.KopSurat).
		SetNillableLatitude(in.Latitude).
		SetNillableLongitude(in.Longitude).
		SetNillableStatusParoki(in.StatusParoki).
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
		SetTimezone(in.Timezone).
		SetActive(in.Active)
	setOrClearParokiUUID(in.KevikepanID, upd.SetKevikepanID, upd.ClearKevikepanID)
	setOrClearParokiUUID(in.IndukID, upd.SetIndukID, upd.ClearIndukID)
	setOrClearParokiString(in.KodeResmi, upd.SetKodeResmi, upd.ClearKodeResmi)
	setOrClearParokiString(in.Pelindung, upd.SetPelindung, upd.ClearPelindung)
	setOrClearParokiString(in.Pastor, upd.SetPastor, upd.ClearPastor)
	setOrClearParokiString(in.Alamat, upd.SetAlamat, upd.ClearAlamat)
	setOrClearParokiUUID(in.KotaID, upd.SetKotaID, upd.ClearKotaID)
	setOrClearParokiUUID(in.KecamatanID, upd.SetKecamatanID, upd.ClearKecamatanID)
	setOrClearParokiUUID(in.KelurahanID, upd.SetKelurahanID, upd.ClearKelurahanID)
	setOrClearParokiString(in.Telepon, upd.SetTelepon, upd.ClearTelepon)
	setOrClearParokiString(in.Fax, upd.SetFax, upd.ClearFax)
	setOrClearParokiString(in.Email, upd.SetEmail, upd.ClearEmail)
	setOrClearParokiString(in.Website, upd.SetWebsite, upd.ClearWebsite)
	setOrClearParokiString(in.LogoURL, upd.SetLogoURL, upd.ClearLogoURL)
	setOrClearParokiString(in.KopSurat, upd.SetKopSurat, upd.ClearKopSurat)
	if in.Latitude != nil && in.Longitude != nil {
		upd.SetLatitude(*in.Latitude).SetLongitude(*in.Longitude)
	} else {
		upd.ClearLatitude().ClearLongitude()
	}
	setOrClearParokiString(in.StatusParoki, upd.SetStatusParoki, upd.ClearStatusParoki)
	return upd.Save(ctx)
}

// UpdateParoki updates mutable fields of a Paroki.
func (r *Repository) UpdateParoki(ctx context.Context, id uuid.UUID, in ParokiUpdate) (*ent.Paroki, error) {
	upd := r.client.Paroki.UpdateOneID(id)
	if in.KeuskupanID != nil {
		upd.SetKeuskupanID(*in.KeuskupanID)
	}
	if in.KevikepanID != nil {
		setOrClearParokiUUID(in.KevikepanID, upd.SetKevikepanID, upd.ClearKevikepanID)
	}
	if in.IndukID != nil {
		setOrClearParokiUUID(in.IndukID, upd.SetIndukID, upd.ClearIndukID)
	}
	if in.Kode != nil {
		upd.SetKode(*in.Kode)
	}
	if in.KodeResmi != nil {
		setOrClearParokiString(in.KodeResmi, upd.SetKodeResmi, upd.ClearKodeResmi)
	}
	if in.Nama != nil {
		upd.SetNama(*in.Nama)
	}
	if in.Pelindung != nil {
		setOrClearParokiString(in.Pelindung, upd.SetPelindung, upd.ClearPelindung)
	}
	if in.Pastor != nil {
		setOrClearParokiString(in.Pastor, upd.SetPastor, upd.ClearPastor)
	}
	if in.Alamat != nil {
		setOrClearParokiString(in.Alamat, upd.SetAlamat, upd.ClearAlamat)
	}
	if in.KotaID != nil {
		setOrClearParokiUUID(in.KotaID, upd.SetKotaID, upd.ClearKotaID)
	}
	if in.KecamatanID != nil {
		setOrClearParokiUUID(in.KecamatanID, upd.SetKecamatanID, upd.ClearKecamatanID)
	}
	if in.KelurahanID != nil {
		setOrClearParokiUUID(in.KelurahanID, upd.SetKelurahanID, upd.ClearKelurahanID)
	}
	if in.Timezone != nil {
		upd.SetTimezone(*in.Timezone)
	}
	if in.Telepon != nil {
		setOrClearParokiString(in.Telepon, upd.SetTelepon, upd.ClearTelepon)
	}
	if in.Fax != nil {
		setOrClearParokiString(in.Fax, upd.SetFax, upd.ClearFax)
	}
	if in.Email != nil {
		setOrClearParokiString(in.Email, upd.SetEmail, upd.ClearEmail)
	}
	if in.Website != nil {
		setOrClearParokiString(in.Website, upd.SetWebsite, upd.ClearWebsite)
	}
	if in.LogoURL != nil {
		setOrClearParokiString(in.LogoURL, upd.SetLogoURL, upd.ClearLogoURL)
	}
	if in.KopSurat != nil {
		setOrClearParokiString(in.KopSurat, upd.SetKopSurat, upd.ClearKopSurat)
	}
	if in.ClearCoordinates {
		upd.ClearLatitude().ClearLongitude()
	} else {
		if in.Latitude != nil {
			upd.SetLatitude(*in.Latitude)
		}
		if in.Longitude != nil {
			upd.SetLongitude(*in.Longitude)
		}
	}
	if in.StatusParoki != nil {
		setOrClearParokiString(in.StatusParoki, upd.SetStatusParoki, upd.ClearStatusParoki)
	}
	if in.Active != nil {
		upd.SetActive(*in.Active)
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

// HasChildParoki reports whether any non-deleted child Paroki references parokiID via induk_id.
func (r *Repository) HasChildParoki(ctx context.Context, parokiID uuid.UUID) (bool, error) {
	return r.client.Paroki.Query().
		Where(paroki.IndukIDEQ(parokiID), paroki.DeletedAtIsNil()).
		Exist(ctx)
}

// HasParokiChildren reports whether any non-deleted child Paroki, Wilayah, Lingkungan, Lookup, Organization, Event, AgentSession, or AgentMemory belongs to parokiID.
func (r *Repository) HasParokiChildren(ctx context.Context, parokiID uuid.UUID) (bool, error) {
	if has, err := r.HasChildParoki(ctx, parokiID); err != nil || has {
		return has, err
	}
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
	if has, err := r.client.AgentSession.Query().Where(agentsession.ParokiIDEQ(parokiID), agentsession.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.AgentMemory.Query().Where(agentmemory.ParokiIDEQ(parokiID)).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Keluarga.Query().Where(keluarga.ParokiIDEQ(parokiID), keluarga.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Umat.Query().Where(umat.ParokiIDEQ(parokiID), umat.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.UmatReferensi.Query().Where(umatreferensi.ParokiIDEQ(parokiID), umatreferensi.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.LiturgiMinistry.Query().Where(liturgiministry.ParokiIDEQ(parokiID)).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.LiturgiPetugas.Query().Where(liturgipetugas.ParokiIDEQ(parokiID)).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.LiturgiLagu.Query().Where(liturgilagu.ParokiIDEQ(parokiID)).Exist(ctx); err != nil || has {
		return has, err
	}
	return r.client.UmatSakramen.Query().Where(umatsakramen.Or(
		umatsakramen.ParokiIDEQ(parokiID),
		umatsakramen.ParokiSakramenIDEQ(parokiID),
	)).Exist(ctx)
}

// SearchParokiMembers returns paginated ParokiMember rows along with a User lookup map.
func (r *Repository) SearchParokiMembers(
	ctx context.Context,
	parokiID uuid.UUID,
	q, role string,
	umatAdmin, liturgiAdmin *bool,
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
	if umatAdmin != nil {
		query = query.Where(parokimember.UmatAdminEQ(*umatAdmin))
	}
	if liturgiAdmin != nil {
		query = query.Where(parokimember.LiturgiAdminEQ(*liturgiAdmin))
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
func (r *Repository) CreateParokiMember(ctx context.Context, parokiID, userID uuid.UUID, role parokimember.Role, umatAdmin, liturgiAdmin bool) (*ent.ParokiMember, error) {
	return r.client.ParokiMember.Create().
		SetParokiID(parokiID).
		SetUserID(userID).
		SetRole(role).
		SetUmatAdmin(umatAdmin).
		SetLiturgiAdmin(liturgiAdmin).
		Save(ctx)
}

// UpdateParokiMember updates a ParokiMember's role, umatAdmin, and/or liturgiAdmin flag.
func (r *Repository) UpdateParokiMember(ctx context.Context, id uuid.UUID, role *parokimember.Role, umatAdmin, liturgiAdmin *bool) (*ent.ParokiMember, error) {
	upd := r.client.ParokiMember.UpdateOneID(id)
	if role != nil {
		upd.SetRole(*role)
	}
	if umatAdmin != nil {
		upd.SetUmatAdmin(*umatAdmin)
	}
	if liturgiAdmin != nil {
		upd.SetLiturgiAdmin(*liturgiAdmin)
	}
	return upd.Save(ctx)
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

func setOrClearParokiUUID(val *uuid.UUID, setFn func(uuid.UUID) *ent.ParokiUpdateOne, clearFn func() *ent.ParokiUpdateOne) {
	if val == nil || *val == uuid.Nil {
		clearFn()
		return
	}
	setFn(*val)
}

func setOrClearParokiString(val *string, setFn func(string) *ent.ParokiUpdateOne, clearFn func() *ent.ParokiUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
