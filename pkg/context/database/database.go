package database

import (
	"context"
	"errors"

	"go.uber.org/multierr"
	"gorm.io/gorm"
)

// List of context keys
var (
	CtxDatabase    = struct{ s string }{"db"}
	CtxTransaction = struct{ s string }{"db:transaction"}
	ctxPrimaryRead = struct{ s string }{"db:primary-read"}
)

// WithPrimaryRead makes Readonly return the master connection, for reads
// that must see the latest data, e.g. reading an object before updating it.
// Replica lag could otherwise return stale data that is then written back.
func WithPrimaryRead(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxPrimaryRead, true)
}

// IsPrimaryRead reports whether reads in the context must go to the master.
func IsPrimaryRead(ctx context.Context) bool {
	v, _ := ctx.Value(ctxPrimaryRead).(bool)
	return v
}

var errNoDatabase = errors.New("database: no connection in the context")

type dbcontext struct {
	master   *gorm.DB
	readonly *gorm.DB
}

// Readonly database connection.
// Inside a transaction (ContextTransaction) it returns the transaction, so
// reads see the transaction's own writes; with WithPrimaryRead it returns
// the master; otherwise the replica when it is configured.
func Readonly(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(CtxTransaction).(*gorm.DB); ok && tx != nil {
		return tx
	}
	dbctx := dbContext(ctx)
	if dbctx == nil {
		return nil
	}
	if dbctx.readonly != nil && !IsPrimaryRead(ctx) {
		return dbctx.readonly
	}
	return dbctx.master
}

// Master database conncetion
func Master(ctx context.Context) *gorm.DB {
	dbctx := dbContext(ctx)
	if dbctx == nil {
		return nil
	}
	return dbctx.master
}

func dbContext(ctx context.Context) *dbcontext {
	dbctxVal := ctx.Value(CtxDatabase)
	if dbctxVal == nil {
		return nil
	}
	return dbctxVal.(*dbcontext)
}

// ContextTransaction get or init new transaction and put object to the context
func ContextTransaction(ctx context.Context, dbs ...*gorm.DB) (*gorm.DB, context.Context, bool, error) {
	var (
		db    *gorm.DB
		curTx = ctx.Value(CtxTransaction)
	)
	if curTx != nil {
		if ttx := curTx.(*gorm.DB); ttx != nil {
			return ttx, ctx, false, nil
		}
	}
	if len(dbs) > 0 && dbs[0] != nil {
		db = dbs[0]
	} else {
		db = Master(ctx)
	}
	if db == nil {
		return nil, ctx, false, errNoDatabase
	}
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, ctx, false, tx.Error
	}
	return tx, context.WithValue(ctx, CtxTransaction, tx), true, nil
}

// ContextTransactionExec executes function in transaction
func ContextTransactionExec(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error, dbs ...*gorm.DB) error {
	tx, ctx, isNew, err := ContextTransaction(ctx, dbs...)
	if err != nil {
		return err
	}
	if !isNew {
		return fn(ctx, tx)
	}
	defer func() {
		if recErr := recover(); recErr != nil {
			if err := tx.Rollback().Error; err != nil {
				panic(err)
			}
			panic(recErr)
		}
	}()
	if err = fn(ctx, tx); err != nil {
		return multierr.Append(err, tx.Rollback().Error)
	}
	return tx.Commit().Error
}

// ContextExecutor returns SQL executor from opened transaction or master connection
func ContextExecutor(ctx context.Context, dbs ...*gorm.DB) *gorm.DB {
	curTx := ctx.Value(CtxTransaction)
	if curTx != nil {
		switch ttx := curTx.(type) {
		case *gorm.DB:
			return ttx
		}
	}
	if len(dbs) > 0 && dbs[0] != nil {
		return dbs[0]
	}
	return Master(ctx)
}

// WithDatabase puts databases to context
func WithDatabase(ctx context.Context, master, slave *gorm.DB) context.Context {
	return context.WithValue(ctx, CtxDatabase, &dbcontext{master: master, readonly: slave})
}
