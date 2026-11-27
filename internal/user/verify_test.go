package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/ref"
	refdomain "github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/user"
	"github.com/paroki/domus/internal/user/domain"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

type stubIdentityExtractor struct {
	result domain.IdentityExtraction
	err    error
}

func (s *stubIdentityExtractor) ExtractIdentityDocument(_ context.Context, _ domain.IdentityExtractInput) (domain.IdentityExtraction, error) {
	if s.err != nil {
		return domain.IdentityExtraction{}, s.err
	}
	return s.result, nil
}

type stubCensusVerifier struct {
	previewFn func(ctx context.Context, parokiID, userID uuid.UUID, ext domain.IdentityExtraction) (domain.VerifyPreviewResult, error)
	confirmFn func(ctx context.Context, in domain.VerifyConfirmInput) (domain.VerifyConfirmResult, error)
	profileFn func(ctx context.Context, parokiID, userID uuid.UUID) (*domain.SelfUmatProfile, error)
}

func (s *stubCensusVerifier) PreviewVerification(ctx context.Context, parokiID, userID uuid.UUID, ext domain.IdentityExtraction) (domain.VerifyPreviewResult, error) {
	if s.previewFn != nil {
		return s.previewFn(ctx, parokiID, userID, ext)
	}
	return domain.VerifyPreviewResult{MatchStatus: "new"}, nil
}

func (s *stubCensusVerifier) VerifyAndSyncCensus(ctx context.Context, in domain.VerifyConfirmInput) (domain.VerifyConfirmResult, error) {
	if s.confirmFn != nil {
		return s.confirmFn(ctx, in)
	}
	return domain.VerifyConfirmResult{}, nil
}

func (s *stubCensusVerifier) GetSelfProfile(ctx context.Context, parokiID, userID uuid.UUID) (*domain.SelfUmatProfile, error) {
	if s.profileFn != nil {
		return s.profileFn(ctx, parokiID, userID)
	}
	return nil, nil
}

type UserVerifySuite struct {
	testutil.ApiTestSuite[any]
	client      *ent.Client
	enforcer    *authz.Enforcer
	sender      *stubSender
	userFeature *user.Feature
	refFeature  *ref.Feature
	extractor   *stubIdentityExtractor
	verifier    *stubCensusVerifier
}

func TestUserVerifySuite(t *testing.T) {
	suite.Run(t, new(UserVerifySuite))
}

func (s *UserVerifySuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	s.client = testutil.NewEntClient(s.T())
	enforcer, err := authz.NewEnforcer(s.client)
	s.Require().NoError(err)
	s.enforcer = enforcer
	s.sender = &stubSender{}
	s.extractor = &stubIdentityExtractor{}
	s.verifier = &stubCensusVerifier{}

	cfg := user.Config{
		SuperAdminEmail: "admin@paroki.org",
		SuperAdminName:  "Super Admin",
		OTPTTL:          5 * time.Minute,
		OTPMaxAttempts:  5,
		OTPCooldown:     60 * time.Second,
		SessionTTL:      24 * time.Hour,
		CookieName:      "domus_session",
	}
	s.userFeature = user.New(s.client, s.enforcer, cfg, slog.New(slog.DiscardHandler), user.WithOTPSender(s.sender))
	s.refFeature = ref.New(s.client, s.enforcer, s.userFeature)

	s.userFeature.SetIdentityExtractor(s.extractor)
	s.userFeature.SetCensusVerifier(s.verifier)

	ctx := context.Background()
	s.Require().NoError(authz.RegisterFeatures(ctx, s.enforcer, s.userFeature, s.refFeature))
	s.Require().NoError(s.userFeature.Bootstrap(ctx))

	s.WithFeatures(s.userFeature, s.refFeature)
}

func (s *UserVerifySuite) loginViaOTP(email string) {
	s.ClearCookies()
	s.Request("/api/auth/otp", http.MethodPost, domain.RequestOTPRequest{Email: email})
	s.OK()
	s.Request("/api/auth/otp/verify", http.MethodPost, domain.VerifyOTPRequest{
		Email: email,
		Code:  s.sender.lastCode,
	})
	s.OK()
	s.Cookies = s.HttpResponse.Cookies()
}

