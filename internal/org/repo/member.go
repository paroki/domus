package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/user"
)

// OrgMemberInput holds fields for creating an OrgMember.
type OrgMemberInput struct {
	ParokiID uuid.UUID
	OrgID    uuid.UUID
	UserID   uuid.UUID
	Role     orgmember.Role
	Status   orgmember.Status
}

// SearchOrgMembers returns paginated OrgMember rows along with a User lookup map.
func (r *Repository) SearchOrgMembers(
	ctx context.Context,
	orgID uuid.UUID,
	q, role, status string,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.OrgMember, map[uuid.UUID]*ent.User, error) {
	query := r.client.OrgMember.Query().Where(orgmember.OrgIDEQ(orgID))
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
			return []*ent.OrgMember{}, map[uuid.UUID]*ent.User{}, nil
		}
		query = query.Where(orgmember.UserIDIn(userIDs...))
	}
	if role != "" {
		query = query.Where(orgmember.RoleEQ(orgmember.Role(role)))
	}
	if status != "" {
		query = query.Where(orgmember.StatusEQ(orgmember.Status(status)))
	}
	if cursorCreatedAt != nil && cursorID != nil {
		query = query.Where(orgmember.Or(
			orgmember.CreatedAtLT(*cursorCreatedAt),
			orgmember.And(orgmember.CreatedAtEQ(*cursorCreatedAt), orgmember.IDLT(*cursorID)),
		))
	}

	rows, err := query.
		Order(orgmember.ByCreatedAt(sql.OrderDesc()), orgmember.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	usersMap := make(map[uuid.UUID]*ent.User)
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

// FindOrgMember returns the OrgMember row for (orgID, userID).
func (r *Repository) FindOrgMember(ctx context.Context, orgID, userID uuid.UUID) (*ent.OrgMember, error) {
	return r.client.OrgMember.Query().
		Where(orgmember.OrgIDEQ(orgID), orgmember.UserIDEQ(userID)).
		Only(ctx)
}

// FindActiveOrgMember returns the active OrgMember row for (orgID, userID).
func (r *Repository) FindActiveOrgMember(ctx context.Context, orgID, userID uuid.UUID) (*ent.OrgMember, error) {
	return r.client.OrgMember.Query().
		Where(
			orgmember.OrgIDEQ(orgID),
			orgmember.UserIDEQ(userID),
			orgmember.StatusEQ(orgmember.StatusActive),
		).
		Only(ctx)
}

// ListActiveOrgRolesForUser returns a map of orgID -> role string for active memberships of userID in orgIDs.
func (r *Repository) ListActiveOrgRolesForUser(ctx context.Context, userID uuid.UUID, orgIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	roles := make(map[uuid.UUID]string)
	if userID == uuid.Nil || len(orgIDs) == 0 {
		return roles, nil
	}
	rows, err := r.client.OrgMember.Query().
		Where(
			orgmember.UserIDEQ(userID),
			orgmember.OrgIDIn(orgIDs...),
			orgmember.StatusEQ(orgmember.StatusActive),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		roles[row.OrgID] = string(row.Role)
	}
	return roles, nil
}

// CreateOrgMember inserts a new OrgMember row.
func (r *Repository) CreateOrgMember(ctx context.Context, in OrgMemberInput) (*ent.OrgMember, error) {
	return r.client.OrgMember.Create().
		SetParokiID(in.ParokiID).
		SetOrgID(in.OrgID).
		SetUserID(in.UserID).
		SetRole(in.Role).
		SetStatus(in.Status).
		Save(ctx)
}

// UpdateOrgMember updates mutable fields of an OrgMember row.
func (r *Repository) UpdateOrgMember(
	ctx context.Context,
	id uuid.UUID,
	role *orgmember.Role,
	status *orgmember.Status,
) (*ent.OrgMember, error) {
	upd := r.client.OrgMember.UpdateOneID(id)
	if role != nil {
		upd.SetRole(*role)
	}
	if status != nil {
		upd.SetStatus(*status)
	}
	return upd.Save(ctx)
}

// DeleteOrgMember hard-deletes an OrgMember row by ID.
func (r *Repository) DeleteOrgMember(ctx context.Context, id uuid.UUID) error {
	return r.client.OrgMember.DeleteOneID(id).Exec(ctx)
}

// CountActiveOrgAdmins counts active admin members in orgID.
func (r *Repository) CountActiveOrgAdmins(ctx context.Context, orgID uuid.UUID) (int, error) {
	return r.client.OrgMember.Query().
		Where(
			orgmember.OrgIDEQ(orgID),
			orgmember.RoleEQ(orgmember.RoleAdmin),
			orgmember.StatusEQ(orgmember.StatusActive),
		).
		Count(ctx)
}
