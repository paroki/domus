package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:build/client
var dist embed.FS

// DistFS returns the embedded SPA client build rooted at build/client.
func DistFS() (fs.FS, error) {
	return fs.Sub(dist, "build/client")
}
