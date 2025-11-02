//go:build ydb || yadb || alldb
// +build ydb yadb alldb

package database

import (
	"context"
	"net/url"
	"time"

	"github.com/demdxx/gocast/v2"
	ydb "github.com/ydb-platform/gorm-driver"
	"gorm.io/gorm"
)

func init() {
	registerDialector(&ydbDialector{}, "ydb", "ydbs", "yadb", "yadbs")
}

type ydbDialector struct{ defaultDialector }

func (d *ydbDialector) PrepareDB(ctx context.Context, db *gorm.DB) (*gorm.DB, error) {
	return db.Set("gorm:ignore_alter_column", true), nil
}

// openYDB opens YDB database connection
// Replace schema ydb:// or yadb:// on grpc:// and ydbs:// or yadbs:// on grpcs://
func (d *ydbDialector) Dialector(ctx context.Context, dsn string, config *gorm.Config) (gorm.Dialector, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "ydb", "yadb":
		u.Scheme = "grpc"
	case "ydbs", "yadbs":
		u.Scheme = "grpcs"
	}

	// Parse options from query parameters
	var (
		query = u.Query()
		opts  []ydb.Option
	)

	// If table_prefix is set in query parameters, use it
	if prefix := query.Get("table_prefix"); prefix != "" {
		opts = append(opts, ydb.WithTablePathPrefix(prefix))
		query.Del("table_prefix")
	}

	// If max open connections are set in query parameters, use them
	if maxOpenConns := query.Get("max_open_conns"); maxOpenConns != "" {
		opts = append(opts, ydb.WithMaxOpenConns(gocast.Int(maxOpenConns)))
		query.Del("max_open_conns")
	}

	// If max idle connections are set in query parameters, use them
	if maxIdleConns := query.Get("max_idle_conns"); maxIdleConns != "" {
		opts = append(opts, ydb.WithMaxIdleConns(gocast.Int(maxIdleConns)))
		query.Del("max_idle_conns")
	}

	// If connection max idle time is set in query parameters, use it
	if connMaxIdleTime := query.Get("conn_max_idle_time"); connMaxIdleTime != "" {
		duration, _ := time.ParseDuration(connMaxIdleTime)
		if duration > 0 {
			opts = append(opts, ydb.WithConnMaxIdleTime(duration))
		}
		query.Del("conn_max_idle_time")
	}

	query.Del("debug")
	u.RawQuery = query.Encode()

	// Disable automatic ping to YDB as it is not required
	config.DisableAutomaticPing = true

	// Open YDB connection
	return ydb.Open(u.String(), opts...), nil
}
