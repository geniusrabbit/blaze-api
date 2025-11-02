//go:build mssql || alldb
// +build mssql alldb

package database

import (
	// _ "gorm.io/gorm/dialects/mssql"
	"context"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func init() {
	registerDialector(&mssqlDialector{}, "mssql", "sqlserver")
}

type mssqlDialector struct{ defaultDialector }

func (d *mssqlDialector) Dialector(ctx context.Context, dsn string, config *gorm.Config) (gorm.Dialector, error) {
	return sqlserver.Open(dsn), nil
}
