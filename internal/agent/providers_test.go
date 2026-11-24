package agent_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paroki/domus/ent/agentprovider"
	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/agent/domain"
	"github.com/paroki/domus/internal/agent/driver"
	"github.com/paroki/domus/internal/platform/crypt"
	"github.com/paroki/domus/internal/platform/errx"
	"github.com/paroki/domus/internal/platform/setting"
	"github.com/paroki/domus/testutil"
	"github.com/stretchr/testify/require"
)

func TestDeleteProvider_InUseGuard(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	store := setting.NewStore(db)
	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	feat := agent.New(
		nil,
		slog.New(slog.DiscardHandler),
		agent.WithDB(db),
		agent.WithCipher(cipher),
		agent.WithSettings(store),
	)
	feat.RegisterSettings(store)
	require.NoError(t, store.Load(ctx))

	prov, err := db.AgentProvider.Create().
		SetName("OpenAI").
		SetProvider(agentprovider.ProviderOpenai).
		SetAPIKey("sk-plain").
		Save(ctx)
	require.NoError(t, err)

	// Bind provider to agent.model.mikael
	_, err = store.Update(ctx, map[string]string{
		agent.SettingModelMikael: prov.ID.String() + ":gpt-4o",
	})
	require.NoError(t, err)

	err = feat.DeleteProvider(ctx, prov.ID)
	require.ErrorIs(t, err, domain.ErrProviderInUse)

	// Unbind -> DeleteProvider succeeds
	_, err = store.Update(ctx, map[string]string{
		agent.SettingModelMikael: "",
	})
	require.NoError(t, err)

	require.NoError(t, feat.DeleteProvider(ctx, prov.ID))
	require.ErrorIs(t, feat.DeleteProvider(ctx, uuid.New()), domain.ErrProviderNotFound)
}

func TestSyncProviderModels_Preserves4xxSentinels(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewEntClient(t)
	cipher, err := crypt.New("0123456789abcdef0123456789abcdef")
	require.NoError(t, err)

	statusCode := http.StatusUnauthorized
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte("upstream error"))
	}))
	t.Cleanup(srv.Close)

	feat := agent.New(
		driver.NewRegistry(srv.Client()),
		slog.New(slog.DiscardHandler),
		agent.WithDB(db),
		agent.WithCipher(cipher),
	)

	prov, err := db.AgentProvider.Create().
		SetName("OpenAI").
		SetProvider(agentprovider.ProviderOpenai).
		SetAPIKey("sk-plain").
		SetBaseURL(srv.URL).
		Save(ctx)
	require.NoError(t, err)

	_, err = feat.SyncProviderModels(ctx, prov.ID)
	require.ErrorIs(t, err, domain.ErrProviderAuthFailed)

	statusCode = http.StatusForbidden
	_, err = feat.SyncProviderModels(ctx, prov.ID)
	require.ErrorIs(t, err, domain.ErrProviderAuthFailed)

	statusCode = http.StatusTooManyRequests
	_, err = feat.SyncProviderModels(ctx, prov.ID)
	require.ErrorIs(t, err, domain.ErrProviderRateLimited)

	var de *errx.Error
	require.True(t, errors.As(err, &de))
	require.Equal(t, errx.KindRateLimited, de.Kind)
}
