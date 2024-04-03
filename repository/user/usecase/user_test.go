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
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/user"
	"github.com/geniusrabbit/blaze-api/repository/user/mocks"
)

type userTestSuite struct {
	suite.Suite

	ctx context.Context

	userRepo    *mocks.MockRepository
	userUsecase user.Usecase
}

func (s *userTestSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.userRepo = mocks.NewMockRepository(ctrl)
	s.userUsecase = NewUserUsecase(s.userRepo)
}

func (s *userTestSuite) TestGet() {
	s.userRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(&model.User{ID: 2}, nil)

	user, err := s.userUsecase.Get(s.ctx, 2)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(2), user.ID)
}

func (s *userTestSuite) TestGetByEmail() {
	const email = "test@mail.com"
	s.userRepo.EXPECT().GetByEmail(s.ctx, email).
		Return(&model.User{ID: 2, Email: email}, nil)

	user, err := s.userUsecase.GetByEmail(s.ctx, email)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(2), user.ID)
	s.Assert().Equal(email, user.Email)
}

func (s *userTestSuite) TestGetCurrent() {
	// s.userRepo.EXPECT().Get(s.ctx, uint64(1)).
	// 	Return(&model.User{ID: 1}, nil)

	user, err := s.userUsecase.Get(s.ctx, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
}

func (s *userTestSuite) TestGetGetError() {
	s.userRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(nil, errors.New("test"))

	user, err := s.userUsecase.Get(s.ctx, 2)
	s.Assert().Error(err)
	s.Assert().Nil(user)
}

func (s *userTestSuite) TestGetByPassword() {
	s.userRepo.EXPECT().GetByPassword(s.ctx, "test@mail.com", "password").
		Return(&model.User{ID: 1}, nil)

	user, err := s.userUsecase.GetByPassword(s.ctx, "test@mail.com", "password")
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
}

func (s *userTestSuite) TestGetByToken() {
	s.userRepo.EXPECT().GetByToken(s.ctx, "token").
		Return(&model.User{ID: 1}, &model.Account{ID: 1}, nil)

	user, account, err := s.userUsecase.GetByToken(s.ctx, "token")
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), user.ID)
	s.Assert().Equal(uint64(1), account.ID)
}

func (s *userTestSuite) TestFetchList() {
	s.userRepo.EXPECT().
		FetchList(s.ctx, &user.ListFilter{AccountID: []uint64{1}},
			gomock.AssignableToTypeOf(&user.ListOrder{}),
			gomock.AssignableToTypeOf(&repository.Pagination{})).
		Return([]*model.User{{ID: 1}, {ID: 2}}, nil)

	users, err := s.userUsecase.FetchList(s.ctx, &user.ListFilter{AccountID: []uint64{1}}, nil, nil)
	s.Assert().NoError(err)
	s.Assert().Equal(2, len(users))
}

func (s *userTestSuite) TestCount() {
	s.userRepo.EXPECT().
		Count(s.ctx, &user.ListFilter{AccountID: []uint64{1}}).
		Return(int64(2), nil)

	count, err := s.userUsecase.Count(s.ctx, &user.ListFilter{AccountID: []uint64{1}})
	s.Assert().NoError(err)
	s.Assert().Equal(int64(2), count)
}

func (s *userTestSuite) TestFetchList_CurrentUser() {
	s.userRepo.EXPECT().
		FetchList(s.ctx, &user.ListFilter{AccountID: []uint64{1}},
			gomock.AssignableToTypeOf(&user.ListOrder{}),
			gomock.AssignableToTypeOf(&repository.Pagination{})).
		Return([]*model.User{{ID: 1}, {ID: 2}}, nil)

	users, err := s.userUsecase.FetchList(s.ctx, &user.ListFilter{AccountID: []uint64{1}}, nil, nil)
	s.Assert().NoError(err)
	s.Assert().Equal(2, len(users))
}

func (s *userTestSuite) TestCreate() {
	s.userRepo.EXPECT().
		Create(s.ctx, gomock.AssignableToTypeOf(&model.User{}), "password").
		Return(uint64(101), nil)

	id, err := s.userUsecase.Store(s.ctx, &model.User{Email: "test@mail.com"}, "password")
	s.Assert().NoError(err)
	s.Assert().Equal(id, uint64(101))
}

func (s *userTestSuite) TestUpdate() {
	s.userRepo.EXPECT().
		Update(s.ctx, gomock.AssignableToTypeOf(&model.User{})).
		Return(nil)

	err := s.userUsecase.Update(s.ctx, &model.User{ID: 101, Email: "test@mail.com"})
	s.Assert().NoError(err)
}

func (s *userTestSuite) TestDelete() {
	s.userRepo.EXPECT().
		Delete(s.ctx, gomock.AssignableToTypeOf(uint64(101))).
		Return(nil)

	err := s.userUsecase.Delete(s.ctx, 1)
	s.Assert().NoError(err)
}

func (s *userTestSuite) TestDelete_NotFound() {
	// s.userRepo.EXPECT().
	// 	Get(s.ctx, gomock.AssignableToTypeOf(uint64(101))).
	// 	Return(nil, sql.ErrNoRows)
	err := s.userUsecase.Delete(s.ctx, 9999)
	s.Assert().EqualError(err, sql.ErrNoRows.Error())
}

func TestUserUsecaseSuite(t *testing.T) {
	suite.Run(t, &userTestSuite{})
}
