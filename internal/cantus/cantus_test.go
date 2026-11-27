package cantus_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/cantus"
	"github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type CantusFeatureSuite struct {
	testutil.ApiTestSuite[domain.CantusResponse]
	client        *ent.Client
	enforcer      *authz.Enforcer
	userFeature   *user.Feature
	cantusFeature *cantus.Feature
	cfg           user.Config
}

func TestCantusFeatureSuite(t *testing.T) {
	suite.Run(t, new(CantusFeatureSuite))
}

func (s *CantusFeatureSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.cfg = user.Config{
		SuperAdminEmail: "super@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     60 * time.Second,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.userFeature = user.New(s.client, s.enforcer, s.cfg, slog.New(slog.DiscardHandler))
	s.cantusFeature = cantus.New(s.client, s.enforcer, s.userFeature)
	s.Require().NoError(authz.RegisterFeatures(context.Background(), s.enforcer, s.userFeature, s.cantusFeature))
	s.Require().NoError(s.userFeature.Bootstrap(context.Background()))
	s.WithFeatures(s.userFeature, s.cantusFeature)
}

func (s *CantusFeatureSuite) issueSessionFor(email string) *http.Cookie {
	ctx := context.Background()
	u, err := s.client.User.Query().Where().All(ctx)
	s.Require().NoError(err)
	var userID uuid.UUID
	for _, row := range u {
		if row.Email == email {
			userID = row.ID
			break
		}
	}
	s.Require().NotEqual(uuid.Nil, userID)
	rawToken := "tok-" + uuid.NewString()
	sum := sha256.Sum256([]byte(rawToken))
	_, err = s.client.UserSession.Create().
		SetUserID(userID).
		SetTokenHash(hex.EncodeToString(sum[:])).
		SetUserAgent("cantus-test").
		SetIPAddress("127.0.0.1").
		SetLastActiveAt(time.Now().UTC()).
		SetExpiresAt(time.Now().UTC().Add(time.Hour)).
		Save(ctx)
	s.Require().NoError(err)
	return &http.Cookie{Name: s.cfg.CookieName, Value: rawToken}
}

func (s *CantusFeatureSuite) createUserWithRoles(email, name string, cantusAdmin bool) (uuid.UUID, *http.Cookie) {
	ctx := context.Background()
	u, err := s.userFeature.EnsureUser(ctx, email, name, "", "email", email)
	s.Require().NoError(err)
	if cantusAdmin {
		s.Require().NoError(s.enforcer.AssignRole(ctx, authz.UserSubject(u.ID), cantus.RoleAdmin, authz.DomainGlobal))
	}
	return u.ID, s.issueSessionFor(email)
}

func (s *CantusFeatureSuite) TestRBAC_G8LeastPrivilegeAndCrossParishAccess() {
	superCookie := s.issueSessionFor("super@paroki.org")
	_, memberCookie := s.createUserWithRoles("member@paroki.org", "Regular Member", false)
	_, cantusAdminCookie := s.createUserWithRoles("cantus@paroki.org", "Cantus Admin", true)

	payload := map[string]any{
		"slug":  "ps-542-tuhan-kasihanilah",
		"jenis": "lagu",
		"judul": "Tuhan Kasihanilah Kami",
	}

	// Regular user (without cantus:admin) -> 403 CANTUS_FORBIDDEN
	s.SetCookie(memberCookie)
	s.Request("/api/cantus", http.MethodPost, payload)
	s.Forbidden()
	s.Equal("CANTUS_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Super Admin (without cantus:admin) -> 403 CANTUS_FORBIDDEN (G8 strict least privilege)
	s.ClearCookies()
	s.SetCookie(superCookie)
	s.Request("/api/cantus", http.MethodPost, payload)
	s.Forbidden()
	s.Equal("CANTUS_FORBIDDEN", s.GetErrorResponse().Error.Code)

	// Cantus Admin -> 201 Created without X-Paroki-Id header
	s.ClearCookies()
	s.SetCookie(cantusAdminCookie)
	s.Request("/api/cantus", http.MethodPost, payload)
	s.Created()
	created := s.GetResponse().Data
	s.True(created.CanManage)
	s.Equal("ps-542-tuhan-kasihanilah", created.Slug)

	// Regular user can read without X-Paroki-Id and sees canManage=false
	s.ClearCookies()
	s.SetCookie(memberCookie)
	s.Request("/api/cantus/"+created.ID.String(), http.MethodGet, nil)
	s.OK()
	fetched := s.GetResponse().Data
	s.False(fetched.CanManage)
	s.Equal(created.ID, fetched.ID)

	// Super Admin without cantus:admin cannot PATCH or DELETE
	s.ClearCookies()
	s.SetCookie(superCookie)
	s.Request("/api/cantus/"+created.ID.String(), http.MethodPatch, map[string]any{"judul": "Hack"})
	s.Forbidden()
	s.Equal("CANTUS_FORBIDDEN", s.GetErrorResponse().Error.Code)

	s.Request("/api/cantus/"+created.ID.String(), http.MethodDelete, nil)
	s.Forbidden()
	s.Equal("CANTUS_FORBIDDEN", s.GetErrorResponse().Error.Code)
}

func (s *CantusFeatureSuite) TestCRUD_ConflictReactivateInUseAndFacade() {
	_, adminCookie := s.createUserWithRoles("cantus-crud@paroki.org", "Cantus CRUD", true)
	s.SetCookie(adminCookie)

	sumber := "PS"
	nomor := "PS-321"
	kategori := "Pembuka"
	nada := "1=D"
	lirik := "Datanglah ya Roh Kudus"
	s.Request("/api/cantus", http.MethodPost, domain.CreateCantusRequest{
		Slug:      "  PS-321_Roh_Kudus ",
		Jenis:     "lagu",
		Sumber:    &sumber,
		Nomor:     &nomor,
		Judul:     "Datanglah Ya Roh Kudus",
		Kategori:  &kategori,
		Tag:       []string{"pentakosta", " ", "pentakosta", "pembuka"},
		NadaDasar: &nada,
		Lirik:     &lirik,
		Kredit:    map[string]string{"syair": "Tim Puji Syukur"},
		Media: []domain.CantusMedia{
			{Type: "pdf", Title: "Partitur", ViewURL: "https://cdn.example.com/ps321.pdf"},
		},
	})
	s.Created()
	created := s.GetResponse().Data
	s.Equal("ps-321_roh_kudus", created.Slug)
	s.Equal([]string{"pentakosta", "pembuka"}, created.Tag)
	s.True(created.Active)

	// Cross-feature GetCantus facade
	facadeRow, err := s.cantusFeature.GetCantus(context.Background(), created.ID)
	s.Require().NoError(err)
	s.Equal(created.Slug, facadeRow.Slug)

	// Duplicate active slug -> 409 CANTUS_CODE_CONFLICT
	s.Request("/api/cantus", http.MethodPost, map[string]any{
		"slug":  "ps-321_roh_kudus",
		"jenis": "lagu",
		"judul": "Duplicate",
	})
	s.Conflict()
	s.Equal("CANTUS_CODE_CONFLICT", s.GetErrorResponse().Error.Code)

	// Partial PATCH update
	newJudul := "Datanglah Roh Kudus (Revisi)"
	activeFalse := false
	s.Request("/api/cantus/"+created.ID.String(), http.MethodPatch, domain.UpdateCantusRequest{
		Judul:  &newJudul,
		Active: &activeFalse,
	})
	s.OK()
	updated := s.GetResponse().Data
	s.Equal(newJudul, updated.Judul)
	s.False(updated.Active)

	// InUseChecker blocks DELETE with 409 CANTUS_IN_USE
	s.cantusFeature.SetInUseChecker(func(_ context.Context, id uuid.UUID) (bool, error) {
		return id == created.ID, nil
	})
	s.Request("/api/cantus/"+created.ID.String(), http.MethodDelete, nil)
	s.Conflict()
	s.Equal("CANTUS_IN_USE", s.GetErrorResponse().Error.Code)

	// Release InUseChecker -> DELETE returns 204
	s.cantusFeature.SetInUseChecker(func(_ context.Context, _ uuid.UUID) (bool, error) {
		return false, nil
	})
	s.Request("/api/cantus/"+created.ID.String(), http.MethodDelete, nil)
	s.NoContent()

	// GET after soft-delete -> 404 CANTUS_NOT_FOUND
	s.Request("/api/cantus/"+created.ID.String(), http.MethodGet, nil)
	s.NotFound()
	s.Equal("CANTUS_NOT_FOUND", s.GetErrorResponse().Error.Code)

	// Recreating with same slug reactivates in-place
	s.Request("/api/cantus", http.MethodPost, map[string]any{
		"slug":  "ps-321_roh_kudus",
		"jenis": "mazmur",
		"judul": "Mazmur Reactivated",
	})
	s.Created()
	reactivated := s.GetResponse().Data
	s.Equal(created.ID, reactivated.ID)
	s.Equal("mazmur", reactivated.Jenis)
	s.Equal("Mazmur Reactivated", reactivated.Judul)
	s.True(reactivated.Active)
	s.Nil(reactivated.Sumber)
}

func (s *CantusFeatureSuite) TestValidation_MediaURLSchemeAndSlug() {
	_, adminCookie := s.createUserWithRoles("cantus-val@paroki.org", "Cantus Val", true)
	s.SetCookie(adminCookie)

	// Invalid slug
	s.Request("/api/cantus", http.MethodPost, map[string]any{
		"slug":  "Invalid Slug!",
		"jenis": "lagu",
		"judul": "Test",
	})
	s.UnprocessableEntity()
	s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)

	// Invalid media URLs (javascript:, data:, relative)
	for _, badURL := range []string{
		"javascript:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"/relative/path.pdf",
		"ftp://files.example.com/a.mp3",
	} {
		s.Request("/api/cantus", http.MethodPost, map[string]any{
			"slug":  "valid-slug-test",
			"jenis": "lagu",
			"judul": "Valid Title",
			"media": []map[string]any{
				{"type": "pdf", "viewUrl": badURL},
			},
		})
		s.UnprocessableEntity()
		s.Equal("VALIDATION_ERROR", s.GetErrorResponse().Error.Code)
	}
}

func (s *CantusFeatureSuite) TestSearch_FiltersTagContainmentAndCursor() {
	_, adminCookie := s.createUserWithRoles("cantus-search@paroki.org", "Cantus Search", true)
	s.SetCookie(adminCookie)

	entries := []map[string]any{
		{
			"slug":      "ps-001-pembuka",
			"jenis":     "lagu",
			"sumber":    "PS",
			"nomor":     "PS-001",
			"judul":     "Bersyukurlah Kepada Tuhan",
			"kategori":  "Pembuka",
			"tag":       []string{"syukur", "paskah"},
			"nadaDasar": "D",
			"kredit":    map[string]string{"lagu": "Antonius"},
		},
		{
			"slug":      "mz-010-tuhan-gembalaku",
			"jenis":     "mazmur",
			"sumber":    "Mazmur",
			"nomor":     "MT-010",
			"judul":     "Tuhan Adalah Gembalaku",
			"kategori":  "Mazmur Tanggapan",
			"tag":       []string{"gembala", "paskah"},
			"nadaDasar": "C",
			"lirik":     "Sekalipun aku berjalan dalam lembah kelam",
		},
		{
			"slug":   "profan-001",
			"jenis":  "profan",
			"judul":  "Mars Lingkungan",
			"tag":    []string{"komunitas"},
			"active": false,
		},
	}
	for _, e := range entries {
		s.Request("/api/cantus", http.MethodPost, e)
		s.Created()
	}

	// JSONB tag containment search ("gembala")
	s.Request("/api/cantus/search", http.MethodPost, domain.SearchCantusRequest{
		Filter: &domain.CantusSearchFilter{Tag: "gembala"},
	})
	s.OK()
	tagRes := s.PagedResponse().Data
	s.Len(tagRes, 1)
	s.Equal("mz-010-tuhan-gembalaku", tagRes[0].Slug)

	// Text search q matching kredit ("Antonius")
	s.Request("/api/cantus/search", http.MethodPost, domain.SearchCantusRequest{Q: "Antonius"})
	s.OK()
	qRes := s.PagedResponse().Data
	s.Len(qRes, 1)
	s.Equal("ps-001-pembuka", qRes[0].Slug)

	// Keyset pagination limit=1 + cursor mismatch -> 400 INVALID_CURSOR
	s.Request("/api/cantus/search", http.MethodPost, domain.SearchCantusRequest{Limit: 1})
	s.OK()
	page1 := s.PagedResponse()
	s.Len(page1.Data, 1)
	s.True(page1.Meta.HasMore)
	s.NotEmpty(page1.Meta.NextCursor)

	// Valid cursor with same query
	s.Request("/api/cantus/search", http.MethodPost, domain.SearchCantusRequest{
		Cursor: page1.Meta.NextCursor,
		Limit:  1,
	})
	s.OK()
	page2 := s.PagedResponse()
	s.Len(page2.Data, 1)
	s.NotEqual(page1.Data[0].ID, page2.Data[0].ID)

	// Cursor mismatch with different filter -> 400 INVALID_CURSOR
	s.Request("/api/cantus/search", http.MethodPost, domain.SearchCantusRequest{
		Q:      "different",
		Cursor: page1.Meta.NextCursor,
		Limit:  1,
	})
	s.BadRequest()
	s.Equal("INVALID_CURSOR", s.GetErrorResponse().Error.Code)
}
