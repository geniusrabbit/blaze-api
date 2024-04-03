package database

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/pkg/context/database"
)

type openFnk func(dsn string) gorm.Dialector

var dialectors = map[string]openFnk{}

// ConnectMasterSlave databases
func ConnectMasterSlave(ctx context.Context, master, slave string, debug bool) (*gorm.DB, *gorm.DB, error) {
	mdb, err := Connect(ctx, master, debug)
	if err != nil {
		return nil, nil, fmt.Errorf("master: %s", err.Error())
	}
	sdb, err := Connect(ctx, slave, debug)
	if err != nil {
		return nil, nil, fmt.Errorf("slave: %s", err.Error())
	}
	return mdb, sdb, nil
}

// Connect to database
func Connect(ctx context.Context, connection string, debug bool) (*gorm.DB, error) {
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
	db, err := gorm.Open(openDriver(connection), &gorm.Config{SkipDefaultTransaction: true})
	if err == nil && debug {
		db = db.Debug()
	}
	return db, err
}

// WithDatabase puts databases to context
func WithDatabase(ctx context.Context, master, slave *gorm.DB) context.Context {
	return database.WithDatabase(ctx, master, slave)
}
