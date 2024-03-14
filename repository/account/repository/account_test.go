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

type testSuite struct {
	testsuite.DatabaseSuite

	accountRepo account.Repository
}

func (s *testSuite) SetupSuite() {
	s.DatabaseSuite.SetupSuite()
	s.accountRepo = New()
}

func (s *testSuite) TestGet() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "title", "description", "created_at"}).
				AddRow(1, 1, "title1", "description1", time.Now()),
		)
	account, err := s.accountRepo.Get(s.Ctx, 1)
	s.NoError(err)
	s.Equal(uint64(1), account.ID)
}

func (s *testSuite) TestGetByTitle() {
	s.Mock.ExpectQuery("SELECT *").
		WithArgs("title1").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "title", "description", "created_at"}).
				AddRow(1, 1, "title1", "description1", time.Now()),
		)
	account, err := s.accountRepo.GetByTitle(s.Ctx, "title1")
	s.NoError(err)
	s.Equal(uint64(1), account.ID)
	s.Equal("title1", account.Title)
}

func (s *testSuite) TestLoadPermissions() {
	s.Mock.ExpectQuery(`SELECT \* FROM "?`+(*model.AccountMember)(nil).TableName()).
		WithArgs(uint64(1), uint64(1)).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "account_id", "user_id", "is_admin", "created_at"}).
				AddRow(1, 1, 1, 1, true, time.Now()),
		)
	s.Mock.ExpectQuery(`SELECT role_id FROM "?` + (*model.M2MAccountMemberRole)(nil).TableName()).
		WithArgs(uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"role_id"}))

	account := &model.Account{ID: 1, Approve: model.ApprovedApproveStatus}
	user := &model.User{ID: 1, Approve: model.ApprovedApproveStatus}
	err := s.accountRepo.LoadPermissions(s.Ctx, account, user)
	s.NoError(err)
	s.NotNil(account.Permissions)
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
		UserID: []uint64{1}, ID: []uint64{1, 2}}, nil, nil)
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
		WithArgs(sqlmock.AnyArg(), "test", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))
	id, err := s.accountRepo.Create(
		s.Ctx,
		&model.Account{
			ID:    101,
			Title: "test",
		})
	s.NoError(err)
	s.Equal(uint64(101), id)
}

func (s *testSuite) TestUpdate() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs("test", sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.accountRepo.Update(
		s.Ctx,
		101, &model.Account{Title: "test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101)).
		WillReturnResult(sqlmock.NewResult(101, 1))
	err := s.accountRepo.Delete(s.Ctx, 101)
	s.NoError(err)
}

func (s *testSuite) TestFetchListMembers() {
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
	members, err := s.accountRepo.FetchListMembers(s.Ctx, nil, nil, nil)
	s.NoError(err)
	s.Equal(2, len(members))
}

func (s *testSuite) TestIsMember() {
	ctx := s.Ctx
	s.Mock.ExpectQuery("SELECT").
		WithArgs(uint64(202), uint64(101)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uint64(1)))
	account := &model.Account{ID: 202}
	user := &model.User{ID: 101}
	ok := s.accountRepo.IsMember(ctx, user.ID, account.ID)
	s.True(ok)
}

func (s *testSuite) TestIsAdmin() {
	ctx := s.Ctx
	s.Mock.ExpectQuery("SELECT").
		WithArgs(uint64(202), uint64(101)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_admin"}).AddRow(uint64(1), true))
	account := &model.Account{ID: 202}
	user := &model.User{ID: 101}
	ok := s.accountRepo.IsAdmin(ctx, user.ID, account.ID)
	s.True(ok)
}

func (s *testSuite) TestLinkMember() {
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
	err := s.accountRepo.LinkMember(s.Ctx, account, true, users...)
	s.NoError(err)
}

func (s *testSuite) TestUnlinkMember() {
	ctx := s.Ctx
	s.Mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), uint64(101), uint64(102)).
		WillReturnResult(sqlmock.NewResult(101, 2))
	account := &model.Account{ID: 101, Title: "test"}
	users := []*model.User{{ID: 101}, {ID: 102}}
	err := s.accountRepo.UnlinkMember(ctx, account, users...)
	s.NoError(err)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
