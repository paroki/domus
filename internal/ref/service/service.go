package service

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/parokimember"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

const (
	roleSuperAdmin  = "user:super-admin"
	roleParokiAdmin = "ref:paroki-admin"
	roleUmatAdmin   = "umat:admin"
)

var (
	codeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)
	e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
)

// UserEnsurer finds or creates a user by email/profile and returns their user ID.
type UserEnsurer interface {
	EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (uuid.UUID, error)
}

// Service implements business logic for keuskupan, paroki, paroki members, and ref facade helpers.
type Service struct {
	repo     *repo.Repository
	enforcer *authz.Enforcer
	users    UserEnsurer
}

// New constructs a ref Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, users UserEnsurer) *Service {
	return &Service{
		repo:     r,
		enforcer: enforcer,
		users:    users,
	}
}

// ValidateWilayahInParoki returns a non-deleted Wilayah if it belongs to parokiID.
func (s *Service) ValidateWilayahInParoki(ctx context.Context, parokiID, wilayahID uuid.UUID) (*domain.WilayahResponse, error) {
	w, err := s.repo.FindWilayahInParoki(ctx, parokiID, wilayahID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrWilayahNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toWilayahResponse(w)
	return &res, nil
}

// ValidateLingkunganInParoki returns a non-deleted Lingkungan if it belongs to parokiID.
func (s *Service) ValidateLingkunganInParoki(ctx context.Context, parokiID, lingkunganID uuid.UUID) (*domain.LingkunganResponse, error) {
	l, err := s.repo.FindLingkunganInParoki(ctx, parokiID, lingkunganID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrLingkunganNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toLingkunganResponse(l)
	return &res, nil
}

// EnsureParokiMember idempotently ensures userID is a member of parokiID.
func (s *Service) EnsureParokiMember(ctx context.Context, parokiID, userID uuid.UUID) (*domain.ParokiMemberResponse, error) {
	if _, err := s.repo.FindParokiByID(ctx, parokiID); ent.IsNotFound(err) {
		return nil, domain.ErrParokiNotFound
	} else if err != nil {
		return nil, err
	}

	u, err := s.repo.FindUserByID(ctx, userID)
	if ent.IsNotFound(err) {
		return nil, domain.ValidationError("userId", "NOT_FOUND", "User not found.")
	}
	if err != nil {
		return nil, err
	}

	var pm *ent.ParokiMember
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindParokiMember(ctx, parokiID, userID)
		if findErr == nil {
			pm = existing
			return nil
		}
		if !ent.IsNotFound(findErr) {
			return findErr
		}
		created, createErr := txRepo.CreateParokiMember(ctx, parokiID, userID, parokimember.RoleMember, false)
		if createErr != nil {
			return createErr
		}
		if err := txRepo.SetUserDefaultParokiIfNil(ctx, userID, parokiID); err != nil {
			return err
		}
		pm = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	res := toParokiMemberResponse(pm, u)
	return &res, nil
}

// HasParokiMember reports whether userID is a member of parokiID.
func (s *Service) HasParokiMember(ctx context.Context, parokiID, userID uuid.UUID) (bool, error) {
	_, err := s.repo.FindParokiMember(ctx, parokiID, userID)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) isSuperAdmin(p authx.Principal) bool {
	if slices.Contains(p.Roles, roleSuperAdmin) {
		return true
	}
	if s.enforcer == nil || p.UserID == uuid.Nil {
		return false
	}
	has, err := s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleSuperAdmin, authz.DomainGlobal)
	return err == nil && has
}

func (s *Service) isParokiAdmin(p authx.Principal, parokiID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || parokiID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleParokiAdmin, authz.ParokiDomain(parokiID))
}

func (s *Service) isUmatAdmin(p authx.Principal, parokiID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || parokiID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleUmatAdmin, authz.ParokiDomain(parokiID))
}

func (s *Service) requireParokiAdminOrSuperAdmin(p authx.Principal, parokiID uuid.UUID) (bool, error) {
	if s.isSuperAdmin(p) {
		return true, nil
	}
	ok, err := s.isParokiAdmin(p, parokiID)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, domain.ErrRefForbidden
	}
	return false, nil
}

func normalizeKode(raw string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(raw))
	if !codeRe.MatchString(norm) {
		return "", domain.ValidationError("kode", "INVALID_CODE", "Kode must be 2-64 lowercase alphanumeric characters, hyphens, or underscores.")
	}
	return norm, nil
}

func normalizeCode(raw string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(raw))
	if !codeRe.MatchString(norm) {
		return "", domain.ValidationError("code", "INVALID_CODE", "Code must be 2-64 lowercase alphanumeric characters, hyphens, or underscores.")
	}
	return norm, nil
}

func normalizePhone(raw string) (string, error) {
	cleaned := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" {
		return "", nil
	}
	if strings.HasPrefix(cleaned, "0") {
		cleaned = "+62" + cleaned[1:]
	} else if strings.HasPrefix(cleaned, "62") && !strings.HasPrefix(cleaned, "+") {
		cleaned = "+" + cleaned
	}
	if !e164Re.MatchString(cleaned) {
		return "", domain.ValidationError("phone", "INVALID_PHONE", "Phone number must be a valid E.164 number.")
	}
	return cleaned, nil
}

func normalizeOptPhone(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	norm, err := normalizePhone(*raw)
	if err != nil {
		return nil, err
	}
	return &norm, nil
}

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizeOptEmail(raw *string) *string {
	if raw == nil {
		return nil
	}
	norm := normalizeEmail(*raw)
	return &norm
}

func trimOptString(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	return &trimmed
}

func toWilayahResponse(w *ent.Wilayah) domain.WilayahResponse {
	return domain.WilayahResponse{
		ID:                 w.ID,
		ParokiID:           w.ParokiID,
		Jenis:              string(w.Jenis),
		Kode:               w.Kode,
		Nama:               w.Nama,
		Koordinator:        w.Koordinator,
		TeleponKoordinator: w.TeleponKoordinator,
		Active:             w.Active,
		CreatedAt:          w.CreatedAt,
		UpdatedAt:          w.UpdatedAt,
	}
}

func toLingkunganResponse(l *ent.Lingkungan) domain.LingkunganResponse {
	return domain.LingkunganResponse{
		ID:           l.ID,
		ParokiID:     l.ParokiID,
		WilayahID:    l.WilayahID,
		Kode:         l.Kode,
		Nama:         l.Nama,
		Pelindung:    l.Pelindung,
		Ketua:        l.Ketua,
		TeleponKetua: l.TeleponKetua,
		Active:       l.Active,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}
