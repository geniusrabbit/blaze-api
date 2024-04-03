package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/geniusrabbit/gosql/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/option"
	"github.com/geniusrabbit/blaze-api/repository/option/mocks"
)

var testOption = model.Option{
	Name:     "opt.name",
	Type:     model.UserOptionType,
	TargetID: 1,
	Value:    *gosql.MustNullableJSON[any](map[string]any{"val": 1}),
}

type testSuite struct {
	suite.Suite

	ctx context.Context

	baseRepo    *mocks.MockRepository
	testUsecase option.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.baseRepo = mocks.NewMockRepository(ctrl)
	s.testUsecase = NewUsecase(s.baseRepo)
}

func (s *testSuite) TestGet() {
	s.baseRepo.EXPECT().Get(s.ctx, testOption.Name, testOption.Type, testOption.TargetID).
		Return(&testOption, nil)

	opt, err := s.testUsecase.Get(s.ctx, testOption.Name, testOption.Type, testOption.TargetID)
	s.NoError(err)
	s.Equal("opt.name", opt.Name)
}

func (s *testSuite) TestGetGetError() {
	s.baseRepo.EXPECT().Get(s.ctx, testOption.Name, testOption.Type, testOption.TargetID).
		Return(nil, errors.New("test"))

	opt, err := s.testUsecase.Get(s.ctx, testOption.Name, testOption.Type, testOption.TargetID)
	s.Error(err)
	s.Nil(opt)
}

func (s *testSuite) TestFetchList() {
	s.baseRepo.EXPECT().
		FetchList(s.ctx,
			gomock.AssignableToTypeOf(&option.Filter{}),
			gomock.AssignableToTypeOf(&option.ListOrder{}),
			gomock.AssignableToTypeOf(&repository.Pagination{})).
		Return([]*model.Option{{Name: "opt.name1"}, {Name: "opt.name2"}}, nil)

	roles, err := s.testUsecase.FetchList(s.ctx,
		&option.Filter{NamePattern: []string{"opt.%"}},
		&option.ListOrder{Name: model.OrderAsc},
		&repository.Pagination{Size: 100})
	s.NoError(err)
	s.Equal(2, len(roles))
}

func (s *testSuite) TestCount() {
	s.baseRepo.EXPECT().
		Count(s.ctx, gomock.AssignableToTypeOf(&option.Filter{})).
		Return(int64(2), nil)

	count, err := s.testUsecase.Count(s.ctx, &option.Filter{Name: []string{"opt.name1", "opt.name2"}})
	s.NoError(err)
	s.Equal(int64(2), count)
}

func (s *testSuite) TestCreate() {
	s.baseRepo.EXPECT().
		Set(s.ctx, gomock.AssignableToTypeOf(&model.Option{})).
		Return(nil)

	err := s.testUsecase.Set(s.ctx, &model.Option{Name: "test1"})
	s.NoError(err)
}

func (s *testSuite) TestDelete() {
	s.baseRepo.EXPECT().
		Get(s.ctx, testOption.Name, testOption.Type, testOption.TargetID).
		Return(&testOption, nil)
	s.baseRepo.EXPECT().
		Delete(s.ctx, testOption.Name, testOption.Type, testOption.TargetID).
		Return(nil)

	err := s.testUsecase.Delete(s.ctx, testOption.Name, testOption.Type, testOption.TargetID)
	s.NoError(err)
}

func (s *testSuite) TestDeleteNotFound() {
	s.baseRepo.EXPECT().
		Get(s.ctx, testOption.Name, testOption.Type, testOption.TargetID).
		Return(nil, sql.ErrNoRows)
	err := s.testUsecase.Delete(s.ctx, testOption.Name, testOption.Type, testOption.TargetID)
	s.EqualError(err, sql.ErrNoRows.Error())
}

func TestSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
