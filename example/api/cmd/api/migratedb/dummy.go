//go:build !postgres && !migrate
// +build !postgres,!migrate

package migratedb

// Migrate dummy action
func Migrate(connet string, dataSources []MigrateSource) error {
	// Do nothing...
	return nil
}
