package service

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/cantus"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/internal/cantus/domain"
	"github.com/paroki/domus/internal/cantus/repo"
	"github.com/paroki/domus/internal/platform/authz"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

// InUseChecker reports whether a cantus entry is currently referenced by another feature.
type InUseChecker func(ctx context.Context, cantusID uuid.UUID) (bool, error)

// Service owns all business logic for the global cantus catalog.
type Service struct {
	repo         *repo.Repository
	enforcer     *authz.Enforcer
	mu           sync.RWMutex
	inUseChecker InUseChecker
}

// New constructs a Service backed by r and enforcer.
func New(r *repo.Repository, enforcer *authz.Enforcer) *Service {
	return &Service{
		repo:     r,
		enforcer: enforcer,
	}
}

// SetInUseChecker registers a cross-feature callback checked before soft-deleting a cantus entry.
func (s *Service) SetInUseChecker(fn InUseChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inUseChecker = fn
}

func (s *Service) getInUseChecker() InUseChecker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.inUseChecker
}

// CanManage reports whether userID holds write permission on cantus:cantus in the global domain.
func (s *Service) CanManage(ctx context.Context, userID uuid.UUID) bool {
	if s.enforcer == nil || userID == uuid.Nil {
		return false
	}
	allowed, err := s.enforcer.Enforce(authz.UserSubject(userID), authz.DomainGlobal, domain.PermCantus, authz.ActionWrite)
	return err == nil && allowed
}

// Search searches non-deleted cantus entries with optional query, filters, and cursor pagination.
func (s *Service) Search(
	ctx context.Context,
	actorID uuid.UUID,
	req domain.SearchCantusRequest,
	cursorCreatedAt *time.Time,
	cursorID *uuid.UUID,
	limit int,
) ([]domain.CantusResponse, bool, error) {
	var filter domain.CantusSearchFilter
	if req.Filter != nil {
		filter = *req.Filter
		filter.Slug = strings.ToLower(strings.TrimSpace(filter.Slug))
		filter.Jenis = strings.ToLower(strings.TrimSpace(filter.Jenis))
		filter.Sumber = strings.TrimSpace(filter.Sumber)
		filter.Kategori = strings.TrimSpace(filter.Kategori)
		filter.Tag = strings.TrimSpace(filter.Tag)
		filter.NadaDasar = strings.TrimSpace(filter.NadaDasar)
	}
	rows, err := s.repo.Search(ctx, strings.TrimSpace(req.Q), filter, cursorCreatedAt, cursorID, limit+1)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	canManage := s.CanManage(ctx, actorID)
	items := make([]domain.CantusResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, toResponse(row, canManage))
	}
	return items, hasMore, nil
}

// Get fetches a single non-deleted cantus entry by ID.
func (s *Service) Get(ctx context.Context, actorID, id uuid.UUID) (domain.CantusResponse, error) {
	row, err := s.repo.FindByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.CantusResponse{}, domain.ErrCantusNotFound
	}
	if err != nil {
		return domain.CantusResponse{}, err
	}
	return toResponse(row, s.CanManage(ctx, actorID)), nil
}

// GetBySlug fetches a single non-deleted cantus entry by normalized slug.
func (s *Service) GetBySlug(ctx context.Context, actorID uuid.UUID, slug string) (domain.CantusResponse, error) {
	normSlug := strings.ToLower(strings.TrimSpace(slug))
	if normSlug == "" {
		return domain.CantusResponse{}, domain.ErrCantusNotFound
	}
	row, err := s.repo.FindBySlug(ctx, normSlug)
	if ent.IsNotFound(err) {
		return domain.CantusResponse{}, domain.ErrCantusNotFound
	}
	if err != nil {
		return domain.CantusResponse{}, err
	}
	return toResponse(row, s.CanManage(ctx, actorID)), nil
}

