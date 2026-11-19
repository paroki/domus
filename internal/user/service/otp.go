package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"time"

	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// RequestOTP generates and dispatches a 6-digit OTP for an existing active user.
func (s *Service) RequestOTP(ctx context.Context, email string) (domain.RequestOTPResponse, error) {
	normEmail := normalizeEmail(email)
	u, err := guardActiveUser(s.repo.FindUserByEmail(ctx, normEmail))
	if err != nil {
		return domain.RequestOTPResponse{}, err
	}

	latest, err := s.repo.FindLatestOTP(ctx, domain.ProviderEmail, u.Email)
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

	if err := s.repo.DeleteOTPsByTarget(ctx, domain.ProviderEmail, u.Email); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.OTPTTL)
	if _, err := s.repo.CreateOTP(ctx, domain.ProviderEmail, u.Email, sha256Hex(code), expiresAt); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	if err := s.otpSender.SendOTP(ctx, domain.ProviderEmail, u.Email, code); err != nil {
		return domain.RequestOTPResponse{}, err
	}

	return domain.RequestOTPResponse{
		Message:          "Verification code sent.",
		ExpiresInSeconds: int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// VerifyOTP validates a 6-digit OTP and creates an authenticated UserSession.
func (s *Service) VerifyOTP(ctx context.Context, email, code, userAgent, ipAddress string) (string, *domain.AuthSessionResponse, error) {
	normEmail := normalizeEmail(email)
	u, err := guardActiveUser(s.repo.FindUserByEmail(ctx, normEmail))
	if err != nil {
		return "", nil, err
	}

	otpRow, err := s.repo.FindLatestOTP(ctx, domain.ProviderEmail, u.Email)
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

	candidateHash := sha256Hex(code)
	if subtle.ConstantTimeCompare([]byte(candidateHash), []byte(otpRow.CodeHash)) != 1 {
		updated, incErr := s.repo.IncrementOTPAttempts(ctx, otpRow.ID)
		if incErr == nil && updated.Attempts >= s.cfg.OTPMaxAttempts {
			_ = s.repo.DeleteOTP(ctx, otpRow.ID)
		}
		return "", nil, domain.ErrOTPInvalid
	}

	var rawToken string
	var resp *domain.AuthSessionResponse
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		if err := txRepo.DeleteOTP(ctx, otpRow.ID); err != nil {
			return err
		}
		if _, err := txRepo.UpsertAccount(ctx, u.ID, domain.ProviderEmail, u.Email, &u.Email, &u.Name, u.AvatarURL, now); err != nil {
			return err
		}
		updatedUser, err := txRepo.UpdateUserLastLogin(ctx, u.ID, now)
		if err != nil {
			return err
		}
		rawToken, resp, err = s.createSession(ctx, txRepo, updatedUser, userAgent, ipAddress)
		return err
	})
	if err != nil {
		return "", nil, err
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
