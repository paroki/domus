package testutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/meta"
	"github.com/paroki/domus/ent/schema"
	"github.com/paroki/domus/ent/user"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntClient_MetaCRUDAndMixins(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	// UUIDv7 IDMixin & TimeMixin
	first, err := client.Meta.Create().SetKey("site.name").SetValue("Domus").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), first.ID.Version())
	assert.False(t, first.CreatedAt.IsZero())
	assert.False(t, first.UpdatedAt.IsZero())

	time.Sleep(2 * time.Millisecond)

	second, err := client.Meta.Create().SetKey("site.lang").SetValue("id").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), second.ID.Version())
	assert.Less(t, first.ID.String(), second.ID.String())

	// Unique constraint on key
	_, err = client.Meta.Create().SetKey("site.name").SetValue("Duplicate").Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	// Read & Update
	got, err := client.Meta.Query().Where(meta.KeyEQ("site.name")).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Domus", got.Value)

	time.Sleep(2 * time.Millisecond)
	updated, err := got.Update().SetValue("Paroki").Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Paroki", updated.Value)
	assert.True(t, updated.CreatedAt.Equal(first.CreatedAt))
	assert.True(t, updated.UpdatedAt.After(first.UpdatedAt) || updated.UpdatedAt.Equal(first.UpdatedAt))

	// Delete
	require.NoError(t, client.Meta.DeleteOneID(updated.ID).Exec(ctx))
	exists, err := client.Meta.Query().Where(meta.ID(updated.ID)).Exist(ctx)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMixins_SoftDeleteAndNewID(t *testing.T) {
	id := schema.NewID()
	assert.Equal(t, uuid.Version(7), id.Version())
	assert.Len(t, schema.SoftDeleteMixin{}.Fields(), 1)
}

func TestUserSchema_SoftDeleteAndUniqueEmail(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("alice@paroki.org").
		SetName("Alice").
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), u.ID.Version())
	assert.Equal(t, user.StatusActive, u.Status)
	assert.Nil(t, u.DeletedAt)

	_, err = client.User.Create().
		SetEmail("alice@paroki.org").
		SetName("Alice Clone").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))

	deletedAt := time.Now().UTC().Truncate(time.Second)
	u, err = u.Update().SetDeletedAt(deletedAt).Save(ctx)
	require.NoError(t, err)
	require.NotNil(t, u.DeletedAt)
	assert.True(t, u.DeletedAt.Equal(deletedAt))
}

func TestUserAccountSchema_UniqueProviderSubject(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("bob@paroki.org").
		SetName("Bob").
		Save(ctx)
	require.NoError(t, err)

	providers := []string{"google", "github", "email", "telegram", "whatsapp"}
	for _, provider := range providers {
		acc, err := client.UserAccount.Create().
			SetUserID(u.ID).
			SetProvider(provider).
			SetSubject("shared-subject-123").
			Save(ctx)
		require.NoError(t, err, "provider %s should allow same subject", provider)
		assert.Equal(t, u.ID, acc.UserID)
	}

	_, err = client.UserAccount.Create().
		SetUserID(u.ID).
		SetProvider("google").
		SetSubject("shared-subject-123").
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestUserSessionSchema_MultiDeviceAndUniqueTokenHash(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	u, err := client.User.Create().
		SetEmail("carol@paroki.org").
		SetName("Carol").
		Save(ctx)
	require.NoError(t, err)

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	s1, err := client.UserSession.Create().
		SetUserID(u.ID).
		SetTokenHash("hash-device-1").
		SetUserAgent("Firefox/130").
		SetIPAddress("127.0.0.1").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, u.ID, s1.UserID)

	s2, err := client.UserSession.Create().
		SetUserID(u.ID).
		SetTokenHash("hash-device-2").
		SetUserAgent("Mobile/1.0").
		SetIPAddress("10.0.0.2").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.NotEqual(t, s1.ID, s2.ID)

	sessions, err := u.QuerySessions().All(ctx)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)

	_, err = client.UserSession.Create().
		SetUserID(u.ID).
		SetTokenHash("hash-device-1").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}

func TestUserOTPSchema_Persistence(t *testing.T) {
	ctx := context.Background()
	client := testutil.NewEntClient(t)

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	otp, err := client.UserOTP.Create().
		SetTarget("dave@paroki.org").
		SetCodeHash("sha256-otp-hash").
		SetExpiresAt(expiresAt).
		Save(ctx)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), otp.ID.Version())
	assert.Equal(t, "email", otp.Channel)
	assert.Equal(t, 0, otp.Attempts)
	assert.Equal(t, "dave@paroki.org", otp.Target)
	assert.Equal(t, "sha256-otp-hash", otp.CodeHash)
}
