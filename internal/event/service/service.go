package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/event/repo"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
)

const roleOrgAdmin = "org:admin"

// OrgInfo holds read-only organization metadata needed by the event service.
type OrgInfo struct {
	ID   uuid.UUID
	Code string
	Name string
	Kind string
}

// OrgReader reads organization metadata and membership roles from the org feature.
type OrgReader interface {
	GetOrganization(ctx context.Context, parokiID, orgID uuid.UUID) (*OrgInfo, error)
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	IsAdmin(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	ListUserOrgs(ctx context.Context, parokiID, userID uuid.UUID) ([]OrgInfo, error)
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
	cfg      Config
}

// New constructs an event Service.
func New(r *repo.Repository, enforcer *authz.Enforcer, orgs OrgReader, cfg Config) *Service {
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
		cfg:      cfg,
	}
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
