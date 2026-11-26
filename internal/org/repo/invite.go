package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/orginvite"
	"github.com/paroki/domus/ent/paroki"
	"github.com/paroki/domus/internal/org/domain"
)

// OrgInviteInput holds fields for creating an OrgInvite.
type OrgInviteInput struct {
	ParokiID  uuid.UUID
	OrgID     uuid.UUID
	Code      string
	Role      orginvite.Role
	MaxUses   *int
	ExpiresAt time.Time
	CreatedBy uuid.UUID
}

// CreateOrgInvite inserts a new OrgInvite row.
func (r *Repository) CreateOrgInvite(ctx context.Context, in OrgInviteInput) (*ent.OrgInvite, error) {
	q := r.client.OrgInvite.Create().
		SetParokiID(in.ParokiID).
		SetOrgID(in.OrgID).
		SetCode(in.Code).
		SetRole(in.Role).
		SetExpiresAt(in.ExpiresAt).
		SetCreatedBy(in.CreatedBy)
	if in.MaxUses != nil {
		q.SetMaxUses(*in.MaxUses)
	}
	return q.Save(ctx)
}

// SearchOrgInvites returns paginated OrgInvite rows for orgID.
func (r *Repository) SearchOrgInvites(
	ctx context.Context,
	orgID uuid.UUID,
	status string,
	now time.Time,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.OrgInvite, error) {
	query := r.client.OrgInvite.Query().Where(orginvite.OrgIDEQ(orgID))
	switch status {
	case "revoked":
		query = query.Where(orginvite.RevokedAtNotNil())
	case "expired":
		query = query.Where(orginvite.RevokedAtIsNil(), orginvite.ExpiresAtLTE(now))
	case "exhausted":
		query = query.Where(
			orginvite.RevokedAtIsNil(),
			orginvite.ExpiresAtGT(now),
			orginvite.MaxUsesNotNil(),
			func(s *sql.Selector) {
				s.Where(sql.ColumnsGTE(s.C(orginvite.FieldUsedCount), s.C(orginvite.FieldMaxUses)))
			},
		)
	case "active":
		query = query.Where(
			orginvite.RevokedAtIsNil(),
			orginvite.ExpiresAtGT(now),
			func(s *sql.Selector) {
				s.Where(sql.Or(
					sql.IsNull(s.C(orginvite.FieldMaxUses)),
					sql.ColumnsLT(s.C(orginvite.FieldUsedCount), s.C(orginvite.FieldMaxUses)),
				))
			},
		)
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(orginvite.Or(
			orginvite.CreatedAtLT(*cursorCreatedAt),
			orginvite.And(orginvite.CreatedAtEQ(*cursorCreatedAt), orginvite.IDLT(*cursorID)),
		))
	}
	return query.
		Order(orginvite.ByCreatedAt(sql.OrderDesc()), orginvite.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

// FindOrgInviteInOrg returns an OrgInvite belonging to orgID.
func (r *Repository) FindOrgInviteInOrg(ctx context.Context, orgID, inviteID uuid.UUID) (*ent.OrgInvite, error) {
	return r.client.OrgInvite.Query().
		Where(orginvite.IDEQ(inviteID), orginvite.OrgIDEQ(orgID)).
		Only(ctx)
}

// RevokeOrgInvite sets revoked_at on an OrgInvite.
func (r *Repository) RevokeOrgInvite(ctx context.Context, inviteID uuid.UUID, at time.Time) error {
	return r.client.OrgInvite.UpdateOneID(inviteID).
		SetRevokedAt(at).
		Exec(ctx)
}

// FindOrgInviteByCode returns an OrgInvite by its unique code.
func (r *Repository) FindOrgInviteByCode(ctx context.Context, code string) (*ent.OrgInvite, error) {
	return r.client.OrgInvite.Query().
		Where(orginvite.CodeEQ(code)).
		Only(ctx)
}

// FindActiveOrgAndParoki returns the active, non-deleted Organization and Paroki for an invite.
func (r *Repository) FindActiveOrgAndParoki(ctx context.Context, orgID, parokiID uuid.UUID) (*ent.Organization, *ent.Paroki, error) {
	orgRow, err := r.client.Organization.Query().
		Where(organization.IDEQ(orgID), organization.DeletedAtIsNil(), organization.ActiveEQ(true)).
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}
	parokiRow, err := r.client.Paroki.Query().
		Where(paroki.IDEQ(parokiID), paroki.DeletedAtIsNil(), paroki.ActiveEQ(true)).
		Only(ctx)
	if err != nil {
		return nil, nil, err
	}
	return orgRow, parokiRow, nil
}

// IncrementOrgInviteUsedCount atomically increments used_count by 1 if max_uses allows.
func (r *Repository) IncrementOrgInviteUsedCount(ctx context.Context, inviteID uuid.UUID, maxUses *int) error {
	q := r.client.OrgInvite.Update().
		Where(orginvite.IDEQ(inviteID), orginvite.RevokedAtIsNil())
	if maxUses != nil {
		q = q.Where(orginvite.UsedCountLT(*maxUses))
	}
	affected, err := q.AddUsedCount(1).Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrInviteMaxUsesReached
	}
	return nil
}
