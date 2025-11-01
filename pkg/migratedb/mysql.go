//go:build migrate && (mysql || mariadb || alldb)

package migratedb

import (
	"database/sql"

	mdatabase "github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/mysql"
)

func init() {
	registerMigrateDriver(migrateMySQLDriver, "mysql", "mariadb", "my")
}

func migrateMySQLDriver(conn *sql.DB, migrateTable string) (mdatabase.Driver, error) {
	return mysql.WithInstance(conn, &mysql.Config{
		MigrationsTable: migrateTable,
	})
}
