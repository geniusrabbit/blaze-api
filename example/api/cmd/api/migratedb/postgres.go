//go:build postgres && migrate
// +build postgres,migrate

package migratedb

import (
	"database/sql"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

// Migrate database schema from many source directories and one target database
func Migrate(connet string, dataSources []string) error {
	connURL, err := url.Parse(connet)
	if err != nil {
		return err
	}
	db, err := sql.Open(connURL.Scheme, connet)
	if err != nil {
		return err
	}
	for _, source := range dataSources {
		sourceURL, err := url.Parse(source)
		if err != nil {
			return err
		}
		migrateTable := "schema_migrations"
		if dir := filepath.Base(sourceURL.Path); dir != "migrations" {
			migrateTable += "_" + strings.ReplaceAll(dir, "-", "_")
		}
		_, _ = db.Exec("update " + migrateTable + " set version=version-1, dirty=false where dirty=true;")
		driver, err := postgres.WithInstance(db, &postgres.Config{MigrationsTable: migrateTable})
		if err != nil {
			return err
		}
		m, err := migrate.NewWithDatabaseInstance(
			source, connURL.Path[1:], driver)
		if err != nil {
			return err
		}
		if err = m.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	}
	return nil
}
