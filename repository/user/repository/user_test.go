package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/password"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

// Password: test
// PassHash: $2a$12$mbz/OdK.Pal.AwOz13RxX.PDqkthADBr.B4UMXerY4QbQeqAiJGma
const (
	defaultPassword     = "test"
	defaultPasswordHash = "$2a$12$mbz/OdK.Pal.AwOz13RxX.PDqkthADBr.B4UMXerY4QbQeqAiJGma"
)

type testSuite struct {
	testsuite.DatabaseSuite

	coreRepo  user.Repository[*testutil.User]
	emailRepo user.EmailRepository[*testutil.User]
	passRepo  user.PasswordRepository[*testutil.User]
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	newUser := func() *testutil.User { return &testutil.User{} }
	s.coreRepo = NewRepository(newUser)
	s.emailRepo = NewEmailRepository(s.coreRepo, newUser)
	s.passRepo = NewPasswordRepository(s.coreRepo, newUser)

	password.SetSalt([]byte("1111111"), 1)
}

func (s *testSuite) SetupTest() {
	s.DatabaseSuite.SetupSuite()
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()),
		)
	userObj, err := s.coreRepo.Get(s.Ctx, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), userObj.GetID())
}

func (s *testSuite) TestGetByEmail() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("test@mail.com", 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "test@mail.com", defaultPasswordHash, time.Now()),
		)
	userObj, err := s.emailRepo.GetByEmail(s.Ctx, "test@mail.com")
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), userObj.GetID())
	s.Assert().Equal("test@mail.com", userObj.GetEmail())
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 2, 100).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()).
				AddRow(2, 1, "email2", defaultPasswordHash, time.Now()),
		)
	users, err := s.coreRepo.FetchList(s.Ctx,
		&user.ListFilter{FilterBase: user.FilterBase{ID: []uint64{1, 2}}},
		&user.ListOrder{OrderBase: user.OrderBase{ID: pkgModels.OrderAsc}},
		&repository.Pagination{Size: 100})
	s.Assert().NoError(err)
	s.Assert().Equal(2, len(users))
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WithArgs(1, 2).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(2),
		)
	count, err := s.coreRepo.Count(s.Ctx,
		&user.ListFilter{FilterBase: user.FilterBase{ID: []uint64{1, 2}}})
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)
}

func (s *testSuite) TestGetByPassword() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "email", "password", "created_at"}).
				AddRow(1, 1, "email1", defaultPasswordHash, time.Now()),
		)

	userObj, err := s.passRepo.GetByPassword(s.Ctx, 1, defaultPassword)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), userObj.GetID())
}

func (s *testSuite) TestCreateWithPassword() {
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	u := testutil.StubWithEmail("test", pkgModels.UndefinedApproveStatus)
	u.SetID(101)
	id, err := s.passRepo.CreateWithPassword(s.Ctx, u, "password")
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(101), id)
}

func (s *testSuite) TestUpdate() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(101, 1))
	u := testutil.StubWithEmail("test", pkgModels.UndefinedApproveStatus)
	u.SetID(101)
	err := s.coreRepo.Update(s.Ctx, u)
	s.Assert().NoError(err)
}

func (s *testSuite) TestDeleteByID() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.coreRepo.Delete(s.Ctx, 101)
	s.Assert().NoError(err)
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
