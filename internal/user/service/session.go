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

// ResolveSession validates a raw session token, refreshes sliding expiry, and returns an authx.Principal.
func (s *Service) ResolveSession(ctx context.Context, rawToken string, headerParishID uuid.UUID) (authx.Principal, error) {
	if rawToken == "" {
		return authx.Principal{}, domain.ErrSessionInvalid
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
		return authx.Principal{}, domain.ErrUserDisabled
	}

	if headerParishID != uuid.Nil {
		isSuper, err := s.hasSuperAdminRole(u.ID)
		if err != nil {
			return authx.Principal{}, err
		}
		if !isSuper {
			ok, err := s.repo.IsActiveParishMember(ctx, u.ID, headerParishID)
			if err != nil {
				return authx.Principal{}, err
			}
			if !ok {
				return authx.Principal{}, errx.Forbidden("FORBIDDEN", "User is not a member of this parish.")
			}
		}
	}

	parishID := resolveParishID(headerParishID, sess.ActiveParishID, u.DefaultParishID)
	if err := s.repo.RefreshSession(ctx, sess.ID, now, now.Add(s.cfg.SessionTTL)); err != nil {
		return authx.Principal{}, err
	}

	roles, err := s.loadRoles(ctx, u.ID, parishID)
	if err != nil {
		return authx.Principal{}, err
	}

	return authx.Principal{
		UserID:    u.ID,
		SessionID: sess.ID,
		Email:     u.Email,
		Name:      u.Name,
		ParishID:  parishID,
		Roles:     roles,
		Channel:   "web",
	}, nil
}

// ResolveByChannel resolves a linked channel identity (e.g. telegram, whatsapp) into an authx.Principal.
func (s *Service) ResolveByChannel(ctx context.Context, provider, subject string, overrideParishID uuid.UUID) (authx.Principal, error) {
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

	parishID := resolveParishID(overrideParishID, acc.ActiveParishID, u.DefaultParishID)
	roles, err := s.loadRoles(ctx, u.ID, parishID)
	if err != nil {
		return authx.Principal{}, err
	}

	return authx.Principal{
		UserID:   u.ID,
		Email:    u.Email,
		Name:     u.Name,
		ParishID: parishID,
		Roles:    roles,
		Channel:  provider,
	}, nil
}

// SwitchChannelParish persists the active parish for a linked channel account (and initializes user default if nil).
func (s *Service) SwitchChannelParish(ctx context.Context, provider, subject string, parishID uuid.UUID) error {
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
		if err := txRepo.UpdateAccountActiveParish(ctx, acc.ID, parishID); err != nil {
			return err
		}
		return txRepo.SetUserDefaultParishIfNil(ctx, acc.UserID, parishID)
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

func (s *Service) guardParishAccess(ctx context.Context, userID, parishID uuid.UUID) error {
	isSuper, err := s.hasSuperAdminRole(userID)
	if err != nil {
		return err
	}
	if isSuper {
		ok, err := s.repo.IsActiveParish(ctx, parishID)
		if err != nil {
			return err
		}
		if !ok {
			return errx.Forbidden("FORBIDDEN", "Parish not found or inactive.")
		}
		return nil
	}
	ok, err := s.repo.IsActiveParishMember(ctx, userID, parishID)
	if err != nil {
		return err
	}
	if !ok {
		return errx.Forbidden("FORBIDDEN", "User is not a member of this parish.")
	}
	return nil
}

func (s *Service) loadGlobalRoles(userID uuid.UUID) ([]string, error) {
	if s.enforcer == nil {
		return []string{}, nil
	}
	return s.enforcer.RolesForUserInDomain(authz.UserSubject(userID), authz.DomainGlobal)
}

func (s *Service) loadRoles(ctx context.Context, userID, parishID uuid.UUID) ([]string, error) {
	globalRoles, err := s.loadGlobalRoles(userID)
	if err != nil {
		return nil, err
	}
	if s.enforcer == nil || parishID == uuid.Nil {
		return globalRoles, nil
	}
	parishRoles, err := s.enforcer.RolesForUserInDomain(authz.UserSubject(userID), authz.ParishDomain(parishID))
	if err != nil {
		return nil, err
	}
	isOrgAdmin, err := s.repo.IsActiveOrgAdminInParish(ctx, userID, parishID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(globalRoles)+len(parishRoles)+1)
	merged := make([]string, 0, len(globalRoles)+len(parishRoles)+1)
	appendRole := func(r string) {
		if _, ok := seen[r]; !ok {
			seen[r] = struct{}{}
			merged = append(merged, r)
		}
	}
	for _, r := range globalRoles {
		appendRole(r)
	}
	for _, r := range parishRoles {
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

func resolveParishID(override uuid.UUID, scoped, fallback *uuid.UUID) uuid.UUID {
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
