package setting

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/paroki/domus/ent"
	entsetting "github.com/paroki/domus/ent/setting"
	"github.com/paroki/domus/internal/platform/crypt"
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
	Key          string `json:"key"`
	Group        string `json:"group"`
	Label        string `json:"label"`
	Description  string `json:"description,omitempty"`
	Secret       bool   `json:"secret"`
	Multiline    bool   `json:"multiline"`
	Placeholder  string `json:"placeholder,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// SettingResponse represents a setting item returned by GET/PATCH /api/settings.
type SettingResponse struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Secret      bool   `json:"secret"`
	Multiline   bool   `json:"multiline"`
	Placeholder string `json:"placeholder,omitempty"`
	Value       string `json:"value"`
	Configured  bool   `json:"configured"`
	Overridden  bool   `json:"overridden"`
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

// StoreOption configures optional Store behavior.
type StoreOption func(*Store)

// WithCipher enables transparent at-rest encryption for Secret=true settings.
func WithCipher(c *crypt.Cipher) StoreOption {
	return func(s *Store) {
		s.cipher = c
	}
}

// Store manages registered setting descriptors, DB persistence, in-memory cache,
// and hot-reload listeners.
type Store struct {
	db          *ent.Client
	cipher      *crypt.Cipher
	writeMu     sync.Mutex
	mu          sync.RWMutex
	descriptors []Descriptor
	descIndex   map[string]int
	values      map[string]string
	listeners   []OnChangeFunc
}

// NewStore constructs a setting Store backed by db.
func NewStore(db *ent.Client, opts ...StoreOption) *Store {
	s := &Store{
		db:        db,
		descIndex: make(map[string]int),
		values:    make(map[string]string),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
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
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	s.mu.RLock()
	cipher := s.cipher
	descIndex := maps.Clone(s.descIndex)
	descriptors := slices.Clone(s.descriptors)
	s.mu.RUnlock()

	loaded := make(map[string]string)
	if s.db != nil {
		rows, err := s.db.Setting.Query().All(ctx)
		if err != nil {
			return fmt.Errorf("load settings: %w", err)
		}
		for _, row := range rows {
			if row.Value == "" {
				continue
			}
			val := row.Value
			if idx, ok := descIndex[row.Key]; ok && descriptors[idx].Secret && cipher != nil && crypt.IsEncryptedString(val) {
				dec, err := cipher.DecryptString(val)
				if err != nil {
					return fmt.Errorf("decrypt setting %s: %w", row.Key, err)
				}
				val = dec
			}
			loaded[row.Key] = val
		}
	}

	s.mu.Lock()
	s.values = loaded
	listeners := make([]OnChangeFunc, len(s.listeners))
	copy(listeners, s.listeners)
	snapshot := s.effectiveMapLocked(s.values)
	s.mu.Unlock()

	for _, fn := range listeners {
		if err := fn(ctx, snapshot); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the effective cached value for key, falling back to Descriptor.DefaultValue.
func (s *Store) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v := s.values[key]; v != "" {
		return v
	}
	if idx, ok := s.descIndex[key]; ok {
		return s.descriptors[idx].DefaultValue
	}
	return ""
}

func (s *Store) effectiveMapLocked(values map[string]string) map[string]string {
	out := make(map[string]string, len(s.descriptors)+len(values))
	for _, d := range s.descriptors {
		if d.DefaultValue != "" {
			out[d.Key] = d.DefaultValue
		}
	}
	for k, v := range values {
		if v != "" {
			out[k] = v
		}
	}
	return out
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
		val := raw
		if val == "" && d.DefaultValue != "" {
			val = d.DefaultValue
		}
		configured := raw != "" || (d.DefaultValue != "" && !d.Secret)
		if d.Secret && raw != "" {
			val = MaskedSecret
		}
		out = append(out, SettingResponse{
			Key:         d.Key,
			Group:       d.Group,
			Label:       d.Label,
			Description: d.Description,
			Secret:      d.Secret,
			Multiline:   d.Multiline,
			Placeholder: d.Placeholder,
			Value:       val,
			Configured:  configured,
			Overridden:  raw != "" && !d.Secret,
		})
	}
	return out
}

// Update validates and applies partial setting updates, invoking OnChange listeners
// before persisting to the database so invalid configurations never corrupt state.
func (s *Store) Update(ctx context.Context, updates map[string]string) ([]SettingResponse, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	s.mu.RLock()
	nextValues := maps.Clone(s.values)
	if nextValues == nil {
		nextValues = make(map[string]string)
	}
	toPersist := make(map[string]string)
	for k, v := range updates {
		idx, ok := s.descIndex[k]
		if !ok {
			s.mu.RUnlock()
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
	effectiveNext := s.effectiveMapLocked(nextValues)
	listeners := slices.Clone(s.listeners)
	descIndex := maps.Clone(s.descIndex)
	descriptors := slices.Clone(s.descriptors)
	cipher := s.cipher
	s.mu.RUnlock()

	for _, fn := range listeners {
		if err := fn(ctx, maps.Clone(effectiveNext)); err != nil {
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
			dbVal := v
			if idx, ok := descIndex[k]; ok && descriptors[idx].Secret && cipher != nil && !crypt.IsEncryptedString(v) {
				enc, err := cipher.EncryptString(v)
				if err != nil {
					return nil, fmt.Errorf("encrypt setting %s: %w", k, err)
				}
				dbVal = enc
			}
			affected, err := s.db.Setting.Update().Where(entsetting.KeyEQ(k)).SetValue(dbVal).Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update setting %s: %w", k, err)
			}
			if affected == 0 {
				if _, err := s.db.Setting.Create().SetKey(k).SetValue(dbVal).Save(ctx); err != nil {
					return nil, fmt.Errorf("create setting %s: %w", k, err)
				}
			}
		}
	}

	s.mu.Lock()
	if s.values == nil {
		s.values = make(map[string]string)
	}
	for k, v := range toPersist {
		if v == "" {
			delete(s.values, k)
		} else {
			s.values[k] = v
		}
	}
	res := s.listLocked()
	s.mu.Unlock()
	return res, nil
}

// PersistValue persists a single registered key directly to the database and cache
// without re-triggering OnChange listeners.
func (s *Store) PersistValue(ctx context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.descIndex[key]
	if !ok {
		return fmt.Errorf("unknown setting key %q", key)
	}
	desc := s.descriptors[idx]
	trimmed := strings.TrimSpace(value)

	if s.db != nil {
		if trimmed == "" {
			if _, err := s.db.Setting.Delete().Where(entsetting.KeyEQ(key)).Exec(ctx); err != nil {
				return fmt.Errorf("delete setting %s: %w", key, err)
			}
		} else {
			dbVal := trimmed
			if desc.Secret && s.cipher != nil && !crypt.IsEncryptedString(trimmed) {
				enc, err := s.cipher.EncryptString(trimmed)
				if err != nil {
					return fmt.Errorf("encrypt setting %s: %w", key, err)
				}
				dbVal = enc
			}
			affected, err := s.db.Setting.Update().Where(entsetting.KeyEQ(key)).SetValue(dbVal).Save(ctx)
			if err != nil {
				return fmt.Errorf("update setting %s: %w", key, err)
			}
			if affected == 0 {
				if _, err := s.db.Setting.Create().SetKey(key).SetValue(dbVal).Save(ctx); err != nil {
					return fmt.Errorf("create setting %s: %w", key, err)
				}
			}
		}
	}

	if s.values == nil {
		s.values = make(map[string]string)
	}
	if trimmed == "" {
		delete(s.values, key)
	} else {
		s.values[key] = trimmed
	}
	return nil
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
