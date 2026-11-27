package repository

import (
	"errors"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type pageItem struct {
	ID   uint64
	Name string
}

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DryRun: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Use(extraClausePlugin.New()); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestNormalizePageSize(t *testing.T) {
	t.Parallel()

	for size, want := range map[int]int{-5: DefaultPageSize, 0: DefaultPageSize, 1: 1, 50: 50, MaxPageSize + 1: MaxPageSize, 1 << 30: MaxPageSize} {
		if got := NormalizePageSize(size); got != want {
			t.Errorf("NormalizePageSize(%d) = %d, want %d", size, got, want)
		}
	}
}

func TestPaginationCapsSize(t *testing.T) {
	t.Parallel()

	stmt := (&Pagination{Size: 1 << 30, Page: 2}).PrepareQuery(dryRunDB(t).Model(&pageItem{})).
		Find(&[]pageItem{}).Statement
	sql := stmt.SQL.String()
	if !strings.Contains(sql, "LIMIT $1 OFFSET $2") {
		t.Fatalf("sql = %s", sql)
	}
	if stmt.Vars[0] != MaxPageSize || stmt.Vars[1] != MaxPageSize {
		t.Fatalf("vars = %v, want limit and offset of %d", stmt.Vars, MaxPageSize)
	}
}

func TestPrepareAfterQueryBindsCursor(t *testing.T) {
	t.Parallel()

	const injection = `1' OR '1'='1`
	db := dryRunDB(t)
	stmt := (&Pagination{After: injection}).
		PrepareAfterQuery(db.Model(&pageItem{}), "id", []OrderingColumn{{Name: "name"}}).
		Find(&[]pageItem{}).Statement
	sql := stmt.SQL.String()

	// Before the fix the cursor was concatenated into the SQL text
	if strings.Contains(sql, injection) || !strings.Contains(sql, "WHERE id = $1") {
		t.Fatalf("cursor is not a bound parameter: %s", sql)
	}
	found := false
	for _, v := range stmt.Vars {
		found = found || v == injection
	}
	if !found {
		t.Fatalf("vars = %v, want the cursor", stmt.Vars)
	}
}

func TestPrepareAfterQueryRejectsBadColumns(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		idCol string
		order string
	}{
		{"id; DROP TABLE x", "name"},
		{"id", "name DESC, (SELECT 1)"},
	} {
		q := (&Pagination{After: "5"}).PrepareAfterQuery(dryRunDB(t).Model(&pageItem{}), tc.idCol, []OrderingColumn{{Name: tc.order}})
		if !errors.Is(q.Error, errInvalidPaginationColumn) {
			t.Errorf("idCol %q, order %q: error = %v", tc.idCol, tc.order, q.Error)
		}
	}
}
