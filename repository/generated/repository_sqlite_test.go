package generated

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/geniusrabbit/blaze-api/pkg/context/database"
	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
)

type genItem struct {
	ID        uint64 `gorm:"primaryKey"`
	Title     string
	Count     int
	Approve   pkgModels.ApproveStatus `gorm:"column:approve_status"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m genItem) GetID() uint64             { return m.ID }
func (m *genItem) SetID(id uint64)          { m.ID = id }
func (m *genItem) SetUpdatedAt(t time.Time) { m.UpdatedAt = t }

func genCtx(t *testing.T) context.Context {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())),
		&gorm.Config{Logger: logger.Discard, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&genItem{}); err != nil {
		t.Fatal(err)
	}
	return database.WithDatabase(context.Background(), db, nil)
}

func TestRepositoryUpdateAndDeleteNotFound(t *testing.T) {
	t.Parallel()

	ctx := genCtx(t)
	repo := NewRepository[genItem, uint64]()

	// Before the fix a missing ID was a silent success
	if err := repo.Update(ctx, 404, &genItem{Title: "x"}); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("update missing: err = %v", err)
	}
	if err := repo.Delete(ctx, 404); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("delete missing: err = %v", err)
	}

	id, err := repo.Create(ctx, &genItem{Title: "a", Count: 1})
	if err != nil {
		t.Fatal(err)
	}
	// Writing the same values again is not "not found"
	if err := repo.Update(ctx, id, &genItem{Title: "a", Count: 1}); err != nil {
		t.Fatalf("update unchanged: %v", err)
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestRepositoryUpdateFieldsSetsZeroValues(t *testing.T) {
	t.Parallel()

	ctx := genCtx(t)
	repo := NewRepository[genItem, uint64]()
	id, _ := repo.Create(ctx, &genItem{Title: "a", Count: 5})

	// Update skips zero values, as before
	_ = repo.Update(ctx, id, &genItem{Title: "b", Count: 0})
	got, _ := repo.Get(ctx, id)
	if got.Title != "b" || got.Count != 5 {
		t.Fatalf("after Update: %+v", got)
	}

	// UpdateFields writes the listed fields, zero included
	if err := repo.UpdateFields(ctx, id, &genItem{Count: 0}, []string{"count"}); err != nil {
		t.Fatal(err)
	}
	got, _ = repo.Get(ctx, id)
	if got.Title != "b" || got.Count != 0 {
		t.Fatalf("after UpdateFields: %+v", got)
	}
}

func TestRepositoryApproverImplementsInterface(t *testing.T) {
	t.Parallel()

	ctx := genCtx(t)
	repo := NewRepository[genItem, uint64]()
	id, _ := repo.Create(ctx, &genItem{Title: "a"})

	// Before the fix the signature didn't match RepositoryApproveIface
	var approver RepositoryApproveIface[uint64] = &RepositoryApprover[genItem, uint64]{IDName: "id", StatusName: "approve_status"}
	if err := approver.Approve(ctx, id); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(ctx, id); got.Approve != pkgModels.ApprovedApproveStatus {
		t.Fatalf("status = %v", got.Approve)
	}
	if err := approver.Reject(ctx, 404); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("reject missing: err = %v", err)
	}
}
