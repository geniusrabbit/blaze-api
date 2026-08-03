//go:build migrate && (clickhouse || ch || alldb)

package migratedb

import (
	"database/sql"

	mdatabase "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/clickhouse"
)

func init() {
	registerMigrateDriver(migrateClickhouseDriver, "clickhouse", "ch")
}

func migrateClickhouseDriver(conn *sql.DB, migrateTable string) (mdatabase.Driver, error) {
	return clickhouse.WithInstance(conn, &clickhouse.Config{
		MigrationsTable: migrateTable,
	})
}
