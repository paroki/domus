package domain

import (
	"time"

	"github.com/google/uuid"
)

// KeuskupanResponse is the public API representation of a Keuskupan.
type KeuskupanResponse struct {
	ID        uuid.UUID `json:"id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Uskup     *string   `json:"uskup,omitempty"`
	Alamat    *string   `json:"alamat,omitempty"`
	Telepon   *string   `json:"telepon,omitempty"`
	Email     *string   `json:"email,omitempty"`
	Website   *string   `json:"website,omitempty"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// KeuskupanSearchFilter filters keuskupan search results.
type KeuskupanSearchFilter struct {
	Active *bool `json:"active,omitempty"`
}

// SearchKeuskupanRequest searches keuskupan with optional query, filter, and cursor pagination.
type SearchKeuskupanRequest struct {
	Q      string                 `json:"q,omitempty"`
	Filter *KeuskupanSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateKeuskupanRequest creates a new Keuskupan.
type CreateKeuskupanRequest struct {
	Kode    string  `json:"kode" validate:"required"`
	Nama    string  `json:"nama" validate:"required,min=1"`
	Uskup   *string `json:"uskup,omitempty"`
	Alamat  *string `json:"alamat,omitempty"`
	Telepon *string `json:"telepon,omitempty"`
	Email   *string `json:"email,omitempty" validate:"omitempty,email"`
	Website *string `json:"website,omitempty"`
	Active  *bool   `json:"active,omitempty"`
}

// UpdateKeuskupanRequest updates an existing Keuskupan.
type UpdateKeuskupanRequest struct {
	Kode    *string `json:"kode,omitempty"`
	Nama    *string `json:"nama,omitempty" validate:"omitempty,min=1"`
	Uskup   *string `json:"uskup,omitempty"`
	Alamat  *string `json:"alamat,omitempty"`
	Telepon *string `json:"telepon,omitempty"`
	Email   *string `json:"email,omitempty" validate:"omitempty,email"`
	Website *string `json:"website,omitempty"`
	Active  *bool   `json:"active,omitempty"`
}

// ParokiResponse is the public API representation of a Paroki.
type ParokiResponse struct {
	ID          uuid.UUID `json:"id"`
	KeuskupanID uuid.UUID `json:"keuskupanId"`
	Kode        string    `json:"kode"`
	Nama        string    `json:"nama"`
	Pelindung   *string   `json:"pelindung,omitempty"`
	Pastor      *string   `json:"pastor,omitempty"`
	Alamat      *string   `json:"alamat,omitempty"`
	Telepon     *string   `json:"telepon,omitempty"`
	Email       *string   `json:"email,omitempty"`
	Website     *string   `json:"website,omitempty"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	KopSurat    *string   `json:"kopSurat,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ParokiSearchFilter filters paroki search results.
type ParokiSearchFilter struct {
	KeuskupanID *uuid.UUID `json:"keuskupanId,omitempty"`
	Active      *bool      `json:"active,omitempty"`
	MemberOnly  *bool      `json:"memberOnly,omitempty"`
}

// SearchParokiRequest searches paroki with optional query, filter, and cursor pagination.
type SearchParokiRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *ParokiSearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateParokiRequest creates a new Paroki.
type CreateParokiRequest struct {
	KeuskupanID uuid.UUID `json:"keuskupanId" validate:"required"`
	Kode        string    `json:"kode" validate:"required"`
	Nama        string    `json:"nama" validate:"required,min=1"`
	Pelindung   *string   `json:"pelindung,omitempty"`
	Pastor      *string   `json:"pastor,omitempty"`
	Alamat      *string   `json:"alamat,omitempty"`
	Telepon     *string   `json:"telepon,omitempty"`
	Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
	Website     *string   `json:"website,omitempty"`
	LogoURL     *string   `json:"logoUrl,omitempty"`
	KopSurat    *string   `json:"kopSurat,omitempty"`
	Active      *bool     `json:"active,omitempty"`
}

// UpdateParokiRequest updates a Paroki (super-admin fields or paroki profile).
type UpdateParokiRequest struct {
	KeuskupanID *uuid.UUID `json:"keuskupanId,omitempty"`
	Kode        *string    `json:"kode,omitempty"`
	Nama        *string    `json:"nama,omitempty" validate:"omitempty,min=1"`
	Pelindung   *string    `json:"pelindung,omitempty"`
	Pastor      *string    `json:"pastor,omitempty"`
	Alamat      *string    `json:"alamat,omitempty"`
	Telepon     *string    `json:"telepon,omitempty"`
	Email       *string    `json:"email,omitempty" validate:"omitempty,email"`
	Website     *string    `json:"website,omitempty"`
	LogoURL     *string    `json:"logoUrl,omitempty"`
	KopSurat    *string    `json:"kopSurat,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}

// UpdateParokiProfileRequest updates mutable profile fields of the current active Paroki.
type UpdateParokiProfileRequest struct {
	Nama      *string `json:"nama,omitempty" validate:"omitempty,min=1"`
	Pelindung *string `json:"pelindung,omitempty"`
	Pastor    *string `json:"pastor,omitempty"`
	Alamat    *string `json:"alamat,omitempty"`
	Telepon   *string `json:"telepon,omitempty"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	Website   *string `json:"website,omitempty"`
	LogoURL   *string `json:"logoUrl,omitempty"`
	KopSurat  *string `json:"kopSurat,omitempty"`
}

// ParokiMemberResponse is the public API representation of a ParokiMember joined with user info.
type ParokiMemberResponse struct {
	ID        uuid.UUID `json:"id"`
	ParokiID  uuid.UUID `json:"parokiId"`
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ParokiMemberSearchFilter filters paroki member search results.
type ParokiMemberSearchFilter struct {
	Role string `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
}

// SearchParokiMembersRequest searches members of a Paroki.
type SearchParokiMembersRequest struct {
	Q      string                    `json:"q,omitempty"`
	Filter *ParokiMemberSearchFilter `json:"filter,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
	Limit  int                       `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// AddParokiMemberRequest adds an existing or invited user as a member of a Paroki.
type AddParokiMemberRequest struct {
	UserID *uuid.UUID `json:"userId,omitempty"`
	Email  string     `json:"email,omitempty" validate:"omitempty,email"`
	Name   string     `json:"name,omitempty"`
	Phone  *string    `json:"phone,omitempty"`
	Role   string     `json:"role,omitempty" validate:"omitempty,oneof=admin member"`
}

// UpdateParokiMemberRequest updates a paroki member's role.
type UpdateParokiMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

// WilayahResponse is the public API representation of a Wilayah.
type WilayahResponse struct {
	ID                 uuid.UUID `json:"id"`
	ParokiID           uuid.UUID `json:"parokiId"`
	Jenis              string    `json:"jenis"`
	Kode               string    `json:"kode"`
	Nama               string    `json:"nama"`
	Koordinator        *string   `json:"koordinator,omitempty"`
	TeleponKoordinator *string   `json:"teleponKoordinator,omitempty"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// LingkunganResponse is the public API representation of a Lingkungan.
type LingkunganResponse struct {
	ID           uuid.UUID `json:"id"`
	ParokiID     uuid.UUID `json:"parokiId"`
	WilayahID    uuid.UUID `json:"wilayahId"`
	Kode         string    `json:"kode"`
	Nama         string    `json:"nama"`
	Pelindung    *string   `json:"pelindung,omitempty"`
	Ketua        *string   `json:"ketua,omitempty"`
	TeleponKetua *string   `json:"teleponKetua,omitempty"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// LookupResponse is the public API representation of a Lookup.
type LookupResponse struct {
	ID        uuid.UUID  `json:"id"`
	ParokiID  *uuid.UUID `json:"parokiId,omitempty"`
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
	ParokiID  uuid.UUID `json:"parokiId"`
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

// WilayahSearchFilter filters wilayah search results.
type WilayahSearchFilter struct {
	Jenis  string `json:"jenis,omitempty" validate:"omitempty,oneof=wilayah stasi"`
	Active *bool  `json:"active,omitempty"`
}

// SearchWilayahRequest searches wilayah within the active paroki.
type SearchWilayahRequest struct {
	Q      string               `json:"q,omitempty"`
	Filter *WilayahSearchFilter `json:"filter,omitempty"`
	Cursor string               `json:"cursor,omitempty"`
	Limit  int                  `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateWilayahRequest creates a new Wilayah or Stasi in the active paroki.
type CreateWilayahRequest struct {
	Jenis              string  `json:"jenis,omitempty" validate:"omitempty,oneof=wilayah stasi"`
	Kode               string  `json:"kode" validate:"required"`
	Nama               string  `json:"nama" validate:"required,min=1"`
	Koordinator        *string `json:"koordinator,omitempty"`
	TeleponKoordinator *string `json:"teleponKoordinator,omitempty"`
}

// UpdateWilayahRequest updates an existing Wilayah or Stasi.
type UpdateWilayahRequest struct {
	Jenis              *string `json:"jenis,omitempty" validate:"omitempty,oneof=wilayah stasi"`
	Kode               *string `json:"kode,omitempty"`
	Nama               *string `json:"nama,omitempty" validate:"omitempty,min=1"`
	Koordinator        *string `json:"koordinator,omitempty"`
	TeleponKoordinator *string `json:"teleponKoordinator,omitempty"`
	Active             *bool   `json:"active,omitempty"`
}

// LingkunganSearchFilter filters lingkungan search results.
type LingkunganSearchFilter struct {
	WilayahID *uuid.UUID `json:"wilayahId,omitempty"`
	Active    *bool      `json:"active,omitempty"`
}

// SearchLingkunganRequest searches lingkungan within the active paroki.
type SearchLingkunganRequest struct {
	Q      string                  `json:"q,omitempty"`
	Filter *LingkunganSearchFilter `json:"filter,omitempty"`
	Cursor string                  `json:"cursor,omitempty"`
	Limit  int                     `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateLingkunganRequest creates a new Lingkungan in the active paroki.
type CreateLingkunganRequest struct {
	WilayahID    uuid.UUID `json:"wilayahId" validate:"required"`
	Kode         string    `json:"kode" validate:"required"`
	Nama         string    `json:"nama" validate:"required,min=1"`
	Pelindung    *string   `json:"pelindung,omitempty"`
	Ketua        *string   `json:"ketua,omitempty"`
	TeleponKetua *string   `json:"teleponKetua,omitempty"`
}

// UpdateLingkunganRequest updates an existing Lingkungan.
type UpdateLingkunganRequest struct {
	WilayahID    *uuid.UUID `json:"wilayahId,omitempty"`
	Kode         *string    `json:"kode,omitempty"`
	Nama         *string    `json:"nama,omitempty" validate:"omitempty,min=1"`
	Pelindung    *string    `json:"pelindung,omitempty"`
	Ketua        *string    `json:"ketua,omitempty"`
	TeleponKetua *string    `json:"teleponKetua,omitempty"`
	Active       *bool      `json:"active,omitempty"`
}

// LookupSearchFilter filters lookup search results.
type LookupSearchFilter struct {
	Category string `json:"category,omitempty"`
	Active   *bool  `json:"active,omitempty"`
	Scope    string `json:"scope,omitempty" validate:"omitempty,oneof=all global paroki"`
}

// SearchLookupsRequest searches global and/or paroki lookups.
type SearchLookupsRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *LookupSearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateLookupRequest creates a global or paroki-scoped Lookup.
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

// SearchPeriodsRequest searches fiscal and liturgical periods in the active paroki.
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

// KevikepanResponse is the public API representation of a Kevikepan.
type KevikepanResponse struct {
	ID          uuid.UUID `json:"id"`
	KeuskupanID uuid.UUID `json:"keuskupanId"`
	Kode        string    `json:"kode"`
	Nama        string    `json:"nama"`
	Vikep       *string   `json:"vikep,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// KevikepanSearchFilter filters kevikepan search results.
type KevikepanSearchFilter struct {
	KeuskupanID *uuid.UUID `json:"keuskupanId,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}

// SearchKevikepanRequest searches kevikepan with optional query, filter, and cursor pagination.
type SearchKevikepanRequest struct {
	Q      string                 `json:"q,omitempty"`
	Filter *KevikepanSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateKevikepanRequest creates a new Kevikepan.
type CreateKevikepanRequest struct {
	KeuskupanID uuid.UUID `json:"keuskupanId" validate:"required"`
	Kode        string    `json:"kode" validate:"required"`
	Nama        string    `json:"nama" validate:"required,min=1"`
	Vikep       *string   `json:"vikep,omitempty"`
	Active      *bool     `json:"active,omitempty"`
}

// UpdateKevikepanRequest updates an existing Kevikepan.
type UpdateKevikepanRequest struct {
	KeuskupanID *uuid.UUID `json:"keuskupanId,omitempty"`
	Kode        *string    `json:"kode,omitempty"`
	Nama        *string    `json:"nama,omitempty" validate:"omitempty,min=1"`
	Vikep       *string    `json:"vikep,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}

// KotaResponse is the public API representation of a Kota.
type KotaResponse struct {
	ID        uuid.UUID `json:"id"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// KotaSearchFilter filters kota search results.
type KotaSearchFilter struct {
	Active *bool `json:"active,omitempty"`
}

// SearchKotaRequest searches kota with optional query, filter, and cursor pagination.
type SearchKotaRequest struct {
	Q      string            `json:"q,omitempty"`
	Filter *KotaSearchFilter `json:"filter,omitempty"`
	Cursor string            `json:"cursor,omitempty"`
	Limit  int               `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateKotaRequest creates a new Kota.
type CreateKotaRequest struct {
	Kode   string `json:"kode" validate:"required"`
	Nama   string `json:"nama" validate:"required,min=1"`
	Active *bool  `json:"active,omitempty"`
}

// UpdateKotaRequest updates an existing Kota.
type UpdateKotaRequest struct {
	Kode   *string `json:"kode,omitempty"`
	Nama   *string `json:"nama,omitempty" validate:"omitempty,min=1"`
	Active *bool   `json:"active,omitempty"`
}

// KecamatanResponse is the public API representation of a Kecamatan.
type KecamatanResponse struct {
	ID        uuid.UUID `json:"id"`
	KotaID    uuid.UUID `json:"kotaId"`
	Kode      string    `json:"kode"`
	Nama      string    `json:"nama"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// KecamatanSearchFilter filters kecamatan search results.
type KecamatanSearchFilter struct {
	KotaID *uuid.UUID `json:"kotaId,omitempty"`
	Active *bool      `json:"active,omitempty"`
}

// SearchKecamatanRequest searches kecamatan with optional query, filter, and cursor pagination.
type SearchKecamatanRequest struct {
	Q      string                 `json:"q,omitempty"`
	Filter *KecamatanSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateKecamatanRequest creates a new Kecamatan.
type CreateKecamatanRequest struct {
	KotaID uuid.UUID `json:"kotaId" validate:"required"`
	Kode   string    `json:"kode" validate:"required"`
	Nama   string    `json:"nama" validate:"required,min=1"`
	Active *bool     `json:"active,omitempty"`
}

// UpdateKecamatanRequest updates an existing Kecamatan.
type UpdateKecamatanRequest struct {
	KotaID *uuid.UUID `json:"kotaId,omitempty"`
	Kode   *string    `json:"kode,omitempty"`
	Nama   *string    `json:"nama,omitempty" validate:"omitempty,min=1"`
	Active *bool      `json:"active,omitempty"`
}

// KelurahanResponse is the public API representation of a Kelurahan.
type KelurahanResponse struct {
	ID          uuid.UUID `json:"id"`
	KecamatanID uuid.UUID `json:"kecamatanId"`
	Kode        string    `json:"kode"`
	Nama        string    `json:"nama"`
	KodePos     *string   `json:"kodePos,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// KelurahanSearchFilter filters kelurahan search results.
type KelurahanSearchFilter struct {
	KecamatanID *uuid.UUID `json:"kecamatanId,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}

// SearchKelurahanRequest searches kelurahan with optional query, filter, and cursor pagination.
type SearchKelurahanRequest struct {
	Q      string                 `json:"q,omitempty"`
	Filter *KelurahanSearchFilter `json:"filter,omitempty"`
	Cursor string                 `json:"cursor,omitempty"`
	Limit  int                    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateKelurahanRequest creates a new Kelurahan.
type CreateKelurahanRequest struct {
	KecamatanID uuid.UUID `json:"kecamatanId" validate:"required"`
	Kode        string    `json:"kode" validate:"required"`
	Nama        string    `json:"nama" validate:"required,min=1"`
	KodePos     *string   `json:"kodePos,omitempty"`
	Active      *bool     `json:"active,omitempty"`
}

// UpdateKelurahanRequest updates an existing Kelurahan.
type UpdateKelurahanRequest struct {
	KecamatanID *uuid.UUID `json:"kecamatanId,omitempty"`
	Kode        *string    `json:"kode,omitempty"`
	Nama        *string    `json:"nama,omitempty" validate:"omitempty,min=1"`
	KodePos     *string    `json:"kodePos,omitempty"`
	Active      *bool      `json:"active,omitempty"`
}
