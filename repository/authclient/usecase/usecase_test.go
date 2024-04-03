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
	"github.com/geniusrabbit/blaze-api/repository/authclient"
	"github.com/geniusrabbit/blaze-api/repository/authclient/mocks"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	authclientRepo    *mocks.MockRepository
	authclientUsecase authclient.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.authclientRepo = mocks.NewMockRepository(ctrl)
	s.authclientUsecase = NewAuthclientUsecase(s.authclientRepo)
}

func (s *testSuite) TestGet() {
	s.authclientRepo.EXPECT().Get(s.ctx, "2").
		Return(&model.AuthClient{ID: "2"}, nil)

	role, err := s.authclientUsecase.Get(s.ctx, "2")
	s.NoError(err)
	s.Equal("2", role.ID)
}

func (s *testSuite) TestGetGetError() {
	s.authclientRepo.EXPECT().Get(s.ctx, "2").
		Return(nil, errors.New("test"))

	role, err := s.authclientUsecase.Get(s.ctx, "2")
	s.Error(err)
	s.Nil(role)
}

func (s *testSuite) TestFetchList() {
	s.authclientRepo.EXPECT().
		FetchList(s.ctx, gomock.AssignableToTypeOf(&authclient.Filter{})).
		Return([]*model.AuthClient{{ID: "1"}, {ID: "2"}}, nil)

	roles, err := s.authclientUsecase.FetchList(s.ctx, &authclient.Filter{
		ID: []string{"1", "2"}, PageSize: 100})
	s.NoError(err)
	s.Equal(2, len(roles))
}

func (s *testSuite) TestCount() {
	s.authclientRepo.EXPECT().
		Count(s.ctx, gomock.AssignableToTypeOf(&authclient.Filter{})).
		Return(int64(2), nil)

	count, err := s.authclientUsecase.Count(s.ctx, &authclient.Filter{ID: []string{"1", "2"}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.authclientRepo.EXPECT().
		Create(s.ctx, gomock.AssignableToTypeOf(&model.AuthClient{})).
		Return("101", nil)

	id, err := s.authclientUsecase.Create(s.ctx, &model.AuthClient{ID: "", Title: "test1"})
	s.NoError(err)
	s.Equal(id, "101")
}

func (s *testSuite) TestUpdate() {
	s.authclientRepo.EXPECT().
		Update(s.ctx, "101", gomock.AssignableToTypeOf(&model.AuthClient{})).
		Return(nil)

	err := s.authclientUsecase.Update(s.ctx, "101", &model.AuthClient{Title: "test-test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.authclientRepo.EXPECT().
		Get(s.ctx, "1").
		Return(&model.AuthClient{ID: "1"}, nil)
	s.authclientRepo.EXPECT().
		Delete(s.ctx, gomock.AssignableToTypeOf("101")).
		Return(nil)

	err := s.authclientUsecase.Delete(s.ctx, "1")
	s.NoError(err)
}

func (s *testSuite) TestDeleteNotFound() {
	s.authclientRepo.EXPECT().
		Get(s.ctx, "9999").
		Return(nil, sql.ErrNoRows)
	err := s.authclientUsecase.Delete(s.ctx, "9999")
	s.EqualError(err, sql.ErrNoRows.Error())
}

func TestRoleSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
