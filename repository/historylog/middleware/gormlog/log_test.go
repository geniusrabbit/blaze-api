package gormlog

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/geniusrabbit/blaze-api/repository/historylog"
	historylogModels "github.com/geniusrabbit/blaze-api/repository/historylog/models"
)

type logItem struct {
	ID          uint64 `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	AccessToken string
	Password    string
	Internal    string `historylog:"-"`
}

func newLogDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())),
		&gorm.Config{Logger: logger.Discard, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	// One connection: the old code wrote history through another connection
	// and deadlocked inside a transaction.
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&logItem{}, &historylogModels.HistoryAction{}); err != nil {
		t.Fatal(err)
	}
	if err := Register(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func historyRows(t *testing.T, db *gorm.DB) []historylogModels.HistoryAction {
	t.Helper()
	var rows []historylogModels.HistoryAction
	if err := db.Order("action_at").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestHistoryLogCreateAndRedaction(t *testing.T) {
	t.Parallel()

	db := newLogDB(t)
	item := &logItem{Name: "a", AccessToken: "tok-123", Password: "pw-123", Internal: "x"}
	if err := db.Create(item).Error; err != nil {
		t.Fatal(err)
	}
	rows := historyRows(t, db)
	if len(rows) != 1 || rows[0].Name != "create" || rows[0].ObjectID != item.ID {
		t.Fatalf("rows = %+v", rows)
	}
	data := rows[0].DataMap()
	// Before the fix only "password"/"secret" columns were skipped: tokens were stored
	for _, col := range []string{"access_token", "password", "internal"} {
		if _, ok := data[col]; ok {
			t.Errorf("history data contains %q: %v", col, data)
		}
	}
	if data["name"] != "a" {
		t.Errorf("history data = %v, want the name", data)
	}
}

func TestHistoryLogRollbackLeavesNoHistory(t *testing.T) {
	t.Parallel()

	db := newLogDB(t)
	done := make(chan error, 1)
	go func() {
		done <- db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&logItem{Name: "rolled-back"}).Error; err != nil {
				return err
			}
			return errors.New("rollback")
		})
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("transaction deadlocked: history was written through another connection")
	}
	if rows := historyRows(t, db); len(rows) != 0 {
		t.Fatalf("history of a rolled-back change: %+v", rows)
	}

	// Committed transactions keep their history
	if err := db.Transaction(func(tx *gorm.DB) error { return tx.Create(&logItem{Name: "kept"}).Error }); err != nil {
		t.Fatal(err)
	}
	if rows := historyRows(t, db); len(rows) != 1 {
		t.Fatalf("rows = %+v, want one", rows)
	}
}

func TestHistoryLogSkipsFailedAndNoopStatements(t *testing.T) {
	t.Parallel()

	db := newLogDB(t)
	if err := db.Create(&logItem{Name: "dup"}).Error; err != nil {
		t.Fatal(err)
	}
	// Unique violation: the old update/delete hooks ran before the statement
	_ = db.Create(&logItem{Name: "dup"}).Error
	if err := db.Model(&logItem{}).Where("name = ?", "missing").Update("name", "x").Error; err != nil {
		t.Fatal(err)
	}
	if rows := historyRows(t, db); len(rows) != 1 {
		t.Fatalf("rows = %+v, want only the successful create", rows)
	}
}

func TestHistoryLogPrimaryKeyFromContext(t *testing.T) {
	t.Parallel()

	db := newLogDB(t)
	item := &logItem{Name: "del"}
	if err := db.Create(item).Error; err != nil {
		t.Fatal(err)
	}
	// Delete by condition: the value has no PK, so it comes from the context
	ctx := historylog.WithPK(context.Background(), item.ID)
	if err := db.WithContext(ctx).Delete(&logItem{}, item.ID).Error; err != nil {
		t.Fatal(err)
	}
	rows := historyRows(t, db)
	if len(rows) != 2 || rows[1].Name != "delete" || rows[1].ObjectID != item.ID {
		t.Fatalf("rows = %+v, want a delete of object %d", rows, item.ID)
	}
}
