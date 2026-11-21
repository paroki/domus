package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/mailx"
	"github.com/paroki/domus/internal/platform/telegram"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// Config holds service-level authentication and session settings.
type Config struct {
	SuperAdminEmail string
	SuperAdminName  string
	OTPTTL          time.Duration
	OTPMaxAttempts  int
	OTPCooldown     time.Duration
	SessionTTL      time.Duration
	GoogleID        string
	GoogleSecret    string
	GitHubID        string
	GitHubSecret    string
}

// LogOTPSender is the default OTPSender that logs dispatched codes via slog.
type LogOTPSender struct {
	logger *slog.Logger
}

// NewLogOTPSender returns a LogOTPSender using logger (or slog.Default if nil).
func NewLogOTPSender(logger *slog.Logger) *LogOTPSender {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogOTPSender{logger: logger}
}

// SendOTP logs the OTP target and code for development and testing.
func (s *LogOTPSender) SendOTP(ctx context.Context, channel, target, code string) error {
	s.logger.InfoContext(ctx, "otp dispatched", "channel", channel, "target", target, "code", code)
	return nil
}

// MailOTPSender delivers email OTP codes via mailx.Mailer.
type MailOTPSender struct {
	mailer *mailx.Mailer
	otpTTL time.Duration
}

// NewMailOTPSender constructs a MailOTPSender backed by mailer.
func NewMailOTPSender(mailer *mailx.Mailer, otpTTL time.Duration) *MailOTPSender {
	return &MailOTPSender{mailer: mailer, otpTTL: otpTTL}
}

// SendOTP sends the 6-digit OTP code to target over email.
func (s *MailOTPSender) SendOTP(ctx context.Context, channel, target, code string) error {
	if channel != "" && channel != domain.ProviderEmail {
		return fmt.Errorf("mailx otp sender: unsupported channel %q", channel)
	}
	ttlMins := max(1, int(s.otpTTL.Minutes()))
	body := fmt.Sprintf(
		"Kode OTP login Domus Anda adalah: %s\n\nKode ini berlaku selama %d menit. Jangan berikan kode ini kepada siapa pun.",
		code,
		ttlMins,
	)
	return s.mailer.Send(ctx, mailx.Message{
		To:       target,
		Subject:  "Kode Verifikasi Login Domus",
		TextBody: body,
	})
}

// InviteJoiner validates and redeems organization invite codes during guest auth onboarding.
type InviteJoiner interface {
	ValidateInviteCode(ctx context.Context, code string) error
	RedeemInviteForUser(ctx context.Context, code string, userID, sessionID uuid.UUID) (parishID uuid.UUID, err error)
}

// Service implements user identity, OTP, OAuth2, and session business logic.
type Service struct {
	cfg            Config
	repo           *repo.Repository
	enforcer       *authz.Enforcer
	otpSender      *CompositeOTPSender
	telegramClient *telegram.Client
	oauthProviders map[string]domain.OAuthProvider
	inviteJoiner   InviteJoiner
}

// SetInviteJoiner wires the organization invite joiner into the auth service.
func (s *Service) SetInviteJoiner(j InviteJoiner) {
	s.inviteJoiner = j
}

// SetTelegramClient wires the optional Telegram Bot API client into the service.
func (s *Service) SetTelegramClient(c *telegram.Client) {
	s.telegramClient = c
}

// New constructs a Service with configured OAuth2 providers and OTP sender.
func New(cfg Config, r *repo.Repository, enforcer *authz.Enforcer, otpSender domain.OTPSender, logger *slog.Logger) *Service {
	var composite *CompositeOTPSender
	if c, ok := otpSender.(*CompositeOTPSender); ok && c != nil {
		composite = c
	} else {
		composite = NewCompositeOTPSender(logger)
		if otpSender != nil {
			composite.SetSender(domain.ProviderEmail, otpSender)
		}
	}
	providers := make(map[string]domain.OAuthProvider)
	if cfg.GoogleID != "" && cfg.GoogleSecret != "" {
		providers[domain.ProviderGoogle] = newGoogleProvider(cfg.GoogleID, cfg.GoogleSecret)
	}
	if cfg.GitHubID != "" && cfg.GitHubSecret != "" {
		providers[domain.ProviderGitHub] = newGitHubProvider(cfg.GitHubID, cfg.GitHubSecret)
	}
	return &Service{
		cfg:            cfg,
		repo:           r,
		enforcer:       enforcer,
		otpSender:      composite,
		oauthProviders: providers,
	}
}

