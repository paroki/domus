package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
)

// Feature-owned Casbin role and resource permission constants.
var (
	RoleAdmin  = authz.Role("cantus", "admin")
	PermCantus = authz.Permission("cantus", "cantus")
)

// Cantus jenis and media type constants.
const (
	JenisLagu   = "lagu"
	JenisMazmur = "mazmur"
	JenisProfan = "profan"

	MediaTypeImage = "image"
	MediaTypePDF   = "pdf"
	MediaTypeAudio = "audio"
	MediaTypeVideo = "video"
)

// CantusMedia represents a single media attachment on a cantus entry.
type CantusMedia struct {
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	ViewURL     string `json:"viewUrl,omitempty"`
	DownloadURL string `json:"downloadUrl,omitempty"`
	S3Key       string `json:"s3Key,omitempty"`
}

// CantusResponse is the public API representation of a Cantus entry.
type CantusResponse struct {
	ID        uuid.UUID         `json:"id"`
	Slug      string            `json:"slug"`
	Jenis     string            `json:"jenis"`
	Sumber    *string           `json:"sumber,omitempty"`
	Nomor     *string           `json:"nomor,omitempty"`
	Judul     string            `json:"judul"`
	Kategori  *string           `json:"kategori,omitempty"`
	Tag       []string          `json:"tag"`
	NadaDasar *string           `json:"nadaDasar,omitempty"`
	Lirik     *string           `json:"lirik,omitempty"`
	Kredit    map[string]string `json:"kredit,omitempty"`
	Media     []CantusMedia     `json:"media,omitempty"`
	Meta      map[string]any    `json:"meta,omitempty"`
	Active    bool              `json:"active"`
	CanManage bool              `json:"canManage"`
	CreatedBy *uuid.UUID        `json:"createdBy,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// CantusSearchFilter filters cantus catalog searches.
type CantusSearchFilter struct {
	Slug      string `json:"slug,omitempty"`
	Jenis     string `json:"jenis,omitempty" validate:"omitempty,oneof=lagu mazmur profan"`
	Sumber    string `json:"sumber,omitempty"`
	Kategori  string `json:"kategori,omitempty"`
	Tag       string `json:"tag,omitempty"`
	NadaDasar string `json:"nadaDasar,omitempty"`
	Active    *bool  `json:"active,omitempty"`
}

// SearchCantusRequest searches the global cantus catalog with query, filters, and cursor pagination.
type SearchCantusRequest struct {
	Q      string              `json:"q,omitempty"`
	Filter *CantusSearchFilter `json:"filter,omitempty"`
	Cursor string              `json:"cursor,omitempty"`
	Limit  int                 `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateCantusRequest creates (or reactivates by slug) a global cantus entry.
type CreateCantusRequest struct {
	Slug      string            `json:"slug" validate:"required"`
	Jenis     string            `json:"jenis" validate:"required,oneof=lagu mazmur profan"`
	Sumber    *string           `json:"sumber,omitempty"`
	Nomor     *string           `json:"nomor,omitempty"`
	Judul     string            `json:"judul" validate:"required,min=1"`
	Kategori  *string           `json:"kategori,omitempty"`
	Tag       []string          `json:"tag,omitempty"`
	NadaDasar *string           `json:"nadaDasar,omitempty"`
	Lirik     *string           `json:"lirik,omitempty"`
	Kredit    map[string]string `json:"kredit,omitempty"`
	Media     []CantusMedia     `json:"media,omitempty"`
	Meta      map[string]any    `json:"meta,omitempty"`
	Active    *bool             `json:"active,omitempty"`
}

// UpdateCantusRequest partially updates a global cantus entry.
type UpdateCantusRequest struct {
	Slug      *string           `json:"slug,omitempty"`
	Jenis     *string           `json:"jenis,omitempty" validate:"omitempty,oneof=lagu mazmur profan"`
	Sumber    *string           `json:"sumber,omitempty"`
	Nomor     *string           `json:"nomor,omitempty"`
	Judul     *string           `json:"judul,omitempty"`
	Kategori  *string           `json:"kategori,omitempty"`
	Tag       []string          `json:"tag,omitempty"`
	NadaDasar *string           `json:"nadaDasar,omitempty"`
	Lirik     *string           `json:"lirik,omitempty"`
	Kredit    map[string]string `json:"kredit,omitempty"`
	Media     []CantusMedia     `json:"media,omitempty"`
	Meta      map[string]any    `json:"meta,omitempty"`
	Active    *bool             `json:"active,omitempty"`
}
