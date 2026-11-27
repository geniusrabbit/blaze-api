package migratedb

import "errors"

// ErrDirtyDatabase is returned when a previous migration failed half-way.
// Fix the schema by hand, then mark the version clean with `migrate force`.
var ErrDirtyDatabase = errors.New("migratedb: database is dirty after a failed migration")

type MigrateSource struct {
	URI                   []string
	SchemaMigrationsTable string
	Models                []any
	// ResetDirty rolls a dirty version back by one and re-runs it instead of
	// failing with ErrDirtyDatabase. Migrations are not transactional, so this
	// can apply a half-done migration twice: use it only for development.
	ResetDirty bool
}
