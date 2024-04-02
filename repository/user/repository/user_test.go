package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
	"github.com/geniusrabbit/blaze-api/repository/user"
)

// make pwgen PASSWORD=test
// go run cmd/pwgen/main.go test
// PwSalt:   1111111
// PwCost:   12
// Password: test
// PassHash: $2a$12$mbz/OdK.Pal.AwOz13RxX.PDqkthADBr.B4UMXerY4QbQeqAiJGma
const (
	defaultPassword     = "test"
	defaultPasswordHash = "$2a$12$mbz/OdK.Pal.AwOz13RxX.PDqkthADBr.B4UMXerY4QbQeqAiJGma"
)

type testSuite struct {
	testsuite.DatabaseSuite

	userRepo user.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.userRepo = New()

	user.SetSalt([]byte("1111111"), 1)
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()),
		)
	user, err := s.userRepo.Get(s.Ctx, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
}

func (s *testSuite) TestGetByEmail() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("test@mail.com", 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "test@mail.com", defaultPasswordHash, time.Now()),
		)
	user, err := s.userRepo.GetByEmail(s.Ctx, "test@mail.com")
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
	s.Assert().Equal("test@mail.com", user.Email)
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 1, 2, 100).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()).
				AddRow(2, 1, "email2", defaultPasswordHash, time.Now()),
		)
	users, err := s.userRepo.FetchList(s.Ctx,
		&user.ListFilter{AccountID: []uint64{1}, UserID: []uint64{1, 2}},
		&user.ListOrder{ID: model.OrderAsc},
		&repository.Pagination{Size: 100})
	s.Assert().NoError(err)
	s.Assert().Equal(2, len(users))
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WithArgs(1, 1, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(2),
		)
	count, err := s.userRepo.Count(s.Ctx,
		&user.ListFilter{AccountID: []uint64{1}, UserID: []uint64{1, 2}})
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)
}

func (s *testSuite) TestGetByPassword() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("email1", 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()),
		)

	user, err := s.userRepo.GetByPassword(s.Ctx, "email1", defaultPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
}

func (s *testSuite) TestGetByToken() {
	tocken := "jBQpj4CbcJRZjznk00mjpgxvLc2QwErx"
	s.Mock.ExpectQuery(`WITH auth_client AS \(`).
		WithArgs(tocken).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "user_id", "account_id", "is_admin", "created_at", "updated_at"}).
				AddRow(1, 1, 1, 1, 0, time.Now(), time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT \* FROM "`+(*model.User)(nil).TableName()+`"`).
		WithArgs(uint64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT \* FROM "`+(*model.Account)(nil).TableName()+`"`).
		WithArgs(uint64(1), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "title", "description", "created_at"}).
				AddRow(1, 1, "title1", "description1", time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT "role_id" FROM `).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}))

	user, account, err := s.userRepo.GetByToken(s.Ctx, tocken)

	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
	s.Assert().Equal(uint64(1), account.ID)
}

func (s *testSuite) TestCreate() {
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs("test", sqlmock.AnyArg(), model.UndefinedApproveStatus, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint(101)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	id, err := s.userRepo.Create(
		s.Ctx,
		&model.User{ID: 101, Email: "test"},
		"password")
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(101), id)
}

func (s *testSuite) TestUpdate() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs("test", sqlmock.AnyArg(), model.UndefinedApproveStatus, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.userRepo.Update(
		s.Ctx,
		&model.User{
			ID:    101,
			Email: "test",
		})
	s.Assert().NoError(err)
}

func (s *testSuite) TestDeleteByID() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.userRepo.Delete(s.Ctx, 101)
	s.Assert().NoError(err)
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
