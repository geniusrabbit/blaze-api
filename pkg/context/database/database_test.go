package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDatabase(t *testing.T) {
	ctx := context.Background()
	ctx = WithDatabase(ctx, &gorm.DB{}, &gorm.DB{})
	assert.NotNil(t, Master(ctx))
	assert.NotNil(t, Readonly(ctx))
}

func openSQLite(t *testing.T, name string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func TestReadonlyRouting(t *testing.T) {
	t.Parallel()

	master, replica := openSQLite(t, t.Name()+"_m"), openSQLite(t, t.Name()+"_r")
	ctx := WithDatabase(context.Background(), master, replica)

	if Readonly(ctx) != replica {
		t.Fatal("plain reads must use the replica")
	}
	// Reads before writes must not see replica lag
	if Readonly(WithPrimaryRead(ctx)) != master {
		t.Fatal("WithPrimaryRead must use the master")
	}

	// Before the fix reads inside a transaction went to the replica and
	// didn't see the transaction's own writes
	err := ContextTransactionExec(ctx, func(txctx context.Context, tx *gorm.DB) error {
		if Readonly(txctx) != tx {
			t.Error("reads inside a transaction must use the transaction")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestContextTransactionErrors(t *testing.T) {
	t.Parallel()

	if _, _, _, err := ContextTransaction(context.Background()); err == nil {
		t.Fatal("transaction without a database didn't fail")
	}

	// Before the fix a failed BEGIN was ignored and a broken tx was returned
	db := openSQLite(t, t.Name())
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()
	if _, _, _, err := ContextTransaction(WithDatabase(context.Background(), db, nil)); err == nil {
		t.Fatal("BEGIN on a closed connection didn't fail")
	}
}
