package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

type testSuite struct {
	testsuite.DatabaseSuite

	authclientRepo authclient.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.authclientRepo = New()
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("1").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "account_id", "user_id", "title", "secret", "created_at"}).
				AddRow("1", 1, 1, "title1", "secret", time.Now()),
		)
	role, err := s.authclientRepo.Get(s.Ctx, "1")
	s.NoError(err)
	s.Equal("1", role.ID)
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("1", "2", 100).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "account_id", "user_id", "title", "secret", "created_at"}).
				AddRow("1", 1, 1, "title1", "secret", time.Now()).
				AddRow("2", 1, 1, "title2", "secret", time.Now()),
		)
	clients, err := s.authclientRepo.FetchList(s.Ctx, &authclient.Filter{
		ID: []string{"1", "2"}, PageSize: 100})
	s.NoError(err)
	s.Equal(2, len(clients))
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WithArgs("1", "2").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(2),
		)
	count, err := s.authclientRepo.Count(s.Ctx, &authclient.Filter{
		ID: []string{"1", "2"}, PageSize: 100})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.Mock.ExpectExec("INSERT INTO").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	id, err := s.authclientRepo.Create(
		s.Ctx,
		&model.AuthClient{
			ID:        "101",
			Title:     "test",
			AccountID: 1,
			UserID:    1,
		})
	s.NoError(err)
	s.Equal("101", id)
}

func (s *testSuite) TestUpdate() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs("test", sqlmock.AnyArg(), "101").
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := s.authclientRepo.Update(s.Ctx, "101", &model.AuthClient{Title: "test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), "101").
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := s.authclientRepo.Delete(s.Ctx, "101")
	s.NoError(err)
}

func TestAuthclientSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
