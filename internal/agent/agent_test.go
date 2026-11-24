package agent_test

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/paroki/domus/internal/agent"
	"github.com/paroki/domus/internal/app"
	"github.com/stretchr/testify/require"
)

func TestFeature_SQLDBAndMCPServerWiring(t *testing.T) {
	sqlDB := &sql.DB{}
	feat := agent.New(nil, slog.New(slog.DiscardHandler), agent.WithSQLDB(sqlDB))
	require.Same(t, sqlDB, feat.SQLDB())
	require.Nil(t, feat.MCPServer())

	_ = app.GetFiber(app.Config{}, slog.New(slog.DiscardHandler), feat)
	require.NotNil(t, feat.MCPServer())
}
