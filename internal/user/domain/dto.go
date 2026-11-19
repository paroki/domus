package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuthProvidersResponse lists enabled OAuth2 provider names.
type AuthProvidersResponse struct {
	Providers []string `json:"providers"`
}

// RequestOTPRequest requests a one-time passcode for passwordless email login.
type RequestOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// RequestOTPResponse acknowledges OTP dispatch and its TTL in seconds.
type RequestOTPResponse struct {
	Message          string `json:"message"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// VerifyOTPRequest verifies a 6-digit OTP code for an email address.
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required,len=6"`
}

// UserResponse is the public API representation of a User.
type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Phone           *string    `json:"phone,omitempty"`
	AvatarURL       *string    `json:"avatarUrl,omitempty"`
	Status          string     `json:"status"`
	DefaultParishID *uuid.UUID `json:"defaultParishId,omitempty"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// AuthSessionResponse is returned upon successful OTP verification.
type AuthSessionResponse struct {
	User           UserResponse `json:"user"`
	ActiveParishID *uuid.UUID   `json:"activeParishId,omitempty"`
	ExpiresAt      time.Time    `json:"expiresAt"`
}
