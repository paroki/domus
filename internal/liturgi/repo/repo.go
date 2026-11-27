package repo

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/liturgilagu"
	"github.com/paroki/domus/ent/liturgiministry"
	"github.com/paroki/domus/ent/liturgipetugas"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/ent/predicate"
	"github.com/paroki/domus/ent/user"
)

// Repository encapsulates Ent persistence for the liturgi feature.
type Repository struct {
	client *ent.Client
}

// New constructs a liturgi Repository.
func New(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// ListMinistries returns all LiturgiMinistry rows for parokiID.
func (r *Repository) ListMinistries(ctx context.Context, parokiID uuid.UUID) ([]*ent.LiturgiMinistry, error) {
	return r.client.LiturgiMinistry.Query().
		Where(liturgiministry.ParokiIDEQ(parokiID)).
		All(ctx)
}

// FindMinistry returns the LiturgiMinistry row for (parokiID, roleSlug), if any.
func (r *Repository) FindMinistry(ctx context.Context, parokiID uuid.UUID, roleSlug liturgiministry.RoleSlug) (*ent.LiturgiMinistry, error) {
	return r.client.LiturgiMinistry.Query().
		Where(
			liturgiministry.ParokiIDEQ(parokiID),
			liturgiministry.RoleSlugEQ(roleSlug),
		).
		Only(ctx)
}

// UpsertMinistry creates or updates a LiturgiMinistry row for (parokiID, roleSlug).
func (r *Repository) UpsertMinistry(
	ctx context.Context,
	parokiID uuid.UUID,
	roleSlug liturgiministry.RoleSlug,
	orgID *uuid.UUID,
	clearOrg bool,
	leaderUserID *uuid.UUID,
	clearLeader bool,
) (*ent.LiturgiMinistry, error) {
	existing, err := r.FindMinistry(ctx, parokiID, roleSlug)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if existing == nil {
		q := r.client.LiturgiMinistry.Create().
			SetParokiID(parokiID).
			SetRoleSlug(roleSlug)
		if orgID != nil && !clearOrg {
			q.SetOrgID(*orgID)
		}
		if leaderUserID != nil && !clearLeader {
			q.SetLeaderUserID(*leaderUserID)
		}
		return q.Save(ctx)
	}

	upd := r.client.LiturgiMinistry.UpdateOneID(existing.ID)
	if clearOrg {
		upd.ClearOrgID()
	} else if orgID != nil {
		upd.SetOrgID(*orgID)
	}
	if clearLeader {
		upd.ClearLeaderUserID()
	} else if leaderUserID != nil {
		upd.SetLeaderUserID(*leaderUserID)
	}
	return upd.Save(ctx)
}

// LoadUsers returns a map of User rows by ID for the given userIDs.
func (r *Repository) LoadUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*ent.User, error) {
	out := make(map[uuid.UUID]*ent.User)
	if len(userIDs) == 0 {
		return out, nil
	}
	rows, err := r.client.User.Query().
		Where(user.IDIn(userIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, u := range rows {
		out[u.ID] = u
	}
	return out, nil
}

// LoadOrgs returns a map of Organization rows by ID for the given orgIDs.
func (r *Repository) LoadOrgs(ctx context.Context, orgIDs []uuid.UUID) (map[uuid.UUID]*ent.Organization, error) {
	out := make(map[uuid.UUID]*ent.Organization)
	if len(orgIDs) == 0 {
		return out, nil
	}
	rows, err := r.client.Organization.Query().
		Where(organization.IDIn(orgIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, o := range rows {
		out[o.ID] = o
	}
	return out, nil
}

// IsActiveUserInParoki reports whether userID is non-deleted and a member of parokiID.
func (r *Repository) IsActiveUserInParoki(ctx context.Context, parokiID, userID uuid.UUID) (bool, error) {
	userExists, err := r.client.User.Query().
		Where(user.IDEQ(userID), user.DeletedAtIsNil()).
		Exist(ctx)
	if err != nil || !userExists {
		return false, err
	}
	return r.client.ParokiMember.Query().
		Where(parokimember.ParokiIDEQ(parokiID), parokimember.UserIDEQ(userID)).
		Exist(ctx)
}

// PetugasEventFilter filters event IDs via liturgi_petugas rows.
type PetugasEventFilter struct {
	ParokiID       uuid.UUID
	Role           string
	AssignedToMe   bool
	AssignedUserID uuid.UUID
	MemberOrgIDs   []uuid.UUID
	AssignedOrgID  *uuid.UUID
	HasDeclined    *bool
}

// ListMatchingEventIDsForPetugas returns distinct event IDs matching the given petugas filters.
func (r *Repository) ListMatchingEventIDsForPetugas(ctx context.Context, in PetugasEventFilter) ([]uuid.UUID, error) {
	preds := []predicate.LiturgiPetugas{
		liturgipetugas.ParokiIDEQ(in.ParokiID),
	}
	if in.Role != "" {
		preds = append(preds, liturgipetugas.RoleEQ(liturgipetugas.Role(in.Role)))
	}
	if in.AssignedToMe {
		mePreds := []predicate.LiturgiPetugas{
			liturgipetugas.UserIDEQ(in.AssignedUserID),
			liturgipetugas.SubstituteUserIDEQ(in.AssignedUserID),
		}
		if len(in.MemberOrgIDs) > 0 {
			mePreds = append(mePreds,
				liturgipetugas.OrgIDIn(in.MemberOrgIDs...),
				liturgipetugas.SubstituteOrgIDIn(in.MemberOrgIDs...),
			)
		}
		preds = append(preds, liturgipetugas.Or(mePreds...))
	}
	if in.AssignedOrgID != nil {
		preds = append(preds, liturgipetugas.Or(
			liturgipetugas.OrgIDEQ(*in.AssignedOrgID),
			liturgipetugas.SubstituteOrgIDEQ(*in.AssignedOrgID),
		))
	}
	if in.HasDeclined != nil && *in.HasDeclined {
		preds = append(preds, liturgipetugas.StatusEQ(liturgipetugas.StatusDeclined))
	}

	rows, err := r.client.LiturgiPetugas.Query().
		Where(preds...).
		Select(liturgipetugas.FieldEventID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[uuid.UUID]struct{}, len(rows))
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.EventID]; !ok {
			seen[row.EventID] = struct{}{}
			ids = append(ids, row.EventID)
		}
	}
	return ids, nil
}

// ListPetugasForEvents returns all LiturgiPetugas rows for eventIDs ordered by role ASC, urutan ASC, created_at ASC.
func (r *Repository) ListPetugasForEvents(ctx context.Context, eventIDs []uuid.UUID) ([]*ent.LiturgiPetugas, error) {
	if len(eventIDs) == 0 {
		return []*ent.LiturgiPetugas{}, nil
	}
	return r.client.LiturgiPetugas.Query().
		Where(liturgipetugas.EventIDIn(eventIDs...)).
		Order(
			liturgipetugas.ByRole(sql.OrderAsc()),
			liturgipetugas.ByUrutan(sql.OrderAsc()),
			liturgipetugas.ByCreatedAt(sql.OrderAsc()),
		).
		All(ctx)
}

// ListLaguForEvents returns all LiturgiLagu rows for eventIDs ordered by urutan ASC, created_at ASC.
func (r *Repository) ListLaguForEvents(ctx context.Context, eventIDs []uuid.UUID) ([]*ent.LiturgiLagu, error) {
	if len(eventIDs) == 0 {
		return []*ent.LiturgiLagu{}, nil
	}
	return r.client.LiturgiLagu.Query().
		Where(liturgilagu.EventIDIn(eventIDs...)).
		Order(
			liturgilagu.ByUrutan(sql.OrderAsc()),
			liturgilagu.ByCreatedAt(sql.OrderAsc()),
		).
		All(ctx)
}
