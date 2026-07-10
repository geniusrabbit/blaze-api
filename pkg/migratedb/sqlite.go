//go:build migrate && (sqlite || sqlite3 || alldb)

package migratedb

import (
	"database/sql"

	mdatabase "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
)

func init() {
	registerMigrateDriver(migrateSQLiteDriver, "sqlite", "sqlite3")
}

func migrateSQLiteDriver(conn *sql.DB, migrateTable string) (mdatabase.Driver, error) {
	return sqlite3.WithInstance(conn, &sqlite3.Config{
		MigrationsTable: migrateTable,
	})
}
