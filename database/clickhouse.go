//go:build clickhouse || alldb
// +build clickhouse alldb

package database

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func init() {
	dialectors["clickhouse"] = openClickhouse
}

func openClickhouse(dsn string) gorm.Dialector {
	return clickhouse.Open(dsn)
}
