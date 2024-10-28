package migratedb

type MigrateSource struct {
	URI                   []string
	SchemaMigrationsTable string
	Models                []any
}
