package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/geniusrabbit/blaze-api/model"
	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository"
	"github.com/geniusrabbit/blaze-api/repository/historylog"
	"github.com/geniusrabbit/blaze-api/repository/historylog/mocks"
)

type testSuite struct {
	suite.Suite

	ctx context.Context

	testRepo    *mocks.MockRepository
	testUsecase historylog.Usecase
}

func (s *testSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.testRepo = mocks.NewMockRepository(ctrl)
	s.testUsecase = NewUsecase(s.testRepo)
}

func (s *testSuite) TestCount() {
	s.testRepo.EXPECT().
		Count(s.ctx, gomock.AssignableToTypeOf(&historylog.Filter{})).
		Return(int64(10), nil)

	count, err := s.testUsecase.Count(s.ctx, &historylog.Filter{})
	s.NoError(err)
	s.Equal(int64(10), count)
}

func (s *testSuite) TestFetchList() {
	uid1, uid2 := uuid.New(), uuid.New()
	s.testRepo.EXPECT().
		FetchList(s.ctx,
			gomock.AssignableToTypeOf(&historylog.Filter{}),
			gomock.AssignableToTypeOf(&historylog.Order{}),
			gomock.AssignableToTypeOf(&repository.Pagination{})).
		Return([]*model.HistoryAction{{ID: uid1}, {ID: uid2}}, nil)

	objs, err := s.testUsecase.FetchList(s.ctx,
		&historylog.Filter{ID: []uuid.UUID{uid1, uid2}},
		&historylog.Order{}, &repository.Pagination{Size: 100})
	s.NoError(err)
	s.Equal(2, len(objs))
}
func TestUsecaseSuite(t *testing.T) {
	suite.Run(t, &testSuite{})
}
