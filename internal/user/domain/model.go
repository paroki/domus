package domain

import "context"

// Role constants owned by the user feature.
const (
	RoleSuperAdmin = "user:super-admin"
	RoleUser       = "user:user"
)

// Provider constants for UserAccount.
const (
	ProviderEmail    = "email"
	ProviderGoogle   = "google"
	ProviderGitHub   = "github"
	ProviderTelegram = "telegram"
	ProviderWhatsApp = "whatsapp"
)

// OTPSender delivers one-time passcodes over a channel (e.g. email, whatsapp, telegram).
type OTPSender interface {
	SendOTP(ctx context.Context, channel, target, code string) error
}

// OAuthUserInfo represents normalized profile fields returned by an OAuth2 provider.
type OAuthUserInfo struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

// OAuthProvider abstracts an OAuth2 login driver (e.g. google, github).
type OAuthProvider interface {
	Name() string
	AuthCodeURL(state, redirectURI string) string
	Exchange(ctx context.Context, code, redirectURI string) (OAuthUserInfo, error)
}