func (s *UserVerifySuite) postMultipart(path string, fields map[string]string, files map[string][]byte) {
	s.T().Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		s.Require().NoError(w.WriteField(k, v))
	}
	for field, data := range files {
		part, err := w.CreateFormFile(field, field+".jpg")
		s.Require().NoError(err)
		_, err = part.Write(data)
		s.Require().NoError(err)
	}
	s.Require().NoError(w.Close())

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, path, &buf)
	for _, c := range s.Cookies {
		req.AddCookie(c)
	}
	for k, vals := range s.Headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := s.App.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	s.Require().NoError(err)
	s.HttpResponse = resp
}

func (s *UserVerifySuite) createParokiAndJoin(email string) uuid.UUID {
	ctx := context.Background()
	s.loginViaOTP("admin@paroki.org")
	s.Request("/api/ref/keuskupan", http.MethodPost, refdomain.CreateKeuskupanRequest{
		Kode: "KAJ-" + uuid.NewString()[:4],
		Nama: "Keuskupan Verify",
	})
	s.Created()
	keuskupan := testutil.DecodeAs[refdomain.KeuskupanResponse](s).Data

	s.Request("/api/ref/paroki", http.MethodPost, refdomain.CreateParokiRequest{
		KeuskupanID: keuskupan.ID,
		Kode:        "PRK-" + uuid.NewString()[:6],
		Nama:        "Paroki Verify",
	})
	s.Created()
	paroki := testutil.DecodeAs[refdomain.ParokiResponse](s).Data

	u, err := s.userFeature.EnsureUser(ctx, email, "User", "", "email", email)
	s.Require().NoError(err)
	_, err = s.refFeature.EnsureParokiMember(ctx, paroki.ID, u.ID)
	s.Require().NoError(err)

	s.loginViaOTP(email)
	s.SetHeader("X-Paroki-Id", paroki.ID.String())
	return paroki.ID
}

