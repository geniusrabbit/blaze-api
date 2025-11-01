//go:build !migrate

package migratedb

import "context"

// Migrate dummy action
func Migrate(ctx context.Context, connet string, dataSources []MigrateSource) error {
	// Do nothing...
	return nil
}
