package repo

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/eventrsvp"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/ent/user"
)

// RSVPSummaryCounts holds raw RSVP and active org member counts for an event.
type RSVPSummaryCounts struct {
	Going              int
	Maybe              int
	NotGoing           int
	TotalActiveMembers int
}

// UpsertRSVP inserts or updates the caller's RSVP row for eventID.
func (r *Repository) UpsertRSVP(
	ctx context.Context,
	parokiID, eventID, userID uuid.UUID,
	status eventrsvp.Status,
	note *string,
) (*ent.EventRSVP, error) {
	existing, err := r.client.EventRSVP.Query().
		Where(eventrsvp.EventIDEQ(eventID), eventrsvp.UserIDEQ(userID)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if ent.IsNotFound(err) {
		q := r.client.EventRSVP.Create().
			SetParokiID(parokiID).
			SetEventID(eventID).
			SetUserID(userID).
			SetStatus(status)
		if note != nil && *note != "" {
			q.SetNote(*note)
		}
		return q.Save(ctx)
	}

	q := r.client.EventRSVP.UpdateOneID(existing.ID).
		SetStatus(status)
	if note == nil || *note == "" {
		q.ClearNote()
	} else {
		q.SetNote(*note)
	}
	return q.Save(ctx)
}

// DeleteRSVP hard-deletes the caller's RSVP row for eventID (idempotent).
func (r *Repository) DeleteRSVP(ctx context.Context, eventID, userID uuid.UUID) error {
	_, err := r.client.EventRSVP.Delete().
		Where(eventrsvp.EventIDEQ(eventID), eventrsvp.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// FindUserRSVP returns the caller's RSVP for eventID, if any.
func (r *Repository) FindUserRSVP(ctx context.Context, eventID, userID uuid.UUID) (*ent.EventRSVP, error) {
	return r.client.EventRSVP.Query().
		Where(eventrsvp.EventIDEQ(eventID), eventrsvp.UserIDEQ(userID)).
		Only(ctx)
}

// FindUserRSVPsForEvents returns a map of eventID -> EventRSVP for userID across eventIDs.
func (r *Repository) FindUserRSVPsForEvents(ctx context.Context, eventIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]*ent.EventRSVP, error) {
	out := make(map[uuid.UUID]*ent.EventRSVP)
	if len(eventIDs) == 0 {
		return out, nil
	}
	rows, err := r.client.EventRSVP.Query().
		Where(eventrsvp.EventIDIn(eventIDs...), eventrsvp.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.EventID] = row
	}
	return out, nil
}

// IsParokiMember reports whether userID has a ParokiMember row in parokiID.
func (r *Repository) IsParokiMember(ctx context.Context, parokiID, userID uuid.UUID) (bool, error) {
	return r.client.ParokiMember.Query().
		Where(parokimember.ParokiIDEQ(parokiID), parokimember.UserIDEQ(userID)).
		Exist(ctx)
}

// IsParokiAdmin reports whether userID has an admin ParokiMember row in parokiID.
func (r *Repository) IsParokiAdmin(ctx context.Context, parokiID, userID uuid.UUID) (bool, error) {
	return r.client.ParokiMember.Query().
		Where(
			parokimember.ParokiIDEQ(parokiID),
			parokimember.UserIDEQ(userID),
			parokimember.RoleEQ(parokimember.RoleAdmin),
		).
		Exist(ctx)
}

// CountActiveParokiMembers returns the number of ParokiMember rows in parokiID whose User is not deleted.
func (r *Repository) CountActiveParokiMembers(ctx context.Context, parokiID uuid.UUID) (int, error) {
	var userIDs []uuid.UUID
	err := r.client.ParokiMember.Query().
		Where(parokimember.ParokiIDEQ(parokiID)).
		Select(parokimember.FieldUserID).
		Scan(ctx, &userIDs)
	if err != nil {
		return 0, err
	}
	if len(userIDs) == 0 {
		return 0, nil
	}
	return r.client.User.Query().
		Where(user.IDIn(userIDs...), user.DeletedAtIsNil()).
		Count(ctx)
}

// ListAttendeeUserIDsForEvent returns user IDs with going or maybe RSVP status for eventID.
func (r *Repository) ListAttendeeUserIDsForEvent(ctx context.Context, eventID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.client.EventRSVP.Query().
		Where(
			eventrsvp.EventIDEQ(eventID),
			eventrsvp.StatusIn(eventrsvp.StatusGoing, eventrsvp.StatusMaybe),
		).
		Select(eventrsvp.FieldUserID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	return ids, nil
}

// GetRSVPSummary returns RSVP counts by status and active member count for an event.
func (r *Repository) GetRSVPSummary(ctx context.Context, parokiID, eventID uuid.UUID, orgID *uuid.UUID) (RSVPSummaryCounts, error) {
	var counts RSVPSummaryCounts
	rows, err := r.client.EventRSVP.Query().
		Where(eventrsvp.EventIDEQ(eventID)).
		All(ctx)
	if err != nil {
		return counts, err
	}
	for _, row := range rows {
		switch row.Status {
		case eventrsvp.StatusGoing:
			counts.Going++
		case eventrsvp.StatusMaybe:
			counts.Maybe++
		case eventrsvp.StatusNotGoing:
			counts.NotGoing++
		}
	}
	if orgID != nil {
		activeCount, err := r.client.OrgMember.Query().
			Where(orgmember.OrgIDEQ(*orgID), orgmember.StatusEQ(orgmember.StatusActive)).
			Count(ctx)
		if err != nil {
			return counts, err
		}
		counts.TotalActiveMembers = activeCount
	} else {
		activeCount, err := r.CountActiveParokiMembers(ctx, parokiID)
		if err != nil {
			return counts, err
		}
		counts.TotalActiveMembers = activeCount
	}
	return counts, nil
}

// GetRSVPSummariesForEvents returns RSVP summaries keyed by eventID for a map of eventID -> optional orgID.
func (r *Repository) GetRSVPSummariesForEvents(ctx context.Context, parokiID uuid.UUID, eventOrgMap map[uuid.UUID]*uuid.UUID) (map[uuid.UUID]RSVPSummaryCounts, error) {
	out := make(map[uuid.UUID]RSVPSummaryCounts, len(eventOrgMap))
	if len(eventOrgMap) == 0 {
		return out, nil
	}

	eventIDs := make([]uuid.UUID, 0, len(eventOrgMap))
	orgIDSet := make(map[uuid.UUID]struct{})
	hasParokiEvent := false
	for evID, orgID := range eventOrgMap {
		eventIDs = append(eventIDs, evID)
		if orgID != nil {
			orgIDSet[*orgID] = struct{}{}
		} else {
			hasParokiEvent = true
		}
	}

	rows, err := r.client.EventRSVP.Query().
		Where(eventrsvp.EventIDIn(eventIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		c := out[row.EventID]
		switch row.Status {
		case eventrsvp.StatusGoing:
			c.Going++
		case eventrsvp.StatusMaybe:
			c.Maybe++
		case eventrsvp.StatusNotGoing:
			c.NotGoing++
		}
		out[row.EventID] = c
	}

	orgMemberCounts := make(map[uuid.UUID]int, len(orgIDSet))
	if len(orgIDSet) > 0 {
		orgIDs := make([]uuid.UUID, 0, len(orgIDSet))
		for orgID := range orgIDSet {
			orgIDs = append(orgIDs, orgID)
		}
		activeMembers, err := r.client.OrgMember.Query().
			Where(orgmember.OrgIDIn(orgIDs...), orgmember.StatusEQ(orgmember.StatusActive)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		for _, m := range activeMembers {
			orgMemberCounts[m.OrgID]++
		}
	}

	var parokiActiveCount int
	if hasParokiEvent {
		parokiActiveCount, err = r.CountActiveParokiMembers(ctx, parokiID)
		if err != nil {
			return nil, err
		}
	}

	for evID, orgID := range eventOrgMap {
		c := out[evID]
		if orgID != nil {
			c.TotalActiveMembers = orgMemberCounts[*orgID]
		} else {
			c.TotalActiveMembers = parokiActiveCount
		}
		out[evID] = c
	}
	return out, nil
}

// SearchRSVPs returns paginated EventRSVP rows for eventID along with User and OrgMember lookup maps.
func (r *Repository) SearchRSVPs(
	ctx context.Context,
	eventID uuid.UUID,
	orgID *uuid.UUID,
	q, status string,
	cursorAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]*ent.EventRSVP, map[uuid.UUID]*ent.User, map[uuid.UUID]*ent.OrgMember, error) {
	query := r.client.EventRSVP.Query().Where(eventrsvp.EventIDEQ(eventID))
	if q != "" {
		userIDs, err := r.client.User.Query().
			Where(user.DeletedAtIsNil(), user.Or(
				user.NameContainsFold(q),
				user.EmailContainsFold(q),
			)).
			IDs(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		if len(userIDs) > 0 {
			query = query.Where(eventrsvp.Or(
				eventrsvp.UserIDIn(userIDs...),
				eventrsvp.NoteContainsFold(q),
			))
		} else {
			query = query.Where(eventrsvp.NoteContainsFold(q))
		}
	}
	if status != "" {
		query = query.Where(eventrsvp.StatusEQ(eventrsvp.Status(status)))
	}
	if cursorAt != nil && cursorID != nil {
		query = query.Where(eventrsvp.Or(
			eventrsvp.CreatedAtLT(*cursorAt),
			eventrsvp.And(eventrsvp.CreatedAtEQ(*cursorAt), eventrsvp.IDLT(*cursorID)),
		))
	}

	rows, err := query.
		Order(eventrsvp.ByCreatedAt(sql.OrderDesc()), eventrsvp.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	if orgID == nil {
		usersMap, err := r.LoadUsers(ctx, ids)
		if err != nil {
			return nil, nil, nil, err
		}
		return rows, usersMap, map[uuid.UUID]*ent.OrgMember{}, nil
	}
	usersMap, membersMap, err := r.LoadUsersAndOrgMembers(ctx, *orgID, ids)
	if err != nil {
		return nil, nil, nil, err
	}
	return rows, usersMap, membersMap, nil
}

// LoadUsers fetches User rows keyed by userID.
func (r *Repository) LoadUsers(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*ent.User, error) {
	usersMap := make(map[uuid.UUID]*ent.User)
	if len(userIDs) == 0 {
		return usersMap, nil
	}
	users, err := r.client.User.Query().Where(user.IDIn(userIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		usersMap[u.ID] = u
	}
	return usersMap, nil
}

// LoadUsersAndOrgMembers fetches User and OrgMember rows keyed by userID for the given orgID.
func (r *Repository) LoadUsersAndOrgMembers(
	ctx context.Context,
	orgID uuid.UUID,
	userIDs []uuid.UUID,
) (map[uuid.UUID]*ent.User, map[uuid.UUID]*ent.OrgMember, error) {
	usersMap, err := r.LoadUsers(ctx, userIDs)
	if err != nil {
		return nil, nil, err
	}
	membersMap := make(map[uuid.UUID]*ent.OrgMember)
	if len(userIDs) == 0 {
		return usersMap, membersMap, nil
	}

	members, err := r.client.OrgMember.Query().
		Where(orgmember.OrgIDEQ(orgID), orgmember.UserIDIn(userIDs...)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, m := range members {
		membersMap[m.UserID] = m
	}

	return usersMap, membersMap, nil
}