// Create creates a new cantus entry or reactivates a soft-deleted entry sharing the same slug.
func (s *Service) Create(ctx context.Context, actorID uuid.UUID, req domain.CreateCantusRequest) (domain.CantusResponse, error) {
	normSlug, err := validateSlug(req.Slug)
	if err != nil {
		return domain.CantusResponse{}, err
	}
	jenis, err := validateJenis(req.Jenis)
	if err != nil {
		return domain.CantusResponse{}, err
	}
	judul := strings.TrimSpace(req.Judul)
	if judul == "" {
		return domain.CantusResponse{}, domain.ValidationError("judul", "REQUIRED", "Judul is required.")
	}
	mediaItems, err := validateAndConvertMedia(req.Media)
	if err != nil {
		return domain.CantusResponse{}, err
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}
	var createdBy *uuid.UUID
	if actorID != uuid.Nil {
		createdBy = &actorID
	}

	params := repo.CreateParams{
		Slug:      normSlug,
		Jenis:     jenis,
		Sumber:    trimOptionalPtr(req.Sumber),
		Nomor:     trimOptionalPtr(req.Nomor),
		Judul:     judul,
		Kategori:  trimOptionalPtr(req.Kategori),
		Tag:       normalizeTags(req.Tag),
		NadaDasar: trimOptionalPtr(req.NadaDasar),
		Lirik:     trimOptionalPtr(req.Lirik),
		Kredit:    normalizeKredit(req.Kredit),
		Media:     mediaItems,
		Meta:      req.Meta,
		Active:    active,
		CreatedBy: createdBy,
	}

	var saved *ent.Cantus
	err = s.repo.WithTx(ctx, func(txRepo *repo.Repository) error {
		existing, txErr := txRepo.FindBySlugIncludeDeleted(ctx, normSlug)
		switch {
		case txErr == nil && existing.DeletedAt == nil:
			return domain.ErrCantusCodeConflict
		case txErr == nil && existing.DeletedAt != nil:
			saved, txErr = txRepo.Reactivate(ctx, existing.ID, params)
		case ent.IsNotFound(txErr):
			saved, txErr = txRepo.Create(ctx, params)
		default:
			return txErr
		}
		if txErr != nil {
			if ent.IsConstraintError(txErr) {
				return domain.ErrCantusCodeConflict
			}
			return txErr
		}
		return nil
	})
	if err != nil {
		return domain.CantusResponse{}, err
	}
	return toResponse(saved, s.CanManage(ctx, actorID)), nil
}

// Update partially updates a non-deleted cantus entry.
func (s *Service) Update(ctx context.Context, actorID, id uuid.UUID, req domain.UpdateCantusRequest) (domain.CantusResponse, error) {
	current, err := s.repo.FindByID(ctx, id)
	if ent.IsNotFound(err) {
		return domain.CantusResponse{}, domain.ErrCantusNotFound
	}
	if err != nil {
		return domain.CantusResponse{}, err
	}

	var params repo.UpdateParams
	if req.Slug != nil {
		normSlug, valErr := validateSlug(*req.Slug)
		if valErr != nil {
			return domain.CantusResponse{}, valErr
		}
		if normSlug != current.Slug {
			existing, findErr := s.repo.FindBySlug(ctx, normSlug)
			if findErr == nil && existing.ID != id {
				return domain.CantusResponse{}, domain.ErrCantusCodeConflict
			}
			if findErr != nil && !ent.IsNotFound(findErr) {
				return domain.CantusResponse{}, findErr
			}
		}
		params.Slug = &normSlug
	}
	if req.Jenis != nil {
		jenis, valErr := validateJenis(*req.Jenis)
		if valErr != nil {
			return domain.CantusResponse{}, valErr
		}
		params.Jenis = &jenis
	}
	if req.Judul != nil {
		trimmed := strings.TrimSpace(*req.Judul)
		if trimmed == "" {
			return domain.CantusResponse{}, domain.ValidationError("judul", "REQUIRED", "Judul cannot be empty.")
		}
		params.Judul = &trimmed
	}
	if req.Sumber != nil {
		trimmed := strings.TrimSpace(*req.Sumber)
		params.Sumber = &trimmed
	}
	if req.Nomor != nil {
		trimmed := strings.TrimSpace(*req.Nomor)
		params.Nomor = &trimmed
	}
	if req.Kategori != nil {
		trimmed := strings.TrimSpace(*req.Kategori)
		params.Kategori = &trimmed
	}
	if req.NadaDasar != nil {
		trimmed := strings.TrimSpace(*req.NadaDasar)
		params.NadaDasar = &trimmed
	}
	if req.Lirik != nil {
		trimmed := strings.TrimSpace(*req.Lirik)
		params.Lirik = &trimmed
	}
	if req.Tag != nil {
		params.Tag = normalizeTags(req.Tag)
	}
	if req.Kredit != nil {
		k := normalizeKredit(req.Kredit)
		if k == nil {
			k = map[string]string{}
		}
		params.Kredit = k
	}
	if req.Media != nil {
		mediaItems, valErr := validateAndConvertMedia(req.Media)
		if valErr != nil {
			return domain.CantusResponse{}, valErr
		}
		if mediaItems == nil {
			mediaItems = []schema.CantusMediaItem{}
		}
		params.Media = mediaItems
	}
	if req.Meta != nil {
		params.Meta = req.Meta
	}
	params.Active = req.Active

	updated, err := s.repo.Update(ctx, id, params)
	if ent.IsNotFound(err) {
		return domain.CantusResponse{}, domain.ErrCantusNotFound
	}
	if ent.IsConstraintError(err) {
		return domain.CantusResponse{}, domain.ErrCantusCodeConflict
	}
	if err != nil {
		return domain.CantusResponse{}, err
	}
	return toResponse(updated, s.CanManage(ctx, actorID)), nil
}

