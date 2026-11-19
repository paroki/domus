package testutil

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/paroki/domus/ent"
	"github.com/paroki/domus/ent/enttest"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var entDBCounter atomic.Uint64

// NewEntClient creates an isolated in-memory SQLite Ent client with migrated schema.
func NewEntClient(t testing.TB) *ent.Client {
	t.Helper()
	id := entDBCounter.Add(1)
	dsn := fmt.Sprintf("file:ent_%d?mode=memory&cache=shared&_pragma=foreign_keys(1)", id)
	db, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "open in-memory sqlite")

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(ent.Driver(drv)))
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}
