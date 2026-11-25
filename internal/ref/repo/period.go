package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/period"
)

// PeriodInput holds fields for creating or reactivating a Period.
type PeriodInput struct {
	ParokiID  uuid.UUID
	Type      period.Type
	Code      string
	Name      string
	StartDate string
	EndDate   string
	IsCurrent bool
}

// SearchPeriods returns paginated non-deleted periods for parokiID.
func (r *Repository) SearchPeriods(
	ctx context.Context,
	parokiID uuid.UUID,
	q, periodType, status string,
	isCurrent *bool,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.Period, error) {
	query := r.client.Period.Query().Where(period.ParokiIDEQ(parokiID), period.DeletedAtIsNil())
	if q != "" {
		query = query.Where(period.Or(
			period.CodeContainsFold(q),
			period.NameContainsFold(q),
		))
	}
	if periodType != "" {
		query = query.Where(period.TypeEQ(period.Type(periodType)))
	}
	if status != "" {
		query = query.Where(period.StatusEQ(period.Status(status)))
	}
	if isCurrent != nil {
		query = query.Where(period.IsCurrentEQ(*isCurrent))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(period.Or(
			period.CreatedAtLT(*cursorCreatedAt),
			period.And(
				period.CreatedAtEQ(*cursorCreatedAt),
				period.IDLT(*cursorID),
			),
		))
	}
	return query.
		Order(ent.Desc(period.FieldCreatedAt), ent.Desc(period.FieldID)).
		Limit(limit).
		All(ctx)
}

// FindPeriodByCodeIncludeDeleted returns a Period by (parokiID, code), including soft-deleted rows.
func (r *Repository) FindPeriodByCodeIncludeDeleted(ctx context.Context, parokiID uuid.UUID, code string) (*ent.Period, error) {
	return r.client.Period.Query().
		Where(period.ParokiIDEQ(parokiID), period.CodeEQ(code)).
		Only(ctx)
}

// HasOverlappingPeriod reports whether any non-deleted Period of the same (parokiID, type) overlaps [startDate, endDate].
func (r *Repository) HasOverlappingPeriod(
	ctx context.Context,
	parokiID uuid.UUID,
	periodType period.Type,
	startDate, endDate string,
	excludeID *uuid.UUID,
) (bool, error) {
	query := r.client.Period.Query().
		Where(
			period.ParokiIDEQ(parokiID),
			period.TypeEQ(periodType),
			period.DeletedAtIsNil(),
			period.StartDateLTE(endDate),
			period.EndDateGTE(startDate),
		)
	if excludeID != nil {
		query = query.Where(period.IDNEQ(*excludeID))
	}
	return query.Exist(ctx)
}

// ClearCurrentPeriod sets is_current = false on sibling periods of the same (parokiID, type).
func (r *Repository) ClearCurrentPeriod(ctx context.Context, parokiID uuid.UUID, periodType period.Type, excludeID *uuid.UUID) error {
	upd := r.client.Period.Update().
		Where(
			period.ParokiIDEQ(parokiID),
			period.TypeEQ(periodType),
			period.IsCurrentEQ(true),
			period.DeletedAtIsNil(),
		)
	if excludeID != nil {
		upd = upd.Where(period.IDNEQ(*excludeID))
	}
	return upd.SetIsCurrent(false).Exec(ctx)
}

// CreatePeriod inserts a new Period.
func (r *Repository) CreatePeriod(ctx context.Context, in PeriodInput) (*ent.Period, error) {
	return r.client.Period.Create().
		SetParokiID(in.ParokiID).
		SetType(in.Type).
		SetCode(in.Code).
		SetName(in.Name).
		SetStartDate(in.StartDate).
		SetEndDate(in.EndDate).
		SetStatus(period.StatusOpen).
		SetIsCurrent(in.IsCurrent).
		Save(ctx)
}

// ReactivatePeriod restores a soft-deleted Period with fresh fields.
func (r *Repository) ReactivatePeriod(ctx context.Context, id uuid.UUID, in PeriodInput) (*ent.Period, error) {
	return r.client.Period.UpdateOneID(id).
		ClearDeletedAt().
		SetType(in.Type).
		SetCode(in.Code).
		SetName(in.Name).
		SetStartDate(in.StartDate).
		SetEndDate(in.EndDate).
		SetStatus(period.StatusOpen).
		SetIsCurrent(in.IsCurrent).
		Save(ctx)
}

// UpdatePeriod updates mutable fields of a Period.
func (r *Repository) UpdatePeriod(
	ctx context.Context,
	id uuid.UUID,
	code, name, startDate, endDate *string,
	status *period.Status,
	isCurrent *bool,
) (*ent.Period, error) {
	upd := r.client.Period.UpdateOneID(id)
	if code != nil {
		upd.SetCode(*code)
	}
	if name != nil {
		upd.SetName(*name)
	}
	if startDate != nil {
		upd.SetStartDate(*startDate)
	}
	if endDate != nil {
		upd.SetEndDate(*endDate)
	}
	if status != nil {
		upd.SetStatus(*status)
	}
	if isCurrent != nil {
		upd.SetIsCurrent(*isCurrent)
	}
	return upd.Save(ctx)
}

// HasOrgMembersByPeriod reports whether any OrgMember in parokiID references periodID.
func (r *Repository) HasOrgMembersByPeriod(ctx context.Context, parokiID, periodID uuid.UUID) (bool, error) {
	return r.client.OrgMember.Query().
		Where(orgmember.ParokiIDEQ(parokiID), orgmember.PeriodIDEQ(periodID)).
		Exist(ctx)
}

// SoftDeletePeriod sets deleted_at and clears is_current on a Period.
func (r *Repository) SoftDeletePeriod(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.client.Period.UpdateOneID(id).
		SetDeletedAt(at).
		SetIsCurrent(false).
		Exec(ctx)
}
