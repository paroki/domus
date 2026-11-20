package service

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/orginvite"
	"github.com/paroki/domus/ent/orgmember"
	"github.com/paroki/domus/internal/org/domain"
	"github.com/paroki/domus/internal/org/repo"
	"github.com/paroki/domus/internal/platform/authx"
)

const inviteCodeAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// CreateOrgInvite creates a time-limited multi-use invite link for orgID (requires org:admin).
func (s *Service) CreateOrgInvite(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.CreateOrgInviteRequest,
) (domain.OrgInviteResponse, error) {
	isOrgAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return domain.OrgInviteResponse{}, err
	}
	if !isOrgAdm {
		return domain.OrgInviteResponse{}, domain.ErrForbidden
	}

	orgRow, err := s.repo.FindOrgInParish(ctx, p.ParishID, orgID)
	if ent.IsNotFound(err) {
		return domain.OrgInviteResponse{}, domain.ErrOrgNotFound
	}
	if err != nil {
		return domain.OrgInviteResponse{}, err
	}

	ttl := s.cfg.DefaultTTL
	if req.TTLSeconds != nil && *req.TTLSeconds > 0 {
		if *req.TTLSeconds > int(s.cfg.MaxTTL.Seconds()) {
			return domain.OrgInviteResponse{}, domain.ValidationError("ttlSeconds", "MAX_TTL_EXCEEDED", "Invite TTL exceeds maximum allowed duration.")
		}
		ttl = time.Duration(*req.TTLSeconds) * time.Second
	}

	role := orginvite.RoleMember
	if req.Role != "" {
		role = orginvite.Role(req.Role)
	}

	code, err := generateInviteCode()
	if err != nil {
		return domain.OrgInviteResponse{}, err
	}

	inv, err := s.repo.CreateOrgInvite(ctx, repo.OrgInviteInput{
		ParishID:        orgRow.ParishID,
		OrgID:           orgID,
		Code:            code,
		Role:            role,
		DefaultPosition: trimOptString(req.DefaultPosition),
		MaxUses:         req.MaxUses,
		ExpiresAt:       time.Now().UTC().Add(ttl),
		CreatedBy:       p.UserID,
	})
	if err != nil {
		return domain.OrgInviteResponse{}, err
	}
	return toOrgInviteResponse(inv), nil
}

// SearchOrgInvites returns paginated invite links for orgID (requires org:admin).
func (s *Service) SearchOrgInvites(
	ctx context.Context,
	p authx.Principal,
	orgID uuid.UUID,
	req domain.SearchOrgInvitesRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.OrgInviteResponse, bool, error) {
	isOrgAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return nil, false, err
	}
	if !isOrgAdm {
		return nil, false, domain.ErrForbidden
	}

	if _, err := s.repo.FindOrgInParish(ctx, p.ParishID, orgID); ent.IsNotFound(err) {
		return nil, false, domain.ErrOrgNotFound
	} else if err != nil {
		return nil, false, err
	}

	var status string
	if req.Filter != nil {
		status = req.Filter.Status
	}
	rows, err := s.repo.SearchOrgInvites(ctx, orgID, status, time.Now().UTC(), cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	items := make([]domain.OrgInviteResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOrgInviteResponse(row))
	}
	return items, hasMore, nil
}

// RevokeOrgInvite idempotently revokes an invite link belonging to orgID (requires org:admin).
func (s *Service) RevokeOrgInvite(ctx context.Context, p authx.Principal, orgID, inviteID uuid.UUID) error {
	isOrgAdm, err := s.isOrgAdmin(p, orgID)
	if err != nil {
		return err
	}
	if !isOrgAdm {
		return domain.ErrForbidden
	}

	if _, err := s.repo.FindOrgInParish(ctx, p.ParishID, orgID); ent.IsNotFound(err) {
		return domain.ErrOrgNotFound
	} else if err != nil {
		return err
	}

	inv, err := s.repo.FindOrgInviteInOrg(ctx, orgID, inviteID)
	if ent.IsNotFound(err) {
		return domain.ErrInviteNotFound
	}
	if err != nil {
		return err
	}
	if inv.RevokedAt != nil {
		return nil
	}
	return s.repo.RevokeOrgInvite(ctx, inv.ID, time.Now().UTC())
}

