package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/diocese"
	"github.com/paroki/domus/ent/parish"
)

// DioceseInput holds mutable fields for creating or updating a Diocese.
type DioceseInput struct {
	Code       string
	Name       string
	BishopName *string
	Address    *string
	Phone      *string
	Email      *string
	Website    *string
	Active     bool
}

// SearchDioceses returns non-deleted dioceses matching optional filters and cursor.
func (r *Repository) SearchDioceses(
	ctx context.Context,
	q string,
	active *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Diocese, error) {
	query := r.client.Diocese.Query().Where(diocese.DeletedAtIsNil())
	if q != "" {
		query = query.Where(diocese.Or(
			diocese.CodeContainsFold(q),
			diocese.NameContainsFold(q),
		))
	}
	if active != nil {
		query = query.Where(diocese.ActiveEQ(*active))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(diocese.Or(
			diocese.CreatedAtLT(*cursorCreatedAt),
			diocese.And(
				diocese.CreatedAtEQ(*cursorCreatedAt),
				diocese.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(diocese.FieldCreatedAt), ent.Desc(diocese.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindDioceseByID returns a non-deleted Diocese by ID.
func (r *Repository) FindDioceseByID(ctx context.Context, id uuid.UUID) (*ent.Diocese, error) {
	return r.client.Diocese.Query().
		Where(diocese.IDEQ(id), diocese.DeletedAtIsNil()).
		Only(ctx)
}

// FindDioceseByCodeIncludeDeleted returns a Diocese by code, including soft-deleted rows.
func (r *Repository) FindDioceseByCodeIncludeDeleted(ctx context.Context, code string) (*ent.Diocese, error) {
	return r.client.Diocese.Query().
		Where(diocese.CodeEQ(code)).
		Only(ctx)
}

// CreateDiocese inserts a new Diocese.
func (r *Repository) CreateDiocese(ctx context.Context, in DioceseInput) (*ent.Diocese, error) {
	return r.client.Diocese.Create().
		SetCode(in.Code).
		SetName(in.Name).
		SetNillableBishopName(in.BishopName).
		SetNillableAddress(in.Address).
		SetNillablePhone(in.Phone).
		SetNillableEmail(in.Email).
		SetNillableWebsite(in.Website).
		SetActive(in.Active).
		Save(ctx)
}

// ReactivateDiocese restores a soft-deleted Diocese with new field values.
func (r *Repository) ReactivateDiocese(ctx context.Context, id uuid.UUID, in DioceseInput) (*ent.Diocese, error) {
	upd := r.client.Diocese.UpdateOneID(id).
		ClearDeletedAt().
		SetCode(in.Code).
		SetName(in.Name).
		SetActive(true)
	setOrClearString(in.BishopName, upd.SetBishopName, upd.ClearBishopName)
	setOrClearString(in.Address, upd.SetAddress, upd.ClearAddress)
	setOrClearString(in.Phone, upd.SetPhone, upd.ClearPhone)
	setOrClearString(in.Email, upd.SetEmail, upd.ClearEmail)
	setOrClearString(in.Website, upd.SetWebsite, upd.ClearWebsite)
	return upd.Save(ctx)
}

// UpdateDiocese updates mutable fields of a Diocese.
func (r *Repository) UpdateDiocese(
	ctx context.Context,
	id uuid.UUID,
	code, name, bishopName, address, phone, email, website *string,
	active *bool,
) (*ent.Diocese, error) {
	upd := r.client.Diocese.UpdateOneID(id)
	if code != nil {
		upd.SetCode(*code)
	}
	if name != nil {
		upd.SetName(*name)
	}
	if bishopName != nil {
		setOrClearString(bishopName, upd.SetBishopName, upd.ClearBishopName)
	}
	if address != nil {
		setOrClearString(address, upd.SetAddress, upd.ClearAddress)
	}
	if phone != nil {
		setOrClearString(phone, upd.SetPhone, upd.ClearPhone)
	}
	if email != nil {
		setOrClearString(email, upd.SetEmail, upd.ClearEmail)
	}
	if website != nil {
		setOrClearString(website, upd.SetWebsite, upd.ClearWebsite)
	}
	if active != nil {
		upd.SetActive(*active)
	}
	return upd.Save(ctx)
}

// SoftDeleteDiocese sets deleted_at and marks the Diocese inactive.
func (r *Repository) SoftDeleteDiocese(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Diocese.UpdateOneID(id).
		SetDeletedAt(at).
		SetActive(false).
		Exec(ctx)
}

// HasActiveParishesByDiocese reports whether any non-deleted active Parish references dioceseID.
func (r *Repository) HasActiveParishesByDiocese(ctx context.Context, dioceseID uuid.UUID) (bool, error) {
	return r.client.Parish.Query().
		Where(parish.DioceseIDEQ(dioceseID), parish.ActiveEQ(true), parish.DeletedAtIsNil()).
		Exist(ctx)
}

func setOrClearString(val *string, setFn func(string) *ent.DioceseUpdateOne, clearFn func() *ent.DioceseUpdateOne) {
	if val == nil || *val == "" {
		clearFn()
		return
	}
	setFn(*val)
}
