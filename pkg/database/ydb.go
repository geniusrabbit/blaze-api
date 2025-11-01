//go:build ydb || yadb || alldb
// +build ydb yadb alldb

package database

import (
	"net/url"
	"time"

	"github.com/demdxx/gocast/v2"
	ydb "github.com/ydb-platform/gorm-driver"
	"gorm.io/gorm"
)

func init() {
	dialectors["ydb"] = openYDB
	dialectors["ydbs"] = openYDB
	dialectors["yadb"] = openYDB
	dialectors["yadbs"] = openYDB
}

// openYDB opens YDB database connection
// Replace schema ydb:// or yadb:// on grpc:// and ydbs:// or yadbs:// on grpcs://
func openYDB(dsn string) gorm.Dialector {
	u, _ := url.Parse(dsn)
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

	u.RawQuery = query.Encode()

	// Open YDB connection
	return ydb.Open(u.String(), opts...)
}