// Delete soft-deletes a cantus entry after checking any registered InUseChecker hook.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return domain.ErrCantusNotFound
		}
		return err
	}
	if checker := s.getInUseChecker(); checker != nil {
		inUse, err := checker(ctx, id)
		if err != nil {
			return err
		}
		if inUse {
			return domain.ErrCantusInUse
		}
	}
	return s.repo.SoftDelete(ctx, id, time.Now().UTC())
}

func validateSlug(raw string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(raw))
	if !slugRegex.MatchString(norm) {
		return "", domain.ValidationError("slug", "INVALID_SLUG", "Slug must be 2-64 lowercase alphanumeric characters, hyphens, or underscores.")
	}
	return norm, nil
}

func validateJenis(raw string) (cantus.Jenis, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case domain.JenisLagu:
		return cantus.JenisLagu, nil
	case domain.JenisMazmur:
		return cantus.JenisMazmur, nil
	case domain.JenisProfan:
		return cantus.JenisProfan, nil
	default:
		return "", domain.ValidationError("jenis", "INVALID_JENIS", "Jenis must be one of: lagu, mazmur, profan.")
	}
}

func validateAndConvertMedia(items []domain.CantusMedia) ([]schema.CantusMediaItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]schema.CantusMediaItem, 0, len(items))
	for _, m := range items {
		mType := strings.ToLower(strings.TrimSpace(m.Type))
		switch mType {
		case domain.MediaTypeImage, domain.MediaTypePDF, domain.MediaTypeAudio, domain.MediaTypeVideo:
		default:
			return nil, domain.ValidationError("media", "INVALID_MEDIA_TYPE", "Media type must be one of: image, pdf, audio, video.")
		}
		viewURL := strings.TrimSpace(m.ViewURL)
		if viewURL != "" {
			if err := validateHTTPURL(viewURL); err != nil {
				return nil, err
			}
		}
		downloadURL := strings.TrimSpace(m.DownloadURL)
		if downloadURL != "" {
			if err := validateHTTPURL(downloadURL); err != nil {
				return nil, err
			}
		}
		out = append(out, schema.CantusMediaItem{
			Type:        mType,
			Title:       strings.TrimSpace(m.Title),
			ViewURL:     viewURL,
			DownloadURL: downloadURL,
			S3Key:       strings.TrimSpace(m.S3Key),
		})
	}
	return out, nil
}

func validateHTTPURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return domain.ValidationError("media", "INVALID_URL", "Media URLs must use http:// or https:// with a valid host.")
	}
	return nil
}

func trimOptionalPtr(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		trimmed := strings.TrimSpace(t)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeKredit(k map[string]string) map[string]string {
	if len(k) == 0 {
		return nil
	}
	out := make(map[string]string, len(k))
	for key, val := range k {
		tk := strings.TrimSpace(key)
		tv := strings.TrimSpace(val)
		if tk != "" && tv != "" {
			out[tk] = tv
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func toResponse(c *ent.Cantus, canManage bool) domain.CantusResponse {
	tags := c.Tag
	if tags == nil {
		tags = []string{}
	}
	var media []domain.CantusMedia
	if len(c.Media) > 0 {
		media = make([]domain.CantusMedia, 0, len(c.Media))
		for _, m := range c.Media {
			media = append(media, domain.CantusMedia{
				Type:        m.Type,
				Title:       m.Title,
				ViewURL:     m.ViewURL,
				DownloadURL: m.DownloadURL,
				S3Key:       m.S3Key,
			})
		}
	}
	return domain.CantusResponse{
		ID:        c.ID,
		Slug:      c.Slug,
		Jenis:     string(c.Jenis),
		Sumber:    c.Sumber,
		Nomor:     c.Nomor,
		Judul:     c.Judul,
		Kategori:  c.Kategori,
		Tag:       tags,
		NadaDasar: c.NadaDasar,
		Lirik:     c.Lirik,
		Kredit:    c.Kredit,
		Media:     media,
		Meta:      c.Meta,
		Active:    c.Active,
		CanManage: canManage,
		CreatedBy: c.CreatedBy,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
