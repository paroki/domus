package testutil_test

import (
	"context"
	"testing"

	"github.com/paroki/domus/api/testutil"
)

func TestNewTestDB_UserCRUD(t *testing.T) {
	client, closeDB := testutil.NewTestDB(t)
	defer closeDB()

	ctx := context.Background()

	// 1. Create a user
	u, err := client.User.Create().
		SetEmail("test@example.com").
		SetStatus("PENDING").
		Save(ctx)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if u.Email != "test@example.com" {
		t.Errorf("expected email to be test@example.com, got %s", u.Email)
	}
	if u.Status != "PENDING" {
		t.Errorf("expected status to be PENDING, got %s", u.Status)
	}
	if u.ID.String() == "" {
		t.Errorf("expected generated UUID, got empty")
	}

	// 2. Query the user back
	got, err := client.User.Get(ctx, u.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if got.Email != "test@example.com" {
		t.Errorf("expected queried email to be test@example.com, got %s", got.Email)
	}
}
