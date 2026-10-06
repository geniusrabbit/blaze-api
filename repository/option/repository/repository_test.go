package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/geniusrabbit/gosql/v2"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/option/models"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

var testOption = models.Option{
	Name:     "opt.name",
	Type:     models.UserOptionType,
	TargetID: 1,
	Value:    *gosql.MustNullableJSON[any](map[string]any{"val": 1}),
}

type testSuite struct {
	testsuite.DatabaseSuite

	testRepo option.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.testRepo = NewOptionRepository(map[string]any{
		"opt.default": 1,
	})
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("opt.name", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}).
				AddRow(models.UserOptionType, uint64(1), "opt.name", `{"val":1}`, time.Now()),
		)
	role, err := s.testRepo.Get(s.Ctx, "opt.name", models.UserOptionType, 1)
	s.NoError(err)
	s.Equal("opt.name", role.Name)
}

func (s *testSuite) TestGetDefault() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("opt.default", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)
	role, err := s.testRepo.Get(s.Ctx, "opt.default", models.SystemOptionType, 0)

	if s.NoError(err) {
		s.Equal("opt.default", role.Name)
		s.Equal(models.SystemOptionType, role.Type)
		s.Equal(uint64(0), role.TargetID)
		if s.NotNil(role.Value.Data) {
			s.Equal(1, *role.Value.Data)
		}
	}
}

func (s *testSuite) TestGetOneOfTypeEmptyPairs() {
	got, err := s.testRepo.GetOneOfType(s.Ctx, "opt.name", nil)
	s.Nil(got)
	s.EqualError(err, "option target pairs are required")
}

func (s *testSuite) TestGetOneOfTypeSingleRow() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(models.AccountOptionType, uint64(5), models.UserOptionType, uint64(1), "opt.name").
		WillReturnRows(
			sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}).
				AddRow(models.UserOptionType, uint64(1), "opt.name", `{"val":1}`, time.Now()),
		)
	got, err := s.testRepo.GetOneOfType(s.Ctx, "opt.name", []option.TargetPair{
		{OptionType: models.AccountOptionType, TargetID: 5},
		{OptionType: models.UserOptionType, TargetID: 1},
	})
	s.NoError(err)
	s.Equal(models.UserOptionType, got.Type)
	s.Equal(uint64(1), got.TargetID)
}

func (s *testSuite) TestGetOneOfTypePairOrder() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(models.AccountOptionType, uint64(5), models.UserOptionType, uint64(1), "opt.name").
		WillReturnRows(
			sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}).
				AddRow(models.UserOptionType, uint64(1), "opt.name", `{"val":1}`, time.Now()).
				AddRow(models.AccountOptionType, uint64(5), "opt.name", `{"val":2}`, time.Now()),
		)
	got, err := s.testRepo.GetOneOfType(s.Ctx, "opt.name", []option.TargetPair{
		{OptionType: models.AccountOptionType, TargetID: 5},
		{OptionType: models.UserOptionType, TargetID: 1},
	})
	s.NoError(err)
	s.Equal(models.AccountOptionType, got.Type)
	s.Equal(uint64(5), got.TargetID)
}

func (s *testSuite) TestGetOneOfTypeEmpty() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(models.UserOptionType, uint64(1), "opt.missing").
		WillReturnRows(sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}))
	got, err := s.testRepo.GetOneOfType(s.Ctx, "opt.missing", []option.TargetPair{
		{OptionType: models.UserOptionType, TargetID: 1},
	})
	s.NoError(err)
	s.Nil(got)
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("opt.name1", "opt.name2", 100).
		WillReturnRows(
			sqlmock.NewRows([]string{"type", "target_id", "name", "value", "created_at"}).
				AddRow(models.UserOptionType, uint64(1), "opt.name1", `{"val":1}`, time.Now()).
				AddRow(models.UserOptionType, uint64(2), "opt.name2", `{"val":2}`, time.Now()),
		)
	list, err := s.testRepo.FetchList(s.Ctx,
		&option.Filter{Name: []string{"opt.name1", "opt.name2"}},
		&option.ListOrder{Name: models.OrderAsc},
		&option.Pagination{Size: 100})
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
		WithArgs(models.UserOptionType, uint64(1), "opt.name", `{"val":1}`, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := s.testRepo.Set(s.Ctx, &testOption)
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), models.UserOptionType, uint64(1), "opt.name").
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := s.testRepo.Delete(s.Ctx, "opt.name", models.UserOptionType, 1)
	s.NoError(err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
