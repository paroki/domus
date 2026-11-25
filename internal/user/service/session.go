package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// ResolveSession validates a raw session token or Personal Access Token (dms_*) and returns an authx.Principal.
func (s *Service) ResolveSession(ctx context.Context, rawToken string, headerParokiID uuid.UUID) (authx.Principal, error) {
	if rawToken == "" {
		return authx.Principal{}, domain.ErrSessionInvalid
	}
	if strings.HasPrefix(rawToken, "dms_") {
		return s.resolvePAT(ctx, rawToken, headerParokiID)
	}

	sess, u, err := s.repo.FindSessionWithUserByTokenHash(ctx, sha256Hex(rawToken))
	if ent.IsNotFound(err) {
		return authx.Principal{}, domain.ErrSessionInvalid
	}
	if err != nil {
		return authx.Principal{}, err
	}

	now := time.Now().UTC()
	if now.After(sess.ExpiresAt) {
		_ = s.repo.DeleteSessionByID(ctx, sess.ID)
		return authx.Principal{}, domain.ErrSessionInvalid
	}

	if u.Status == user.StatusDisabled {
		_ = s.repo.DeleteSessionsByUserID(ctx, u.ID)
		_ = s.repo.DeleteTokensByUserID(ctx, u.ID)
		return authx.Principal{}, domain.ErrUserDisabled
	}

	parokiID, roles, err := s.resolveTenantAccess(ctx, u, headerParokiID, sess.ActiveParokiID)
	if err != nil {
		return authx.Principal{}, err
	}

	if err := s.repo.RefreshSession(ctx, sess.ID, now, now.Add(s.cfg.SessionTTL)); err != nil {
		return authx.Principal{}, err
	}

	return authx.Principal{
		UserID:    u.ID,
		SessionID: sess.ID,
		Email:     u.Email,
		Name:      u.Name,
		ParokiID:  parokiID,
		Roles:     roles,
		Channel:   "web",
	}, nil
}

func (s *Service) resolvePAT(ctx context.Context, rawToken string, headerParokiID uuid.UUID) (authx.Principal, error) {
	tok, u, err := s.repo.FindTokenWithUserByTokenHash(ctx, sha256Hex(rawToken))
	if ent.IsNotFound(err) {
		return authx.Principal{}, domain.ErrSessionInvalid
	}
	if err != nil {
		return authx.Principal{}, err
	}

	now := time.Now().UTC()
	if tok.ExpiresAt != nil && now.After(*tok.ExpiresAt) {
		_ = s.repo.DeleteTokenByID(ctx, tok.ID)
		return authx.Principal{}, domain.ErrSessionInvalid
	}

	if u.Status == user.StatusDisabled {
		_ = s.repo.DeleteSessionsByUserID(ctx, u.ID)
		_ = s.repo.DeleteTokensByUserID(ctx, u.ID)
		return authx.Principal{}, domain.ErrUserDisabled
	}

	parokiID, roles, err := s.resolveTenantAccess(ctx, u, headerParokiID, nil)
	if err != nil {
		return authx.Principal{}, err
	}

	if err := s.repo.UpdateTokenLastUsed(ctx, tok.ID, now); err != nil {
		return authx.Principal{}, err
	}

	return authx.Principal{
		UserID:   u.ID,
		Email:    u.Email,
		Name:     u.Name,
		ParokiID: parokiID,
		Roles:    roles,
		Channel:  "api",
	}, nil
}

func (s *Service) resolveTenantAccess(ctx context.Context, u *ent.User, headerParokiID uuid.UUID, activeParokiID *uuid.UUID) (uuid.UUID, []string, error) {
	if headerParokiID != uuid.Nil {
		isSuper, err := s.hasSuperAdminRole(u.ID)
		if err != nil {
			return uuid.Nil, nil, err
		}
		if !isSuper {
			ok, err := s.repo.IsActiveParokiMember(ctx, u.ID, headerParokiID)
			if err != nil {
				return uuid.Nil, nil, err
			}
			if !ok {
				return uuid.Nil, nil, errx.Forbidden("FORBIDDEN", "User is not a member of this paroki.")
			}
		}
	}

	parokiID := resolveParokiID(headerParokiID, activeParokiID, u.DefaultParokiID)
	roles, err := s.loadRoles(ctx, u.ID, parokiID)
	if err != nil {
		return uuid.Nil, nil, err
	}
	return parokiID, roles, nil
}

