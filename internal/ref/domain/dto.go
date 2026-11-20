package domain

import (
	"time"

	"github.com/google/uuid"
)

// DioceseResponse is the public API representation of a Diocese.
type DioceseResponse struct {
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	BishopName *string   `json:"bishopName,omitempty"`
	Address    *string   `json:"address,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	Email      *string   `json:"email,omitempty"`
	Website    *string   `json:"website,omitempty"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// DioceseSearchFilter filters diocese search results.
type DioceseSearchFilter struct {
	Active *bool `json:"active,omitempty"`
}

// SearchDiocesesRequest searches dioceses with optional query, filter, and cursor pagination.
type SearchDiocesesRequest struct {
	Q      string               `json:"q,omitempty"`
	Filter *DioceseSearchFilter `json:"filter,omitempty"`
	Cursor string               `json:"cursor,omitempty"`
	Limit  int                  `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateDioceseRequest creates a new Diocese.
type CreateDioceseRequest struct {
	Code       string  `json:"code" validate:"required"`
	Name       string  `json:"name" validate:"required,min=1"`
	BishopName *string `json:"bishopName,omitempty"`
	Address    *string `json:"address,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email"`
	Website    *string `json:"website,omitempty"`
	Active     *bool   `json:"active,omitempty"`
}

// UpdateDioceseRequest updates an existing Diocese.
type UpdateDioceseRequest struct {
	Code       *string `json:"code,omitempty"`
	Name       *string `json:"name,omitempty" validate:"omitempty,min=1"`
	BishopName *string `json:"bishopName,omitempty"`
	Address    *string `json:"address,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email"`
	Website    *string `json:"website,omitempty"`
	Active     *bool   `json:"active,omitempty"`
}

// ParishResponse is the public API representation of a Parish.
type ParishResponse struct {
	ID          uuid.UUID `json:"id"`
	DioceseID   uuid.UUID `json:"dioceseId"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	PatronSaint *string   `json:"patronSaint,omitempty"`
	PastorName  *string   `json:"pastorName,omitempty"`
	Address     *string   `json:"address,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Email       *string   `json:"email,omitempty"`
	Website     *string   `json:"website,omitempty"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	Letterhead  *string   `json:"letterhead,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ParishSearchFilter filters parish search results.
type ParishSearchFilter struct {
	DioceseID  *uuid.UUID `json:"dioceseId,omitempty"`
	Active     *bool      `json:"active,omitempty"`
	MemberOnly *bool      `json:"memberOnly,omitempty"`
}

// SearchParishesRequest searches parishes with optional query, filter, and cursor pagination.
type SearchParishesRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *ParishSearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateParishRequest creates a new Parish.
type CreateParishRequest struct {
	DioceseID   uuid.UUID `json:"dioceseId" validate:"required"`
	Code        string    `json:"code" validate:"required"`
	Name        string    `json:"name" validate:"required,min=1"`
	PatronSaint *string   `json:"patronSaint,omitempty"`
	PastorName  *string   `json:"pastorName,omitempty"`
	Address     *string   `json:"address,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
	Website     *string   `json:"website,omitempty"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	Letterhead  *string   `json:"letterhead,omitempty"`
	Active      *bool     `json:"active,omitempty"`
}

// UpdateParishRequest updates a Parish (super-admin fields or parish profile).
type UpdateParishRequest struct {
	DioceseID   *uuid.UUID `json:"dioceseId,omitempty"`
	Code        *string    `json:"code,omitempty"`
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=1"`
	PatronSaint *string    `json:"patronSaint,omitempty"`
	PastorName  *string    `json:"pastorName,omitempty"`
	Address     *string    `json:"address,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	Email       *string    `json:"email,omitempty" validate:"omitempty,email"`
	Website     *string    `json:"website,omitempty"`
	LogoURL     *string    `json:"logoUrl,omitempty"`
	Letterhead  *string    `json:"letterhead,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}

// UpdateParishProfileRequest updates mutable profile fields of the current active Parish.
type UpdateParishProfileRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	PatronSaint *string `json:"patronSaint,omitempty"`
	PastorName  *string `json:"pastorName,omitempty"`
	Address     *string `json:"address,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty" validate:"omitempty,email"`
	Website     *string `json:"website,omitempty"`
	LogoURL     *string `json:"logoUrl,omitempty"`
	Letterhead  *string `json:"letterhead,omitempty"`
}

// ParishMemberResponse is the public API representation of a ParishMember joined with user info.
type ParishMemberResponse struct {
	ID        uuid.UUID `json:"id"`
	ParishID  uuid.UUID `json:"parishId"`
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ParishMemberSearchFilter filters parish member search results.
type ParishMemberSearchFilter struct {
	Role string `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
}

// SearchParishMembersRequest searches members of a Parish.
type SearchParishMembersRequest struct {
	Q      string                    `json:"q,omitempty"`
	Filter *ParishMemberSearchFilter `json:"filter,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
	Limit  int                       `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// AddParishMemberRequest adds an existing or invited user as a member of a Parish.
type AddParishMemberRequest struct {
	UserID *uuid.UUID `json:"userId,omitempty"`
	Email  string     `json:"email,omitempty" validate:"omitempty,email"`
	Name   string     `json:"name,omitempty"`
	Phone  *string    `json:"phone,omitempty"`
	Role   string     `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
}

// UpdateParishMemberRequest updates a parish member's role.
type UpdateParishMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

// ZoneResponse is the public API representation of a Zone.
type ZoneResponse struct {
	ID               uuid.UUID `json:"id"`
	ParishID         uuid.UUID `json:"parishId"`
	Type             string    `json:"type"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	CoordinatorName  *string   `json:"coordinatorName,omitempty"`
	CoordinatorPhone *string   `json:"coordinatorPhone,omitempty"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// WardResponse is the public API representation of a Ward.
type WardResponse struct {
	ID          uuid.UUID `json:"id"`
	ParishID    uuid.UUID `json:"parishId"`
	ZoneID      uuid.UUID `json:"zoneId"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	PatronSaint *string   `json:"patronSaint,omitempty"`
	HeadName    *string   `json:"headName,omitempty"`
	HeadPhone   *string   `json:"headPhone,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// LookupResponse is the public API representation of a Lookup.
type LookupResponse struct {
	ID        uuid.UUID  `json:"id"`
	ParishID  *uuid.UUID `json:"parishId,omitempty"`
	Category  string     `json:"category"`
	Code      string     `json:"code"`
	Label     string     `json:"label"`
	SortOrder int        `json:"sortOrder"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// PeriodResponse is the public API representation of a Period.
type PeriodResponse struct {
	ID        uuid.UUID `json:"id"`
	ParishID  uuid.UUID `json:"parishId"`
	Type      string    `json:"type"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	StartDate string    `json:"startDate"`
	EndDate   string    `json:"endDate"`
	Status    string    `json:"status"`
	IsCurrent bool      `json:"isCurrent"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ZoneSearchFilter filters zone search results.
type ZoneSearchFilter struct {
	Type   string `json:"type,omitempty" validate:"omitempty,oneof=zone station"`
	Active *bool  `json:"active,omitempty"`
}

// SearchZonesRequest searches zones within the active parish.
type SearchZonesRequest struct {
	Q      string            `json:"q,omitempty"`
	Filter *ZoneSearchFilter `json:"filter,omitempty"`
	Cursor string            `json:"cursor,omitempty"`
	Limit  int               `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateZoneRequest creates a new Zone or Station in the active parish.
type CreateZoneRequest struct {
	Type             string  `json:"type,omitempty" validate:"omitempty,oneof=zone station"`
	Code             string  `json:"code" validate:"required"`
	Name             string  `json:"name" validate:"required,min=1"`
	CoordinatorName  *string `json:"coordinatorName,omitempty"`
	CoordinatorPhone *string `json:"coordinatorPhone,omitempty"`
}

// UpdateZoneRequest updates an existing Zone or Station.
type UpdateZoneRequest struct {
	Type             *string `json:"type,omitempty" validate:"omitempty,oneof=zone station"`
	Code             *string `json:"code,omitempty"`
	Name             *string `json:"name,omitempty" validate:"omitempty,min=1"`
	CoordinatorName  *string `json:"coordinatorName,omitempty"`
	CoordinatorPhone *string `json:"coordinatorPhone,omitempty"`
	Active           *bool   `json:"active,omitempty"`
}

// WardSearchFilter filters ward search results.
type WardSearchFilter struct {
	ZoneID *uuid.UUID `json:"zoneId,omitempty"`
	Active *bool      `json:"active,omitempty"`
}

// SearchWardsRequest searches wards within the active parish.
type SearchWardsRequest struct {
	Q      string            `json:"q,omitempty"`
	Filter *WardSearchFilter `json:"filter,omitempty"`
	Cursor string            `json:"cursor,omitempty"`
	Limit  int               `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateWardRequest creates a new Ward in the active parish.
type CreateWardRequest struct {
	ZoneID      uuid.UUID `json:"zoneId" validate:"required"`
	Code        string    `json:"code" validate:"required"`
	Name        string    `json:"name" validate:"required,min=1"`
	PatronSaint *string   `json:"patronSaint,omitempty"`
	HeadName    *string   `json:"headName,omitempty"`
	HeadPhone   *string   `json:"headPhone,omitempty"`
}

// UpdateWardRequest updates an existing Ward.
type UpdateWardRequest struct {
	ZoneID      *uuid.UUID `json:"zoneId,omitempty"`
	Code        *string    `json:"code,omitempty"`
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=1"`
	PatronSaint *string    `json:"patronSaint,omitempty"`
	HeadName    *string    `json:"headName,omitempty"`
	HeadPhone   *string    `json:"headPhone,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}

// LookupSearchFilter filters lookup search results.
type LookupSearchFilter struct {
	Category string `json:"category,omitempty"`
	Active   *bool  `json:"active,omitempty"`
	Scope    string `json:"scope,omitempty" validate:"omitempty,oneof=all global parish"`
}

// SearchLookupsRequest searches global and/or parish lookups.
type SearchLookupsRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *LookupSearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateLookupRequest creates a global or parish-scoped Lookup.
type CreateLookupRequest struct {
	Category  string `json:"category" validate:"required"`
	Code      string `json:"code" validate:"required"`
	Label     string `json:"label" validate:"required,min=1"`
	SortOrder int    `json:"sortOrder,omitempty"`
	Global    bool   `json:"global,omitempty"`
}

// UpdateLookupRequest updates an existing Lookup.
type UpdateLookupRequest struct {
	Category  *string `json:"category,omitempty"`
	Code      *string `json:"code,omitempty"`
	Label     *string `json:"label,omitempty" validate:"omitempty,min=1"`
	SortOrder *int    `json:"sortOrder,omitempty"`
	Active    *bool   `json:"active,omitempty"`
}

// PeriodSearchFilter filters period search results.
type PeriodSearchFilter struct {
	Type      string `json:"type,omitempty" validate:"omitempty,oneof=fiscal liturgical"`
	Status    string `json:"status,omitempty" validate:"omitempty,oneof=open closed"`
	IsCurrent *bool  `json:"isCurrent,omitempty"`
}

// SearchPeriodsRequest searches fiscal and liturgical periods in the active parish.
type SearchPeriodsRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *PeriodSearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreatePeriodRequest creates a fiscal or liturgical Period.
type CreatePeriodRequest struct {
	Type      string `json:"type,omitempty" validate:"omitempty,oneof=fiscal liturgical"`
	Code      string `json:"code" validate:"required"`
	Name      string `json:"name" validate:"required,min=1"`
	StartDate string `json:"startDate" validate:"required"`
	EndDate   string `json:"endDate" validate:"required"`
	IsCurrent bool   `json:"isCurrent,omitempty"`
}

// UpdatePeriodRequest updates an existing Period.
type UpdatePeriodRequest struct {
	Code      *string `json:"code,omitempty"`
	Name      *string `json:"name,omitempty" validate:"omitempty,min=1"`
	StartDate *string `json:"startDate,omitempty"`
	EndDate   *string `json:"endDate,omitempty"`
	Status    *string `json:"status,omitempty" validate:"omitempty,oneof=open closed"`
	IsCurrent *bool   `json:"isCurrent,omitempty"`
}
