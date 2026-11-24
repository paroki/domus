package setting_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	entsetting "github.com/paroki/domus/ent/setting"
	"github.com/paroki/domus/internal/platform/authx"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/httpx"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/suite"
)

func TestKey(t *testing.T) {
	if got := setting.Key("user", "telegram.bot_token"); got != "user.telegram.bot_token" {
		t.Fatalf("unexpected key: %q", got)
	}
	for _, bad := range [][2]string{
		{"User", "foo"},
		{"user", "Foo"},
		{"user", ""},
		{"", "foo"},
		{"user_x", "foo"},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for Key(%q, %q)", bad[0], bad[1])
				}
			}()
			setting.Key(bad[0], bad[1])
		}()
	}
}

type mockSettingFeature struct {
	called bool
}

func (m *mockSettingFeature) RegisterSettings(s *setting.Store) {
	m.called = true
	s.Register(setting.Descriptor{
		Key:   setting.Key("user", "telegram.bot_token"),
		Group: "Telegram Bot",
		Label: "Bot Token",
	})
}

func TestRegisterFeatures(t *testing.T) {
	s := setting.NewStore(nil)
	mf := &mockSettingFeature{}
	setting.RegisterFeatures(s, mf, struct{}{})
	if !mf.called {
		t.Fatal("expected RegisterSettings to be called")
	}
	if len(s.List(context.Background())) != 1 {
		t.Fatalf("expected 1 registered setting")
	}
}

type SettingSuite struct {
	testutil.ApiTestSuite[SettingSuite]
	store    *setting.Store
	enforcer *authz.Enforcer
	adminID  uuid.UUID
	userID   uuid.UUID
}

func TestSettingSuite(t *testing.T) {
	suite.Run(t, new(SettingSuite))
}

func (s *SettingSuite) SetupTest() {
	s.ApiTestSuite.SetupTest()
	ctx := context.Background()
	db := testutil.NewEntClient(s.T())

	enforcer, err := authz.NewMemoryEnforcer()
	s.Require().NoError(err)
	s.enforcer = enforcer

	permUsers := authz.Permission("user", "users")
	roleAdmin := authz.Role("user", "super-admin")
	s.Require().NoError(enforcer.AddPermissionForRole(ctx, roleAdmin, authz.DomainGlobal, permUsers, authz.ActionRead))
	s.Require().NoError(enforcer.AddPermissionForRole(ctx, roleAdmin, authz.DomainGlobal, permUsers, authz.ActionWrite))

	s.adminID = uuid.New()
	s.userID = uuid.New()
	s.Require().NoError(enforcer.AssignRole(ctx, authz.UserSubject(s.adminID), roleAdmin, authz.DomainGlobal))

	store := setting.NewStore(db)
	store.Register(setting.Descriptor{
		Key:    setting.Key("user", "telegram.bot_username"),
		Group:  "Telegram Bot",
		Label:  "Bot Username",
		Secret: false,
	})
	store.Register(setting.Descriptor{
		Key:    setting.Key("user", "telegram.bot_token"),
		Group:  "Telegram Bot",
		Label:  "Bot Token",
		Secret: true,
	})
	store.OnChange(func(_ context.Context, values map[string]string) error {
		if values["user.telegram.bot_username"] == "invalid" {
			return errors.New("username tidak valid")
		}
		return nil
	})
	s.Require().NoError(store.Load(ctx))
	s.store = store

	authMW := func(c fiber.Ctx) error {
		switch c.Get("X-Test-User") {
		case "admin":
			c.SetContext(authx.WithPrincipal(c.Context(), authx.Principal{UserID: s.adminID}))
		case "user":
			c.SetContext(authx.WithPrincipal(c.Context(), authx.Principal{UserID: s.userID}))
		}
		return c.Next()
	}
	requireAuth := func(c fiber.Ctx) error {
		if p, ok := authx.FromContext(c.Context()); !ok || p.UserID == uuid.Nil {
			return httpx.Unauthorized("unauthorized")
		}
		return c.Next()
	}

	feat := setting.NewFeature(
		store,
		authMW,
		requireAuth,
		enforcer.Require(permUsers, authz.ActionRead),
		enforcer.Require(permUsers, authz.ActionWrite),
	)
	s.WithFeatures(feat)
}

