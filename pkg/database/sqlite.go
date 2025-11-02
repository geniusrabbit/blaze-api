//go:build sqlite || alldb
// +build sqlite alldb

package database

import (
	"context"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	registerDialector(&sqliteDialector{}, "sqlite", "sqlite3")
}

type sqliteDialector struct{ defaultDialector }

func (d *sqliteDialector) Dialector(ctx context.Context, dsn string, config *gorm.Config) (gorm.Dialector, error) {
	return sqlite.Open(dsn), nil
}
