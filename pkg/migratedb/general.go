//go:build migrate

package migratedb

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/demdxx/gocast/v2"
	"github.com/geniusrabbit/blaze-api/pkg/database"
	"github.com/golang-migrate/migrate"
	mdatabase "github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/mysql"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
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
	db, err := database.Connect(context.Background(), connet, false)
	if err != nil {
		return err
	}
	if conn, _ := db.DB(); conn != nil {
		defer conn.Close()
	}
	conn, _ := db.DB()

	// Process all data sources
	for _, source := range dataSources {
		if len(source.Models) > 0 {
			// Auto migrate models
			if err := db.AutoMigrate(source.Models...); err != nil {
				return err
			}
		}
		if len(source.URI) == 0 {
			continue
		}

		// Define schema_migrations table name
		migrateTable := gocast.Or(source.SchemaMigrationsTable, "schema_migrations")

		// Print log
		fmt.Println("Migrate source:", source.URI, "to", migrateTable)

		// Reset dirty migrations state to be able to run migrations again
		_ = db.Exec("update " + migrateTable + " set version=version-1, dirty=false where dirty=true;")

		// Process migrations sources and apply them to the database
		for _, uri := range source.URI {
			var (
				driver mdatabase.Driver
				err    error
			)
			// Configure the database driver
			switch connURL.Scheme {
			case "postgres", "postgresql", "pg":
				driver, err = postgres.WithInstance(conn, &postgres.Config{MigrationsTable: migrateTable})
			case "sqlite3", "sqlite":
				driver, err = sqlite3.WithInstance(conn, &sqlite3.Config{MigrationsTable: migrateTable})
			case "mysql":
				driver, err = mysql.WithInstance(conn, &mysql.Config{MigrationsTable: migrateTable})
			default:
				err = fmt.Errorf("unsupported database driver: %s", connURL.Scheme)
			}
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