// ResolveByChannel resolves a linked channel identity (e.g. telegram, whatsapp) into an authx.Principal.
func (s *Service) ResolveByChannel(ctx context.Context, provider, subject string, overrideParokiID uuid.UUID) (authx.Principal, error) {
	normSubject, err := normalizeChannelSubject(provider, subject)
	if err != nil {
		return authx.Principal{}, err
	}
	acc, err := s.repo.FindAccountByProviderSubject(ctx, provider, normSubject)
	if ent.IsNotFound(err) {
		return authx.Principal{}, domain.ErrUserNotFound
	}
	if err != nil {
		return authx.Principal{}, err
	}

	u, err := guardActiveUser(s.repo.FindUserByID(ctx, acc.UserID))
	if err != nil {
		return authx.Principal{}, err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateAccountLastUsed(ctx, acc.ID, now); err != nil {
		return authx.Principal{}, err
	}

	parokiID := resolveParokiID(overrideParokiID, acc.ActiveParokiID, u.DefaultParokiID)
	roles, err := s.loadRoles(ctx, u.ID, parokiID)
	if err != nil {
		return authx.Principal{}, err
	}

	return authx.Principal{
		UserID:   u.ID,
		Email:    u.Email,
		Name:     u.Name,
		ParokiID: parokiID,
		Roles:    roles,
		Channel:  provider,
	}, nil
}

// SwitchChannelParoki persists the active paroki for a linked channel account (and initializes user default if nil).
func (s *Service) SwitchChannelParoki(ctx context.Context, provider, subject string, parokiID uuid.UUID) error {
	normSubject, err := normalizeChannelSubject(provider, subject)
	if err != nil {
		return err
	}
	acc, err := s.repo.FindAccountByProviderSubject(ctx, provider, normSubject)
	if ent.IsNotFound(err) {
		return domain.ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	return s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if err := txRepo.UpdateAccountActiveParoki(ctx, acc.ID, parokiID); err != nil {
			return err
		}
		return txRepo.SetUserDefaultParokiIfNil(ctx, acc.UserID, parokiID)
	})
}

// BootstrapSuperAdmin idempotently seeds the initial Super Admin user, email account, and Casbin roles.
func (s *Service) BootstrapSuperAdmin(ctx context.Context) error {
	email := normalizeEmail(s.cfg.SuperAdminEmail)
	if email == "" {
		return nil
	}
	name := strings.TrimSpace(s.cfg.SuperAdminName)
	if name == "" {
		name = "Super Admin"
	}

	var userID uuid.UUID
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		u, err := txRepo.FindUserByEmail(ctx, email)
		if ent.IsNotFound(err) {
			u, err = txRepo.CreateUser(ctx, email, name, nil)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		if _, err := txRepo.UpsertAccount(ctx, u.ID, domain.ProviderEmail, email, &u.Email, &u.Name, u.AvatarURL, time.Now().UTC()); err != nil {
			return err
		}
		userID = u.ID
		return nil
	})
	if err != nil {
		return err
	}

	if s.enforcer != nil {
		sub := authz.UserSubject(userID)
		if err := s.enforcer.AssignRole(ctx, sub, domain.RoleSuperAdmin, authz.DomainGlobal); err != nil {
			return err
		}
		if err := s.enforcer.AssignRole(ctx, sub, domain.RoleUser, authz.DomainGlobal); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) hasSuperAdminRole(userID uuid.UUID) (bool, error) {
	if s.enforcer == nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(userID), domain.RoleSuperAdmin, authz.DomainGlobal)
}

func (s *Service) guardParokiAccess(ctx context.Context, userID, parokiID uuid.UUID) error {
	isSuper, err := s.hasSuperAdminRole(userID)
	if err != nil {
		return err
	}
	if isSuper {
		ok, err := s.repo.IsActiveParoki(ctx, parokiID)
		if err != nil {
			return err
		}
		if !ok {
			return errx.Forbidden("FORBIDDEN", "Paroki not found or inactive.")
		}
		return nil
	}
	ok, err := s.repo.IsActiveParokiMember(ctx, userID, parokiID)
	if err != nil {
		return err
	}
	if !ok {
		return errx.Forbidden("FORBIDDEN", "User is not a member of this paroki.")
	}
	return nil
}

func (s *Service) loadGlobalRoles(userID uuid.UUID) ([]string, error) {
	if s.enforcer == nil {
		return []string{}, nil
	}
	return s.enforcer.RolesForUserInDomain(authz.UserSubject(userID), authz.DomainGlobal)
}

func (s *Service) loadRoles(ctx context.Context, userID, parokiID uuid.UUID) ([]string, error) {
	globalRoles, err := s.loadGlobalRoles(userID)
	if err != nil {
		return nil, err
	}
	if s.enforcer == nil || parokiID == uuid.Nil {
		return globalRoles, nil
	}
	parokiRoles, err := s.enforcer.RolesForUserInDomain(authz.UserSubject(userID), authz.ParokiDomain(parokiID))
	if err != nil {
		return nil, err
	}
	isOrgAdmin, err := s.repo.IsActiveOrgAdminInParoki(ctx, userID, parokiID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(globalRoles)+len(parokiRoles)+1)
	merged := make([]string, 0, len(globalRoles)+len(parokiRoles)+1)
	appendRole := func(r string) {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			merged = append(merged, r)
		}
	}
	for _, r := range globalRoles {
		appendRole(r)
	}
	for _, r := range parokiRoles {
		appendRole(r)
	}
	if isOrgAdmin {
		appendRole(authz.Role("org", "admin"))
	}
	return merged, nil
}

func normalizeChannelSubject(provider, rawSubject string) (string, error) {
	subject := strings.TrimSpace(rawSubject)
	if provider == domain.ProviderWhatsApp && subject != "" && !strings.Contains(subject, "@") {
		return normalizePhone(subject)
	}
	return subject, nil
}

func resolveParokiID(override uuid.UUID, scoped, fallback *uuid.UUID) uuid.UUID {
	if override != uuid.Nil {
		return override
	}
	if scoped != nil && *scoped != uuid.Nil {
		return *scoped
	}
	if fallback != nil && *fallback != uuid.Nil {
		return *fallback
	}
	return uuid.Nil
}
