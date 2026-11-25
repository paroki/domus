package authx

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/httpx"
)

type ctxKey struct{}

// ErrTenantRequired is returned when an endpoint requires an active paroki tenant.
var ErrTenantRequired = httpx.NewError(fiber.StatusBadRequest, "TENANT_REQUIRED", "An active paroki tenant is required.")

// Principal represents an authenticated user across HTTP, MCP, and bot channels.
type Principal struct {
	UserID    uuid.UUID `json:"userId"`
	SessionID uuid.UUID `json:"sessionId,omitempty"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ParokiID  uuid.UUID `json:"parokiId,omitempty"`
	Roles     []string  `json:"roles,omitempty"`
	Channel   string    `json:"channel,omitempty"`
}

// Subject returns the Casbin subject identifier ("user:<uuid>").
func (p Principal) Subject() string {
	return "user:" + p.UserID.String()
}

// HasTenant reports whether a paroki tenant is selected on this Principal.
func (p Principal) HasTenant() bool {
	return p.ParokiID != uuid.Nil
}

// ParokiDomain returns the Casbin paroki domain ("paroki:<uuid>") or "" if unset.
func (p Principal) ParokiDomain() string {
	if p.ParokiID == uuid.Nil {
		return ""
	}
	return "paroki:" + p.ParokiID.String()
}

// WithPrincipal attaches p to ctx.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// FromContext extracts the Principal from ctx, if present.
func FromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

// UserContext extracts the Principal from ctx, if present.
func UserContext(ctx context.Context) (Principal, bool) {
	return FromContext(ctx)
}

// RequireTenant returns the authenticated Principal if it has an active ParokiID.
func RequireTenant(ctx context.Context) (Principal, error) {
	p, ok := FromContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return Principal{}, httpx.Unauthorized("Authentication is required.")
	}
	if p.ParokiID == uuid.Nil {
		return Principal{}, ErrTenantRequired
	}
	return p, nil
}
