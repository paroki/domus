package repo

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentmessage"
	"github.com/paroki/domus/ent/agentsession"
	"github.com/paroki/domus/internal/agent/domain"
)

// CreateSession creates a new AgentSession scoped to (parishID, userID).
func (r *Repository) CreateSession(
	ctx context.Context,
	parishID, userID uuid.UUID,
	title string,
	providerID *uuid.UUID,
	modelID *string,
) (domain.SessionResponse, error) {
	q := r.client.AgentSession.Create().
		SetParishID(parishID).
		SetUserID(userID).
		SetLastMessageAt(time.Now().UTC())
	if trimmed := strings.TrimSpace(title); trimmed != "" {
		q.SetTitle(trimmed)
	}
	if providerID != nil {
		q.SetProviderID(*providerID)
	}
	if modelID != nil && strings.TrimSpace(*modelID) != "" {
		q.SetModelID(strings.TrimSpace(*modelID))
	}
	row, err := q.Save(ctx)
	if err != nil {
		return domain.SessionResponse{}, err
	}
	return toSessionResponse(row), nil
}

// GetSession returns an active AgentSession scoped to (parishID, userID).
func (r *Repository) GetSession(ctx context.Context, parishID, userID, sessionID uuid.UUID) (domain.SessionResponse, error) {
	row, err := r.client.AgentSession.Query().
		Where(
			agentsession.IDEQ(sessionID),
			agentsession.ParishIDEQ(parishID),
			agentsession.UserIDEQ(userID),
			agentsession.DeletedAtIsNil(),
		).
		Only(ctx)
	if ent.IsNotFound(err) {
		return domain.SessionResponse{}, domain.ErrSessionNotFound
	}
	if err != nil {
		return domain.SessionResponse{}, err
	}
	return toSessionResponse(row), nil
}

// UpdateSessionTitle updates the title of an active AgentSession scoped to (parishID, userID).
func (r *Repository) UpdateSessionTitle(ctx context.Context, parishID, userID, sessionID uuid.UUID, title string) (domain.SessionResponse, error) {
	if _, err := r.GetSession(ctx, parishID, userID, sessionID); err != nil {
		return domain.SessionResponse{}, err
	}
	row, err := r.client.AgentSession.UpdateOneID(sessionID).
		SetTitle(strings.TrimSpace(title)).
		Save(ctx)
	if err != nil {
		return domain.SessionResponse{}, err
	}
	return toSessionResponse(row), nil
}

// SoftDeleteSession soft-deletes an active AgentSession scoped to (parishID, userID).
func (r *Repository) SoftDeleteSession(ctx context.Context, parishID, userID, sessionID uuid.UUID) error {
	affected, err := r.client.AgentSession.Update().
		Where(
			agentsession.IDEQ(sessionID),
			agentsession.ParishIDEQ(parishID),
			agentsession.UserIDEQ(userID),
			agentsession.DeletedAtIsNil(),
		).
		SetDeletedAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

// SearchSessions returns keyset-paginated active sessions for (parishID, userID) ordered by last_message_at DESC, id DESC.
func (r *Repository) SearchSessions(
	ctx context.Context,
	parishID, userID uuid.UUID,
	cursorLastMessageAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.SessionResponse, bool, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	q := r.client.AgentSession.Query().
		Where(
			agentsession.ParishIDEQ(parishID),
			agentsession.UserIDEQ(userID),
			agentsession.DeletedAtIsNil(),
		)
	if cursorLastMessageAt != nil && cursorID != nil {
		q = q.Where(agentsession.Or(
			agentsession.LastMessageAtLT(*cursorLastMessageAt),
			agentsession.And(
				agentsession.LastMessageAtEQ(*cursorLastMessageAt),
				agentsession.IDLT(*cursorID),
			),
		))
	}

	rows, err := q.
		Order(ent.Desc(agentsession.FieldLastMessageAt), ent.Desc(agentsession.FieldID)).
		Limit(limit + 1).
		All(ctx)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	out := make([]domain.SessionResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toSessionResponse(row))
	}
	return out, hasMore, nil
}

// AppendMessages appends a batch of messages to an active session scoped to (parishID, userID)
// and updates session.last_message_at.
func (r *Repository) AppendMessages(
	ctx context.Context,
	parishID, userID, sessionID uuid.UUID,
	msgs []domain.AppendMessageInput,
) ([]domain.MessageResponse, error) {
	if _, err := r.GetSession(ctx, parishID, userID, sessionID); err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	builders := make([]*ent.AgentMessageCreate, 0, len(msgs))
	for _, m := range msgs {
		b := tx.AgentMessage.Create().
			SetSessionID(sessionID).
			SetParishID(parishID).
			SetUserID(userID).
			SetRole(agentmessage.Role(m.Role)).
			SetContent(m.Content)
		if strings.TrimSpace(m.AgentName) != "" {
			b.SetAgentName(strings.TrimSpace(m.AgentName))
		}
		if len(m.ToolCalls) > 0 {
			b.SetToolCalls(m.ToolCalls)
		}
		if m.ToolCallID != nil {
			b.SetToolCallID(*m.ToolCallID)
		}
		if m.ToolName != nil {
			b.SetToolName(*m.ToolName)
		}
		if m.Meta != nil {
			b.SetMeta(m.Meta)
		}
		builders = append(builders, b)
	}

	rows, err := tx.AgentMessage.CreateBulk(builders...).Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.AgentSession.UpdateOneID(sessionID).SetLastMessageAt(time.Now().UTC()).Exec(ctx); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	out := make([]domain.MessageResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toMessageResponse(row))
	}
	return out, nil
}

// ListMessages lists messages in chronological order for an active session scoped to (parishID, userID).
func (r *Repository) ListMessages(ctx context.Context, parishID, userID, sessionID uuid.UUID, limit int) ([]domain.MessageResponse, error) {
	if _, err := r.GetSession(ctx, parishID, userID, sessionID); err != nil {
		return nil, err
	}
	q := r.client.AgentMessage.Query().
		Where(
			agentmessage.SessionIDEQ(sessionID),
			agentmessage.ParishIDEQ(parishID),
			agentmessage.UserIDEQ(userID),
		).
		Order(ent.Asc(agentmessage.FieldCreatedAt), ent.Asc(agentmessage.FieldID))
	if limit > 0 {
		q = q.Limit(limit)
	}
	rows, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MessageResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toMessageResponse(row))
	}
	return out, nil
}

func toSessionResponse(row *ent.AgentSession) domain.SessionResponse {
	return domain.SessionResponse{
		ID:            row.ID,
		ParishID:      row.ParishID,
		UserID:        row.UserID,
		Title:         row.Title,
		ProviderID:    row.ProviderID,
		ModelID:       row.ModelID,
		LastMessageAt: row.LastMessageAt,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func toMessageResponse(row *ent.AgentMessage) domain.MessageResponse {
	return domain.MessageResponse{
		ID:         row.ID,
		SessionID:  row.SessionID,
		ParishID:   row.ParishID,
		UserID:     row.UserID,
		AgentName:  row.AgentName,
		Role:       string(row.Role),
		Content:    row.Content,
		ToolCalls:  row.ToolCalls,
		ToolCallID: row.ToolCallID,
		ToolName:   row.ToolName,
		Meta:       row.Meta,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
