package testutil_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/meta"
	"github.com/paroki/domus/ent/schema"
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
