package repository_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pkgModels "github.com/geniusrabbit/blaze-api/pkg/models"
	"github.com/geniusrabbit/blaze-api/pkg/database"
	"github.com/geniusrabbit/blaze-api/repository/account"
	accountrepo "github.com/geniusrabbit/blaze-api/repository/account/repository"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type testSuite struct {
	testsuite.DatabaseSuite

	accountRepo account.SessionRepository[*testutil.User, *testAccount]
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.accountRepo = accountrepo.NewSessionRepository(
		func() *testutil.User { return new(testutil.User) },
		func() *testAccount { return new(testAccount) },
		func() *account.Member[*testutil.User, *testAccount] {
			return new(account.Member[*testutil.User, *testAccount])
		},
	)
}

func (s *testSuite) SetupTest() {
	if err := s.Mock.ExpectationsWereMet(); err != nil {
		s.T().Log(err)
	}
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	s.BaseDB = db
	s.Mock = mock
	s.DB, err = gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{SkipDefaultTransaction: true})
	s.Require().NoError(err)
	s.Ctx = database.WithDatabase(s.Ctx, s.DB, s.DB)
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "title", "description", "created_at"}).
				AddRow(1, 1, "title1", "description1", time.Now()),
		)
	accountObj, err := s.accountRepo.Get(s.Ctx, 1)
	s.NoError(err)
	s.Equal(uint64(1), accountObj.ID)
}

func (s *testSuite) TestLoadPermissions() {
	s.Mock.ExpectQuery(`SELECT \* FROM "account_member"`).
		WithArgs(uint64(1), uint64(1)).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "approve_status", "account_id", "user_id", "is_admin", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, 1, 1, 1, true, time.Now(), time.Now(), nil),
		)
	s.Mock.ExpectQuery(`SELECT role_id FROM "?`).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}))

	accountObj := testAccountStub(1)
	accountObj.Approve = pkgModels.ApprovedApproveStatus
	userObj := testutil.Stub(1)
	userObj.Approve = pkgModels.ApprovedApproveStatus
	err := s.accountRepo.LoadPermissions(s.Ctx, accountObj, userObj)
	s.NoError(err)
	s.NotNil(accountObj.Permissions)
}

func (s *testSuite) TestFetchList() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1, 2, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "title", "description", "created_at"}).
				AddRow(1, 1, "title1", "description1", time.Now()).
				AddRow(2, 1, "title2", "description2", time.Now()),
		)
	accounts, err := s.accountRepo.FetchList(s.Ctx, &account.Filter{
		UserID: []uint64{1}, ID: []uint64{1, 2}})
	s.NoError(err)
	s.Equal(2, len(accounts))
}

func (s *testSuite) TestCount() {
	s.Mock.ExpectQuery("SELECT count").
		WithArgs(1, 2, 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(2),
		)
	count, err := s.accountRepo.Count(s.Ctx, &account.Filter{
		UserID: []uint64{1}, ID: []uint64{1, 2}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(
			pkgModels.UndefinedApproveStatus,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			nil,
			"test",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	acc := testAccountFromProfile("test")
	id, err := s.accountRepo.Create(s.Ctx, acc)
	s.NoError(err)
	s.Equal(uint64(101), id)
}

func (s *testSuite) TestUpdate() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), "test", uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	acc := testAccountFromProfile("test")
	err := s.accountRepo.Update(s.Ctx, 101, acc)
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.accountRepo.Delete(s.Ctx, 101)
	s.NoError(err)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
