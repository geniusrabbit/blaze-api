package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/demdxx/gocast/v2"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/database"
)

// ConnectMasterSlave databases
func ConnectMasterSlave(ctx context.Context, master, slave string) (*gorm.DB, *gorm.DB, error) {
	mdb, err := Connect(ctx, master)
	if err != nil {
		return nil, nil, fmt.Errorf("master: %s", err.Error())
	}
	if slave == "" {
		return mdb, mdb, nil
	}
	sdb, err := Connect(ctx, slave)
	if err != nil {
		return nil, nil, fmt.Errorf("slave: %s", err.Error())
	}
	return mdb, sdb, nil
}

// Connect to database
func Connect(ctx context.Context, connection string) (*gorm.DB, error) {
	var (
		i      = strings.Index(connection, "://")
		driver = connection[:i]
	)

	if driver == "mysql" {
		connection = connection[i+3:]
	}

	openDriver := dialectors[driver]
	if openDriver == nil {
		return nil, fmt.Errorf(`unsupported database driver %s`, driver)
	}

	var (
		config         = &gorm.Config{SkipDefaultTransaction: true}
		dialector, err = openDriver.Dialector(ctx, connection, config)
	)

	if err != nil {
		return nil, err
	}

	// Open gorm DB
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	// Set debug mode
	if openDriver.IsDebug(ctx, connection) {
		db = db.Debug()
	}

	// Prepare DB
	db, err = openDriver.PrepareDB(ctx, db)
	if err != nil {
		return nil, err
	}

	// Apply connection pool limits, if requested via the DSN query
	// parameters (see applyPoolOptions doc).
	if err := applyPoolOptions(connection, db); err != nil {
		return nil, err
	}

	return db, nil
}

// applyPoolOptions configures the underlying database/sql connection pool
// from optional DSN query parameters, so every dialector backed by
// database/sql (postgres, mysql, mssql, sqlite, clickhouse — not ydb, which
// manages pooling via its own driver options, see ydb.go) can be tuned the
// same way any consumer already tunes e.g. `sslmode` or `debug`, without
// consumers needing their own Go code to call db.DB().SetMaxOpenConns(...).
// Recognized parameters (all optional, unrecognized/invalid values are
// ignored and leave the database/sql default in place):
//   - max_open_conns: maximum number of open connections (int)
//   - max_idle_conns: maximum number of idle connections kept in the pool (int)
//   - conn_max_lifetime: maximum amount of time a connection may be reused, e.g. "30m" (duration)
//   - conn_max_idle_time: maximum amount of time a connection may sit idle before being closed, e.g. "5m" (duration)
//
// Left unset, database/sql defaults apply (MaxOpenConns unlimited,
// MaxIdleConns 2, no lifetime/idle-time limit) — that is what every
// consumer got before this option existed, so this is purely opt-in.
func applyPoolOptions(dsn string, db *gorm.DB) error {
	parsedDSN, err := url.Parse(dsn)
	if err != nil || len(parsedDSN.Query()) == 0 {
		// Non-URL DSNs (e.g. some ODBC-style strings) simply can't carry
		// pool params this way — nothing to apply.
		return nil
	}

	query := parsedDSN.Query()
	maxOpenConns := query.Get("max_open_conns")
	maxIdleConns := query.Get("max_idle_conns")
	connMaxLifetime := query.Get("conn_max_lifetime")
	connMaxIdleTime := query.Get("conn_max_idle_time")
	if maxOpenConns == "" && maxIdleConns == "" && connMaxLifetime == "" && connMaxIdleTime == "" {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if maxOpenConns != "" {
		sqlDB.SetMaxOpenConns(gocast.Int(maxOpenConns))
	}
	if maxIdleConns != "" {
		sqlDB.SetMaxIdleConns(gocast.Int(maxIdleConns))
	}
	if connMaxLifetime != "" {
		if d, err := time.ParseDuration(connMaxLifetime); err == nil {
			sqlDB.SetConnMaxLifetime(d)
		}
	}
	if connMaxIdleTime != "" {
		if d, err := time.ParseDuration(connMaxIdleTime); err == nil {
			sqlDB.SetConnMaxIdleTime(d)
		}
	}
	return nil
}

// WithDatabase puts databases to context
func WithDatabase(ctx context.Context, master, slave *gorm.DB) context.Context {
	return database.WithDatabase(ctx, master, slave)
}
