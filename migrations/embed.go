package migrations

import "embed"

// FS embeds all versioned SQL migration files and atlas.sum.
//
//go:embed *.sql atlas.sum
var FS embed.FS
