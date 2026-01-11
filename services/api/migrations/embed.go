package migrations

import "embed"

// PostgresFS contains all PostgreSQL migration files embedded at compile time.
// Use this with golang-migrate's iofs source.
//
//go:embed postgres/*.sql
var PostgresFS embed.FS
