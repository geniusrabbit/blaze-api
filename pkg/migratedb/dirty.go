package migratedb

import (
	"fmt"

	mdatabase "github.com/golang-migrate/migrate/v4/database"
)

// resolveDirty fails on a dirty migration version, or rolls it back by one
// when the source allows it (ResetDirty).
//
// A dirty version means a migration failed half-way; migrations are not
// transactional, so re-running it automatically can corrupt the schema.
func resolveDirty(driver mdatabase.Driver, source MigrateSource, migrateTable string) error {
	v, dirty, err := driver.Version()
	if err != nil || !dirty {
		return nil
	}
	if !source.ResetDirty {
		return fmt.Errorf("%w: %s version %d (fix the schema, then run `migrate force %d`)",
			ErrDirtyDatabase, migrateTable, v, v)
	}
	// Reset via the driver API (Postgres UPDATE / ClickHouse append INSERT).
	// Do not run dialect-specific SQL here — CH TinyLog version tables reject UPDATE.
	if err := driver.SetVersion(v-1, false); err != nil {
		return fmt.Errorf("reset dirty migration version %d: %w", v, err)
	}
	return nil
}
