//go:build clickhouse || alldb
// +build clickhouse alldb

package database

import (
	"context"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func init() {
	registerDialector(&clickhouseDialector{}, "ch", "clickhouse")
}

type clickhouseDialector struct{ defaultDialector }

func (d *clickhouseDialector) Dialector(ctx context.Context, dns string, config *gorm.Config) (gorm.Dialector, error) {
	return clickhouse.Open(dns), nil
}
