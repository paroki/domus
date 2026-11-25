package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationResponse is the public API representation of an Organization.
type OrganizationResponse struct {
	ID           uuid.UUID  `json:"id"`
	ParokiID     uuid.UUID  `json:"parokiId"`
	ParentID     *uuid.UUID `json:"parentId,omitempty"`
	WilayahID    *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID *uuid.UUID `json:"lingkunganId,omitempty"`
	Kind         string     `json:"kind"`
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	Description  *string    `json:"description,omitempty"`
	Active       bool       `json:"active"`
	MyRole       *string    `json:"myRole,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// OrganizationSearchFilter filters organization search results.
type OrganizationSearchFilter struct {
	Kind       string     `json:"kind,omitempty" validate:"omitempty,oneof=dpp bgkp section categorical wilayah lingkungan community"`
	Active     *bool      `json:"active,omitempty"`
	ParentID   *uuid.UUID `json:"parentId,omitempty"`
	MyOrgsOnly *bool      `json:"myOrgsOnly,omitempty"`
}

// SearchOrganizationsRequest searches organizations within the active paroki.
type SearchOrganizationsRequest struct {
	Q      string                    `json:"q,omitempty"`
	Filter *OrganizationSearchFilter `json:"filter,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
	Limit  int                       `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateOrganizationRequest creates a new Organization in the active paroki.
type CreateOrganizationRequest struct {
	ParentID           *uuid.UUID `json:"parentId,omitempty"`
	WilayahID          *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID       *uuid.UUID `json:"lingkunganId,omitempty"`
	Kind               string     `json:"kind,omitempty" validate:"omitempty,oneof=dpp bgkp section categorical wilayah lingkungan community"`
	Code               string     `json:"code" validate:"required"`
	Name               string     `json:"name" validate:"required,min=1"`
	Description        *string    `json:"description,omitempty"`
	Active             *bool      `json:"active,omitempty"`
	InitialAdminUserID *uuid.UUID `json:"initialAdminUserId,omitempty"`
	InitialAdminEmail  string     `json:"initialAdminEmail,omitempty" validate:"omitempty,email"`
	InitialAdminName   string     `json:"initialAdminName,omitempty"`
}

// UpdateOrganizationRequest updates an existing Organization.
type UpdateOrganizationRequest struct {
	ParentID     *uuid.UUID `json:"parentId,omitempty"`
	WilayahID    *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID *uuid.UUID `json:"lingkunganId,omitempty"`
	Kind         *string    `json:"kind,omitempty" validate:"omitempty,oneof=dpp bgkp section categorical wilayah lingkungan community"`
	Code         *string    `json:"code,omitempty"`
	Name         *string    `json:"name,omitempty" validate:"omitempty,min=1"`
	Description  *string    `json:"description,omitempty"`
	Active       *bool      `json:"active,omitempty"`
}

// OrgMemberResponse is the public API representation of an OrgMember joined with user info.
type OrgMemberResponse struct {
	ID        uuid.UUID  `json:"id"`
	ParokiID  uuid.UUID  `json:"parokiId"`
	OrgID     uuid.UUID  `json:"orgId"`
	UserID    uuid.UUID  `json:"userId"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Phone     *string    `json:"phone,omitempty"`
	Role      string     `json:"role"`
	Position  *string    `json:"position,omitempty"`
	PeriodID  *uuid.UUID `json:"periodId,omitempty"`
	StartDate *string    `json:"startDate,omitempty"`
	EndDate   *string    `json:"endDate,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// OrgMemberSearchFilter filters organization member search results.
type OrgMemberSearchFilter struct {
	Role     string     `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
	Status   string     `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	PeriodID *uuid.UUID `json:"periodId,omitempty"`
}

// SearchOrgMembersRequest searches members of an Organization.
type SearchOrgMembersRequest struct {
	Q      string                 `json:"q,omitempty"`
	Filter *OrgMemberSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// AddOrgMemberRequest adds an existing or onboarded user as a member of an Organization.
type AddOrgMemberRequest struct {
	UserID    *uuid.UUID `json:"userId,omitempty"`
	Email     string     `json:"email,omitempty" validate:"omitempty,email"`
	Name      string     `json:"name,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Role      string     `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
	Position  *string    `json:"position,omitempty"`
	PeriodID  *uuid.UUID `json:"periodId,omitempty"`
	StartDate *string    `json:"startDate,omitempty"`
	EndDate   *string    `json:"endDate,omitempty"`
	Status    string     `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// UpdateOrgMemberRequest updates an organization member's role, position, term, or status.
type UpdateOrgMemberRequest struct {
	Role      *string    `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
	Position  *string    `json:"position,omitempty"`
	PeriodID  *uuid.UUID `json:"periodId,omitempty"`
	StartDate *string    `json:"startDate,omitempty"`
	EndDate   *string    `json:"endDate,omitempty"`
	Status    *string    `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// OrgInviteResponse is the public API representation of an OrgInvite.
type OrgInviteResponse struct {
	ID              uuid.UUID  `json:"id"`
	ParokiID        uuid.UUID  `json:"parokiId"`
	OrgID           uuid.UUID  `json:"orgId"`
	Code            string     `json:"code"`
	Role            string     `json:"role"`
	DefaultPosition *string    `json:"defaultPosition,omitempty"`
	MaxUses         *int       `json:"maxUses,omitempty"`
	UsedCount       int        `json:"usedCount"`
	ExpiresAt       time.Time  `json:"expiresAt"`
	RevokedAt       *time.Time `json:"revokedAt,omitempty"`
	CreatedBy       uuid.UUID  `json:"createdBy"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// OrgInvitePreviewResponse is the public preview representation of an OrgInvite.
type OrgInvitePreviewResponse struct {
	Code            string    `json:"code"`
	OrgID           uuid.UUID `json:"orgId"`
	OrgCode         string    `json:"orgCode"`
	OrgName         string    `json:"orgName"`
	OrgKind         string    `json:"orgKind"`
	ParokiID        uuid.UUID `json:"parokiId"`
	ParokiCode      string    `json:"parokiCode"`
	ParokiName      string    `json:"parokiName"`
	Role            string    `json:"role"`
	DefaultPosition *string   `json:"defaultPosition,omitempty"`
	ExpiresAt       time.Time `json:"expiresAt"`
	Status          string    `json:"status"`
	Valid           bool      `json:"valid"`
}

// CreateOrgInviteRequest creates a new time-limited multi-use invite link for an Organization.
type CreateOrgInviteRequest struct {
	Role            string  `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
	DefaultPosition *string `json:"defaultPosition,omitempty"`
	MaxUses         *int    `json:"maxUses,omitempty" validate:"omitempty,min=1"`
	TTLSeconds      *int    `json:"ttlSeconds,omitempty"`
}

// OrgInviteSearchFilter filters organization invite search results.
type OrgInviteSearchFilter struct {
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active expired revoked exhausted"`
}

// SearchOrgInvitesRequest searches invite links for an Organization.
type SearchOrgInvitesRequest struct {
	Filter *OrgInviteSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// RedeemOrgInviteResponse is returned upon successful invite redemption.
type RedeemOrgInviteResponse struct {
	Organization OrganizationResponse `json:"organization"`
	Membership   OrgMemberResponse    `json:"membership"`
	ParokiID     uuid.UUID            `json:"parokiId"`
}
