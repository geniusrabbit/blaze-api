package repository

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/testsuite"
)

type testMemberSuite struct {
	testsuite.DatabaseSuite

	memberRepo account.MemberRepository
}

func (s *testMemberSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.memberRepo = NewMemberRepository()
}

func (s *testMemberSuite) TestFetchListMembers() {
	s.Mock.ExpectQuery(`SELECT \* FROM "account_member"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "user_id", "account_id", "created_at"}).
				AddRow(1, 1, 101, 1, time.Now()).
				AddRow(2, 1, 102, 1, time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT \* FROM "account_base"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "approve_status", "title", "description", "updated_at", "created_at"}).
				AddRow(1, 1, "title", "desc", time.Now(), time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT \* FROM "m2m_account_member_role"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"member_id", "role_id", "created_at"}).
				AddRow(1, 1, time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT \* FROM "rbac_role"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "name", "created_at"}).
				AddRow(1, `test`, time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT \* FROM "account_user"`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "approve_status", "email", "created_at"}).
				AddRow(101, 1, "mail@", time.Now()).
				AddRow(102, 1, "mail@", time.Now()),
		)
	members, err := s.memberRepo.FetchListMembers(s.Ctx, nil, nil, nil)
	s.NoError(err)
	s.Equal(2, len(members))
}

func (s *testMemberSuite) TestIsMember() {
	ctx := s.Ctx
	s.Mock.ExpectQuery("SELECT").
		WithArgs(uint64(202), uint64(101)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	account := &model.Account{ID: 202}
	user := &model.User{ID: 101}
	ok := s.memberRepo.IsMember(ctx, user.ID, account.ID)
	s.True(ok)
}

func (s *testMemberSuite) TestIsAdmin() {
	ctx := s.Ctx
	s.Mock.ExpectQuery("SELECT").
		WithArgs(uint64(202), uint64(101)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_admin"}).AddRow(uint64(1), true))
	account := &model.Account{ID: 202}
	user := &model.User{ID: 101}
	ok := s.memberRepo.IsAdmin(ctx, user.ID, account.ID)
	s.True(ok)
}

func (s *testMemberSuite) TestLinkMember() {
	s.Mock.ExpectBegin()
	// stmt := s.Mock.ExpectPrepare("INSERT INTO")
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(model.ApprovedApproveStatus, uint64(101), uint64(101), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	s.Mock.ExpectQuery("INSERT INTO").
		WithArgs(model.ApprovedApproveStatus, uint64(101), uint64(102), true, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(102))
	s.Mock.ExpectCommit()

	account := &model.Account{ID: 101, Title: "test"}
	users := []*model.User{{ID: 101}, {ID: 102}}
	err := s.memberRepo.LinkMember(s.Ctx, account, true, users...)
	s.NoError(err)
}

func (s *testMemberSuite) TestUnlinkMember() {
	ctx := s.Ctx
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101), uint64(102)).
		WillReturnResult(sqlmock.NewResult(101, 2))
	account := &model.Account{ID: 101, Title: "test"}
	users := []*model.User{{ID: 101}, {ID: 102}}
	err := s.memberRepo.UnlinkMember(ctx, account, users...)
	s.NoError(err)
}

func TestMemberSuite(t *testing.T) {
	suite.Run(t, &testMemberSuite{})
}
