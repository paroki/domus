package service

import (
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/parishmember"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
)

const (
	roleSuperAdmin  = "user:super-admin"
	roleParishAdmin = "ref:parish-admin"
)

var (
	codeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)
	e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
)

// UserEnsurer finds or creates a user by email/profile and returns their user ID.
type UserEnsurer interface {
	EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (uuid.UUID, error)
}

// Service implements business logic for dioceses, parishes, parish members, and ref facade helpers.
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

// ValidateZoneInParish returns a non-deleted Zone if it belongs to parishID.
func (s *Service) ValidateZoneInParish(ctx context.Context, parishID, zoneID uuid.UUID) (*domain.ZoneResponse, error) {
	z, err := s.repo.FindZoneInParish(ctx, parishID, zoneID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrZoneNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toZoneResponse(z)
	return &res, nil
}

// ValidateWardInParish returns a non-deleted Ward if it belongs to parishID.
func (s *Service) ValidateWardInParish(ctx context.Context, parishID, wardID uuid.UUID) (*domain.WardResponse, error) {
	w, err := s.repo.FindWardInParish(ctx, parishID, wardID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrWardNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toWardResponse(w)
	return &res, nil
}

// ValidatePeriodInParish returns a non-deleted Period if it belongs to parishID.
func (s *Service) ValidatePeriodInParish(ctx context.Context, parishID, periodID uuid.UUID) (*domain.PeriodResponse, error) {
	p, err := s.repo.FindPeriodInParish(ctx, parishID, periodID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toPeriodResponse(p)
	return &res, nil
}

// EnsureParishMember idempotently ensures userID is a member of parishID.
func (s *Service) EnsureParishMember(ctx context.Context, parishID, userID uuid.UUID) (*domain.ParishMemberResponse, error) {
	if _, err := s.repo.FindParishByID(ctx, parishID); ent.IsNotFound(err) {
		return nil, domain.ErrParishNotFound
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

	var pm *ent.ParishMember
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindParishMember(ctx, parishID, userID)
		if findErr == nil {
			pm = existing
			return nil
		}
		if !ent.IsNotFound(findErr) {
			return findErr
		}
		created, createErr := txRepo.CreateParishMember(ctx, parishID, userID, parishmember.RoleMember)
		if createErr != nil {
			return createErr
		}
		if err := txRepo.SetUserDefaultParishIfNil(ctx, userID, parishID); err != nil {
			return err
		}
		pm = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	res := toParishMemberResponse(pm, u)
	return &res, nil
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

func (s *Service) isParishAdmin(p authx.Principal, parishID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || parishID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleParishAdmin, authz.ParishDomain(parishID))
}

func (s *Service) requireParishAdminOrSuperAdmin(p authx.Principal, parishID uuid.UUID) (bool, error) {
	if s.isSuperAdmin(p) {
		return true, nil
	}
	ok, err := s.isParishAdmin(p, parishID)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, domain.ErrRefForbidden
	}
	return false, nil
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

func toZoneResponse(z *ent.Zone) domain.ZoneResponse {
	return domain.ZoneResponse{
		ID:               z.ID,
		ParishID:         z.ParishID,
		Type:             string(z.Type),
		Code:             z.Code,
		Name:             z.Name,
		CoordinatorName:  z.CoordinatorName,
		CoordinatorPhone: z.CoordinatorPhone,
		Active:           z.Active,
		CreatedAt:        z.CreatedAt,
		UpdatedAt:        z.UpdatedAt,
	}
}

func toWardResponse(w *ent.Ward) domain.WardResponse {
	return domain.WardResponse{
		ID:          w.ID,
		ParishID:    w.ParishID,
		ZoneID:      w.ZoneID,
		Code:        w.Code,
		Name:        w.Name,
		PatronSaint: w.PatronSaint,
		HeadName:    w.HeadName,
		HeadPhone:   w.HeadPhone,
		Active:      w.Active,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func toPeriodResponse(p *ent.Period) domain.PeriodResponse {
	return domain.PeriodResponse{
		ID:        p.ID,
		ParishID:  p.ParishID,
		Type:      string(p.Type),
		Code:      p.Code,
		Name:      p.Name,
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
		Status:    string(p.Status),
		IsCurrent: p.IsCurrent,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
