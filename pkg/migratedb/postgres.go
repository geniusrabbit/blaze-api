//go:build migrate && (postgres || pgsql || pg || alldb)

package migratedb

import (
	"database/sql"

	mdatabase "github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/lib/pq"
)

func init() {
	registerMigrateDriver(migratePostgresDriver, "postgresql", "postgres", "pgsql", "pg")
}

func migratePostgresDriver(conn *sql.DB, migrateTable string) (mdatabase.Driver, error) {
	return postgres.WithInstance(conn, &postgres.Config{
		MigrationsTable: migrateTable,
	})
}
