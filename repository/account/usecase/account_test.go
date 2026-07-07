package usecase_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/mocks"
	"github.com/geniusrabbit/blaze-api/repository/account/usecase"
	usermocks "github.com/geniusrabbit/blaze-api/repository/user/mocks"
	"github.com/geniusrabbit/blaze-api/repository/user/testutil"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	userRepo       *usermocks.MockRepository[*testutil.User]
	accountRepo    *mocks.MockSessionRepository[*testutil.User, *testAccount]
	memberRepo     *mocks.MockMemberRepository[*testutil.User, *testAccount]
	accountUsecase account.Usecase[*testutil.User, *testAccount]
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.userRepo = usermocks.NewMockRepository[*testutil.User](ctrl)
	s.accountRepo = mocks.NewMockSessionRepository[*testutil.User, *testAccount](ctrl)
	s.memberRepo = mocks.NewMockMemberRepository[*testutil.User, *testAccount](ctrl)
	s.accountUsecase = usecase.NewAccountUsecase(s.userRepo, s.accountRepo, s.memberRepo)
}

func (s *testSuite) TestGet() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(testAccountStub(2), nil)

	accountObj, err := s.accountUsecase.Get(s.ctx, 2)
	s.NoError(err)
	s.Equal(uint64(2), accountObj.ID)
}

func (s *testSuite) TestGetCurrent() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(1)).
		Return(testAccountStub(1), nil)

	accountObj, err := s.accountUsecase.Get(s.ctx, 1)
	s.NoError(err)
	s.Equal(uint64(1), accountObj.ID)
}

func (s *testSuite) TestGetGetError() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(nil, errors.New("test"))

	accountObj, err := s.accountUsecase.Get(s.ctx, 2)
	s.Error(err)
	s.Nil(accountObj)
}

func (s *testSuite) TestFetchList() {
	s.accountRepo.EXPECT().
		FetchList(s.ctx, gomock.AssignableToTypeOf(&account.Filter{})).
		Return([]*testAccount{testAccountStub(1), testAccountStub(2)}, nil)

	accounts, err := s.accountUsecase.FetchList(s.ctx, &account.Filter{
		UserID: []uint64{1}, ID: []uint64{1, 2}})
	s.NoError(err)
	s.Equal(2, len(accounts))
}

func (s *testSuite) TestCount() {
	s.accountRepo.EXPECT().
		Count(s.ctx, gomock.AssignableToTypeOf(&account.Filter{})).
		Return(int64(2), nil)

	count, err := s.accountUsecase.Count(s.ctx, &account.Filter{
		UserID: []uint64{1}, ID: []uint64{1, 2}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestUpdate() {
	s.accountRepo.EXPECT().
		Update(gomock.AssignableToTypeOf(s.ctx),
			uint64(101), gomock.AssignableToTypeOf(&testAccount{})).
		Return(nil)

	acc := testAccountStub(101)
	acc.Title = "test-test"
	_, err := s.accountUsecase.Update(s.ctx, acc)
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.accountRepo.EXPECT().
		Get(gomock.AssignableToTypeOf(s.ctx), uint64(1)).
		Return(testAccountStub(1), nil)
	s.accountRepo.EXPECT().
		Delete(gomock.AssignableToTypeOf(s.ctx), gomock.AssignableToTypeOf(uint64(1))).
		Return(nil)

	err := s.accountUsecase.Delete(s.ctx, 1)
	s.NoError(err)
}

func (s *testSuite) TestDeleteNotFound() {
	s.accountRepo.EXPECT().
		Get(s.ctx, uint64(9999)).
		Return(nil, sql.ErrNoRows)
	err := s.accountUsecase.Delete(s.ctx, 9999)
	s.EqualError(err, sql.ErrNoRows.Error())
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
