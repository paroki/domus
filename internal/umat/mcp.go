package umat

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/umat/domain"
	"github.com/paroki/domus/internal/user"
)

// UmatSearchInput is the input schema for umat_search (never accepts raw NIK).
type UmatSearchInput struct {
	Query            string `json:"query,omitempty"`
	KeluargaID       string `json:"keluargaId,omitempty"`
	WilayahID        string `json:"wilayahId,omitempty"`
	LingkunganID     string `json:"lingkunganId,omitempty"`
	JenisKelamin     string `json:"jenisKelamin,omitempty"`
	Agama            string `json:"agama,omitempty"`
	StatusPernikahan string `json:"statusPernikahan,omitempty"`
	Almarhum         *bool  `json:"almarhum,omitempty"`
	Limit            int    `json:"limit,omitempty"`
}

// MCPUmat is the PII-masked MCP representation of an Umat.
type MCPUmat struct {
	ID               string    `json:"id"`
	ParokiID         string    `json:"parokiId"`
	KeluargaID       *string   `json:"keluargaId,omitempty"`
	WilayahID        *string   `json:"wilayahId,omitempty"`
	LingkunganID     *string   `json:"lingkunganId,omitempty"`
	UserID           *string   `json:"userId,omitempty"`
	Nama             string    `json:"nama"`
	NamaBaptis       *string   `json:"namaBaptis,omitempty"`
	NIKMasked        *string   `json:"nikMasked,omitempty"`
	JenisKelamin     string    `json:"jenisKelamin"`
	TempatLahir      *string   `json:"tempatLahir,omitempty"`
	TanggalLahir     *string   `json:"tanggalLahir,omitempty"`
	GolonganDarah    *string   `json:"golonganDarah,omitempty"`
	Suku             *string   `json:"suku,omitempty"`
	Agama            *string   `json:"agama,omitempty"`
	Kewarganegaraan  *string   `json:"kewarganegaraan,omitempty"`
	Negara           *string   `json:"negara,omitempty"`
	HubunganKeluarga *string   `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string   `json:"statusPernikahan,omitempty"`
	Almarhum         bool      `json:"almarhum"`
	TanggalMeninggal *string   `json:"tanggalMeninggal,omitempty"`
	Alamat           *string   `json:"alamat,omitempty"`
	KotaID           *string   `json:"kotaId,omitempty"`
	KecamatanID      *string   `json:"kecamatanId,omitempty"`
	KelurahanID      *string   `json:"kelurahanId,omitempty"`
	Telepon          *string   `json:"telepon,omitempty"`
	Email            *string   `json:"email,omitempty"`
	FotoURL          *string   `json:"fotoUrl,omitempty"`
	Pendidikan       *string   `json:"pendidikan,omitempty"`
	Jurusan          *string   `json:"jurusan,omitempty"`
	Pekerjaan        *string   `json:"pekerjaan,omitempty"`
	Profesi          *string   `json:"profesi,omitempty"`
	Keterampilan     *string   `json:"keterampilan,omitempty"`
	Disabilitas      *string   `json:"disabilitas,omitempty"`
	AktivitasSosial  *string   `json:"aktivitasSosial,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// UmatSearchOutput is the output schema for umat_search.
type UmatSearchOutput struct {
	Umat []MCPUmat `json:"umat"`
}

// UmatGetInput is the input schema for umat_get.
type UmatGetInput struct {
	UmatID string `json:"umatId"`
}

// MCPUmatKeluargaSummary is the PII-masked household summary inside MCPUmatDetail.
type MCPUmatKeluargaSummary struct {
	ID            string  `json:"id"`
	NamaKeluarga  string  `json:"namaKeluarga"`
	NomorKKMasked *string `json:"nomorKkMasked,omitempty"`
}

// MCPUmatSakramen is the MCP representation of a sacramental record.
type MCPUmatSakramen struct {
	ID               string    `json:"id"`
	ParokiID         string    `json:"parokiId"`
	UmatID           string    `json:"umatId"`
	Jenis            string    `json:"jenis"`
	Tanggal          *string   `json:"tanggal,omitempty"`
	KeuskupanID      *string   `json:"keuskupanId,omitempty"`
	ParokiSakramenID *string   `json:"parokiSakramenId,omitempty"`
	KeuskupanNama    *string   `json:"keuskupanNama,omitempty"`
	ParokiNama       *string   `json:"parokiNama,omitempty"`
	Buku             *string   `json:"buku,omitempty"`
	Nomor            *string   `json:"nomor,omitempty"`
	Catatan          *string   `json:"catatan,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// MCPUmatDetail is the PII-masked MCP detail representation of an Umat (never exposes raw NIK or NomorKK).
type MCPUmatDetail struct {
	ID               string                  `json:"id"`
	ParokiID         string                  `json:"parokiId"`
	KeluargaID       *string                 `json:"keluargaId,omitempty"`
	WilayahID        *string                 `json:"wilayahId,omitempty"`
	LingkunganID     *string                 `json:"lingkunganId,omitempty"`
	UserID           *string                 `json:"userId,omitempty"`
	Nama             string                  `json:"nama"`
	NamaBaptis       *string                 `json:"namaBaptis,omitempty"`
	NIKMasked        *string                 `json:"nikMasked,omitempty"`
	JenisKelamin     string                  `json:"jenisKelamin"`
	TempatLahir      *string                 `json:"tempatLahir,omitempty"`
	TanggalLahir     *string                 `json:"tanggalLahir,omitempty"`
	GolonganDarah    *string                 `json:"golonganDarah,omitempty"`
	Suku             *string                 `json:"suku,omitempty"`
	Agama            *string                 `json:"agama,omitempty"`
	Kewarganegaraan  *string                 `json:"kewarganegaraan,omitempty"`
	Negara           *string                 `json:"negara,omitempty"`
	HubunganKeluarga *string                 `json:"hubunganKeluarga,omitempty"`
	StatusPernikahan *string                 `json:"statusPernikahan,omitempty"`
	Almarhum         bool                    `json:"almarhum"`
	TanggalMeninggal *string                 `json:"tanggalMeninggal,omitempty"`
	Alamat           *string                 `json:"alamat,omitempty"`
	KotaID           *string                 `json:"kotaId,omitempty"`
	KecamatanID      *string                 `json:"kecamatanId,omitempty"`
	KelurahanID      *string                 `json:"kelurahanId,omitempty"`
	Telepon          *string                 `json:"telepon,omitempty"`
	Email            *string                 `json:"email,omitempty"`
	FotoURL          *string                 `json:"fotoUrl,omitempty"`
	Pendidikan       *string                 `json:"pendidikan,omitempty"`
	Jurusan          *string                 `json:"jurusan,omitempty"`
	Pekerjaan        *string                 `json:"pekerjaan,omitempty"`
	Profesi          *string                 `json:"profesi,omitempty"`
	Keterampilan     *string                 `json:"keterampilan,omitempty"`
	Disabilitas      *string                 `json:"disabilitas,omitempty"`
	AktivitasSosial  *string                 `json:"aktivitasSosial,omitempty"`
	Keluarga         *MCPUmatKeluargaSummary `json:"keluarga,omitempty"`
	Sakramen         []MCPUmatSakramen       `json:"sakramen"`
	CreatedAt        time.Time               `json:"createdAt"`
	UpdatedAt        time.Time               `json:"updatedAt"`
}

// UmatListKeluargaInput is the input schema for umat_list_keluarga (never accepts raw NomorKK).
type UmatListKeluargaInput struct {
	Query        string `json:"query,omitempty"`
	WilayahID    string `json:"wilayahId,omitempty"`
	LingkunganID string `json:"lingkunganId,omitempty"`
	Aktif        *bool  `json:"aktif,omitempty"`
	Limit        int    `json:"limit,omitempty"`
}

// MCPKeluarga is the PII-masked MCP representation of a Keluarga.
type MCPKeluarga struct {
	ID            string    `json:"id"`
	ParokiID      string    `json:"parokiId"`
	WilayahID     *string   `json:"wilayahId,omitempty"`
	LingkunganID  *string   `json:"lingkunganId,omitempty"`
	NomorKKMasked *string   `json:"nomorKkMasked,omitempty"`
	NamaKeluarga  string    `json:"namaKeluarga"`
	Alamat        *string   `json:"alamat,omitempty"`
	KotaID        *string   `json:"kotaId,omitempty"`
	KecamatanID   *string   `json:"kecamatanId,omitempty"`
	KelurahanID   *string   `json:"kelurahanId,omitempty"`
	KodePos       *string   `json:"kodePos,omitempty"`
	StatusRumah   *string   `json:"statusRumah,omitempty"`
	StatusEkonomi *string   `json:"statusEkonomi,omitempty"`
	Aktif         bool      `json:"aktif"`
	Catatan       *string   `json:"catatan,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// UmatListKeluargaOutput is the output schema for umat_list_keluarga.
type UmatListKeluargaOutput struct {
	Keluarga []MCPKeluarga `json:"keluarga"`
}

// UmatGetKeluargaInput is the input schema for umat_get_keluarga.
type UmatGetKeluargaInput struct {
	KeluargaID string `json:"keluargaId"`
}

// MCPKeluargaAnggotaSummary is the PII-masked member summary inside MCPKeluargaDetail.
type MCPKeluargaAnggotaSummary struct {
	ID               string  `json:"id"`
	Nama             string  `json:"nama"`
	NamaBaptis       *string `json:"namaBaptis,omitempty"`
	NIKMasked        *string `json:"nikMasked,omitempty"`
	JenisKelamin     string  `json:"jenisKelamin"`
	HubunganKeluarga *string `json:"hubunganKeluarga,omitempty"`
	Almarhum         bool    `json:"almarhum"`
}

// MCPKeluargaDetail is the PII-masked MCP detail representation of a Keluarga (never exposes raw NomorKK or NIK).
type MCPKeluargaDetail struct {
	ID            string                      `json:"id"`
	ParokiID      string                      `json:"parokiId"`
	WilayahID     *string                     `json:"wilayahId,omitempty"`
	LingkunganID  *string                     `json:"lingkunganId,omitempty"`
	NomorKKMasked *string                     `json:"nomorKkMasked,omitempty"`
	NamaKeluarga  string                      `json:"namaKeluarga"`
	Alamat        *string                     `json:"alamat,omitempty"`
	KotaID        *string                     `json:"kotaId,omitempty"`
	KecamatanID   *string                     `json:"kecamatanId,omitempty"`
	KelurahanID   *string                     `json:"kelurahanId,omitempty"`
	KodePos       *string                     `json:"kodePos,omitempty"`
	StatusRumah   *string                     `json:"statusRumah,omitempty"`
	StatusEkonomi *string                     `json:"statusEkonomi,omitempty"`
	Aktif         bool                        `json:"aktif"`
	Catatan       *string                     `json:"catatan,omitempty"`
	Anggota       []MCPKeluargaAnggotaSummary `json:"anggota"`
	CreatedAt     time.Time                   `json:"createdAt"`
	UpdatedAt     time.Time                   `json:"updatedAt"`
}

// RegisterMCP registers read-only PII-masked umat_* tools on the MCP server.
func (f *Feature) RegisterMCP(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "umat_search",
		Description: "Search parishioner profiles in the active paroki (PII-masked; requires umat:admin).",
	}, f.mcpUmatSearch)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "umat_get",
		Description: "Get a parishioner profile with family summary and ordered sacraments (PII-masked; requires umat:admin).",
	}, f.mcpUmatGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "umat_list_keluarga",
		Description: "List household (Kartu Keluarga) records in the active paroki (PII-masked; requires umat:admin).",
	}, f.mcpUmatListKeluarga)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "umat_get_keluarga",
		Description: "Get a household (Kartu Keluarga) record and its members (PII-masked; requires umat:admin).",
	}, f.mcpUmatGetKeluarga)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "umat_list_referensi",
		Description: "List active Data Umat reference dictionary entries (global merged with parish overrides).",
	}, f.mcpUmatListReferensi)
}

