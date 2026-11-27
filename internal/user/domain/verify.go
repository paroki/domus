package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// IdentityImageInput represents a validated image payload for AI vision extraction.
type IdentityImageInput struct {
	MimeType string
	Data     []byte
}

// IdentityExtractInput contains the uploaded identity document images for extraction.
type IdentityExtractInput struct {
	KTP          *IdentityImageInput
	KartuPelajar *IdentityImageInput
	KK           *IdentityImageInput
}

// ExtractedKTP holds structured fields extracted from a KTP / KIA / Kartu Pelajar image.
type ExtractedKTP struct {
	NIK              string  `json:"nik"`
	Nama             string  `json:"nama"`
	TempatLahir      *string `json:"tempatLahir,omitempty"`
	TanggalLahir     *string `json:"tanggalLahir,omitempty"`
	JenisKelamin     *string `json:"jenisKelamin,omitempty"`
	Alamat           *string `json:"alamat,omitempty"`
	Agama            *string `json:"agama,omitempty"`
	GolonganDarah    *string `json:"golonganDarah,omitempty"`
	StatusPernikahan *string `json:"statusPernikahan,omitempty"`
	Pekerjaan        *string `json:"pekerjaan,omitempty"`
}

// ExtractedKKMember holds a single family member row extracted from a Kartu Keluarga image.
type ExtractedKKMember struct {
	Nama             string  `json:"nama"`
	NIK              string  `json:"nik"`
	JenisKelamin     string  `json:"jenisKelamin"`
	TempatLahir      *string `json:"tempatLahir,omitempty"`
	TanggalLahir     *string `json:"tanggalLahir,omitempty"`
	Agama            *string `json:"agama,omitempty"`
	Pendidikan       *string `json:"pendidikan,omitempty"`
	Pekerjaan        *string `json:"pekerjaan,omitempty"`
	StatusPernikahan *string `json:"statusPernikahan,omitempty"`
	HubunganKeluarga *string `json:"hubunganKeluarga,omitempty"`
}

// ExtractedKK holds structured fields extracted from a Kartu Keluarga image.
type ExtractedKK struct {
	NomorKK      string              `json:"nomorKk"`
	NamaKeluarga string              `json:"namaKeluarga"`
	Alamat       *string             `json:"alamat,omitempty"`
	Anggota      []ExtractedKKMember `json:"anggota"`
}

// IdentityExtraction is the structured JSON output from AI vision extraction.
type IdentityExtraction struct {
	KTP *ExtractedKTP `json:"ktp,omitempty"`
	KK  *ExtractedKK  `json:"kk,omitempty"`
}

// VerifyPreviewResult is returned by CensusVerifier.PreviewVerification.
type VerifyPreviewResult struct {
	MatchStatus   string     `json:"matchStatus"`
	MatchedUmatID *uuid.UUID `json:"matchedUmatId,omitempty"`
}

// VerifyConfirmDocument is a single document image to encrypt and store during confirmation.
type VerifyConfirmDocument struct {
	Jenis    string
	MimeType string
	Data     []byte
}

// VerifyConfirmInput is passed to CensusVerifier.VerifyAndSyncCensus.
type VerifyConfirmInput struct {
	ParokiID     uuid.UUID
	UserID       uuid.UUID
	Payload      IdentityExtraction
	MemberIndex  int
	WilayahID    *uuid.UUID
	LingkunganID *uuid.UUID
	Documents    []VerifyConfirmDocument
}

// VerifyConfirmResult is returned by CensusVerifier.VerifyAndSyncCensus and POST /api/me/verify/confirm.
type VerifyConfirmResult struct {
	UmatID       uuid.UUID  `json:"umatId"`
	NomorInduk   string     `json:"nomorInduk"`
	KeluargaID   *uuid.UUID `json:"keluargaId,omitempty"`
	Nama         string     `json:"nama"`
	JenisKelamin string     `json:"jenisKelamin"`
	TanggalLahir *string    `json:"tanggalLahir,omitempty"`
	Verified     bool       `json:"verified"`
	VerifiedAt   *time.Time `json:"verifiedAt,omitempty"`
}

// SelfDokumenMeta is document metadata (no bytes) returned in GET /api/me/umat.
type SelfDokumenMeta struct {
	ID         uuid.UUID  `json:"id"`
	UmatID     *uuid.UUID `json:"umatId,omitempty"`
	KeluargaID *uuid.UUID `json:"keluargaId,omitempty"`
	Jenis      string     `json:"jenis"`
	MimeType   string     `json:"mimeType"`
	SizeBytes  int        `json:"sizeBytes"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// SelfUmatInfo is the caller's or family member's umat info in GET /api/me/umat.
type SelfUmatInfo struct {
	ID               uuid.UUID  `json:"id"`
	NomorInduk       string     `json:"nomorInduk"`
	Nama             string     `json:"nama"`
	NIK              *string    `json:"nik,omitempty"`
	JenisKelamin     string     `json:"jenisKelamin"`
	TempatLahir      *string    `json:"tempatLahir,omitempty"`
	TanggalLahir     *string    `json:"tanggalLahir,omitempty"`
	Agama            *string    `json:"agama,omitempty"`
	GolonganDarah    *string    `json:"golonganDarah,omitempty"`
	HubunganKeluarga *string    `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string    `json:"statusPernikahan,omitempty"`
	Pekerjaan        *string    `json:"pekerjaan,omitempty"`
	Alamat           *string    `json:"alamat,omitempty"`
	WilayahID        *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID     *uuid.UUID `json:"lingkunganId,omitempty"`
	UserID           *uuid.UUID `json:"userId,omitempty"`
}

// SelfKeluargaInfo is the household info with decrypted nomorKk in GET /api/me/umat.
type SelfKeluargaInfo struct {
	ID              uuid.UUID  `json:"id"`
	NomorRegistrasi string     `json:"nomorRegistrasi"`
	NomorKK         *string    `json:"nomorKk,omitempty"`
	NamaKeluarga    string     `json:"namaKeluarga"`
	Alamat          *string    `json:"alamat,omitempty"`
	WilayahID       *uuid.UUID `json:"wilayahId,omitempty"`
	LingkunganID    *uuid.UUID `json:"lingkunganId,omitempty"`
}

// SelfUmatProfile is returned by CensusVerifier.GetSelfProfile and GET /api/me/umat.
type SelfUmatProfile struct {
	Umat     *SelfUmatInfo     `json:"umat"`
	Keluarga *SelfKeluargaInfo `json:"keluarga,omitempty"`
	Members  []SelfUmatInfo    `json:"members"`
	Dokumen  []SelfDokumenMeta `json:"dokumen"`
}

// VerifyExtractResponse is returned by POST /api/me/verify/extract.
type VerifyExtractResponse struct {
	Extraction  IdentityExtraction `json:"extraction"`
	MatchStatus string             `json:"matchStatus"`
}

// IdentityExtractor extracts structured identity fields from document images via AI Vision.
type IdentityExtractor interface {
	ExtractIdentityDocument(ctx context.Context, in IdentityExtractInput) (IdentityExtraction, error)
}

// CensusVerifier previews, confirms, and reads parishioner self-verification profiles.
type CensusVerifier interface {
	PreviewVerification(ctx context.Context, parokiID, userID uuid.UUID, extraction IdentityExtraction) (VerifyPreviewResult, error)
	VerifyAndSyncCensus(ctx context.Context, in VerifyConfirmInput) (VerifyConfirmResult, error)
	GetSelfProfile(ctx context.Context, parokiID, userID uuid.UUID) (*SelfUmatProfile, error)
}
