package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/umat"
)

const (
	roleParokiAdmin = "ref:paroki-admin"
	roleOrgAdmin    = "org:admin"
	roleOrgMember   = "org:member"
)

var (
	codeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)
	e164Re = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
)

// UserEnsurer finds or creates a user by email/profile and returns their user ID.
type UserEnsurer interface {
	EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (uuid.UUID, error)
	GetUserName(ctx context.Context, userID uuid.UUID) (string, error)
}

// UmatReader reads PII-free parishioner summaries from internal/umat.
type UmatReader interface {
	GetBasic(ctx context.Context, parokiID, umatID uuid.UUID) (*umat.UmatBasic, error)
	GetByUserID(ctx context.Context, parokiID, userID uuid.UUID) (*umat.UmatBasic, error)
}

// RefValidator validates cross-feature reference entities within a paroki.
type RefValidator interface {
	ValidateWilayahInParoki(ctx context.Context, parokiID, wilayahID uuid.UUID) error
	ValidateLingkunganInParoki(ctx context.Context, parokiID, lingkunganID uuid.UUID) error
	EnsureParokiMember(ctx context.Context, parokiID, userID uuid.UUID) error
}

// SessionSwitcher switches the active paroki on an existing user session.
type SessionSwitcher interface {
	SwitchSessionParoki(ctx context.Context, userID, sessionID, parokiID uuid.UUID) error
}

// InviteConfig holds invite TTL configuration.
type InviteConfig struct {
	DefaultTTL time.Duration
	MaxTTL     time.Duration
}

// Service implements business logic for organizations, members, invites, and facade helpers.
type Service struct {
	repo     *repo.Repository
	enforcer *authz.Enforcer
	users    UserEnsurer
	refs     RefValidator
	umats    UmatReader
	sessions SessionSwitcher
	cfg      InviteConfig
}

// SetUmatReader attaches an optional UmatReader to the org Service.
func (s *Service) SetUmatReader(u UmatReader) {
	s.umats = u
}

// New constructs an org Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, users UserEnsurer, refs RefValidator, sessions SessionSwitcher, cfg InviteConfig) *Service {
	if cfg.DefaultTTL <= 0 {
		cfg.DefaultTTL = 24 * time.Hour
	}
	if cfg.MaxTTL <= 0 {
		cfg.MaxTTL = 168 * time.Hour
	}
	return &Service{
		repo:     r,
		enforcer: enforcer,
		users:    users,
		refs:     refs,
		sessions: sessions,
		cfg:      cfg,
	}
}

// GetOrganization returns a non-deleted Organization in parokiID.
func (s *Service) GetOrganization(ctx context.Context, parokiID, orgID uuid.UUID) (*domain.OrganizationResponse, error) {
	row, err := s.repo.FindOrgInParoki(ctx, parokiID, orgID)
	if ent.IsNotFound(err) {
		return nil, domain.ErrOrgNotFound
	}
	if err != nil {
		return nil, err
	}
	res := toOrgResponse(row, nil)
	return &res, nil
}

// IsMember reports whether userID is an active member of orgID.
func (s *Service) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	_, err := s.repo.FindActiveOrgMember(ctx, orgID, userID)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// IsAdmin reports whether userID is an active admin of orgID.
func (s *Service) IsAdmin(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	m, err := s.repo.FindActiveOrgMember(ctx, orgID, userID)
	if ent.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return m.Role == orgmember.RoleAdmin, nil
}

// ListUserOrgs returns active, non-deleted Organizations in parokiID where userID is an active member.
func (s *Service) ListUserOrgs(ctx context.Context, parokiID, userID uuid.UUID) ([]domain.OrganizationResponse, error) {
	rows, err := s.repo.ListUserOrgsInParoki(ctx, parokiID, userID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	rolesMap, err := s.repo.ListActiveOrgRolesForUser(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.OrganizationResponse, 0, len(rows))
	for _, row := range rows {
		var myRole *string
		if r, ok := rolesMap[row.ID]; ok {
			myRole = &r
		}
		out = append(out, toOrgResponse(row, myRole))
	}
	return out, nil
}

func (s *Service) isParokiAdmin(p authx.Principal, parokiID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || parokiID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleParokiAdmin, authz.ParokiDomain(parokiID))
}

func (s *Service) isOrgAdmin(p authx.Principal, orgID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || orgID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleOrgAdmin, authz.OrgDomain(orgID))
}

func (s *Service) isOrgMember(p authx.Principal, orgID uuid.UUID) (bool, error) {
	if s.enforcer == nil || p.UserID == uuid.Nil || orgID == uuid.Nil {
		return false, nil
	}
	return s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleOrgMember, authz.OrgDomain(orgID))
}

func (s *Service) syncMemberCasbin(ctx context.Context, orgID, userID uuid.UUID, role orgmember.Role, status orgmember.Status) error {
	if s.enforcer == nil {
		return nil
	}
	sub := authz.UserSubject(userID)
	dom := authz.OrgDomain(orgID)
	if status == orgmember.StatusInactive {
		if err := s.enforcer.RevokeRole(ctx, sub, roleOrgAdmin, dom); err != nil {
			return err
		}
		return s.enforcer.RevokeRole(ctx, sub, roleOrgMember, dom)
	}
	if role == orgmember.RoleAdmin {
		if err := s.enforcer.AssignRole(ctx, sub, roleOrgAdmin, dom); err != nil {
			return err
		}
		return s.enforcer.RevokeRole(ctx, sub, roleOrgMember, dom)
	}
	if err := s.enforcer.AssignRole(ctx, sub, roleOrgMember, dom); err != nil {
		return err
	}
	return s.enforcer.RevokeRole(ctx, sub, roleOrgAdmin, dom)
}

func (s *Service) revokeMemberCasbin(ctx context.Context, orgID, userID uuid.UUID) error {
	if s.enforcer == nil {
		return nil
	}
	sub := authz.UserSubject(userID)
	dom := authz.OrgDomain(orgID)
	if err := s.enforcer.RevokeRole(ctx, sub, roleOrgAdmin, dom); err != nil {
		return err
	}
	return s.enforcer.RevokeRole(ctx, sub, roleOrgMember, dom)
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

func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func trimOptString(raw *string) *string {
	if raw == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*raw)
	return &trimmed
}

func toOrgResponse(o *ent.Organization, myRole *string) domain.OrganizationResponse {
	return domain.OrganizationResponse{
		ID:           o.ID,
		ParokiID:     o.ParokiID,
		ParentID:     o.ParentID,
		WilayahID:    o.WilayahID,
		LingkunganID: o.LingkunganID,
		Kind:         string(o.Kind),
		Code:         o.Code,
		Name:         o.Name,
		Description:  o.Description,
		Active:       o.Active,
		MyRole:       myRole,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

func toOrgMemberResponse(m *ent.OrgMember, u *ent.User) domain.OrgMemberResponse {
	res := domain.OrgMemberResponse{
		ID:        m.ID,
		ParokiID:  m.ParokiID,
		OrgID:     m.OrgID,
		UserID:    m.UserID,
		Role:      string(m.Role),
		Status:    string(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if u != nil {
		res.Email = u.Email
		res.Name = u.Name
		res.Phone = u.Phone
	}
	return res
}