func (f *Feature) authorizeMCP(ctx context.Context, perm string) (user.Principal, error) {
	p, ok := user.UserContext(ctx)
	if !ok || p.UserID == uuid.Nil {
		return user.Principal{}, httpx.Unauthorized("Authentication is required.")
	}
	tp, err := user.RequireTenant(ctx)
	if err != nil {
		return user.Principal{}, err
	}
	if f.enforcer == nil {
		return user.Principal{}, httpx.Forbidden("Forbidden.")
	}
	dom := authz.ParokiDomain(tp.ParokiID)
	if perm == PermReferensi {
		dom = authz.DomainGlobal
	}
	allowed, err := f.enforcer.Enforce(authz.UserSubject(tp.UserID), dom, perm, authz.ActionRead)
	if err != nil {
		return user.Principal{}, err
	}
	if !allowed {
		return user.Principal{}, httpx.Forbidden("Forbidden.")
	}
	return tp, nil
}

func normalizeMCPLimit(limit int) int {
	if limit <= 0 {
		return 25
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func parseOptUUID(raw, fieldName string) (*uuid.UUID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, httpx.BadRequest("Invalid " + fieldName + ".")
	}
	return &id, nil
}

func uuidPtrStr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func (f *Feature) mcpUmatSearch(ctx context.Context, _ *mcp.CallToolRequest, in UmatSearchInput) (*mcp.CallToolResult, UmatSearchOutput, error) {
	tp, err := f.authorizeMCP(ctx, PermUmat)
	if err != nil {
		return nil, UmatSearchOutput{}, err
	}
	kelID, err := parseOptUUID(in.KeluargaID, "keluargaId")
	if err != nil {
		return nil, UmatSearchOutput{}, err
	}
	wilID, err := parseOptUUID(in.WilayahID, "wilayahId")
	if err != nil {
		return nil, UmatSearchOutput{}, err
	}
	lingID, err := parseOptUUID(in.LingkunganID, "lingkunganId")
	if err != nil {
		return nil, UmatSearchOutput{}, err
	}
	items, _, _, err := f.service.SearchUmat(ctx, tp.ParokiID, domain.SearchUmatRequest{
		Q: in.Query,
		Filter: &domain.UmatSearchFilter{
			KeluargaID:       kelID,
			WilayahID:        wilID,
			LingkunganID:     lingID,
			JenisKelamin:     in.JenisKelamin,
			Agama:            in.Agama,
			StatusPernikahan: in.StatusPernikahan,
			Almarhum:         in.Almarhum,
		},
	}, normalizeMCPLimit(in.Limit))
	if err != nil {
		return nil, UmatSearchOutput{}, err
	}
	out := make([]MCPUmat, 0, len(items))
	for _, it := range items {
		out = append(out, MCPUmat{
			ID:               it.ID.String(),
			ParokiID:         it.ParokiID.String(),
			KeluargaID:       uuidPtrStr(it.KeluargaID),
			WilayahID:        uuidPtrStr(it.WilayahID),
			LingkunganID:     uuidPtrStr(it.LingkunganID),
			UserID:           uuidPtrStr(it.UserID),
			Nama:             it.Nama,
			NamaBaptis:       it.NamaBaptis,
			NIKMasked:        it.NIKMasked,
			JenisKelamin:     it.JenisKelamin,
			TempatLahir:      it.TempatLahir,
			TanggalLahir:     it.TanggalLahir,
			GolonganDarah:    it.GolonganDarah,
			Suku:             it.Suku,
			Agama:            it.Agama,
			Kewarganegaraan:  it.Kewarganegaraan,
			Negara:           it.Negara,
			HubunganKeluarga: it.HubunganKeluarga,
			StatusPernikahan: it.StatusPernikahan,
			Almarhum:         it.Almarhum,
			TanggalMeninggal: it.TanggalMeninggal,
			Alamat:           it.Alamat,
			KotaID:           uuidPtrStr(it.KotaID),
			KecamatanID:      uuidPtrStr(it.KecamatanID),
			KelurahanID:      uuidPtrStr(it.KelurahanID),
			Telepon:          it.Telepon,
			Email:            it.Email,
			FotoURL:          it.FotoURL,
			Pendidikan:       it.Pendidikan,
			Jurusan:          it.Jurusan,
			Pekerjaan:        it.Pekerjaan,
			Profesi:          it.Profesi,
			Keterampilan:     it.Keterampilan,
			Disabilitas:      it.Disabilitas,
			AktivitasSosial:  it.AktivitasSosial,
			CreatedAt:        it.CreatedAt,
			UpdatedAt:        it.UpdatedAt,
		})
	}
	return nil, UmatSearchOutput{Umat: out}, nil
}

func (f *Feature) mcpUmatGet(ctx context.Context, _ *mcp.CallToolRequest, in UmatGetInput) (*mcp.CallToolResult, MCPUmatDetail, error) {
	tp, err := f.authorizeMCP(ctx, PermUmat)
	if err != nil {
		return nil, MCPUmatDetail{}, err
	}
	umatID, err := uuid.Parse(in.UmatID)
	if err != nil {
		return nil, MCPUmatDetail{}, httpx.BadRequest("Invalid umatId.")
	}
	d, err := f.service.GetUmat(ctx, tp.ParokiID, umatID)
	if err != nil {
		return nil, MCPUmatDetail{}, err
	}
	var kel *MCPUmatKeluargaSummary
	if d.Keluarga != nil {
		kel = &MCPUmatKeluargaSummary{
			ID:            d.Keluarga.ID.String(),
			NamaKeluarga:  d.Keluarga.NamaKeluarga,
			NomorKKMasked: d.Keluarga.NomorKKMasked,
		}
	}
	sak := make([]MCPUmatSakramen, 0, len(d.Sakramen))
	for _, s := range d.Sakramen {
		sak = append(sak, MCPUmatSakramen{
			ID:               s.ID.String(),
			ParokiID:         s.ParokiID.String(),
			UmatID:           s.UmatID.String(),
			Jenis:            s.Jenis,
			Tanggal:          s.Tanggal,
			KeuskupanID:      uuidPtrStr(s.KeuskupanID),
			ParokiSakramenID: uuidPtrStr(s.ParokiSakramenID),
			KeuskupanNama:    s.KeuskupanNama,
			ParokiNama:       s.ParokiNama,
			Buku:             s.Buku,
			Nomor:            s.Nomor,
			Catatan:          s.Catatan,
			CreatedAt:        s.CreatedAt,
			UpdatedAt:        s.UpdatedAt,
		})
	}
	return nil, MCPUmatDetail{
		ID:               d.ID.String(),
		ParokiID:         d.ParokiID.String(),
		KeluargaID:       uuidPtrStr(d.KeluargaID),
		WilayahID:        uuidPtrStr(d.WilayahID),
		LingkunganID:     uuidPtrStr(d.LingkunganID),
		UserID:           uuidPtrStr(d.UserID),
		Nama:             d.Nama,
		NamaBaptis:       d.NamaBaptis,
		NIKMasked:        d.NIKMasked,
		JenisKelamin:     d.JenisKelamin,
		TempatLahir:      d.TempatLahir,
		TanggalLahir:     d.TanggalLahir,
		GolonganDarah:    d.GolonganDarah,
		Suku:             d.Suku,
		Agama:            d.Agama,
		Kewarganegaraan:  d.Kewarganegaraan,
		Negara:           d.Negara,
		HubunganKeluarga: d.HubunganKeluarga,
		StatusPernikahan: d.StatusPernikahan,
		Almarhum:         d.Almarhum,
		TanggalMeninggal: d.TanggalMeninggal,
		Alamat:           d.Alamat,
		KotaID:           uuidPtrStr(d.KotaID),
		KecamatanID:      uuidPtrStr(d.KecamatanID),
		KelurahanID:      uuidPtrStr(d.KelurahanID),
		Telepon:          d.Telepon,
		Email:            d.Email,
		FotoURL:          d.FotoURL,
		Pendidikan:       d.Pendidikan,
		Jurusan:          d.Jurusan,
		Pekerjaan:        d.Pekerjaan,
		Profesi:          d.Profesi,
		Keterampilan:     d.Keterampilan,
		Disabilitas:      d.Disabilitas,
		AktivitasSosial:  d.AktivitasSosial,
		Keluarga:         kel,
		Sakramen:         sak,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}, nil
}

func (f *Feature) mcpUmatListKeluarga(ctx context.Context, _ *mcp.CallToolRequest, in UmatListKeluargaInput) (*mcp.CallToolResult, UmatListKeluargaOutput, error) {
	tp, err := f.authorizeMCP(ctx, PermKeluarga)
	if err != nil {
		return nil, UmatListKeluargaOutput{}, err
	}
	wilID, err := parseOptUUID(in.WilayahID, "wilayahId")
	if err != nil {
		return nil, UmatListKeluargaOutput{}, err
	}
	lingID, err := parseOptUUID(in.LingkunganID, "lingkunganId")
	if err != nil {
		return nil, UmatListKeluargaOutput{}, err
	}
	items, _, _, err := f.service.SearchKeluarga(ctx, tp.ParokiID, domain.SearchKeluargaRequest{
		Q: in.Query,
		Filter: &domain.KeluargaSearchFilter{
			WilayahID:    wilID,
			LingkunganID: lingID,
			Aktif:        in.Aktif,
		},
	}, normalizeMCPLimit(in.Limit))
	if err != nil {
		return nil, UmatListKeluargaOutput{}, err
	}
	out := make([]MCPKeluarga, 0, len(items))
	for _, it := range items {
		out = append(out, MCPKeluarga{
			ID:            it.ID.String(),
			ParokiID:      it.ParokiID.String(),
			WilayahID:     uuidPtrStr(it.WilayahID),
			LingkunganID:  uuidPtrStr(it.LingkunganID),
			NomorKKMasked: it.NomorKKMasked,
			NamaKeluarga:  it.NamaKeluarga,
			Alamat:        it.Alamat,
			KotaID:        uuidPtrStr(it.KotaID),
			KecamatanID:   uuidPtrStr(it.KecamatanID),
			KelurahanID:   uuidPtrStr(it.KelurahanID),
			KodePos:       it.KodePos,
			StatusRumah:   it.StatusRumah,
			StatusEkonomi: it.StatusEkonomi,
			Aktif:         it.Aktif,
			Catatan:       it.Catatan,
			CreatedAt:     it.CreatedAt,
			UpdatedAt:     it.UpdatedAt,
		})
	}
	return nil, UmatListKeluargaOutput{Keluarga: out}, nil
}

func (f *Feature) mcpUmatGetKeluarga(ctx context.Context, _ *mcp.CallToolRequest, in UmatGetKeluargaInput) (*mcp.CallToolResult, MCPKeluargaDetail, error) {
	tp, err := f.authorizeMCP(ctx, PermKeluarga)
	if err != nil {
		return nil, MCPKeluargaDetail{}, err
	}
	keluargaID, err := uuid.Parse(in.KeluargaID)
	if err != nil {
		return nil, MCPKeluargaDetail{}, httpx.BadRequest("Invalid keluargaId.")
	}
	d, err := f.service.GetKeluarga(ctx, tp.ParokiID, keluargaID)
	if err != nil {
		return nil, MCPKeluargaDetail{}, err
	}
	anggota := make([]MCPKeluargaAnggotaSummary, 0, len(d.Anggota))
	for _, a := range d.Anggota {
		anggota = append(anggota, MCPKeluargaAnggotaSummary{
			ID:               a.ID.String(),
			Nama:             a.Nama,
			NamaBaptis:       a.NamaBaptis,
			NIKMasked:        a.NIKMasked,
			JenisKelamin:     a.JenisKelamin,
			HubunganKeluarga: a.HubunganKeluarga,
			Almarhum:         a.Almarhum,
		})
	}
	return nil, MCPKeluargaDetail{
		ID:            d.ID.String(),
		ParokiID:      d.ParokiID.String(),
		WilayahID:     uuidPtrStr(d.WilayahID),
		LingkunganID:  uuidPtrStr(d.LingkunganID),
		NomorKKMasked: d.NomorKKMasked,
		NamaKeluarga:  d.NamaKeluarga,
		Alamat:        d.Alamat,
		KotaID:        uuidPtrStr(d.KotaID),
		KecamatanID:   uuidPtrStr(d.KecamatanID),
		KelurahanID:   uuidPtrStr(d.KelurahanID),
		KodePos:       d.KodePos,
		StatusRumah:   d.StatusRumah,
		StatusEkonomi: d.StatusEkonomi,
		Aktif:         d.Aktif,
		Catatan:       d.Catatan,
		Anggota:       anggota,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}, nil
}

// UmatListReferensiInput is the input schema for umat_list_referensi.
type UmatListReferensiInput struct {
	Category string `json:"category" jsonschema:"required,Referensi category (agama, golongan_darah, suku, hubungan_keluarga, status_pernikahan, status_rumah, status_ekonomi, pendidikan, pekerjaan, profesi, disabilitas, aktivitas_sosial)"`
	Query    string `json:"q,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// MCPReferensiItem is a single active Data Umat referensi dictionary entry.
type MCPReferensiItem struct {
	ID        string  `json:"id"`
	ParokiID  *string `json:"parokiId,omitempty"`
	Category  string  `json:"category"`
	Code      string  `json:"code"`
	Label     string  `json:"label"`
	SortOrder int     `json:"sortOrder"`
}

// UmatListReferensiOutput is the output of umat_list_referensi.
type UmatListReferensiOutput struct {
	Items []MCPReferensiItem `json:"items"`
}

func (f *Feature) mcpUmatListReferensi(ctx context.Context, _ *mcp.CallToolRequest, in UmatListReferensiInput) (*mcp.CallToolResult, UmatListReferensiOutput, error) {
	tp, err := f.authorizeMCP(ctx, PermReferensi)
	if err != nil {
		return nil, UmatListReferensiOutput{}, err
	}
	if strings.TrimSpace(in.Category) == "" {
		return nil, UmatListReferensiOutput{}, domain.ValidationError("category", "REQUIRED", "Category is required.")
	}
	activeTrue := true
	items, _, err := f.service.SearchReferensi(ctx, tp, domain.SearchUmatReferensiRequest{
		Q: in.Query,
		Filter: &domain.UmatReferensiSearchFilter{
			Category: in.Category,
			Active:   &activeTrue,
		},
	}, nil, nil, normalizeMCPLimit(in.Limit))
	if err != nil {
		return nil, UmatListReferensiOutput{}, err
	}
	out := make([]MCPReferensiItem, 0, len(items))
	for _, it := range items {
		out = append(out, MCPReferensiItem{
			ID:        it.ID.String(),
			ParokiID:  uuidPtrStr(it.ParokiID),
			Category:  it.Category,
			Code:      it.Code,
			Label:     it.Label,
			SortOrder: it.SortOrder,
		})
	}
	return nil, UmatListReferensiOutput{Items: out}, nil
}
