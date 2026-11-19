package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/internal/user/repo"
)

// Config holds service-level authentication and session settings.
type Config struct {
	OTPTTL         time.Duration
	OTPMaxAttempts int
	OTPCooldown    time.Duration
	SessionTTL     time.Duration
	GoogleID       string
	GoogleSecret   string
	GitHubID       string
	GitHubSecret   string
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

// Service implements user identity, OTP, OAuth2, and session business logic.
type Service struct {
	cfg            Config
	repo           *repo.Repository
	otpSender      domain.OTPSender
	oauthProviders map[string]domain.OAuthProvider
}

// New constructs a Service with configured OAuth2 providers and OTP sender.
func New(cfg Config, r *repo.Repository, otpSender domain.OTPSender, logger *slog.Logger) *Service {
	if otpSender == nil {
		otpSender = NewLogOTPSender(logger)
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
		otpSender:      otpSender,
		oauthProviders: providers,
	}
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
	if provider == "" {
		provider = domain.ProviderEmail
	}
	if subject == "" {
		subject = normEmail
	}

	var out *domain.UserResponse
	err := s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		u, err := txRepo.FindUserByEmail(ctx, normEmail)
		if ent.IsNotFound(err) {
			var phonePtr *string
			if p := strings.TrimSpace(phone); p != "" {
				phonePtr = &p
			}
			u, err = txRepo.CreateUser(ctx, normEmail, strings.TrimSpace(name), phonePtr)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if u.Status == user.StatusDisabled {
			return domain.ErrUserDisabled
		}

		now := time.Now().UTC()
		if _, err := txRepo.UpsertAccount(ctx, u.ID, provider, subject, &u.Email, &u.Name, u.AvatarURL, now); err != nil {
			return err
		}
		res := toUserResponse(u)
		out = &res
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) createSession(ctx context.Context, r *repo.Repository, u *ent.User, userAgent, ipAddress string) (string, *domain.AuthSessionResponse, error) {
	rawToken, err := randomToken(32)
	if err != nil {
		return "", nil, err
	}
	expiresAt := time.Now().UTC().Add(s.cfg.SessionTTL)
	if _, err := r.CreateSession(ctx, u.ID, sha256Hex(rawToken), userAgent, ipAddress, u.DefaultParishID, expiresAt); err != nil {
		return "", nil, err
	}
	return rawToken, &domain.AuthSessionResponse{
		User:           toUserResponse(u),
		ActiveParishID: u.DefaultParishID,
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
