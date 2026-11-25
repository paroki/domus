package migrations

import "embed"

// FS embeds versioned SQL migration files and atlas.sum when present.
// During pre-release development only embed.go exists, so fs.Glob(FS, "*.sql") returns empty.
//
//go:embed *
var FS embed.FS