// PreviewOrgInvite returns public preview metadata for an invite code (even when expired/revoked/exhausted).
func (s *Service) PreviewOrgInvite(ctx context.Context, code string) (domain.OrgInvitePreviewResponse, error) {
	normCode := strings.TrimSpace(code)
	if normCode == "" {
		return domain.OrgInvitePreviewResponse{}, domain.ErrInviteNotFound
	}
	inv, err := s.repo.FindOrgInviteByCode(ctx, normCode)
	if ent.IsNotFound(err) {
		return domain.OrgInvitePreviewResponse{}, domain.ErrInviteNotFound
	}
	if err != nil {
		return domain.OrgInvitePreviewResponse{}, err
	}

	orgRow, parishRow, err := s.repo.FindActiveOrgAndParish(ctx, inv.OrgID, inv.ParishID)
	if ent.IsNotFound(err) {
		return domain.OrgInvitePreviewResponse{}, domain.ErrInviteNotFound
	}
	if err != nil {
		return domain.OrgInvitePreviewResponse{}, err
	}

	status := computeInviteStatus(inv, time.Now().UTC())
	return domain.OrgInvitePreviewResponse{
		Code:            inv.Code,
		OrgID:           orgRow.ID,
		OrgCode:         orgRow.Code,
		OrgName:         orgRow.Name,
		OrgKind:         string(orgRow.Kind),
		ParishID:        parishRow.ID,
		ParishCode:      parishRow.Code,
		ParishName:      parishRow.Name,
		Role:            string(inv.Role),
		DefaultPosition: inv.DefaultPosition,
		ExpiresAt:       inv.ExpiresAt,
		Status:          status,
		Valid:           status == "active",
	}, nil
}

// ValidateInviteCode verifies that code refers to an active, non-expired, non-revoked, non-exhausted invite.
func (s *Service) ValidateInviteCode(ctx context.Context, code string) error {
	_, _, _, err := s.validateUsableInvite(ctx, s.repo, code, time.Now().UTC())
	return err
}

// RedeemInvite redeems an invite code for the authenticated principal p.
func (s *Service) RedeemInvite(ctx context.Context, p authx.Principal, code string) (domain.RedeemOrgInviteResponse, error) {
	return s.redeemInviteInternal(ctx, code, p.UserID, p.SessionID)
}

// RedeemInviteForUser redeems an invite code for userID and switches sessionID's active parish if non-nil.
func (s *Service) RedeemInviteForUser(ctx context.Context, code string, userID, sessionID uuid.UUID) (uuid.UUID, error) {
	res, err := s.redeemInviteInternal(ctx, code, userID, sessionID)
	if err != nil {
		return uuid.Nil, err
	}
	return res.ParishID, nil
}

func (s *Service) redeemInviteInternal(ctx context.Context, code string, userID, sessionID uuid.UUID) (domain.RedeemOrgInviteResponse, error) {
	inv, orgRow, _, err := s.validateUsableInvite(ctx, s.repo, code, time.Now().UTC())
	if err != nil {
		return domain.RedeemOrgInviteResponse{}, err
	}

	if s.refs != nil {
		if err := s.refs.EnsureParishMember(ctx, inv.ParishID, userID); err != nil {
			return domain.RedeemOrgInviteResponse{}, err
		}
	}

	var savedMember *ent.OrgMember
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		txInv, _, _, err := s.validateUsableInvite(ctx, txRepo, code, time.Now().UTC())
		if err != nil {
			return err
		}

		existing, findErr := txRepo.FindOrgMember(ctx, txInv.OrgID, userID)
		switch {
		case findErr == nil && existing.Status == orgmember.StatusActive:
			var rolePtr *orgmember.Role
			if txInv.Role == orginvite.RoleAdmin && existing.Role != orgmember.RoleAdmin {
				r := orgmember.RoleAdmin
				rolePtr = &r
			}
			var posPtr *string
			if existing.Position == nil && txInv.DefaultPosition != nil {
				posPtr = txInv.DefaultPosition
			}
			if rolePtr != nil || posPtr != nil {
				upd, updErr := txRepo.UpdateOrgMember(ctx, existing.ID, rolePtr, posPtr, nil, nil, nil, nil)
				if updErr != nil {
					return updErr
				}
				savedMember = upd
			} else {
				savedMember = existing
			}
			return nil

		case findErr == nil && existing.Status == orgmember.StatusInactive:
			activeStatus := orgmember.StatusActive
			newRole := existing.Role
			if txInv.Role == orginvite.RoleAdmin || existing.Role != orgmember.RoleAdmin {
				newRole = orgmember.Role(txInv.Role)
			}
			var posPtr *string
			if existing.Position == nil && txInv.DefaultPosition != nil {
				posPtr = txInv.DefaultPosition
			}
			upd, updErr := txRepo.UpdateOrgMember(ctx, existing.ID, &newRole, posPtr, nil, nil, nil, &activeStatus)
			if updErr != nil {
				return updErr
			}
			savedMember = upd
			return txRepo.IncrementOrgInviteUsedCount(ctx, txInv.ID, txInv.MaxUses)

		case ent.IsNotFound(findErr):
			created, createErr := txRepo.CreateOrgMember(ctx, repo.OrgMemberInput{
				ParishID: txInv.ParishID,
				OrgID:    txInv.OrgID,
				UserID:   userID,
				Role:     orgmember.Role(txInv.Role),
				Position: txInv.DefaultPosition,
				Status:   orgmember.StatusActive,
			})
			if createErr != nil {
				return createErr
			}
			savedMember = created
			return txRepo.IncrementOrgInviteUsedCount(ctx, txInv.ID, txInv.MaxUses)

		default:
			return findErr
		}
	})
	if err != nil {
		return domain.RedeemOrgInviteResponse{}, err
	}

	if err := s.syncMemberCasbin(ctx, inv.OrgID, userID, savedMember.Role, orgmember.StatusActive); err != nil {
		return domain.RedeemOrgInviteResponse{}, err
	}

	if sessionID != uuid.Nil && s.sessions != nil {
		if err := s.sessions.SwitchSessionParish(ctx, userID, sessionID, inv.ParishID); err != nil {
			return domain.RedeemOrgInviteResponse{}, err
		}
	}

	u, err := s.repo.FindUserByID(ctx, userID)
	if err != nil && !ent.IsNotFound(err) {
		return domain.RedeemOrgInviteResponse{}, err
	}
	myRole := string(savedMember.Role)
	return domain.RedeemOrgInviteResponse{
		Organization: toOrgResponse(orgRow, &myRole),
		Membership:   toOrgMemberResponse(savedMember, u),
		ParishID:     inv.ParishID,
	}, nil
}

