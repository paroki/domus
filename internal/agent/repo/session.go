package repo

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentmessage"
	"github.com/paroki/domus/ent/agentsession"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/internal/agent/domain"
)

// CreateSession creates a new AgentSession scoped to (parokiID, userID).
func (r *Repository) CreateSession(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	title string,
	providerID *uuid.UUID,
	modelID *string,
) (domain.SessionResponse, error) {
	q := r.client.AgentSession.Create().
		SetParokiID(parokiID).
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

// GetSession returns an active AgentSession scoped to (parokiID, userID).
func (r *Repository) GetSession(ctx context.Context, parokiID, userID, sessionID uuid.UUID) (domain.SessionResponse, error) {
	row, err := r.client.AgentSession.Query().
		Where(
			agentsession.IDEQ(sessionID),
			agentsession.ParokiIDEQ(parokiID),
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

// UpdateSession updates title and/or provider/model override on an active AgentSession scoped to (parokiID, userID).
func (r *Repository) UpdateSession(
	ctx context.Context,
	parokiID, userID, sessionID uuid.UUID,
	title *string,
	clearModel bool,
	providerID *uuid.UUID,
	modelID *string,
) (domain.SessionResponse, error) {
	if _, err := r.GetSession(ctx, parokiID, userID, sessionID); err != nil {
		return domain.SessionResponse{}, err
	}
	upd := r.client.AgentSession.UpdateOneID(sessionID)
	if title != nil {
		upd.SetTitle(strings.TrimSpace(*title))
	}
	if clearModel {
		upd.ClearProviderID().ClearModelID()
	} else {
		if providerID != nil {
			upd.SetProviderID(*providerID)
		}
		if modelID != nil {
			upd.SetModelID(strings.TrimSpace(*modelID))
		}
	}
	row, err := upd.Save(ctx)
	if err != nil {
		return domain.SessionResponse{}, err
	}
	return toSessionResponse(row), nil
}

// SoftDeleteSession soft-deletes an active AgentSession scoped to (parokiID, userID).
func (r *Repository) SoftDeleteSession(ctx context.Context, parokiID, userID, sessionID uuid.UUID) error {
	affected, err := r.client.AgentSession.Update().
		Where(
			agentsession.IDEQ(sessionID),
			agentsession.ParokiIDEQ(parokiID),
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

// SearchSessions returns keyset-paginated active sessions for (parokiID, userID) ordered by last_message_at DESC, id DESC.
func (r *Repository) SearchSessions(
	ctx context.Context,
	parokiID, userID uuid.UUID,
	q string,
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

	query := r.client.AgentSession.Query().
		Where(
			agentsession.ParokiIDEQ(parokiID),
			agentsession.UserIDEQ(userID),
			agentsession.DeletedAtIsNil(),
		)
	if trimmedQ := strings.TrimSpace(q); trimmedQ != "" {
		query = query.Where(agentsession.TitleContainsFold(trimmedQ))
	}
	if cursorLastMessageAt != nil && cursorID != nil {
		query = query.Where(agentsession.Or(
			agentsession.LastMessageAtLT(*cursorLastMessageAt),
			agentsession.And(
				agentsession.LastMessageAtEQ(*cursorLastMessageAt),
				agentsession.IDLT(*cursorID),
			),
		))
	}

	rows, err := query.
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

// AppendMessages appends a batch of messages to an active session scoped to (parokiID, userID)
// and updates session.last_message_at.
func (r *Repository) AppendMessages(
	ctx context.Context,
	parokiID, userID, sessionID uuid.UUID,
	msgs []domain.AppendMessageInput,
) ([]domain.MessageResponse, error) {
	if _, err := r.GetSession(ctx, parokiID, userID, sessionID); err != nil {
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
			SetParokiID(parokiID).
			SetUserID(userID).
			SetRole(agentmessage.Role(m.Role)).
			SetContent(m.Content)
		if strings.TrimSpace(m.AgentName) != "" {
			b.SetAgentName(strings.TrimSpace(m.AgentName))
		}
		if len(m.ToolCalls) > 0 {
			tc := make([]schema.ToolCallRecord, len(m.ToolCalls))
			for i, t := range m.ToolCalls {
				tc[i] = schema.ToolCallRecord(t)
			}
			b.SetToolCalls(tc)
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

// ListMessages lists messages in chronological order for an active session scoped to (parokiID, userID).
func (r *Repository) ListMessages(ctx context.Context, parokiID, userID, sessionID uuid.UUID, limit int) ([]domain.MessageResponse, error) {
	if _, err := r.GetSession(ctx, parokiID, userID, sessionID); err != nil {
		return nil, err
	}
	q := r.client.AgentMessage.Query().
		Where(
			agentmessage.SessionIDEQ(sessionID),
			agentmessage.ParokiIDEQ(parokiID),
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
		ParokiID:      row.ParokiID,
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
	var toolCalls []domain.ToolCallRecord
	if len(row.ToolCalls) > 0 {
		toolCalls = make([]domain.ToolCallRecord, len(row.ToolCalls))
		for i, t := range row.ToolCalls {
			toolCalls[i] = domain.ToolCallRecord(t)
		}
	}
	return domain.MessageResponse{
		ID:         row.ID,
		SessionID:  row.SessionID,
		ParokiID:   row.ParokiID,
		UserID:     row.UserID,
		AgentName:  row.AgentName,
		Role:       string(row.Role),
		Content:    row.Content,
		ToolCalls:  toolCalls,
		ToolCallID: row.ToolCallID,
		ToolName:   row.ToolName,
		Meta:       row.Meta,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
