package repo

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/agentmessage"
	"github.com/paroki/domus/ent/agentsession"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/internal/agent/domain"
)

// AppendMessage inserts a turn message into an active session and updates session activity timestamps.
func (r *Repository) AppendMessage(ctx context.Context, in domain.CreateMessageInput) (*ent.AgentMessage, error) {
	var msg *ent.AgentMessage
	err := r.WithTx(ctx, func(txRepo *Repository) error {
		if _, err := txRepo.FindSessionByID(ctx, in.ParishID, in.UserID, in.SessionID); err != nil {
			return err
		}

		cre := txRepo.client.AgentMessage.Create().
			SetParishID(in.ParishID).
			SetUserID(in.UserID).
			SetSessionID(in.SessionID).
			SetRole(agentmessage.Role(in.Role)).
			SetContent(in.Content)
		if in.AgentName != "" {
			cre.SetAgentName(in.AgentName)
		}
		if in.Thinking != nil {
			cre.SetThinking(*in.Thinking)
		}
		if len(in.ToolCalls) > 0 {
			tcs := make([]schema.AgentToolCall, len(in.ToolCalls))
			for i, tc := range in.ToolCalls {
				tcs[i] = schema.AgentToolCall{
					ID:        tc.ID,
					Name:      tc.Name,
					Arguments: tc.Arguments,
					AgentName: tc.AgentName,
				}
			}
			cre.SetToolCalls(tcs)
		}
		if in.ToolCallID != nil {
			cre.SetToolCallID(*in.ToolCallID)
		}
		if in.ToolName != nil {
			cre.SetToolName(*in.ToolName)
		}

		var err error
		msg, err = cre.Save(ctx)
		if err != nil {
			return err
		}

		return txRepo.client.AgentSession.UpdateOneID(in.SessionID).
			SetLastMessageAt(msg.CreatedAt).
			SetUpdatedAt(msg.CreatedAt).
			Exec(ctx)
	})
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// ListRecentMessages returns the latest limit messages for a session in chronological ASC order.
func (r *Repository) ListRecentMessages(ctx context.Context, parishID, sessionID uuid.UUID, limit int) ([]*ent.AgentMessage, error) {
	if limit <= 0 {
		limit = domain.HistoryWindow
	}
	rows, err := r.client.AgentMessage.Query().
		Where(
			agentmessage.ParishIDEQ(parishID),
			agentmessage.SessionIDEQ(sessionID),
			agentmessage.HasSessionWith(agentsession.DeletedAtIsNil()),
		).
		Order(agentmessage.ByCreatedAt(sql.OrderDesc()), agentmessage.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	slices.Reverse(rows)
	return rows, nil
}

// SearchSessionMessages searches messages within a single session using keyset pagination or ParadeDB BM25.
func (r *Repository) SearchSessionMessages(ctx context.Context, in domain.SearchSessionMessagesInput) ([]*ent.AgentMessage, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = domain.HistoryWindow
	}
	q := strings.TrimSpace(in.Q)

	if q != "" && r.dialect == dialect.Postgres && r.db != nil {
		rows, err := r.searchMessagesBM25Postgres(ctx, in.ParishID, in.UserID, &in.SessionID, q, limit)
		if err == nil {
			return rows, nil
		}
	}

	query := r.client.AgentMessage.Query().
		Where(
			agentmessage.ParishIDEQ(in.ParishID),
			agentmessage.UserIDEQ(in.UserID),
			agentmessage.SessionIDEQ(in.SessionID),
			agentmessage.HasSessionWith(agentsession.DeletedAtIsNil()),
		)
	if q != "" {
		query = query.Where(agentmessage.ContentContainsFold(q))
	}
	if in.CursorAt != nil && in.CursorID != nil {
		query = query.Where(agentmessage.Or(
			agentmessage.CreatedAtLT(*in.CursorAt),
			agentmessage.And(
				agentmessage.CreatedAtEQ(*in.CursorAt),
				agentmessage.IDLT(*in.CursorID),
			),
		))
	}
	return query.
		Order(agentmessage.ByCreatedAt(sql.OrderDesc()), agentmessage.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

// SearchUserHistory searches episodic conversation history across all active sessions and channels of a user.
func (r *Repository) SearchUserHistory(ctx context.Context, in domain.SearchUserHistoryInput) ([]*ent.AgentMessage, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = domain.HistoryWindow
	}
	q := strings.TrimSpace(in.Q)

	if q != "" && r.dialect == dialect.Postgres && r.db != nil {
		rows, err := r.searchMessagesBM25Postgres(ctx, in.ParishID, in.UserID, nil, q, limit)
		if err == nil {
			return rows, nil
		}
	}

	query := r.client.AgentMessage.Query().
		Where(
			agentmessage.ParishIDEQ(in.ParishID),
			agentmessage.UserIDEQ(in.UserID),
			agentmessage.HasSessionWith(agentsession.DeletedAtIsNil()),
		)
	if q != "" {
		query = query.Where(agentmessage.ContentContainsFold(q))
	}
	if in.CursorAt != nil && in.CursorID != nil {
		query = query.Where(agentmessage.Or(
			agentmessage.CreatedAtLT(*in.CursorAt),
			agentmessage.And(
				agentmessage.CreatedAtEQ(*in.CursorAt),
				agentmessage.IDLT(*in.CursorID),
			),
		))
	}
	return query.
		Order(agentmessage.ByCreatedAt(sql.OrderDesc()), agentmessage.ByID(sql.OrderDesc())).
		Limit(limit).
		All(ctx)
}

func (r *Repository) searchMessagesBM25Postgres(
	ctx context.Context,
	parishID, userID uuid.UUID,
	sessionID *uuid.UUID,
	q string,
	limit int,
) ([]*ent.AgentMessage, error) {
	args := []any{q, parishID, userID}
	sessionFilter := ""
	if sessionID != nil {
		args = append(args, *sessionID)
		sessionFilter = " AND m.session_id = $4"
		args = append(args, limit)
	} else {
		args = append(args, limit)
	}
	limitParam := "$4"
	if sessionID != nil {
		limitParam = "$5"
	}

	rawSQL := `
		SELECT m.id, m.created_at, m.updated_at, m.parish_id, m.user_id, m.session_id,
		       m.agent_name, m.role, m.content, m.thinking, m.tool_calls, m.tool_call_id, m.tool_name
		FROM agent_messages m
		JOIN agent_sessions s ON s.id = m.session_id
		WHERE m.parish_id = $2
		  AND m.user_id = $3
		  AND s.deleted_at IS NULL
		  AND m.content @@@ $1` + sessionFilter + `
		ORDER BY paradedb.score(m.id) DESC, m.created_at DESC, m.id DESC
		LIMIT ` + limitParam

	rows, err := r.db.QueryContext(ctx, rawSQL, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []*ent.AgentMessage
	for rows.Next() {
		var m ent.AgentMessage
		var role string
		var thinking, toolCallID, toolName *string
		var toolCallsRaw []byte
		if err := rows.Scan(
			&m.ID, &m.CreatedAt, &m.UpdatedAt, &m.ParishID, &m.UserID, &m.SessionID,
			&m.AgentName, &role, &m.Content, &thinking, &toolCallsRaw, &toolCallID, &toolName,
		); err != nil {
			return nil, err
		}
		m.Role = agentmessage.Role(role)
		m.Thinking = thinking
		m.ToolCallID = toolCallID
		m.ToolName = toolName
		if len(toolCallsRaw) > 0 {
			var tcs []schema.AgentToolCall
			if err := json.Unmarshal(toolCallsRaw, &tcs); err == nil {
				m.ToolCalls = tcs
			}
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}
