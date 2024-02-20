//go:build postgres && migrate
// +build postgres,migrate

package migratedb

import (
	"database/sql"
	"errors"
	"net/url"

	"github.com/demdxx/gocast/v2"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

// Migrate database schema from many source directories and one target database
func Migrate(connet string, dataSources []MigrateSource) error {
	// Parse connection string
	connURL, err := url.Parse(connet)
	if err != nil {
		return err
	}

	// Open database connection
	db, err := sql.Open(connURL.Scheme, connet)
	if err != nil {
		return err
	}
	defer db.Close()

	// Process all data sources
	for _, source := range dataSources {
		// Define schema_migrations table name
		migrateTable := gocast.Or(source.SchemaMigrationsTable, "schema_migrations")

		// Reset dirty migrations state to be able to run migrations again
		_, _ = db.Exec("update " + migrateTable + " set version=version-1, dirty=false where dirty=true;")

		// Process migrations sources and apply them to the database
		for _, uri := range source.URI {
			// Configure the database driver
			driver, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: migrateTable})
			if err != nil {
				return err
			}

			// Init migration instance
			migInst, err := migrate.NewWithDatabaseInstance(uri, connURL.Path[1:], driver)
			if err != nil {
				return err
			}

			// Apply migrations
			if err = migInst.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
		}
	}
	return nil
}
