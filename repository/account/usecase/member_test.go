package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/geniusrabbit/blaze-api/pkg/context/session"
	"github.com/geniusrabbit/blaze-api/repository/account"
	"github.com/geniusrabbit/blaze-api/repository/account/mocks"
	accountModels "github.com/geniusrabbit/blaze-api/repository/account/models"
	usermocks "github.com/geniusrabbit/blaze-api/repository/user/mocks"
	userModels "github.com/geniusrabbit/blaze-api/repository/user/models"
)

type testMemberSuite struct {
	suite.Suite

	ctx context.Context

	userRepo      *usermocks.MockRepository
	accountRepo   *mocks.MockRepository
	memberRepo    *mocks.MockMemberRepository
	memberUsecase account.MemberUsecase
}

func (s *testMemberSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	s.ctx = session.WithUserAccountDevelop(context.TODO())
	s.userRepo = usermocks.NewMockRepository(ctrl)
	s.accountRepo = mocks.NewMockRepository(ctrl)
	s.memberRepo = mocks.NewMockMemberRepository(ctrl)
	s.memberUsecase = NewMemberUsecase(s.userRepo, s.accountRepo, s.memberRepo)
}

func (s *testMemberSuite) TestFetchListMembers() {
	s.memberRepo.EXPECT().
		FetchListMembers(s.ctx, gomock.AssignableToTypeOf((*account.MemberFilter)(nil))).
		Return([]*accountModels.AccountMember{{ID: 1}, {ID: 2}}, nil)

	members, err := s.memberUsecase.FetchListMembers(s.ctx,
		&account.MemberFilter{AccountID: []uint64{1}, UserID: []uint64{1, 2}},
	)

	s.NoError(err)
	s.Equal(2, len(members))
}

func (s *testMemberSuite) TestLinkMember() {
	s.memberRepo.EXPECT().
		LinkMember(s.ctx, gomock.AssignableToTypeOf(&accountModels.Account{}),
			true, gomock.AssignableToTypeOf(&userModels.User{})).
		Return(nil)

	account := &accountModels.Account{ID: 1}
	user := &userModels.User{ID: 101}
	err := s.memberUsecase.LinkMember(s.ctx, account, true, user)
	s.NoError(err)
}

func (s *testMemberSuite) TestUnlinkMember() {
	s.memberRepo.EXPECT().
		UnlinkMember(s.ctx, gomock.AssignableToTypeOf(&accountModels.Account{}),
			gomock.AssignableToTypeOf(&userModels.User{})).
		Return(nil)

	account := &accountModels.Account{ID: 1}
	user := &userModels.User{ID: 101}
	err := s.memberUsecase.UnlinkMember(s.ctx, account, user)
	s.NoError(err)
}

func TestAccountMemberSuite(t *testing.T) {
	suite.Run(t, &testMemberSuite{})
}
