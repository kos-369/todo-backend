// Package migrations embeds the SQL migration files, so a Go binary
// or test can apply them without reading files from disk.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
