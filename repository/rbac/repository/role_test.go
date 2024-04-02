package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

type testSuite struct {
	testsuite.DatabaseSuite

	roleRepo rbac.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.roleRepo = New()
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "title", "description", "created_at"}).
				AddRow(1, 1, "title1", "description1", time.Now()),
		)
	role, err := s.roleRepo.Get(s.Ctx, 1)
	s.NoError(err)
	s.Equal(uint64(1), role.ID)
}

func (s *testSuite) TestGetByName() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("system:manager").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "title", "created_at"}).
				AddRow(1, "system:manager", "manager", time.Now()),
		)
	role, err := s.roleRepo.GetByName(s.Ctx, "system:manager")
	s.NoError(err)
	s.Equal(uint64(1), role.ID)
	s.Equal("manager", role.Title)
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "type", "title", "name", "created_at"}).
				AddRow(1, "role", "title1", "test1", time.Now()).
				AddRow(2, "role", "title2", "test2", time.Now()),
		)
	roles, err := s.roleRepo.FetchList(s.Ctx, &rbac.Filter{
		ID: []uint64{1, 2}}, nil, nil)
	s.NoError(err)
	s.Equal(2, len(roles))
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	count, err := s.roleRepo.Count(s.Ctx, &rbac.Filter{})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	id, err := s.roleRepo.Create(
		s.Ctx,
		&model.Role{
			ID:    101,
			Name:  "test",
			Title: "test",
		})
	s.NoError(err)
	s.Equal(uint64(101), id)
}

func (s *testSuite) TestUpdate() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs("test", sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.roleRepo.Update(s.Ctx, 101, &model.Role{Title: "test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.roleRepo.Delete(s.Ctx, 101)
	s.NoError(err)
}

func TestRoleSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
