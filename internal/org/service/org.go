package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/organization"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

// SearchOrganizations returns paginated non-deleted organizations in p.ParokiID.
func (s *Service) SearchOrganizations(
	ctx context.Context,
	p authx.Principal,
	req domain.SearchOrganizationsRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.OrganizationResponse, bool, error) {
	var kind string
	var active *bool
	var parentID, memberUserID *uuid.UUID
	if req.Filter != nil {
		kind = req.Filter.Kind
		active = req.Filter.Active
		parentID = req.Filter.ParentID
		if req.Filter.MyOrgsOnly != nil && *req.Filter.MyOrgsOnly {
			uid := p.UserID
			memberUserID = &uid
		}
	}

	rows, err := s.repo.SearchOrgs(
		ctx,
		p.ParokiID,
		strings.TrimSpace(req.Q),
		kind,
		active,
		parentID,
		memberUserID,
		cursorCreatedAt,
		cursorID,
		limit+1,
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	orgIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		orgIDs = append(orgIDs, row.ID)
	}
	rolesMap, err := s.repo.ListActiveOrgRolesForUser(ctx, p.UserID, orgIDs)
	if err != nil {
		return nil, false, err
	}

	items := make([]domain.OrganizationResponse, 0, len(rows))
	for _, row := range rows {
		var myRole *string
		if r, ok := rolesMap[row.ID]; ok {
			myRole = &r
		}
		items = append(items, toOrgResponse(row, myRole))
	}
	return items, hasMore, nil
}

// GetOrganizationForCaller returns a single non-deleted Organization in p.ParokiID with myRole populated.
func (s *Service) GetOrganizationForCaller(ctx context.Context, p authx.Principal, orgID uuid.UUID) (domain.OrganizationResponse, error) {
	row, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, orgID)
	if ent.IsNotFound(err) {
		return domain.OrganizationResponse{}, domain.ErrOrgNotFound
	}
	if err != nil {
		return domain.OrganizationResponse{}, err
	}
	var myRole *string
	if m, err := s.repo.FindActiveOrgMember(ctx, orgID, p.UserID); err == nil {
		r := string(m.Role)
		myRole = &r
	} else if !ent.IsNotFound(err) {
		return domain.OrganizationResponse{}, err
	}
	return toOrgResponse(row, myRole), nil
}

