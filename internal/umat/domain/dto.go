package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
)

// Feature-owned Casbin roles and resource permissions.
var (
	RoleAdmin    = authz.Role("umat", "admin")
	PermKeluarga = authz.Permission("umat", "keluarga")
	PermUmat     = authz.Permission("umat", "umat")
)

// MaskPII deterministically masks a sensitive identifier (e.g. NIK or Nomor KK) for list/summary responses.
func MaskPII(plain string) string {
	if plain == "" {
		return ""
	}
	r := []rune(plain)
	if len(r) > 4 {
		return strings.Repeat("*", len(r)-4) + string(r[len(r)-4:])
	}
	if len(r) >= 2 {
		return "****" + string(r[len(r)-2:])
	}
	return "****" + plain
}

// KeluargaListItemResponse is the masked list representation of a Keluarga.
type KeluargaListItemResponse struct {
	ID            uuid.UUID  `json:"id"`
	ParokiID      uuid.UUID  `json:"parokiId"`
	WilayahID     *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID  *uuid.UUID `json:"lingkunganId,omitempty"`
	NomorKKMasked *string    `json:"nomorKkMasked,omitempty"`
	NamaKeluarga  string     `json:"namaKeluarga"`
	Alamat        *string    `json:"alamat,omitempty"`
	KotaID        *uuid.UUID `json:"kotaId,omitempty"`
	KecamatanID   *uuid.UUID `json:"kecamatanId,omitempty"`
	KelurahanID   *uuid.UUID `json:"kelurahanId,omitempty"`
	KodePos       *string    `json:"kodePos,omitempty"`
	StatusRumah   *string    `json:"statusRumah,omitempty"`
	StatusEkonomi *string    `json:"statusEkonomi,omitempty"`
	Aktif         bool       `json:"aktif"`
	Catatan       *string    `json:"catatan,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// KeluargaAnggotaSummary is the masked summary of a member inside KeluargaDetailResponse.
type KeluargaAnggotaSummary struct {
	ID               uuid.UUID `json:"id"`
	Nama             string    `json:"nama"`
	NamaBaptis       *string   `json:"namaBaptis,omitempty"`
	NIKMasked        *string   `json:"nikMasked,omitempty"`
	JenisKelamin     string    `json:"jenisKelamin"`
	HubunganKeluarga *string   `json:"hubunganKeluarga,omitempty"`
	Almarhum         bool      `json:"almarhum"`
}

// KeluargaDetailResponse is the single-entity representation of a Keluarga including decrypted NomorKK and Anggota.
type KeluargaDetailResponse struct {
	ID            uuid.UUID                `json:"id"`
	ParokiID      uuid.UUID                `json:"parokiId"`
	WilayahID     *uuid.UUID               `json:"wilayahId,omitempty"`
	LingkunganID  *uuid.UUID               `json:"lingkunganId,omitempty"`
	NomorKK       *string                  `json:"nomorKk,omitempty"`
	NomorKKMasked *string                  `json:"nomorKkMasked,omitempty"`
	NamaKeluarga  string                   `json:"namaKeluarga"`
	Alamat        *string                  `json:"alamat,omitempty"`
	KotaID        *uuid.UUID               `json:"kotaId,omitempty"`
	KecamatanID   *uuid.UUID               `json:"kecamatanId,omitempty"`
	KelurahanID   *uuid.UUID               `json:"kelurahanId,omitempty"`
	KodePos       *string                  `json:"kodePos,omitempty"`
	StatusRumah   *string                  `json:"statusRumah,omitempty"`
	StatusEkonomi *string                  `json:"statusEkonomi,omitempty"`
	Aktif         bool                     `json:"aktif"`
	Catatan       *string                  `json:"catatan,omitempty"`
	Anggota       []KeluargaAnggotaSummary `json:"anggota"`
	CreatedAt     time.Time                `json:"createdAt"`
	UpdatedAt     time.Time                `json:"updatedAt"`
}

// KeluargaSearchFilter filters Keluarga search results.
type KeluargaSearchFilter struct {
	WilayahID    *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID *uuid.UUID `json:"lingkunganId,omitempty"`
	Aktif        *bool      `json:"aktif,omitempty"`
	NomorKK      string     `json:"nomorKk,omitempty"`
}

// SearchKeluargaRequest searches Keluarga records in the active paroki.
type SearchKeluargaRequest struct {
	Q      string                `json:"q,omitempty"`
	Sort   string                `json:"sort,omitempty" validate:"omitempty,oneof=-createdAt namaKeluarga -namaKeluarga"`
	Filter *KeluargaSearchFilter `json:"filter,omitempty"`
	Cursor string                `json:"cursor,omitempty"`
	Limit  int                   `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateKeluargaRequest creates a new Keluarga record.
type CreateKeluargaRequest struct {
	WilayahID     *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID  *uuid.UUID `json:"lingkunganId,omitempty"`
	NomorKK       *string    `json:"nomorKk,omitempty"`
	NamaKeluarga  string     `json:"namaKeluarga" validate:"required"`
	Alamat        *string    `json:"alamat,omitempty"`
	KotaID        *uuid.UUID `json:"kotaId,omitempty"`
	KecamatanID   *uuid.UUID `json:"kecamatanId,omitempty"`
	KelurahanID   *uuid.UUID `json:"kelurahanId,omitempty"`
	KodePos       *string    `json:"kodePos,omitempty"`
	StatusRumah   *string    `json:"statusRumah,omitempty"`
	StatusEkonomi *string    `json:"statusEkonomi,omitempty"`
	Aktif         *bool      `json:"aktif,omitempty"`
	Catatan       *string    `json:"catatan,omitempty"`
}

// UpdateKeluargaRequest partially updates a Keluarga record.
type UpdateKeluargaRequest struct {
	WilayahID     *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID  *uuid.UUID `json:"lingkunganId,omitempty"`
	NomorKK       *string    `json:"nomorKk,omitempty"`
	NamaKeluarga  *string    `json:"namaKeluarga,omitempty"`
	Alamat        *string    `json:"alamat,omitempty"`
	KotaID        *uuid.UUID `json:"kotaId,omitempty"`
	KecamatanID   *uuid.UUID `json:"kecamatanId,omitempty"`
	KelurahanID   *uuid.UUID `json:"kelurahanId,omitempty"`
	KodePos       *string    `json:"kodePos,omitempty"`
	StatusRumah   *string    `json:"statusRumah,omitempty"`
	StatusEkonomi *string    `json:"statusEkonomi,omitempty"`
	Aktif         *bool      `json:"aktif,omitempty"`
	Catatan       *string    `json:"catatan,omitempty"`
}

// UmatKeluargaSummary is the masked household summary embedded in UmatDetailResponse.
type UmatKeluargaSummary struct {
	ID            uuid.UUID `json:"id"`
	NamaKeluarga  string    `json:"namaKeluarga"`
	NomorKKMasked *string   `json:"nomorKkMasked,omitempty"`
}

// UmatSakramenResponse represents a sacramental history record for an Umat.
type UmatSakramenResponse struct {
	ID               uuid.UUID  `json:"id"`
	ParokiID         uuid.UUID  `json:"parokiId"`
	UmatID           uuid.UUID  `json:"umatId"`
	Jenis            string     `json:"jenis"`
	Tanggal          *string    `json:"tanggal,omitempty"`
	KeuskupanID      *uuid.UUID `json:"keuskupanId,omitempty"`
	ParokiSakramenID *uuid.UUID `json:"parokiSakramenId,omitempty"`
	KeuskupanNama    *string    `json:"keuskupanNama,omitempty"`
	ParokiNama       *string    `json:"parokiNama,omitempty"`
	Buku             *string    `json:"buku,omitempty"`
	Nomor            *string    `json:"nomor,omitempty"`
	Catatan          *string    `json:"catatan,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// UmatListItemResponse is the masked list representation of an Umat.
type UmatListItemResponse struct {
	ID               uuid.UUID  `json:"id"`
	ParokiID         uuid.UUID  `json:"parokiId"`
	KeluargaID       *uuid.UUID `json:"keluargaId,omitempty"`
	WilayahID        *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID     *uuid.UUID `json:"lingkunganId,omitempty"`
	UserID           *uuid.UUID `json:"userId,omitempty"`
	Nama             string     `json:"nama"`
	NamaBaptis       *string    `json:"namaBaptis,omitempty"`
	NIKMasked        *string    `json:"nikMasked,omitempty"`
	JenisKelamin     string     `json:"jenisKelamin"`
	TempatLahir      *string    `json:"tempatLahir,omitempty"`
	TanggalLahir     *string    `json:"tanggalLahir,omitempty"`
	GolonganDarah    *string    `json:"golonganDarah,omitempty"`
	Suku             *string    `json:"suku,omitempty"`
	Agama            *string    `json:"agama,omitempty"`
	Kewarganegaraan  *string    `json:"kewarganegaraan,omitempty"`
	Negara           *string    `json:"negara,omitempty"`
	HubunganKeluarga *string    `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string    `json:"statusPernikahan,omitempty"`
	Almarhum         bool       `json:"almarhum"`
	TanggalMeninggal *string    `json:"tanggalMeninggal,omitempty"`
	Alamat           *string    `json:"alamat,omitempty"`
	KotaID           *uuid.UUID `json:"kotaId,omitempty"`
	KecamatanID      *uuid.UUID `json:"kecamatanId,omitempty"`
	KelurahanID      *uuid.UUID `json:"kelurahanId,omitempty"`
	Telepon          *string    `json:"telepon,omitempty"`
	Email            *string    `json:"email,omitempty"`
	FotoURL          *string    `json:"fotoUrl,omitempty"`
	Pendidikan       *string    `json:"pendidikan,omitempty"`
	Jurusan          *string    `json:"jurusan,omitempty"`
	Pekerjaan        *string    `json:"pekerjaan,omitempty"`
	Profesi          *string    `json:"profesi,omitempty"`
	Keterampilan     *string    `json:"keterampilan,omitempty"`
	Disabilitas      *string    `json:"disabilitas,omitempty"`
	AktivitasSosial  *string    `json:"aktivitasSosial,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// UmatDetailResponse is the single-entity representation of an Umat including decrypted NIK, Keluarga summary, and Sakramen.
type UmatDetailResponse struct {
	ID               uuid.UUID              `json:"id"`
	ParokiID         uuid.UUID              `json:"parokiId"`
	KeluargaID       *uuid.UUID             `json:"keluargaId,omitempty"`
	WilayahID        *uuid.UUID             `json:"wilayahId,omitempty"`
	LingkunganID     *uuid.UUID             `json:"lingkunganId,omitempty"`
	UserID           *uuid.UUID             `json:"userId,omitempty"`
	Nama             string                 `json:"nama"`
	NamaBaptis       *string                `json:"namaBaptis,omitempty"`
	NIK              *string                `json:"nik,omitempty"`
	NIKMasked        *string                `json:"nikMasked,omitempty"`
	JenisKelamin     string                 `json:"jenisKelamin"`
	TempatLahir      *string                `json:"tempatLahir,omitempty"`
	TanggalLahir     *string                `json:"tanggalLahir,omitempty"`
	GolonganDarah    *string                `json:"golonganDarah,omitempty"`
	Suku             *string                `json:"suku,omitempty"`
	Agama            *string                `json:"agama,omitempty"`
	Kewarganegaraan  *string                `json:"kewarganegaraan,omitempty"`
	Negara           *string                `json:"negara,omitempty"`
	HubunganKeluarga *string                `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string                `json:"statusPernikahan,omitempty"`
	Almarhum         bool                   `json:"almarhum"`
	TanggalMeninggal *string                `json:"tanggalMeninggal,omitempty"`
	Alamat           *string                `json:"alamat,omitempty"`
	KotaID           *uuid.UUID             `json:"kotaId,omitempty"`
	KecamatanID      *uuid.UUID             `json:"kecamatanId,omitempty"`
	KelurahanID      *uuid.UUID             `json:"kelurahanId,omitempty"`
	Telepon          *string                `json:"telepon,omitempty"`
	Email            *string                `json:"email,omitempty"`
	FotoURL          *string                `json:"fotoUrl,omitempty"`
	Pendidikan       *string                `json:"pendidikan,omitempty"`
	Jurusan          *string                `json:"jurusan,omitempty"`
	Pekerjaan        *string                `json:"pekerjaan,omitempty"`
	Profesi          *string                `json:"profesi,omitempty"`
	Keterampilan     *string                `json:"keterampilan,omitempty"`
	Disabilitas      *string                `json:"disabilitas,omitempty"`
	AktivitasSosial  *string                `json:"aktivitasSosial,omitempty"`
	Keluarga         *UmatKeluargaSummary   `json:"keluarga,omitempty"`
	Sakramen         []UmatSakramenResponse `json:"sakramen"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
}

// UmatSearchFilter filters Umat search results.
type UmatSearchFilter struct {
	KeluargaID       *uuid.UUID `json:"keluargaId,omitempty"`
	WilayahID        *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID     *uuid.UUID `json:"lingkunganId,omitempty"`
	JenisKelamin     string     `json:"jenisKelamin,omitempty" validate:"omitempty,oneof=pria wanita"`
	Agama            string     `json:"agama,omitempty"`
	StatusPernikahan string     `json:"statusPernikahan,omitempty"`
	Almarhum         *bool      `json:"almarhum,omitempty"`
	TanggalLahir     string     `json:"tanggalLahir,omitempty"`
	NIK              string     `json:"nik,omitempty"`
}

// SearchUmatRequest searches Umat records in the active paroki.
type SearchUmatRequest struct {
	Q      string            `json:"q,omitempty"`
	Sort   string            `json:"sort,omitempty" validate:"omitempty,oneof=-createdAt nama -nama"`
	Filter *UmatSearchFilter `json:"filter,omitempty"`
	Cursor string            `json:"cursor,omitempty"`
	Limit  int               `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// CreateUmatRequest creates a new Umat record.
type CreateUmatRequest struct {
	KeluargaID       *uuid.UUID `json:"keluargaId,omitempty"`
	WilayahID        *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID     *uuid.UUID `json:"lingkunganId,omitempty"`
	UserID           *uuid.UUID `json:"userId,omitempty"`
	Nama             string     `json:"nama" validate:"required"`
	NamaBaptis       *string    `json:"namaBaptis,omitempty"`
	NIK              *string    `json:"nik,omitempty"`
	JenisKelamin     string     `json:"jenisKelamin" validate:"required,oneof=pria wanita"`
	TempatLahir      *string    `json:"tempatLahir,omitempty"`
	TanggalLahir     *string    `json:"tanggalLahir,omitempty"`
	GolonganDarah    *string    `json:"golonganDarah,omitempty"`
	Suku             *string    `json:"suku,omitempty"`
	Agama            *string    `json:"agama,omitempty"`
	Kewarganegaraan  *string    `json:"kewarganegaraan,omitempty"`
	Negara           *string    `json:"negara,omitempty"`
	HubunganKeluarga *string    `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string    `json:"statusPernikahan,omitempty"`
	Almarhum         *bool      `json:"almarhum,omitempty"`
	TanggalMeninggal *string    `json:"tanggalMeninggal,omitempty"`
	Alamat           *string    `json:"alamat,omitempty"`
	KotaID           *uuid.UUID `json:"kotaId,omitempty"`
	KecamatanID      *uuid.UUID `json:"kecamatanId,omitempty"`
	KelurahanID      *uuid.UUID `json:"kelurahanId,omitempty"`
	Telepon          *string    `json:"telepon,omitempty"`
	Email            *string    `json:"email,omitempty"`
	FotoURL          *string    `json:"fotoUrl,omitempty"`
	Pendidikan       *string    `json:"pendidikan,omitempty"`
	Jurusan          *string    `json:"jurusan,omitempty"`
	Pekerjaan        *string    `json:"pekerjaan,omitempty"`
	Profesi          *string    `json:"profesi,omitempty"`
	Keterampilan     *string    `json:"keterampilan,omitempty"`
	Disabilitas      *string    `json:"disabilitas,omitempty"`
	AktivitasSosial  *string    `json:"aktivitasSosial,omitempty"`
}

// UpdateUmatRequest partially updates an Umat record.
type UpdateUmatRequest struct {
	KeluargaID       *uuid.UUID `json:"keluargaId,omitempty"`
	WilayahID        *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID     *uuid.UUID `json:"lingkunganId,omitempty"`
	UserID           *uuid.UUID `json:"userId,omitempty"`
	Nama             *string    `json:"nama,omitempty"`
	NamaBaptis       *string    `json:"namaBaptis,omitempty"`
	NIK              *string    `json:"nik,omitempty"`
	JenisKelamin     *string    `json:"jenisKelamin,omitempty" validate:"omitempty,oneof=pria wanita"`
	TempatLahir      *string    `json:"tempatLahir,omitempty"`
	TanggalLahir     *string    `json:"tanggalLahir,omitempty"`
	GolonganDarah    *string    `json:"golonganDarah,omitempty"`
	Suku             *string    `json:"suku,omitempty"`
	Agama            *string    `json:"agama,omitempty"`
	Kewarganegaraan  *string    `json:"kewarganegaraan,omitempty"`
	Negara           *string    `json:"negara,omitempty"`
	HubunganKeluarga *string    `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string    `json:"statusPernikahan,omitempty"`
	Almarhum         *bool      `json:"almarhum,omitempty"`
	TanggalMeninggal *string    `json:"tanggalMeninggal,omitempty"`
	Alamat           *string    `json:"alamat,omitempty"`
	KotaID           *uuid.UUID `json:"kotaId,omitempty"`
	KecamatanID      *uuid.UUID `json:"kecamatanId,omitempty"`
	KelurahanID      *uuid.UUID `json:"kelurahanId,omitempty"`
	Telepon          *string    `json:"telepon,omitempty"`
	Email            *string    `json:"email,omitempty"`
	FotoURL          *string    `json:"fotoUrl,omitempty"`
	Pendidikan       *string    `json:"pendidikan,omitempty"`
	Jurusan          *string    `json:"jurusan,omitempty"`
	Pekerjaan        *string    `json:"pekerjaan,omitempty"`
	Profesi          *string    `json:"profesi,omitempty"`
	Keterampilan     *string    `json:"keterampilan,omitempty"`
	Disabilitas      *string    `json:"disabilitas,omitempty"`
	AktivitasSosial  *string    `json:"aktivitasSosial,omitempty"`
}

// UpsertUmatSakramenRequest upserts a sacramental record keyed by (paroki_id, umat_id, jenis).
type UpsertUmatSakramenRequest struct {
	Jenis            string     `json:"jenis" validate:"required,oneof=baptis komuni_pertama krisma perkawinan imamat perminyakan kematian"`
	Tanggal          *string    `json:"tanggal,omitempty"`
	KeuskupanID      *uuid.UUID `json:"keuskupanId,omitempty"`
	ParokiSakramenID *uuid.UUID `json:"parokiSakramenId,omitempty"`
	KeuskupanNama    *string    `json:"keuskupanNama,omitempty"`
	ParokiNama       *string    `json:"parokiNama,omitempty"`
	Buku             *string    `json:"buku,omitempty"`
	Nomor            *string    `json:"nomor,omitempty"`
	Catatan          *string    `json:"catatan,omitempty"`
}

// UpdateUmatSakramenRequest partially updates an existing sacramental record.
type UpdateUmatSakramenRequest struct {
	Jenis            *string    `json:"jenis,omitempty" validate:"omitempty,oneof=baptis komuni_pertama krisma perkawinan imamat perminyakan kematian"`
	Tanggal          *string    `json:"tanggal,omitempty"`
	KeuskupanID      *uuid.UUID `json:"keuskupanId,omitempty"`
	ParokiSakramenID *uuid.UUID `json:"parokiSakramenId,omitempty"`
	KeuskupanNama    *string    `json:"keuskupanNama,omitempty"`
	ParokiNama       *string    `json:"parokiNama,omitempty"`
	Buku             *string    `json:"buku,omitempty"`
	Nomor            *string    `json:"nomor,omitempty"`
	Catatan          *string    `json:"catatan,omitempty"`
}
