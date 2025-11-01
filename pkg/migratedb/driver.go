//go:build migrate

package migratedb

import (
	"database/sql"
	"fmt"

	mdatabase "github.com/golang-migrate/migrate/database"
)

type migrateDriverFunc func(conn *sql.DB, migrateTable string) (mdatabase.Driver, error)

var migrateDrivers = map[string]migrateDriverFunc{}

func migrateDriver(schema string, conn *sql.DB, migrateTable string) (mdatabase.Driver, error) {
	if driver, ok := migrateDrivers[schema]; ok {
		return driver(conn, migrateTable)
	}
	return nil, fmt.Errorf("unsupported database driver: %s", schema)
}

func registerMigrateDriver(driver migrateDriverFunc, schema ...string) {
	for _, s := range schema {
		migrateDrivers[s] = driver
	}
}
