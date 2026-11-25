package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// GetMe returns the authenticated user's profile with active paroki and global roles.
func (s *Service) GetMe(ctx context.Context, p authx.Principal) (domain.UserResponse, error) {
	u, err := guardActiveUser(s.repo.FindUserByID(ctx, p.UserID))
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.enrichMeResponse(ctx, u, p)
}

// GetUser returns a non-deleted active user's profile by ID.
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*domain.UserResponse, error) {
	u, err := guardActiveUser(s.repo.FindUserByID(ctx, userID))
	if err != nil {
		return nil, err
	}
	res := toUserResponse(u)
	return &res, nil
}

// UpdateMe updates the authenticated user's mutable profile fields.
func (s *Service) UpdateMe(ctx context.Context, p authx.Principal, req domain.UpdateMeRequest) (domain.UserResponse, error) {
	current, err := guardActiveUser(s.repo.FindUserByID(ctx, p.UserID))
	if err != nil {
		return domain.UserResponse{}, err
	}

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

	var genderPtr *user.Gender
	if req.Gender != nil {
		trimmed := strings.TrimSpace(*req.Gender)
		if trimmed == "" {
			return domain.UserResponse{}, domain.ValidationError("gender", "REQUIRED", "Gender is required.")
		}
		g := user.Gender(trimmed)
		if g != user.GenderPria && g != user.GenderWanita {
			return domain.UserResponse{}, domain.ValidationError("gender", "INVALID_GENDER", "Gender must be pria or wanita.")
		}
		genderPtr = &g
	}

	var birthDatePtr *string
	if req.BirthDate != nil {
		trimmed := strings.TrimSpace(*req.BirthDate)
		if trimmed == "" {
			return domain.UserResponse{}, domain.ValidationError("birthDate", "REQUIRED", "Birth date is required.")
		}
		parsed, parseErr := time.Parse("2006-01-02", trimmed)
		if parseErr != nil {
			return domain.UserResponse{}, domain.ValidationError("birthDate", "INVALID_DATE", "Birth date must be formatted as YYYY-MM-DD.")
		}
		now := time.Now().UTC()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if parsed.After(today.AddDate(-13, 0, 0)) || parsed.Before(today.AddDate(-120, 0, 0)) {
			return domain.UserResponse{}, domain.ValidationError("birthDate", "INVALID_DATE", "Age must be between 13 and 120 years.")
		}
		birthDatePtr = &trimmed
	}

	if req.Onboarded != nil && *req.Onboarded {
		effectiveGender := current.Gender
		if genderPtr != nil {
			effectiveGender = genderPtr
		}
		if effectiveGender == nil || *effectiveGender == "" {
			return domain.UserResponse{}, domain.ValidationError("gender", "REQUIRED", "Gender is required to complete onboarding.")
		}
		effectiveBirthDate := current.BirthDate
		if birthDatePtr != nil {
			effectiveBirthDate = birthDatePtr
		}
		if effectiveBirthDate == nil || *effectiveBirthDate == "" {
			return domain.UserResponse{}, domain.ValidationError("birthDate", "REQUIRED", "Birth date is required to complete onboarding.")
		}
	}

	var avatarPtr *string
	if req.AvatarURL != nil {
		trimmed := strings.TrimSpace(*req.AvatarURL)
		avatarPtr = &trimmed
	}

	u, err := s.repo.UpdateUser(ctx, p.UserID, namePtr, phonePtr, genderPtr, birthDatePtr, avatarPtr, req.DefaultParokiID, nil, req.Onboarded)
	if ent.IsNotFound(err) {
		return domain.UserResponse{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.UserResponse{}, err
	}
	return s.enrichMeResponse(ctx, u, p)
}

func (s *Service) enrichMeResponse(ctx context.Context, u *ent.User, p authx.Principal) (domain.UserResponse, error) {
	roles, err := s.loadRoles(ctx, u.ID, p.ParokiID)
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
	if p.ParokiID != uuid.Nil {
		pid := p.ParokiID
		res.ActiveParokiID = &pid
	}
	return res, nil
}

// SwitchSessionParoki updates active_paroki_id on the current session and initializes User.default_paroki_id if nil.
func (s *Service) SwitchSessionParoki(ctx context.Context, p authx.Principal, parokiID uuid.UUID) error {
	if p.SessionID == uuid.Nil {
		return domain.ErrSessionNotFound
	}
	if err := s.guardParokiAccess(ctx, p.UserID, parokiID); err != nil {
		return err
	}
	return s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if err := txRepo.UpdateSessionActiveParoki(ctx, p.SessionID, p.UserID, parokiID); err != nil {
			if ent.IsNotFound(err) {
				return domain.ErrSessionNotFound
			}
			return err
		}
		return txRepo.SetUserDefaultParokiIfNil(ctx, p.UserID, parokiID)
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
			ActiveParokiID: row.ActiveParokiID,
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

// CreateMyToken creates a new Personal Access Token for the authenticated user and returns the plaintext token once.
func (s *Service) CreateMyToken(ctx context.Context, p authx.Principal, req domain.CreateUserTokenRequest) (domain.CreateUserTokenResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.CreateUserTokenResponse{}, domain.ValidationError("name", "REQUIRED", "Token name is required.")
	}
	if len(name) > 100 {
		return domain.CreateUserTokenResponse{}, domain.ValidationError("name", "MAX_LENGTH", "Token name must be at most 100 characters.")
	}

	now := time.Now().UTC()
	var expiresAt *time.Time
	if req.ExpiresInDays != nil {
		if *req.ExpiresInDays < 1 || *req.ExpiresInDays > 365 {
			return domain.CreateUserTokenResponse{}, domain.ValidationError("expiresInDays", "OUT_OF_RANGE", "Expiration must be between 1 and 365 days.")
		}
		exp := now.AddDate(0, 0, *req.ExpiresInDays)
		expiresAt = &exp
	}

	suffix, err := randomToken(32)
	if err != nil {
		return domain.CreateUserTokenResponse{}, err
	}
	rawToken := "dms_" + suffix
	tokenPrefix := rawToken[:12]
	tokenHash := sha256Hex(rawToken)

	tok, err := s.repo.CreateUserToken(ctx, p.UserID, name, tokenPrefix, tokenHash, expiresAt)
	if err != nil {
		return domain.CreateUserTokenResponse{}, err
	}
	return domain.CreateUserTokenResponse{
		UserTokenResponse: toUserTokenResponse(tok),
		Token:             rawToken,
	}, nil
}

// SearchMyTokens returns active non-expired Personal Access Tokens for the authenticated user.
func (s *Service) SearchMyTokens(
	ctx context.Context,
	p authx.Principal,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.UserTokenResponse, bool, error) {
	rows, err := s.repo.SearchUserTokens(ctx, p.UserID, time.Now().UTC(), cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]domain.UserTokenResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toUserTokenResponse(row))
	}
	return items, hasMore, nil
}

// RevokeMyToken revokes a Personal Access Token owned by the authenticated user.
func (s *Service) RevokeMyToken(ctx context.Context, p authx.Principal, tokenID uuid.UUID) error {
	deleted, err := s.repo.DeleteUserTokenByID(ctx, tokenID, p.UserID)
	if err != nil {
		return err
	}
	if deleted == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}

func toUserTokenResponse(tok *ent.UserToken) domain.UserTokenResponse {
	return domain.UserTokenResponse{
		ID:          tok.ID,
		Name:        tok.Name,
		TokenPrefix: tok.TokenPrefix,
		LastUsedAt:  tok.LastUsedAt,
		ExpiresAt:   tok.ExpiresAt,
		CreatedAt:   tok.CreatedAt,
	}
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
		ActiveParokiID: row.ActiveParokiID,
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
