package config

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/paroki/domus/api/ent/enttest"
	"github.com/paroki/domus/api/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureEntCli_AutoSetCreatorAndUpdater(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent_hook_test?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ConfigureEntCli(client)

	user1ID := core.GenerateID()
	user2ID := core.GenerateID()

	// Seed users first because Diocese has foreign key edges to User
	_, err := client.User.Create().
		SetID(user1ID).
		SetName("User One").
		SetEmail("user1@example.com").
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.User.Create().
		SetID(user2ID).
		SetName("User Two").
		SetEmail("user2@example.com").
		Save(context.Background())
	require.NoError(t, err)

	// Test 1: Create Diocese with User 1 in context (creator and updater auto-set)
	ctxUser1 := core.ContextWithUser(context.Background(), core.AuthenticatedUser{
		ID:    user1ID,
		Name:  "User One",
		Email: "user1@example.com",
	})

	diocese, err := client.Diocese.Create().
		SetName("Keuskupan Surabaya").
		Save(ctxUser1)
	require.NoError(t, err)
	assert.Equal(t, user1ID, diocese.CreatedBy)
	assert.Equal(t, user1ID, diocese.UpdatedBy)

	// Test 2: Update Diocese with User 2 in context (updater auto-set)
	ctxUser2 := core.ContextWithUser(context.Background(), core.AuthenticatedUser{
		ID:    user2ID,
		Name:  "User Two",
		Email: "user2@example.com",
	})

	updatedDiocese, err := client.Diocese.UpdateOneID(diocese.ID).
		SetName("Keuskupan Surabaya Updated").
		Save(ctxUser2)
	require.NoError(t, err)
	assert.Equal(t, user1ID, updatedDiocese.CreatedBy)
	assert.Equal(t, user2ID, updatedDiocese.UpdatedBy)
}
