package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// RequestOTP generates and dispatches a 6-digit OTP for an existing active user (or email guest invite onboarding).
func (s *Service) RequestOTP(ctx context.Context, req domain.RequestOTPRequest) (domain.RequestOTPResponse, error) {
	channel, rawTarget, targetField, err := resolveOTPChannelAndTarget(req.Channel, req.Target, req.Email)
	if err != nil {
		return domain.RequestOTPResponse{}, err
	}
	if !s.otpSender.HasChannel(channel) {
		return domain.RequestOTPResponse{}, domain.ErrAuthChannelDisabled
	}

	inviteCode := strings.TrimSpace(req.InviteCode)
	if inviteCode != "" && channel != domain.ProviderEmail {
		return domain.RequestOTPResponse{}, domain.ValidationError("channel", "INVALID_CHANNEL", "Invite code onboarding is only supported over email.")
	}
	if inviteCode != "" && s.inviteJoiner != nil {
		if err := s.inviteJoiner.ValidateInviteCode(ctx, inviteCode); err != nil {
			return domain.RequestOTPResponse{}, err
		}
	}

	_, _, otpTarget, err := s.resolveOTPRecipient(ctx, channel, rawTarget, targetField, inviteCode != "" && s.inviteJoiner != nil)
	if err != nil {
		return domain.RequestOTPResponse{}, err
	}

	latest, err := s.repo.FindLatestOTP(ctx, channel, otpTarget)
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

	if err := s.repo.DeleteOTPsByTarget(ctx, channel, otpTarget); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.OTPTTL)
	if _, err := s.repo.CreateOTP(ctx, channel, otpTarget, sha256Hex(code), expiresAt); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	if err := s.otpSender.SendOTP(ctx, channel, otpTarget, code); err != nil {
		return domain.RequestOTPResponse{}, err
	}

	return domain.RequestOTPResponse{
		Message:          "Verification code sent.",
		ExpiresInSeconds: int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// VerifyOTP validates a 6-digit OTP and creates an authenticated UserSession (and redeems inviteCode if present).
func (s *Service) VerifyOTP(ctx context.Context, req domain.VerifyOTPRequest, userAgent, ipAddress string) (string, *domain.AuthSessionResponse, error) {
	channel, rawTarget, targetField, err := resolveOTPChannelAndTarget(req.Channel, req.Target, req.Email)
	if err != nil {
		return "", nil, err
	}
	if !s.otpSender.HasChannel(channel) {
		return "", nil, domain.ErrAuthChannelDisabled
	}

	inviteCode := strings.TrimSpace(req.InviteCode)
	if inviteCode != "" && channel != domain.ProviderEmail {
		return "", nil, domain.ValidationError("channel", "INVALID_CHANNEL", "Invite code onboarding is only supported over email.")
	}
	if inviteCode != "" && s.inviteJoiner != nil {
		if err := s.inviteJoiner.ValidateInviteCode(ctx, inviteCode); err != nil {
			return "", nil, err
		}
	}

	allowMissingEmail := inviteCode != "" && s.inviteJoiner != nil
	u, matchedAcc, otpTarget, err := s.resolveOTPRecipient(ctx, channel, rawTarget, targetField, allowMissingEmail)
	if err != nil {
		return "", nil, err
	}

	otpRow, err := s.verifyStoredOTP(ctx, channel, otpTarget, req.Code)
	if err != nil {
		return "", nil, err
	}

	now := time.Now().UTC()
	if allowMissingEmail {
		var phoneStr string
		if req.Phone != nil {
			normPhone, err := normalizePhone(*req.Phone)
			if err != nil {
				return "", nil, err
			}
			phoneStr = normPhone
		}
		ensured, err := s.EnsureUser(ctx, otpTarget, req.Name, phoneStr, domain.ProviderEmail, otpTarget)
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
		if channel == domain.ProviderEmail {
			if _, err := txRepo.UpsertAccount(ctx, u.ID, domain.ProviderEmail, u.Email, &u.Email, &u.Name, u.AvatarURL, now); err != nil {
				return err
			}
		} else if matchedAcc != nil {
			if err := txRepo.UpdateAccountLastUsed(ctx, matchedAcc.ID, now); err != nil {
				return err
			}
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

	if allowMissingEmail {
		parishID, err := s.inviteJoiner.RedeemInviteForUser(ctx, inviteCode, updatedUser.ID, sessionID)
		if err != nil {
			return "", nil, err
		}
		reloaded, err := s.repo.FindUserByID(ctx, updatedUser.ID)
		if err == nil {
			updatedUser = reloaded
		}
		resp, err = s.buildAuthSessionResponse(ctx, s.repo, updatedUser, &parishID, resp.ExpiresAt)
		if err != nil {
			return "", nil, err
		}
	}
	return rawToken, resp, nil
}

func resolveOTPChannelAndTarget(rawChannel, rawTarget, rawEmail string) (channel, target, targetField string, err error) {
	channel = strings.ToLower(strings.TrimSpace(rawChannel))
	if channel == "" {
		channel = domain.ProviderEmail
	}
	switch channel {
	case domain.ProviderEmail, domain.ProviderWhatsApp, domain.ProviderTelegram:
	default:
		return "", "", "", domain.ValidationError("channel", "INVALID_CHANNEL", "OTP channel must be email, whatsapp, or telegram.")
	}

	t := strings.TrimSpace(rawTarget)
	e := strings.TrimSpace(rawEmail)
	if t != "" && e != "" && t != e {
		return "", "", "", domain.ValidationError("target", "CONFLICTING_TARGET", "Target and email must not differ when both are provided.")
	}
	if t != "" {
		return channel, t, "target", nil
	}
	if e != "" {
		return channel, e, "email", nil
	}
	return "", "", "", domain.ValidationError("target", "REQUIRED", "Target or email is required.")
}

func (s *Service) resolveOTPRecipient(
	ctx context.Context,
	channel, rawTarget, targetField string,
	allowMissingEmail bool,
) (*ent.User, *ent.UserAccount, string, error) {
	switch channel {
	case domain.ProviderEmail:
		addr, err := mail.ParseAddress(rawTarget)
		if err != nil || !strings.Contains(addr.Address, "@") || !strings.EqualFold(addr.Address, rawTarget) {
			return nil, nil, "", domain.ValidationError(targetField, "EMAIL", "Must be a valid email address.")
		}
		normEmail := normalizeEmail(addr.Address)
		u, err := s.repo.FindUserByEmail(ctx, normEmail)
		switch {
		case err == nil:
			if u.Status == user.StatusDisabled {
				return nil, nil, "", domain.ErrUserDisabled
			}
			return u, nil, normEmail, nil
		case ent.IsNotFound(err) && allowMissingEmail:
			return nil, nil, normEmail, nil
		case ent.IsNotFound(err):
			return nil, nil, "", domain.ErrUserNotFound
		default:
			return nil, nil, "", err
		}

	case domain.ProviderWhatsApp:
		e164, err := normalizePhoneTarget(rawTarget)
		if err != nil {
			return nil, nil, "", err
		}
		acc, err := s.repo.FindAccountByProviderSubject(ctx, domain.ProviderWhatsApp, e164)
		if ent.IsNotFound(err) {
			return nil, nil, "", domain.ErrUserNotFound
		}
		if err != nil {
			return nil, nil, "", err
		}
		u, err := guardActiveUser(s.repo.FindUserByID(ctx, acc.UserID))
		if err != nil {
			return nil, nil, "", err
		}
		return u, acc, e164, nil

	case domain.ProviderTelegram:
		var u *ent.User
		if strings.Contains(rawTarget, "@") {
			var err error
			u, err = guardActiveUser(s.repo.FindUserByEmail(ctx, normalizeEmail(rawTarget)))
			if err != nil {
				return nil, nil, "", err
			}
		} else {
			e164, err := normalizePhoneTarget(rawTarget)
			if err != nil {
				return nil, nil, "", err
			}
			waAcc, err := s.repo.FindAccountByProviderSubject(ctx, domain.ProviderWhatsApp, e164)
			if ent.IsNotFound(err) {
				return nil, nil, "", domain.ErrUserNotFound
			}
			if err != nil {
				return nil, nil, "", err
			}
			u, err = guardActiveUser(s.repo.FindUserByID(ctx, waAcc.UserID))
			if err != nil {
				return nil, nil, "", err
			}
		}
		tgAcc, err := s.repo.FindUserAccountByUserAndProvider(ctx, u.ID, domain.ProviderTelegram)
		if ent.IsNotFound(err) {
			return nil, nil, "", domain.ErrUserNotFound
		}
		if err != nil {
			return nil, nil, "", err
		}
		return u, tgAcc, tgAcc.Subject, nil

	default:
		return nil, nil, "", domain.ValidationError("channel", "INVALID_CHANNEL", "OTP channel must be email, whatsapp, or telegram.")
	}
}

func normalizePhoneTarget(raw string) (string, error) {
	e164, err := normalizePhone(raw)
	if err != nil || e164 == "" {
		return "", domain.ValidationError("target", "INVALID_PHONE", "Phone number must be a valid E.164 number.")
	}
	return e164, nil
}

func (s *Service) verifyStoredOTP(ctx context.Context, channel, target, rawCode string) (*ent.UserOTP, error) {
	otpRow, err := s.repo.FindLatestOTP(ctx, channel, target)
	if ent.IsNotFound(err) {
		return nil, domain.ErrOTPInvalid
	}
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if now.After(otpRow.ExpiresAt) {
		_ = s.repo.DeleteOTP(ctx, otpRow.ID)
		return nil, domain.ErrOTPExpired
	}

	candidateHash := sha256Hex(rawCode)
	if subtle.ConstantTimeCompare([]byte(candidateHash), []byte(otpRow.CodeHash)) != 1 {
		updated, incErr := s.repo.IncrementOTPAttempts(ctx, otpRow.ID)
		if incErr == nil && updated.Attempts >= s.cfg.OTPMaxAttempts {
			_ = s.repo.DeleteOTP(ctx, otpRow.ID)
		}
		return nil, domain.ErrOTPInvalid
	}
	return otpRow, nil
}

func generateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
