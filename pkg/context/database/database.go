package database

import (
	"context"

	"go.uber.org/multierr"
	"gorm.io/gorm"
)

// List of context keys
var (
	CtxDatabase    = struct{ s string }{"db"}
	CtxTransaction = struct{ s string }{"db:transaction"}
)

type dbcontext struct {
	master   *gorm.DB
	readonly *gorm.DB
}

// Readonly database conncetion
func Readonly(ctx context.Context) *gorm.DB {
	dbctx := dbContext(ctx)
	if dbctx == nil {
		return nil
	}
	if dbctx.readonly != nil {
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
	tx := db.Begin()
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
