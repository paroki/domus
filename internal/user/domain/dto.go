package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
)

// AuthProvidersResponse lists enabled OAuth2 provider names and OTP channels.
type AuthProvidersResponse struct {
	Providers   []string `json:"providers"`
	OTPChannels []string `json:"otpChannels"`
}

// RequestOTPRequest requests a one-time passcode for passwordless login over email, WhatsApp, or Telegram.
type RequestOTPRequest struct {
	Channel    string `json:"channel,omitempty"`
	Target     string `json:"target,omitempty"`
	Email      string `json:"email,omitempty"`
	InviteCode string `json:"inviteCode,omitempty"`
}

// RequestOTPResponse acknowledges OTP dispatch and its TTL in seconds.
type RequestOTPResponse struct {
	Message          string `json:"message"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// VerifyOTPRequest verifies a 6-digit OTP code for email, WhatsApp, or Telegram login.
type VerifyOTPRequest struct {
	Channel    string  `json:"channel,omitempty"`
	Target     string  `json:"target,omitempty"`
	Email      string  `json:"email,omitempty"`
	Code       string  `json:"code" validate:"required,len=6"`
	InviteCode string  `json:"inviteCode,omitempty"`
	Name       string  `json:"name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
}

// UserResponse is the public API representation of a User.
type UserResponse struct {
	ID              uuid.UUID              `json:"id"`
	Email           string                 `json:"email"`
	Name            string                 `json:"name"`
	Phone           *string                `json:"phone,omitempty"`
	Gender          *string                `json:"gender,omitempty"`
	BirthDate       *string                `json:"birthDate,omitempty"`
	AvatarURL       *string                `json:"avatarUrl,omitempty"`
	Status          string                 `json:"status"`
	Onboarded       bool                   `json:"onboarded"`
	LinkedProviders []string               `json:"linkedProviders,omitempty"`
	ActiveParokiID  *uuid.UUID             `json:"activeParokiId,omitempty"`
	DefaultParokiID *uuid.UUID             `json:"defaultParokiId,omitempty"`
	Roles           []string               `json:"roles,omitempty"`
	RoleDetails     []authz.RoleDescriptor `json:"roleDetails,omitempty"`
	LastLoginAt     *time.Time             `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// AuthSessionResponse is returned upon successful OTP verification.
type AuthSessionResponse struct {
	User           UserResponse `json:"user"`
	ActiveParokiID *uuid.UUID   `json:"activeParokiId,omitempty"`
	ExpiresAt      time.Time    `json:"expiresAt"`
}

// UpdateMeRequest updates the authenticated user's profile fields.
type UpdateMeRequest struct {
	Name            *string    `json:"name,omitempty" validate:"omitempty,min=1"`
	Phone           *string    `json:"phone,omitempty"`
	Gender          *string    `json:"gender,omitempty"`
	BirthDate       *string    `json:"birthDate,omitempty"`
	AvatarURL       *string    `json:"avatarUrl,omitempty" validate:"omitempty,url"`
	DefaultParokiID *uuid.UUID `json:"defaultParokiId,omitempty"`
	Onboarded       *bool      `json:"onboarded,omitempty"`
}

// LinkAccountOTPRequest requests a 6-digit OTP to link a WhatsApp account to the current user.
type LinkAccountOTPRequest struct {
	Channel string `json:"channel" validate:"required,oneof=whatsapp"`
	Target  string `json:"target" validate:"required"`
}

// VerifyLinkAccountRequest verifies a 6-digit OTP to link a WhatsApp account to the current user.
type VerifyLinkAccountRequest struct {
	Channel string `json:"channel" validate:"required,oneof=whatsapp"`
	Target  string `json:"target" validate:"required"`
	Code    string `json:"code" validate:"required,len=6"`
}

// TelegramLinkTokenResponse returns a single-use deep link to connect Telegram to the current user.
type TelegramLinkTokenResponse struct {
	DeepLinkURL      string `json:"deepLinkUrl"`
	BotUsername      string `json:"botUsername"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// SwitchSessionParokiRequest updates the active paroki on the current web session.
type SwitchSessionParokiRequest struct {
	ActiveParokiID uuid.UUID `json:"activeParokiId" validate:"required"`
}

// UserSessionResponse is the public API representation of an active UserSession (never exposes token_hash).
type UserSessionResponse struct {
	ID             uuid.UUID  `json:"id"`
	UserAgent      string     `json:"userAgent"`
	IPAddress      string     `json:"ipAddress"`
	ActiveParokiID *uuid.UUID `json:"activeParokiId,omitempty"`
	IsCurrent      bool       `json:"isCurrent"`
	LastActiveAt   time.Time  `json:"lastActiveAt"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// SearchSessionsRequest holds cursor pagination parameters for listing user sessions.
type SearchSessionsRequest struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// UserAccountResponse is the public API representation of a linked UserAccount.
type UserAccountResponse struct {
	ID             uuid.UUID  `json:"id"`
	Provider       string     `json:"provider"`
	Subject        string     `json:"subject"`
	Email          *string    `json:"email,omitempty"`
	DisplayName    *string    `json:"displayName,omitempty"`
	AvatarURL      *string    `json:"avatarUrl,omitempty"`
	ActiveParokiID *uuid.UUID `json:"activeParokiId,omitempty"`
	LastUsedAt     *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// SearchAccountsRequest holds cursor pagination parameters for listing linked accounts.
type SearchAccountsRequest struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateUserTokenRequest creates a new Personal Access Token for the current user.
type CreateUserTokenRequest struct {
	Name          string `json:"name" validate:"required,min=1,max=100"`
	ExpiresInDays *int   `json:"expiresInDays,omitempty" validate:"omitempty,min=1,max=365"`
}

// UserTokenResponse is the public API representation of a UserToken (never exposes token_hash or plaintext token).
type UserTokenResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"tokenPrefix"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// CreateUserTokenResponse is returned once on token creation and includes the plaintext token.
type CreateUserTokenResponse struct {
	UserTokenResponse
	Token string `json:"token"`
}

// SearchTokensRequest holds cursor pagination parameters for listing Personal Access Tokens.
type SearchTokensRequest struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// UserSearchFilter filters the admin user directory search.
type UserSearchFilter struct {
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active disabled"`
}

// SearchUsersRequest searches the user directory with optional query, filter, and cursor pagination.
type SearchUsersRequest struct {
	Q      string            `json:"q,omitempty"`
	Filter *UserSearchFilter `json:"filter,omitempty"`
	Cursor string            `json:"cursor,omitempty"`
	Limit  int               `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// AdminCreateUserRequest creates a new user (and optionally appoints Super Admin).
type AdminCreateUserRequest struct {
	Email        string  `json:"email" validate:"required,email"`
	Name         string  `json:"name" validate:"required,min=1"`
	Phone        *string `json:"phone,omitempty"`
	IsSuperAdmin bool    `json:"isSuperAdmin"`
}

// AdminUpdateUserRequest updates user profile, status, or Super Admin role.
type AdminUpdateUserRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Phone        *string `json:"phone,omitempty"`
	Status       *string `json:"status,omitempty" validate:"omitempty,oneof=active disabled"`
	IsSuperAdmin *bool   `json:"isSuperAdmin,omitempty"`
}
