package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// RequestLinkAccountOTP sends a 6-digit verification OTP to link a WhatsApp number to the current user.
func (s *Service) RequestLinkAccountOTP(ctx context.Context, p authx.Principal, req domain.LinkAccountOTPRequest) (domain.RequestOTPResponse, error) {
	if !s.otpSender.HasChannel(domain.ProviderWhatsApp) {
		return domain.RequestOTPResponse{}, domain.ErrAuthChannelDisabled
	}

	e164, err := normalizePhoneTarget(req.Target)
	if err != nil {
		return domain.RequestOTPResponse{}, err
	}

	existing, err := s.repo.FindAccountByProviderSubject(ctx, domain.ProviderWhatsApp, e164)
	if err == nil && existing.UserID != p.UserID {
		return domain.RequestOTPResponse{}, domain.ErrUserAccountConflict
	}
	if err != nil && !ent.IsNotFound(err) {
		return domain.RequestOTPResponse{}, err
	}

	scopedTarget := p.UserID.String() + ":" + e164
	latest, err := s.repo.FindLatestOTP(ctx, channelWhatsAppLink, scopedTarget)
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

	if err := s.repo.DeleteOTPsByTarget(ctx, channelWhatsAppLink, scopedTarget); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.OTPTTL)
	if _, err := s.repo.CreateOTP(ctx, channelWhatsAppLink, scopedTarget, sha256Hex(code), expiresAt); err != nil {
		return domain.RequestOTPResponse{}, err
	}
	if err := s.otpSender.SendOTP(ctx, channelWhatsAppLink, e164, code); err != nil {
		return domain.RequestOTPResponse{}, err
	}

	return domain.RequestOTPResponse{
		Message:          "Verification code sent.",
		ExpiresInSeconds: int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// VerifyLinkAccount verifies a WhatsApp OTP and links the verified WhatsApp account to the current user.
func (s *Service) VerifyLinkAccount(ctx context.Context, p authx.Principal, req domain.VerifyLinkAccountRequest) (domain.UserAccountResponse, error) {
	if !s.otpSender.HasChannel(domain.ProviderWhatsApp) {
		return domain.UserAccountResponse{}, domain.ErrAuthChannelDisabled
	}

	e164, err := normalizePhoneTarget(req.Target)
	if err != nil {
		return domain.UserAccountResponse{}, err
	}

	scopedTarget := p.UserID.String() + ":" + e164
	otpRow, err := s.verifyStoredOTP(ctx, channelWhatsAppLink, scopedTarget, req.Code)
	if err != nil {
		return domain.UserAccountResponse{}, err
	}

	now := time.Now().UTC()
	var linked *ent.UserAccount
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, accErr := txRepo.FindAccountByProviderSubject(ctx, domain.ProviderWhatsApp, e164)
		if accErr == nil && existing.UserID != p.UserID {
			return domain.ErrUserAccountConflict
		}
		if accErr != nil && !ent.IsNotFound(accErr) {
			return accErr
		}
		if err := txRepo.DeleteOTP(ctx, otpRow.ID); err != nil {
			return err
		}
		var repErr error
		linked, repErr = txRepo.ReplaceChannelAccount(ctx, p.UserID, domain.ProviderWhatsApp, e164, nil, now)
		if repErr != nil {
			return repErr
		}
		return txRepo.SetUserPhone(ctx, p.UserID, e164)
	})
	if err != nil {
		return domain.UserAccountResponse{}, err
	}
	return toUserAccountResponse(linked), nil
}

// CreateTelegramLinkToken creates a single-use deep-link token for connecting Telegram via /start.
func (s *Service) CreateTelegramLinkToken(ctx context.Context, p authx.Principal) (domain.TelegramLinkTokenResponse, error) {
	if s.telegramClient == nil {
		return domain.TelegramLinkTokenResponse{}, domain.ErrAuthChannelDisabled
	}

	token, err := randomToken(24)
	if err != nil {
		return domain.TelegramLinkTokenResponse{}, err
	}

	target := p.UserID.String()
	if err := s.repo.DeleteOTPsByTarget(ctx, domain.ProviderTelegram, target); err != nil {
		return domain.TelegramLinkTokenResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.OTPTTL)
	if _, err := s.repo.CreateOTP(ctx, domain.ProviderTelegram, target, sha256Hex(token), expiresAt); err != nil {
		return domain.TelegramLinkTokenResponse{}, err
	}

	return domain.TelegramLinkTokenResponse{
		DeepLinkURL:      s.telegramClient.DeepLinkURL(token),
		BotUsername:      s.telegramClient.BotUsername(),
		ExpiresInSeconds: int(s.cfg.OTPTTL.Seconds()),
	}, nil
}

// CompleteTelegramLink consumes a Telegram deep-link token and links chatID to the owning user.
func (s *Service) CompleteTelegramLink(ctx context.Context, token, chatID, displayName string) error {
	trimmedToken := strings.TrimSpace(token)
	if trimmedToken == "" {
		return domain.ErrOTPInvalid
	}

	otpRow, err := s.repo.FindOTPByChannelAndHash(ctx, domain.ProviderTelegram, sha256Hex(trimmedToken))
	if ent.IsNotFound(err) {
		return domain.ErrOTPInvalid
	}
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if now.After(otpRow.ExpiresAt) {
		_ = s.repo.DeleteOTP(ctx, otpRow.ID)
		return domain.ErrOTPExpired
	}

	userID, err := uuid.Parse(otpRow.Target)
	if err != nil {
		return domain.ErrOTPInvalid
	}
	if _, err := guardActiveUser(s.repo.FindUserByID(ctx, userID)); err != nil {
		return err
	}

	cleanChatID := strings.TrimSpace(chatID)
	var displayNamePtr *string
	if trimmedName := strings.TrimSpace(displayName); trimmedName != "" {
		displayNamePtr = &trimmedName
	}

	return s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, accErr := txRepo.FindAccountByProviderSubject(ctx, domain.ProviderTelegram, cleanChatID)
		if accErr == nil && existing.UserID != userID {
			return domain.ErrUserAccountConflict
		}
		if accErr != nil && !ent.IsNotFound(accErr) {
			return accErr
		}
		if err := txRepo.DeleteOTP(ctx, otpRow.ID); err != nil {
			return err
		}
		_, err := txRepo.ReplaceChannelAccount(ctx, userID, domain.ProviderTelegram, cleanChatID, displayNamePtr, now)
		return err
	})
}
