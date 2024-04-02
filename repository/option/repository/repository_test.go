package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/geniusrabbit/gosql/v2"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

var testOption = model.Option{
	Name:     "opt.name",
	Type:     model.UserOptionType,
	TargetID: 1,
	Value:    *gosql.MustNullableJSON[any](map[string]any{"val": 1}),
}

type testSuite struct {
	testsuite.DatabaseSuite

	testRepo option.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.testRepo = New()
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("opt.name", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}).
				AddRow(model.UserOptionType, uint64(1), "opt.name", `{"val":1}`, time.Now()),
		)
	role, err := s.testRepo.Get(s.Ctx, "opt.name", model.UserOptionType, 1)
	s.NoError(err)
	s.Equal("opt.name", role.Name)
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("opt.name1", "opt.name2", 100).
		WillReturnRows(
			sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}).
				AddRow(model.UserOptionType, uint64(1), "opt.name1", `{"val":1}`, time.Now()).
				AddRow(model.UserOptionType, uint64(2), "opt.name2", `{"val":2}`, time.Now()),
		)
	list, err := s.testRepo.FetchList(s.Ctx,
		&option.Filter{Name: []string{"opt.name1", "opt.name2"}},
		&option.ListOrder{Name: model.OrderAsc},
		&repository.Pagination{Size: 100})
	s.NoError(err)
	s.Equal(2, len(list))
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WithArgs("opt.name1", "opt.name2").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(2),
		)
	count, err := s.testRepo.Count(s.Ctx,
		&option.Filter{Name: []string{"opt.name1", "opt.name2"}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestSet() {
	s.Mock.ExpectExec("INSERT INTO").
		WithArgs(model.UserOptionType, uint64(1), "opt.name", `{"val":1}`, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := s.testRepo.Set(s.Ctx, &testOption)
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), model.UserOptionType, uint64(1), "opt.name").
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := s.testRepo.Delete(s.Ctx, "opt.name", model.UserOptionType, 1)
	s.NoError(err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
