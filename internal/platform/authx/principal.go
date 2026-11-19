package authx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

// Principal represents an authenticated user across HTTP, MCP, and bot channels.
type Principal struct {
	UserID   uuid.UUID `json:"userId"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	ParishID uuid.UUID `json:"parishId,omitempty"`
	Roles    []string  `json:"roles,omitempty"`
	Channel  string    `json:"channel,omitempty"`
}

// Subject returns the Casbin subject identifier ("user:<uuid>").
func (p Principal) Subject() string {
	return "user:" + p.UserID.String()
}

// HasTenant reports whether a parish tenant is selected on this Principal.
func (p Principal) HasTenant() bool {
	return p.ParishID != uuid.Nil
}

// ParishDomain returns the Casbin parish domain ("parish:<uuid>") or "" if unset.
func (p Principal) ParishDomain() string {
	if p.ParishID == uuid.Nil {
		return ""
	}
	return "parish:" + p.ParishID.String()
}

// WithPrincipal attaches p to ctx.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// UserContext extracts the Principal from ctx, if present.
func UserContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}
