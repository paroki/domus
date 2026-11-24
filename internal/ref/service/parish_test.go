package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paroki/domus/internal/platform/authz"
	"github.com/paroki/domus/internal/ref/domain"
	"github.com/paroki/domus/internal/ref/repo"
	"github.com/paroki/domus/internal/ref/service"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestDeleteParish_AgentChildrenGuard(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	enforcer, err := authz.NewEnforcer(db)
	require.NoError(t, err)

	svc := service.New(repo.New(db), enforcer, nil)

	diocese, err := db.Diocese.Create().SetCode("kaj").SetName("KAJ").Save(ctx)
	require.NoError(t, err)

	userID := uuid.New()

	// 1. Parish with active AgentSession -> 409 REF_IN_USE
	p1, err := db.Parish.Create().SetDioceseID(diocese.ID).SetCode("p1").SetName("Paroki 1").Save(ctx)
	require.NoError(t, err)
	sess, err := db.AgentSession.Create().
		SetParishID(p1.ID).
		SetUserID(userID).
		SetLastMessageAt(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	err = svc.DeleteParish(ctx, p1.ID)
	require.ErrorIs(t, err, domain.ErrInUse)

	// Soft-deleting the session unblocks parish soft-delete
	require.NoError(t, db.AgentSession.UpdateOneID(sess.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))
	require.NoError(t, svc.DeleteParish(ctx, p1.ID))

	// 2. Parish with AgentMemory -> 409 REF_IN_USE
	p2, err := db.Parish.Create().SetDioceseID(diocese.ID).SetCode("p2").SetName("Paroki 2").Save(ctx)
	require.NoError(t, err)
	_, err = db.AgentMemory.Create().
		SetParishID(p2.ID).
		SetUserID(userID).
		SetContent("Preferensi bahasa Indonesia").
		Save(ctx)
	require.NoError(t, err)

	err = svc.DeleteParish(ctx, p2.ID)
	require.ErrorIs(t, err, domain.ErrInUse)
}
