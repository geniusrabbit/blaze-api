package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/mocks"
	usermocks "github.com/geniusrabbit/blaze-api/repository/user/mocks"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	userRepo       *usermocks.MockRepository
	accountRepo    *mocks.MockRepository
	accountUsecase account.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.userRepo = usermocks.NewMockRepository(ctrl)
	s.accountRepo = mocks.NewMockRepository(ctrl)
	s.accountUsecase = NewAccountUsecase(s.userRepo, s.accountRepo)
}

func (s *testSuite) TestGet() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(&model.Account{ID: 2}, nil)

	account, err := s.accountUsecase.Get(s.ctx, 2)
	s.NoError(err)
	s.Equal(uint64(2), account.ID)
}

func (s *testSuite) TestGetByTitle() {
	const title = "title"
	s.accountRepo.EXPECT().GetByTitle(s.ctx, title).
		Return(&model.Account{ID: 2, Title: title}, nil)

	account, err := s.accountUsecase.GetByTitle(s.ctx, title)
	s.NoError(err)
	s.Equal(uint64(2), account.ID)
	s.Equal(title, account.Title)
}

func (s *testSuite) TestGetCurrent() {
	s.accountRepo.EXPECT().Get(s.ctx, uint64(1)).
		Return(&model.Account{ID: 1}, nil)

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
		FetchList(s.ctx, gomock.AssignableToTypeOf(&account.Filter{}), nil, nil).
		Return([]*model.Account{{ID: 1}, {ID: 2}}, nil)

	accounts, err := s.accountUsecase.FetchList(s.ctx, &account.Filter{
		UserID: []uint64{1}, ID: []uint64{1, 2}}, nil, nil)
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
		Create(s.ctx, gomock.AssignableToTypeOf(&model.Account{})).
		Return(uint64(101), nil)

	id, err := s.accountUsecase.Store(s.ctx, &model.Account{ID: 0, Title: "test1"})
	s.NoError(err)
	s.Equal(id, uint64(101))
}

func (s *testSuite) TestUpdate() {
	s.accountRepo.EXPECT().
		Update(s.ctx, uint64(101), gomock.AssignableToTypeOf(&model.Account{})).
		Return(nil)

	_, err := s.accountUsecase.Store(s.ctx, &model.Account{ID: 101, Title: "test-test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.accountRepo.EXPECT().
		Get(s.ctx, uint64(1)).
		Return(&model.Account{ID: 1}, nil)
	s.accountRepo.EXPECT().
		Delete(s.ctx, gomock.AssignableToTypeOf(uint64(1))).
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

func (s *testSuite) TestFetchListMembers() {
	s.accountRepo.EXPECT().
		FetchListMembers(s.ctx, gomock.AssignableToTypeOf((*account.MemberFilter)(nil)), nil, nil).
		Return([]*model.AccountMember{{ID: 1}, {ID: 2}}, nil)

	members, err := s.accountUsecase.FetchListMembers(s.ctx,
		&account.MemberFilter{AccountID: []uint64{1}, UserID: []uint64{1, 2}},
		nil, nil,
	)

	s.NoError(err)
	s.Equal(2, len(members))
}

func (s *testSuite) TestLinkMember() {
	s.accountRepo.EXPECT().
		LinkMember(s.ctx, gomock.AssignableToTypeOf(&model.Account{}),
			true, gomock.AssignableToTypeOf(&model.User{})).
		Return(nil)

	account := &model.Account{ID: 1}
	user := &model.User{ID: 101}
	err := s.accountUsecase.LinkMember(s.ctx, account, true, user)
	s.NoError(err)
}

func (s *testSuite) TestUnlinkMember() {
	s.accountRepo.EXPECT().
		UnlinkMember(s.ctx, gomock.AssignableToTypeOf(&model.Account{}),
			gomock.AssignableToTypeOf(&model.User{})).
		Return(nil)

	account := &model.Account{ID: 1}
	user := &model.User{ID: 101}
	err := s.accountUsecase.UnlinkMember(s.ctx, account, user)
	s.NoError(err)
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
