package usecase

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
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	usermocks "github.com/geniusrabbit/blaze-api/repository/user/mocks"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	userRepo       *usermocks.MockRepository
	accountRepo    *mocks.MockRepository
	memberRepo     *mocks.MockMemberRepository
	accountUsecase account.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.userRepo = usermocks.NewMockRepository(ctrl)
	s.accountRepo = mocks.NewMockRepository(ctrl)
	s.memberRepo = mocks.NewMockMemberRepository(ctrl)
	s.accountUsecase = NewAccountUsecase(s.userRepo, s.accountRepo, s.memberRepo)
}

func (s *testSuite) TestGet() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(&accountModels.Account{ID: 2}, nil)

	account, err := s.accountUsecase.Get(s.ctx, 2)
	s.NoError(err)
	s.Equal(uint64(2), account.ID)
}

func (s *testSuite) TestGetCurrent() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(1)).
		Return(&accountModels.Account{ID: 1}, nil)

	account, err := s.accountUsecase.Get(s.ctx, 1)
	s.NoError(err)
	s.Equal(uint64(1), account.ID)
}

func (s *testSuite) TestGetGetError() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(nil, errors.New("test"))

	account, err := s.accountUsecase.Get(s.ctx, 2)
	s.Error(err)
	s.Nil(account)
}

func (s *testSuite) TestFetchList() {
	s.accountRepo.EXPECT().
		FetchList(s.ctx, gomock.AssignableToTypeOf(&account.Filter{})).
		Return([]*accountModels.Account{{ID: 1}, {ID: 2}}, nil)

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

func (s *testSuite) TestStore() {
	s.accountRepo.EXPECT().
		Create(s.ctx, gomock.AssignableToTypeOf(&accountModels.Account{})).
		Return(uint64(101), nil)

	id, err := s.accountUsecase.Store(s.ctx, &accountModels.Account{ID: 0, Title: "test1"})
	s.NoError(err)
	s.Equal(id, uint64(101))
}

func (s *testSuite) TestUpdate() {
	s.accountRepo.EXPECT().
		Update(gomock.AssignableToTypeOf(s.ctx),
			uint64(101), gomock.AssignableToTypeOf(&accountModels.Account{})).
		Return(nil)

	_, err := s.accountUsecase.Store(s.ctx, &accountModels.Account{ID: 101, Title: "test-test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.accountRepo.EXPECT().
		Get(gomock.AssignableToTypeOf(s.ctx), uint64(1)).
		Return(&accountModels.Account{ID: 1}, nil)
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
