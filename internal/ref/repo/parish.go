package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/lookup"
	"github.com/paroki/domus/ent/parish"
	"github.com/paroki/domus/ent/parishmember"
	"github.com/paroki/domus/ent/period"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/ent/ward"
	"github.com/paroki/domus/ent/zone"
)

// ParishInput holds fields for creating or reactivating a Parish.
type ParishInput struct {
	DioceseID   uuid.UUID
	Code        string
	Name        string
	PatronSaint *string
	PastorName  *string
	Address     *string
	Phone       *string
	Email       *string
	Website     *string
	LogoURL     *string
	Letterhead  *string
	Active      bool
}

// SearchParishes returns non-deleted parishes matching optional filters and cursor.
func (r *Repository) SearchParishes(
	ctx context.Context,
	q string,
	dioceseID *uuid.UUID,
	active *bool,
	memberUserID *uuid.UUID,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Parish, error) {
	query := r.client.Parish.Query().Where(parish.DeletedAtIsNil())
	if q != "" {
		query = query.Where(parish.Or(
			parish.CodeContainsFold(q),
			parish.NameContainsFold(q),
		))
	}
	if dioceseID != nil && *dioceseID != uuid.Nil {
		query = query.Where(parish.DioceseIDEQ(*dioceseID))
	}
	if active != nil {
		query = query.Where(parish.ActiveEQ(*active))
	}
	if memberUserID != nil && *memberUserID != uuid.Nil {
		query = query.Where(parish.HasMembersWith(parishmember.UserIDEQ(*memberUserID)))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(parish.Or(
			parish.CreatedAtLT(*cursorCreatedAt),
			parish.And(
				parish.CreatedAtEQ(*cursorCreatedAt),
				parish.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(parish.FieldCreatedAt), ent.Desc(parish.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindParishByID returns a non-deleted Parish by ID.
func (r *Repository) FindParishByID(ctx context.Context, id uuid.UUID) (*ent.Parish, error) {
	return r.client.Parish.Query().
		Where(parish.IDEQ(id), parish.DeletedAtIsNil()).
		Only(ctx)
}

// FindParishByCodeIncludeDeleted returns a Parish by code, including soft-deleted rows.
func (r *Repository) FindParishByCodeIncludeDeleted(ctx context.Context, code string) (*ent.Parish, error) {
	return r.client.Parish.Query().
		Where(parish.CodeEQ(code)).
		Only(ctx)
}

// CreateParish inserts a new Parish.
func (r *Repository) CreateParish(ctx context.Context, in ParishInput) (*ent.Parish, error) {
	return r.client.Parish.Create().
		SetDioceseID(in.DioceseID).
		SetCode(in.Code).
		SetName(in.Name).
		SetNillablePatronSaint(in.PatronSaint).
		SetNillablePastorName(in.PastorName).
		SetNillableAddress(in.Address).
		SetNillablePhone(in.Phone).
		SetNillableEmail(in.Email).
		SetNillableWebsite(in.Website).
		SetNillableLogoURL(in.LogoURL).
		SetNillableLetterhead(in.Letterhead).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateParish restores a soft-deleted Parish with new field values.
func (r *Repository) ReactivateParish(ctx context.Context, id uuid.UUID, in ParishInput) (*ent.Parish, error) {
	upd := r.client.Parish.UpdateOneID(id).
		ClearDeletedAt().
		SetDioceseID(in.DioceseID).
		SetCode(in.Code).
		SetName(in.Name).
		SetActive(true)
	setOrClearParishString(in.PatronSaint, upd.SetPatronSaint, upd.ClearPatronSaint)
	setOrClearParishString(in.PastorName, upd.SetPastorName, upd.ClearPastorName)
	setOrClearParishString(in.Address, upd.SetAddress, upd.ClearAddress)
	setOrClearParishString(in.Phone, upd.SetPhone, upd.ClearPhone)
	setOrClearParishString(in.Email, upd.SetEmail, upd.ClearEmail)
	setOrClearParishString(in.Website, upd.SetWebsite, upd.ClearWebsite)
	setOrClearParishString(in.LogoURL, upd.SetLogoURL, upd.ClearLogoURL)
	setOrClearParishString(in.Letterhead, upd.SetLetterhead, upd.ClearLetterhead)
	return upd.Save(ctx)
}

// UpdateParish updates mutable fields of a Parish.
func (r *Repository) UpdateParish(
	ctx context.Context,
	id uuid.UUID,
	dioceseID *uuid.UUID,
	code, name, patronSaint, pastorName, address, phone, email, website, logoURL, letterhead *string,
	active *bool,
) (*ent.Parish, error) {
	upd := r.client.Parish.UpdateOneID(id)
	if dioceseID != nil {
		upd.SetDioceseID(*dioceseID)
	}
	if code != nil {
		upd.SetCode(*code)
	}
	if name != nil {
		upd.SetName(*name)
	}
	if patronSaint != nil {
		setOrClearParishString(patronSaint, upd.SetPatronSaint, upd.ClearPatronSaint)
	}
	if pastorName != nil {
		setOrClearParishString(pastorName, upd.SetPastorName, upd.ClearPastorName)
	}
	if address != nil {
		setOrClearParishString(address, upd.SetAddress, upd.ClearAddress)
	}
	if phone != nil {
		setOrClearParishString(phone, upd.SetPhone, upd.ClearPhone)
	}
	if email != nil {
		setOrClearParishString(email, upd.SetEmail, upd.ClearEmail)
	}
	if website != nil {
		setOrClearParishString(website, upd.SetWebsite, upd.ClearWebsite)
	}
	if logoURL != nil {
		setOrClearParishString(logoURL, upd.SetLogoURL, upd.ClearLogoURL)
	}
	if letterhead != nil {
		setOrClearParishString(letterhead, upd.SetLetterhead, upd.ClearLetterhead)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteParish sets deleted_at and marks the Parish inactive.
func (r *Repository) SoftDeleteParish(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Parish.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasParishChildren reports whether any non-deleted Zone, Ward, Lookup, or Period belongs to parishID.
func (r *Repository) HasParishChildren(ctx context.Context, parishID uuid.UUID) (bool, error) {
	if has, err := r.client.Zone.Query().Where(zone.ParishIDEQ(parishID), zone.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Ward.Query().Where(ward.ParishIDEQ(parishID), ward.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	if has, err := r.client.Lookup.Query().Where(lookup.ParishIDEQ(parishID), lookup.DeletedAtIsNil()).Exist(ctx); err != nil || has {
		return has, err
	}
	return r.client.Period.Query().Where(period.ParishIDEQ(parishID), period.DeletedAtIsNil()).Exist(ctx)
}

// SearchParishMembers returns paginated ParishMember rows along with a User lookup map.
func (r *Repository) SearchParishMembers(
	ctx context.Context,
	parishID uuid.UUID,
	q, role string,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.ParishMember, map[uuid.UUID]*ent.User, error) {
	query := r.client.ParishMember.Query().Where(parishmember.ParishIDEQ(parishID))
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
			return []*ent.ParishMember{}, map[uuid.UUID]*ent.User{}, nil
		}
		query = query.Where(parishmember.UserIDIn(userIDs...))
	}
	if role != "" {
		query = query.Where(parishmember.RoleEQ(parishmember.Role(role)))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(parishmember.Or(
			parishmember.CreatedAtLT(*cursorCreatedAt),
			parishmember.And(
				parishmember.CreatedAtEQ(*cursorCreatedAt),
				parishmember.IDLT(*cursorID),
			),
		))
	}
	rows, err := query.
		Order(ent.Desc(parishmember.FieldCreatedAt), ent.Desc(parishmember.FieldID)).
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

// FindParishMember returns the ParishMember record for (parishID, userID).
func (r *Repository) FindParishMember(ctx context.Context, parishID, userID uuid.UUID) (*ent.ParishMember, error) {
	return r.client.ParishMember.Query().
		Where(parishmember.ParishIDEQ(parishID), parishmember.UserIDEQ(userID)).
		Only(ctx)
}

// CreateParishMember inserts a new ParishMember record.
func (r *Repository) CreateParishMember(ctx context.Context, parishID, userID uuid.UUID, role parishmember.Role) (*ent.ParishMember, error) {
	return r.client.ParishMember.Create().
		SetParishID(parishID).
		SetUserID(userID).
		SetRole(role).
		Save(ctx)
}

// UpdateParishMemberRole updates a ParishMember's role.
func (r *Repository) UpdateParishMemberRole(ctx context.Context, id uuid.UUID, role parishmember.Role) (*ent.ParishMember, error) {
	return r.client.ParishMember.UpdateOneID(id).
		SetRole(role).
		Save(ctx)
}

// DeleteParishMember hard-deletes a ParishMember record.
func (r *Repository) DeleteParishMember(ctx context.Context, id uuid.UUID) error {
	return r.client.ParishMember.DeleteOneID(id).Exec(ctx)
}

// CountParishAdmins returns the number of admins in parishID.
func (r *Repository) CountParishAdmins(ctx context.Context, parishID uuid.UUID) (int, error) {
	return r.client.ParishMember.Query().
		Where(parishmember.ParishIDEQ(parishID), parishmember.RoleEQ(parishmember.RoleAdmin)).
		Count(ctx)
}

func setOrClearParishString(val *string, setFn func(string) *ent.ParishUpdateOne, clearFn func() *ent.ParishUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
