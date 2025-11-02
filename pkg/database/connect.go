package database

import (
	"context"
	"fmt"
	"strings"

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

	return db, nil
}

// WithDatabase puts databases to context
func WithDatabase(ctx context.Context, master, slave *gorm.DB) context.Context {
	return database.WithDatabase(ctx, master, slave)
}
