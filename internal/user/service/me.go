package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// GetMe returns the authenticated user's profile with active parish and global roles.
func (s *Service) GetMe(ctx context.Context, p authx.Principal) (domain.UserResponse, error) {
	u, err := guardActiveUser(s.repo.FindUserByID(ctx, p.UserID))
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.enrichMeResponse(ctx, u, p)
}

// UpdateMe updates the authenticated user's mutable profile fields.
func (s *Service) UpdateMe(ctx context.Context, p authx.Principal, req domain.UpdateMeRequest) (domain.UserResponse, error) {
	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.UserResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}

	var phonePtr *string
	if req.Phone != nil {
		norm, err := normalizePhone(*req.Phone)
		if err != nil {
			return domain.UserResponse{}, err
		}
		phonePtr = &norm
	}

	var avatarPtr *string
	if req.AvatarURL != nil {
		trimmed := strings.TrimSpace(*req.AvatarURL)
		avatarPtr = &trimmed
	}

	u, err := s.repo.UpdateUser(ctx, p.UserID, namePtr, phonePtr, avatarPtr, req.DefaultParishID, nil, req.Onboarded)
	if ent.IsNotFound(err) {
		return domain.UserResponse{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.enrichMeResponse(ctx, u, p)
}

func (s *Service) enrichMeResponse(ctx context.Context, u *ent.User, p authx.Principal) (domain.UserResponse, error) {
	roles, err := s.loadRoles(ctx, u.ID, p.ParishID)
	if err != nil {
		return domain.UserResponse{}, err
	}
	providers, err := s.repo.ListAccountProvidersByUserID(ctx, u.ID)
	if err != nil {
		return domain.UserResponse{}, err
	}
	res := toUserResponse(u)
	res.Roles = roles
	res.RoleDetails = s.describeRoles(roles)
	res.LinkedProviders = providers
	if p.ParishID != uuid.Nil {
		pid := p.ParishID
		res.ActiveParishID = &pid
	}
	return res, nil
}

// SwitchSessionParish updates active_parish_id on the current session and initializes User.default_parish_id if nil.
func (s *Service) SwitchSessionParish(ctx context.Context, p authx.Principal, parishID uuid.UUID) error {
	if p.SessionID == uuid.Nil {
		return domain.ErrSessionNotFound
	}
	if err := s.guardParishAccess(ctx, p.UserID, parishID); err != nil {
		return err
	}
	return s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if err := txRepo.UpdateSessionActiveParish(ctx, p.SessionID, p.UserID, parishID); err != nil {
			if ent.IsNotFound(err) {
				return domain.ErrSessionNotFound
			}
			return err
		}
		return txRepo.SetUserDefaultParishIfNil(ctx, p.UserID, parishID)
	})
}

// SearchMySessions returns active non-expired sessions for the authenticated user.
func (s *Service) SearchMySessions(
	ctx context.Context,
	p authx.Principal,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.UserSessionResponse, bool, error) {
	rows, err := s.repo.SearchUserSessions(ctx, p.UserID, time.Now().UTC(), cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.UserSessionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.UserSessionResponse{
			ID:             row.ID,
			UserAgent:      row.UserAgent,
			IPAddress:      row.IPAddress,
			ActiveParishID: row.ActiveParishID,
			IsCurrent:      p.SessionID != uuid.Nil && row.ID == p.SessionID,
			LastActiveAt:   row.LastActiveAt,
			ExpiresAt:      row.ExpiresAt,
			CreatedAt:      row.CreatedAt,
		})
	}
	return items, hasMore, nil
}

// RevokeMySession revokes a session owned by the authenticated user and reports whether it was the current session.
func (s *Service) RevokeMySession(ctx context.Context, p authx.Principal, sessionID uuid.UUID) (bool, error) {
	deleted, err := s.repo.DeleteUserSessionByID(ctx, sessionID, p.UserID)
	if err != nil {
		return false, err
	}
	if deleted == 0 {
		return false, domain.ErrSessionNotFound
	}
	return p.SessionID != uuid.Nil && sessionID == p.SessionID, nil
}

// SearchMyAccounts returns linked provider accounts for the authenticated user.
func (s *Service) SearchMyAccounts(
	ctx context.Context,
	p authx.Principal,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.UserAccountResponse, bool, error) {
	rows, err := s.repo.SearchUserAccounts(ctx, p.UserID, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.UserAccountResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUserAccountResponse(row))
	}
	return items, hasMore, nil
}

func toUserAccountResponse(row *ent.UserAccount) domain.UserAccountResponse {
	return domain.UserAccountResponse{
		ID:             row.ID,
		Provider:       row.Provider,
		Subject:        row.Subject,
		Email:          row.Email,
		DisplayName:    row.DisplayName,
		AvatarURL:      row.AvatarURL,
		ActiveParishID: row.ActiveParishID,
		LastUsedAt:     row.LastUsedAt,
		CreatedAt:      row.CreatedAt,
	}
}

// UnlinkMyAccount removes a linked provider account unless it is the user's last remaining account.
func (s *Service) UnlinkMyAccount(ctx context.Context, p authx.Principal, accountID uuid.UUID) error {
	return s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if _, err := txRepo.FindUserAccountByID(ctx, accountID, p.UserID); ent.IsNotFound(err) {
			return domain.ErrAccountNotFound
		} else if err != nil {
			return err
		}
		count, err := txRepo.CountUserAccounts(ctx, p.UserID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domain.ErrLastAccountRequired
		}
		_, err = txRepo.DeleteUserAccountByID(ctx, accountID, p.UserID)
		return err
	})
}