// CreateOrganization creates an Organization (or reactivates a soft-deleted one with the same code).
func (s *Service) CreateOrganization(ctx context.Context, p authx.Principal, req domain.CreateOrganizationRequest) (domain.OrganizationResponse, error) {
	isParokiAdm, err := s.isParokiAdmin(p, p.ParokiID)
	if err != nil {
		return domain.OrganizationResponse{}, err
	}
	if !isParokiAdm {
		return domain.OrganizationResponse{}, domain.ErrForbidden
	}

	code, err := normalizeCode(req.Code)
	if err != nil {
		return domain.OrganizationResponse{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return domain.OrganizationResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
	}

	if err := s.validateOrgRefs(ctx, p.ParokiID, uuid.Nil, req.ParentID, req.WilayahID, req.LingkunganID); err != nil {
		return domain.OrganizationResponse{}, err
	}

	var initialAdminID uuid.UUID
	if req.InitialAdminUserID != nil && *req.InitialAdminUserID != uuid.Nil {
		u, err := s.repo.FindUserByID(ctx, *req.InitialAdminUserID)
		if ent.IsNotFound(err) {
			return domain.OrganizationResponse{}, domain.ValidationError("initialAdminUserId", "NOT_FOUND", "User not found.")
		}
		if err != nil {
			return domain.OrganizationResponse{}, err
		}
		initialAdminID = u.ID
	} else if email := normalizeEmail(req.InitialAdminEmail); email != "" {
		existingUser, err := s.repo.FindUserByEmail(ctx, email)
		if err == nil {
			initialAdminID = existingUser.ID
		} else if !ent.IsNotFound(err) {
			return domain.OrganizationResponse{}, err
		} else {
			if s.users == nil {
				return domain.OrganizationResponse{}, domain.ValidationError("initialAdminEmail", "UNAVAILABLE", "User onboarding is unavailable.")
			}
			adminName := strings.TrimSpace(req.InitialAdminName)
			if adminName == "" {
				adminName = email
			}
			uid, err := s.users.EnsureUser(ctx, email, adminName, "", "email", email)
			if err != nil {
				return domain.OrganizationResponse{}, err
			}
			initialAdminID = uid
		}
	}
	if initialAdminID != uuid.Nil && s.refs != nil {
		if err := s.refs.EnsureParokiMember(ctx, p.ParokiID, initialAdminID); err != nil {
			return domain.OrganizationResponse{}, err
		}
	}

	kind := organization.KindSection
	if req.Kind != "" {
		kind = organization.Kind(req.Kind)
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	in := repo.OrgInput{
		ParokiID:     p.ParokiID,
		ParentID:     req.ParentID,
		WilayahID:    req.WilayahID,
		LingkunganID: req.LingkunganID,
		Kind:         kind,
		Code:         code,
		Name:         name,
		Description:  trimOptString(req.Description),
		Active:       active,
	}

	var out *ent.Organization
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, findErr := txRepo.FindOrgByCodeIncludeDeleted(ctx, p.ParokiID, code)
		switch {
		case findErr == nil && existing.DeletedAt == nil:
			return domain.ErrCodeConflict
		case findErr == nil && existing.DeletedAt != nil:
			out, findErr = txRepo.ReactivateOrg(ctx, existing.ID, in)
			if findErr != nil {
				return findErr
			}
		case ent.IsNotFound(findErr):
			out, findErr = txRepo.CreateOrg(ctx, in)
			if findErr != nil {
				if ent.IsConstraintError(findErr) {
					return domain.ErrCodeConflict
				}
				return findErr
			}
		default:
			return findErr
		}

		if initialAdminID != uuid.Nil {
			existingMember, err := txRepo.FindOrgMember(ctx, out.ID, initialAdminID)
			switch {
			case ent.IsNotFound(err):
				if _, err := txRepo.CreateOrgMember(ctx, repo.OrgMemberInput{
					ParokiID: p.ParokiID,
					OrgID:    out.ID,
					UserID:   initialAdminID,
					Role:     orgmember.RoleAdmin,
					Status:   orgmember.StatusActive,
				}); err != nil {
					return err
				}
			case err == nil:
				adminRole := orgmember.RoleAdmin
				activeStatus := orgmember.StatusActive
				if _, err := txRepo.UpdateOrgMember(ctx, existingMember.ID, &adminRole, nil, nil, nil, nil, &activeStatus); err != nil {
					return err
				}
			default:
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.OrganizationResponse{}, err
	}

	var myRole *string
	if initialAdminID != uuid.Nil {
		if err := s.syncMemberCasbin(ctx, out.ID, initialAdminID, orgmember.RoleAdmin, orgmember.StatusActive); err != nil {
			return domain.OrganizationResponse{}, err
		}
		if initialAdminID == p.UserID {
			r := string(orgmember.RoleAdmin)
			myRole = &r
		}
	}
	return toOrgResponse(out, myRole), nil
}

// UpdateOrganization updates mutable fields of an Organization (requires org:admin or ref:paroki-admin).
func (s *Service) UpdateOrganization(ctx context.Context, p authx.Principal, id uuid.UUID, req domain.UpdateOrganizationRequest) (domain.OrganizationResponse, error) {
	existing, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, id)
	if ent.IsNotFound(err) {
		return domain.OrganizationResponse{}, domain.ErrOrgNotFound
	}
	if err != nil {
		return domain.OrganizationResponse{}, err
	}

	isOrgAdm, err := s.isOrgAdmin(p, id)
	if err != nil {
		return domain.OrganizationResponse{}, err
	}
	isParokiAdm, err := s.isParokiAdmin(p, p.ParokiID)
	if err != nil {
		return domain.OrganizationResponse{}, err
	}
	if !isOrgAdm && !isParokiAdm {
		return domain.OrganizationResponse{}, domain.ErrForbidden
	}

	if err := s.validateOrgRefs(ctx, p.ParokiID, id, req.ParentID, req.WilayahID, req.LingkunganID); err != nil {
		return domain.OrganizationResponse{}, err
	}

	var kindPtr *organization.Kind
	if req.Kind != nil {
		k := organization.Kind(*req.Kind)
		kindPtr = &k
	}
	var codePtr *string
	if req.Code != nil {
		normCode, err := normalizeCode(*req.Code)
		if err != nil {
			return domain.OrganizationResponse{}, err
		}
		if normCode != existing.Code {
			if _, findErr := s.repo.FindOrgByCodeIncludeDeleted(ctx, p.ParokiID, normCode); findErr == nil {
				return domain.OrganizationResponse{}, domain.ErrCodeConflict
			} else if !ent.IsNotFound(findErr) {
				return domain.OrganizationResponse{}, findErr
			}
		}
		codePtr = &normCode
	}
	var namePtr *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return domain.OrganizationResponse{}, domain.ValidationError("name", "REQUIRED", "Name cannot be empty.")
		}
		namePtr = &trimmed
	}

	updated, err := s.repo.UpdateOrg(
		ctx,
		id,
		req.ParentID,
		req.WilayahID,
		req.LingkunganID,
		kindPtr,
		codePtr,
		namePtr,
		trimOptString(req.Description),
		req.Active,
	)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.OrganizationResponse{}, domain.ErrCodeConflict
		}
		return domain.OrganizationResponse{}, err
	}

	var myRole *string
	if m, err := s.repo.FindActiveOrgMember(ctx, id, p.UserID); err == nil {
		r := string(m.Role)
		myRole = &r
	} else if !ent.IsNotFound(err) {
		return domain.OrganizationResponse{}, err
	}
	return toOrgResponse(updated, myRole), nil
}