func (s *SettingSuite) TestAuthzAndCRUD() {
	// Unauthenticated -> 401
	s.Request("/api/settings", http.MethodGet, nil)
	s.Unauthorized()

	// Non-super-admin -> 403
	s.SetHeader("X-Test-User", "user")
	s.Request("/api/settings", http.MethodGet, nil)
	s.Forbidden()
	s.Request("/api/settings", http.MethodPatch, map[string]any{
		"values": map[string]string{"user.telegram.bot_username": "domus_bot"},
	})
	s.Forbidden()

	// Super Admin -> 200 GET empty
	s.SetHeader("X-Test-User", "admin")
	s.Request("/api/settings", http.MethodGet, nil)
	s.OK()
	items := testutil.DecodeAs[[]setting.SettingResponse](s).Data
	s.Len(items, 2)
	s.False(items[0].Configured)
	s.Equal("", items[1].Value)
	s.False(items[1].Configured)

	// Unknown key -> 400 SETTING_KEY_UNKNOWN
	s.Request("/api/settings", http.MethodPatch, map[string]any{
		"values": map[string]string{"user.unknown_key": "abc"},
	})
	s.BadRequest()
	errEnv := s.GetErrorResponse()
	s.Equal("SETTING_KEY_UNKNOWN", errEnv.Error.Code)

	// Update valid values -> 200, secret masked
	s.Request("/api/settings", http.MethodPatch, map[string]any{
		"values": map[string]string{
			"user.telegram.bot_username": "domus_bot",
			"user.telegram.bot_token":    "secret-123",
		},
	})
	s.OK()
	items = testutil.DecodeAs[[]setting.SettingResponse](s).Data
	s.Equal("domus_bot", items[0].Value)
	s.True(items[0].Configured)
	s.Equal(setting.MaskedSecret, items[1].Value)
	s.True(items[1].Configured)
	s.Equal("secret-123", s.store.Get("user.telegram.bot_token"))

	// Masked secret leaves existing secret unchanged
	s.Request("/api/settings", http.MethodPatch, map[string]any{
		"values": map[string]string{
			"user.telegram.bot_username": "domus_bot_2",
			"user.telegram.bot_token":    setting.MaskedSecret,
		},
	})
	s.OK()
	s.Equal("domus_bot_2", s.store.Get("user.telegram.bot_username"))
	s.Equal("secret-123", s.store.Get("user.telegram.bot_token"))

	// Validation error -> 422 SETTING_VALIDATION_ERROR, state unchanged
	s.Request("/api/settings", http.MethodPatch, map[string]any{
		"values": map[string]string{
			"user.telegram.bot_username": "invalid",
		},
	})
	s.UnprocessableEntity()
	errEnv = s.GetErrorResponse()
	s.Equal("SETTING_VALIDATION_ERROR", errEnv.Error.Code)
	s.Equal("domus_bot_2", s.store.Get("user.telegram.bot_username"))

	// Clear values -> empty & unconfigured
	s.Request("/api/settings", http.MethodPatch, map[string]any{
		"values": map[string]string{
			"user.telegram.bot_username": "",
			"user.telegram.bot_token":    "",
		},
	})
	s.OK()
	items = testutil.DecodeAs[[]setting.SettingResponse](s).Data
	s.False(items[0].Configured)
	s.False(items[1].Configured)
	s.Equal("", s.store.Get("user.telegram.bot_token"))
}

func (s *SettingSuite) TestWithCipher_DefaultValue_And_PersistValue() {
	ctx := context.Background()
	db := testutil.NewEntClient(s.T())
	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	s.Require().NoError(err)

	store := setting.NewStore(db, setting.WithCipher(cipher))
	store.Register(setting.Descriptor{
		Key:    setting.Key("agent", "provider.gemini.api_key"),
		Group:  "AI Provider",
		Label:  "Gemini API Key",
		Secret: true,
	})
	store.Register(setting.Descriptor{
		Key:          setting.Key("agent", "gabriel.model"),
		Group:        "AI Agent",
		Label:        "Gabriel Model",
		Multiline:    true,
		DefaultValue: "gemini-2.5-flash",
	})

	var onChangeCount int
	var lastEffective map[string]string
	store.OnChange(func(_ context.Context, values map[string]string) error {
		onChangeCount++
		lastEffective = values
		return nil
	})

	s.Require().NoError(store.Load(ctx))
	s.Equal(1, onChangeCount)
	s.Equal("gemini-2.5-flash", store.Get("agent.gabriel.model"))
	s.Equal("gemini-2.5-flash", lastEffective["agent.gabriel.model"])

	listed := store.List(ctx)
	s.Require().Len(listed, 2)
	s.False(listed[0].Overridden)
	s.True(listed[1].Multiline)
	s.True(listed[1].Configured)
	s.False(listed[1].Overridden)

	// Update secret -> encrypted in DB, plaintext in Get/OnChange, masked in List
	listed, err = store.Update(ctx, map[string]string{
		"agent.provider.gemini.api_key": "AIza-super-secret",
		"agent.gabriel.model":           "gemini-2.5-pro",
	})
	s.Require().NoError(err)
	s.Equal(2, onChangeCount)
	s.Equal("AIza-super-secret", store.Get("agent.provider.gemini.api_key"))
	s.Equal("AIza-super-secret", lastEffective["agent.provider.gemini.api_key"])
	s.Equal("gemini-2.5-pro", store.Get("agent.gabriel.model"))
	s.False(listed[0].Overridden)
	s.True(listed[1].Overridden)

	row, err := db.Setting.Query().Where(entsetting.KeyEQ("agent.provider.gemini.api_key")).Only(ctx)
	s.Require().NoError(err)
	s.True(crypt.IsEncryptedString(row.Value))
	s.NotEqual("AIza-super-secret", row.Value)

	// Reload into a fresh store -> decrypts from DB
	store2 := setting.NewStore(db, setting.WithCipher(cipher))
	store2.Register(setting.Descriptor{
		Key:    setting.Key("agent", "provider.gemini.api_key"),
		Group:  "AI Provider",
		Label:  "Gemini API Key",
		Secret: true,
	})
	store2.Register(setting.Descriptor{
		Key:          setting.Key("agent", "gabriel.model"),
		Group:        "AI Agent",
		Label:        "Gabriel Model",
		DefaultValue: "gemini-2.5-flash",
	})
	s.Require().NoError(store2.Load(ctx))
	s.Equal("AIza-super-secret", store2.Get("agent.provider.gemini.api_key"))

	// PersistValue updates DB and cache without triggering OnChange
	s.Require().NoError(store.PersistValue(ctx, "agent.gabriel.model", "gemini-2.5-flash-lite"))
	s.Equal(2, onChangeCount)
	s.Equal("gemini-2.5-flash-lite", store.Get("agent.gabriel.model"))

	// Clearing setting with DefaultValue deletes override and restores default
	listed, err = store.Update(ctx, map[string]string{
		"agent.gabriel.model": "",
	})
	s.Require().NoError(err)
	s.Equal("gemini-2.5-flash", store.Get("agent.gabriel.model"))
	s.False(listed[1].Overridden)
}
