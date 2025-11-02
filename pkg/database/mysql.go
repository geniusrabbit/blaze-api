//go:build mysql || alldb
// +build mysql alldb

package database

import (
	// _ "gorm.io/gorm/dialects/mysql"
	"context"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	registerDialector(&mysqlDialector{}, "mysql", "mariadb")
}

type mysqlDialector struct{ defaultDialector }

func (d *mysqlDialector) Dialector(ctx context.Context, dsn string, config *gorm.Config) (gorm.Dialector, error) {
	return mysql.Open(dsn), nil
}
