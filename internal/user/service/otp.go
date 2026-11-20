package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// RequestOTP generates and dispatches a 6-digit OTP for an existing active user or guest invite onboarding.
func (s *Service) RequestOTP(ctx context.Context, req domain.RequestOTPRequest) (domain.RequestOTPResponse, error) {
	normEmail := normalizeEmail(req.Email)
	inviteCode := strings.TrimSpace(req.InviteCode)

	if inviteCode != "" && s.inviteJoiner != nil {
		if err := s.inviteJoiner.ValidateInviteCode(ctx, inviteCode); err != nil {
			return domain.RequestOTPResponse{}, err
		}
	}

	u, err := s.repo.FindUserByEmail(ctx, normEmail)
	switch {
	case err == nil:
		if u.Status == user.StatusDisabled {
			return domain.RequestOTPResponse{}, domain.ErrUserDisabled
		}
	case ent.IsNotFound(err):
		if inviteCode == "" || s.inviteJoiner == nil {
			return domain.RequestOTPResponse{}, domain.ErrUserNotFound
		}
	default:
		return domain.RequestOTPResponse{}, err
	}

	latest, err := s.repo.FindLatestOTP(ctx, domain.ProviderEmail, normEmail)
	if err == nil && time.Since(latest.CreatedAt) < s.cfg.OTPCooldown {
		return domain.RequestOTPResponse{}, domain.ErrOTPRateLimited
	}
	if err != nil && !ent.IsNotFound(err) {
		return domain.RequestOTPResponse{}, err
	}

	code, err := generateOTPCode()
	if err != nil {
		return domain.RequestOTPResponse{}, err
	}

	if err := s.repo.DeleteOTPsByTarget(ctx, domain.ProviderEmail, normEmail); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.OTPTTL)
	if _, err := s.repo.CreateOTP(ctx, domain.ProviderEmail, normEmail, sha256Hex(code), expiresAt); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	if err := s.otpSender.SendOTP(ctx, domain.ProviderEmail, normEmail, code); err != nil {
		return domain.RequestOTPResponse{}, err
	}

	return domain.RequestOTPResponse{
		Message:          "Verification code sent.",
		ExpiresInSeconds: int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// VerifyOTP validates a 6-digit OTP and creates an authenticated UserSession (and redeems inviteCode if present).
func (s *Service) VerifyOTP(ctx context.Context, req domain.VerifyOTPRequest, userAgent, ipAddress string) (string, *domain.AuthSessionResponse, error) {
	normEmail := normalizeEmail(req.Email)
	inviteCode := strings.TrimSpace(req.InviteCode)

	if inviteCode != "" && s.inviteJoiner != nil {
		if err := s.inviteJoiner.ValidateInviteCode(ctx, inviteCode); err != nil {
			return "", nil, err
		}
	}

	var u *ent.User
	if inviteCode == "" || s.inviteJoiner == nil {
		var err error
		u, err = guardActiveUser(s.repo.FindUserByEmail(ctx, normEmail))
		if err != nil {
			return "", nil, err
		}
	}

	otpRow, err := s.repo.FindLatestOTP(ctx, domain.ProviderEmail, normEmail)
	if ent.IsNotFound(err) {
		return "", nil, domain.ErrOTPInvalid
	}
	if err != nil {
		return "", nil, err
	}

	now := time.Now().UTC()
	if now.After(otpRow.ExpiresAt) {
		_ = s.repo.DeleteOTP(ctx, otpRow.ID)
		return "", nil, domain.ErrOTPExpired
	}

	candidateHash := sha256Hex(req.Code)
	if subtle.ConstantTimeCompare([]byte(candidateHash), []byte(otpRow.CodeHash)) != 1 {
		updated, incErr := s.repo.IncrementOTPAttempts(ctx, otpRow.ID)
		if incErr == nil && updated.Attempts >= s.cfg.OTPMaxAttempts {
			_ = s.repo.DeleteOTP(ctx, otpRow.ID)
		}
		return "", nil, domain.ErrOTPInvalid
	}

	if inviteCode != "" && s.inviteJoiner != nil {
		var phoneStr string
		if req.Phone != nil {
			normPhone, err := normalizePhone(*req.Phone)
			if err != nil {
				return "", nil, err
			}
			phoneStr = normPhone
		}
		ensured, err := s.EnsureUser(ctx, normEmail, req.Name, phoneStr, domain.ProviderEmail, normEmail)
		if err != nil {
			return "", nil, err
		}
		u, err = guardActiveUser(s.repo.FindUserByID(ctx, ensured.ID))
		if err != nil {
			return "", nil, err
		}
	}

	var rawToken string
	var sessionID uuid.UUID
	var resp *domain.AuthSessionResponse
	var updatedUser *ent.User
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if err := txRepo.DeleteOTP(ctx, otpRow.ID); err != nil {
			return err
		}
		if _, err := txRepo.UpsertAccount(ctx, u.ID, domain.ProviderEmail, u.Email, &u.Email, &u.Name, u.AvatarURL, now); err != nil {
			return err
		}
		var updErr error
		updatedUser, updErr = txRepo.UpdateUserLastLogin(ctx, u.ID, now)
		if updErr != nil {
			return updErr
		}
		rawToken, sessionID, resp, updErr = s.createSessionWithID(ctx, txRepo, updatedUser, userAgent, ipAddress)
		return updErr
	})
	if err != nil {
		return "", nil, err
	}

	if inviteCode != "" && s.inviteJoiner != nil {
		parishID, err := s.inviteJoiner.RedeemInviteForUser(ctx, inviteCode, updatedUser.ID, sessionID)
		if err != nil {
			return "", nil, err
		}
		reloaded, err := s.repo.FindUserByID(ctx, updatedUser.ID)
		if err == nil {
			updatedUser = reloaded
		}
		resp, err = s.buildAuthSessionResponse(updatedUser, &parishID, resp.ExpiresAt)
		if err != nil {
			return "", nil, err
		}
	}
	return rawToken, resp, nil
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
