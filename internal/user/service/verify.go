package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/user/domain"
)

const (
	maxVerifyImageBytes   = 2 * 1024 * 1024
	maxVerifyExtractPerHr = 10
)

func validateVerifyImage(field string, img *domain.IdentityImageInput) error {
	if img == nil {
		return nil
	}
	if len(img.Data) == 0 {
		return domain.ValidationError(field, "REQUIRED", field+" file is empty.")
	}
	if len(img.Data) > maxVerifyImageBytes {
		return domain.ValidationError(field, "FILE_TOO_LARGE", field+" exceeds 2 MiB maximum size.")
	}
	detected := http.DetectContentType(img.Data)
	switch detected {
	case "image/jpeg", "image/png", "image/webp":
		img.MimeType = detected
		return nil
	default:
		return domain.ValidationError(field, "INVALID_MIME", field+" must be image/jpeg, image/png, or image/webp.")
	}
}

func (s *Service) checkVerifyExtractRateLimit(userID uuid.UUID, now time.Time) error {
	s.verifyRateMu.Lock()
	defer s.verifyRateMu.Unlock()

	if s.verifyExtractHits == nil {
		s.verifyExtractHits = make(map[uuid.UUID][]time.Time)
	}
	cutoff := now.Add(-time.Hour)
	hits := s.verifyExtractHits[userID]
	kept := hits[:0]
	for _, ts := range hits {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= maxVerifyExtractPerHr {
		s.verifyExtractHits[userID] = kept
		return domain.ErrVerifyRateLimited
	}
	kept = append(kept, now)
	s.verifyExtractHits[userID] = kept
	return nil
}

// ExtractIdentity validates uploaded identity images, enforces rate limiting, runs AI vision extraction, and previews census matching.
func (s *Service) ExtractIdentity(
	ctx context.Context,
	p authx.Principal,
	in domain.IdentityExtractInput,
) (domain.VerifyExtractResponse, error) {
	if p.ParokiID == uuid.Nil {
		return domain.VerifyExtractResponse{}, authx.ErrTenantRequired
	}
	if in.KTP == nil && in.KK == nil {
		return domain.VerifyExtractResponse{}, domain.ValidationError("kk", "REQUIRED", "KK is required when KTP is not provided.")
	}
	if err := validateVerifyImage("ktp", in.KTP); err != nil {
		return domain.VerifyExtractResponse{}, err
	}
	if err := validateVerifyImage("kartuPelajar", in.KartuPelajar); err != nil {
		return domain.VerifyExtractResponse{}, err
	}
	if err := validateVerifyImage("kk", in.KK); err != nil {
		return domain.VerifyExtractResponse{}, err
	}
	if s.identityExtractor == nil {
		return domain.VerifyExtractResponse{}, domain.ErrVerifyUnavailable
	}
	if err := s.checkVerifyExtractRateLimit(p.UserID, time.Now().UTC()); err != nil {
		return domain.VerifyExtractResponse{}, err
	}

	ext, err := s.identityExtractor.ExtractIdentityDocument(ctx, in)
	if err != nil {
		return domain.VerifyExtractResponse{}, err
	}

	matchStatus := "new"
	if s.censusVerifier != nil {
		preview, err := s.censusVerifier.PreviewVerification(ctx, p.ParokiID, p.UserID, ext)
		if err != nil {
			return domain.VerifyExtractResponse{}, err
		}
		if preview.MatchStatus != "" {
			matchStatus = preview.MatchStatus
		}
	}

	return domain.VerifyExtractResponse{
		Extraction:  ext,
		MatchStatus: matchStatus,
	}, nil
}

// ConfirmVerification re-validates reviewed identity data and documents, syncs census records, and marks the user verified.
func (s *Service) ConfirmVerification(
	ctx context.Context,
	p authx.Principal,
	in domain.VerifyConfirmInput,
) (domain.VerifyConfirmResult, error) {
	if p.ParokiID == uuid.Nil {
		return domain.VerifyConfirmResult{}, authx.ErrTenantRequired
	}
	if s.censusVerifier == nil {
		return domain.VerifyConfirmResult{}, domain.ErrVerifyUnavailable
	}

	in.ParokiID = p.ParokiID
	in.UserID = p.UserID

	res, err := s.censusVerifier.VerifyAndSyncCensus(ctx, in)
	if err != nil {
		return domain.VerifyConfirmResult{}, err
	}

	var genderPtr *user.Gender
	switch res.JenisKelamin {
	case string(user.GenderPria):
		g := user.GenderPria
		genderPtr = &g
	case string(user.GenderWanita):
		g := user.GenderWanita
		genderPtr = &g
	}

	now := time.Now().UTC()
	if _, err := s.repo.MarkUserVerified(ctx, p.UserID, res.Nama, genderPtr, res.TanggalLahir, now); err != nil {
		return domain.VerifyConfirmResult{}, err
	}

	res.Verified = true
	res.VerifiedAt = &now
	return res, nil
}

// GetMyUmat returns the caller's linked umat, keluarga (with decrypted NIK/nomorKk), members, and document metadata.
func (s *Service) GetMyUmat(ctx context.Context, p authx.Principal) (*domain.SelfUmatProfile, error) {
	if p.ParokiID == uuid.Nil {
		return nil, authx.ErrTenantRequired
	}
	if s.censusVerifier == nil {
		return nil, domain.ErrVerifyUnavailable
	}
	profile, err := s.censusVerifier.GetSelfProfile(ctx, p.ParokiID, p.UserID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return &domain.SelfUmatProfile{
			Members: []domain.SelfUmatInfo{},
			Dokumen: []domain.SelfDokumenMeta{},
		}, nil
	}
	return profile, nil
}