// EnabledOTPChannels returns enabled login OTP channel names in deterministic order.
func (s *Service) EnabledOTPChannels() []string {
	return s.otpSender.EnabledChannels()
}

// RegisterOAuthProvider registers or overrides an OAuth2 provider driver (useful for tests/extensions).
func (s *Service) RegisterOAuthProvider(p domain.OAuthProvider) {
	if p != nil {
		s.oauthProviders[p.Name()] = p
	}
}

// Logout revokes the session associated with rawToken.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return s.repo.DeleteSessionByTokenHash(ctx, sha256Hex(rawToken))
}

// EnsureUser finds or creates a user for invite-based onboarding and links the given provider account.
func (s *Service) EnsureUser(ctx context.Context, email, name, phone, provider, subject string) (*domain.UserResponse, error) {
	normEmail := normalizeEmail(email)
	if normEmail == "" {
		return nil, domain.ErrUserNotFound
	}
	if strings.TrimSpace(name) == "" {
		name = normEmail
	}
	normPhone, err := normalizePhone(phone)
	if err != nil {
		return nil, err
	}
	if provider == "" {
		provider = domain.ProviderEmail
	}
	normSubject, err := normalizeChannelSubject(provider, subject)
	if err != nil {
		return nil, err
	}
	if normSubject == "" {
		if provider == domain.ProviderWhatsApp && normPhone != "" {
			normSubject = normPhone
		} else {
			normSubject = normEmail
		}
	}

	var phonePtr *string
	if normPhone != "" {
		phonePtr = &normPhone
	}

	var out *domain.UserResponse
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		u, err := txRepo.FindUserByEmailIncludeDeleted(ctx, normEmail)
		switch {
		case ent.IsNotFound(err):
			u, err = txRepo.CreateUser(ctx, normEmail, strings.TrimSpace(name), phonePtr)
			if err != nil {
				return err
			}
		case err != nil:
			return err
		case u.DeletedAt != nil:
			u, err = txRepo.ReactivateUser(ctx, u.ID, strings.TrimSpace(name), phonePtr)
			if err != nil {
				return err
			}
		case u.Status == user.StatusDisabled:
			return domain.ErrUserDisabled
		}

		now := time.Now().UTC()
		if _, err := txRepo.UpsertAccount(ctx, u.ID, provider, normSubject, &u.Email, &u.Name, u.AvatarURL, now); err != nil {
			return err
		}
		res := toUserResponse(u)
		out = &res
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.enforcer != nil {
		if err := s.enforcer.AssignRole(ctx, authz.UserSubject(out.ID), domain.RoleUser, authz.DomainGlobal); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Service) createSessionWithID(ctx context.Context, r *repo.Repository, u *ent.User, userAgent, ipAddress string) (string, uuid.UUID, *domain.AuthSessionResponse, error) {
	rawToken, err := randomToken(32)
	if err != nil {
		return "", uuid.Nil, nil, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.SessionTTL)
	sess, err := r.CreateSession(ctx, u.ID, sha256Hex(rawToken), userAgent, ipAddress, u.DefaultParishID, expiresAt)
	if err != nil {
		return "", uuid.Nil, nil, err
	}
	resp, err := s.buildAuthSessionResponse(u, u.DefaultParishID, expiresAt)
	if err != nil {
		return "", uuid.Nil, nil, err
	}
	return rawToken, sess.ID, resp, nil
}

func (s *Service) buildAuthSessionResponse(u *ent.User, parishIDPtr *uuid.UUID, expiresAt time.Time) (*domain.AuthSessionResponse, error) {
	userResp := toUserResponse(u)
	var activeParishID uuid.UUID
	if parishIDPtr != nil {
		activeParishID = *parishIDPtr
	}
	roles, err := s.loadRoles(u.ID, activeParishID)
	if err != nil {
		return nil, err
	}
	userResp.Roles = roles
	userResp.ActiveParishID = parishIDPtr
	return &domain.AuthSessionResponse{
		User:           userResp,
		ActiveParishID: parishIDPtr,
		ExpiresAt:      expiresAt,
	}, nil
}

func guardActiveUser(u *ent.User, err error) (*ent.User, error) {
	if ent.IsNotFound(err) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if u.Status == user.StatusDisabled {
		return nil, domain.ErrUserDisabled
	}
	return u, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func randomToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func toUserResponse(u *ent.User) domain.UserResponse {
	return domain.UserResponse{
		ID:              u.ID,
		Email:           u.Email,
		Name:            u.Name,
		Phone:           u.Phone,
		AvatarURL:       u.AvatarURL,
		Status:          string(u.Status),
		DefaultParishID: u.DefaultParishID,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
