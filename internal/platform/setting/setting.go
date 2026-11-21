package setting

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/ent"
	entsetting "github.com/paroki/domus/ent/setting"
	"github.com/paroki/domus/internal/platform/httpx"
)

// MaskedSecret is the placeholder returned for configured secret settings and
// accepted on PATCH to leave an existing secret unchanged.
const MaskedSecret = "••••••••"

var keyRe = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9_]*)+$`)

// Key builds and validates a namespaced setting key (<scope>.<name>).
// It panics if the resulting key does not match ^[a-z][a-z0-9-]*(\.[a-z][a-z0-9_]*)+$.
func Key(scope, name string) string {
	k := scope + "." + name
	if !keyRe.MatchString(k) {
		panic(fmt.Sprintf("setting.Key: invalid key %q", k))
	}
	return k
}

// Descriptor defines metadata for a registered application setting.
type Descriptor struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret"`
	Placeholder string `json:"placeholder,omitempty"`
}

// SettingResponse represents a setting item returned by GET/PATCH /api/settings.
type SettingResponse struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret"`
	Placeholder string `json:"placeholder,omitempty"`
	Value       string `json:"value"`
	Configured  bool   `json:"configured"`
}

// UpdateSettingsRequest is the payload for PATCH /api/settings.
type UpdateSettingsRequest struct {
	Values map[string]string `json:"values"`
}

// SettingFeature is implemented by feature facades that register their own settings.
type SettingFeature interface {
	RegisterSettings(s *Store)
}

// RegisterFeatures calls RegisterSettings on every feature implementing SettingFeature.
func RegisterFeatures(s *Store, features ...any) {
	for _, f := range features {
		if sf, ok := f.(SettingFeature); ok {
			sf.RegisterSettings(s)
		}
	}
}

// OnChangeFunc is called on Load and before Update persists changes.
// Returning an error aborts Update with 422 SETTING_VALIDATION_ERROR.
type OnChangeFunc func(ctx context.Context, values map[string]string) error

// Store manages registered setting descriptors, DB persistence, in-memory cache,
// and hot-reload listeners.
type Store struct {
	db          *ent.Client
	mu          sync.RWMutex
	descriptors []Descriptor
	descIndex   map[string]int
	values      map[string]string
	listeners   []OnChangeFunc
}

// NewStore constructs a setting Store backed by db.
func NewStore(db *ent.Client) *Store {
	return &Store{
		db:        db,
		descIndex: make(map[string]int),
		values:    make(map[string]string),
	}
}

// Register records a setting Descriptor in registration order, deduplicated by Key.
func (s *Store) Register(desc Descriptor) {
	if !keyRe.MatchString(desc.Key) {
		panic(fmt.Sprintf("setting.Register: invalid key %q", desc.Key))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx, exists := s.descIndex[desc.Key]; exists {
		s.descriptors[idx] = desc
		return
	}
	s.descIndex[desc.Key] = len(s.descriptors)
	s.descriptors = append(s.descriptors, desc)
}

// OnChange registers a listener invoked on Load and before Update persists changes.
func (s *Store) OnChange(fn OnChangeFunc) {
	if fn == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, fn)
}

// Load reads all persisted settings from the database into the cache and
// invokes all registered OnChange listeners.
func (s *Store) Load(ctx context.Context) error {
	loaded := make(map[string]string)
	if s.db != nil {
		rows, err := s.db.Setting.Query().All(ctx)
		if err != nil {
			return fmt.Errorf("load settings: %w", err)
		}
		for _, row := range rows {
			if row.Value != "" {
				loaded[row.Key] = row.Value
			}
		}
	}

	s.mu.Lock()
	s.values = loaded
	listeners := make([]OnChangeFunc, len(s.listeners))
	copy(listeners, s.listeners)
	snapshot := maps.Clone(s.values)
	s.mu.Unlock()

	for _, fn := range listeners {
		if err := fn(ctx, maps.Clone(snapshot)); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the raw cached value for key, or "" if unset.
func (s *Store) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values[key]
}

// List returns all registered settings in registration order with secrets masked.
func (s *Store) List(_ context.Context) []SettingResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listLocked()
}

func (s *Store) listLocked() []SettingResponse {
	out := make([]SettingResponse, 0, len(s.descriptors))
	for _, d := range s.descriptors {
		raw := s.values[d.Key]
		configured := raw != ""
		val := raw
		if d.Secret && configured {
			val = MaskedSecret
		}
		out = append(out, SettingResponse{
			Key:         d.Key,
			Group:       d.Group,
			Label:       d.Label,
			Description: d.Description,
			Secret:      d.Secret,
			Placeholder: d.Placeholder,
			Value:       val,
			Configured:  configured,
		})
	}
	return out
}

