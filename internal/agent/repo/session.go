package repo

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentsession"
	"github.com/paroki/domus/internal/agent/domain"
)

// CreateSession inserts a new AgentSession for (parishID, userID).
func (r *Repository) CreateSession(ctx context.Context, in domain.CreateSessionInput) (*ent.AgentSession, error) {
	ch := in.Channel
	if ch == "" {
		ch = domain.ChannelWeb
	}
	cre := r.client.AgentSession.Create().
		SetParishID(in.ParishID).
		SetUserID(in.UserID).
		SetChannel(agentsession.Channel(ch))
	if title := strings.TrimSpace(in.Title); title != "" {
		cre.SetTitle(title)
	}
	return cre.Save(ctx)
}

// FindSessionByID returns a non-deleted AgentSession owned by (parishID, userID).
func (r *Repository) FindSessionByID(ctx context.Context, parishID, userID, sessionID uuid.UUID) (*ent.AgentSession, error) {
	s, err := r.client.AgentSession.Query().
		Where(
			agentsession.IDEQ(sessionID),
			agentsession.ParishIDEQ(parishID),
			agentsession.UserIDEQ(userID),
			agentsession.DeletedAtIsNil(),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, domain.ErrSessionNotFound
	}
	return s, err
}

// FindOrCreateChannelSession resumes the most recent active session for a messaging channel or creates one.
func (r *Repository) FindOrCreateChannelSession(
	ctx context.Context,
	parishID, userID uuid.UUID,
	ch domain.Channel,
	title string,
) (*ent.AgentSession, error) {
	s, err := r.client.AgentSession.Query().
		Where(
			agentsession.ParishIDEQ(parishID),
			agentsession.UserIDEQ(userID),
			agentsession.ChannelEQ(agentsession.Channel(ch)),
			agentsession.DeletedAtIsNil(),
		).
		Order(agentsession.ByLastMessageAt(sql.OrderDesc()), agentsession.ByID(sql.OrderDesc())).
		First(ctx)
	if err == nil {
		return s, nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}
	return r.CreateSession(ctx, domain.CreateSessionInput{
		ParishID: parishID,
		UserID:   userID,
		Channel:  ch,
		Title:    title,
	})
}

// SearchSessions returns keyset-paginated non-deleted sessions for (parishID, userID).
func (r *Repository) SearchSessions(ctx context.Context, in domain.SearchSessionsInput) ([]*ent.AgentSession, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = domain.HistoryWindow
	}
	query := r.client.AgentSession.Query().
		Where(
			agentsession.ParishIDEQ(in.ParishID),
			agentsession.UserIDEQ(in.UserID),
			agentsession.DeletedAtIsNil(),
		)
	if in.Channel != "" {
		query = query.Where(agentsession.ChannelEQ(agentsession.Channel(in.Channel)))
	}
	if in.CursorAt != nil && in.CursorID != nil {
		query = query.Where(agentsession.Or(
			agentsession.LastMessageAtLT(*in.CursorAt),
			agentsession.And(
				agentsession.LastMessageAtEQ(*in.CursorAt),
				agentsession.IDLT(*in.CursorID),
			),
		))
	}
	return query.
		Order(agentsession.ByLastMessageAt(sql.OrderDesc()), agentsession.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

// UpdateSessionTitle renames a non-deleted AgentSession owned by (parishID, userID).
func (r *Repository) UpdateSessionTitle(ctx context.Context, parishID, userID, sessionID uuid.UUID, title string) (*ent.AgentSession, error) {
	if _, err := r.FindSessionByID(ctx, parishID, userID, sessionID); err != nil {
		return nil, err
	}
	return r.client.AgentSession.UpdateOneID(sessionID).
		SetTitle(strings.TrimSpace(title)).
		Save(ctx)
}

// SoftDeleteSession soft-deletes a non-deleted AgentSession owned by (parishID, userID).
func (r *Repository) SoftDeleteSession(ctx context.Context, parishID, userID, sessionID uuid.UUID) error {
	if _, err := r.FindSessionByID(ctx, parishID, userID, sessionID); err != nil {
		return err
	}
	return r.client.AgentSession.UpdateOneID(sessionID).
		SetDeletedAt(time.Now().UTC()).
		Exec(ctx)
}