func (s *Service) validateUsableInvite(ctx context.Context, r *repo.Repository, code string, now time.Time) (*ent.OrgInvite, *ent.Organization, *ent.Parish, error) {
	normCode := strings.TrimSpace(code)
	if normCode == "" {
		return nil, nil, nil, domain.ErrInviteNotFound
	}
	inv, err := r.FindOrgInviteByCode(ctx, normCode)
	if ent.IsNotFound(err) {
		return nil, nil, nil, domain.ErrInviteNotFound
	}
	if err != nil {
		return nil, nil, nil, err
	}
	orgRow, parishRow, err := r.FindActiveOrgAndParish(ctx, inv.OrgID, inv.ParishID)
	if ent.IsNotFound(err) {
		return nil, nil, nil, domain.ErrInviteNotFound
	}
	if err != nil {
		return nil, nil, nil, err
	}
	if inv.RevokedAt != nil {
		return nil, nil, nil, domain.ErrInviteRevoked
	}
	if !now.Before(inv.ExpiresAt) {
		return nil, nil, nil, domain.ErrInviteExpired
	}
	if inv.MaxUses != nil && inv.UsedCount >= *inv.MaxUses {
		return nil, nil, nil, domain.ErrInviteMaxUsesReached
	}
	return inv, orgRow, parishRow, nil
}

func computeInviteStatus(inv *ent.OrgInvite, now time.Time) string {
	switch {
	case inv.RevokedAt != nil:
		return "revoked"
	case !now.Before(inv.ExpiresAt):
		return "expired"
	case inv.MaxUses != nil && inv.UsedCount >= *inv.MaxUses:
		return "exhausted"
	default:
		return "active"
	}
}

func generateInviteCode() (string, error) {
	const length = 12
	maxIdx := big.NewInt(int64(len(inviteCodeAlphabet)))
	buf := make([]byte, length)
	for i := range buf {
		n, err := rand.Int(rand.Reader, maxIdx)
		if err != nil {
			return "", err
		}
		buf[i] = inviteCodeAlphabet[n.Int64()]
	}
	return string(buf), nil
}

func toOrgInviteResponse(inv *ent.OrgInvite) domain.OrgInviteResponse {
	return domain.OrgInviteResponse{
		ID:              inv.ID,
		ParishID:        inv.ParishID,
		OrgID:           inv.OrgID,
		Code:            inv.Code,
		Role:            string(inv.Role),
		DefaultPosition: inv.DefaultPosition,
		MaxUses:         inv.MaxUses,
		UsedCount:       inv.UsedCount,
		ExpiresAt:       inv.ExpiresAt,
		RevokedAt:       inv.RevokedAt,
		CreatedBy:       inv.CreatedBy,
		CreatedAt:       inv.CreatedAt,
		UpdatedAt:       inv.UpdatedAt,
	}
}