func (s *UserVerifySuite) TestVerifyEndpoints_Validation_Confirm_GetMe_RateLimit() {
	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01}
	pngBytes := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n', 0, 0, 0, 0}

	parokiID := s.createParokiAndJoin("budi@paroki.org")

	// 1. Validation: without ktp, kk is required
	s.postMultipart("/api/me/verify/extract", nil, map[string][]byte{
		"kartuPelajar": jpegBytes,
	})
	s.UnprocessableEntity()

	// 2. Validation: > 2 MiB rejected with 422
	tooBig := make([]byte, 2*1024*1024+1)
	copy(tooBig, jpegBytes)
	s.postMultipart("/api/me/verify/extract", nil, map[string][]byte{
		"ktp": tooBig,
	})
	s.UnprocessableEntity()

	// 3. Validation: invalid mime (plain text) rejected with 422
	s.postMultipart("/api/me/verify/extract", nil, map[string][]byte{
		"ktp": []byte("not an image"),
	})
	s.UnprocessableEntity()

	// 4. Happy path: KTP only extract -> new, confirm -> verified=true + GET /api/me enriched
	genderPria := "pria"
	birth := "1992-05-10"
	umatID := uuid.New()
	kelID := uuid.New()
	s.extractor.result = domain.IdentityExtraction{
		KTP: &domain.ExtractedKTP{
			NIK:          "3174011005920001",
			Nama:         "Budi Santoso",
			JenisKelamin: &genderPria,
			TanggalLahir: &birth,
		},
	}
	s.verifier.confirmFn = func(_ context.Context, in domain.VerifyConfirmInput) (domain.VerifyConfirmResult, error) {
		s.Require().Equal(parokiID, in.ParokiID)
		s.Require().NotEmpty(in.Documents)
		return domain.VerifyConfirmResult{
			UmatID:       umatID,
			NomorInduk:   "NIU-2026-000001",
			KeluargaID:   &kelID,
			Nama:         "Budi Santoso",
			JenisKelamin: "pria",
			TanggalLahir: &birth,
		}, nil
	}
	nik := "3174011005920001"
	kk := "3174011005920099"
	s.verifier.profileFn = func(_ context.Context, _, _ uuid.UUID) (*domain.SelfUmatProfile, error) {
		return &domain.SelfUmatProfile{
			Umat: &domain.SelfUmatInfo{
				ID:           umatID,
				NomorInduk:   "NIU-2026-000001",
				Nama:         "Budi Santoso",
				NIK:          &nik,
				JenisKelamin: "pria",
			},
			Keluarga: &domain.SelfKeluargaInfo{
				ID:              kelID,
				NomorRegistrasi: "KK-2026-000001",
				NomorKK:         &kk,
				NamaKeluarga:    "Keluarga Santoso",
			},
			Members: []domain.SelfUmatInfo{{ID: umatID, NomorInduk: "NIU-2026-000001", Nama: "Budi Santoso", JenisKelamin: "pria"}},
			Dokumen: []domain.SelfDokumenMeta{{ID: uuid.New(), Jenis: "ktp", MimeType: "image/jpeg", SizeBytes: len(jpegBytes)}},
		}, nil
	}

	s.postMultipart("/api/me/verify/extract", nil, map[string][]byte{
		"ktp": jpegBytes,
	})
	s.OK()
	extResp := testutil.DecodeAs[domain.VerifyExtractResponse](s).Data
	s.Equal("new", extResp.MatchStatus)

	payloadBytes, err := json.Marshal(s.extractor.result)
	s.Require().NoError(err)
	s.postMultipart("/api/me/verify/confirm", map[string]string{
		"payload": string(payloadBytes),
	}, map[string][]byte{
		"ktp": pngBytes,
	})
	s.OK()
	confResp := testutil.DecodeAs[domain.VerifyConfirmResult](s).Data
	s.True(confResp.Verified)
	s.Equal("NIU-2026-000001", confResp.NomorInduk)

	s.Request("/api/me", http.MethodGet, nil)
	s.OK()
	me := testutil.DecodeAs[domain.UserResponse](s).Data
	s.True(me.Verified)
	s.NotNil(me.VerifiedAt)
	s.Require().NotNil(me.UmatID)
	s.Equal(umatID, *me.UmatID)
	s.Require().NotNil(me.NomorInduk)
	s.Equal("NIU-2026-000001", *me.NomorInduk)
	s.Equal("Budi Santoso", me.Name)

	s.Request("/api/me/umat", http.MethodGet, nil)
	s.OK()
	profile := testutil.DecodeAs[domain.SelfUmatProfile](s).Data
	s.Require().NotNil(profile.Umat)
	s.Equal("3174011005920001", *profile.Umat.NIK)
	s.Require().NotNil(profile.Keluarga)
	s.Equal("3174011005920099", *profile.Keluarga.NomorKK)
	s.Len(profile.Dokumen, 1)

	// 5. Rate limit: 9 more extract calls succeed (1 already used above = 10 total), 11th returns 429 VERIFY_RATE_LIMITED
	for range 9 {
		s.postMultipart("/api/me/verify/extract", nil, map[string][]byte{"ktp": jpegBytes})
		s.OK()
	}
	s.postMultipart("/api/me/verify/extract", nil, map[string][]byte{"ktp": jpegBytes})
	s.Equal(http.StatusTooManyRequests, s.HttpResponse.StatusCode)
	s.Equal("VERIFY_RATE_LIMITED", s.GetErrorResponse().Error.Code)

	// 6. Conflict propagated from CensusVerifier
	s.verifier.confirmFn = func(_ context.Context, _ domain.VerifyConfirmInput) (domain.VerifyConfirmResult, error) {
		return domain.VerifyConfirmResult{}, errx.Conflict("UMAT_USER_ALREADY_LINKED", "User is already linked.")
	}
	s.postMultipart("/api/me/verify/confirm", map[string]string{
		"payload": string(payloadBytes),
	}, map[string][]byte{
		"ktp": jpegBytes,
	})
	s.Conflict()
	s.Equal("UMAT_USER_ALREADY_LINKED", s.GetErrorResponse().Error.Code)
}
