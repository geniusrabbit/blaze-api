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
	"github.com/geniusrabbit/blaze-api/repository/rbac"
	"github.com/geniusrabbit/blaze-api/repository/rbac/mocks"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	roleRepo    *mocks.MockRepository
	roleUsecase rbac.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.roleRepo = mocks.NewMockRepository(ctrl)
	s.roleUsecase = New(s.roleRepo)
}

func (s *testSuite) TestGet() {
	s.roleRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(&model.Role{ID: 2}, nil)

	role, err := s.roleUsecase.Get(s.ctx, 2)
	s.NoError(err)
	s.Equal(uint64(2), role.ID)
}

func (s *testSuite) TestGetByName() {
	const name = "test"
	s.roleRepo.EXPECT().GetByName(s.ctx, name).
		Return(&model.Role{ID: 2, Name: name}, nil)

	role, err := s.roleUsecase.GetByName(s.ctx, name)
	s.NoError(err)
	s.Equal(uint64(2), role.ID)
	s.Equal(name, role.Name)
}

func (s *testSuite) TestGetGetError() {
	s.roleRepo.EXPECT().Get(s.ctx, uint64(2)).
		Return(nil, errors.New("test"))

	role, err := s.roleUsecase.Get(s.ctx, 2)
	s.Error(err)
	s.Nil(role)
}

func (s *testSuite) TestFetchList() {
	s.roleRepo.EXPECT().
		FetchList(s.ctx, gomock.AssignableToTypeOf(&rbac.Filter{}), nil, nil).
		Return([]*model.Role{{ID: 1}, {ID: 2}}, nil)

	roles, err := s.roleUsecase.FetchList(s.ctx, &rbac.Filter{ID: []uint64{1, 2}}, nil, nil)
	s.NoError(err)
	s.Equal(2, len(roles))
}

func (s *testSuite) TestCount() {
	s.roleRepo.EXPECT().
		Count(s.ctx, gomock.AssignableToTypeOf(&rbac.Filter{})).
		Return(int64(2), nil)

	count, err := s.roleUsecase.Count(s.ctx, &rbac.Filter{ID: []uint64{1, 2}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.roleRepo.EXPECT().
		Create(s.ctx, gomock.AssignableToTypeOf(&model.Role{})).
		Return(uint64(101), nil)

	id, err := s.roleUsecase.Create(s.ctx, &model.Role{ID: 0, Title: "test1"})
	s.NoError(err)
	s.Equal(id, uint64(101))
}

func (s *testSuite) TestUpdate() {
	s.roleRepo.EXPECT().
		Update(s.ctx, uint64(101), gomock.AssignableToTypeOf(&model.Role{})).
		Return(nil)

	err := s.roleUsecase.Update(s.ctx, 101, &model.Role{Title: "test-test"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.roleRepo.EXPECT().
		Get(s.ctx, gomock.AssignableToTypeOf(uint64(101))).
		Return(&model.Role{ID: 1}, nil)
	s.roleRepo.EXPECT().
		Delete(s.ctx, gomock.AssignableToTypeOf(uint64(101))).
		Return(nil)

	err := s.roleUsecase.Delete(s.ctx, 1)
	s.NoError(err)
}

func (s *testSuite) TestDeleteNotFound() {
	s.roleRepo.EXPECT().
		Get(s.ctx, gomock.AssignableToTypeOf(uint64(101))).
		Return(nil, sql.ErrNoRows)
	err := s.roleUsecase.Delete(s.ctx, 9999)
	s.EqualError(err, sql.ErrNoRows.Error())
}

func TestRoleUsecaseSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