// Update validates and applies partial setting updates, invoking OnChange listeners
// before persisting to the database so invalid configurations never corrupt state.
func (s *Store) Update(ctx context.Context, updates map[string]string) ([]SettingResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nextValues := maps.Clone(s.values)
	if nextValues == nil {
		nextValues = make(map[string]string)
	}

	toPersist := make(map[string]string)
	for k, v := range updates {
		idx, ok := s.descIndex[k]
		if !ok {
			return nil, httpx.NewError(
				fiber.StatusBadRequest,
				"SETTING_KEY_UNKNOWN",
				fmt.Sprintf("Kunci pengaturan tidak dikenal: %s", k),
			)
		}
		desc := s.descriptors[idx]
		if desc.Secret && v == MaskedSecret {
			continue
		}
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			delete(nextValues, k)
			toPersist[k] = ""
		} else {
			nextValues[k] = trimmed
			toPersist[k] = trimmed
		}
	}

	for _, fn := range s.listeners {
		if err := fn(ctx, maps.Clone(nextValues)); err != nil {
			return nil, httpx.NewError(
				fiber.StatusUnprocessableEntity,
				"SETTING_VALIDATION_ERROR",
				err.Error(),
			)
		}
	}

	if s.db != nil {
		for k, v := range toPersist {
			if v == "" {
				if _, err := s.db.Setting.Delete().Where(entsetting.KeyEQ(k)).Exec(ctx); err != nil {
					return nil, fmt.Errorf("delete setting %s: %w", k, err)
				}
				continue
			}
			affected, err := s.db.Setting.Update().Where(entsetting.KeyEQ(k)).SetValue(v).Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update setting %s: %w", k, err)
			}
			if affected == 0 {
				if _, err := s.db.Setting.Create().SetKey(k).SetValue(v).Save(ctx); err != nil {
					return nil, fmt.Errorf("create setting %s: %w", k, err)
				}
			}
		}
	}

	s.values = nextValues
	return s.listLocked(), nil
}

// Feature exposes HTTP endpoints for reading and updating application settings.
type Feature struct {
	store        *Store
	authMW       fiber.Handler
	requireAuth  fiber.Handler
	requireRead  fiber.Handler
	requireWrite fiber.Handler
}

// NewFeature constructs the settings HTTP Feature.
func NewFeature(
	store *Store,
	authMW fiber.Handler,
	requireAuth fiber.Handler,
	requireRead fiber.Handler,
	requireWrite fiber.Handler,
) *Feature {
	return &Feature{
		store:        store,
		authMW:       authMW,
		requireAuth:  requireAuth,
		requireRead:  requireRead,
		requireWrite: requireWrite,
	}
}

// Register mounts /settings endpoints under /api.
func (f *Feature) Register(api fiber.Router) {
	g := api.Group("/settings", f.authMW, f.requireAuth)
	g.Get("", f.requireRead, f.List)
	g.Patch("", f.requireWrite, f.Update)
}

// List returns all registered application settings with secrets masked.
//
//	@Summary		List application settings
//	@Description	Returns all registered application settings with secret values masked.
//	@ID				listSettings
//	@Tags			Settings
//	@Produce		json
//	@Success		200	{object}	httpx.Envelope[[]SettingResponse]
//	@Failure		401	{object}	httpx.Envelope[any]
//	@Failure		403	{object}	httpx.Envelope[any]
//	@Router			/settings [get]
func (f *Feature) List(c fiber.Ctx) error {
	return httpx.OK(c, f.store.List(c.Context()))
}

// Update validates and persists application setting updates.
//
//	@Summary		Update application settings
//	@Description	Updates application settings, validates configuration listeners, and hot-reloads active clients.
//	@ID				updateSettings
//	@Tags			Settings
//	@Accept			json
//	@Produce		json
//	@Param			body	body		UpdateSettingsRequest	true	"Setting key-value updates"
//	@Success		200		{object}	httpx.Envelope[[]SettingResponse]
//	@Failure		400		{object}	httpx.Envelope[any]
//	@Failure		401		{object}	httpx.Envelope[any]
//	@Failure		403		{object}	httpx.Envelope[any]
//	@Failure		422		{object}	httpx.Envelope[any]
//	@Router			/settings [patch]
func (f *Feature) Update(c fiber.Ctx) error {
	var req UpdateSettingsRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return httpx.BadRequest("Payload JSON tidak valid.")
	}
	if req.Values == nil {
		return httpx.ValidationError(httpx.FieldError{
			Field:   "values",
			Code:    "required",
			Message: "values wajib diisi.",
		})
	}
	items, err := f.store.Update(c.Context(), req.Values)
	if err != nil {
		return err
	}
	return httpx.OK(c, items)
}
