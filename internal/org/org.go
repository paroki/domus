package org

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/handler"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/org/service"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref"
	"github.com/paroki/domus/internal/user"
)

// Re-export read-only DTO types for downstream features.
type (
	Organization  = domain.OrganizationResponse
	Member        = domain.OrgMemberResponse
	Invite        = domain.OrgInviteResponse
	InvitePreview = domain.OrgInvitePreviewResponse
)

// Feature-owned Casbin roles and resource permissions.
var (
	RoleAdmin   = authz.Role("org", "admin")
	RoleMember  = authz.Role("org", "member")
	PermOrgs    = authz.Permission("org", "orgs")
	PermMembers = authz.Permission("org", "members")
	PermInvites = authz.Permission("org", "invites")
)

type userEnsurerAdapter struct {
	users *user.Feature
}

func (a userEnsurerAdapter) EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (uuid.UUID, error) {
	u, err := a.users.EnsureUser(ctx, email, name, phone, provider, subject)
	if err != nil {
		return uuid.Nil, err
	}
	return u.ID, nil
}

type refValidatorAdapter struct {
	refs *ref.Feature
}

func (a refValidatorAdapter) ValidateZoneInParish(ctx context.Context, parishID, zoneID uuid.UUID) error {
	_, err := a.refs.ValidateZoneInParish(ctx, parishID, zoneID)
	return err
}

func (a refValidatorAdapter) ValidateWardInParish(ctx context.Context, parishID, wardID uuid.UUID) error {
	_, err := a.refs.ValidateWardInParish(ctx, parishID, wardID)
	return err
}

func (a refValidatorAdapter) ValidatePeriodInParish(ctx context.Context, parishID, periodID uuid.UUID) error {
	_, err := a.refs.ValidatePeriodInParish(ctx, parishID, periodID)
	return err
}

func (a refValidatorAdapter) EnsureParishMember(ctx context.Context, parishID, userID uuid.UUID) error {
	_, err := a.refs.EnsureParishMember(ctx, parishID, userID)
	return err
}

// Feature wires organization management and cross-feature helpers.
type Feature struct {
	service *service.Service
	handler *handler.Handler
	users   *user.Feature
	cfg     Config
}

// New constructs the org Feature.
func New(db *ent.Client, enforcer *authz.Enforcer, users *user.Feature, refs *ref.Feature, cfg Config) *Feature {
	r := repo.New(db)
	var ensurer service.UserEnsurer
	if users != nil {
		ensurer = userEnsurerAdapter{users: users}
	}
	var refVal service.RefValidator
	if refs != nil {
		refVal = refValidatorAdapter{refs: refs}
	}
	svc := service.New(r, enforcer, ensurer, refVal)
	h := handler.New(svc)
	return &Feature{
		service: svc,
		handler: h,
		users:   users,
		cfg:     cfg,
	}
}

// Register mounts /orgs routes under /api.
func (f *Feature) Register(api fiber.Router) {
	f.handler.Register(
		api,
		f.users.AuthenticateMiddleware(),
		f.users.RequireAuth(),
		requireTenantMiddleware(),
		f.users.Require(PermOrgs, authz.ActionRead),
	)
}

// RegisterPermissions registers org feature policies with the Casbin enforcer.
func (f *Feature) RegisterPermissions(ctx context.Context, e *authz.Enforcer) error {
	if e == nil {
		return nil
	}
	policies := []struct {
		role string
		perm string
		act  authz.Action
	}{
		{user.RoleUser, PermOrgs, authz.ActionRead},
		{ref.RoleParishAdmin, PermOrgs, authz.ActionRead},
		{ref.RoleParishAdmin, PermOrgs, authz.ActionWrite},
		{RoleAdmin, PermOrgs, authz.ActionRead},
		{RoleAdmin, PermOrgs, authz.ActionWrite},
		{RoleAdmin, PermMembers, authz.ActionRead},
		{RoleAdmin, PermMembers, authz.ActionWrite},
		{RoleAdmin, PermInvites, authz.ActionRead},
		{RoleAdmin, PermInvites, authz.ActionWrite},
		{RoleMember, PermOrgs, authz.ActionRead},
		{RoleMember, PermMembers, authz.ActionRead},
	}
	for _, p := range policies {
		if err := e.AddPermissionForRole(ctx, p.role, authz.DomainGlobal, p.perm, p.act); err != nil {
			return err
		}
	}
	return nil
}

// GetOrganization returns a non-deleted Organization in parishID.
func (f *Feature) GetOrganization(ctx context.Context, parishID, orgID uuid.UUID) (*Organization, error) {
	return f.service.GetOrganization(ctx, parishID, orgID)
}

// IsMember reports whether userID is an active member of orgID.
func (f *Feature) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return f.service.IsMember(ctx, orgID, userID)
}

// IsAdmin reports whether userID is an active admin of orgID.
func (f *Feature) IsAdmin(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	return f.service.IsAdmin(ctx, orgID, userID)
}

// ListUserOrgs returns active, non-deleted Organizations in parishID where userID is an active member.
func (f *Feature) ListUserOrgs(ctx context.Context, parishID, userID uuid.UUID) ([]Organization, error) {
	return f.service.ListUserOrgs(ctx, parishID, userID)
}

func requireTenantMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if _, err := user.RequireTenant(c.Context()); err != nil {
			return err
		}
		return c.Next()
	}
}
