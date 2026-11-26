package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
)

const (
	roleOrgAdmin    = "org:admin"
	roleParokiAdmin = "ref:paroki-admin"
)

// DefaultParokiTimezone is the fallback IANA timezone when a paroki has none configured.
const DefaultParokiTimezone = "Asia/Makassar"

// OrgInfo holds read-only organization metadata needed by the event service.
type OrgInfo struct {
	ID   uuid.UUID
	Code string
	Name string
	Kind string
}

// ParokiInfo holds read-only parish metadata needed by the event service.
type ParokiInfo struct {
	ID       uuid.UUID
	Kode     string
	Nama     string
	Timezone string
}

// OrgReader reads organization metadata and membership roles from the org feature.
type OrgReader interface {
	GetOrganization(ctx context.Context, parokiID, orgID uuid.UUID) (*OrgInfo, error)
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	IsAdmin(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	ListUserOrgs(ctx context.Context, parokiID, userID uuid.UUID) ([]OrgInfo, error)
}

// ParokiReader reads active parish metadata (including timezone) and lookups from the ref feature.
type ParokiReader interface {
	GetParoki(ctx context.Context, parokiID uuid.UUID) (*ParokiInfo, error)
	HasActiveLookup(ctx context.Context, parokiID uuid.UUID, category, code string) (bool, error)
}

// Config holds runtime configuration for the event service.
type Config struct {
	DefaultVisibility string
	RSVPGracePeriod   time.Duration
}

// Service implements business logic for parish and organizational events.
type Service struct {
	repo     *repo.Repository
	enforcer *authz.Enforcer
	orgs     OrgReader
	parokis  ParokiReader
	cfg      Config
}

// New constructs an event Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, orgs OrgReader, parokis ParokiReader, cfg Config) *Service {
	if cfg.DefaultVisibility == "" {
		cfg.DefaultVisibility = "public"
	}
	if cfg.RSVPGracePeriod <= 0 {
		cfg.RSVPGracePeriod = 30 * time.Minute
	}
	return &Service{
		repo:     r,
		enforcer: enforcer,
		orgs:     orgs,
		parokis:  parokis,
		cfg:      cfg,
	}
}

// ResolveParokiLocation resolves the IANA time.Location for parokiID, falling back to Asia/Makassar.
func (s *Service) ResolveParokiLocation(ctx context.Context, parokiID uuid.UUID) *time.Location {
	tz := DefaultParokiTimezone
	if s.parokis != nil {
		if info, err := s.parokis.GetParoki(ctx, parokiID); err == nil && info != nil && info.Timezone != "" {
			tz = info.Timezone
		}
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	loc, _ := time.LoadLocation(DefaultParokiTimezone)
	if loc != nil {
		return loc
	}
	return time.UTC
}

func (s *Service) isOrgAdmin(ctx context.Context, p authx.Principal, orgID uuid.UUID) (bool, error) {
	if s.enforcer != nil {
		hasRole, err := s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleOrgAdmin, authz.OrgDomain(orgID))
		if err != nil {
			return false, err
		}
		if !hasRole {
			return false, nil
		}
	}
	return s.orgs.IsAdmin(ctx, orgID, p.UserID)
}

func (s *Service) isParokiAdmin(ctx context.Context, p authx.Principal) (bool, error) {
	if s.enforcer != nil {
		hasRole, err := s.enforcer.HasRoleInDomain(authz.UserSubject(p.UserID), roleParokiAdmin, authz.ParokiDomain(p.ParokiID))
		if err != nil {
			return false, err
		}
		if hasRole {
			return true, nil
		}
	}
	return s.repo.IsParokiAdmin(ctx, p.ParokiID, p.UserID)
}
