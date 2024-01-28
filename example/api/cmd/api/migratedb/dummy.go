//go:build !postgres && !migrate
// +build !postgres,!migrate

package migratedb

// Migrate dummy action
func Migrate(connet string, dataSources []string) error {
	// Do nothing...
	return nil
}
