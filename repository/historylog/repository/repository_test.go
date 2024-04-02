package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

type testSuite struct {
	testsuite.DatabaseSuite

	testRepo historylog.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.testRepo = New()
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	count, err := s.testRepo.Count(s.Ctx, &historylog.Filter{
		ID:          []uuid.UUID{uuid.New(), uuid.New()},
		Name:        []string{"name1", "name2"},
		UserID:      []uint64{1, 2},
		AccountID:   []uint64{1, 2},
		ObjectID:    []uint64{1, 2},
		ObjectIDStr: []string{"1", "2"},
		ObjectType:  []string{"type1", "type2"},
	})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestFetchList() {
	uid1, uid2 := uuid.New(), uuid.New()
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(uid1, uid2, 100).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "user_id", "account_id", "object_id", "object_id_str", "object_type", "created_at"}).
				AddRow(uid1, 1, 1, 111, "111", "model:User", time.Now()).
				AddRow(uid2, 1, 1, 112, "112", "model:User", time.Now()),
		)
	objs, err := s.testRepo.FetchList(s.Ctx,
		&historylog.Filter{ID: []uuid.UUID{uid1, uid2}},
		&historylog.Order{ID: 1, Name: -1},
		&repository.Pagination{Size: 100})
	s.NoError(err)
	s.Equal(2, len(objs))
}

func TestSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
