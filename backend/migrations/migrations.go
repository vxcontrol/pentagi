package migrations

import "embed"

//go:embed sql/*.sql
var EmbedMigrations embed.FS
