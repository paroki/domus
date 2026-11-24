package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionLockWaitsWhenCancelling(t *testing.T) {
	svc := New(nil, nil, nil, nil, nil, nil)
	sid := uuid.New()

	require.True(t, svc.acquireSessionLock(context.Background(), sid))

	// Active turn (not cancelling) rejects concurrent acquire immediately.
	assert.False(t, svc.acquireSessionLock(context.Background(), sid))

	// Once marked cancelling, a subsequent acquire waits until unlockSession.
	svc.markSessionCancelling(sid)

	acquired := make(chan bool, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		acquired <- svc.acquireSessionLock(ctx, sid)
	}()

	select {
	case <-acquired:
		t.Fatal("expected acquireSessionLock to wait until unlockSession")
	case <-time.After(30 * time.Millisecond):
	}

	svc.unlockSession(sid)

	select {
	case ok := <-acquired:
		assert.True(t, ok)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for acquireSessionLock after unlockSession")
	}

	svc.unlockSession(sid)
}