// DeleteOrganization soft-deletes an Organization if it has no non-deleted child organizations or events.
func (s *Service) DeleteOrganization(ctx context.Context, p authx.Principal, id uuid.UUID) error {
	isParokiAdm, err := s.isParokiAdmin(p, p.ParokiID)
	if err != nil {
		return err
	}
	if !isParokiAdm {
		return domain.ErrForbidden
	}

	if _, err := s.repo.FindOrgInParoki(ctx, p.ParokiID, id); ent.IsNotFound(err) {
		return domain.ErrOrgNotFound
	} else if err != nil {
		return err
	}

	inUse, err := s.repo.HasChildOrgs(ctx, id)
	if err != nil {
		return err
	}
	if !inUse {
		inUse, err = s.repo.HasOrgEvents(ctx, id)
		if err != nil {
			return err
		}
	}
	if inUse {
		return domain.ErrInUse
	}
	return s.repo.SoftDeleteOrg(ctx, id, time.Now().UTC())
}

func (s *Service) validateOrgRefs(ctx context.Context, parokiID, selfID uuid.UUID, parentID, wilayahID, lingkunganID *uuid.UUID) error {
	if parentID != nil && *parentID != uuid.Nil {
		if selfID != uuid.Nil && *parentID == selfID {
			return domain.ValidationError("parentId", "INVALID_PARENT", "Organization cannot be its own parent.")
		}
		parent, err := s.repo.FindOrgInParoki(ctx, parokiID, *parentID)
		if ent.IsNotFound(err) {
			return domain.ErrOrgNotFound
		} else if err != nil {
			return err
		}
		if selfID != uuid.Nil {
			visited := map[uuid.UUID]struct{}{parent.ID: {}}
			cur := parent.ParentID
			for cur != nil && *cur != uuid.Nil {
				if *cur == selfID {
					return domain.ValidationError("parentId", "INVALID_PARENT", "Organization hierarchy cannot contain a cycle.")
				}
				if _, seen := visited[*cur]; seen {
					break
				}
				visited[*cur] = struct{}{}
				anc, err := s.repo.FindOrgInParoki(ctx, parokiID, *cur)
				if ent.IsNotFound(err) {
					break
				}
				if err != nil {
					return err
				}
				cur = anc.ParentID
			}
		}
	}
	if wilayahID != nil && *wilayahID != uuid.Nil && s.refs != nil {
		if err := s.refs.ValidateWilayahInParoki(ctx, parokiID, *wilayahID); err != nil {
			return err
		}
	}
	if lingkunganID != nil && *lingkunganID != uuid.Nil && s.refs != nil {
		if err := s.refs.ValidateLingkunganInParoki(ctx, parokiID, *lingkunganID); err != nil {
			return err
		}
	}
	return nil
}
