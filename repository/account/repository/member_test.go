package repository_test

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
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

type testMemberSuite struct {
	testsuite.DatabaseSuite

	memberRepo account.MemberRepository[*testutil.User, *testAccount]
}

func (s *testMemberSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.memberRepo = accountrepo.NewMemberRepositoryFor(func() *account.Member[*testutil.User, *testAccount] {
		return new(account.Member[*testutil.User, *testAccount])
	})
}

func (s *testMemberSuite) SetupTest() {
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

func (s *testMemberSuite) TestFetchListMembers() {
	s.Mock.ExpectQuery(`SELECT \* FROM "account_member"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "approve_status", "user_id", "account_id", "is_admin", "created_at", "updated_at", "deleted_at"}).
				AddRow(1, 1, 101, 1, false, time.Now(), time.Now(), nil).
				AddRow(2, 1, 102, 1, false, time.Now(), time.Now(), nil),
		)
	members, err := s.memberRepo.FetchListMembers(s.Ctx)
	s.NoError(err)
	s.Equal(2, len(members))
}

func (s *testMemberSuite) TestIsMember() {
	ctx := s.Ctx
	s.Mock.ExpectQuery(`SELECT count\(\*\) FROM "account_member"`).
		WithArgs(uint64(202), uint64(101)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	accountObj := testAccountStub(202)
	user := testutil.Stub(101)
	ok := s.memberRepo.IsMember(ctx, user.ID, accountObj.ID)
	s.True(ok)
}

func (s *testMemberSuite) TestIsAdmin() {
	ctx := s.Ctx
	s.Mock.ExpectQuery(`SELECT \* FROM "account_member"`).
		WithArgs(uint64(202), uint64(101), 1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "approve_status", "account_id", "user_id", "is_admin", "created_at", "updated_at", "deleted_at"}).
				AddRow(uint64(1), 1, 202, 101, true, time.Now(), time.Now(), nil),
		)
	accountObj := testAccountStub(202)
	user := testutil.Stub(101)
	ok := s.memberRepo.IsAdmin(ctx, user.ID, accountObj.ID)
	s.True(ok)
}

func (s *testMemberSuite) TestLinkMember() {
	s.Mock.ExpectBegin()
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(pkgModels.ApprovedApproveStatus, uint64(101), uint64(101), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(pkgModels.ApprovedApproveStatus, uint64(101), uint64(102), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(102))
	s.Mock.ExpectCommit()

	accountObj := testAccountStub(101)
	accountObj.Title = "test"
	users := []*testutil.User{testutil.Stub(101), testutil.Stub(102)}
	err := s.memberRepo.LinkMember(s.Ctx, accountObj, true, users...)
	s.NoError(err)
}

func (s *testMemberSuite) TestUnlinkMember() {
	ctx := s.Ctx
	s.Mock.ExpectExec(`UPDATE "account_member" SET "deleted_at"`).
		WithArgs(sqlmock.AnyArg(), uint64(101), uint64(101), uint64(102)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	accountObj := testAccountStub(101)
	accountObj.Title = "test"
	users := []*testutil.User{testutil.Stub(101), testutil.Stub(102)}
	err := s.memberRepo.UnlinkMember(ctx, accountObj, users...)
	s.NoError(err)
}

func TestMemberSuite(t *testing.T) {
	suite.Run(t, &testMemberSuite{})
}
